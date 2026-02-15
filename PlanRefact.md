# Plan de Refactorisation Exhaustif — FibGo

**Version** : 1.0
**Date** : 2026-02-14
**Auteur** : Généré par Claude Code (équipe de 5 agents d'audit parallèles)
**Approche** : Top-Down (Architecture d'abord)
**Statut** : Complet — 65+ problèmes identifiés, 6 phases, 42 tâches — **TOUTES COMPLÉTÉES**

---

## Tableau de Suivi des Tâches

> **INSTRUCTION** : Après chaque tâche complétée avec succès (build + tests passent), mettre à jour ce tableau en changeant le statut à `FAIT` et en inscrivant la date de complétion. Ne jamais marquer une tâche `FAIT` si `go test -race ./...` échoue.

### Progression globale : 42 / 42 tâches complétées

| ID    | Tâche                                                        | Phase | Sévérité | Statut       | Date       |
|-------|--------------------------------------------------------------|-------|----------|--------------|------------|
| T1.1  | Supprimer l'alias déprécié `MultiplicationStrategy`          | P1    | Critique | `FAIT`       | 2026-02-14 |
| T1.2  | Extraire `internal/progress/` (Observer & Progress)          | P1    | Élevé    | `FAIT`       | 2026-02-14 |
| T1.3  | Extraire `internal/fibonacci/threshold/`                     | P1    | Moyen    | `FAIT`       | 2026-02-14 |
| T1.4  | Extraire `internal/fibonacci/memory/`                        | P1    | Moyen    | `FAIT`       | 2026-02-14 |
| T1.5  | Réorganiser les fichiers restants dans `fibonacci/`          | P1    | Faible   | `FAIT`       | 2026-02-14 |
| T1.6  | Remplacer `GlobalFactory()` par injection de dépendances     | P1    | Critique | `FAIT`       | 2026-02-14 |
| T1.7  | Convertir les `init()` en initialisations explicites         | P1    | Moyen    | `FAIT`       | 2026-02-14 |
| T2.1  | Découpler orchestration de `config.AppConfig`                | P2    | Critique | `FAIT`       | 2026-02-14 |
| T2.2  | Segmenter l'interface `ResultPresenter`                      | P2    | Critique | `FAIT`       | 2026-02-14 |
| T2.3  | Découpler la calibration du CLI                              | P2    | Moyen    | `FAIT`       | 2026-02-14 |
| T2.4  | Découpler TUI/CLI des types `orchestration.CalculationResult`| P2    | Moyen    | `FAIT`       | 2026-02-14 |
| T2.5  | Réduire les responsabilités de `app.go`                      | P2    | Élevé    | `FAIT`       | 2026-02-14 |
| T2.6  | Centraliser la logique de seuils                             | P2    | Moyen    | `FAIT`       | 2026-02-14 |
| T2.7  | Extraire l'agrégation de progression dans orchestration      | P2    | Moyen    | `FAIT`       | 2026-02-14 |
| T3.1  | Éliminer le `goto` dans `fft_recursion.go`                   | P3    | Élevé    | `FAIT`       | 2026-02-14 |
| T3.2  | Extraire les chemins parallèle/séquentiel de FFT             | P3    | Moyen    | `FAIT`       | 2026-02-14 |
| T3.3  | Consolider la duplication exécution parallèle                | P3    | Moyen    | `FAIT`       | 2026-02-14 |
| T3.4  | Réduire la complexité de `DynamicThresholdManager`           | P3    | Élevé    | `FAIT`       | 2026-02-14 |
| T3.5  | Standardiser l'enveloppement d'erreurs                       | P3    | Moyen    | `FAIT`       | 2026-02-14 |
| T3.6  | Remplacer les nombres magiques par des constantes            | P3    | Faible   | `FAIT`       | 2026-02-14 |
| T3.7  | Corriger le bug de nommage `colorCyan` (TUI)                 | P3    | Faible   | `FAIT`       | 2026-02-14 |
| T3.8  | Corriger l'inconsistance `IntToBigInt` dans FFT              | P3    | Faible   | `FAIT`       | 2026-02-14 |
| T4.1  | Optimiser le cache FFT (copies redondantes)                  | P4    | Élevé    | `FAIT`       | 2026-02-14 |
| T4.2  | Réduire `MaxPooledBitLen`                                    | P4    | Élevé    | `FAIT`       | 2026-02-14 |
| T4.3  | Optimiser le buffer de channel progress                      | P4    | Moyen    | `FAIT`       | 2026-02-14 |
| T4.4  | Corriger l'allocation du hash de cache                       | P4    | Faible   | `FAIT`       | 2026-02-14 |
| T4.5  | Ajuster le dimensionnement de l'arena                        | P4    | Faible   | `FAIT`       | 2026-02-14 |
| T4.6  | Évaluer et documenter le pool warming                        | P4    | Moyen    | `FAIT`       | 2026-02-14 |
| T5.1  | Corriger la race condition `Freeze()/Notify()`               | P5    | Critique | `FAIT`       | 2026-02-14 |
| T5.2  | Augmenter la couverture de `app.go` (≥75%)                   | P5    | Moyen    | `FAIT`       | 2026-02-14 |
| T5.3  | Ajouter des tests de concurrence pools/factory               | P5    | Moyen    | `FAIT`       | 2026-02-14 |
| T5.4  | Intégrer le logging structuré (zerolog)                      | P5    | Élevé    | `FAIT`       | 2026-02-14 |
| T5.5  | Étendre les tests E2E                                        | P5    | Moyen    | `FAIT`       | 2026-02-14 |
| T5.6  | Enrichir les fuzz targets (identités supplémentaires)        | P5    | Faible   | `FAIT`       | 2026-02-14 |
| T5.7  | Élargir la hiérarchie d'erreurs                              | P5    | Moyen    | `FAIT`       | 2026-02-14 |
| T6.1  | Unifier le système de thèmes TUI/CLI                         | P6    | Moyen    | `FAIT`       | 2026-02-14 |
| T6.2  | Nettoyer les conventions de nommage CLI                      | P6    | Moyen    | `FAIT`       | 2026-02-14 |
| T6.3  | Extraire `FormatBytes` dans le package partagé               | P6    | Faible   | `FAIT`       | 2026-02-14 |
| T6.4  | Refactorer la génération de complétion shell                 | P6    | Moyen    | `FAIT`       | 2026-02-14 |
| T6.5  | Refactorer le modèle TUI (composition)                       | P6    | Moyen    | `FAIT`       | 2026-02-14 |
| T6.6  | Rendre l'override d'env déclaratif                           | P6    | Moyen    | `FAIT`       | 2026-02-14 |
| T6.7  | Corriger la thread-safety de `programRef` (TUI)              | P6    | Faible   | `FAIT`       | 2026-02-14 |

### Résumé par phase

| Phase | Description                          | Total | Fait | Restant | Progression |
|-------|--------------------------------------|-------|------|---------|-------------|
| P1    | Restructuration packages & globals   | 7     | 7    | 0       | 100%        |
| P2    | Découplage couches & interfaces      | 7     | 7    | 0       | 100%        |
| P3    | Qualité code & complexité            | 8     | 8    | 0       | 100%        |
| P4    | Performance & mémoire                | 6     | 6    | 0       | 100%        |
| P5    | Tests & observabilité                | 7     | 7    | 0       | 100%        |
| P6    | TUI, CLI & polish                    | 7     | 7    | 0       | 100%        |
| **—** | **Total**                            | **42**| **42**| **0**  | **100%**    |

---

## Résumé Exécutif

Ce document constitue le plan de refactorisation exhaustif du projet **FibGo** — un calculateur Fibonacci haute performance en Go (103 fichiers source, 91 fichiers test, 32 552 lignes, 17 packages). Le plan est fondé sur un audit complet réalisé par 5 agents spécialisés analysant en parallèle :

1. **Architecture & Structure des packages**
2. **Qualité du code & Complexité**
3. **Performance & Gestion mémoire**
4. **Tests & Observabilité**
5. **TUI, CLI & Configuration**

### Bilan de l'audit

| Domaine                  | Critique | Élevé | Moyen | Faible | Total |
|--------------------------|----------|-------|-------|--------|-------|
| Architecture & Packages  | 4        | 4     | 4     | 4      | 16    |
| Qualité du code          | 0        | 0     | 2     | 8      | 10    |
| Performance & Mémoire    | 0        | 2     | 3     | 2      | 7     |
| Tests & Observabilité    | 1        | 2     | 7     | 3      | 13    |
| TUI, CLI & Config        | 0        | 0     | 7     | 12     | 19    |
| **Total**                | **5**    | **8** | **23**| **29** | **65**|

### Les 5 Problèmes Critiques

| #  | Problème                                                          | Fichier principal         |
|----|-------------------------------------------------------------------|---------------------------|
| C1 | État global `GlobalFactory` — viole l'injection de dépendances  | `registry.go:218-239`     |
| C2 | Alias déprécié `MultiplicationStrategy` toujours en usage        | `strategy.go:86-91`       |
| C3 | Package `fibonacci` — God Package (60 fichiers, ~10K lignes)     | `internal/fibonacci/`     |
| C4 | Dépendance bidirectionnelle orchestration ↔ présentation         | `orchestrator.go`, `app.go` |
| C5 | Race conditions potentielles dans Observer `Freeze()/Notify()`   | `observer.go:135-146`     |

### Stratégie

**Approche Top-Down** : Restructurer les packages et interfaces d'abord (P1-P2), puis corriger la qualité/performance dans la nouvelle structure (P3-P4), consolider les tests (P5), et polir la présentation (P6).

**Principe directeur** : Chaque tâche doit être vérifiable par `go test -race ./...` après complétion. Aucune tâche ne casse le build.

---

## Table des Matières

- [Phase 1 : Restructuration des packages & suppression de l'état global](#phase-1--restructuration-des-packages--suppression-de-létat-global)
- [Phase 2 : Découplage des couches & nettoyage des interfaces](#phase-2--découplage-des-couches--nettoyage-des-interfaces)
- [Phase 3 : Qualité du code & réduction de la complexité](#phase-3--qualité-du-code--réduction-de-la-complexité)
- [Phase 4 : Performance & gestion mémoire](#phase-4--performance--gestion-mémoire)
- [Phase 5 : Tests, observabilité & robustesse](#phase-5--tests-observabilité--robustesse)
- [Phase 6 : TUI, CLI & polish de la présentation](#phase-6--tui-cli--polish-de-la-présentation)
- [Matrice de dépendances entre phases](#matrice-de-dépendances-entre-phases)
- [Estimation d'effort & timeline](#estimation-deffort--timeline)

---

# Phase 1 : Restructuration des packages & suppression de l'état global

> **Objectif** : Éclater le package monolithique `fibonacci` (60 fichiers, ~10K lignes) en sous-packages cohérents, éliminer l'état global, et supprimer le code déprécié.
>
> **Prérequis** : Aucun
> **Critère de succès global** : `go test -race -cover ./...` passe à 100%, aucune variable globale mutable exportée, package `fibonacci` réduit à ≤20 fichiers.

---

## T1.1 — Supprimer l'alias déprécié `MultiplicationStrategy`

### Contexte

`strategy.go:86-91` définit un alias de type déprécié :

```go
// Deprecated: Use Multiplier or DoublingStepExecutor instead.
type MultiplicationStrategy = DoublingStepExecutor
```

Cet alias est toujours exporté et potentiellement utilisé par du code externe. Sa présence crée de la confusion entre trois interfaces (`Multiplier`, `DoublingStepExecutor`, `MultiplicationStrategy`).

### Tâches détaillées

1. **Inventaire des usages** : Exécuter `grep -r "MultiplicationStrategy" internal/` pour identifier tous les fichiers consommateurs.
2. **Migration** : Remplacer chaque occurrence par `DoublingStepExecutor` (pour les usages nécessitant `ExecuteStep`) ou `Multiplier` (pour les usages ne nécessitant que `Multiply/Square`).
3. **Suppression** : Retirer l'alias de `strategy.go`.
4. **Documentation** : Mettre à jour `CLAUDE.md` pour retirer les mentions de `MultiplicationStrategy`.

### Fichiers touchés

| Fichier | Action |
|---------|--------|
| `internal/fibonacci/strategy.go` | Supprimer lignes 86-91 |
| Tous les fichiers référençant `MultiplicationStrategy` | Remplacer par l'interface appropriée |
| `CLAUDE.md` | Retirer les mentions de l'alias déprécié |

### Critères de validation

- `grep -r "MultiplicationStrategy" .` retourne 0 résultats (hors historique git)
- `go build ./...` compile sans erreur
- `go test -race ./...` passe à 100%

### Risques & Mitigation

| Risque | Probabilité | Impact | Mitigation |
|--------|-------------|--------|------------|
| Code externe utilisant l'alias | Faible | Moyen | Le module est `internal/`, pas de consommateurs externes |
| Confusion Multiplier vs DoublingStepExecutor | Moyen | Faible | Ajouter des commentaires guidant le choix |

---

## T1.2 — Extraire `internal/progress/` (Observer & Progress)

### Contexte

Le pattern Observer (3 fichiers, ~413 lignes) est actuellement dans `internal/fibonacci/` mais n'est pas spécifique aux algorithmes Fibonacci. Il est consommé par l'orchestration, le CLI, et le TUI.

### Fichiers à déplacer

| Source | Destination | Contenu |
|--------|-------------|---------|
| `fibonacci/observer.go` (146 lignes) | `progress/observer.go` | `ProgressObserver`, `ProgressSubject`, `Freeze()` |
| `fibonacci/observers.go` (139 lignes) | `progress/observers.go` | `ChannelObserver`, `LoggingObserver`, `NoOpObserver` |
| `fibonacci/progress.go` (128 lignes) | `progress/progress.go` | `ProgressUpdate`, `ProgressCallback`, `ReportStepProgress()`, `powersOf4` |

### Tâches détaillées

1. **Créer** `internal/progress/` avec les 3 fichiers.
2. **Mettre à jour les imports** dans tous les consommateurs :
   - `internal/fibonacci/calculator.go` — utilise `ProgressSubject`, `ChannelObserver`
   - `internal/fibonacci/doubling_framework.go` — utilise `ProgressCallback`, `ReportStepProgress`
   - `internal/fibonacci/matrix_framework.go` — idem
   - `internal/orchestration/orchestrator.go` — utilise `ProgressUpdate`
   - `internal/orchestration/interfaces.go` — référence `ProgressUpdate` dans `ProgressReporter`
   - `internal/cli/presenter.go` — utilise `ProgressUpdate`
   - `internal/cli/ui_display.go` — utilise `ProgressUpdate`
   - `internal/tui/bridge.go` — utilise `ProgressUpdate`
   - `internal/tui/messages.go` — potentiellement
3. **Déplacer les fichiers test** correspondants vers `internal/progress/`.
4. **Vérifier** que les types exportés restent accessibles.

### Pseudocode de la structure cible

```
internal/progress/
├── observer.go        # ProgressObserver interface, ProgressSubject
├── observers.go       # ChannelObserver, LoggingObserver, NoOpObserver
├── progress.go        # ProgressUpdate, ProgressCallback, ReportStepProgress
├── observer_test.go
├── observers_test.go
└── progress_test.go
```

### Critères de validation

- `go build ./...` compile sans erreur
- `go test -race ./...` passe
- Aucun import de `internal/fibonacci` dans `internal/progress/`
- `internal/progress/` n'a aucune dépendance vers `fibonacci`

---

## T1.3 — Extraire `internal/fibonacci/threshold/` (Seuils dynamiques)

### Contexte

La gestion des seuils dynamiques (389 lignes dans `dynamic_threshold.go` + `threshold_types.go` + parties de `options.go` et `constants.go`) est un sous-système autonome qui peut être isolé.

### Fichiers à déplacer / refactorer

| Source | Destination | Contenu |
|--------|-------------|---------|
| `fibonacci/dynamic_threshold.go` (389 lignes) | `fibonacci/threshold/manager.go` | `DynamicThresholdManager`, analyse FFT/parallèle |
| `fibonacci/threshold_types.go` | `fibonacci/threshold/types.go` | `IterationMetric`, types de seuils |
| Constantes de seuils de `fibonacci/constants.go` | `fibonacci/threshold/constants.go` | `DefaultParallelThreshold`, `DefaultFFTThreshold`, etc. |

### Tâches détaillées

1. **Créer** `internal/fibonacci/threshold/` (sous-package).
2. **Extraire** `DynamicThresholdManager` et types associés.
3. **Extraire** les constantes de seuils depuis `constants.go`.
4. **Mettre à jour** `doubling_framework.go` pour importer depuis `threshold/`.
5. **Vérifier** l'absence de dépendances circulaires.

### Critères de validation

- `dynamic_threshold.go` supprimé de `internal/fibonacci/`
- `go test -race ./internal/fibonacci/threshold/...` passe
- `go test -race ./...` passe

---

## T1.4 — Extraire `internal/fibonacci/memory/` (Arena, GC, Budget)

### Contexte

La gestion mémoire (arena bump allocator, GC controller, budget mémoire) est un sous-système orthogonal aux algorithmes.

### Fichiers à déplacer

| Source | Destination | Contenu |
|--------|-------------|---------|
| `fibonacci/arena.go` | `fibonacci/memory/arena.go` | `CalculationArena`, allocation bump |
| `fibonacci/gc_control.go` | `fibonacci/memory/gc_control.go` | `GCController`, désactivation GC |
| `fibonacci/memory_budget.go` | `fibonacci/memory/budget.go` | Estimation mémoire, validation budget |

### Tâches détaillées

1. **Créer** `internal/fibonacci/memory/`.
2. **Déplacer** les 3 fichiers et leurs tests.
3. **Mettre à jour** les imports dans les algorithmes (`fastdoubling.go`, `matrix.go`, `fft_based.go`).
4. **Vérifier** que `memory/` ne dépend pas de `fibonacci` (dépendance unidirectionnelle).

### Critères de validation

- `go test -race ./internal/fibonacci/memory/...` passe
- Pas de dépendance circulaire
- `go test -race ./...` passe

---

## T1.5 — Réorganiser les fichiers restants dans `fibonacci/`

### Contexte

Après extraction de `progress/`, `threshold/`, et `memory/`, le package `fibonacci` devrait passer de ~60 à ~25-30 fichiers. Regrouper logiquement les fichiers restants.

### Structure cible

```
internal/fibonacci/
├── calculator.go          # Calculator interface, FibCalculator decorator
├── registry.go            # CalculatorFactory, DefaultFactory
├── strategy.go            # Multiplier, DoublingStepExecutor, strategies
├── options.go             # Options struct (allégé, sans constantes de seuils)
├── constants.go           # Constantes non-seuils restantes
├── common.go              # Task semaphore, state pool, executeTasks
├── fastdoubling.go        # OptimizedFastDoubling (coreCalculator)
├── matrix.go              # MatrixExponentiation (coreCalculator)
├── fft_based.go           # FFTBasedCalculator (coreCalculator)
├── calculator_gmp.go      # GMPCalculator (build tag gmp)
├── doubling_framework.go  # DoublingFramework
├── matrix_framework.go    # MatrixFramework
├── matrix_ops.go          # Opérations matricielles
├── matrix_types.go        # Types matriciels
├── fft.go                 # Wrappers FFT (mulFFT, sqrFFT, smartMultiply)
├── generator.go           # SequenceGenerator interface
├── generator_iterative.go # IterativeGenerator
├── modular.go             # FastDoublingMod
├── testing.go             # Test helpers exportés
└── (fichiers _test.go correspondants)
```

### Critères de validation

- Package `fibonacci/` contient ≤25 fichiers `.go` (hors tests)
- `go test -race ./...` passe
- Aucun fichier orphelin

---

## T1.6 — Remplacer `GlobalFactory()` par injection de dépendances

### Contexte

`registry.go:218-239` définit une variable globale mutable :

```go
var globalFactory = NewDefaultFactory()

func GlobalFactory() *DefaultFactory {
    return globalFactory
}
```

`app.go:50` utilise directement cette globale : `factory := fibonacci.GlobalFactory()`.

### Tâches détaillées

1. **Modifier `app.New()`** pour accepter un `CalculatorFactory` en paramètre (ou créer le factory explicitement).
2. **Modifier `Application`** pour stocker le factory injecté.
3. **Conserver `GlobalFactory()` pour compatibilité** mais marquer comme déprécié.
4. **Modifier les tests** pour injecter des factories mock/isolés.

### Pseudocode de la modification

```go
// app.go — AVANT
func New(args []string, errWriter io.Writer) (*Application, int) {
    factory := fibonacci.GlobalFactory()
    // ...
}

// app.go — APRÈS
func New(args []string, errWriter io.Writer, opts ...AppOption) (*Application, int) {
    app := &Application{ErrWriter: errWriter}
    for _, opt := range opts {
        opt(app)
    }
    if app.Factory == nil {
        app.Factory = fibonacci.NewDefaultFactory()
    }
    // ...
}

type AppOption func(*Application)

func WithFactory(f fibonacci.CalculatorFactory) AppOption {
    return func(a *Application) { a.Factory = f }
}
```

### Critères de validation

- `app.New()` sans options crée un factory par défaut (pas de breaking change)
- Les tests peuvent injecter un factory custom
- `go test -race ./...` passe

---

## T1.7 — Convertir les `init()` en initialisations explicites

### Contexte

Plusieurs `init()` ont des effets de bord cachés :

| Fichier | `init()` | Effet |
|---------|----------|-------|
| `calculator_gmp.go:29` | `RegisterCalculator("gmp", ...)` | Enregistrement non-déterministe |
| `progress.go:51` | `PrecomputePowers4()` | Table de lookup pré-calculée |
| `matrix_ops.go:15` | Initialisation matrice | Setup matriciel |

### Tâches détaillées

1. **`calculator_gmp.go`** : Remplacer `init()` par une fonction `RegisterGMPCalculator(factory)` appelée explicitement depuis `app.go` si le build tag `gmp` est actif.
2. **`progress.go`** : Conserver `init()` car c'est une table de lookup immuable (acceptable).
3. **`matrix_ops.go`** : Évaluer si l'init peut être lazy ; si non, documenter.

### Critères de validation

- `grep -r "func init()" internal/fibonacci/` ne retourne que des initialisations de données immuables
- Les tests ne dépendent plus d'un ordre d'exécution des `init()`
- `go test -race ./...` passe

---

# Phase 2 : Découplage des couches & nettoyage des interfaces

> **Objectif** : Éliminer les dépendances bidirectionnelles entre orchestration et présentation, segmenter les interfaces trop larges, et clarifier les responsabilités de chaque couche.
>
> **Prérequis** : Phase 1 (extraction des packages)
> **Critère de succès global** : Aucune dépendance circulaire, interfaces ≤3 méthodes, orchestration indépendant de `config.AppConfig`.

---

## T2.1 — Découpler orchestration de `config.AppConfig`

### Contexte

`orchestrator.go:56` reçoit `config.AppConfig` directement :

```go
func ExecuteCalculations(ctx context.Context, calculators []fibonacci.Calculator,
    cfg config.AppConfig, // ← Couplage avec la couche application
    progressReporter ProgressReporter, out io.Writer) []CalculationResult
```

L'orchestration traduit ensuite les champs config en `fibonacci.Options` (lignes 64-68). Cette traduction devrait être dans `app.go`.

### Tâches détaillées

1. **Modifier `ExecuteCalculations()`** pour accepter `fibonacci.Options` + `uint64` (N) + `time.Duration` (timeout) au lieu de `config.AppConfig`.
2. **Déplacer** la traduction `AppConfig → Options` dans `app.go`.
3. **Mettre à jour** `AnalyzeComparisonResults()` de la même façon.
4. **Mettre à jour** tous les appelants.

### Pseudocode

```go
// orchestrator.go — APRÈS
func ExecuteCalculations(
    ctx context.Context,
    calculators []fibonacci.Calculator,
    n uint64,
    opts fibonacci.Options,
    progressReporter ProgressReporter,
    out io.Writer,
) []CalculationResult

// app.go — conversion dans l'appelant
opts := fibonacci.Options{
    ParallelThreshold: a.Config.Threshold,
    FFTThreshold:      a.Config.FFTThreshold,
    StrassenThreshold: a.Config.StrassenThreshold,
    // ...
}
results := orchestration.ExecuteCalculations(ctx, calcs, a.Config.N, opts, reporter, out)
```

### Critères de validation

- `internal/orchestration/` n'importe plus `internal/config/`
- `go build ./...` compile
- `go test -race ./...` passe

---

## T2.2 — Segmenter l'interface `ResultPresenter`

### Contexte

`orchestration/interfaces.go:54-97` définit une interface avec 4 méthodes mêlant affichage, formatage et gestion d'erreurs :

```go
type ResultPresenter interface {
    PresentComparisonTable(...)
    PresentResult(...)
    FormatDuration(...)
    HandleError(...)
}
```

### Tâches détaillées

1. **Scinder** en 3 interfaces distinctes :

```go
type ResultPresenter interface {
    PresentComparisonTable(results []CalculationResult, out io.Writer)
    PresentResult(result CalculationResult, n uint64, verbose, details, showValue bool, out io.Writer)
}

type DurationFormatter interface {
    FormatDuration(d time.Duration) string
}

type ErrorHandler interface {
    HandleError(err error, duration time.Duration, out io.Writer) int
}
```

2. **Mettre à jour** les implémentations CLI et TUI.
3. **Mettre à jour** les appelants dans `orchestrator.go` et `app.go`.

### Critères de validation

- Aucune interface n'a plus de 3 méthodes
- `go build ./...` compile
- `go test -race ./...` passe

---

## T2.3 — Découpler la calibration du CLI

### Contexte

`calibration/calibration.go:12` importe `internal/cli` pour afficher la progression :

```go
import "github.com/agbru/fibcalc/internal/cli"
// ...
go cli.DisplayProgress(&wg, progressChan, 1, out)
```

Cela empêche l'utilisation de la calibration en mode TUI ou headless.

### Tâches détaillées

1. **Créer** une interface `CalibrationProgressDisplayer` dans le package `calibration`.
2. **Injecter** le displayer depuis `app.go` au lieu d'importer `cli` directement.
3. **Supprimer** l'import de `cli` dans `calibration`.

### Critères de validation

- `internal/calibration/` n'importe plus `internal/cli/`
- La calibration fonctionne en mode CLI et TUI
- `go test -race ./...` passe

---

## T2.4 — Découpler TUI/CLI des types `orchestration.CalculationResult`

### Contexte

Les interfaces de présentation utilisent `orchestration.CalculationResult` dans leurs signatures, créant un couplage entre la couche présentation et la couche orchestration.

### Tâches détaillées

1. **Définir** un type `PresentationResult` dans le package de l'interface (ou dans `orchestration/interfaces.go`).
2. **Convertir** `CalculationResult → PresentationResult` à la frontière orchestration/présentation.
3. **Mettre à jour** les implémentations CLI et TUI.

### Critères de validation

- Les packages `cli/` et `tui/` n'importent `orchestration` que pour l'interface (pas les types concrets)
- `go test -race ./...` passe

---

## T2.5 — Réduire les responsabilités de `app.go`

### Contexte

`app.go` (370 lignes) mélange : lifecycle, dispatching, calibration, calcul CLI, last-digits, analyse des résultats, et I/O fichier.

### Tâches détaillées

1. **Extraire** le dispatching de modes dans `app/dispatcher.go` :

```go
type ModeDispatcher struct{}

func (d *ModeDispatcher) SelectMode(cfg config.AppConfig) Mode {
    if cfg.Completion != "" { return CompletionMode{} }
    if cfg.Calibrate       { return CalibrationMode{} }
    if cfg.TUI             { return TUIMode{} }
    return CLIMode{}
}

type Mode interface {
    Run(ctx context.Context, app *Application, out io.Writer) int
}
```

2. **Extraire** la logique de calcul dans `app/calculator_runner.go`.
3. **Réduire** `app.go` à ≤100 lignes (lifecycle + composition).

### Critères de validation

- `app.go` ≤100 lignes
- Chaque mode dans son propre fichier
- `go test -race ./...` passe

---

## T2.6 — Centraliser la logique de seuils

### Contexte

La résolution des seuils est dispersée : `config.go` (validation), `app.go` (adaptive), `constants.go` (defaults), `options.go` (normalisation).

### Tâches détaillées

1. **Créer** `internal/config/thresholds.go` avec `ValidateThresholds()` et `ApplyAdaptiveThresholds()`.
2. **Déplacer** la logique de `applyAdaptiveThresholds()` depuis `app.go`.
3. **Documenter** la chaîne de résolution : CLI flags > env vars > profil calibration > estimation adaptative > défauts statiques.

### Critères de validation

- Un seul fichier gère toute la résolution de seuils
- `go test -race ./...` passe

---

## T2.7 — Extraire l'agrégation de progression dans orchestration

### Contexte

L'agrégation de progression (calcul de la moyenne, ETA) est dupliquée entre CLI (`cli/ui_display.go:42`) et TUI (`tui/bridge.go:40-60`). Les deux créent un `format.NewProgressWithETA()` et agrègent indépendamment.

### Tâches détaillées

1. **Déplacer** l'agrégation dans `orchestration/` (ou un package dédié).
2. **Envoyer** des updates pré-agrégées aux reporters.
3. **Simplifier** les implémentations CLI et TUI.

### Critères de validation

- Aucune duplication de logique d'agrégation
- `go test -race ./...` passe

---

# Phase 3 : Qualité du code & réduction de la complexité

> **Objectif** : Éliminer le code dupliqué, réduire la complexité cyclomatique, standardiser les patterns d'erreur, et supprimer les nombres magiques.
>
> **Prérequis** : Phase 1 (packages restructurés)
> **Critère de succès global** : Aucune fonction >15 de complexité cyclomatique, 0 nombre magique non-documenté.

---

## T3.1 — Éliminer le `goto` dans `fft_recursion.go`

### Contexte

`bigfft/fft_recursion.go:138,152` utilise un `goto Reconstruct` pour sauter à une section de reconstruction après le fork parallèle/séquentiel.

### Tâches détaillées

1. **Extraire** la logique de reconstruction dans une fonction `executeReconstruction()`.
2. **Restructurer** le if-else parallèle/séquentiel pour appeler la fonction extraite.
3. **Supprimer** le label `Reconstruct:` et le `goto`.

### Critères de validation

- `grep "goto" internal/bigfft/` retourne 0 résultats
- Benchmarks : pas de régression >2%
- `go test -race ./...` passe

---

## T3.2 — Extraire les chemins parallèle/séquentiel de `executeDoublingStepFFT`

### Contexte

`fibonacci/fft.go:83-237` (155 lignes) gère les deux chemins (parallèle et séquentiel) inline, avec ~40 lignes de duplication.

### Tâches détaillées

1. **Extraire** `executeFFTTransformsParallel(ctx, ...)` (~80 lignes).
2. **Extraire** `executeFFTTransformsSequential(ctx, ...)` (~45 lignes).
3. **Réduire** `executeDoublingStepFFT` à ~30 lignes (setup + dispatch).

### Critères de validation

- `executeDoublingStepFFT` ≤50 lignes
- Benchmarks : pas de régression >2%
- `go test -race ./...` passe

---

## T3.3 — Consolider la duplication dans l'exécution parallèle

### Contexte

Le pattern d'exécution parallèle (3 goroutines + channels/WaitGroup + error collection) est dupliqué entre `fft.go:111-191` et `doubling_framework.go:67-122`.

### Tâches détaillées

1. **Créer** un helper générique `executeParallel3(ctx, op1, op2, op3 func() error) error`.
2. **Refactorer** les deux sites d'appel pour utiliser le helper.
3. **Tester** le helper isolément.

### Critères de validation

- Un seul implémentation du pattern parallèle 3-voies
- `go test -race ./...` passe

---

## T3.4 — Réduire la complexité de `DynamicThresholdManager`

### Contexte

`analyzeFFTThreshold()` (54 lignes) et `analyzeParallelThreshold()` (49 lignes) dans `dynamic_threshold.go` partagent une structure quasi-identique mais avec des métriques différentes.

### Tâches détaillées

1. **Extraire** un pattern commun `analyzeThreshold(metrics, speedupThreshold, adjustmentFactor) int`.
2. **Extraire** `filterMetricsByMode(metrics, mode)` et `calculateSpeedupRatio(avgA, avgB)`.
3. **Réduire** la complexité cyclomatique de chaque méthode à ≤10.

### Critères de validation

- Complexité cyclomatique de chaque méthode d'analyse ≤10
- `go test -race ./...` passe

---

## T3.5 — Standardiser l'enveloppement d'erreurs

### Contexte

Certaines fonctions retournent des erreurs sans contexte (`return err`), d'autres les enveloppent (`return fmt.Errorf("context: %w", err)`). Incohérent.

### Fichiers concernés

| Fichier | Lignes | Problème |
|---------|--------|----------|
| `bigfft/fft_recursion.go:85` | Bare `fmt.Errorf` sans wrapping | Pas de `%w` |
| `fibonacci/fft.go:101-108` | `return err` sans contexte | Debugging difficile |
| `cli/completion.go:90,134,180` | `return err` sans contexte | Pas de trace |
| `bigfft/scan.go:67,72,76` | `return err` sans contexte | Pas de trace |

### Tâches détaillées

1. **Ajouter** du contexte à chaque `return err` : `return fmt.Errorf("transform FK failed: %w", err)`.
2. **Standardiser** le format : `"<composant> <opération> failed: %w"`.
3. **Vérifier** que `errors.Is()` et `errors.As()` fonctionnent toujours.

### Critères de validation

- `grep -rn "return err$" internal/` retourne uniquement des cas justifiés (interface satisfaction)
- `go test -race ./...` passe

---

## T3.6 — Remplacer les nombres magiques par des constantes nommées

### Contexte

Plusieurs nombres magiques sont utilisés sans explication :

| Valeur | Fichier | Usage |
|--------|---------|-------|
| `0.69424` | `fastdoubling.go:101`, `fft_based.go:53`, `arena.go:24` | Approximation de log2(φ) |
| `+2` | `fft.go:91` | Marge de sécurité FFT |
| `0.7 / 0.3` | `tui/metrics.go:60` | Facteurs EMA |
| `60 / 100` | `tui/model.go:269` | Split layout 60/40 |
| `7` | `tui/model.go:261` | Hauteur fixe du panel métriques |

### Tâches détaillées

1. **Définir** `FibonacciGrowthFactor = 0.69424` dans `fibonacci/constants.go` avec documentation : `// log2(phi), where phi ≈ 1.618 (golden ratio)`.
2. **Définir** `FFTSafetyMarginWords = 2` dans `fibonacci/fft.go`.
3. **Définir** les constantes TUI dans `tui/model.go` : `LogsPanelWidthPercent = 60`, `EMASmoothFactor = 0.7`, etc.
4. **Remplacer** tous les usages littéraux.

### Critères de validation

- Aucun nombre magique non-documenté dans le code
- `go test -race ./...` passe

---

## T3.7 — Corriger le bug de nommage `colorCyan` dans le TUI

### Contexte

`tui/styles.go:15` définit `colorCyan = lipgloss.Color("#FF8C00")` — c'est du **orange**, pas du cyan. Confusion sémantique.

### Tâches détaillées

1. **Renommer** `colorCyan` en `colorAccentOrange`.
2. **Mettre à jour** tous les usages dans `styles.go`.

### Critères de validation

- Aucune variable de couleur mal nommée
- `go test -race ./...` passe

---

## T3.8 — Corriger l'inconsistance `IntToBigInt` dans FFT séquentiel

### Contexte

`fibonacci/fft.go:204,219,234` — en mode séquentiel, les valeurs de retour de `IntToBigInt()` sont ignorées, contrairement au mode parallèle (lignes 143, 162, 181) où elles sont utilisées.

### Tâches détaillées

1. **Assigner** explicitement les valeurs de retour :

```go
// AVANT
p1.IntToBigInt(s.T3)
// APRÈS
s.T3 = p1.IntToBigInt(s.T3)
```

### Critères de validation

- Usage cohérent entre modes parallèle et séquentiel
- `go test -race ./...` passe

---

# Phase 4 : Performance & gestion mémoire

> **Objectif** : Corriger les inefficiences de mémoire et performance identifiées, optimiser le cache FFT, ajuster les pools.
>
> **Prérequis** : Phase 1 (packages restructurés)
> **Critère de succès global** : Aucune régression benchmark >2%, réduction de l'empreinte mémoire pour les gros calculs.

---

## T4.1 — Optimiser le cache FFT (copies redondantes)

### Contexte

`bigfft/fft_cache.go:171-187` (Get) et `222-242` (Put) effectuent des copies complètes des données à chaque accès, avec allocation de backing buffers. Pour F(10M), cela représente ~300KB par accès.

### Tâches détaillées

1. **Option A (recommandée)** : Retourner des références avec garantie d'immutabilité documentée. Ajouter une assertion runtime vérifiant qu'aucune mutation n'a lieu.
2. **Option B** : Implémenter un copy-on-write wrapper.
3. **Mesurer** l'impact avec `go test -bench=BenchmarkFib -benchmem`.

### Critères de validation

- Benchmark FFT : amélioration ≥5% pour les calculs itératifs
- `go test -race ./...` passe (pas de data race)

---

## T4.2 — Réduire `MaxPooledBitLen`

### Contexte

`fibonacci/common.go:34` définit `MaxPooledBitLen = 100_000_000` (12.5 MB par `big.Int`). C'est trop large — les objets de 12.5 MB restent dans le pool et empêchent le GC.

### Tâches détaillées

1. **Réduire** à `MaxPooledBitLen = 50_000_000` (6.25 MB).
2. **Mesurer** l'impact mémoire avec `go test -bench=BenchmarkFib -benchmem`.
3. **Optionnel** : Implémenter un pool à deux niveaux (petits <5MB, grands <50MB).

### Critères de validation

- RSS pour F(100M) réduit de ≥10%
- Pas de régression benchmark >2%

---

## T4.3 — Optimiser le buffer de channel progress

### Contexte

`orchestration/orchestrator.go:38,58` — `ProgressBufferMultiplier = 50` est excessif pour un seul calculateur.

### Tâches détaillées

1. **Réduire** le multiplier de 50 à 5 (ou implémenter un buffer adaptatif).
2. **Ajouter** un `select` avec timeout pour le backpressure gracieux.

### Critères de validation

- Pas de blocage du channel progress dans les tests
- Réduction de ~5-10KB de mémoire par invocation

---

## T4.4 — Corriger l'allocation du hash de cache

### Contexte

`bigfft/fft_cache.go:103-115` — `buf := make([]byte, 8)` est alloué à chaque calcul de clé de cache.

### Tâches détaillées

1. **Remplacer** par une variable stack-allocated : `var buf [8]byte`.
2. **Utiliser** `buf[:]` pour les appels `Write()`.

### Critères de validation

- `go test -bench=BenchmarkCacheKey -benchmem` montre 0 allocations
- `go test -race ./...` passe

---

## T4.5 — Ajuster le dimensionnement de l'arena

### Contexte

`fibonacci/arena.go:26` alloue toujours 10 temporaires, mais pendant les étapes FFT, l'arena est épuisée après 5-10 itérations.

### Tâches détaillées

1. **Augmenter** l'estimation : `totalWords := wordsPerInt * 15` pour les calculs FFT.
2. **Adapter** le nombre de temporaires en fonction de N (5 pour petit N, 15 pour gros N).
3. **Ajouter** un log de warning quand le fallback heap est déclenché.

### Critères de validation

- Moins de fallback heap pour les calculs FFT ≥100K
- `go test -race ./...` passe

---

## T4.6 — Évaluer et documenter le pool warming

### Contexte

`bigfft/pool_warming.go:30-98` — le pré-chauffage des pools n'a pas de benchmark démontrant son utilité.

### Tâches détaillées

1. **Ajouter** un benchmark : `BenchmarkFib10M_WithWarming` vs `BenchmarkFib10M_WithoutWarming`.
2. **Si bénéfice <2%** : documenter comme optionnel ou supprimer.
3. **Si bénéfice >2%** : déclencher le warming AVANT le calcul, pas après.

### Critères de validation

- Résultats de benchmark documentés
- Décision prise et documentée (garder/supprimer/déplacer)

---

# Phase 5 : Tests, observabilité & robustesse

> **Objectif** : Combler les lacunes de tests, ajouter du logging structuré, corriger les race conditions, et étendre les tests E2E.
>
> **Prérequis** : Phases 1-3 (structure stable, code nettoyé)
> **Critère de succès global** : Couverture ≥80% globale, app.go ≥75%, 0 race condition sous `-race`.

---

## T5.1 — Corriger la race condition `Freeze()/Notify()`

### Contexte

`observer.go:135-146` — `Freeze()` crée un snapshot lock-free, mais si un observer panique dans `Update()`, le comportement est indéfini. De plus, aucun test ne vérifie le comportement concurrent.

### Tâches détaillées

1. **Ajouter** un recovery de panic dans la callback Freeze :

```go
func (s *ProgressSubject) Freeze(calcIndex int) ProgressCallback {
    s.mu.RLock()
    snapshot := make([]ProgressObserver, len(s.observers))
    copy(snapshot, s.observers)
    s.mu.RUnlock()

    return func(progress float64) {
        for _, observer := range snapshot {
            func() {
                defer func() {
                    if r := recover(); r != nil {
                        // Log panic, ne pas propager
                    }
                }()
                observer.Update(calcIndex, progress)
            }()
        }
    }
}
```

2. **Ajouter** des tests de concurrence :
   - `TestProgressSubject_FreezeRaceConditions` : register/unregister pendant freeze
   - `TestProgressSubject_FreezeSnapshot` : vérifier l'isolation du snapshot

### Critères de validation

- `go test -race -count=100 ./internal/progress/...` passe sans data race
- Couverture de `Freeze()` ≥90%

---

## T5.2 — Augmenter la couverture de `app.go` (63.5% → ≥75%)

### Contexte

`internal/app/` a la couverture la plus basse (63.5%). Fonctions non testées : `applyAdaptiveThresholds()`, `runCalibration()`, chemins TUI.

### Tâches détaillées

1. **Tester** `applyAdaptiveThresholds()` avec différents CPU counts/architectures.
2. **Tester** le dispatching de modes (completion, calibration, TUI, CLI).
3. **Tester** les chemins d'erreur (timeout, cancellation, mémoire insuffisante).

### Critères de validation

- Couverture `internal/app/` ≥75%
- `go test -race ./internal/app/...` passe

---

## T5.3 — Ajouter des tests de concurrence pour les pools et le factory

### Contexte

Aucun test de concurrence pour `CalculatorFactory.Create()` concurrent, `CalculationStatePool` concurrent, ou `matrixStatePool` concurrent.

### Tâches détaillées

1. **Ajouter** `TestCalculatorFactory_ConcurrentCreation` : 10 goroutines appelant `Create("fast")`.
2. **Ajouter** `TestCalculationStatePool_ConcurrentAllocation` : 100 goroutines Get/Put.
3. **Exécuter** avec `-race -count=100`.

### Critères de validation

- 0 data race détectée
- `go test -race -count=100 ./internal/fibonacci/...` passe

---

## T5.4 — Intégrer le logging structuré (zerolog)

### Contexte

Zerolog est importé mais à peine utilisé. Aucun logging aux points de décision critiques.

### Tâches détaillées

1. **Ajouter** du logging aux points clés :
   - Sélection d'algorithme (`registry.go`)
   - Ajustement de seuils (`dynamic_threshold.go`)
   - Cache FFT hits/misses (`fft_cache.go`)
   - Déclenchement GC controller (`gc_control.go`)
   - Distribution des tâches parallèles (`common.go`)
2. **Utiliser** les niveaux appropriés : `Debug` pour les décisions, `Info` pour les résultats, `Warn` pour les anomalies.
3. **Ajouter** un contexte structuré (nom du calculateur, N, itération).

### Critères de validation

- `FIBCALC_VERBOSE=true ./fibcalc -n 1000000` produit des logs structurés
- Les logs ne dégradent pas les performances (vérifié par benchmark)
- `go test -race ./...` passe

---

## T5.5 — Étendre les tests E2E

### Contexte

Seulement 8 tests E2E. Manquent : erreurs, timeout, signaux, fichier de sortie, calibration.

### Tâches détaillées

1. **Ajouter** des tests pour :
   - Flags invalides (exit code 4)
   - `--output-file` (vérifier le contenu du fichier)
   - `--memory-limit` avec N trop grand (exit code approprié)
   - `--timeout` avec validation réelle (>1s)
   - Shell completion (bash, zsh, fish, powershell — vérifier la sortie non-vide)
2. **Structurer** les tests par catégorie.

### Critères de validation

- ≥15 tests E2E couvrant les cas nominaux et d'erreur
- `go test -race ./test/e2e/...` passe

---

## T5.6 — Enrichir les fuzz targets avec des identités supplémentaires

### Contexte

`FuzzFibonacciIdentities` ne teste que l'identité de doubling. Manquent : Cassini, addition, GCD.

### Tâches détaillées

1. **Ajouter** l'identité de Cassini : `F(n-1)·F(n+1) - F(n)² = (-1)^n`.
2. **Ajouter** l'identité d'addition : `F(m+n) = F(m)·F(n+1) + F(m-1)·F(n)`.
3. **Enrichir** le corpus avec des valeurs limites : `2^10, 2^20, F(12)=144, F(13)=233`.

### Critères de validation

- `go test -fuzz=FuzzFibonacciIdentities -fuzztime=30s ./internal/fibonacci/` passe

---

## T5.7 — Élargir la hiérarchie d'erreurs

### Contexte

Seulement 2 types d'erreurs custom (`ConfigError`, `CalculationError`). Manquent : `TimeoutError`, `ValidationError`, `MemoryError`.

### Tâches détaillées

1. **Ajouter** `TimeoutError` avec `Operation` et `Limit`.
2. **Ajouter** `ValidationError` avec `Field` et `Message`.
3. **Ajouter** `MemoryError` avec `Requested`, `Available`, `Limit`.
4. **Mettre à jour** `HandleCalculationError()` pour utiliser `errors.As()` avec les nouveaux types.

### Critères de validation

- Les tests de `HandleCalculationError` couvrent tous les nouveaux types
- `go test -race ./internal/errors/...` passe avec couverture ≥95%

---

# Phase 6 : TUI, CLI & polish de la présentation

> **Objectif** : Améliorer la maintenabilité de la couche présentation, unifier le theming, nettoyer les conventions de nommage.
>
> **Prérequis** : Phase 2 (interfaces découplées)
> **Critère de succès global** : Convention Display*/Format*/Write*/Print* respectée à 100%, TUI respecte NO_COLOR.

---

## T6.1 — Unifier le système de thèmes TUI/CLI

### Contexte

Le TUI (`tui/styles.go`) définit des couleurs hardcodées sans utiliser le système de thèmes de `internal/ui/themes.go`. Le TUI ignore `NO_COLOR`.

### Tâches détaillées

1. **Créer** un `TUITheme` dans `internal/ui/themes.go` avec les couleurs du thème orange.
2. **Refactorer** `tui/styles.go` pour utiliser les couleurs du thème courant.
3. **Vérifier** que `NO_COLOR=1` désactive les couleurs dans le TUI.

### Critères de validation

- `NO_COLOR=1 ./fibcalc --tui -n 100` n'affiche aucun code ANSI
- `go test -race ./internal/tui/...` passe

---

## T6.2 — Nettoyer les conventions de nommage CLI

### Contexte

La convention `Display*/Format*/Write*/Print*` est documentée mais partiellement suivie. Plusieurs fonctions de `cli/ui.go` sont de simples délégations vers `format/`.

### Tâches détaillées

1. **Supprimer** les wrappers triviaux (`FormatExecutionDuration`, `FormatNumberString`, `FormatETA`) — les appelants importent `format/` directement.
2. **Renommer** les helpers internes pour cohérence (`displayResultHeader` → `_displayResultHeader` ou l'exporter si nécessaire).
3. **Documenter** la convention dans les commentaires de package.

### Critères de validation

- Aucun wrapper trivial (délégation pure) dans `cli/`
- Convention documentée en haut de `cli/output.go`

---

## T6.3 — Extraire `FormatBytes` dans le package partagé

### Contexte

`cli/presenter.go:122` et `tui/metrics.go:137` définissent la même fonction `formatBytes()`.

### Tâches détaillées

1. **Ajouter** `FormatBytes(b uint64) string` dans `internal/format/numbers.go`.
2. **Supprimer** les implémentations dupliquées.
3. **Mettre à jour** les imports.

### Critères de validation

- `grep -rn "func formatBytes" internal/` retourne 1 seul résultat (dans `format/`)
- `go test -race ./...` passe

---

## T6.4 — Refactorer la génération de complétion shell

### Contexte

`cli/completion.go:33-260` — chaque fonction shell (bash, zsh, fish, powershell) hardcode la liste de flags. Ajouter un flag nécessite de modifier 4 fonctions.

### Tâches détaillées

1. **Créer** un registre de flags :

```go
type FlagCompletion struct {
    Long    string
    Short   string
    Help    string
    Values  []string // pour les options enum
}

var flagRegistry = []FlagCompletion{...}
```

2. **Générer** les scripts de complétion depuis le registre.
3. **Tester** que la sortie est syntaxiquement valide pour chaque shell.

### Critères de validation

- Ajouter un nouveau flag ne nécessite qu'une entrée dans le registre
- `go test -race ./internal/cli/...` passe

---

## T6.5 — Refactorer le modèle TUI (composition)

### Contexte

`tui/model.go` — le `Model` struct a 13 champs mélangeant état UI, état d'exécution, et dimensions.

### Tâches détaillées

1. **Extraire** un `ExecutionState` struct : `ctx`, `cancel`, `calculators`, `generation`, `done`, `exitCode`.
2. **Extraire** un `LayoutManager` struct : `width`, `height`, `layoutPanels()`, `metricsWidth()`, `metricsHeight()`.
3. **Réduire** `Model` à la composition de composants + `ExecutionState` + `LayoutManager`.

### Critères de validation

- `Model` struct ≤8 champs
- `go test -race ./internal/tui/...` passe

---

## T6.6 — Rendre l'override d'env déclaratif

### Contexte

`config/env.go:100-180` — l'application des env vars est procédurale et répétitive (3 lignes par champ).

### Tâches détaillées

1. **Définir** un tableau déclaratif d'overrides :

```go
type envOverride struct {
    flagName string
    envKey   string
    aliases  []string
    apply    func(*AppConfig, string) error
}

var overrides = []envOverride{
    {"n", "N", nil, func(c *AppConfig, v string) error { /* ... */ }},
    // ...
}
```

2. **Itérer** sur le tableau dans `applyEnvOverrides()`.
3. **Ajouter** un warning pour les valeurs invalides (ex: `FIBCALC_VERBOSE=maybe`).

### Critères de validation

- Ajouter un nouveau env var = 1 entrée dans le tableau
- `go test -race ./internal/config/...` passe

---

## T6.7 — Corriger la thread-safety de `programRef` (TUI)

### Contexte

`tui/bridge.go:19-27` — `programRef.program` est muté sans synchronisation. Si une goroutine appelle `Send()` pendant l'assignation de `program`, il y a une race condition.

### Tâches détaillées

1. **Ajouter** un `sync.RWMutex` à `programRef`.
2. **Créer** `SetProgram(p *tea.Program)` avec lock.
3. **Modifier** `Send()` pour utiliser `RLock`.

### Critères de validation

- `go test -race -count=100 ./internal/tui/...` passe
- Pas de data race détectée

---

# Matrice de dépendances entre phases

```
Phase 1 (Restructuration packages)
  ↓
Phase 2 (Découplage couches)     ← dépend de P1
  ↓
Phase 3 (Qualité code)           ← dépend de P1 (structure stable)
  ↓
Phase 4 (Performance)            ← dépend de P1 (packages isolés)
  ↓
Phase 5 (Tests & observabilité)  ← dépend de P1-P3 (code nettoyé)
  ↓
Phase 6 (TUI/CLI polish)         ← dépend de P2 (interfaces découplées)
```

**Parallélisation possible :**
- P3 et P4 peuvent être exécutées en parallèle (domaines indépendants après P1)
- P5 et P6 peuvent être exécutées en parallèle (domaines indépendants après P2-P3)

```
P1 ──→ P2 ──→ P5
 │      │      ↑
 ├──→ P3 ──→──┤
 │             │
 └──→ P4 ──→──┘
        │
        └──→ P6
```

---

# Estimation d'effort & timeline

## Par phase

| Phase | Tâches | Effort estimé | Risque | Parallélisable |
|-------|--------|---------------|--------|----------------|
| P1 — Restructuration | 7 | 3-4 jours | Élevé (beaucoup de fichiers) | Non (fondation) |
| P2 — Découplage | 7 | 2-3 jours | Moyen | Non (dépend de P1) |
| P3 — Qualité code | 8 | 2-3 jours | Faible (mécanique) | Oui (avec P4) |
| P4 — Performance | 6 | 2-3 jours | Moyen (benchmarks) | Oui (avec P3) |
| P5 — Tests | 7 | 3-4 jours | Faible | Oui (avec P6) |
| P6 — TUI/CLI | 7 | 2-3 jours | Faible | Oui (avec P5) |
| **Total** | **52** | **14-20 jours** | — | — |

## Par sévérité

| Sévérité | Nombre de tâches | Phases concernées |
|----------|-----------------|-------------------|
| Critique (C1-C5) | 5 | P1, P2, P5 |
| Élevé | 8 | P1, P2, P3 |
| Moyen | 23 | P2-P6 |
| Faible | 16 | P3, P6 |

## Métriques cibles post-refactorisation

| Métrique | Avant | Après (cible) |
|----------|-------|---------------|
| Fichiers dans `fibonacci/` | 60 | ≤25 |
| Lignes dans `fibonacci/` | ~10K | ~4K |
| Variables globales mutables | 3+ | 0 |
| Complexité cyclomatique max | ~14 | ≤12 |
| Couverture tests globale | ~85% | ≥80% |
| Couverture `app.go` | 63.5% | ≥75% |
| Interfaces >3 méthodes | 1 (`ResultPresenter`) | 0 |
| Dépendances circulaires | 1 | 0 |
| `goto` statements | 1 | 0 |
| Nombres magiques non-documentés | ~8 | 0 |
| Wrappers de délégation pure | ~3 | 0 |
| Fonctions dupliquées | 3 (`formatBytes`, parallel exec, progress aggregation) | 0 |

---

## Principes de la refactorisation (issus de CLAUDE.md)

1. **Changements chirurgicaux** : Chaque tâche ne touche que les fichiers strictement nécessaires.
2. **Vérification continue** : `go test -race ./...` après chaque tâche — pas de commit qui casse le build.
3. **Simplicité** : Pas de sur-ingénierie. Les extractions de packages ne créent pas d'abstractions inutiles.
4. **Goal-driven** : Chaque tâche a des critères de validation mesurables.
5. **Conventions commits** : `refactor(<scope>): <description>` pour chaque tâche.

---

**Ce plan est opérationnel si :** les tâches P1 sont exécutées en séquence (T1.1 → T1.7), puis P2 en séquence, puis P3-P4 en parallèle, puis P5-P6 en parallèle. Chaque tâche est auto-contenue et vérifiable indépendamment.
