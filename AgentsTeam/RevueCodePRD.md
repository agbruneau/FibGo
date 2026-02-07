# Rapport de Revue de Code Consolide — FibGo

**Date** : 2026-02-07
**Equipe** : 4 reviewers specialises (perf, archi, fiabilite, tests)

## Vue d'ensemble

| Reviewer | Critiques | Moderes | Suggestions | Positifs |
|----------|-----------|---------|-------------|----------|
| **perf** | 3 | 7 | 5 | 8 |
| **archi** | 2 | 6 | 5 | 9 |
| **fiabilite** | 0 | 5 | 6 | 10 |
| **tests** | 2 | 6 | 5 | 10+ |
| **Total** | **7** | **24** | **21** | **37** |

**Verdict global** : Codebase de haute qualite avec une architecture bien structuree, des patterns GoF correctement implementes, et une fiabilite solide (0 critique fiabilite). Les problemes critiques sont concentres sur du **code mort dangereux** (Karatsuba, ASM AVX2), un **cache FFT contre-productif** (SHA-256 + deep copy), et des **incoherences de naming/couplage** architecturales.

---

## PROBLEMES CRITIQUES (7) — Par priorite d'action

### P1. Cache FFT : SHA-256 + deep copy annulent le benefice (perf C2+C3)

**Fichiers** : `internal/bigfft/fft_cache.go`

- `computeKey()` utilise SHA-256 (crypto) pour des cles de cache LRU — surdimensionne 10-50x
- `flattenPolyData()` alloue O(n) juste pour le hashing
- `Get()` fait une deep copy de ~4MB par cache hit (128 fermats * 4097 mots * 8 bytes)
- `Put()` fait la meme deep copy

**Recoupement fiabilite** : Pas de race condition ici (mutex protege le cache), mais l'overhead memoire est confirme.

**Action** : Profiler avec/sans cache. Si le cache est benefique, remplacer SHA-256 par xxhash/FNV-1a et envisager un COW (copy-on-write) au lieu du deep copy.

### P2. Karatsuba custom : code mort 3-3.5x plus lent, toujours expose publiquement (perf C1)

**Fichier** : `internal/bigfft/karatsuba.go`

- `bolt.md` documente que ce code est plus lent que `math/big.Mul` a toutes les tailles
- ~970 allocations recursives par multiplication (add/sub/assemble)
- `KaratsubaMultiply()` reste exporte — un appelant externe pourrait l'utiliser par erreur
- `smartMultiply()` ne l'utilise plus (transition directe math/big -> FFT), confirmant qu'il est mort

**Recoupement archi** : Le code mort viole le principe YAGNI et pollue l'API publique.

**Action** : Supprimer ou rendre `internal` (unexport). Si conserve pour reference, le marquer `Deprecated`.

### P3. Assembleur AVX2 : 226 lignes de code mort + dispatch no-op (perf M1+M2)

**Fichiers** : `internal/bigfft/arith_amd64.s`, `cpu_amd64.go`

- `selectAVX2Impl()` est un no-op (corps vide) — jamais active
- `addVVAvx2`/`subVVAvx2` sont des boucles scalaires deguisees en AVX2
- `bolt.md` confirme que c'est intentionnel (2-2.5x plus lent que math/big)
- Infrastructure de dispatch entiere (function pointers, `UseAVX2()`, `UseDefault()`) inutilisee

**Action** : Nettoyer le code ASM mort et l'infrastructure de dispatch.

### P4. Confusion nominale `ProgressReporter` : deux types homonymes (archi 1.1)

**Fichiers** : `internal/fibonacci/progress.go:25` vs `internal/orchestration/interfaces.go:19`

- `fibonacci.ProgressReporter` = type fonction `func(float64)`
- `orchestration.ProgressReporter` = interface avec `DisplayProgress()`
- Meme nom, semantiques completement differentes

**Recoupement fiabilite** : Les deux types coexistent sans probleme technique, mais le risque de confusion est reel.

**Action** : Renommer `fibonacci.ProgressReporter` -> `ProgressCallback` ou `ProgressFunc`.

### P5. `config` importe `fibonacci` — couplage ascendant (archi 1.2)

**Fichier** : `internal/config/config.go:12`

- `config.ToCalculationOptions()` retourne `fibonacci.Options` -> la config connait le business
- Constantes dupliquees : `DefaultFFTThreshold` dans `config` ET `fibonacci/constants.go`
- `app.go:99` compare `cfg.Threshold == fibonacci.DefaultParallelThreshold`

**Action** : Deplacer `Options` dans un package partage ou inverser la dependance.

### P6. Race detector impossible sans CGO (tests C2)

- `go test -race` echoue sur Windows sans CGO
- Les tests de concurrence existent mais ne prouvent rien sans le race detector
- Impact direct sur la confiance dans les findings de **fiabilite**

**Recoupement fiabilite** : M3 (ProgressState non thread-safe) et M4 (programRef.Send race) seraient detectables par le race detector s'il fonctionnait.

**Action** : Activer CGO dans le CI ou tester sur Linux avec `-race`.

### P7. `cmd/generate-golden` a 29.4% de couverture (tests C1)

- Seul package sous le seuil de 75%
- `main()`, generation JSON, ecriture fichier non testes

**Action** : Ajouter des tests ou exclure explicitement du seuil de couverture (outil utilitaire).

---

## PROBLEMES MODERES (24) — Regroupes par theme

### Concurrence & Annulation (fiabilite M1, M2 + archi 2.6)

- Goroutines de multiplication parallele (`fft.go:115-182`, `doubling_framework.go:64-108`) non annulables par le contexte — l'app peut mettre du temps a repondre a SIGINT sur de grands N
- `errgroup.WithContext` dans l'orchestrateur ne sert qu'au contexte parent (les goroutines retournent toujours `nil`) — un `sync.WaitGroup` simple serait plus honnete

### Thread-safety fragile (fiabilite M3, M4)

- `ProgressState`/`ProgressWithETA` : commentaire "designed to be safe for concurrent use" incorrect (single goroutine en pratique)
- `programRef.program` dans le TUI : pas de barriere memoire explicite lors de l'ecriture/lecture

### Performance (perf M3-M7)

- Triple semaphore (Fibonacci `NumCPU*2` + FFT `NumCPU` + Karatsuba) = potentiel `NumCPU*3` goroutines
- `acquireWordSlice` fait un `clear()` O(n) meme quand inutile (jusqu'a 128MB)
- `getWordSlicePoolIndex` : recherche lineaire au lieu de bitwise O(1)
- `polyFromNat` : allocation de slice headers non poolee
- `DynamicThresholdManager.ShouldAdjust()` : Lock exclusif a chaque iteration (impact faible)

### Architecture (archi 2.1-2.5)

- TUI importe CLI (`bridge.go:10`) — couplage horizontal entre presentations
- `MultiplicationStrategy.ExecuteStep()` prend `*CalculationState` — viole ISP
- `GlobalFactory()` retourne `*DefaultFactory` (concret) au lieu de `CalculatorFactory` (interface)
- `CalculatorFactory.Register()` expose `coreCalculator` (unexported) — semi-ouvert
- `GetCalculatorsToRun` dans le package CLI au lieu d'orchestration

### Tests (tests M3-M8)

- Tests FFT avec `t.Logf` au lieu de `t.Error` — ne peuvent pas echouer
- `TestFFTThresholdVariations` : verification potentiellement executee avant les sous-tests paralleles
- Pas de test pour `N = math.MaxUint64` avec timeout/cancellation
- Mocks gomock generes mais jamais utilises (code mort)
- E2E tests minimalistes (2 cas seulement)
- Tests d'env vars sans `t.Parallel()` (correct mais fragile)

### Fiabilite (fiabilite M5)

- `ErrorCollector` ne capture que la premiere erreur (intentionnel, mais limite le debug)

---

## RECOUPEMENTS INTER-REVIEWERS

### Performance x Fiabilite/Architecture

| Trouvaille perf | Challenge fiabilite/archi | Verdict |
|-----------------|--------------------------|---------|
| C1: Karatsuba mort expose | archi: viole YAGNI, pollue l'API | **Supprimer** |
| C2+C3: Cache FFT overhead | fiabilite: pas de race, mais overhead confirme | **Profiler puis decider** |
| M1: AVX2 no-op | archi: code mort = dette technique | **Nettoyer** |
| M3: Triple semaphore | fiabilite: pas de deadlock detecte, mais contention possible | **Documenter l'invariant** |
| P7: Semaphore non-bloquant FFT | fiabilite P7: confirme excellent pattern | **Conserver** |

### Architecture x Tests

| Trouvaille archi | Challenge tests | Verdict |
|------------------|-----------------|---------|
| 2.6: errgroup retourne nil | tests: orchestration couverte a 98% | **Refactorer vers WaitGroup ou documenter** |
| 1.1: ProgressReporter homonyme | tests: bien couvert mais confus | **Renommer** |

### Fiabilite x Tests

| Trouvaille fiabilite | Challenge tests | Verdict |
|----------------------|-----------------|---------|
| M1+M2: goroutines non annulables | tests C2: race detector impossible | **Risque non verifiable actuellement** |
| M3: ProgressState non thread-safe | tests: pas de test concurrent dessus | **Ajouter test ou corriger commentaire** |

---

## ACTIONS PRIORITAIRES

### Priorite 1 — Quick wins (effort faible, impact eleve)

1. Renommer `fibonacci.ProgressReporter` -> `ProgressCallback`
2. Supprimer/unexport `KaratsubaMultiply()` dans bigfft
3. Corriger les assertions `t.Logf` -> `t.Errorf` dans `fft_test.go`
4. Supprimer les mocks gomock inutilises

### Priorite 2 — Nettoyage (effort moyen)

5. Nettoyer le code assembleur AVX2 mort + infrastructure dispatch
6. Profiler le cache FFT (avec/sans) et decider de son sort
7. Deplacer `GetCalculatorsToRun` de CLI vers orchestration
8. Extraire les utilitaires partages CLI/TUI dans `internal/format`

### Priorite 3 — Ameliorations structurelles (effort eleve)

9. Resoudre le couplage `config` -> `fibonacci` (inverser la dependance)
10. Separer `MultiplicationStrategy` en `Multiplier` + `DoublingStepExecutor`
11. Propager le contexte dans les goroutines de multiplication parallele
12. Activer le race detector dans le CI (CGO ou Linux)
13. Enrichir les tests E2E (8+ scenarios manquants identifies)

---

## POINTS FORTS DU CODEBASE

Le codebase presente des qualites remarquables unanimement saluees par les 4 reviewers :

- **Architecture en couches exemplaire** — direction des dependances respectee, decouplage par interfaces
- **Patterns GoF bien implementes** — Decorator, Factory+Registry, Strategy, Observer, Null Object
- **Fiabilite solide** — 0 critique fiabilite, signal handling idiomatique, panic recovery dans bigfft, sync.Pool avec eviction
- **Tests de haute qualite** — fuzz testing, property-based (4 identites mathematiques), golden files, couverture >75% partout sauf un outil utilitaire
- **Optimisations intelligentes** — bump allocator, swap de pointeurs big.Int, FFT transform reuse, semaphore non-bloquant
- **Code testable** — injection de dependances, `run()` pattern dans main, config parametrable

---

## RAPPORTS DETAILLES PAR REVIEWER

### Reviewer Performance (perf)

#### Problemes Critiques

**C1. Custom Karatsuba dans `bigfft/karatsuba.go` : toujours present et allocationnel**

Fichier : `internal/bigfft/karatsuba.go`, lignes 167-235

Le fichier `bolt.md` documente clairement que le Karatsuba custom est 3-3.5x plus lent que `math/big.Mul` a toutes les tailles d'operandes (512 a 100K bits), a cause de ~970 allocations recursives par multiplication.

Allocations identifiees dans le hot path :
- Ligne 266 : `z := make(nat, len(x)+1)` dans `add()` — nouvelle allocation a chaque appel recursif
- Ligne 277 : `z := make(nat, len(x))` dans `sub()` — nouvelle allocation a chaque appel recursif
- Ligne 315 : `res := make(nat, size+1)` dans `assemble()` — nouvelle allocation a chaque combinaison
- Ligne 247 : `result := make(nat, len(x)+m)` dans `multiplyAsymmetric()` — allocation de base

**C2. FFT Transform Cache : SHA-256 hashing sur le hot path**

Fichier : `internal/bigfft/fft_cache.go`, lignes 101-120

Le cache utilise SHA-256 pour calculer les cles. Pour un operande de 500K bits (~8000 mots) :
1. Allocation d'un `sha256.New()` + etat interne
2. Allocation de `buf := make([]byte, 8)` a chaque appel
3. 8000 iterations de hashing SHA-256

De plus, `flattenPolyData()` (lignes 309-323) cree une copie complete des donnees du polynome juste pour le hashing.

**C3. Cache Get() : Deep copy sur cache hit**

Fichier : `internal/bigfft/fft_cache.go`, lignes 148-160

A chaque cache hit, deep copy complete de tous les PolValues. Pour K=128 et n=4096 mots : 128 * 4097 * 8 = ~4MB de copies. Le `Put()` fait la meme chose.

#### Problemes Moderes

**M1. `selectAVX2Impl()` est un no-op**

Fichier : `internal/bigfft/cpu_amd64.go`, lignes 227-229. Fonction vide — les dispatchers AVX2 ne sont jamais actives.

**M2. Assembleur addVV/subVV non deroule — pas d'avantage sur scalar**

Fichier : `internal/bigfft/arith_amd64.s`, lignes 26-110. Boucles scalaires (ADC/SBB) sans deroulement, fonctionnellement identiques aux implementations Go `go:linkname`.

**M3. Triple semaphore independant**

`common.go:26` (NumCPU*2) vs `fft_recursion.go:17` (NumCPU) vs Karatsuba. Jusqu'a NumCPU*3 goroutines simultanees.

**M4. `acquireWordSlice` fait un `clear()` complet meme quand inutile**

Fichier : `internal/bigfft/pool.go`, lignes 54-64. Memset potentiellement 128MB par allocation de grande classe.

**M5. `getWordSlicePoolIndex` : recherche lineaire**

Fichier : `internal/bigfft/pool.go`, lignes 35-42. Puisque les tailles sont des puissances de 4, un `bits.Len()` permettrait O(1).

**M6. `polyFromNat` alloue des slices non poolees**

Fichier : `internal/bigfft/fft_poly.go`, lignes 19-38.

**M7. `DynamicThresholdManager.ShouldAdjust()` prend un mutex Lock a chaque check**

Fichier : `internal/fibonacci/dynamic_threshold.go`, lignes 158-191. Impact faible (iterations logarithmiques).

#### Suggestions

- S1. Bump allocator fallback silencieux — ajouter un compteur de fallbacks
- S2. Pool pre-warming potentiellement obsolete avec le bump allocator
- S3. `fermat.Mul` : transition big.Int OK grace a l'escape analysis
- S4. Seuils FFT internes hardcodes (non ajustables dynamiquement)
- S5. `defer` dans les hot paths FFT — OK avec Go 1.25+

#### Points Positifs

- P1. Architecture smartMultiply a 2 niveaux excellente (math/big -> bigfft directe)
- P2. Bump allocator bien concu avec fallback transparent
- P3. Object pooling du CalculationState avec guard MaxPooledBitLen = 4M bits
- P4. Pattern de swap des big.Int (rotation de pointeurs sans copie)
- P5. FFT transform reuse dans executeDoublingStepFFT (3 multiplications pointwise)
- P6. Pool sizing bien calibre (classes facteur 4x, 64 mots a 16M mots)
- P7. Parallelisation avec semaphore non-bloquant dans FFT
- P8. Contention eliminee entre parallelisme Fibonacci et FFT

---

### Reviewer Architecture (archi)

#### Problemes Critiques

**1.1 Confusion nominale ProgressReporter**

`fibonacci.ProgressReporter` (type fonction) vs `orchestration.ProgressReporter` (interface). Deux abstractions totalement differentes avec le meme nom.

**1.2 config importe fibonacci — violation de la direction des dependances**

`config.ToCalculationOptions()` retourne `fibonacci.Options`, creant un couplage descendant. Constantes dupliquees entre les packages.

#### Problemes Moderes

- 2.1 TUI importe CLI (`bridge.go:10`) — couplage horizontal
- 2.2 Interface `MultiplicationStrategy` trop large (ISP) — `ExecuteStep` couple au fast doubling
- 2.3 Variable globale mutable `globalFactory` — `GlobalFactory()` retourne un type concret
- 2.4 `CalculatorFactory` expose `coreCalculator` (unexported) — interface semi-ouverte
- 2.5 `GetCalculatorsToRun` dans le package CLI — logique metier mal placee
- 2.6 `ExecuteCalculations` retourne toujours nil dans errgroup — `sync.WaitGroup` suffirait

#### Suggestions

- 3.1 `CalculationState` leaks implementation details dans l'interface Strategy
- 3.2 Panic dans `NewCalculator` au lieu de retourner une erreur
- 3.3 Observer synchrone dans `ProgressSubject.Notify` — risque de blocage si observer bloque
- 3.4 Duplication des constantes de seuil (fibonacci/constants.go, config/config.go, app.go)
- 3.5 `Register` retourne toujours nil

#### Points Positifs

- 4.1 Decorator pattern bien implemente (FibCalculator)
- 4.2 Factory + Registry avec double-check locking correct
- 4.3 Observer pattern flexible et bien decouple
- 4.4 Interface-based decoupling pour la presentation
- 4.5 Generics bien utilises pour les taches (`executeTasks[T any, PT]`)
- 4.6 Configuration testable (parametres injectes)
- 4.7 Architecture en couches respectee (direction des dependances)
- 4.8 Entrypoint testable (`run()` pattern)
- 4.9 Conventions de nommage coherentes (Display*/Format*/Write*)

---

### Reviewer Fiabilite (fiabilite)

#### Problemes Critiques

Aucun probleme critique bloquant detecte.

#### Problemes Moderes

- M1. Goroutines non annulables dans `executeDoublingStepFFT` (parallel path) — pas de propagation du contexte
- M2. Goroutines dans `executeDoublingStepMultiplications` sans contexte — `wg.Wait()` bloque meme si contexte annule
- M3. `ProgressState` et `ProgressWithETA` non thread-safe — commentaire "designed to be safe" incorrect
- M4. `programRef.Send` possible race sur `program` dans le TUI — pas de barriere memoire explicite
- M5. `ErrorCollector` ne capture que la premiere erreur — limite le debug multi-erreur

#### Suggestions

- S1. Ajouter contexte dans les goroutines paralleles de multiplication
- S2. Pas de panic recovery dans les goroutines de `executeTasks`
- S3. `ChannelObserver` drop silencieusement les updates — ajouter un compteur
- S4. `fftState` pool ne libere jamais les buffers fermat internes (intentionnel mais non documente)
- S5. Generation counter overflow theorique dans le TUI (academique)
- S6. `Unregister` du `ProgressSubject` jamais utilise (code mort)

#### Points Positifs

- P1. Excellent signal handling (`signal.NotifyContext`)
- P2. Orchestration solide avec `errgroup` (close + displayWg.Wait)
- P3. Panic recovery systematique dans bigfft (4 fonctions publiques)
- P4. Error wrapping correct et coherent (`%w`, `Unwrap()`, `errors.Is`)
- P5. sync.Pool bien utilise avec eviction des gros objets (MaxPooledBitLen)
- P6. Non-blocking channel send dans ChannelObserver
- P7. Task semaphore pour limiter la concurrence (`sync.Once`)
- P8. Context checks dans la boucle de calcul
- P9. TUI generation counter pour les messages stales
- P10. Separation propre des interfaces de presentation

---

### Reviewer Tests (tests)

#### Couverture

| Package | Couverture | Statut |
|---------|-----------|--------|
| `internal/errors` | 100.0% | OK |
| `internal/parallel` | 100.0% | OK |
| `internal/sysmon` | 100.0% | OK |
| `internal/testutil` | 100.0% | OK |
| `internal/ui` | 100.0% | OK |
| `internal/config` | 99.2% | OK |
| `internal/orchestration` | 98.0% | OK |
| `internal/metrics` | 95.7% | OK |
| `internal/tui` | 93.6% | OK |
| `internal/fibonacci` | 88.7% | OK |
| `internal/bigfft` | 88.6% | OK |
| `internal/calibration` | 88.5% | OK |
| `internal/app` | 88.0% | OK |
| `internal/cli` | 86.3% | OK |
| `cmd/fibcalc` | 75.0% | Seuil |
| `cmd/generate-golden` | **29.4%** | SOUS SEUIL |

#### Problemes Critiques

- C1. `cmd/generate-golden` a 29.4% de couverture
- C2. Race detector impossible sans CGO sur Windows

#### Problemes Moderes

- M3. Tests FFT avec assertions faibles (`t.Logf` au lieu de `t.Error`)
- M4. Verification de resultats dans tests paralleles potentiellement non atteinte
- M5. Pas de test pour `N = math.MaxUint64`
- M6. Tests de configuration manipulent l'environnement sans `t.Parallel()`
- M7. Mocks gomock generes mais jamais utilises
- M8. E2E tests minimalistes (2 cas seulement)

#### Suggestions

- S9. Golden file pourrait couvrir plus de N
- S10. Limites des fuzz tests a documenter
- S11. Example tests supplementaires possibles
- S12. Absence de benchmark pour le mode comparaison (orchestration)
- S13. `TestStrassenOptionsPrecedence` manipule un etat global

#### Points Positifs

- Fuzz testing impressionnant (4 fonctions)
- Property-based testing exemplaire (4 identites mathematiques : Cassini, recurrence, doublement, GCD)
- Golden file testing avec outil de generation dedie
- Tests d'edge cases complets (N=0,1,2,92,93,94, puissances de 2, context cancelled)
- Concurrence bien testee (50 goroutines)
- Table-driven tests systematiques avec `t.Parallel()`
- Tests CLI e2e avec double strategie (run() + binaire compile)
- Observer pattern bien couvert (nil channel, full channel, nil observer)
- Spy pattern dans l'orchestration
- Tests de copie semantique
