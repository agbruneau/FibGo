# Rapport : Number Theoretic Transform (NTT)

## Resume executif

L'implementation actuelle de FibGo utilise une FFT de type Schonhage-Strassen operant dans l'anneau Z/(2^n+1)Z (arithmetique de Fermat). Le remplacement par une NTT multi-primes (Small Primes NTT) offrirait des calculs **exacts sans erreur d'arrondi** et une meilleure compatibilite avec les architectures SIMD modernes (AVX2/AVX-512), mais introduirait un **overhead significatif** lie a la reconstruction CRT et aux multiplications modulaires. **Recommandation : Approfondir** -- la NTT multi-primes est une alternative viable a long terme pour les tres grands nombres (F(10^9)+), mais l'effort d'integration est substantiel (~4-6 semaines) et le gain n'est pas garanti pour la plage F(10^6)-F(10^8) ou l'implementation Fermat actuelle est deja bien optimisee.

## 1. Description de l'approche

### Principe algorithmique

La Number Theoretic Transform (NTT) est l'analogue de la DFT (Discrete Fourier Transform) dans un corps fini Z/pZ au lieu de C. Elle permet des convolutions exactes (sans erreurs d'arrondi) sur des sequences d'entiers.

Pour la multiplication de grands entiers, l'approche **Small Primes NTT** (multi-primes) fonctionne ainsi :

1. **Decomposition** : L'entier est decompose en coefficients de base B (typiquement B = 2^30 ou 2^62)
2. **Transformation** : On effectue K NTT independantes modulo K primes p_1, ..., p_K choisis de la forme p_i = k_i * 2^m + 1 (NTT-friendly)
3. **Multiplication pointwise** : Multiplication element par element dans chaque domaine transforme
4. **Transformation inverse** : K NTT inverses
5. **Reconstruction CRT** : Le Chinese Remainder Theorem reconstruit le resultat exact a partir des K residus

### Complexite theorique

- **NTT forward/inverse** : O(N log N) operations modulaires par prime
- **Multiplication pointwise** : O(N) multiplications modulaires par prime
- **CRT reconstruction** : O(N * K) operations sur mots etendus
- **Complexite totale** : O(K * N log N) operations modulaires, soit O(N log N log log N) comme Schonhage-Strassen

### Variantes

| Variante | Description | Avantages |
|----------|-------------|-----------|
| **Small Primes NTT (32-bit)** | 3-9 primes ~31 bits | Rapide, limite a ~10^9 bits |
| **Small Primes NTT (64-bit)** | 3-9 primes ~63 bits | Plus lent, pas de limite pratique |
| **Hybrid NTT** (y-cruncher 2008) | Fermat + NTT via CRT | Compromis memoire/vitesse, obsolete depuis 2013 |
| **Generalized Fermat Primes** | Primes r^(2^lambda)+1 | Base theorique du O(n log n) de Harvey-van der Hoeven |

### Primes NTT-friendly typiques

Pour une NTT 32-bit avec transform length 2^27 :
- p1 = 2^32 - 2^20 + 1 = 4293918721 (operations modulaires efficaces via shifts)
- p2 = 2^32 - 2^21 + 1 = 4291821569
- p3 = 2^32 - 2^22 + 1 = 4287627265

Pour une NTT 64-bit :
- p = 18446744069414584321 (= Phi_6(2^32), proche de 2^64)
- Primes de la forme k*2^m + 1 avec m >= transform_size

## 2. Analyse du code actuel

L'implementation actuelle de FibGo est une FFT de type Schonhage-Strassen pure, operant dans Z/(2^(w*W)+1)Z ou W est la taille de mot machine. C'est fondamentalement une **NTT dans un anneau de Fermat** plutot qu'une NTT dans un corps premier.

### Points forts du code actuel

1. **Racines de l'unite gratuites** : Dans Z/(2^n+1)Z, les puissances de 2 sont des racines de l'unite. Le "twiddle factor" est un simple shift (fermat.Shift), ce qui est bien plus rapide qu'une multiplication modulaire
2. **Pas de CRT** : Un seul anneau suffit, pas besoin de reconstruction multi-primes
3. **Bump allocator** : L'allocateur O(1) minimise la pression GC avec excellente localite cache
4. **Transform cache** : Cache LRU thread-safe pour reutiliser les transformations FFT (15-30% speedup)
5. **Parallelisme** : Recursion FFT parallelisee avec semaphore et seuil adaptatif
6. **Squaring optimise** : Une seule transformation pour le carre (33% d'economie)

### Points faibles motivant l'exploration NTT

1. **Arithmetic Fermat couteuse** : `fermat.Mul()` appelle `big.Int.Mul` pour les coefficients > 30 mots (fermat.go:149-204), ce qui est sous-optimal pour des coefficients de taille intermediaire
2. **Overflow handling complexe** : La normalisation `fermat.norm()` et la reduction modulo 2^n+1 ajoutent de la complexite et des branches conditionnelles
3. **ShiftHalf non-trivial** : La multiplication par sqrt(2) mod 2^n+1 (fermat.go:103-118) necessite deux shifts et une soustraction
4. **Pas d'exploitation SIMD** : Les operations fermat sont scalaires; une NTT avec primes < 2^31 permettrait d'utiliser directement les instructions SIMD (8 multiplications 32-bit en parallele avec AVX2)
5. **Coefficient size croissant** : Pour les tres grands N, les coefficients Fermat deviennent eux-memes tres grands, necessitant des multiplications recursives couteuses

### Fichiers analyses

| Fichier | Lignes | Observations cles |
|---------|--------|-------------------|
| `internal/bigfft/fermat.go` | 218 | Type `fermat` = `nat` (alias []big.Word). Arithmetique modulo 2^(w*W)+1. Shift par puissances de 2 pour les twiddle factors. Mul delegue a `big.Int.Mul` au-dessus de 30 mots (smallMulThreshold). Norm() effectue reduction avec branches conditionnelles. |
| `internal/bigfft/fft_core.go` | 104 | Point d'entree fourier() et variantes (fourierWithState, fourierWithBump). Alloue tmp/tmp2 fermats depuis pool ou bump allocator. Delegue a fourierRecursive. |
| `internal/bigfft/fft_recursion.go` | 139 | Recursion FFT unifiee (fourierRecursiveUnified). Parallelisme controle par semaphore (NumCPU tokens), seuil k>=4, profondeur max 3. Reconstruction butterfly : ShiftHalf + Add/Sub. Pool allocator pour goroutines paralleles (thread-safety). |
| `internal/bigfft/fft_poly.go` | 416 | Couche polynomiale : Poly (coefficients nat) et PolValues (valeurs transformees fermat). Transform/InvTransform, Mul/Sqr pointwise. MulCached/SqrCached avec cache LRU. Clone() pour usage concurrent. IntTo() avec reutilisation de buffer. |
| `internal/bigfft/fft.go` | 225 | API publique : Mul, MulTo, Sqr, SqrTo. Seuil FFT = 1800 mots (~115kbits). fftSize() et fftSizeThreshold[] pour selection des parametres k,m. valueSize() pour taille des coefficients. |
| `internal/fibonacci/fft.go` | 239 | Integration : mulFFT/sqrFFT delegue a bigfft. smartMultiply/smartSquare avec seuil configurable. executeDoublingStepFFT() : transforme FK, FK1, T4 une seule fois puis 3 multiplications pointwise en parallele. |

## 3. Comparaison theorique

### Complexite

| Operation | Actuel (Fermat FFT / SS) | Alternative (NTT multi-primes) | Avantage |
|-----------|--------------------------|-------------------------------|----------|
| Forward transform | O(N log N) shifts + adds | O(K * N log N) mulmods | **Actuel** : shifts >> mulmods |
| Twiddle factors | O(1) : simple shift binaire | O(1) : precomputed mulmod | **Actuel** : shift = 1-2 cycles vs mulmod = 3-5 cycles |
| Pointwise multiply | O(N) fermat.Mul (cher pour gros coefficients) | O(K * N) mulmods 32/64-bit | **NTT** pour gros N : mulmods restent O(1) |
| Inverse transform | O(N log N) shifts + adds | O(K * N log N) mulmods | **Actuel** |
| Reconstruction | Aucune (un seul anneau) | O(N * K) CRT reconstruction | **Actuel** : pas d'overhead CRT |
| Memoire | ~2*K*(n+1) mots | ~K*N*K_primes + CRT buffers | **Actuel** pour K petit |
| Parallelisme SIMD | Non exploite (scalaire) | 8x32-bit ou 4x64-bit AVX2 | **NTT** : facteur 4-8x potentiel |
| Scalabilite (>10^8 bits) | Coefficients recursifs | Mulmods constants | **NTT** pour tres grands N |

### Constantes cachees

Le diable est dans les constantes. Les analyses de y-cruncher (reference dans le domaine) montrent :

1. **FFT flottante** est ~10x plus rapide que NTT et SSA pour les tailles moyennes (cache-resident)
2. **NTT** est ~3-4x plus econome en memoire que FFT flottante
3. **Fermat FFT** (approche actuelle de FibGo) se situe entre les deux : shifts rapides mais coefficients croissants

Pour FibGo specifiquement :
- Le **shift gratuit** de l'approche Fermat economise un facteur ~2-3x sur les twiddle factors par rapport a la NTT
- Mais la **multiplication Fermat** (fermat.Mul) devient couteuse quand les coefficients depassent le cache L1
- La **NTT avec SIMD** pourrait recuperer le facteur 2-3x perdu sur les twiddles grace au parallelisme vectoriel

Le **crossover estime** ou la NTT multi-primes depasse la Fermat FFT se situe autour de **50-100 millions de bits** de resultat, soit environ F(7*10^7) a F(1.5*10^8).

## 4. Estimation de performance

### Donnees de reference FibGo

- F(10^6) : ~694,241 bits (~10,848 mots 64-bit)
- F(10^7) : ~6,942,419 bits (~108,475 mots 64-bit)
- F(10^8) : ~69,424,191 bits (~1,084,753 mots 64-bit)
- F(10^9) : ~694,241,913 bits (~10,847,530 mots 64-bit)
- F(10^10) : ~6,942,419,134 bits (~108,475,299 mots 64-bit)

### Estimation par taille

Les estimations ci-dessous sont basees sur l'analyse des complexites et les ratios observes dans y-cruncher et GMP :

| N | Bits F(n) | Actuel Fermat FFT (estime) | NTT 64-bit 3-primes (estime) | Ratio NTT/Fermat |
|---|-----------|---------------------------|-------------------------------|------------------|
| 10^6 | 694K | ~5ms | ~15ms | 3.0x plus lent |
| 10^7 | 6.9M | ~80ms | ~150ms | 1.9x plus lent |
| 10^8 | 69M | ~1.2s | ~1.5s | 1.25x plus lent |
| 10^9 | 694M | ~18s | ~15s | **1.2x plus rapide** |
| 10^10 | 6.9G | ~300s | ~200s | **1.5x plus rapide** |

**Notes importantes** :
- Ces estimations supposent une NTT AVX2-optimisee avec 3 primes 63-bit
- Sans SIMD, la NTT serait ~2-3x plus lente que les chiffres ci-dessus
- L'implementation Go pure (sans assembly) ajouterait ~30-50% d'overhead supplementaire
- Les temps incluent le Fast Doubling complet (log2(N) iterations de doubling)

### Seuil de crossover

Le crossover NTT vs Fermat FFT depend fortement de l'implementation :

| Scenario | Crossover estime |
|----------|-----------------|
| NTT Go pur (sans SIMD) | ~F(5*10^9) -- probablement jamais rentable |
| NTT Go + assembly AVX2 | ~F(10^8) a F(10^9) |
| NTT Go + assembly AVX-512 | ~F(5*10^7) a F(10^8) |
| Optimisation Fermat actuelle + SIMD | Repousse le crossover vers ~F(10^9)+ |

**Conclusion** : Pour la plage cible typique de FibGo (F(10^6) a F(10^8)), la Fermat FFT actuelle est probablement plus rapide. La NTT ne devient avantageuse que pour les calculs extremes (F(10^9)+) et **seulement avec une implementation SIMD dediee**.

## 5. Effort d'integration

### Fichiers a modifier

| Fichier | Type de modification | Effort estime |
|---------|---------------------|---------------|
| `internal/bigfft/ntt.go` (nouveau) | Nouveau : type nttPrime, operations modulaires (mulmod, addmod, submod), racines de l'unite precomputed | 3-4 jours |
| `internal/bigfft/ntt_primes.go` (nouveau) | Nouveau : selection et validation de primes NTT-friendly, table de primes precomputed | 1-2 jours |
| `internal/bigfft/ntt_transform.go` (nouveau) | Nouveau : forward/inverse NTT, butterfly Cooley-Tukey et Gentleman-Sande, parallelisme | 3-4 jours |
| `internal/bigfft/ntt_crt.go` (nouveau) | Nouveau : reconstruction CRT multi-primes, arithmetique etendue | 2-3 jours |
| `internal/bigfft/ntt_amd64.s` (nouveau) | Nouveau : assembly AVX2/AVX-512 pour butterfly NTT et mulmod | 5-7 jours |
| `internal/bigfft/fft.go` | Modifier : ajouter seuil de selection NTT vs Fermat FFT | 0.5 jour |
| `internal/bigfft/fft_poly.go` | Modifier : adapter Poly pour supporter les deux backends | 1-2 jours |
| `internal/fibonacci/fft.go` | Modifier : ajouter option NTT dans smartMultiply/smartSquare | 0.5 jour |
| `internal/fibonacci/options.go` | Modifier : ajouter NTTThreshold, NTTPrimeCount | 0.5 jour |
| `internal/fibonacci/constants.go` | Modifier : ajouter defaults NTT | 0.5 jour |
| `internal/bigfft/ntt_test.go` (nouveau) | Nouveau : tests unitaires, golden file, fuzz, property-based | 2-3 jours |
| `internal/bigfft/ntt_bench_test.go` (nouveau) | Nouveau : benchmarks comparatifs NTT vs Fermat FFT | 1 jour |
| `internal/calibration/` | Modifier : calibration NTT threshold | 1 jour |

### Effort total estime

| Phase | Effort | Description |
|-------|--------|-------------|
| Phase 1 : NTT core (Go pur) | 2 semaines | Arithmetique modulaire, transform, CRT, tests |
| Phase 2 : Integration | 1 semaine | Branchement dans le pipeline existant, options, calibration |
| Phase 3 : SIMD (optionnel) | 2-3 semaines | Assembly AVX2/AVX-512 pour les hot paths |
| Phase 4 : Tuning | 1 semaine | Benchmarks comparatifs, selection de seuils, profiling |
| **Total sans SIMD** | **~4 semaines** | Fonctionnel mais probablement plus lent que Fermat FFT |
| **Total avec SIMD** | **~6-7 semaines** | Potentiellement plus rapide pour F(10^8)+ |

## 6. Risques

| Risque | Probabilite | Impact | Mitigation |
|--------|-------------|--------|------------|
| **Performance inferieure sans SIMD** : La NTT Go pur pourrait etre 2-3x plus lente que la Fermat FFT actuelle pour la plage F(10^6)-F(10^8) | Elevee | Eleve | Benchmarker tot avec prototype minimal avant investissement complet. Ne pas remplacer Fermat FFT mais ajouter NTT comme option |
| **Complexite CRT** : La reconstruction CRT avec 3+ primes 63-bit necessite de l'arithmetique 128-bit/192-bit, non native en Go | Moyenne | Moyen | Utiliser `math/bits.Mul64` et `math/bits.Add64` pour l'arithmetique etendue. Alternative : primes 31-bit avec plus de primes |
| **Regression pour petits N** : La NTT a plus d'overhead fixe (precomputation twiddle, K transforms au lieu de 1) | Elevee | Faible | Seuil de crossover configurable, la Fermat FFT reste le defaut pour petits N |
| **Maintenance accrue** : Deux backends FFT a maintenir (Fermat + NTT) augmente la charge de maintenance | Moyenne | Moyen | Interface TempAllocator et strategy pattern existants facilitent l'abstraction. Tests golden file partagés |
| **Assembly non-portable** : Le code SIMD est specifique x86-64 et necessite maintenance separee | Moyenne | Moyen | Fallback Go pur obligatoire (arith_generic.go pattern existant). Build tags |
| **Primes insuffisants** : Pour les tres grands N (>10^18 bits), le nombre de primes disponibles peut limiter la NTT 32-bit | Faible | Faible | NTT 64-bit comme fallback. En pratique, F(10^10) = ~7*10^9 bits est bien en dessous de la limite |
| **Correctness bugs** : La reconstruction CRT et l'arithmetique modulaire sont sujettes a des bugs subtils (off-by-one, overflow) | Moyenne | Eleve | Tests exhaustifs : golden file, fuzz testing, comparaison avec Fermat FFT, property-based testing avec gopter |

## 7. Recommandation

**Recommandation : Approfondir** avec un prototype limite avant engagement complet.

### Justification

1. **La Fermat FFT actuelle est bien optimisee** pour la plage F(10^6)-F(10^8) : bump allocator, transform cache, squaring optimise, parallelisme. Un remplacement pur ne serait probablement pas plus rapide sans SIMD.

2. **Le potentiel existe pour les tres grands N** : La NTT multi-primes avec SIMD est l'approche dominante dans les logiciels de pointe (y-cruncher post-2013, GMP pour les tailles au-dela de Toom). Pour F(10^9)+, elle serait vraisemblablement superieure.

3. **L'effort est significatif** : 4-7 semaines pour une implementation complete, dont la moitie en assembly SIMD. Sans SIMD, le ROI est incertain.

4. **Comparaison avec GMP** : GMP utilise deja une variante Schonhage-Strassen (Fermat FFT) similaire a FibGo. Le passage a la NTT n'est pertinent que si l'on vise des performances superieures a GMP via SIMD, ce qui est un objectif ambitieux.

### Plan d'action propose

1. **Sprint 1 (3 jours)** : Prototype NTT Go pur minimal (1 prime, pas de parallelisme) et benchmark contre Fermat FFT actuelle pour F(10^7) et F(10^8). Si la NTT est > 5x plus lente, abandon.

2. **Sprint 2 (5 jours)** : Si le prototype est prometteur, implementation complete avec 3 primes, CRT, parallelisme. Benchmark comparatif complet.

3. **Sprint 3 (2-3 semaines)** : Si les benchmarks montrent un crossover < F(10^8), implementation SIMD AVX2 des hot paths (butterfly, mulmod).

4. **Decision finale** : Basee sur les benchmarks du Sprint 2/3. Critere de succes : NTT plus rapide que Fermat FFT pour F(10^8) avec au plus 20% d'overhead memoire supplementaire.

### Alternative a considerer

Avant d'investir dans la NTT, il serait peut-etre plus rentable d'**optimiser la Fermat FFT actuelle avec SIMD** (vectoriser fermat.Shift, fermat.Add, fermat.Sub). Cela apporterait un gain immediat de 2-4x sur les hot paths pour un effort bien moindre (~1-2 semaines).
