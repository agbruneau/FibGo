# Plan de Test Exhaustif - Mode TUI

## Sommaire

Le mode TUI (Terminal User Interface) du projet FibGo utilise le framework Bubble Tea (architecture Elm) pour offrir un tableau de bord interactif affichant la progression des calculs de Fibonacci, les metriques systeme et les resultats.

Ce document recense l'inventaire complet des tests, les anomalies identifiees et corrigees, la matrice de couverture et la procedure de regression.

---

## 1. Architecture TUI

```
Model (model.go)           -- Modele racine, gestion des messages
  +-- HeaderModel           -- Barre de titre (version, elapsed)
  +-- LogsModel (logs.go)   -- Journal scrollable avec viewport
  +-- MetricsModel           -- Metriques memoire et vitesse
  +-- ChartModel             -- Sparkline de progression
  +-- FooterModel            -- Barre de statut et raccourcis
  +-- KeyMap (keymap.go)     -- Bindings clavier

Bridge (bridge.go)          -- Pont orchestration <-> TUI
  +-- programRef             -- Reference thread-safe au tea.Program
  +-- TUIProgressReporter    -- Impl. ProgressReporter
  +-- TUIResultPresenter     -- Impl. ResultPresenter

Messages (messages.go)      -- Types de messages Bubble Tea
Styles (styles.go)          -- Palette btop-inspired (lipgloss)
```

---

## 2. Inventaire des Tests (109 tests, 8 fichiers)

### model_test.go (40 tests)

| Test | Couverture |
|------|-----------|
| `TestNewModel` | Construction initiale |
| `TestNewModel_WithCalculators` | Injection de calculateurs |
| `TestModel_Update_WindowSize` | Message WindowSizeMsg |
| `TestModel_Update_ProgressMsg` | Progression (chart, logs, metrics) |
| `TestModel_Update_ProgressMsg_Paused` | Progression ignoree en pause |
| `TestModel_Update_CalculationComplete` | Signal de fin |
| `TestModel_Update_ErrorMsg` | Gestion d'erreur |
| `TestModel_View_Initializing` | Affichage avant resize |
| `TestModel_View_WithSize` | Rendu apres resize |
| `TestModel_HandleKey_Pause` | Touche espace (pause/resume) |
| `TestModel_HandleKey_Reset` | Touche r (reset chart/metrics) |
| `TestModel_Update_ComparisonResultsMsg` | Resultats de comparaison |
| `TestModel_Update_FinalResultMsg` | Resultat final |
| `TestModel_Update_ProgressDoneMsg` | Signal fin progression |
| `TestModel_Update_ContextCancelledMsg` | Annulation contexte |
| `TestModel_Update_CalculationComplete_NonZeroExitCode` | Code sortie non-zero |
| `TestModel_Update_MemStatsMsg` | Stats memoire |
| `TestModel_Update_TickMsg_NotPaused` | Tick actif |
| `TestModel_Update_TickMsg_Paused` | Tick en pause |
| `TestModel_HandleKey_Quit_Q` | Touche q |
| `TestModel_HandleKey_Quit_CtrlC` | Ctrl+C |
| `TestModel_HandleKey_ScrollKeys` | Up/Down/PgUp/PgDn/j/k |
| `TestModel_HandleKey_Unknown` | Touche inconnue |
| `TestModel_Update_UnknownMsg` | Message inconnu |
| `TestModel_LayoutPanels_VerySmallTerminal` | Terminal minuscule |
| `TestModel_LayoutPanels_MinBodyHeight` | Clamping bodyHeight |
| `TestModel_HandleKey_Reset_ClearsMetrics` | Reset remet metriques a zero |
| `TestModel_Init_ReturnsCommands` | Init() retourne batch |
| `TestSampleMemStatsCmd_ReturnsMemStatsMsg` | sampleMemStatsCmd() |
| `TestWatchContextCmd_SendsOnCancel` | watchContextCmd() cancel |
| `TestWatchContextCmd_AlreadyCancelled` | watchContextCmd() deja annule |
| `TestModel_layoutPanels_Percentages` | Proportions 60/40 |
| `TestModel_layoutPanels_MinimumSizes` | Clamp minimum |
| `TestModel_metricsHeight_ConsistentWithLayout` | Coherence metricsHeight/layout |
| `TestModel_View_ContainsAllComponents` | Tous composants presents |
| `TestModel_Update_VeryWideTerminal` | Terminal tres large (500) |
| `TestModel_metricsHeight_SmallTerminal` | metricsHeight petit terminal |
| `TestTickCmd_ReturnsCmd` | tickCmd() |
| `TestStartCalculationCmd_ReturnsCompleteMsg` | startCalculationCmd() execution |
| `TestModel_HandleKey_Quit_CancelsContext` | Quit annule le contexte |

### logs_test.go (16 tests)

| Test | Couverture |
|------|-----------|
| `TestLogsModel_AddProgressEntry` | Ajout progression |
| `TestLogsModel_AddProgressEntry_Complete` | 100% OK |
| `TestLogsModel_AddResults` | Resultats de comparaison |
| `TestLogsModel_AddResults_WithError` | Resultat en erreur |
| `TestLogsModel_AddFinalResult` | Resultat final |
| `TestLogsModel_AddFinalResult_NilResult` | Resultat nil |
| `TestLogsModel_AddError` | Ajout erreur |
| `TestLogsModel_AlgoName_OutOfBounds` | Mapping index/nom |
| `TestLogsModel_View` | Rendu viewport |
| `TestLogsModel_AutoScroll` | Auto-scroll actif |
| `TestLogsModel_Update_ScrollKeys` | Navigation clavier |
| `TestLogsModel_AddProgressEntry_BoundedGrowth` | Limite 10000 entries |
| `TestLogsModel_algoName_NegativeProducesUnknown` | Index negatif -> Unknown |
| `TestLogsModel_AutoScroll_DisablesOnScrollUp` | Desactivation auto-scroll |
| `TestLogsModel_AddProgressEntry_StressTest` | Test de charge (10000 updates) |
| `TestLogsModel_SetSize` | Dimensions viewport |

### bridge_test.go (14 tests)

| Test | Couverture |
|------|-----------|
| `TestTUIProgressReporter_DrainsChannel` | Drain du canal |
| `TestTUIProgressReporter_ZeroCalculators` | 0 calculateurs |
| `TestTUIResultPresenter_FormatDuration` | Formatage duree |
| `TestProgramRef_Send_NilProgram` | Send sans programme |
| `TestTUIResultPresenter_PresentComparisonTable` | Envoi comparaison |
| `TestTUIResultPresenter_PresentResult` | Envoi resultat |
| `TestTUIResultPresenter_HandleError_Timeout` | Erreur timeout |
| `TestTUIResultPresenter_HandleError_Canceled` | Erreur annulation |
| `TestTUIResultPresenter_HandleError_Generic` | Erreur generique |
| `TestTUIResultPresenter_HandleError_Nil` | Erreur nil |
| `TestTUIProgressReporter_MultipleCalculators` | Multi-calculateurs |
| `TestProgramRef_Send_Concurrent` | Thread safety Send() |
| `TestTUIResultPresenter_HandleError_PassesDuration` | Verification fix duration |
| `TestTUIProgressReporter_EmptyChannel` | Canal vide |

### chart_test.go (11 tests)

| Test | Couverture |
|------|-----------|
| `TestChartModel_AddDataPoint` | Ajout points |
| `TestChartModel_AddDataPoint_Overflow` | Depassement maxPoints |
| `TestChartModel_Reset` | Remise a zero |
| `TestChartModel_RenderSparkline` | Rendu sparkline |
| `TestChartModel_View` | Vue complete |
| `TestChartModel_RenderSparkline_NegativeValues` | Valeurs negatives |
| `TestChartModel_RenderSparkline_OverOneValues` | Valeurs > 1.0 |
| `TestChartModel_RenderSparkline_AllCharacters` | Mapping complet bloc/valeur |
| `TestChartModel_RenderSparkline_Boundaries` | Limites 0.0 et 1.0 |
| `TestChartModel_SetSize_VeryWide` | Terminal tres large |
| `TestChartModel_SetSize_TrimsData` | Trim sur resize |

### metrics_test.go (11 tests)

| Test | Couverture |
|------|-----------|
| `TestMetricsModel_UpdateMemStats` | Mise a jour stats memoire |
| `TestMetricsModel_UpdateProgress` | Calcul vitesse |
| `TestMetricsModel_UpdateProgress_Smoothing` | Lissage EMA |
| `TestMetricsModel_UpdateProgress_TooFast` | dt < 0.05s ignore |
| `TestMetricsModel_UpdateProgress_NoForward` | Pas de progression |
| `TestMetricsModel_View` | Vue complete |
| `TestFormatBytes` | Formatage octets |
| `TestFormatBytes_Boundaries` | Limites KB/MB/GB |
| `TestMetricsModel_UpdateProgress_RapidUpdates` | 1000 updates rapides |
| `TestMetricsModel_SetSize` | Dimensions |
| `TestFormatMetricLine` | Ligne metrique formatee |

### footer_test.go (9 tests)

| Test | Couverture |
|------|-----------|
| `TestFooterModel_View_Running` | Statut Running |
| `TestFooterModel_View_Paused` | Statut Paused |
| `TestFooterModel_View_Done` | Statut Done |
| `TestFooterModel_View_Error` | Statut Error |
| `TestFooterModel_View_ErrorPrecedence` | Priorite Error > Done > Paused |
| `TestFooterModel_View_Shortcuts` | Raccourcis affiches |
| `TestFooterModel_View_NarrowWidth` | Largeur etroite |
| `TestFooterModel_SetWidth_Negative` | Largeur negative |
| `TestFooterModel_SetWidth_Zero` | Largeur zero |

### header_test.go (6 tests)

| Test | Couverture |
|------|-----------|
| `TestHeaderModel_View_ContainsTitle` | Titre FibGo Monitor |
| `TestHeaderModel_View_ContainsVersion` | Version affichee |
| `TestHeaderModel_View_ContainsElapsed` | Temps ecoule |
| `TestHeaderModel_View_NarrowWidth` | Largeur etroite |
| `TestHeaderModel_View_ZeroWidth` | Largeur zero |
| `TestSpaces` | Fonction utilitaire spaces() |

### keymap_test.go (2 tests)

| Test | Couverture |
|------|-----------|
| `TestDefaultKeyMap_AllBindingsDefined` | Tous les bindings definis et actifs |
| `TestDefaultKeyMap_QuitKeys` | Quit contient q et ctrl+c |

---

## 3. Anomalies Identifiees et Corrigees

### Anomalie 1 : Index negatif produit "Algo--1"

- **Fichier :** `internal/tui/logs.go:136-141`
- **Probleme :** `fmt.Sprintf("Algo-%d", index)` avec index=-1 donnait "Algo--1"
- **Correction :** Retourner `"Unknown"` pour index < 0, `fmt.Sprintf("Algo %d", index)` pour index >= len
- **Test :** `TestLogsModel_algoName_NegativeProducesUnknown`

### Anomalie 2 : Croissance illimitee des entrees de log

- **Fichier :** `internal/tui/logs.go:59`
- **Probleme :** `l.entries = append(l.entries, entry)` sans limite, fuite memoire potentielle
- **Correction :** Ajout `const maxLogEntries = 10000` et appel `trimEntries()` dans `AddProgressEntry`, `AddResults`, `AddFinalResult`, `AddError`
- **Test :** `TestLogsModel_AddProgressEntry_BoundedGrowth`

### Anomalie 3 : Styles morts (dead code)

- **Fichier :** `internal/tui/styles.go:47-48, 113-118`
- **Probleme :** `logEntryStyle`, `progressFullStyle`, `progressEmptyStyle` definis mais jamais utilises
- **Correction :** Suppression des 3 styles inutilises
- **Verification :** Compilation sans erreur

### Anomalie 4 : HandleError passe duration=0

- **Fichier :** `internal/tui/bridge.go:95`
- **Probleme :** `apperrors.HandleCalculationError(err, 0, io.Discard, nil)` ignorait la duree reelle
- **Correction :** Remplacement de `0` par `duration`
- **Test :** `TestTUIResultPresenter_HandleError_PassesDuration`

### Anomalie 5 : Nombre magique dans metricsHeight()

- **Fichier :** `internal/tui/model.go:214-220`
- **Probleme :** `m.height - 6` au lieu de constantes, duplication de logique avec `layoutPanels()`
- **Correction :** Extraction de `headerHeight`, `footerHeight`, `minBodyHeight` en constantes package-level, utilisees dans `layoutPanels()` et `metricsHeight()`
- **Test :** `TestModel_metricsHeight_ConsistentWithLayout`

---

## 4. Matrice de Couverture

| Fichier | Avant | Apres |
|---------|-------|-------|
| `bridge.go` | 90% | 94% |
| `chart.go` | 95% | 100% |
| `footer.go` | 100% | 100% |
| `header.go` | 100% | 100% |
| `keymap.go` | 0% | 100% |
| `logs.go` | 90% | 100% |
| `messages.go` | 100% | 100% |
| `metrics.go` | 90% | 100% |
| `model.go` | 85% | 90% |
| `styles.go` | 100% | 100% |
| **Total** | **90.5%** | **95.7%** |

> Note : `model.go` contient `Run()` qui demarre un `tea.Program` reel et ne peut pas etre teste unitairement. Les 4.3% non couverts correspondent principalement a cette fonction.

---

## 5. Fichiers Modifies

| Fichier | Modification |
|---------|-------------|
| `internal/tui/logs.go` | Fix anomalies 1 (algoName) et 2 (maxLogEntries + trimEntries) |
| `internal/tui/styles.go` | Fix anomalie 3 (suppression logEntryStyle, progressFullStyle, progressEmptyStyle) |
| `internal/tui/bridge.go` | Fix anomalie 4 (duration au lieu de 0) |
| `internal/tui/model.go` | Fix anomalie 5 (constantes headerHeight, footerHeight, minBodyHeight) |
| `internal/tui/model_test.go` | +13 nouveaux tests |
| `internal/tui/logs_test.go` | +5 nouveaux tests |
| `internal/tui/bridge_test.go` | +2 nouveaux tests |
| `internal/tui/chart_test.go` | +3 nouveaux tests |
| `internal/tui/metrics_test.go` | +3 nouveaux tests |
| `internal/tui/footer_test.go` | +2 nouveaux tests |
| `internal/tui/keymap_test.go` | Nouveau fichier, 2 tests |
| `Plan-Test-TUI.md` | Ce document |

---

## 6. Procedure de Regression

```bash
# Tests complets (tous les packages)
go test -v -count=1 ./...

# Coverage TUI detaillee
go test -v -cover ./internal/tui/

# Profil de couverture
go test -coverprofile=coverage.out ./internal/tui/
go tool cover -func=coverage.out

# Tests rapides (skip slow/stress tests)
go test -v -short ./...

# Tests avec race detector
go test -v -race ./internal/tui/
```

### Criteres de succes

- 16/16 packages passants
- TUI coverage >= 95%
- Aucun FAIL dans la regression
- Aucun data race detecte
