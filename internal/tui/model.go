package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/agbru/fibcalc/internal/config"
	"github.com/agbru/fibcalc/internal/fibonacci"
	"github.com/agbru/fibcalc/internal/orchestration"
)

// Model is the legacy model type for backward compatibility.
// The new HTOP-style dashboard uses DashboardModel instead.
// This type is kept for existing tests and as documentation of the old structure.
type Model struct {
	// Configuration
	config      config.AppConfig
	calculators []fibonacci.Calculator
	ctx         context.Context
	cancel      context.CancelFunc

	// UI state
	currentView View
	prevView    View
	keys        KeyMap
	styles      Styles
	help        help.Model
	width       int
	height      int
	ready       bool

	// View-specific state (legacy)
	homeState       HomeState
	calculatorState CalculatorState
	progressState   ProgressState
	resultsState    ResultsState
	comparisonState ComparisonState
	settingsState   SettingsState
	helpState       HelpState

	// Error handling
	lastError error
}

// HomeState holds state for the home view (legacy).
type HomeState struct {
	cursor int
}

// CalculatorState holds state for the calculator view (legacy).
type CalculatorState struct {
	inputN         string
	selectedAlgo   int
	focusedField   int
	availableAlgos []string
}

// ProgressState holds state for the progress view (legacy).
type ProgressState struct {
	n              uint64
	algorithm      string
	numCalculators int
	progresses     []float64
	startTime      int64
	done           bool
	progressChan   chan fibonacci.ProgressUpdate
}

// ResultsState holds state for the results view (legacy).
type ResultsState struct {
	result   *orchestration.CalculationResult
	n        uint64
	showHex  bool
	showFull bool
}

// ComparisonState holds state for the comparison view (legacy).
type ComparisonState struct {
	results      []orchestration.CalculationResult
	n            uint64
	cursor       int
	showDetails  bool
	progressChan chan fibonacci.ProgressUpdate
	progresses   []float64
	inProgress   bool
}

// SettingsState holds state for the settings view (legacy).
type SettingsState struct {
	cursor       int
	themeIndex   int
	themeOptions []string
}

// HelpState holds state for the help view (legacy).
type HelpState struct {
	scrollOffset int
}

// NewModel creates a new legacy Model.
// This is kept for backward compatibility with tests.
// New code should use NewDashboardModel instead.
func NewModel(cfg config.AppConfig, calculators []fibonacci.Calculator) Model {
	ctx, cancel := context.WithCancel(context.Background())

	// Get algorithm names
	algoNames := make([]string, 0, len(calculators))
	for _, c := range calculators {
		if c != nil {
			algoNames = append(algoNames, c.Name())
		}
	}

	return Model{
		config:      cfg,
		calculators: calculators,
		ctx:         ctx,
		cancel:      cancel,
		currentView: ViewHome,
		keys:        DefaultKeyMap(),
		styles:      DefaultStyles(),
		help:        help.New(),
		homeState: HomeState{
			cursor: 0,
		},
		calculatorState: CalculatorState{
			inputN:         fmt.Sprintf("%d", cfg.N),
			selectedAlgo:   0,
			availableAlgos: append([]string{"all"}, algoNames...),
		},
		settingsState: SettingsState{
			themeOptions: []string{"dark", "light", "none"},
			themeIndex:   0,
		},
	}
}

// Init initializes the model (legacy - not used by new dashboard).
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tickCmd(100*time.Millisecond),
	)
}

// Update handles messages (legacy - not used by new dashboard).
func (m Model) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View renders the model (legacy - not used by new dashboard).
func (m Model) View() string {
	return "Legacy Model - use DashboardModel instead"
}

// maxInt returns the maximum of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// renderProgressBar renders a progress bar (legacy helper for tests).
func renderProgressBar(progress float64, width int, styles Styles) string {
	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	filledStr := ""
	emptyStr := ""
	for i := 0; i < filled; i++ {
		filledStr += "█"
	}
	for i := filled; i < width; i++ {
		emptyStr += "░"
	}

	return styles.ProgressFilled.Render(filledStr) + styles.ProgressEmpty.Render(emptyStr)
}
