# Rapport : Optimisation Schönhage-Strassen

## Resume executif

L'implementation Schönhage-Strassen actuelle de FibGo est deja bien optimisee avec un bump allocator O(1), un cache LRU pour les transformees, un systeme de pools multi-niveaux, et un parallelisme adaptatif. Cependant, plusieurs optimisations significatives restent possibles : (1) pre-calcul et cache des twiddle factors (shifts) pour eviter les recalculs dans la boucle Reconstruct, (2) amelioration de la localite cache dans `fourierRecursiveUnified` via un passage iteratif-recursif hybride, (3) optimisation du squaring en frequence (pointwise) via une routine specialisee `fermat.Sqr()`, et (4) ajustement des seuils de parallelisme FFT. **Recommandation : Approfondir et implementer les optimisations 1, 3 et 4, qui offrent un gain estime cumule de 15-25% pour les grandes tailles (N > 1M) avec un effort modere.**

## 1. Description de l'approche

L'implementation actuelle utilise l'algorithme classique de Schönhage-Strassen operant dans Z/(2^(n*W)+1)Z avec :
- FFT recursive de type Cooley-Tukey (DIT) sur des nombres de Fermat
- Twiddle factors calcules par des shifts (rotation cyclique modulo 2^n+1) via `fermat.ShiftHalf()`
- Multiplication pointwise dans le domaine frequentiel
- Parallelisme via goroutines avec semaphore de concurrence

Les gains identifies sont :
- **Twiddle factor caching** : 5-10% sur les grandes tailles en evitant les recalculs de shifts
- **Squaring specialise** : 5-8% en exploitant x=y dans `fermat.Mul(x,y)`
- **Seuils de parallelisme FFT** : 3-7% par ajustement fin des constantes
- **Localite cache amelioree** : 3-5% via restructuration de la boucle Reconstruct

## 2. Analyse du code actuel

### Points forts
1. **Architecture TempAllocator** : L'interface `TempAllocator` avec `PoolAllocator` et `BumpAllocatorAdapter` est elegante et permet de choisir la strategie d'allocation sans duplication de code.
2. **Bump Allocator** : Allocation O(1) avec fallback automatique. Excellent pour la localite cache des FFT temporaires.
3. **Cache LRU pour transformees** : Le `TransformCache` (128 entrees, seuil 100K bits) avec statistiques atomiques (hits/misses/evictions) est bien concu pour les calculs iteratifs de Fibonacci.
4. **Pools multi-niveaux** : 10 classes de taille pour `wordSlicePools` (64 a 16M mots), 9 pour `fermatPools`, avec indexation O(1) par bits.Len.
5. **Parallelisme adaptatif** : Semaphore non-bloquant (`select`/`default`) pour eviter la contention excessive.
6. **Detection CPU** : Detection AVX2/AVX-512/BMI2/ADX pour adapter les strategies.
7. **go:linkname vers math/big** : Reutilisation des routines assembleur optimisees de Go (addVV, subVV, etc.) sans duplication.

### Faiblesses
1. **Twiddle factors recalcules a chaque fois** : Dans `fourierRecursiveUnified`, la boucle Reconstruct (ligne 127-131) appelle `tmp.ShiftHalf(dst2[i], i*omega2shift, tmp2)` a chaque papillon. Ces shifts sont coûteux et pourraient etre pre-calcules.
2. **Pas de squaring specialise dans fermat.Mul** : La multiplication pointwise `buf.Mul(p.Values[i], p.Values[i])` dans `sqr()` (fft_poly.go:387) appelle `fermat.Mul(x, y)` generique meme quand x==y. Un squaring specialise serait plus rapide.
3. **Parallelisme FFT trop conservateur** : `ParallelFFTRecursionThreshold=4` et `MaxParallelFFTDepth=3` sont des valeurs codees en dur sans calibration. Pour des FFT de taille K=4096+ sur des machines >8 coeurs, ces seuils sont trop conservateurs.
4. **Cache LRU potentiellement sous-dimensionne** : 128 entrees max pour le cache de transformees. Pour les calculs iteratifs avec fast doubling (qui reutilisent beaucoup les memes transformees), un cache plus grand pourrait augmenter le hit rate.
5. **Pas de sqrt(2) trick** : L'implementation utilise des shifts par sqrt(2) via `ShiftHalf`, mais ne tire pas parti du "sqrt(2) trick" de GMP qui permet de doubler la taille du FFT sans augmenter la taille des coefficients, en utilisant le fait que sqrt(2) = 2^(3n/4) - 2^(n/4) mod (2^n+1).
6. **Vectorisation indirecte uniquement** : Les operations `addVV`/`subVV`/`addMulVVW` sont deleguees a `math/big` via `go:linkname`. Aucune vectorisation AVX2/AVX-512 directe n'est utilisee dans le code bigfft lui-meme, malgre la detection CPU.
7. **Pool warming calibration fixe** : Les seuils de pre-chauffage (2/4/5/6 buffers selon N) sont heuristiques et non calibres empiriquement.

### Fichiers analyses

| Fichier | Lignes | Observations cles |
|---------|--------|-------------------|
| `fft_core.go` | 105 | Point d'entree FFT. 3 variantes : `fourier`, `fourierWithState`, `fourierWithBump`. `fftmulTo`/`fftsqrTo` sont les fonctions principales avec bump allocator et cache. |
| `fft_recursion.go` | 140 | Coeur recursif `fourierRecursiveUnified`. Parallelisme via semaphore non-bloquant. Seuils : k>=4, depth<3. La boucle Reconstruct (L127-131) est le hotpath. |
| `fft_cache.go` | 449 | Cache LRU thread-safe (128 entrees, min 100K bits). Hash FNV-1a. Deep copy des valeurs cachees avec backing buffer contigu. Statistiques atomiques. |
| `bump.go` | 243 | Bump allocator O(1) avec fallback make(). `AllocFermat`, `AllocFermatSlice`. `EstimateBumpCapacity` heuristique avec 10% de marge. |
| `pool.go` | 464 | 4 types de pools (wordSlice, fermat, natSlice, fermatSlice). Indexation O(1) par bits.Len. `fftState` pool combine tmp+tmp2. |
| `pool_warming.go` | 99 | Pre-chauffage adaptatif (2-6 buffers selon N). `EnsurePoolsWarmed` avec atomic CAS. |
| `cpu_amd64.go` | 170 | Detection AVX2/AVX-512/BMI2/ADX via golang.org/x/sys/cpu. Resultat en `CPUFeatures` struct. |
| `memory_est.go` | 77 | Estimation heuristique des besoins memoire pour F(n). Utilise log2(phi) pour estimer la taille de F(n). |
| `fermat.go` | 219 | Arithmetique modulo 2^(n*W)+1. `Shift`, `ShiftHalf` (twiddle factors), `Add`, `Sub`, `Mul`. `smallMulThreshold=30` pour basicMul vs big.Int.Mul. |
| `fft_poly.go` | 417 | Polynomes et transformees. `transform`, `invTransform`, `mul` (pointwise), `sqr` (pointwise), `Clone`. |
| `fft.go` | 226 | API publique (Mul, MulTo, Sqr, SqrTo). `fftSizeThreshold` table. `valueSize` calcul. `fftThreshold=1800` mots. |
| `arith_decl.go` | 66 | go:linkname vers math/big : addVV, subVV, addVW, subVW, shlVU, mulAddVWW, addMulVVW. |
| `arith_amd64.go` | 33 | Wrappers publics AddVV/SubVV/AddMulVVW delegant aux fonctions linkname. |
| `arith_generic.go` | 37 | Fallback portable identique pour non-amd64. |
| `allocator.go` | 110 | Interface TempAllocator + PoolAllocator + BumpAllocatorAdapter. Pattern strategy. |

## 3. Optimisations proposees

### 3.1 Pre-calcul des twiddle factors (shifts)

**Description** : Dans `fourierRecursiveUnified` (fft_recursion.go:126-131), la boucle Reconstruct calcule `tmp.ShiftHalf(dst2[i], i*omega2shift, tmp2)` pour chaque papillon. La valeur `i*omega2shift` est deterministe et pourrait etre pre-calculee une seule fois par niveau de recursion, puis reutilisee.

**Mecanisme** : Creer une table de twiddle factors pre-calcules indexee par (size, backward). Comme les shifts sont dans Z/(2^(n*W)+1)Z, on peut stocker les resultats de `ShiftHalf` pre-calcules pour les tailles les plus frequentes, ou au minimum pre-calculer les offsets `i*omega2shift` et les stocker dans un tableau pour eviter les multiplications repetees.

**Alternative GMP** : GMP utilise le fait que les twiddle factors sont de simples puissances de 2 (shifts), et optimise en combinant le shift avec l'addition papillon dans une seule operation. Cela evite les ecritures intermediaires en memoire.

**Impact estime** : 5-10% d'acceleration sur la FFT pour les grandes tailles (K >= 256). La boucle Reconstruct represente ~40% du temps FFT.

**Effort** : Moyen (2-3 jours). Modifier `fourierRecursiveUnified` et potentiellement ajouter un cache de twiddle factors.

### 3.2 Squaring specialise dans le domaine frequentiel

**Description** : Dans `fft_poly.go:384-389`, la fonction `sqr()` calcule `buf.Mul(p.Values[i], p.Values[i])`. La fonction `fermat.Mul(x, y)` (fermat.go:149-205) ne detecte pas que x==y et utilise le chemin generique. Pour le squaring, `basicMul(z, x, x)` peut etre remplace par un `basicSqr(z, x)` qui evite la moitie des multiplications intermediaires (symetrie de la matrice de produits partiels).

**Mecanisme** : Ajouter une methode `fermat.Sqr(x)` qui :
- Pour n < smallMulThreshold : utilise une version `basicSqr` exploitant la symetrie
- Pour n >= smallMulThreshold : utilise `big.Int.Mul(x, x)` (Go l'optimise deja en interne pour le squaring)
- Modifier `sqr()` dans fft_poly.go pour appeler `buf.Sqr(p.Values[i])` au lieu de `buf.Mul(p.Values[i], p.Values[i])`

**Impact estime** : 5-8% d'acceleration sur le squaring total (qui represente ~50% des operations dans fast doubling). Pour `basicMul` avec n=29 (seuil), on economise ~30% des produits partiels.

**Effort** : Faible (1-2 jours). Ajouter `fermat.Sqr` et `basicSqr`, modifier `sqr()`.

### 3.3 Ajustement des seuils de parallelisme FFT

**Description** : Les constantes `ParallelFFTRecursionThreshold=4` et `MaxParallelFFTDepth=3` dans fft_recursion.go sont fixees en dur. Sur des machines >8 coeurs, `MaxParallelFFTDepth=3` ne cree que 2^3=8 branches paralleles au maximum, ce qui sous-utilise les coeurs. De plus, le seuil k>=4 (FFT de taille 16) est potentiellement trop bas pour amortir l'overhead des goroutines.

**Mecanisme** :
1. Rendre ces seuils configurables via `Options` ou au minimum via des variables package
2. Augmenter `MaxParallelFFTDepth` a `log2(NumCPU)` pour les machines a beaucoup de coeurs
3. Augmenter `ParallelFFTRecursionThreshold` a 5 ou 6 pour les petites FFT ou l'overhead goroutine domine
4. Integrer dans le systeme de calibration existant (`internal/calibration`)

**Impact estime** : 3-7% d'acceleration, tres dependant du hardware. Gain majeur sur >16 coeurs.

**Effort** : Faible (1 jour). Modifier les constantes, ajouter la configurabilite.

### 3.4 Amelioration de la localite cache dans la boucle Reconstruct

**Description** : La boucle Reconstruct dans `fourierRecursiveUnified` (L127-131) accede a `dst1[i]`, `dst2[i]`, et `tmp` de maniere lineaire. Cependant, pour les grandes FFT, `dst1` et `dst2` sont des slices de fermat dont les elements pointent vers des zones memoire potentiellement eloignees (surtout quand le pool allocator est utilise au lieu du bump allocator).

**Mecanisme** :
- Utiliser le bump allocator de maniere plus agressive pour la boucle Reconstruct
- Implementer un "blocking" de la boucle Reconstruct (traiter par blocs de taille cache-friendly, par exemple 4-8 papillons a la fois) pour maximiser les hits L1
- Pour les goroutines paralleles dans le FFT recursif, utiliser systematiquement le bump allocator au lieu du pool allocator (actuellement le code utilise le pool pour les goroutines paralleles a cause du thread-safety)

**Impact estime** : 3-5% pour les grandes tailles (K >= 1024). Negligeable pour les petites tailles deja en L1.

**Effort** : Moyen (2-3 jours). Modification delicate de `fourierRecursiveUnified`.

### 3.5 Augmentation de la taille du cache de transformees

**Description** : Le `TransformCache` est dimensionne a 128 entrees max avec un seuil minimum de 100K bits. Pour les calculs de Fibonacci iteratifs utilisant fast doubling, le pattern d'acces aux transformees est tres regulier (memes valeurs reutilisees a travers les iterations). Un cache plus grand pourrait augmenter le hit rate.

**Mecanisme** :
1. Augmenter `MaxEntries` a 256 ou 512 pour les grandes calculations
2. Ajouter un mecanisme adaptatif basee sur la taille du calcul (similaire a `PreWarmPools`)
3. Ajouter des metriques de hit rate dans les logs pour monitorer l'efficacite
4. Considerer un cache a deux niveaux (L1 petit/rapide, L2 grand/plus lent) pour les tres grandes tailles

**Impact estime** : 2-5% si le hit rate actuel est bas (< 50%). Negligeable si le hit rate est deja eleve. Necessite des mesures.

**Effort** : Faible (0.5 jour). Modifier la configuration par defaut. Le mecanisme adaptatif prendrait plus de temps.

### 3.6 Sqrt(2) Trick a la GMP

**Description** : L'implementation actuelle utilise `ShiftHalf` qui calcule sqrt(2) mod (2^n+1) comme 2^(3n/4) - 2^(n/4). C'est le "sqrt(2) trick" classique. Cependant, GMP va plus loin : il utilise ce fait pour choisir des tailles de FFT plus grandes (K plus grand) sans augmenter la taille des coefficients n. Concretement, quand sqrt(2) est disponible comme racine primitive, on peut doubler K tout en gardant le meme n, ce qui reduit la taille des produits partiels.

L'implementation actuelle de FibGo utilise deja `ShiftHalf` correctement pour les twiddle factors de demi-entier, mais ne tire pas parti de la possibilite d'augmenter K pour un n donne. Ce serait un changement structurel dans `fftSize()`.

**Impact estime** : 5-15% potentiel pour les grandes tailles en reduisant le nombre de mots par coefficient. Mais c'est un changement complexe qui modifie le coeur de l'algorithme.

**Effort** : Eleve (5-7 jours). Modification de `fftSize`, `valueSize`, et potentiellement tout le pipeline FFT.

### 3.7 Vectorisation SIMD directe pour les operations Fermat

**Description** : Actuellement, `addVV`, `subVV`, `addMulVVW` sont delegues a `math/big` via `go:linkname`. Les routines assembleur de Go sont bonnes mais generiques. Pour les operations specifiques a la FFT modulo 2^n+1 (shift, add, sub, norm), des routines AVX2/AVX-512 specialisees pourraient etre plus rapides car elles connaissent la structure du probleme.

**Mecanisme** : Ecrire des routines assembleur Go (`.s` files) pour :
- `fermat.Shift` optimise avec VPSHLDQ (AVX-512) ou VPSLLVQ/VPSRLVQ (AVX2)
- `fermat.Add`/`fermat.Sub` avec propagation de retenue vectorisee
- `fermat.norm` optimise

**Impact estime** : 10-20% potentiel sur les operations Fermat, qui sont omnipresentes. Mais les routines math/big de Go sont deja bien optimisees, donc le gain reel depend de la proportion du temps passe dans ces routines vs. les appels recursifs.

**Effort** : Tres eleve (7-10 jours). Ecriture d'assembleur Go, tests, benchmarks, maintenance multi-plateforme.

### 3.8 Harvey-van der Hoeven O(n log n) : Analyse de faisabilite

**Description** : L'algorithme de Harvey-van der Hoeven (2019) atteint la complexite optimale O(n log n) pour la multiplication d'entiers, confirmant la conjecture de Schönhage-Strassen. Il utilise un "Gaussian resampling" et des transformees de Fourier multidimensionnelles.

**Verdict : Non implementable en pratique pour FibGo.**

Raisons :
1. Les auteurs eux-memes indiquent que l'algorithme n'est pas concu pour etre pratique et que les constantes sont enormes
2. Une implementation non-optimisee dans Mathemagix est un ordre de grandeur plus lente que GMP
3. Le crossover theorique avec Schönhage-Strassen se situe a des tailles d'entiers astronomiques (probablement > 2^(2^30) bits)
4. La complexite d'implementation est considerable (Gaussian resampling, Crandall/Fagin transforms)
5. Le gain asymptotique O(n log n) vs O(n log n log log n) est negligeable en pratique pour toute taille raisonnable de Fibonacci

**Recommandation** : Rejeter. L'effort serait enorme pour un gain nul en pratique.

## 4. Estimation de performance

### Estimation par taille

| N | Bits F(n) | Temps actuel (ref) | Apres opt 3.1+3.2+3.3 (estime) | Ratio |
|---|-----------|--------------------|---------------------------------|-------|
| 100,000 | ~69K | 1.00x | ~0.92x | 1.09x |
| 500,000 | ~347K | 1.00x | ~0.88x | 1.14x |
| 1,000,000 | ~694K | 1.00x | ~0.85x | 1.18x |
| 5,000,000 | ~3.5M | 1.00x | ~0.82x | 1.22x |
| 10,000,000 | ~6.9M | 1.00x | ~0.80x | 1.25x |
| 50,000,000 | ~34.7M | 1.00x | ~0.78x | 1.28x |

### Gain cumule estime

| Optimisation | Gain estime | Confiance |
|-------------|-------------|-----------|
| 3.1 Twiddle factor pre-calcul | 5-10% | Moyenne-Haute |
| 3.2 Squaring specialise | 5-8% | Haute |
| 3.3 Seuils parallelisme | 3-7% | Moyenne |
| 3.4 Localite cache Reconstruct | 3-5% | Moyenne |
| 3.5 Cache transformees agrandi | 2-5% | Basse (necessite mesures) |
| 3.6 Sqrt(2) trick complet | 5-15% | Basse (complexe) |
| 3.7 SIMD Fermat | 10-20% | Moyenne (effort eleve) |
| **Total cumule (3.1-3.5)** | **15-25%** | **Moyenne** |
| **Total avec 3.6-3.7** | **25-45%** | **Basse** |

Note : Les gains ne sont pas parfaitement additifs car certaines optimisations touchent les memes chemins critiques.

## 5. Effort d'integration

### Fichiers a modifier

| Fichier | Type de modification | Effort estime |
|---------|---------------------|---------------|
| `internal/bigfft/fft_recursion.go` | Twiddle cache + seuils parallelisme | 3 jours |
| `internal/bigfft/fermat.go` | Ajouter `Sqr()`, `basicSqr()` | 1 jour |
| `internal/bigfft/fft_poly.go` | Modifier `sqr()` pour utiliser `fermat.Sqr()` | 0.5 jour |
| `internal/bigfft/fft_cache.go` | Config adaptative, metriques enrichies | 1 jour |
| `internal/bigfft/fft_core.go` | Integration twiddle cache dans `fourierWithBump` | 1 jour |
| `internal/fibonacci/constants.go` | Nouveaux seuils configurables | 0.5 jour |
| `internal/fibonacci/options.go` | Options pour seuils FFT parallelisme | 0.5 jour |
| `internal/calibration/microbench.go` | Benchmarks pour calibrer les seuils FFT | 1 jour |
| Tests associes | Tests unitaires, benchmarks, fuzz | 2 jours |
| **Total** | | **~10.5 jours** |

### Effort total estime

- **Phase 1 (Quick wins : 3.2 + 3.3 + 3.5)** : 3 jours - Gain estime 10-15%
- **Phase 2 (Twiddle cache : 3.1 + 3.4)** : 5 jours - Gain supplementaire 5-10%
- **Phase 3 (Avance : 3.6 + 3.7)** : 12-17 jours - Gain supplementaire 10-20%, risque eleve

## 6. Risques

| Risque | Probabilite | Impact | Mitigation |
|--------|-------------|--------|------------|
| Regression de performance pour les petites tailles (N < 100K) | Moyenne | Moyen | Benchmarks avant/apres pour chaque taille. Seuils conditionnels. |
| Bugs subtils dans le twiddle factor caching (coherence cache) | Moyenne | Eleve | Golden file tests, comparaison avec l'implementation actuelle pour 10K+ valeurs |
| Gains reels inferieurs aux estimations | Haute | Moyen | Profiling rigoureux avant implementation. Abandonner si gain < 3%. |
| go:linkname casse avec une future version de Go | Faible | Eleve | Veille sur les releases Go. Fallback pur-Go deja present (arith_generic.go). |
| Complexite accrue rendant le code moins maintenable | Moyenne | Moyen | Documentation exhaustive. Garder les chemins simples par defaut. |
| Le squaring `big.Int.Mul(x,x)` est deja optimise en interne par Go | Haute | Faible | Ne concerne que basicSqr pour n < 30 mots. Benchmark pour confirmer le gain. |
| Le cache de transformees consomme trop de memoire | Faible | Faible | Limiter MaxEntries. Ajouter un mecanisme d'eviction basee sur la memoire. |

## 7. Recommandation

**Recommandation : Approfondir et implementer en phases.**

### Phase 1 - Quick Wins (Priorite Haute, 3 jours)
1. **3.2 Squaring specialise** : Le gain le plus sur avec l'effort le plus faible. `fermat.Sqr()` est isole et facile a tester.
2. **3.3 Seuils parallelisme** : Rendre configurables et ajuster. Integrer dans le systeme de calibration.
3. **3.5 Cache transformees** : Simple augmentation de MaxEntries + metriques pour valider.

### Phase 2 - Optimisations Structurelles (Priorite Moyenne, 5 jours)
4. **3.1 Twiddle factor pre-calcul** : Gain significatif mais necessite une refactorisation delicate de `fourierRecursiveUnified`.
5. **3.4 Localite cache** : Complementaire a 3.1, optimise le meme chemin critique.

### Phase 3 - Optimisations Avancees (Priorite Basse, decision apres Phase 1-2)
6. **3.6 Sqrt(2) trick complet** : Seulement si les gains des Phases 1-2 sont insuffisants. Risque eleve.
7. **3.7 SIMD Fermat** : Seulement si le profiling montre que les operations Fermat sont le goulot d'etranglement apres Phases 1-2.

### Rejection explicite
- **3.8 Harvey-van der Hoeven** : Rejete. Non applicable en pratique.

### Justification
L'implementation actuelle est deja mature et bien optimisee. Les gains les plus accessibles (Phase 1) offrent un ratio effort/gain excellent. Les Phases 2-3 doivent etre validees par du profiling apres Phase 1. Le focus devrait rester sur l'algorithme SS actuel plutot que de migrer vers un autre algorithme de multiplication, car les optimisations proposees sont incrementales et a faible risque.

### Comparaison avec les alternatives (NTT, Montgomery)
L'optimisation de l'implementation SS existante est complementaire a l'evaluation d'alternatives comme NTT ou Montgomery. Si les gains cumules des Phases 1-2 atteignent 15-20%, cela pourrait rendre moins attractif le cout de migration vers un algorithme entierement different. Inversement, si les gains sont decevants (< 10%), cela renforcerait l'argument pour une migration.

### References
- Gaudry, Kruppa, Zimmermann. "A GMP-based implementation of Schonhage-Strassen's large integer multiplication algorithm." ISSAC 2007.
- Harvey, van der Hoeven. "Integer multiplication in time O(n log n)." Annals of Mathematics, 2021.
- GMP Manual, Section "FFT Multiplication" - https://gmplib.org/manual/FFT-Multiplication
- Fabian Giesen. "Notes on FFTs: for implementers." 2023 - https://fgiesen.wordpress.com/2023/03/19/notes-on-ffts-for-implementers/
- FLINT 3.1 Release Notes (fft_small module) - https://fredrikj.net/blog/2024/02/whats-new-in-flint-3-1/
