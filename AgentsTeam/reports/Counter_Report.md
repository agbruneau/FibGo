# Contre-rapport : Revue adversariale

## Synthese des objections

| Proposition | Objections critiques | Objections mineures | Verdict preliminaire |
|-------------|---------------------|---------------------|---------------------|
| NTT multi-primes | 3 (overhead CRT, perte squaring optimise, Go pur impraticable) | 2 (complexite maintenance, regression petits N) | **Douteux** -- viable uniquement avec SIMD pour F(10^9)+, ROI incertain |
| Montgomery | 2 (inadaptation fondamentale, overhead systematique) | 1 (complexite inutile) | **Rejete** -- accord avec le chercheur, mais objections additionnelles |
| SS ameliore | 2 (gains surestimates, squaring basicSqr inefficace) | 3 (twiddle cache risque, seuils non valides, sqrt(2) trick hors portee) | **Viable** -- mais gains realistes de 8-15% plutot que 15-25% |

---

## Objections detaillees

---

### NTT multi-primes

#### Objection 1 : L'overhead CRT annule l'avantage SIMD pour la plage cible

- **Categorie** : Performance
- **Severite** : Bloquante
- **Argument** : Le rapport estime un crossover NTT vs Fermat FFT a F(10^8)-F(10^9) avec SIMD. Cependant, cette estimation sous-evalue l'overhead CRT. Avec K=3 primes 63-bit, la reconstruction CRT necessite de l'arithmetique 192-bit (3 * 64 bits) pour chaque coefficient. En Go, `math/bits.Mul64` + `math/bits.Add64` produisent du code scalaire sans vectorisation. Pour N coefficients du polynome FFT, c'est O(N) multiplications 128-bit, chacune coutant ~10-15 cycles sur x86-64 (vs 1 cycle pour un shift Fermat dans `fermat.Shift` a `fermat.go:48-101`). Pour F(10^8) avec K_FFT ~4096 coefficients et 3 primes, le CRT ajoute ~4096 * 3 * 15 = ~184K cycles **par iteration de doubling**. Sur les ~27 iterations du Fast Doubling pour F(10^8) (log2(10^8) ~ 27), cela fait ~5M cycles supplementaires -- soit ~1.5ms a 3GHz. C'est faible en absolu mais representatif d'un overhead constant qui erode le gain SIMD (estime a 2-4x sur les butterfly, soit ~50-75% de gain). Le rapport NTT estime un ratio NTT/Fermat de 1.25x plus lent pour F(10^8) -- cela semble **optimiste** car il ne comptabilise que la multiplication pointwise et ignore la reconstruction CRT dans les estimations detaillees.
- **Fichiers concernes** : `internal/bigfft/fermat.go:48-101` (Shift gratuit actuel), `internal/bigfft/fft_poly.go:331-355` (multiplication pointwise actuelle)
- **Critere de resolution** : Presenter un prototype benchmark Go montrant le cout CRT reel par coefficient pour 3 primes 63-bit, compare au cout d'un `fermat.Mul` pour des coefficients de taille equivalente. Si le CRT coute < 2x le shift Fermat, l'objection est caduque.

#### Objection 2 : Perte du squaring optimise -- l'avantage 33% de la FFT Fermat disparait

- **Categorie** : Performance
- **Severite** : Majeure
- **Argument** : L'implementation actuelle exploite un avantage fondamental du squaring FFT : une seule transformee au lieu de deux (`fftsqrTo` a `fft_core.go:87-104`). Le Fast Doubling effectue 2 squarings + 1 multiplication par iteration (F(k)^2, F(k+1)^2, F(k)*(2F(k+1)-F(k))). Avec la FFT Fermat, le squaring economise 33% du cout de transform. Avec la NTT multi-primes, cette economie existe aussi MAIS elle est multipliee par K primes : au lieu de 1 transformee economisee, on economise K transformees -- sauf que le cout de base est K fois plus eleve. L'avantage relatif reste le meme (33%), mais le cout absolu des K transforms est plus lourd. Plus important : le rapport NTT ne mentionne pas comment le squaring interagit avec le cache de transformees (`fft_cache.go:305-331`). L'implementation actuelle cache les transformees FFT Fermat et les reutilise pour le squaring -- un seul cache suffit. Avec K primes NTT, il faudrait K caches independants ou un cache K fois plus grand, augmentant la pression memoire.
- **Fichiers concernes** : `internal/bigfft/fft_core.go:87-104` (fftsqrTo), `internal/bigfft/fft_cache.go:411-448` (SqrCachedWithBump), `internal/fibonacci/fft.go:81-239` (executeDoublingStepFFT avec transform reuse)
- **Critere de resolution** : Demontrer que le schema de caching NTT multi-primes peut etre aussi efficace que le cache Fermat actuel (128 entrees, 15-30% speedup) sans depasser 2x la consommation memoire cache.

#### Objection 3 : Implementation Go pur = regression garantie, assembly = hors perimetre raisonnable

- **Categorie** : Risque / Maintenabilite
- **Severite** : Majeure
- **Argument** : Le rapport NTT admet lui-meme que "sans SIMD, la NTT serait ~2-3x plus lente" et que la "NTT Go pur" a un crossover a ~F(5*10^9) -- soit "probablement jamais rentable". Cela signifie que le Phase 1 du plan (prototype Go pur, 2 semaines) aboutirait a un code **garanti plus lent** que l'existant. Pour que la NTT devienne competitive, il faut l'assembly SIMD (Phase 3, 2-3 semaines supplementaires). Or, le projet utilise deja `go:linkname` vers `math/big` pour les operations vectorielles (`arith_decl.go:34-65`), ce qui est fragile (avertissement explicite lignes 1-17). Ajouter un second lot de code assembly custom (`ntt_amd64.s`) doublerait cette surface de fragilite. De plus, l'assembly NTT (butterfly, mulmod modular) est significativement plus complexe que les simples `addVV`/`subVV` deja linkes, car il necessite des reductions modulaires SIMD (Barrett ou Montgomery reduction vectorisee).
- **Fichiers concernes** : `internal/bigfft/arith_decl.go:1-17` (warning go:linkname), `internal/bigfft/arith_amd64.go` (wrappers actuels), `internal/bigfft/cpu_amd64.go:64-85` (detection CPU existante)
- **Critere de resolution** : (1) Confirmer que le prototype Go pur Phase 1 sera benchmarke contre Fermat FFT avec un critere d'abandon clair (> 5x plus lent = stop). (2) Pour la Phase 3 assembly, fournir une estimation de lignes d'assembleur necessaires et un plan de maintenance (qui maintient le code quand Go change de version ?).

#### Objection 4 (mineure) : Complexite de maintenance -- deux backends FFT

- **Categorie** : Maintenabilite
- **Severite** : Mineure
- **Argument** : Le rapport propose d'ajouter la NTT "comme option" sans remplacer la Fermat FFT. Cela implique de maintenir deux backends de multiplication FFT en parallele (6+ nouveaux fichiers proposes). Le pattern `TempAllocator` (`allocator.go:10-35`) facilite l'abstraction des allocations mais ne couvre pas la divergence algorithmique (Fermat ring vs prime fields). Chaque optimisation future (ex: les optimisations SS du rapport ss-researcher) devrait etre implementee dans les deux backends.
- **Fichiers concernes** : `internal/bigfft/allocator.go:10-35` (interface TempAllocator), `internal/bigfft/fft.go:40-52` (dispatch actuel par seuil)
- **Critere de resolution** : Definir une interface commune `FFTBackend` avec des methodes `ForwardTransform`, `PointwiseMul`, `InverseTransform`, `PointwiseSqr` qui abstrait les deux implementations. Demontrer que cette interface ne penalise pas les performances (pas d'indirection sur le hot path).

#### Objection 5 (mineure) : Les estimations de crossover dependent de y-cruncher, pas de FibGo

- **Categorie** : Performance
- **Severite** : Mineure
- **Argument** : Le rapport utilise les benchmarks y-cruncher comme reference pour estimer les ratios NTT/Fermat. Mais y-cruncher est ecrit en C++ avec du SIMD intensif et des optimisations memoire avancees. Extrapoler ces ratios a une implementation Go est hasardeux. En particulier, Go a un overhead GC qui n'existe pas en C++, et les goroutines ont un overhead de scheduling qui n'existe pas avec les threads natifs utilises par y-cruncher. Le ratio "NTT Go" / "Fermat FFT Go" pourrait etre significativement different du ratio "NTT C++" / "Fermat FFT C++".
- **Fichiers concernes** : N/A (estimation theorique)
- **Critere de resolution** : Les estimations de performance devraient etre etiquetees comme "extrapolees de y-cruncher" et pas presentees comme des predictions fiables. Un prototype Go est le seul moyen de valider.

---

### Montgomery

#### Objection 1 : Accord sur le rejet, mais l'analyse du "gap Karatsuba-FFT" est incomplete

- **Categorie** : Performance
- **Severite** : Majeure
- **Argument** : Le rapport Montgomery identifie correctement un "gap" entre Karatsuba (~10K bits) et FFT (~500K bits via `DefaultFFTThreshold` a `constants.go:28`). Il propose NTT et Toom-Cook comme alternatives mais ne quantifie pas le cout de ce gap. En realite, ce gap est deja couvert par `math/big` qui utilise Toom-Cook 3-way et Toom-Cook 4-way en interne (seuils non controles par FibGo). La vraie question est : **le seuil FFT de 500K bits est-il optimal ?** Le rapport ne tente pas d'abaisser `DefaultFFTThreshold` et de mesurer l'impact. Or, `fftThreshold` dans `fft.go:35` est a 1800 mots (~115K bits) au niveau bigfft, tandis que `DefaultFFTThreshold` dans `constants.go:28` est a 500K bits au niveau fibonacci. Il y a donc deja un ecart : bigfft considere que FFT est efficace a partir de ~115K bits, mais fibonacci n'y bascule qu'a 500K bits. Cet ecart de ~385K bits est-il justifie par les benchmarks ou est-ce un haut seuil conservateur ?
- **Fichiers concernes** : `internal/bigfft/fft.go:31-35` (fftThreshold = 1800 mots), `internal/fibonacci/constants.go:22-28` (DefaultFFTThreshold = 500,000), `internal/fibonacci/fft.go:45-59` (smartMultiply dispatch)
- **Critere de resolution** : Benchmarker F(N) pour N = 500K, 1M, 2M, 5M avec des seuils FFT de 100K, 200K, 300K, 400K, 500K bits. Si un seuil plus bas est systematiquement plus rapide, la recommandation "optimiser les seuils existants" devient la priorite #1 avant tout changement algorithmique.

#### Objection 2 : Le rapport sous-estime l'utilite de Montgomery pour un mode Fibonacci modulaire futur

- **Categorie** : Integration
- **Severite** : Mineure
- **Argument** : Le rapport mentionne en passant que Montgomery serait utile pour `F(n) mod M`, puis ecarte ce scenario. Cependant, le calcul de Fibonacci modulaire est un cas d'usage courant en cryptographie et en competitive programming (F(n) mod 10^9+7 par exemple). Si FibGo souhaitait un jour ajouter ce mode, Montgomery serait la technique optimale. Le rapport devrait explicitement recommander de **preserver la possibilite architecturale** d'ajouter Montgomery plus tard via l'interface `Multiplier` existante (`strategy.go:29-59`), meme si l'implementation immediate est rejetee.
- **Fichiers concernes** : `internal/fibonacci/strategy.go:29-59` (interface Multiplier)
- **Critere de resolution** : Ajouter une note dans la conclusion du rapport reconnaissant explicitement que le Strategy pattern actuel permet l'ajout futur de Montgomery sans modification architecturale. Aucune action immediate requise.

#### Objection 3 (mineure) : Les estimations de overhead Montgomery (25-30%) sont imprecises

- **Categorie** : Performance
- **Severite** : Mineure
- **Argument** : Le rapport estime un ralentissement de ~25% avec Montgomery. Mais cette estimation est presentee comme uniforme pour toutes les tailles (F(10^3) a F(10^10)), ce qui est improbable. Pour les tres petites tailles (< 1K bits), le REDC overhead est proportionnellement plus eleve (setup cost domine). Pour les tres grandes tailles, le REDC est O(n) vs M(n) = O(n log n), donc l'overhead relatif diminue asymptotiquement vers 0%. L'estimation devrait etre variable : ~50% overhead pour petits N, ~10% pour tres grands N.
- **Fichiers concernes** : N/A (estimation theorique)
- **Critere de resolution** : Affiner les estimations avec une analyse asymptotique montrant la decroissance du ratio overhead REDC / cout multiplication total en fonction de la taille.

---

### SS ameliore (Schonhage-Strassen optimise)

#### Objection 1 : Le gain cumule de 15-25% est surestime -- estimation realiste : 8-15%

- **Categorie** : Performance
- **Severite** : Majeure
- **Argument** : Le rapport estime des gains additifs pour les optimisations 3.1-3.5. Cependant, les gains ne sont **pas additifs** (le rapport le note mais ne corrige pas les estimations). Plus specifiquement :

  (a) **Twiddle factor pre-calcul (3.1, "5-10%")** : Le rapport affirme que la boucle Reconstruct represente ~40% du temps FFT. Verifions : dans `fourierRecursiveUnified` (`fft_recursion.go:126-131`), la boucle fait 3 operations par element : `ShiftHalf`, `Sub`, `Add`. Le `ShiftHalf` (`fermat.go:106-118`) fait 2 `Shift` + 1 `Sub` (pour k impair) ou 1 `Shift` (pour k pair). Le pre-calcul des twiddle factors economiserait le re-calcul du shift mais PAS l'application du shift (qui est l'operation couteuse). Un tableau pre-calcule de `fermat` shifts necessite K elements de taille n+1 mots, soit potentiellement des dizaines de MB pour les grandes FFT. Le gain reel est probablement **3-5%** et non 5-10%.

  (b) **Squaring specialise (3.2, "5-8%")** : Le rapport propose `basicSqr` pour n < 30 mots (`smallMulThreshold` a `fermat.go:11`). Pour ces tailles, `basicMul` fait n^2 operations. `basicSqr` avec symetrie fait n*(n+1)/2 operations, soit ~50% d'economie sur les produits partiels. Cependant, `basicMul` n'est utilise que pour les **tres petits** coefficients (< 30 mots = < 1920 bits). Pour les grandes FFT ou les coefficients sont > 30 mots, `fermat.Mul` utilise `big.Int.Mul` (`fermat.go:158-173`) qui optimise DEJA le squaring en interne (Go 1.20+ detecte x==y dans `math/big`). Le gain reel ne concerne donc que les petites FFT (K petit, coefficients < 30 mots), soit un gain **conditionnel de 3-5%** sur les petites tailles et **~0%** sur les grandes tailles.

  (c) **Seuils parallelisme (3.3, "3-7%")** : Augmenter `MaxParallelFFTDepth` de 3 a log2(NumCPU) est risque. Avec 16 coeurs, depth=4 cree 16 branches -- mais chaque branche de la FFT recursive traite des donnees de taille K/16 qui pourraient deja etre en L1 cache. La parallelisation de ces sous-problemes en cache cree de la contention memoire (cache line bouncing) et de l'overhead goroutine (~300 bytes stack + scheduling). Le gain n'est pas garanti et pourrait meme etre negatif. Estimation realiste : **1-3%** sur >16 coeurs, **0% ou negatif** sur 4-8 coeurs.

  Gain realiste cumule : 3-5% + 3-5% + 1-3% + 1-3% (cache Reconstruct) + 0-2% (cache agrandissement) = **8-15%** et non 15-25%.

- **Fichiers concernes** : `internal/bigfft/fft_recursion.go:126-131` (boucle Reconstruct), `internal/bigfft/fermat.go:11` (smallMulThreshold=30), `internal/bigfft/fermat.go:149-173` (Mul dispatch), `internal/bigfft/fft_recursion.go:28-33` (seuils parallelisme)
- **Critere de resolution** : Pour chaque optimisation proposee, presenter un micro-benchmark isole (pas une estimation) montrant le gain sur le chemin critique concerne. Par exemple : benchmark de `fermat.Mul(x,x)` vs `fermat.Sqr(x)` pour n=10, 20, 29, 50, 100 mots. Si le gain mesure pour une optimisation est < 2%, la deprioritiser.

#### Objection 2 : basicSqr pour n < 30 mots est inutile quand big.Int detecte deja le squaring

- **Categorie** : Performance
- **Severite** : Majeure
- **Argument** : Le rapport propose d'ajouter `fermat.Sqr()` avec `basicSqr()` specialise. Examinons le chemin critique. Dans `fft_poly.go:384-389`, `sqr()` appelle `buf.Mul(p.Values[i], p.Values[i])`. Cela execute `fermat.Mul(x, y)` avec x == y (meme slice). Pour n >= `smallMulThreshold` (30), le code a `fermat.go:158-173` cree deux `big.Int` (`xi`, `yi`) pointant vers les memes donnees puis appelle `zi.Mul(&xi, &yi)`. Go's `math/big.Int.Mul` detecte que `x == y` (memes pointeurs internes) et utilise son algorithme de squaring optimise en interne (Karatsuba squaring). Donc pour n >= 30, **le squaring est deja optimise**. Le gain ne concerne que `basicMul` pour n < 30 (utilise dans la multiplication pointwise pour les petites FFT). Or, pour les grandes valeurs de N (la cible principale), les coefficients FFT font bien plus de 30 mots. Le gain est donc **negligeable pour la plage cible F(10^6)+**.

  Cependant, il y a un subtilite : dans `fermat.Mul`, `xi.SetBits(x)` et `yi.SetBits(y)` creent des `big.Int` qui **partagent** les memes mots sous-jacents. Quand on appelle `zi.Mul(&xi, &yi)`, est-ce que `math/big` detecte que `x.Bits()` et `y.Bits()` pointent vers la meme memoire ? Pas necessairement -- il compare les pointeurs `*big.Int` (`&xi != &yi`), pas les slices internes. Il faudrait verifier si Go detecte le squaring via les pointeurs de slice ou les pointeurs de `big.Int`. Si c'est par `big.Int`, alors le squaring n'est PAS detecte et le gain de `basicSqr` serait reel meme pour n >= 30.

- **Fichiers concernes** : `internal/bigfft/fermat.go:149-173` (Mul avec big.Int), `internal/bigfft/fft_poly.go:384-389` (sqr pointwise)
- **Critere de resolution** : Ecrire un benchmark simple : `big.Int.Mul(x, y)` ou x et y sont des `big.Int` distincts mais avec `SetBits` pointant vers la meme slice. Comparer avec `big.Int.Mul(x, x)` (meme pointeur). Si les temps sont identiques, le squaring est bien detecte par les slices internes et `basicSqr` est inutile pour n >= 30. Si les temps different, `fermat.Sqr` est justifie pour TOUTES les tailles.

#### Objection 3 (mineure) : Le twiddle factor cache cree un risque de coherence

- **Categorie** : Risque
- **Severite** : Mineure
- **Argument** : Pre-calculer les twiddle factors implique de les stocker dans un cache indexe par (size, backward, n). La FFT recursive descend dans des sous-problemes de taille decroissante (size-1 a chaque niveau). Les twiddle factors dependent du `omega2shift = (4*n*_W) >> size` (`fft_recursion.go:51`), qui change a chaque niveau de recursion. Un cache valide au niveau k n'est pas valide au niveau k-1. Le cache devrait donc etre indexe par niveau de recursion, pas globalement. Cela complique considerablement l'implementation et reduit le gain (le cache est "cold" au premier appel de chaque niveau).
- **Fichiers concernes** : `internal/bigfft/fft_recursion.go:50-52` (calcul omega2shift dependant de size)
- **Critere de resolution** : Preciser l'architecture du cache de twiddle factors : per-level vs global, strategie d'invalidation, et cout memoire estime pour K=4096, n=1024.

#### Objection 4 (mineure) : Les seuils de parallelisme ne sont pas valides empiriquement

- **Categorie** : Risque
- **Severite** : Mineure
- **Argument** : Le rapport recommande d'augmenter `MaxParallelFFTDepth` a `log2(NumCPU)` sans benchmark. Sur une machine 16 coeurs, cela donnerait depth=4 (16 branches). Mais la FFT a deja un semaphore limite a `NumCPU` tokens (`fft_recursion.go:18-23`). Augmenter la depth sans augmenter les tokens ne creerait que de la contention. De plus, le `select/default` non-bloquant (`fft_recursion.go:80-114`) fait que les branches excessives tombent dans le fallback sequentiel, rendant l'augmentation de depth potentiellement sans effet.
- **Fichiers concernes** : `internal/bigfft/fft_recursion.go:18-23` (semaphore = NumCPU), `internal/bigfft/fft_recursion.go:79-114` (parallelisme avec select/default)
- **Critere de resolution** : Benchmark avec depth=3,4,5 et tokens=NumCPU, 2*NumCPU pour F(10^7) sur une machine 16+ coeurs. Presenter les resultats avant de modifier les constantes.

#### Objection 5 (mineure) : Le sqrt(2) trick (3.6) est presente comme "moyen effort" alors qu'il modifie le coeur de l'algorithme

- **Categorie** : Risque
- **Severite** : Mineure
- **Argument** : Le rapport estime 5-7 jours pour le sqrt(2) trick et un gain de 5-15%. C'est sous-estime en effort. Le trick modifie `fftSize()` (`fft.go:193-207`), `valueSize()` (`fft.go:215-225`), et potentiellement toute la chaine `fftSizeThreshold` (`fft.go:184-188`). Il change la relation fondamentale entre K (taille FFT) et n (taille des coefficients), ce qui impacte le bump allocator capacity estimation (`bump.go:209-242`), le cache de transformees (cle de cache changee), et les pools (tailles de classe). C'est un changement structurel de 10-15 jours realistes avec tests complets.
- **Fichiers concernes** : `internal/bigfft/fft.go:184-207` (fftSizeThreshold, fftSize), `internal/bigfft/fft.go:215-225` (valueSize), `internal/bigfft/bump.go:209-242` (EstimateBumpCapacity), `internal/bigfft/fft_cache.go:103-115` (cle de cache)
- **Critere de resolution** : Requalifier l'effort a 10-15 jours et la placer en Phase 3 uniquement. Pas de Phase 2.

---

## Elements de comparaison transversaux

### 1. Ratio effort/gain

| Proposition | Effort total | Gain estime (realiste) | Ratio jours/pourcent |
|-------------|-------------|----------------------|---------------------|
| NTT (Go pur) | 4 semaines (20j) | 0% (regression probable) | N/A (negatif) |
| NTT (avec SIMD) | 7 semaines (35j) | 10-20% pour F(10^9)+ seulement | 2-3.5 j/% |
| Montgomery | 3 semaines (15j) | -25% (regression) | N/A (negatif) |
| SS Phase 1 (quick wins) | 3 jours | 5-8% | 0.4-0.6 j/% |
| SS Phase 2 (twiddle+cache) | 5 jours | 3-7% | 0.7-1.7 j/% |
| SS Phase 3 (SIMD+sqrt2) | 15 jours | 10-20% (incertain) | 0.75-1.5 j/% |
| **Abaisser le seuil FFT** | **0.5 jour** | **2-5%** (a valider) | **0.1-0.25 j/%** |

L'optimisation au meilleur ROI est potentiellement la plus simple : **ajuster `DefaultFFTThreshold`** de 500K a une valeur optimale determinee par benchmark. C'est une modification d'une seule ligne dans `constants.go:28`.

### 2. Risque de regression

| Proposition | Risque regression petits N | Risque regression grands N | Risque maintenance |
|-------------|---------------------------|---------------------------|-------------------|
| NTT | Eleve (2-3x plus lent) | Moyen (dependent SIMD) | Eleve (deux backends) |
| Montgomery | Eleve (~25% plus lent) | Eleve (~25% plus lent) | Moyen |
| SS ameliore | Faible (seuils conditionnels) | Faible | Faible |

### 3. Compatibilite avec l'architecture existante

Toutes les propositions s'integrent via le pattern Strategy (`Multiplier`/`DoublingStepExecutor` dans `strategy.go`), ce qui est un point fort de l'architecture actuelle. Cependant :

- **NTT** necessiterait une nouvelle strategie `NTTStrategy` ET un nouveau backend dans `bigfft/`. L'interface `TempAllocator` (`allocator.go`) supporte les deux allocateurs mais ne couvre pas la divergence algorithmique.
- **Montgomery** s'integrerait proprement comme une `MontgomeryStrategy` mais serait inutilisee (overhead sans gain).
- **SS ameliore** modifie le code existant sans ajouter de nouvelle strategie -- c'est l'approche la moins invasive.

---

## Questions ouvertes pour les chercheurs

### Pour ntt-researcher :
1. Quelle est votre estimation du cout CRT par coefficient en cycles CPU pour 3 primes 63-bit en Go pur ? Avez-vous un prototype ?
2. Comment gerez-vous le cache de transformees avec K primes ? K caches separes ou un cache unifie K fois plus grand ?
3. Etes-vous d'accord que le prototype Go pur Phase 1 sera probablement plus lent ? Si oui, pourquoi investir 2 semaines avant de valider la Phase 3 SIMD qui est le vrai differentiateur ?

### Pour montgomery-researcher :
1. Etes-vous d'accord que l'ecart entre `fftThreshold` (1800 mots = ~115K bits) et `DefaultFFTThreshold` (500K bits) merite investigation comme "quick win" independant ?
2. Avez-vous envisage Montgomery specifiquement pour l'arithmetique Fermat interne (`fermat.Mul` dans `fermat.go:149-205`) plutot que pour la multiplication de haut niveau ? Montgomery pourrait accelerer les multiplications modulaires internes a la FFT si le modulus 2^n+1 est fixe pendant toute la FFT.

### Pour ss-researcher :
1. Pouvez-vous verifier si `big.Int.Mul(xi, yi)` detecte le squaring quand xi et yi sont des `big.Int` distincts construits avec `SetBits` pointant vers la meme memoire ? C'est critique pour evaluer le gain reel de `fermat.Sqr`.
2. Avez-vous profile la boucle Reconstruct (`fft_recursion.go:126-131`) pour confirmer qu'elle represente ~40% du temps FFT ? Si c'est < 25%, le twiddle factor cache est deprioritise.
3. Pour le sqrt(2) trick, etes-vous d'accord que l'effort est plutot 10-15 jours que 5-7 jours vu l'impact sur `fftSizeThreshold`, `valueSize`, `EstimateBumpCapacity`, et le cache ?
