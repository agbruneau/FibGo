# Recommandation finale : Recherche exploratoire bigfft

## Decision : Scenario C - Optimisation incrementale

L'evaluation de trois alternatives au systeme de multiplication FFT actuel (Schonhage-Strassen / Fermat) dans `internal/bigfft/` conclut qu'**aucune alternative ne justifie un remplacement**, mais que des **optimisations significatives de l'implementation existante** sont realisables avec un excellent ratio effort/gain.

---

## Matrice de scores

| Critere | Poids | NTT | Montgomery | SS ameliore |
|---------|-------|-----|------------|-------------|
| Perf N moyen (10^6-10^7) | 25% | 1.5/5 | 1/5 | 3.5/5 |
| Perf N grand (10^8-10^10) | 20% | 3/5 | 1/5 | 3/5 |
| Memoire | 15% | 2.5/5 | 2/5 | 4/5 |
| Scalabilite | 10% | 4/5 | 1/5 | 3/5 |
| Effort integration | 15% | 1.5/5 | 2/5 | 4.5/5 |
| Maintenabilite | 10% | 2/5 | 2.5/5 | 4.5/5 |
| Risque regression | 5% | 2/5 | 1.5/5 | 4/5 |
| **Score pondere** | **100%** | **2.15** | **1.43** | **3.64** |

### Verdict par seuil

- **NTT : 2.15 < 2.5 -> Rejet** (pour implementation immediate ; approfondir a long terme pour F(10^9)+)
- **Montgomery : 1.43 < 2.5 -> Rejet** (inadapte a la multiplication non-modulaire)
- **SS ameliore : 3.64 > 3.5 -> Recommandation forte d'adoption**

---

## Justification

### 1. Montgomery -- REJETE (consensus unanime)

La multiplication de Montgomery est concue exclusivement pour l'arithmetique **modulaire** avec modulus fixe. Le calcul de Fibonacci par Fast Doubling effectue des multiplications d'**entiers libres** sans modulus, rendant Montgomery :
- **Inapplicable** : pas de modulus N dans `F(2k) = F(k) * (2*F(k+1) - F(k))`
- **Contre-productif** : un modulus artificiel transformerait REDC en overhead pur (~5-50% de ralentissement selon la taille)
- **Inferieur** : la reutilisation de transformees FFT dans `executeDoublingStepFFT()` est deja plus efficace que l'amortissement Montgomery

Le Devil's Advocate et le chercheur Montgomery sont en accord total. L'architecture Strategy pattern existante (`Multiplier`/`DoublingStepExecutor`) permet l'ajout futur de Montgomery si un mode `F(n) mod M` est ajoute.

### 2. NTT multi-primes -- REJETE pour implementation immediate, APPROFONDIR a long terme

La NTT multi-primes est theoriquement superieure pour les tres grands nombres (F(10^9)+) grace a la vectorisation SIMD des operations modulaires, mais :

- **Go pur = regression garantie** : 2-5x plus lent que la Fermat FFT actuelle pour toute la plage F(10^6)-F(10^8)
- **Overhead CRT significatif** : la reconstruction Chinese Remainder Theorem ajoute un facteur constant non-negligeable (K=3 multiplications 128-bit par coefficient)
- **Assembly SIMD requis** : 3-4 semaines d'effort en assembleur complexe (Barrett/Montgomery reduction vectorisee) pour un gain incertain
- **Crossover corrige** : apres revue adversariale, le crossover NTT/Fermat se situe a **F(10^9)** (et non F(10^8) comme initialement estime), et uniquement avec SIMD AVX2+

Le NTT researcher accepte ces corrections et recommande de prioriser les optimisations SS avant tout investissement NTT.

### 3. Schonhage-Strassen ameliore -- ADOPTER (optimisations incrementales)

L'implementation actuelle est mature et bien optimisee, mais plusieurs gains significatifs ont ete identifies :

**Decouverte majeure de la revue adversariale** : `big.Int.Mul` dans `fermat.Mul` ne detecte **PAS** le squaring car il compare les pointeurs `*Int` (`&xi != &yi`) et non les slices sous-jacentes. Le squaring pointwise n'est donc jamais optimise dans le chemin actuel. Corriger cela via `fermat.Sqr()` apporterait un gain de **8-12%** sur toutes les tailles -- c'est le quick win le plus impactant identifie.

---

## Quick Win transversal : optimisation du seuil FFT

Le Devil's Advocate a identifie un ecart significatif entre les deux seuils FFT :
- `bigfft.fftThreshold` = 1800 mots (~115K bits) : seuil interne ou la FFT est plus rapide que `math/big.Mul`
- `DefaultFFTThreshold` = 500,000 bits : seuil dans `smartMultiply()` ou FibGo bascule vers la FFT

L'ecart de ~385K bits signifie que des operandes entre 115K et 500K bits utilisent `math/big.Mul` alors que la FFT serait potentiellement plus rapide. **Abaisser ce seuil** apres calibration est le quick win au meilleur ROI (0.5 jour, potentiellement 2-5% de gain).

---

## Prochaines etapes

### Phase 0 -- Quick Win immediat (0.5 jour)
1. **Calibrer et ajuster `DefaultFFTThreshold`** (`constants.go:28`) : lancer `fibcalc --calibrate` et benchmarker avec des seuils de 150K, 200K, 300K, 400K bits. Abaisser si le benchmark le justifie.

### Phase 1 -- Quick Wins SS (3 jours, gain estime 8-14%)
1. **`fermat.Sqr()` specialise** (priorite #1, 1-2 jours) : Ajouter une methode `Sqr(x)` dans `fermat.go` qui passe `&xi, &xi` (meme pointeur) a `big.Int.Mul` pour n >= 30, et utilise `basicSqr` avec symetrie pour n < 30. Modifier `sqr()` dans `fft_poly.go` pour appeler `fermat.Sqr()`.
2. **Cache transformees agrandi** (0.5 jour) : Augmenter `MaxEntries` de 128 a 256, ajouter des metriques de hit rate.
3. **Seuils parallelisme** (conditionnel, 1 jour) : Benchmarker `MaxParallelFFTDepth=3,4,5` avec `NumCPU` et `2*NumCPU` tokens pour F(10^7). Ajuster seulement si gain > 2%.

### Phase 2 -- Optimisations structurelles SS (5 jours, gain supplementaire 3-7%)
4. **Twiddle per-invocation + fused butterfly** (3 jours) : Pre-calculer les offsets `i*omega2shift` par niveau de recursion (alloues depuis bump allocator), fusionner l'application du shift avec le papillon pour reduire les passes memoire.
5. **Localite cache Reconstruct** (2 jours) : Blocking de la boucle Reconstruct par blocs de 4-8 papillons pour maximiser les hits L1.

### Phase 3 -- Decision apres Phases 1-2 (si gain < 10%)
6. Si les gains des Phases 1-2 sont decevants (< 10%), envisager un **prototype NTT Go pur** (3-5 jours) pour calibrer le ratio reel NTT/Fermat et decider de la Phase SIMD.
7. Si les gains sont satisfaisants (>= 15%), documenter la decision et passer a d'autres optimisations.

### Non recommande
- **Harvey-van der Hoeven O(n log n)** : rejete unanimement. Non implementable en pratique (constantes astronomiques, crossover au-dela de 2^(2^30) bits).
- **Montgomery pour multiplication libre** : rejete. Inapplicable sans modulus.
- **NTT sans SIMD** : regression garantie, pas de valeur ajoutee.

---

## Fichiers d'action prioritaires

| Fichier | Modification | Phase | Priorite |
|---------|-------------|-------|----------|
| `internal/fibonacci/constants.go:28` | Ajuster `DefaultFFTThreshold` apres calibration | 0 | Critique |
| `internal/bigfft/fermat.go` | Ajouter `Sqr()`, `basicSqr()` | 1 | Critique |
| `internal/bigfft/fft_poly.go:384-389` | Modifier `sqr()` pour utiliser `fermat.Sqr()` | 1 | Critique |
| `internal/bigfft/fft_cache.go` | Augmenter `MaxEntries`, metriques hit rate | 1 | Haute |
| `internal/bigfft/fft_recursion.go` | Twiddle offsets pre-calcules, seuils configurables | 2 | Moyenne |
| `internal/bigfft/fft_core.go` | Integration twiddle pre-calcul | 2 | Moyenne |

---

## Livrables de la recherche

| Document | Contenu |
|----------|---------|
| `AgentsTeam/reports/NTT_Report.md` | Rapport complet NTT multi-primes |
| `AgentsTeam/reports/Montgomery_Report.md` | Rapport complet Montgomery |
| `AgentsTeam/reports/SS_Optimization_Report.md` | Rapport optimisations SS avec 8 propositions |
| `AgentsTeam/reports/Counter_Report.md` | Contre-rapport adversarial avec objections detaillees |
| `AgentsTeam/reports/NTT_Response.md` | Reponses NTT aux objections (estimations corrigees) |
| `AgentsTeam/reports/Montgomery_Response.md` | Reponses Montgomery (seuil FFT comme quick win) |
| `AgentsTeam/reports/SS_Response.md` | Reponses SS (decouverte squaring non-detecte) |
| `AgentsTeam/reports/Final_Recommendation.md` | Ce document |

---

*Recherche realisee par l'equipe bigfft-research : ntt-researcher, montgomery-researcher, ss-researcher, devils-advocate. Coordonnee par le leader.*
