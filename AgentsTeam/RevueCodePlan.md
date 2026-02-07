# Plan d'implementation — Revue de Code FibGo

**Date** : 2026-02-07
**Base** : RevueCode.md (7 critiques, 24 moderes, 21 suggestions)
**Strategie** : 4 phases sequentielles, quick wins d'abord, structurel ensuite

---

## Suivi d'avancement

**Derniere mise a jour** : 2026-02-07
**Session** : Execution complete — 18/18 taches DONE
**Verification finale** : go build OK, go test 18/18 OK, go vet OK

| ID | Tache | Phase | Statut | Notes |
|----|-------|-------|--------|-------|
| 1.1 | Renommer `ProgressReporter` -> `ProgressCallback` | 1 | DONE | 13 fichiers modifies, build+tests OK |
| 1.2 | Corriger assertions faibles `fft_test.go` | 1 | DONE | 3x `t.Logf` -> `t.Errorf` |
| 1.3 | Corriger race condition tests de seuils | 1 | DONE | Group subtest pattern applique |
| 1.4 | Supprimer mocks gomock inutilises | 1 | DONE | 4 fichiers + 4 directives supprimes |
| 1.5 | Corriger commentaire `ProgressState` | 1 | DONE | Commentaire thread-safety corrige |
| 2.1 | Supprimer Karatsuba custom | 2 | DONE | karatsuba.go + tests supprimes, docs MAJ |
| 2.2 | Nettoyer ASM AVX2 mort | 2 | DONE | arith_amd64.s supprime, arith_amd64.go simplifie |
| 2.3 | Deplacer `GetCalculatorsToRun` | 2 | DONE | cli -> orchestration |
| 2.4 | Extraire `internal/format` | 2 | DONE | 3 fichiers format/ crees, TUI+calibration MAJ |
| 3.1 | Evaluer/optimiser cache FFT | 3 | DONE | SHA-256->FNV-1a, flattenPolyData elimine |
| 3.2 | Resoudre couplage config->fibonacci | 3 | DONE | Import fibonacci retire, constantes unifiees |
| 3.3 | Separer `MultiplicationStrategy` (ISP) | 3 | DONE | Multiplier + DoublingStepExecutor + type alias |
| 3.4 | Propager contexte goroutines paralleles | 3 | DONE | 6 goroutines avec ctx.Err() |
| 3.5 | Activer race detector CI | 3 | DONE | .github/workflows/ci.yml cree |
| 3.6 | Enrichir tests E2E | 3 | DONE | 6 nouveaux cas (2->8 total) |
| 4.1 | Optimiser `acquireWordSlice` | 4 | DONE | acquireWordSliceUnsafe ajoute + tests |
| 4.2 | Optimiser `getWordSlicePoolIndex` O(1) | 4 | DONE | 4 fonctions optimisees, 20x plus rapide |
| 4.3 | Documenter invariant semaphores | 4 | DONE | Refs croisees common.go + fft_recursion.go |

---

## Phase 1 — Quick wins (effort faible, impact eleve)

### 1.1 Renommer `fibonacci.ProgressReporter` -> `ProgressCallback`

**Probleme** : Deux types `ProgressReporter` homonymes avec des semantiques differentes (revue archi 1.1).
- `internal/fibonacci/progress.go:18-25` : `type ProgressReporter func(progress float64)` (type fonction)
- `internal/orchestration/interfaces.go:11-30` : `type ProgressReporter interface { DisplayProgress(...) }` (interface)

**Fichiers a modifier** :

| Fichier | Modification |
|---------|-------------|
| `internal/fibonacci/progress.go:25` | Renommer `ProgressReporter` -> `ProgressCallback` |
| `internal/fibonacci/calculator.go` | Mettre a jour toutes les references au type (declaration `var reporter ProgressReporter` etc.) |
| `internal/fibonacci/observer.go` | `AsProgressReporter` -> `AsProgressCallback`, retour `ProgressCallback` |
| `internal/fibonacci/observers.go` | Toute reference au type dans les observers |
| `internal/fibonacci/doubling_framework.go` | Parametres/variables de type `ProgressReporter` -> `ProgressCallback` |
| `internal/fibonacci/mocks/mock_calculator.go` | Regenerer avec `go generate ./internal/fibonacci/...` |
| Fichiers `*_test.go` dans `internal/fibonacci/` | Toute reference au type |

**Verification** :
```bash
go build ./...
go test -short ./internal/fibonacci/ ./internal/orchestration/
```

**Risque** : Faible. Renommage interne au package `fibonacci`. L'interface `orchestration.ProgressReporter` reste inchangee.

---

### 1.2 Corriger les assertions faibles dans `fft_test.go`

**Probleme** : Tests FFT utilisent `t.Logf` au lieu de `t.Errorf` — ne peuvent pas echouer (revue tests M3).

**Fichier** : `internal/fibonacci/fft_test.go`

**Modifications** :
- Ligne 34 : Remplacer `t.Logf("executeDoublingStepFFT returned error (may be expected): %v", err)` par `t.Errorf("executeDoublingStepFFT returned unexpected error: %v", err)`
- Ligne 57 : Idem (mode parallele)
- Ligne 81 : Idem (petits nombres)

Si certaines erreurs sont reellement attendues dans des conditions specifiques, utiliser `t.Skipf` avec un commentaire explicite plutot que `t.Logf`.

**Verification** :
```bash
go test -v -run TestExecuteDoublingStepFFT ./internal/fibonacci/
```

---

### 1.3 Corriger la race condition dans les tests de seuils

**Probleme** : `TestFFTThresholdVariations` et `TestStrassenThresholdVariations` verifient les resultats avant que les sous-tests paralleles ne terminent (revue tests M4).

**Fichier** : `internal/fibonacci/fibonacci_edge_test.go`

**Modification pour `TestFFTThresholdVariations` (lignes 105-143)** :

Deplacer la verification de coherence dans un sous-test `t.Run("Verify", ...)` qui s'execute apres les sous-tests de calcul. En Go, les sous-tests paralleles d'un parent sont garantis de terminer avant que le parent ne continue APRES le dernier `t.Run`. Cependant, le pattern actuel place la verification DANS le meme scope que les `t.Run` paralleles.

**Solution** : Encapsuler les sous-tests dans un `t.Run("group", func(t *testing.T) { ... })` qui attend leur completion avant la verification :

```go
func TestFFTThresholdVariations(t *testing.T) {
    t.Parallel()
    calc := NewCalculator(&OptimizedFastDoubling{})
    ctx := context.Background()
    n := uint64(50000)
    thresholds := []int{0, 1000, 10000, 100000, 1000000}
    var results []*big.Int
    var mu sync.Mutex

    t.Run("calculations", func(t *testing.T) {
        for _, threshold := range thresholds {
            t.Run(fmt.Sprintf("Threshold=%d", threshold), func(t *testing.T) {
                t.Parallel()
                // ... calcul et append dans results ...
            })
        }
    }) // <-- ici, tous les sous-tests paralleles sont termines

    // Verification APRES completion de tous les sous-tests
    if len(results) > 1 {
        for i := 1; i < len(results); i++ {
            if results[0].Cmp(results[i]) != 0 {
                t.Errorf("Results differ...")
            }
        }
    }
}
```

Appliquer le meme pattern a `TestStrassenThresholdVariations` (lignes 145-183).

**Verification** :
```bash
go test -v -run TestFFTThresholdVariations -run TestStrassenThresholdVariations ./internal/fibonacci/
```

---

### 1.4 Supprimer les mocks gomock inutilises

**Probleme** : Mocks gomock generes mais jamais importes dans aucun test (revue tests M7).

**Fichiers a supprimer** :
- `internal/fibonacci/mocks/mock_calculator.go`
- `internal/fibonacci/mocks/mock_generator.go`
- `internal/fibonacci/mocks/mock_strategy.go`
- `internal/cli/mocks/mock_ui.go`

**Pre-verification** : Confirmer qu'aucun fichier de test n'importe ces packages :
```bash
# Verifier qu'aucun import n'existe
grep -r "fibonacci/mocks" --include="*.go" .
grep -r "cli/mocks" --include="*.go" .
```

Si des imports existent, ne supprimer que les mocks non importes.

**Post-verification** : Supprimer aussi les directives `//go:generate mockgen` dans les fichiers source correspondants si les mocks ne sont plus desires, OU conserver les directives et documenter que les mocks sont disponibles mais optionnels.

**Verification** :
```bash
go build ./...
go test -short ./...
```

---

### 1.5 Corriger le commentaire trompeur de `ProgressState`

**Probleme** : Commentaire "designed to be safe for concurrent use" incorrect — `ProgressState` n'est utilise que depuis un seul goroutine (revue fiabilite M3).

**Fichier** : `internal/cli/ui.go:124`

**Modification** : Remplacer le commentaire par une description honnete :
```go
// ProgressState tracks progress for multiple calculators.
// Note: This type is NOT thread-safe. It is designed to be accessed
// from a single goroutine (the select loop in DisplayProgress).
```

---

## Phase 2 — Nettoyage de code mort (effort moyen)

### 2.1 Unexport/supprimer le Karatsuba custom

**Probleme** : `KaratsubaMultiply` est 3-3.5x plus lent que `math/big.Mul`, n'est plus appele par `smartMultiply()`, mais reste exporte (revue perf C1).

**Fichier principal** : `internal/bigfft/karatsuba.go` (348 lignes)

**Exports a supprimer/unexporter** :
- `KaratsubaMultiply(x, y *big.Int) *big.Int` (ligne 76)
- `KaratsubaMultiplyTo(z, x, y *big.Int) *big.Int` (ligne 85)
- `KaratsubaSqr(x *big.Int) *big.Int` (ligne 118)
- `KaratsubaSqrTo(z, x *big.Int) *big.Int` (ligne 125)
- `SetKaratsubaThreshold(threshold int)` (ligne 141)
- `GetKaratsubaThreshold() int` (ligne 150)
- `DefaultKaratsubaThreshold` (ligne 18)
- `DefaultParallelKaratsubaThreshold` (ligne 22)
- `MaxKaratsubaParallelDepth` (ligne 26)

**Pre-verification** : Confirmer que rien n'appelle ces fonctions en dehors des tests et de la documentation :
```bash
grep -r "KaratsubaMultiply\|KaratsubaSqr\|SetKaratsubaThreshold\|GetKaratsubaThreshold" --include="*.go" .
```

**Strategie** :
- **Option A (recommandee)** : Supprimer entierement `karatsuba.go` et `karatsuba_test.go`. Le code est documente comme plus lent a toutes les tailles et n'est plus dans le chemin d'execution. Si besoin futur, l'historique git conserve le code.
- **Option B** : Unexporter toutes les fonctions (minuscule initiale) et ajouter un commentaire `// Deprecated: slower than math/big.Mul at all sizes. Kept for reference only.`

**Fichiers de test a adapter** :
- `internal/bigfft/karatsuba_test.go` : Supprimer (Option A) ou adapter les references (Option B)

**References documentaires a mettre a jour** :
- `Docs/algorithms/BIGFFT.md` (ligne 33)
- `Docs/algorithms/FFT.md`
- `Docs/algorithms/FAST_DOUBLING.md`
- `Docs/ARCHITECTURE.md`
- `Docs/PERFORMANCE.md`
- `CLAUDE.md` (si Karatsuba y est mentionne)

**Verification** :
```bash
go build ./...
go test -short ./internal/bigfft/ ./internal/fibonacci/
```

---

### 2.2 Nettoyer le code assembleur AVX2 mort et l'infrastructure de dispatch

**Probleme** : 226 lignes d'assembleur + infrastructure de dispatch inutilises. `selectAVX2Impl()` est un no-op. `bolt.md` confirme que c'est 2-2.5x plus lent (revue perf M1+M2).

**Fichiers concernes** :

| Fichier | Action |
|---------|--------|
| `internal/bigfft/arith_amd64.s` (225 lignes) | Supprimer entierement |
| `internal/bigfft/arith_amd64.go` (154 lignes) | Simplifier : garder uniquement `AddVV`, `SubVV`, `AddMulVVW` qui delegent directement aux `go:linkname`. Supprimer `UseAVX2()`, `UseDefault()`, `AddVVAuto`, `SubVVAuto`, `AddMulVVWAuto`, les function pointers, les declarations `go:noescape`, et `MinSIMDVectorLen`. |
| `internal/bigfft/cpu_amd64.go` (239 lignes) | Supprimer `selectAVX2Impl()` (ligne 227-229), `selectAVX512Impl()` (ligne 232-239), `selectImplementation()` (ligne 183-195). Conserver la detection CPU (`HasAVX2()`, `HasAVX512()`, etc.) car elle pourrait servir a l'avenir. |
| `internal/bigfft/arith_amd64_test.go` | Supprimer les tests des fonctions AVX2 supprimees : `TestAddVVAvx2`, `TestSubVVAvx2`, `TestAddMulVVWAvx2`, `TestUseAVX2`, `TestUseDefault`, `TestAddVVAuto`, `TestSubVVAuto`, `TestAddMulVVWAuto` et les benchmarks AVX2. Conserver les tests des fonctions `AddVV`, `SubVV`, `AddMulVVW` de base. |

**Post-simplification de `arith_amd64.go`** : Le fichier devrait ressembler a `arith_generic.go` — 3 fonctions publiques qui delegent directement aux `go:linkname` :
```go
//go:build amd64

package bigfft

import "math/big"

func AddVV(z, x, y []big.Word) big.Word  { return addVV(z, x, y) }
func SubVV(z, x, y []big.Word) big.Word  { return subVV(z, x, y) }
func AddMulVVW(z, x []big.Word, y big.Word) big.Word { return addMulVVW(z, x, y) }
```

**Alternative** : Fusionner `arith_amd64.go` et `arith_generic.go` en un seul `arith.go` sans build tags, puisque les deux font la meme chose (delegation aux `go:linkname`).

**Documentation a mettre a jour** :
- `Docs/algorithms/BIGFFT.md` (lignes 556-569)
- `Docs/BUILD.md` (ligne 86)
- `CLAUDE.md` : retirer la mention `arith_amd64.go`, `arith_amd64.s`

**Verification** :
```bash
go build ./...
go test -short ./internal/bigfft/
```

**Risque** : Moyen. La suppression du fichier `.s` necessite de verifier que les declarations `go:noescape` dans `arith_amd64.go` ne creent pas d'erreurs de link. Si les fonctions ASM sont supprimees, leurs declarations Go doivent l'etre aussi.

---

### 2.3 Deplacer `GetCalculatorsToRun` de CLI vers orchestration

**Probleme** : Logique metier (selection des calculateurs) placee dans le package de presentation CLI (revue archi 2.5). Le TUI l'appelle aussi via `cli.GetCalculatorsToRun`.

**Fichier source** : `internal/cli/calculate.go:13-38`

**Fichier destination** : `internal/orchestration/calculator_selection.go` (nouveau)

**Modifications** :

| Fichier | Action |
|---------|--------|
| `internal/orchestration/calculator_selection.go` | Creer avec la fonction `GetCalculatorsToRun` deplacee |
| `internal/cli/calculate.go` | Supprimer la fonction. Si le fichier ne contient que cette fonction, supprimer le fichier. |
| `internal/app/app.go:186` | Changer `cli.GetCalculatorsToRun(...)` -> `orchestration.GetCalculatorsToRun(...)` |
| `internal/app/app.go:199` | Idem |
| `internal/app/app.go` imports | Ajouter `orchestration`, retirer `cli` si plus utilise |
| `internal/cli/calculate_test.go` | Deplacer vers `internal/orchestration/calculator_selection_test.go` |

**Verification** :
```bash
go build ./...
go test -short ./internal/orchestration/ ./internal/app/ ./internal/cli/
```

**Risque** : Faible. Deplacement pur sans changement de logique.

---

### 2.4 Extraire les utilitaires partages CLI/TUI dans `internal/format`

**Probleme** : TUI importe CLI pour `NewProgressWithETA()`, `UpdateWithETA()`, et `FormatExecutionDuration()` — couplage horizontal entre presentations (revue archi 2.1).

**Fichier source** : `internal/tui/bridge.go:10` importe `internal/cli`

**Fonctions a extraire** (utilisees par le TUI depuis le CLI) :
- `cli.NewProgressWithETA(numCalculators int) *ProgressWithETA` (bridge.go:48)
- `cli.(*ProgressWithETA).UpdateWithETA(...)` (bridge.go:51)
- `cli.FormatExecutionDuration(d time.Duration) string` (bridge.go:89, header.go:57, logs.go:111,127,143,146)
- `cli.FormatETA(d time.Duration) string` (chart.go:84, metrics.go:80)
- `cli.FormatNumberString(s string) string` (logs.go:146)

**Note** : Le couplage TUI->CLI est plus large que prevu. 5 fichiers TUI importent CLI :
- `internal/tui/bridge.go` (FormatExecutionDuration, NewProgressWithETA)
- `internal/tui/chart.go` (FormatExecutionDuration, FormatETA)
- `internal/tui/header.go` (FormatExecutionDuration)
- `internal/tui/logs.go` (FormatExecutionDuration, FormatNumberString)
- `internal/tui/metrics.go` (FormatETA)

De plus, 2 fichiers hors TUI importent aussi CLI :
- `internal/calibration/calibration.go`
- `internal/calibration/io.go`

**Fichiers a creer** :
- `internal/format/progress_eta.go` : Deplacer `ProgressWithETA`, `NewProgressWithETA`, `UpdateWithETA`, `FormatETA` depuis `internal/cli/progress_eta.go`
- `internal/format/duration.go` : Deplacer `FormatExecutionDuration` depuis `internal/cli/ui.go:23`
- `internal/format/numbers.go` : Deplacer `FormatNumberString` depuis `internal/cli/ui_format.go:15`

**Modifications** :

| Fichier | Action |
|---------|--------|
| `internal/format/progress_eta.go` | Nouveau : `ProgressWithETA` + `NewProgressWithETA` + `UpdateWithETA` + `FormatETA` |
| `internal/format/duration.go` | Nouveau : `FormatExecutionDuration` |
| `internal/format/numbers.go` | Nouveau : `FormatNumberString` |
| `internal/cli/progress_eta.go` | Supprimer les fonctions deplacees, ou faire deleguer vers `format` |
| `internal/cli/ui.go` | Supprimer `FormatExecutionDuration` ou faire deleguer |
| `internal/cli/ui_format.go` | Supprimer `FormatNumberString` ou faire deleguer |
| `internal/tui/bridge.go` | Remplacer `import cli` par `import format` |
| `internal/tui/chart.go` | Idem |
| `internal/tui/header.go` | Idem |
| `internal/tui/logs.go` | Idem |
| `internal/tui/metrics.go` | Idem |
| `internal/calibration/calibration.go` | Adapter si les fonctions CLI utilisees sont deplacees |
| `internal/calibration/io.go` | Idem |
| `internal/cli/` fichiers restants | Importer `format` au lieu de references locales |

**Verification** :
```bash
go build ./...
go test -short ./internal/cli/ ./internal/tui/ ./internal/format/
```

**Risque** : Moyen. Necessite de tracer toutes les references a ces fonctions dans le package CLI pour ne rien casser.

---

## Phase 3 — Ameliorations structurelles (effort eleve)

### 3.1 Evaluer et optimiser le cache FFT

**Probleme** : SHA-256 pour les cles de cache + deep copy a chaque Get/Put annulent potentiellement le benefice du cache (revue perf C2+C3).

**Fichier** : `internal/bigfft/fft_cache.go`

**Etape 1 — Profiler** :
```bash
# Benchmark avec cache active (defaut)
go test -bench=BenchmarkFibonacci -benchmem ./internal/fibonacci/ -run=^$
# Benchmark avec cache desactive
# (modifier temporairement DefaultTransformCacheConfig().Enabled = false)
go test -bench=BenchmarkFibonacci -benchmem ./internal/fibonacci/ -run=^$
```

Comparer les resultats. Si le cache n'apporte pas de gain mesurable, le desactiver par defaut.

**Etape 2 — Si le cache est benefique, optimiser** :

2a. Remplacer SHA-256 par un hash rapide (lignes 99-120) :
```go
import "hash/fnv"

func computeKey(data nat, k uint, n int) uint64 {
    h := fnv.New64a()
    buf := make([]byte, 8)
    binary.LittleEndian.PutUint64(buf, uint64(k))
    h.Write(buf)
    binary.LittleEndian.PutUint64(buf, uint64(n))
    h.Write(buf)
    for _, word := range data {
        binary.LittleEndian.PutUint64(buf, uint64(word))
        h.Write(buf)
    }
    return h.Sum64()
}
```

Adapter la structure `cacheEntry` pour utiliser `uint64` au lieu de `[32]byte` comme cle.

2b. Eliminer `flattenPolyData()` (lignes 308-323) : Hasher directement les coefficients du `Poly` sans copie intermediaire.

2c. Evaluer une approche COW (copy-on-write) pour eviter les deep copies dans `Get()` et `Put()`. Alternative : utiliser un compteur de references et ne copier que si le buffer est partage.

**Verification** :
```bash
go test -bench=. -benchmem ./internal/bigfft/
go test -short ./internal/bigfft/ ./internal/fibonacci/
```

**Risque** : Eleve. Changement de la semantique du cache. Les benchmarks avant/apres sont obligatoires.

---

### 3.2 Resoudre le couplage `config` -> `fibonacci`

**Probleme** : `config.go` importe `fibonacci` pour `Options` et `ToCalculationOptions()`. Constantes dupliquees entre les deux packages (revue archi 1.2).

**Fichier principal** : `internal/config/config.go:12,83-91`

**Strategie** : Inverser la dependance. La config ne devrait pas connaitre `fibonacci.Options`.

**Option A — Fonction dans fibonacci** :
1. Supprimer `ToCalculationOptions()` de `config.go`
2. Creer `internal/fibonacci/config_adapter.go` avec :
   ```go
   func OptionsFromConfig(parallel, fft, strassen int) Options {
       return Options{
           ParallelThreshold: parallel,
           FFTThreshold:      fft,
           StrassenThreshold: strassen,
       }
   }
   ```
3. Dans `orchestrator.go:67`, remplacer `cfg.ToCalculationOptions()` par `fibonacci.OptionsFromConfig(cfg.Threshold, cfg.FFTThreshold, cfg.StrassenThreshold)`

**Option B — Package partage** :
1. Creer `internal/types/options.go` avec le type `CalculationOptions`
2. Faire dependre `config` et `fibonacci` de `types`
3. Plus propre mais ajoute un package

**Constantes** : Supprimer les constantes dupliquees dans `config.go` (lignes 31-37) et utiliser `fibonacci.DefaultParallelThreshold`, `fibonacci.DefaultFFTThreshold`, `fibonacci.DefaultStrassenThreshold` dans le parsing de config via l'app layer.

**Fichiers a modifier** :

| Fichier | Action |
|---------|--------|
| `internal/config/config.go` | Supprimer import `fibonacci`, supprimer `ToCalculationOptions()`, supprimer constantes dupliquees |
| `internal/config/config.go` | Utiliser des valeurs par defaut locales (0) et laisser `fibonacci.normalizeOptions` appliquer les defauts |
| `internal/app/app.go` | Construire `fibonacci.Options` directement depuis `cfg` au lieu d'appeler `cfg.ToCalculationOptions()` |
| `internal/orchestration/orchestrator.go:67` | Adapter l'appel |
| `internal/config/config_extra_test.go` | Adapter les tests |

**Verification** :
```bash
go build ./...
go test -short ./internal/config/ ./internal/app/ ./internal/orchestration/ ./internal/fibonacci/
```

**Risque** : Moyen. Changement d'API interne affectant config, app, et orchestration.

---

### 3.3 Separer `MultiplicationStrategy` en deux interfaces

**Probleme** : `ExecuteStep(*CalculationState)` dans `MultiplicationStrategy` couple toutes les strategies au fast doubling (revue archi 2.2, ISP).

**Fichier** : `internal/fibonacci/strategy.go:28-75`

**Modification** :
```go
// Multiplier defines pure multiplication/squaring operations.
type Multiplier interface {
    Multiply(z, x, y *big.Int, opts Options) (*big.Int, error)
    Square(z, x *big.Int, opts Options) (*big.Int, error)
    Name() string
}

// DoublingStepExecutor extends Multiplier with doubling-step-aware execution.
type DoublingStepExecutor interface {
    Multiplier
    ExecuteStep(ctx context.Context, s *CalculationState, opts Options, inParallel bool) error
}
```

**Fichiers a modifier** :

| Fichier | Action |
|---------|--------|
| `internal/fibonacci/strategy.go` | Separer l'interface, adapter `AdaptiveStrategy`, `FFTOnlyStrategy`, `KaratsubaStrategy` |
| `internal/fibonacci/doubling_framework.go` | Utiliser `DoublingStepExecutor` la ou `ExecuteStep` est appele |
| `internal/fibonacci/calculator.go` | Adapter les references a `MultiplicationStrategy` |
| `internal/fibonacci/fastdoubling.go` | Adapter |
| `internal/fibonacci/mocks/mock_strategy.go` | Regenerer ou supprimer (cf. 1.4) |
| Tests dans `internal/fibonacci/` | Adapter les assertions de type |

**Verification** :
```bash
go build ./...
go test -short ./internal/fibonacci/
```

**Risque** : Eleve. Changement d'interface qui impacte toutes les strategies et leurs consommateurs.

---

### 3.4 Propager le contexte dans les goroutines de multiplication parallele

**Probleme** : Les goroutines de multiplication parallele ne verifient pas le contexte — l'app peut etre lente a repondre a SIGINT sur de grands N (revue fiabilite M1+M2).

**Fichiers** :
- `internal/fibonacci/fft.go:115-174` (executeDoublingStepFFT, chemin parallele)
- `internal/fibonacci/doubling_framework.go:65-108` (executeDoublingStepMultiplications, chemin parallele)

**Modification pour `fft.go`** :
Ajouter une verification `ctx.Err()` avant chaque operation FFT lourde dans les goroutines :
```go
go func() {
    if err := ctx.Err(); err != nil {
        resChan <- result{nil, fmt.Errorf("canceled before FFT mul: %w", err)}
        return
    }
    v, err := fkPoly.Mul(&t4Poly)
    // ...
}()
```

**Note** : Le contexte (`ctx`) est deja un parametre de `executeDoublingStepFFT` (verifie dans l'exploration). Il suffit de l'utiliser dans les goroutines.

**Modification pour `doubling_framework.go`** :
Meme pattern — verifier `ctx.Err()` au debut de chaque goroutine :
```go
go func() {
    defer wg.Done()
    if err := ctx.Err(); err != nil {
        ec.SetError(fmt.Errorf("canceled before multiply: %w", err))
        return
    }
    // ... operation existante ...
}()
```

**Limitation** : Ces verifications ne permettent que de detecter une annulation AVANT le debut de l'operation. Pour interrompre une operation FFT en cours, il faudrait propager le contexte plus profondement dans `bigfft.MulTo`/`bigfft.SqrTo`, ce qui est un changement beaucoup plus invasif (hors scope).

**Verification** :
```bash
go test -v -run TestContextCancellation ./internal/fibonacci/
go test -short ./internal/fibonacci/
```

---

### 3.5 Activer le race detector dans le CI

**Probleme** : `go test -race` echoue sur Windows sans CGO. Les tests de concurrence ne prouvent rien sans le race detector (revue tests C2).

**Options** :

| Option | Effort | Fiabilite |
|--------|--------|-----------|
| A. Activer CGO sur Windows CI (installer GCC via chocolatey/msys2) | Moyen | Elevee |
| B. Ajouter un job CI Linux avec `-race` | Faible | Elevee |
| C. Utiliser `-race` uniquement en local avec documentation | Tres faible | Faible |

**Recommandation** : Option B. Ajouter un job Linux dans la CI (GitHub Actions) :
```yaml
  test-race:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: go test -race -short ./...
```

**Fichier a modifier** : `.github/workflows/ci.yml` (ou equivalent)

---

### 3.6 Enrichir les tests E2E

**Probleme** : Seulement 2 cas de test E2E (revue tests M8).

**Fichier** : `test/e2e/cli_e2e_test.go`

**Cas de test a ajouter** :

```go
// Ajouter dans la table de tests existante :
{
    name:     "All Algorithms Comparison",
    args:     []string{"-n", "100", "--algo", "all", "-c"},
    wantOut:  "F(100)",
    wantCode: 0,
},
{
    name:     "Quiet Mode",
    args:     []string{"-n", "10", "--quiet", "-c"},
    wantOut:  "55",
    wantCode: 0,
},
{
    name:     "Very Short Timeout",
    args:     []string{"-n", "10000000", "--timeout", "1ms"},
    wantOut:  "",        // timeout avant resultat
    wantCode: 2,         // exit code timeout
},
{
    name:     "Invalid N Zero",
    args:     []string{"-n", "0", "-c"},
    wantOut:  "F(0) = 0",
    wantCode: 0,
},
{
    name:     "Large N",
    args:     []string{"-n", "1000", "-c"},
    wantOut:  "F(1000)",
    wantCode: 0,
},
{
    name:     "Unknown Algorithm",
    args:     []string{"-n", "10", "--algo", "nonexistent"},
    wantOut:  "",
    wantCode: 1,         // exit code erreur config
},
{
    name:     "Env Variable Override",
    args:     []string{"-c"},
    env:      map[string]string{"FIBCALC_N": "10"},
    wantOut:  "F(10) = 55",
    wantCode: 0,
},
{
    name:     "Version Flag",
    args:     []string{"--version"},
    wantOut:  "fibcalc",
    wantCode: 0,
},
```

**Note** : Adapter la structure de test pour supporter le champ `env` et la verification du code de sortie via `exec.ExitError`.

**Verification** :
```bash
go test -v ./test/e2e/
```

---

## Phase 4 — Optimisations performance (effort variable)

### 4.1 Optimiser `acquireWordSlice` : eviter le `clear()` inutile

**Probleme** : `clear(slice)` dans `pool.go:59` nettoie toute la slice meme quand l'appelant va immediatement ecrire dedans (revue perf M4).

**Fichier** : `internal/bigfft/pool.go:54-64`

**Modification** : Ajouter une variante sans clear :
```go
func acquireWordSliceUnsafe(size int) []big.Word {
    idx := getWordSlicePoolIndex(size)
    if idx < 0 {
        return make([]big.Word, size)
    }
    slice := wordSlicePools[idx].Get().([]big.Word)
    return slice[:size]
}
```

Modifier les callsites qui ecrivent immediatement (ex: `copy` dans `PoolAllocator.AllocFermatSlice`) pour utiliser `acquireWordSliceUnsafe`. Garder `acquireWordSlice` avec `clear` pour les callsites qui lisent avant d'ecrire.

**Verification** :
```bash
go test -short ./internal/bigfft/
go test -bench=. -benchmem ./internal/bigfft/
```

---

### 4.2 Optimiser `getWordSlicePoolIndex` : O(1) bitwise

**Probleme** : Recherche lineaire sur 10 elements au lieu d'un calcul O(1) (revue perf M5).

**Fichier** : `internal/bigfft/pool.go:35-42`

**Modification** :
```go
func getWordSlicePoolIndex(size int) int {
    if size <= 0 || size > wordSliceSizes[len(wordSliceSizes)-1] {
        return -1
    }
    // Tailles sont des puissances de 4 a partir de 64 (= 4^3)
    // bits.Len(63) = 6, bits.Len(64) = 7
    // index = (bits.Len(uint(size-1)) - 6) / 2
    // mais verifier les edge cases avec la table existante
    idx := (bits.Len(uint(size-1)) - 5) / 2
    if idx < 0 {
        idx = 0
    }
    if idx >= len(wordSliceSizes) {
        return -1
    }
    return idx
}
```

**Important** : Verifier par des tests unitaires que le mapping est identique a la recherche lineaire pour TOUTES les tailles. Garder la table `wordSliceSizes` comme reference pour les tests.

Appliquer le meme pattern a `getFermatPoolIndex`, `getNatSlicePoolIndex`, `getFermatSlicePoolIndex`.

---

### 4.3 Documenter l'invariant des semaphores multiples

**Probleme** : Triple semaphore independant sans documentation de l'interaction (revue perf M3).

**Fichiers** :
- `internal/fibonacci/common.go:26` (NumCPU*2)
- `internal/bigfft/fft_recursion.go:17` (NumCPU)

**Modification** : Ajouter un commentaire documentant l'invariant :
```go
// getTaskSemaphore returns a semaphore limiting Fibonacci-level parallelism
// to NumCPU*2 goroutines. This is independent from the FFT-level semaphore
// (bigfft/fft_recursion.go, NumCPU goroutines). When both are active, up to
// NumCPU*3 goroutines may be active simultaneously. This is mitigated by
// ShouldParallelizeMultiplication() which disables Fibonacci-level parallelism
// when FFT is active (except for operands > ParallelFFTThreshold = 10M bits).
```

---

## Dependances entre taches

```
Phase 1 (parallele) :
  1.1 (ProgressReporter rename)  -- independant
  1.2 (fft_test assertions)      -- independant
  1.3 (edge test race fix)       -- independant
  1.4 (supprimer mocks)          -- independant
  1.5 (commentaire ProgressState)-- independant

Phase 2 (sequentiel partiel) :
  2.1 (Karatsuba)   -- independant
  2.2 (ASM AVX2)    -- independant, mais tester apres 2.1 (meme package bigfft)
  2.3 (GetCalcsToRun) -- independant
  2.4 (CLI/TUI extract) -- independant de 2.3 mais tester ensemble

Phase 3 (sequentiel) :
  3.1 (cache FFT)    -- apres 2.1+2.2 (meme package)
  3.2 (config coupling) -- independant
  3.3 (ISP strategy) -- apres 1.1 (meme package fibonacci)
  3.4 (context propagation) -- apres 3.3 (meme fichiers)
  3.5 (CI race)      -- independant
  3.6 (E2E tests)    -- apres 2.3 (GetCalcsToRun deplace)

Phase 4 (parallele) :
  4.1 (clear optim)   -- apres 2.1+2.2
  4.2 (pool index O(1)) -- apres 4.1 (meme fichier)
  4.3 (doc semaphores) -- independant
```

## Verification finale

Apres toutes les phases :
```bash
go build ./...
go test -v -short ./...
go test -cover ./...         # verifier >= 75% partout
go vet ./...
golangci-lint run ./...      # si disponible
```

## Resume

| Phase | Taches | Effort estime | Fichiers impactes |
|-------|--------|---------------|-------------------|
| 1 | 5 quick wins | Faible | ~10 fichiers |
| 2 | 4 nettoyages | Moyen | ~15 fichiers + suppressions |
| 3 | 6 ameliorations | Eleve | ~20 fichiers |
| 4 | 3 optimisations | Variable | ~5 fichiers |
| **Total** | **18 taches** | | **~50 fichiers** |
