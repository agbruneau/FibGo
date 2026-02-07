# Sommaire executif — Recherche exploratoire bigfft

**Projet** : FibGo — Evaluation d'alternatives au systeme de multiplication FFT
**Date** : 2026-02-07
**Equipe** : ntt-researcher, montgomery-researcher, ss-researcher, devils-advocate
**Methode** : Recherche comparative + revue adversariale (6 phases, 12 taches)

---

## Decision

**Scenario C — Optimisation incrementale de l'implementation Schonhage-Strassen existante.**

Aucune alternative ne justifie un remplacement du systeme actuel. Des optimisations significatives de l'implementation existante dans `internal/bigfft/` sont realisables avec un excellent ratio effort/gain.

---

## Scores

| Alternative           |   Score pondere   | Verdict                                                     |
| --------------------- | :----------------: | ----------------------------------------------------------- |
| NTT multi-primes      |      2.15 / 5      | Rejet (< 2.5) — approfondir a long terme pour F(10^9)+     |
| Montgomery            |      1.43 / 5      | Rejet (< 2.5) — inadapte a la multiplication non-modulaire |
| **SS ameliore** | **3.64 / 5** | **Adoption recommandee** (> 3.5)                      |

---

## Decouverte majeure

`big.Int.Mul` ne detecte **pas** le squaring dans `fermat.Mul` : la comparaison se fait sur les pointeurs `*Int` (`&xi != &yi`), pas sur les slices sous-jacentes. Le squaring pointwise n'est donc jamais optimise dans le chemin actuel. Corriger cela via `fermat.Sqr()` apporterait **8-12% de gain** sur toutes les tailles — c'est le quick win le plus impactant identifie.

---

## Plan d'action

| Phase       | Contenu                                                               | Effort              | Gain estime      |
| ----------- | --------------------------------------------------------------------- | ------------------- | ---------------- |
| **0** | Calibrer et abaisser `DefaultFFTThreshold` (ecart 115K–500K bits)  | 0.5 jour            | 2-5%             |
| **1** | `fermat.Sqr()` specialise + cache agrandi + seuils parallelisme     | 3 jours             | 8-14%            |
| **2** | Twiddle per-invocation + fused butterfly + localite cache Reconstruct | 5 jours             | +3-7%            |
|             | **Total Phases 0-2**                                            | **8.5 jours** | **10-18%** |

Phase 3 (conditionnelle) : sqrt(2) trick (10-15j, +5-15%) ou prototype NTT Go pur (3-5j) si gains Phases 1-2 < 10%.

---

## Fichiers prioritaires

| Fichier                                 | Modification                                      | Phase |
| --------------------------------------- | ------------------------------------------------- | :---: |
| `internal/fibonacci/constants.go:28`  | Ajuster `DefaultFFTThreshold` apres calibration |   0   |
| `internal/bigfft/fermat.go`           | Ajouter `Sqr()` / `basicSqr()`                |   1   |
| `internal/bigfft/fft_poly.go:384-389` | Appeler `fermat.Sqr()` dans `sqr()`           |   1   |
| `internal/bigfft/fft_cache.go`        | `MaxEntries` 128 → 256, metriques hit rate     |   1   |
| `internal/bigfft/fft_recursion.go`    | Twiddle offsets pre-calcules                      |   2   |
| `internal/bigfft/fft_core.go`         | Integration twiddle pre-calcul                    |   2   |

---

## Rejets

- **Montgomery** : Inapplicable sans modulus fixe. Consensus unanime (chercheur + adversaire).
- **NTT Go pur** : Regression garantie (2-5x plus lent). Crossover a F(10^9) uniquement avec SIMD AVX2+.
- **Harvey-van der Hoeven O(n log n)** : Constantes astronomiques, crossover au-dela de 2^(2^30) bits.

---

## Livrables

| # | Document                              | Statut |
| - | ------------------------------------- | :-----: |
| 1 | `reports/NTT_Report.md`             | Termine |
| 2 | `reports/Montgomery_Report.md`      | Termine |
| 3 | `reports/SS_Optimization_Report.md` | Termine |
| 4 | `reports/Counter_Report.md`         | Termine |
| 5 | `reports/NTT_Response.md`           | Termine |
| 6 | `reports/Montgomery_Response.md`    | Termine |
| 7 | `reports/SS_Response.md`            | Termine |
| 8 | `reports/Final_Recommendation.md`   | Termine |
