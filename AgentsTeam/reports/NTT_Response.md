# Reponse aux objections NTT -- ntt-researcher

## Resume

Sur les 5 objections du devils-advocate, j'accepte partiellement les 3 majeures et les 2 mineures. Elles renforcent ma recommandation initiale ("Approfondir avec prototype limite") plutot que de l'invalider. Les objections pointent des faiblesses reelles dans mes estimations que je corrige ci-dessous. La conclusion reste : la NTT multi-primes est une piste a long terme pour F(10^9)+, mais **pas une priorite immediate** par rapport aux optimisations SS (meilleur ROI).

---

## Reponses point par point

### Objection 1 : L'overhead CRT annule l'avantage SIMD pour la plage cible

**Verdict : Partiellement acceptee. Les estimations sont corrigees a la hausse.**

L'analyse du devils-advocate sur le cout CRT est correcte dans sa structure mais contient une erreur d'echelle. Clarifions :

1. **Cout CRT par coefficient** : Pour 3 primes 63-bit, la reconstruction CRT necessite :
   - 2 soustractions modulaires 64-bit (2 cycles chacune)
   - 2 multiplications 64-bit avec resultat 128-bit via `math/bits.Mul64` (~3-4 cycles chacune)
   - 1 reduction modulo p3 (~5 cycles)
   - Total : **~15-20 cycles par coefficient** pour 3 primes

   C'est effectivement ~15x plus cher qu'un simple shift Fermat (~1-2 cycles). Le devils-advocate a raison sur ce point.

2. **Mais l'echelle est sous-estimee dans l'objection** : Pour F(10^8), ce n'est pas K_FFT = 4096 coefficients qui comptent, mais le **nombre total de coefficients dans la FFT**. Avec k~10 (FFT size 2^10 = 1024), on a 1024 coefficients, pas 4096. Le cout CRT par iteration de doubling serait : 1024 * 15 cycles * 3 = ~46K cycles, soit ~15us a 3GHz. Sur 27 iterations, c'est ~405us total -- negligeable par rapport au temps total de ~1.2s.

3. **Le vrai cout CRT** n'est pas dans la reconstruction elle-meme mais dans le fait qu'on effectue **3 transforms forward + 3 transforms inverse au lieu de 1+1**. C'est un facteur multiplicatif de K=3 sur la partie la plus couteuse (les transforms eux-memes).

**Estimation corrigee** : Mon ratio NTT/Fermat de 1.25x pour F(10^8) etait effectivement **optimiste**. Estimation corrigee :

| N | Ancien ratio | Ratio corrige |
|---|-------------|---------------|
| 10^6 | 3.0x plus lent | 3.5x plus lent |
| 10^7 | 1.9x plus lent | 2.2x plus lent |
| 10^8 | 1.25x plus lent | 1.5-1.8x plus lent |
| 10^9 | 1.2x plus rapide | 1.0-1.1x (quasi egal) |
| 10^10 | 1.5x plus rapide | 1.2-1.3x plus rapide |

Le crossover se deplace vers F(10^9) plutot que F(10^8). Cela renforce la conclusion que **la NTT n'est pas rentable pour la plage F(10^6)-F(10^8)**.

**Critere de resolution** : Accepte. Un prototype benchmark Go mesurant le cout CRT reel serait la premiere etape du Sprint 1. Si le CRT coute > 3x le shift Fermat (ce qui semble probable), le crossover se confirmerait au-dela de F(10^9).

---

### Objection 2 : Perte du squaring optimise

**Verdict : Partiellement acceptee. Le cache est le vrai probleme, pas le squaring lui-meme.**

1. **Le squaring NTT reste optimise** : Avec la NTT multi-primes, le squaring economise exactement K transforms (au lieu de 2K transforms pour un mul). L'avantage relatif de 33% est preserve. Le devils-advocate le reconnait mais souleve le cout absolu -- ce point est deja couvert par l'Objection 1 (facteur multiplicatif K).

2. **Le cache de transformees est le vrai probleme** : L'objection sur le cache est valide et c'est un point que j'avais insuffisamment traite. Avec K=3 primes :
   - **Option A : K caches independants** (3 x 128 entrees) -- triple la memoire cache (~3x), mais simple a implementer
   - **Option B : Cache unifie** avec cles incluant l'indice du prime -- meme nombre d'entrees (128) mais chaque entree est 3x plus petite (1 prime au lieu du Fermat complet). L'empreinte memoire totale serait comparable.
   - **Option C : Cache par triplet** -- cacher le triplet (transform_p1, transform_p2, transform_p3) comme une seule entree. C'est l'approche la plus naturelle : le cache a 128 entrees, chacune 3x plus grande en memoire mais conceptuellement identique au cache Fermat actuel.

   **Recommandation** : Option C. La consommation memoire cache augmenterait d'environ 2x (les valeurs NTT sont des mots 64-bit simples vs les fermat qui sont des multi-mots), pas 3x. Le hit rate resterait identique car la logique d'acces est la meme.

3. **Interaction avec `executeDoublingStepFFT`** : L'objection mentionne `fft_cache.go:411-448` (SqrCachedWithBump). Dans l'implementation NTT, cette logique serait repliquee : transformer une seule fois, cacher le triplet, reutiliser pour squaring et multiplication. La structure de `executeDoublingStepFFT` (transformer FK, FK1, T4 une fois puis 3 operations pointwise) resterait identique -- seul le nombre de transforms par operande passe de 1 a K.

**Conclusion** : L'objection ne remet pas en cause la faisabilite mais ajoute ~2x de memoire cache. Cela reste gerable avec l'Option C.

---

### Objection 3 : Go pur = regression garantie, assembly = hors perimetre

**Verdict : Largement acceptee. C'est l'objection la plus forte.**

1. **Go pur est une regression garantie** : Je confirme. Mon rapport l'indiquait deja clairement ("probablement jamais rentable"). Le Sprint 1 (prototype Go pur) est un **outil de validation**, pas un deliverable. Son but est de :
   - Valider la correctness de l'implementation CRT
   - Mesurer le ratio reel NTT/Fermat en Go pur pour calibrer les estimations
   - Decider si le Sprint 3 (SIMD) vaut l'investissement

   **Critere d'abandon confirme** : Si le prototype Go pur est > 5x plus lent que Fermat FFT pour F(10^7), abandon. Si 3-5x, poursuite conditionee au Sprint 3.

2. **Assembly NTT : complexite reelle** : L'objection est valide. L'assembly NTT butterfly avec reduction modulaire Barrett/Montgomery est significativement plus complexe que `addVV`/`subVV`. Estimation de complexite :
   - **Butterfly AVX2** : ~200-300 lignes d'assembleur (4 multiplications 64x64->128 par butterfly, avec reduction Barrett)
   - **Mulmod SIMD** : ~100-150 lignes (4-wide Barrett reduction)
   - **Total** : ~400-600 lignes d'assembleur, vs ~50 lignes pour les wrappers `arith_amd64.go` actuels
   - **Maintenance** : Chaque changement de version Go pourrait casser les `go:linkname` existants ET le nouvel assembleur. Le cout de maintenance double effectivement.

3. **Proposition revisee** : Au lieu de custom assembly, explorer l'utilisation de **CGo + intrinsics C** pour la NTT SIMD. Avantages :
   - Les intrinsics `_mm256_*` sont stables et portables entre compilateurs
   - Le compilateur C optimise automatiquement le scheduling
   - Inconvenient : overhead CGo (~100ns par appel), mais amortissable si on batch les operations

   Alternative : utiliser `go:linkname` vers le package `internal/runtime` pour acceder aux SIMD intrinsics de Go 1.25+ (si disponibles).

4. **Plan de maintenance** : Le code assembly serait maintenu par quiconque maintient deja `arith_amd64.go` et `cpu_amd64.go`. Le pattern existant (build tags + fallback generique) serait replique.

**Conclusion** : L'objection est la plus forte car elle remet en question le ROI de la Phase 3. Mon estimation d'effort pour la Phase 3 est requalifiee de 2-3 semaines a **3-4 semaines** (incluant la complexite de la reduction modulaire SIMD).

---

### Objection 4 (mineure) : Deux backends FFT

**Verdict : Acceptee avec mitigation.**

L'objection est valide. Maintenir deux backends est un cout reel. Cependant :

1. **Interface FFTBackend** : La proposition d'une interface commune est excellente et realisable. Les methodes seraient :
   ```
   type FFTBackend interface {
       ForwardTransform(input []nat, k uint, m int) (TransformedValues, error)
       PointwiseMul(a, b TransformedValues) (TransformedValues, error)
       PointwiseSqr(a TransformedValues) (TransformedValues, error)
       InverseTransform(v TransformedValues) (Poly, error)
   }
   ```
   Avec `TransformedValues` etant un type opaque (PolValues Fermat ou [K]NTTValues).

2. **Indirection sur le hot path** : En Go, un appel d'interface ajoute ~2ns d'overhead (vtable lookup). Pour des operations qui prennent des microsecondes a millisecondes, c'est negligeable. L'interface serait utilisee au niveau `fftmulTo`/`fftsqrTo`, pas au niveau butterfly individuel.

3. **Mitigation** : Si la NTT n'atteint pas les criteres de succes apres le Sprint 2, elle serait **supprimee** du code, pas maintenue indefiniment. Le dual-backend n'existerait que pendant la phase d'evaluation.

---

### Objection 5 (mineure) : Estimations basees sur y-cruncher

**Verdict : Acceptee integralement.**

L'objection est entierement valide. Mes estimations extrapolent des ratios C++ SIMD vers Go, ce qui est methodologiquement faible. Corrections :

1. **Toutes les estimations de performance dans mon rapport** devraient etre etiquetees comme "extrapolees de y-cruncher/GMP, a valider par prototype Go".

2. **Facteurs specifiques a Go non pris en compte** :
   - GC pauses : non-deterministes, peuvent ajouter ~1-5% d'overhead sur les calculs longs
   - Goroutine scheduling : ~300 bytes/goroutine + overhead scheduler vs threads natifs
   - Compilateur Go : genere du code ~1.5-2x plus lent que GCC -O3 pour le code numerique (pas d'auto-vectorisation, pas de SLP vectorization)
   - `math/bits.Mul64` : compile en une seule instruction `MULQ` sur x86-64, donc le penalty Go est faible pour l'arithmetique scalaire

3. **La seule validation fiable** est un prototype Go benchmark. C'est exactement ce que propose le Sprint 1.

---

## Reponses aux questions directes

### Q1 : Cout CRT par coefficient en cycles CPU pour 3 primes 63-bit en Go pur ?

Estimation theorique : **15-20 cycles par coefficient**.

Decomposition :
- `math/bits.Mul64(a, b)` -> instruction MULQ -> 3 cycles latence, 1 cycle throughput sur x86-64
- `math/bits.Add64(a, b, carry)` -> ADCQ -> 1 cycle
- Reconstruction CRT Garner pour 3 primes : 2 multiplications 64-bit, 2 additions 128-bit, 1 reduction modulo -> ~15 cycles

Je n'ai pas de prototype. Le Sprint 1 (3 jours) produirait ce benchmark comme premiere deliverable.

### Q2 : Cache de transformees avec K primes ?

**Cache unifie par triplet** (Option C decrite dans la reponse a l'Objection 2) :
- Meme nombre d'entrees que le cache Fermat actuel (128)
- Chaque entree contient le triplet (NTT_p1, NTT_p2, NTT_p3)
- Meme logique LRU, meme cle de cache (hash de l'input)
- Memoire : ~2x le cache Fermat actuel (les valeurs NTT sont des mots 64-bit simples vs multi-mots Fermat, mais il y en a K=3 sets)

### Q3 : Pourquoi investir 2 semaines en Phase 1 Go pur si c'est garanti plus lent ?

Les 2 semaines du Sprint 1 ne sont pas un investissement a perte. Elles servent a :

1. **Valider la correctness** : L'implementation CRT est sujette a des bugs subtils. Un prototype Go pur permet de valider contre les resultats Fermat FFT existants (golden files, fuzz tests).

2. **Mesurer le ratio reel** : Si le prototype Go pur est 2x plus lent (au lieu de 3-5x estime), le SIMD n'a besoin que d'un gain de 2x pour etre rentable (faisable avec AVX2). Si c'est 10x plus lent, meme AVX-512 ne suffirait pas.

3. **Derisquer la Phase 3** : Sans la Phase 1, on investirait directement 3-4 semaines en assembly SIMD sans savoir si l'algorithme est correct ou si les ratios sont realistes.

**Cependant**, j'accepte l'objection implicite : 2 semaines est trop long pour un prototype de validation. **Revision** : Sprint 1 reduit a **3-5 jours** avec scope minimal :
- 1 seul prime (pas de CRT) pour valider le butterfly NTT
- Benchmark NTT 1-prime vs Fermat FFT pour calibrer le cout de base
- CRT ajoute en Sprint 2 seulement si le butterfly NTT est < 3x plus lent que le butterfly Fermat

---

## Estimations revisees

A la lumiere des objections, voici les estimations corrigees :

### Performance

| N | Ratio NTT/Fermat (Go pur) | Ratio NTT/Fermat (AVX2) | Ratio NTT/Fermat (AVX-512) |
|---|--------------------------|------------------------|---------------------------|
| 10^6 | 4-5x plus lent | 2-3x plus lent | 1.5-2x plus lent |
| 10^7 | 3-4x plus lent | 1.5-2x plus lent | 1-1.5x plus lent |
| 10^8 | 2.5-3x plus lent | 1.2-1.5x plus lent | 0.8-1.2x (quasi egal) |
| 10^9 | 2-2.5x plus lent | 0.8-1.0x (quasi egal) | 0.6-0.8x plus rapide |
| 10^10 | 1.5-2x plus lent | 0.6-0.8x plus rapide | 0.5-0.7x plus rapide |

### Effort

| Phase | Ancien estime | Estime revise |
|-------|--------------|---------------|
| Sprint 1 : Prototype NTT Go pur | 2 semaines | **3-5 jours** (scope reduit) |
| Sprint 2 : NTT complet + CRT | - | **5-7 jours** (si Sprint 1 prometteur) |
| Sprint 3 : SIMD | 2-3 semaines | **3-4 semaines** (complexite Barrett) |
| Sprint 4 : Tuning | 1 semaine | 1 semaine |
| **Total** | **4-7 semaines** | **5-7 semaines** |

### Recommandation finale revisee

**Recommandation maintenue : Approfondir**, mais avec un scope reduit et des criteres d'abandon plus stricts :

1. Les optimisations SS (rapport ss-researcher) devraient etre **prioritaires** -- meilleur ROI immediat (8-15% pour ~8 jours de travail)
2. La NTT est un **investissement a long terme** pour F(10^9)+, a commencer seulement apres les quick wins SS
3. Sprint 1 NTT reduit a 3-5 jours comme outil de validation/decision
4. Decision continue/abandon basee sur les ratios mesures du Sprint 1
