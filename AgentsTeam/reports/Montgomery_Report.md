# Rapport : Multiplication de Montgomery

## Resume executif

La multiplication de Montgomery est une technique de **multiplication modulaire** optimisee, concue pour accelerer les chaines d'operations `a*b mod N` en evitant les divisions couteuses. Apres analyse approfondie du code FibGo et de la litterature, **la recommandation est de REJETER** l'integration de Montgomery comme strategie de multiplication dans le projet. Montgomery est fondamentalement inadapte a la multiplication d'entiers libres (non modulaire) requise par le calcul de Fibonacci via Fast Doubling. Le cout d'adaptation (introduction d'un modulus artificiel, conversions Montgomery form) annulerait tout benefice theorique et ajouterait une complexite significative sans gain de performance.

## 1. Description de l'approche

### Principe algorithmique

La multiplication de Montgomery (Peter L. Montgomery, 1985) repose sur une representation speciale des nombres : la "Montgomery form". Pour un modulus N et un auxiliaire R (typiquement une puissance de 2, R > N, gcd(R, N) = 1) :

- **Montgomery form** : `a_bar = a * R mod N`
- **Montgomery product** : `MonPro(a_bar, b_bar) = a_bar * b_bar * R^{-1} mod N`
- **Reduction REDC** : Remplace la division par N par une division par R (simple shift si R = 2^k)

L'algorithme REDC :
```
function REDC(T):
    m = (T mod R) * N' mod R    // N' = -N^{-1} mod R
    t = (T + m*N) / R
    if t >= N: return t - N
    else: return t
```

### Complexite theorique

| Operation | Complexite |
|-----------|-----------|
| Conversion vers Montgomery form | O(n) (une multiplication + REDC) |
| MonPro (produit Montgomery) | O(n^2) pour schoolbook, O(n^1.585) avec Karatsuba interne |
| Conversion depuis Montgomery form | O(n) (un REDC) |
| Chaine de k multiplications mod N | O(k * M(n)) ou M(n) est le cout de multiplication |

### Variantes

1. **CIOS** (Coarsely Integrated Operand Scanning) : Entrelace multiplication et reduction mot par mot
2. **FIOS** (Finely Integrated Operand Scanning) : Integration plus fine, meilleure localite cache
3. **SOS** (Separated Operand Scanning) : Separation multiplication/reduction, plus simple a vectoriser
4. **Montgomery avec Karatsuba** : Combine Karatsuba pour la multiplication interne et Montgomery pour la reduction
5. **Batched SIMD Montgomery** : Exploitation AVX-512 VPMADD52 pour paralleliser sur plusieurs operandes independants

## 2. Analyse du code actuel

### Points forts de l'implementation actuelle

1. **Architecture strategie bien concue** : Le pattern Strategy (`Multiplier` / `DoublingStepExecutor`) dans `strategy.go` permet d'ajouter de nouvelles strategies sans modifier le code existant.
2. **Seuils adaptatifs** : `AdaptiveStrategy` choisit entre `math/big` (Karatsuba) et FFT selon la taille des operandes, avec seuils configurables.
3. **FFT transform reuse** : `executeDoublingStepFFT()` dans `fft.go` transforme les operandes une seule fois et reutilise les transformees pour les 3 multiplications du doubling step -- une optimisation analogue a ce que Montgomery offre pour les chaines modulaires.
4. **Arithmetique vectorielle** : `arith_decl.go` utilise `go:linkname` vers les fonctions assembleur de `math/big` (addVV, subVV, mulAddVWW) qui exploitent deja les instructions SIMD du CPU.
5. **Arithmetique Fermat** : `fermat.go` effectue deja de l'arithmetique modulaire (mod 2^n+1) pour la FFT, ce qui est conceptuellement similaire a Montgomery mais avec un modulus de Fermat optimise pour les shifts.

### Faiblesses exploitables (perspective Montgomery)

1. **Gap entre Karatsuba et FFT** : Le seuil FFT est a ~500,000 bits. Entre ~10,000 et ~500,000 bits, seul `math/big.Mul` (Karatsuba/Toom-Cook interne) est utilise. C'est le "gap" ou une technique intermediaire pourrait theoriquement aider.
2. **Pas d'optimisation Toom-Cook explicite** : Le code repose sur l'implementation interne de Go `math/big` qui utilise Karatsuba puis Toom-Cook. Aucun controle sur les seuils internes.

### Fichiers analyses

| Fichier | Lignes | Observations cles |
|---------|--------|-------------------|
| `internal/bigfft/fft.go` | 226 | Seuil FFT = 1800 words (~115 kbits). Fonctions Mul/Sqr/MulTo/SqrTo avec dispatch adaptatif. fftSizeThreshold pour selectionner la taille FFT. |
| `internal/bigfft/arith_amd64.go` | 33 | Wrappers AddVV/SubVV/AddMulVVW delegant vers les fonctions assembleur de math/big via go:linkname. Pas d'implementation AVX2 custom. |
| `internal/bigfft/arith_generic.go` | 37 | Fallback portable identique au amd64 (memes delegations go:linkname). |
| `internal/bigfft/arith_decl.go` | 66 | Declarations go:linkname vers addVV, subVV, addVW, subVW, shlVU, mulAddVWW, addMulVVW de math/big. Base vectorielle commune. |
| `internal/bigfft/fermat.go` | 219 | Arithmetique mod 2^(w*W)+1. Mul utilise basicMul ou big.Int.Mul selon seuil (30 words). Deja une forme de reduction modulaire specialisee. |
| `internal/fibonacci/strategy.go` | 188 | 3 strategies : AdaptiveStrategy (Karatsuba+FFT), FFTOnlyStrategy, KaratsubaStrategy. Interface Multiplier (narrow) et DoublingStepExecutor (wide). |
| `internal/fibonacci/constants.go` | 73 | Seuils : ParallelThreshold=4096, FFTThreshold=500,000, StrassenThreshold=3072, ParallelFFTThreshold=5,000,000. |
| `internal/fibonacci/fft.go` | 239 | smartMultiply/smartSquare (dispatch 2-tiers). executeDoublingStepFFT avec reutilisation de transformees FFT et parallelisation. |
| `internal/fibonacci/options.go` | 86 | Options configurable : seuils, cache FFT, seuils dynamiques. normalizeOptions() pour valeurs par defaut. |

## 3. Comparaison theorique

### Complexite

| Operation | Actuel (Karatsuba/SS) | Montgomery (theorique) | Avantage |
|-----------|----------------------|----------------------|----------|
| Multiplication simple (< 500K bits) | O(n^1.585) Karatsuba via math/big | O(n^2) schoolbook + REDC, ou O(n^1.585) Karatsuba + REDC | **Actuel** : Montgomery ajoute le cout REDC sans reduire la complexite de la multiplication sous-jacente |
| Multiplication simple (> 500K bits) | O(n log n log log n) SS/FFT | O(n log n log log n) multiplication + O(n) REDC | **Actuel** : Montgomery ajoute un overhead REDC inutile |
| Chaine de k multiplications mod N | O(k * M(n)) sans optimisation | O(2*M(n) + k*M(n)) avec amortissement | **Montgomery** : Mais seulement en contexte modulaire |
| Doubling step (3 mults) | O(3 * M(n)) avec FFT reuse | O(2*M(n) + 3*MonPro(n)) | **Actuel** : La reutilisation FFT est deja plus efficace que Montgomery pour 3 operations |
| Squaring x^2 | O(0.67 * M(n)) avec FFT (1 seule transformee) | O(M(n)) MonPro complet | **Actuel** : L'optimisation squaring FFT est superieure |

### Constantes cachees

1. **Conversion Montgomery form** : Chaque entree/sortie de Montgomery form coute une multiplication + REDC. Pour le Fast Doubling, chaque iteration utilise des valeurs intermediaires F(k) et F(k+1) qui changent a chaque step. Il n'y a **pas de modulus fixe** a travers les iterations, donc les conversions ne s'amortissent pas.

2. **Pas de modulus naturel pour Fibonacci** : F(n) est un entier libre, sans modulus. Introduire un modulus artificiel (ex: 2^B+1 pour B > bitlen(F(n))) transformerait Montgomery en une surcouche inutile car la "reduction" serait triviale (le resultat est toujours < modulus).

3. **Fermat vs Montgomery** : L'arithmetique Fermat dans `fermat.go` EST deja une forme optimisee de reduction modulaire (mod 2^n+1), parfaitement adaptee a la FFT. Montgomery n'apporte rien de plus dans ce contexte.

4. **Branch misprediction** : Montgomery REDC a une branche conditionnelle finale (`if t >= N: t -= N`) qui est mal predite ~50% du temps pour des entiers aleatoires, ajoutant de la latence.

## 4. Estimation de performance

### Estimation par taille

| N | Bits F(n) | Actuel (estime) | Montgomery (estime) | Ratio |
|---|-----------|-----------------|---------------------|-------|
| 10^3 | ~700 | < 1ms (Karatsuba) | < 1ms (overhead conversion) | ~1.0x (equivalent) |
| 10^4 | ~7,000 | ~1ms (Karatsuba) | ~1.3ms (Karatsuba + REDC overhead) | **0.77x** (Montgomery plus lent) |
| 10^5 | ~70,000 | ~10ms (Karatsuba) | ~13ms (Karatsuba + REDC overhead) | **0.77x** (Montgomery plus lent) |
| 10^6 | ~700,000 | ~200ms (FFT) | ~260ms (FFT + REDC inutile) | **0.77x** (Montgomery plus lent) |
| 10^7 | ~7,000,000 | ~3s (FFT parallele) | ~4s (FFT + overhead) | **0.75x** (Montgomery plus lent) |
| 10^8 | ~70,000,000 | ~50s (FFT parallele) | ~65s (FFT + overhead) | **0.77x** (Montgomery plus lent) |
| 10^9 | ~700,000,000 | ~15min (FFT parallele) | ~20min (FFT + overhead) | **0.75x** (Montgomery plus lent) |
| 10^10 | ~7,000,000,000 | ~4h (FFT parallele) | ~5.3h (FFT + overhead) | **0.75x** (Montgomery plus lent) |

**Note** : Les estimations Montgomery incluent le surcout de conversion vers/depuis Montgomery form a chaque iteration du Fast Doubling, car il n'y a pas de modulus fixe permettant de rester en forme Montgomery entre iterations.

### Seuil de crossover

**Il n'existe pas de seuil de crossover favorable a Montgomery** dans ce contexte. Montgomery ne peut surpasser les methodes actuelles que dans un scenario de multiplication modulaire avec modulus fixe et longues chaines d'operations -- conditions absentes du calcul de Fibonacci par Fast Doubling.

Le seul scenario ou Montgomery apporterait un benefice serait le calcul de F(n) mod M (Fibonacci modulaire), ou une chaine de ~log2(n) multiplications modulaires avec M fixe permettrait d'amortir les conversions. Mais ce n'est pas le cas d'usage principal de FibGo qui calcule F(n) exact.

## 5. Effort d'integration

### Fichiers a modifier

| Fichier | Type de modification | Effort estime |
|---------|---------------------|---------------|
| `internal/fibonacci/strategy.go` | Ajouter `MontgomeryStrategy` implementant `DoublingStepExecutor` | 2-3 jours |
| `internal/fibonacci/montgomery.go` (nouveau) | Implementation REDC, conversion form, MonPro | 3-5 jours |
| `internal/fibonacci/montgomery_test.go` (nouveau) | Tests unitaires, golden file validation | 2-3 jours |
| `internal/fibonacci/constants.go` | Ajouter `DefaultMontgomeryThreshold` | 0.5 jour |
| `internal/fibonacci/options.go` | Ajouter `MontgomeryThreshold` dans Options | 0.5 jour |
| `internal/fibonacci/fft.go` | Modifier smartMultiply pour tier 3 Montgomery | 1 jour |
| `internal/fibonacci/registry.go` | Enregistrer la variante Montgomery | 0.5 jour |
| `internal/bigfft/montgomery.go` (nouveau) | Potentiel module Montgomery bas-niveau | 3-5 jours |
| Documentation | README, CLAUDE.md | 1 jour |

### Effort total estime

**14-20 jours-developpeur** pour une integration complete, incluant :
- Implementation de l'arithmetique Montgomery (REDC, form conversion)
- Integration dans le framework de strategies
- Tests et benchmarks
- Documentation

C'est un investissement tres eleve pour un gain de performance **negatif** (ralentissement attendu de ~25%).

## 6. Risques

| Risque | Probabilite | Impact | Mitigation |
|--------|-------------|--------|-----------|
| **Ralentissement systematique** : Montgomery ajoute un overhead REDC sans benefice en multiplication non-modulaire | Elevee (95%) | Eleve : Degradation de 20-30% sur toutes les tailles | Benchmarks comparatifs avant merge. Mais le resultat negatif est quasi-certain d'apres l'analyse theorique. |
| **Complexite du code accrue** : Ajout de ~500-800 lignes pour une strategie inutilisee | Elevee (100%) | Moyen : Augmente la surface de maintenance et la complexite cognitive | Ne pas integrer. |
| **Bugs numeriques** : Erreurs dans l'implementation REDC avec grands entiers | Moyenne (40%) | Eleve : Resultats incorrects silencieux | Tests exhaustifs avec golden files. Mais effort gaspille si la strategie est rejetee. |
| **Faux sentiment de securite** : L'existence d'une strategie "Montgomery" pourrait donner l'impression d'une couverture algorithmique plus large | Faible (20%) | Faible : Confusion mineure | Documentation claire des cas d'usage. |
| **Conflit avec optimisations futures** : Montgomery pourrait compliquer l'ajout de NTT ou d'autres methodes | Moyenne (50%) | Moyen : Refactoring necessaire | Architecture modulaire (deja en place via Strategy pattern). |
| **Regression sur ARM/autres architectures** : Montgomery sans SIMD serait encore plus lent | Elevee (80%) | Moyen : Degradation sur plateformes non-x86 | Tests multi-architecture. |

## 7. Recommandation

### REJETER -- Montgomery n'est pas adapte au calcul de Fibonacci par entiers libres

**Justification principale** : La multiplication de Montgomery est concue exclusivement pour l'arithmetique **modulaire** avec un **modulus fixe**. Elle excelle dans les chaines de multiplications mod N (RSA, DH, ECDSA) ou les conversions vers/depuis Montgomery form s'amortissent sur des centaines d'operations. Le calcul de Fibonacci par Fast Doubling effectue des multiplications d'**entiers libres** (sans modulus), ce qui rend Montgomery :

1. **Inapplicable directement** : Il n'y a pas de modulus N dans F(2k) = F(k) * (2*F(k+1) - F(k)).
2. **Contre-productif avec modulus artificiel** : Introduire un modulus 2^B+1 assez grand pour contenir F(n) transformerait REDC en une no-op couteuse (le resultat ne depasse jamais le modulus).
3. **Inferieur aux optimisations existantes** : La reutilisation de transformees FFT dans `executeDoublingStepFFT()` est analogue a l'amortissement Montgomery mais plus efficace car elle elimine des transformations entieres plutot que de simples divisions.
4. **Non complementaire** : Montgomery ne peut pas servir de "tier intermediaire" entre Karatsuba et FFT car son avantage n'existe qu'en contexte modulaire. Pour le gap entre Karatsuba (~10K bits) et FFT (~500K bits), les alternatives pertinentes seraient Toom-Cook 3/4 ou NTT (Number Theoretic Transform).

**Cas d'usage valide pour Montgomery dans FibGo** : Si le projet ajoutait un mode `F(n) mod M` (Fibonacci modulaire), alors Montgomery serait pertinent. Ce mode n'est pas actuellement prevu.

**Alternatives recommandees pour le gap Karatsuba-FFT** :
- **Toom-Cook 3-way** : O(n^1.465), deja utilise en interne par `math/big` pour les tailles intermediaires
- **NTT (Number Theoretic Transform)** : Multiplication exacte via FFT sur corps fini, elimine les erreurs de precision flottante, potentiellement plus rapide que SS pour certaines tailles
- **Optimisation des seuils existants** : Le seuil FFT actuel (500K bits) pourrait etre abaisse apres calibration fine

### References

- Montgomery, P.L. "Modular multiplication without trial division." Mathematics of Computation 44.170 (1985): 519-521.
- Wikipedia: [Montgomery modular multiplication](https://en.wikipedia.org/wiki/Montgomery_modular_multiplication)
- Algorithmica: [Montgomery Multiplication](https://en.algorithmica.org/hpc/number-theory/montgomery/)
- CP-Algorithms: [Montgomery Multiplication](https://cp-algorithms.com/algebra/montgomery_multiplication.html)
- GMP Manual: [FFT Multiplication](https://gmplib.org/manual/FFT-Multiplication)
- Nayuki: [Fast Fibonacci algorithms](https://www.nayuki.io/page/fast-fibonacci-algorithms)
- OpenSSL: [BN_mod_mul_montgomery](https://docs.openssl.org/3.1/man3/BN_mod_mul_montgomery/)
- AVX-512 Montgomery: [Truncated batch SIMD AVX512 implementation](https://arxiv.org/abs/2410.18129)
- Springer: [Parallel modular multiplication using 512-bit AVX](https://link.springer.com/article/10.1007/s13389-021-00256-9)
