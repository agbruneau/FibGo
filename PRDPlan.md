# Plan : Rédaction exhaustive du PRD.md pour le portage Rust de FibGo

## Livrable

Le contenu complet du plan de tâches sera écrit dans le fichier **`prdplan.md`** à la racine du projet (`C:\Users\agbru\OneDrive\Documents\GitHub\FibGo\prdplan.md`).

## Contexte

FibGo est un calculateur Fibonacci haute performance en Go (102 fichiers source, 85 fichiers test, 38 docs, 17 packages). Un premier PRD existe (`PRD-Claude 1.md`, 1280 lignes, 13 sections) mais reste insuffisant pour guider un développeur sans ambiguïté. Ce plan structure les tâches pour produire un PRD exhaustif couvrant chaque aspect du portage vers Rust.

### Lacunes identifiées dans le PRD existant

1. Pas de mapping fichier-par-fichier Go → Rust avec notes de migration spécifiques
2. Documentation algorithmique superficielle (formules sans détails d'itération bit-à-bit, séquences de swap de pointeurs, logique de parallélisation)
3. Détails FFT manquants (arithmétique Fermat, récursion, sélection de paramètres, polynômes)
4. Système Observer non spécifié (Freeze() lock-free, modèle géométrique de progression, pré-calcul puissances de 4)
5. Seuils dynamiques non détaillés (ring buffer, hystérésis, algorithme d'ajustement)
6. Pas de design alternatif pour Arena/GC Controller en Rust (puisque pas de GC)
7. Spécification TUI lacunaire (filtrage par génération, programRef, ring buffer sparklines, rendu Braille, layout adaptatif)
8. Mapping idiomatique Go→Rust superficiel (au-delà des correspondances de surface)
9. Pas de critères d'acceptation par fonctionnalité
10. Pas d'évaluation comparative des dépendances (num-bigint vs rug, ratatui vs alternatives)
11. Pas de plan de validation croisée Go/Rust
12. Pas de registre de risques par composant
13. Pas de diagrammes de flux de données pour la version Rust
14. Pas de spécification CLAUDE.md pour le projet Rust
15. Pas de stratégie de détection de régression de performance

---

## Phase 1 : Fondations & Analyse des exigences (12 tâches)

### T1.1 — Raffinement de la vision et critères de succès

- **Objectif** : Réécrire la Section 1 avec des critères de succès quantitatifs et vérifiables
- **Recherche** : PRD existant S1, CONTRIBUTING.md, README.md
- **Livrable** : Vision + matrice de critères (≥3 critères testables par objectif O1-O6)
- **Critère** : Chaque objectif a une méthode de vérification explicite

### T1.2 — Analyse des lacunes documentée

- **Objectif** : Formaliser les 15 lacunes identifiées avec priorisation
- **Recherche** : Comparaison PRD existant vs CLAUDE.md (14 patterns), code source
- **Livrable** : Tableau des lacunes (lacune × impact × priorité × section PRD cible)
- **Critère** : Toutes les lacunes ont une stratégie de résolution

### T1.3 — Guide approfondi des idiomes Rust

- **Objectif** : Traduire chaque pattern Go en Rust idiomatique avec exemples de code
- **Recherche** : 13 patterns du tableau existant S3.3 + patterns du CLAUDE.md
- **Livrable** : Guide de traduction avec snippets Go/Rust côte-à-côte pour chaque pattern
- **Critère** : Chaque pattern a un exemple avant/après + notes de performance

### T1.4 — Matrice d'évaluation des dépendances

- **Objectif** : Évaluer chaque crate Rust candidate avec critères objectifs
- **Recherche** : Docs de num-bigint, rug, ratatui, clap, rayon, crossbeam, bumpalo, etc.
- **Livrable** : Matrice décision (crate × performance × sécurité × ergonomie × maturité)
- **Critère** : Chaque dépendance a une recommandation justifiée + alternative documentée

### T1.5 — Registre de risques par composant

- **Objectif** : Identifier les risques de portage par module avec mitigation
- **Recherche** : Analyse de complexité des fichiers Go critiques (dynamic_threshold.go, fft_cache.go, etc.)
- **Livrable** : Registre (module × risque × probabilité × impact × mitigation)
- **Critère** : Les 7 crates Rust cibles ont chacune ≥2 risques identifiés avec mitigation

### T1.6 — Plan de validation croisée Go/Rust

- **Objectif** : Protocole pour garantir que Rust produit des résultats identiques à Go
- **Recherche** : fibonacci_golden.json, fibonacci_golden_test.go, fuzz targets
- **Livrable** : Protocole de validation (N ranges, algorithmes, golden files, diff automatique)
- **Critère** : Couvre N ∈ {0, 1, 93, 1K, 10K, 100K, 1M, 10M, 100M} × 3 algorithmes

### T1.7 — Baselines de performance et détection de régression

- **Objectif** : Documenter les performances Go de référence + seuils d'alerte Rust
- **Recherche** : Benchmarks Go, Docs/PERFORMANCE.md
- **Livrable** : Tableaux de baselines (algo × N × temps × mémoire) + critères de régression (<5%)
- **Critère** : 6 valeurs de N × 3 algorithmes avec temps et mémoire de référence

### T1.8 — Exigences non fonctionnelles

- **Objectif** : Spécifier les NFR au-delà de la parité fonctionnelle
- **Recherche** : Caractéristiques binaire Go (taille, startup), profils mémoire
- **Livrable** : Spécification NFR (startup <50ms, binaire <5MB stripped, peak RSS, etc.)
- **Critère** : ≥8 catégories NFR avec cibles quantitatives

### T1.9 — Matrice de cross-compilation détaillée

- **Objectif** : Instructions de build par plateforme avec dépendances et problèmes connus
- **Recherche** : Makefile build-all, arith_amd64.go, cpu_amd64.go
- **Livrable** : Guide par triple cible (5 triples) avec toolchain + troubleshooting
- **Critère** : Chaque triple a instructions + dépendances + problèmes connus

### T1.10 — Revue de compatibilité des licences

- **Objectif** : Vérifier la compatibilité des licences de toutes les dépendances Rust
- **Recherche** : Licences de num-bigint, rug (LGPL via GMP), ratatui, etc.
- **Livrable** : Matrice de licences + recommandation de licence projet
- **Critère** : Toutes les dépendances vérifiées + avis sur GMP/LGPL

### T1.11 — Timeline de migration raffinée

- **Objectif** : Estimer l'effort par phase avec chemin critique
- **Recherche** : LoC Go par package, multiplicateur Go→Rust estimé
- **Livrable** : Timeline avec dépendances inter-tâches et chemin critique
- **Critère** : 7 phases avec estimation de durée + graphe de dépendances

### T1.12 — Spécification CLAUDE.md pour le projet Rust

- **Objectif** : Concevoir le CLAUDE.md du projet Rust (commandes Cargo, architecture, conventions)
- **Recherche** : CLAUDE.md FibGo existant (500 lignes)
- **Livrable** : Contenu CLAUDE.md complet en annexe du PRD
- **Critère** : Miroir structurel du CLAUDE.md Go avec contenu Rust-spécifique

---

## Phase 2 : Spécifications algorithmiques détaillées (18 tâches)

### T2.1 — Fast Doubling : itération bit-à-bit et transitions d'état

- **Source** : `internal/fibonacci/doubling_framework.go` (ExecuteDoublingLoop)
- **Livrable** : Pseudocode avec itération MSB→LSB, mises à jour FK/FK1, swaps de pointeurs T1/T2/T3
- **Critère** : Diagramme de machine à états + pseudocode exécutable

### T2.2 — Fast Doubling : logique de parallélisation

- **Source** : `internal/fibonacci/doubling_framework.go` (executeDoublingStepMultiplications)
- **Livrable** : Arbre de décision parallèle/séquentiel + protocole de synchronisation
- **Critère** : Montre vérification du seuil, fork-join 3 voies, collecte d'erreurs

### T2.3 — Matrix Exponentiation : gestion d'état et pooling

- **Source** : `internal/fibonacci/matrix_framework.go`, `matrix_ops.go`, `matrix_types.go`
- **Livrable** : Cycle de vie matrixState + protocole de pooling + optimisation symétrique
- **Critère** : Diagramme d'état + séquence acquire/use/return

### T2.4 — Strassen : logique de basculement et fallback

- **Source** : `internal/fibonacci/matrix_ops.go` (shouldUseStrassen)
- **Livrable** : Critères d'activation Strassen + structure 7 multiplications + fallback
- **Critère** : Arbre de décision avec seuil en bits

### T2.5 — FFT : sélection des paramètres (k, n, modulus Fermat)

- **Source** : `internal/bigfft/fft.go`, `fft_core.go` (fftSize, chooseFFTParams)
- **Livrable** : Algorithme de sélection avec seuils de taille et contraintes
- **Critère** : Montre comment k et n sont choisis à partir de la longueur en bits

### T2.6 — Arithmétique de Fermat (Shift, Mul, Sqr, normalisation)

- **Source** : `internal/bigfft/fermat.go`
- **Livrable** : Sémantique de chaque opération Fermat avec formules + cas limites
- **Critère** : Chaque opération a formule mathématique + pseudocode + edge cases

### T2.7 — FFT : structure récursive et cas de base

- **Source** : `internal/bigfft/fft_recursion.go`
- **Livrable** : Arbre de récursion + seuil de base + opérations butterfly
- **Critère** : Montre profondeur, cas de base, nombre d'opérations butterfly

### T2.8 — Opérations polynomiales pour FFT

- **Source** : `internal/bigfft/fft_poly.go`
- **Livrable** : Spécification des opérations polynomiales (évaluation, extraction)
- **Critère** : Chaque fonction a définition mathématique + notes d'implémentation

### T2.9 — Réutilisation de transformée FFT pour le squaring

- **Source** : `internal/bigfft/fft.go` (sqrFFT vs mulFFT)
- **Livrable** : Optimisation squaring avec diagramme de flux
- **Critère** : Montre FFT unique + carré point-par-point + FFT inverse

### T2.10 — Stratégie adaptative : logique de sélection

- **Source** : `internal/fibonacci/strategy.go` (AdaptiveStrategy)
- **Livrable** : Algorithme de sélection avec comparaisons de seuils
- **Critère** : Flowchart avec vérification FFT threshold, taille opérande, fallback Karatsuba

### T2.11 — Fast Doubling modulaire (--last-digits)

- **Source** : `internal/fibonacci/modular.go`
- **Livrable** : Algorithme modulaire complet + preuve mémoire O(K)
- **Critère** : Opérations mod 10^K, preuve de correction, analyse mémoire

### T2.12 — Fast path itératif (n ≤ 93)

- **Source** : `internal/fibonacci/calculator.go` (fast path)
- **Livrable** : Spécification du chemin rapide u64
- **Critère** : Boucle itérative, prévention overflow, seuil de transition (93)

### T2.13 — Retour de résultat zero-copy

- **Source** : Pattern de "vol de pointeur" dans les frameworks
- **Livrable** : Technique de transfert de propriété + mapping std::mem::replace
- **Critère** : Séquence de swap, protocole de retour au pool, preuve zero-memcpy

### T2.14 — Méthodologie de comparaison inter-algorithmes

- **Source** : `internal/orchestration/orchestrator.go` (AnalyzeComparisonResults)
- **Livrable** : Protocole de comparaison avec détection de mismatch
- **Critère** : Comparaison bit-à-bit, génération d'erreur mismatch, tri par vitesse

### T2.15 — Générateur de séquence et optimisation Skip

- **Source** : `internal/fibonacci/generator_iterative.go`
- **Livrable** : Algorithme Skip() avec seuil de basculement
- **Critère** : Skip itératif (<1000) vs Calculator O(log n), mise à jour d'état

### T2.16 — Sélection de calculateur depuis la config

- **Source** : `internal/orchestration/calculator_selection.go`
- **Livrable** : Algorithme de sélection avec interaction factory
- **Critère** : Parsing "all"/"fast"/"matrix"/"fft", lookup factory, erreur si introuvable

### T2.17 — Preuves de correction algorithmique

- **Source** : `Docs/algorithms/FAST_DOUBLING.md`, `MATRIX.md`, `FFT.md`
- **Livrable** : Arguments de correction par algorithme avec identités
- **Critère** : Chaque algorithme a énoncé d'identité + esquisse de preuve

### T2.18 — Carte de couverture de tests par algorithme

- **Source** : `internal/fibonacci/*_test.go`
- **Livrable** : Matrice (composant × type de test : unit/golden/property/fuzz)
- **Critère** : Couverture documentée par composant algorithmique

---

## Phase 3 : Système Observer & suivi de progression (10 tâches)

### T3.1 — Architecture du pattern Observer

- **Source** : `internal/fibonacci/observer.go`
- **Livrable** : Diagramme UML + cycle de vie Register/Unregister/Notify + thread safety

### T3.2 — Mécanisme Freeze() pour snapshots lock-free

- **Source** : `internal/fibonacci/observer.go` (Freeze method)
- **Livrable** : Spécification Freeze avec sémantique de snapshot + analyse concurrence

### T3.3 — Modèle géométrique de travail (progression)

- **Source** : `Docs/algorithms/PROGRESS_BAR_ALGORITHM.md`
- **Livrable** : Formules de la série géométrique puissance-de-4 + exemples numériques

### T3.4 — Pré-calcul des puissances de 4

- **Source** : `internal/fibonacci/progress.go` (PrecomputePowers4)
- **Livrable** : Tableau global [64]float64 + optimisation zero-allocation

### T3.5 — Seuil de reporting de progression (1%)

- **Source** : `internal/fibonacci/progress.go` (ReportStepProgress)
- **Livrable** : Logique de déclenchement conditionnel avec forçage première/dernière itération

### T3.6 — ChannelObserver : pont vers l'UI

- **Source** : `internal/fibonacci/observers.go`
- **Livrable** : Envoi non-bloquant, capacité canal, pattern select/default

### T3.7 — LoggingObserver : throttling temporel

- **Source** : `internal/fibonacci/observers.go`
- **Livrable** : Throttling basé sur le temps, sélection de niveau de log

### T3.8 — NoOpObserver : null object pattern

- **Source** : `internal/fibonacci/observers.go`
- **Livrable** : Implémentation vide + scénarios d'utilisation (tests, quiet mode)

### T3.9 — Structure ProgressUpdate

- **Source** : Définition struct ProgressUpdate
- **Livrable** : Spécification de tous les champs avec sémantique

### T3.10 — Cycle de vie d'enregistrement des observers

- **Source** : `internal/fibonacci/calculator.go` (CalculateWithObservers)
- **Livrable** : Diagramme de séquence (enregistrement → notification → nettoyage)

---

## Phase 4 : Gestion mémoire & concurrence (12 tâches)

### T4.1 — CalculationArena : bump allocator pour états de calcul

- **Source** : `internal/fibonacci/arena.go`
- **Livrable** : Protocole d'allocation + mapping vers bumpalo en Rust

### T4.2 — GC Controller : stratégie Go et alternative Rust

- **Source** : `internal/fibonacci/gc_control.go`
- **Livrable** : Stratégie Go (disable GC + soft limit) + explication RAII Rust

### T4.3 — Pool d'objets BigInt avec classes de taille

- **Source** : `internal/bigfft/pool.go`
- **Livrable** : Cycle acquire/release, cap 100M bits, classes de taille (puissances de 4)

### T4.4 — Pré-chauffage des pools

- **Source** : `internal/bigfft/pool_warming.go`
- **Livrable** : Prédiction de taille, nombre de pré-allocations, seuil de déclenchement

### T4.5 — Bump allocator FFT

- **Source** : `internal/bigfft/bump.go`
- **Livrable** : Alloc O(1) par bump de pointeur + mapping bumpalo::Bump

### T4.6 — Validation du budget mémoire

- **Source** : `internal/fibonacci/memory_budget.go`
- **Livrable** : Formules d'estimation (state + FFT + cache + overhead) + logique de validation

### T4.7 — Estimation mémoire FFT

- **Source** : `internal/bigfft/memory_est.go`
- **Livrable** : Composantes mémoire FFT (transformée, temporaires, pool) + total

### T4.8 — Pooling d'état CalculationState/matrixState

- **Source** : `internal/fibonacci/common.go`, `matrix_ops.go`
- **Livrable** : Cycle de vie pool (acquire → reset → use → return)

### T4.9 — Modèle de concurrence complet

- **Source** : Orchestrator (errgroup), DoublingFramework (parallel), TUI (event loop)
- **Livrable** : Catalogue de 4 patterns + mapping Go→Rust (rayon, crossbeam, etc.)

### T4.10 — Sémaphore de tâches et exécution générique

- **Source** : `internal/fibonacci/common.go` (taskSemaphore, executeTasks[T,PT])
- **Livrable** : Dimensionnement sémaphore (2×NumCPU) + générique avec contrainte pointeur

### T4.11 — Pattern de collecte d'erreurs parallèles

- **Source** : `internal/parallel/errors.go`
- **Livrable** : ErrorCollector (SetError/Err) + sémantique first-error + atomiques

### T4.12 — Protocole d'annulation coopérative

- **Source** : context.Context checks dans les boucles algorithmiques
- **Livrable** : Points de vérification, propagation d'erreur, mapping Arc `<AtomicBool>`

---

## Phase 5 : Seuils dynamiques & calibration (8 tâches)

### T5.1 — Architecture DynamicThresholdManager

- **Source** : `internal/fibonacci/dynamic_threshold.go`
- **Livrable** : Ring buffer [20]IterationMetric, metricsHead, algorithme d'ajustement

### T5.2 — Mécanisme d'hystérésis

- **Source** : `internal/fibonacci/dynamic_threshold.go` (HysteresisMargin 0.15)
- **Livrable** : Marge 15%, condition d'ajustement, analyse de stabilité

### T5.3 — Collecte de métriques par itération

- **Source** : `internal/fibonacci/threshold_types.go`
- **Livrable** : Champs IterationMetric, protocole de timing, catégorisation FFT/parallèle

### T5.4 — Algorithme d'ajustement des seuils

- **Source** : `internal/fibonacci/dynamic_threshold.go` (AdjustThresholds)
- **Livrable** : Pseudocode avec vérification speedup FFT (1.2×) et parallèle (1.1×)

### T5.5 — Format du profil de calibration

- **Source** : `internal/calibration/profile.go`
- **Livrable** : Schéma JSON + règles de validation (CPU/arch/version)

### T5.6 — Modes de calibration (complet, auto, caché)

- **Source** : `internal/calibration/calibration.go`
- **Livrable** : Flowchart de la chaîne de fallback 3 niveaux

### T5.7 — Estimation adaptative des seuils

- **Source** : `internal/calibration/adaptive.go`
- **Livrable** : Formules f(cores) pour parallel, f(arch) pour FFT, f(CPU) pour Strassen

### T5.8 — Stratégie de micro-benchmarks

- **Source** : `internal/calibration/microbench.go`
- **Livrable** : Cas de test, méthodologie de timing, algorithme de sélection de seuil

---

## Phase 6 : Spécification détaillée du TUI (10 tâches)

### T6.1 — Architecture Elm du modèle Bubble Tea → ratatui

- **Source** : `internal/tui/model.go`
- **Livrable** : Structure Model (45 champs) + cycle Init/Update/View + mapping ratatui

### T6.2 — Filtrage des messages par génération

- **Source** : `internal/tui/model.go` (generation counter)
- **Livrable** : Incrément à chaque restart, filtrage des messages obsolètes

### T6.3 — Pattern programRef pour accès tea.Program

- **Source** : `internal/tui/bridge.go`
- **Livrable** : Champ programRef, setProgram, SendToProgram, thread safety

### T6.4 — Catalogue des types de messages TUI

- **Source** : `internal/tui/messages.go`
- **Livrable** : Tableau des 11 types (ProgressMsg, TickMsg, MemStatsMsg, etc.) avec champs

### T6.5 — Algorithme de layout 60/40 adaptatif

- **Source** : `internal/tui/model.go` (layoutPanels)
- **Livrable** : Calcul de tailles, gestion du resize, tailles minimales

### T6.6 — Ring buffer pour sparklines CPU/mémoire

- **Source** : `internal/tui/chart.go`, `sparkline.go`
- **Livrable** : Buffer circulaire fixe, pointeur head, wrapping, rendu sparkline

### T6.7 — Rendu Braille pour graphiques

- **Source** : `internal/tui/sparkline.go`
- **Livrable** : Mapping Unicode Braille (U+2800), calcul de hauteur, plot des valeurs

### T6.8 — Panneau Logs scrollable

- **Source** : `internal/tui/logs.go`
- **Livrable** : Viewport offset, limites de scroll, cap 10K entrées, auto-scroll

### T6.9 — Pipeline de collecte de métriques système

- **Source** : `internal/tui/metrics.go`, `internal/sysmon/sysmon.go`
- **Livrable** : Intégration sysinfo, intervalle 1s, types (RSS, CPU%, threads)

### T6.10 — Bridge TUI (ProgressReporter/ResultPresenter)

- **Source** : `internal/tui/bridge.go`
- **Livrable** : Conversion progress channel → messages ratatui, présentation résultats

---

## Phase 7 : Intégration, tests & finalisation (8 tâches)

### T7.1 — Mapping fichier-par-fichier Go → Rust

- **Livrable** : Tableau (100+ fichiers Go × crate Rust cible × fichier Rust × priorité × notes)
- **Critère** : Chaque fichier .go a une destination dans la structure Cargo workspace

### T7.2 — Contrats d'interface (traits) avec pré/postconditions

- **Livrable** : Spécification formelle par trait (Calculator, Multiplier, DoublingStepExecutor, ProgressObserver, etc.)
- **Critère** : Chaque méthode de trait a préconditions, postconditions, invariants

### T7.3 — Diagrammes de flux de données (Rust-spécifiques)

- **Livrable** : 5 DFD (CLI flow, TUI flow, orchestration, algorithme, FFT) avec frontières de crates et transferts d'ownership
- **Critère** : Chaque DFD montre les types concrets + les canaux + le cycle de vie des données

### T7.4 — Catalogue de cas limites

- **Livrable** : Tableau (composant × cas limite × traitement attendu) avec ≥50 cas
- **Critère** : Cas n=0, n=1, n=93, n=94, overflow, timeout, cancel, mismatch, etc.

### T7.5 — Carte de propagation des erreurs

- **Livrable** : Hiérarchie FibError + chemins de propagation + codes de sortie
- **Critère** : Chaque variant d'erreur a son chemin de propagation documenté

### T7.6 — Spécification de la frontière FFI (feature gmp/rug)

- **Livrable** : Blocs unsafe, preuves de sûreté, mapping build tags → features Cargo
- **Critère** : Chaque usage de rug::Integer documenté avec justification de safety

### T7.7 — Scénarios de tests d'intégration

- **Livrable** : ≥20 scénarios E2E avec setup, exécution, vérification, résultat attendu
- **Critère** : Couvre tous les modes (CLI, TUI, calibration, completion) et codes de sortie

### T7.8 — Structure de documentation du projet Rust

- **Livrable** : Arborescence docs/ avec contenu attendu par fichier
- **Critère** : Miroir de la structure FibGo (38 docs) adapté à Rust

---

## Résumé

| Phase           | Tâches      | Focus                                                         |
| --------------- | ------------ | ------------------------------------------------------------- |
| Phase 1         | 12           | Fondations, exigences, évaluation dépendances, risques      |
| Phase 2         | 18           | Algorithmes détaillés (Fast Doubling, Matrix, FFT, Modular) |
| Phase 3         | 10           | Observer, progression, modèle géométrique                  |
| Phase 4         | 12           | Mémoire (arena, pool, bump), concurrence, annulation         |
| Phase 5         | 8            | Seuils dynamiques, calibration, profils                       |
| Phase 6         | 10           | TUI (layout, messages, sparklines, bridge)                    |
| Phase 7         | 8            | Intégration (migration map, DFD, edge cases, tests)          |
| **Total** | **78** | **PRD estimé : 4000-5000 lignes**                      |

### Dépendances critiques

```
Phase 1 (fondations) ──→ Phase 2 (algorithmes) ──→ Phase 7.1 (migration map)
Phase 1 ──→ Phase 3 (observer) ──→ Phase 6.10 (TUI bridge)
Phase 1 ──→ Phase 4 (mémoire) ──→ Phase 7.2 (contrats)
Phase 2 ──→ Phase 5 (seuils dynamiques)
Phase 6 (TUI) ──→ Phase 7.3 (DFD)
```

### Fichiers Go critiques à lire en profondeur

- `internal/fibonacci/doubling_framework.go` — Boucle Fast Doubling, parallélisation, intégration progression
- `internal/fibonacci/observer.go` — Pattern Observer, mécanisme Freeze(), thread-safety
- `internal/bigfft/fft_core.go` — Moteur FFT, sélection paramètres, structure récursive
- `internal/bigfft/fermat.go` — Arithmétique Fermat (Shift, Mul, Sqr, normalisation)
- `internal/fibonacci/dynamic_threshold.go` — Ring buffer, hystérésis, ajustement runtime
- `internal/tui/model.go` — Architecture Elm, filtrage par génération, programRef
- `internal/bigfft/fft_cache.go` — Cache LRU, thread-safety, hashing FNV
- `internal/fibonacci/arena.go` — Bump allocator, pré-allocation contiguë
- `Docs/algorithms/PROGRESS_BAR_ALGORITHM.md` — Modèle géométrique, pré-calcul puissances de 4

### Approche de rédaction

Le PRD sera rédigé de manière incrémentale, phase par phase. Chaque tâche implique :

1. **Lecture** des fichiers source Go identifiés
2. **Extraction** de la logique, des constantes, des seuils, des algorithmes
3. **Traduction** en spécification Rust idiomatique avec exemples de code
4. **Rédaction** de la section PRD correspondante avec critères d'acceptation

Le document final sera un fichier `PRD.md` unique et autosuffisant, permettant à un développeur Rust de porter FibGo sans ambiguïté.
