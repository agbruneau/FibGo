package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/agbru/fibcalc/internal/config"
	"github.com/agbru/fibcalc/internal/fibonacci"
	"github.com/agbru/fibcalc/internal/orchestration"
	"github.com/agbru/fibcalc/internal/ui"
)

// DashboardModel is the root model for the single-screen HTOP-style TUI.
type DashboardModel struct {
	// Configuration
	config      config.AppConfig
	calculators []fibonacci.Calculator
	ctx         context.Context
	cancel      context.CancelFunc

	// UI Framework
	keys   KeyMap
	styles Styles
	help   help.Model
	width  int
	height int
	ready  bool

	// Consolidated dashboard state
	input       InputState
	algorithms  AlgorithmTableState
	calculation CalculationState
	results     ResultsDisplayState

	// Focus and overlays
	focusedSection Section
	helpOverlay    bool

	// Error handling
	lastError error
}

// InputState holds the input section state.
type InputState struct {
	n           string
	cursorPos   int
	inputActive bool
	buttonIndex int // 0 = Calculate, 1 = Compare
}

// AlgorithmTableState holds the algorithm comparison table state.
type AlgorithmTableState struct {
	names      []string
	progresses []float64
	durations  []time.Duration
	statuses   []AlgoStatus
	cursor     int
}

// CalculationState holds the active calculation state.
type CalculationState struct {
	active       bool
	n            uint64
	mode         CalcMode
	startTime    time.Time
	progressChan chan fibonacci.ProgressUpdate
}

// ResultsDisplayState holds the results display state.
type ResultsDisplayState struct {
	hasResults  bool
	results     []orchestration.CalculationResult
	n           uint64
	showHex     bool
	showFull    bool
	showDetails bool // When false, only show bits count; when true, show full result
	cursor      int
	consistent  bool
}

// NewDashboardModel creates a new dashboard model.
func NewDashboardModel(cfg config.AppConfig, calculators []fibonacci.Calculator) DashboardModel {
	ctx, cancel := context.WithCancel(context.Background())

	// Get algorithm names
	algoNames := make([]string, len(calculators))
	progresses := make([]float64, len(calculators))
	durations := make([]time.Duration, len(calculators))
	statuses := make([]AlgoStatus, len(calculators))

	for i, c := range calculators {
		algoNames[i] = c.Name()
		statuses[i] = StatusIdle
	}

	return DashboardModel{
		config:      cfg,
		calculators: calculators,
		ctx:         ctx,
		cancel:      cancel,
		keys:        DefaultKeyMap(),
		styles:      DefaultStyles(),
		help:        help.New(),
		input: InputState{
			n:           fmt.Sprintf("%d", cfg.N),
			cursorPos:   len(fmt.Sprintf("%d", cfg.N)),
			inputActive: true,
		},
		algorithms: AlgorithmTableState{
			names:      algoNames,
			progresses: progresses,
			durations:  durations,
			statuses:   statuses,
			cursor:     0,
		},
		focusedSection: SectionInput,
	}
}

// Init initializes the dashboard model.
func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tickCmd(100*time.Millisecond),
	)
}

// Update handles messages and updates the dashboard model.
func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)
	case ProgressMsg:
		return m.handleProgressUpdate(msg)
	case ProgressDoneMsg:
		return m, nil
	case CalculationResultMsg:
		return m.handleCalculationResult(msg)
	case ComparisonResultsMsg:
		return m.handleComparisonResults(msg)
	case ErrorMsg:
		m.lastError = msg.Err
		return m, nil
	case ThemeChangedMsg:
		return m.handleThemeChanged(msg)
	case TickMsg:
		return m, tickCmd(100 * time.Millisecond)
	}
	return m, nil
}

func (m DashboardModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Help overlay toggle
	if key.Matches(msg, m.keys.Help) {
		m.helpOverlay = !m.helpOverlay
		return m, nil
	}

	// If help overlay is shown, only allow closing it
	if m.helpOverlay {
		if key.Matches(msg, m.keys.Escape) {
			m.helpOverlay = false
		}
		return m, nil
	}

	// Global shortcuts
	switch {
	case key.Matches(msg, m.keys.Quit):
		m.cancel()
		return m, tea.Quit
	case key.Matches(msg, m.keys.Theme):
		return m.cycleTheme()
	case key.Matches(msg, m.keys.Tab):
		m.focusedSection = m.focusedSection.Next()
		m.input.inputActive = (m.focusedSection == SectionInput)
		return m, nil
	case key.Matches(msg, m.keys.ShiftTab):
		m.focusedSection = m.focusedSection.Prev()
		m.input.inputActive = (m.focusedSection == SectionInput)
		return m, nil
	case key.Matches(msg, m.keys.Calculate):
		return m.startSingleCalculation()
	case key.Matches(msg, m.keys.Compare):
		return m.startComparison()
	case key.Matches(msg, m.keys.Escape):
		return m.handleEscape()
	case key.Matches(msg, m.keys.Hex):
		m.results.showHex = !m.results.showHex
		return m, nil
	case key.Matches(msg, m.keys.Full):
		m.results.showFull = !m.results.showFull
		return m, nil
	case key.Matches(msg, m.keys.Details):
		m.results.showDetails = !m.results.showDetails
		return m, nil
	case key.Matches(msg, m.keys.Save):
		return m.saveResult()
	}

	// Section-specific handling
	return m.updateFocusedSection(msg)
}

func (m DashboardModel) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.ready = true
	m.help.Width = msg.Width
	return m, nil
}

func (m DashboardModel) handleProgressUpdate(msg ProgressMsg) (tea.Model, tea.Cmd) {
	idx := msg.Update.CalculatorIndex
	if idx >= 0 && idx < len(m.algorithms.progresses) {
		m.algorithms.progresses[idx] = msg.Update.Value
	}
	// Continue listening for progress
	if m.calculation.progressChan != nil && m.calculation.active {
		return m, listenForProgress(m.calculation.progressChan)
	}
	return m, nil
}

func (m DashboardModel) handleCalculationResult(msg CalculationResultMsg) (tea.Model, tea.Cmd) {
	m.calculation.active = false

	// Update results
	m.results.hasResults = true
	m.results.results = []orchestration.CalculationResult{msg.Result}
	m.results.n = msg.N
	m.results.consistent = true

	// Update algorithm status
	for i, name := range m.algorithms.names {
		if name == msg.Result.Name {
			m.algorithms.statuses[i] = StatusComplete
			m.algorithms.durations[i] = msg.Result.Duration
			m.algorithms.progresses[i] = 1.0
		}
	}

	// Focus results section
	m.focusedSection = SectionResults
	m.input.inputActive = false

	return m, nil
}

func (m DashboardModel) handleComparisonResults(msg ComparisonResultsMsg) (tea.Model, tea.Cmd) {
	m.calculation.active = false

	// Update results
	m.results.hasResults = true
	m.results.results = msg.Results
	m.results.n = msg.N

	// Check consistency
	m.results.consistent = checkConsistency(msg.Results)

	// Update all algorithm statuses
	for i := range m.algorithms.names {
		if i < len(msg.Results) {
			if msg.Results[i].Err != nil {
				m.algorithms.statuses[i] = StatusError
			} else {
				m.algorithms.statuses[i] = StatusComplete
				m.algorithms.durations[i] = msg.Results[i].Duration
			}
			m.algorithms.progresses[i] = 1.0
		}
	}

	// Focus results section
	m.focusedSection = SectionResults
	m.input.inputActive = false

	return m, nil
}

func (m DashboardModel) handleThemeChanged(msg ThemeChangedMsg) (tea.Model, tea.Cmd) {
	ui.SetTheme(msg.ThemeName)
	m.styles.RefreshStyles()
	return m, nil
}

func (m DashboardModel) handleEscape() (tea.Model, tea.Cmd) {
	if m.calculation.active {
		// Cancel current calculation
		m.cancel()
		m.ctx, m.cancel = context.WithCancel(context.Background())
		m.calculation.active = false
		// Reset algorithm statuses
		for i := range m.algorithms.statuses {
			if m.algorithms.statuses[i] == StatusRunning {
				m.algorithms.statuses[i] = StatusIdle
				m.algorithms.progresses[i] = 0
			}
		}
		return m, nil
	}
	return m, nil
}

func (m DashboardModel) cycleTheme() (tea.Model, tea.Cmd) {
	themes := []string{"dark", "light", "none"}
	current := ui.GetCurrentTheme().Name
	nextIdx := 0
	for i, t := range themes {
		if t == current {
			nextIdx = (i + 1) % len(themes)
			break
		}
	}
	return m, func() tea.Msg {
		return ThemeChangedMsg{ThemeName: themes[nextIdx]}
	}
}

func (m DashboardModel) updateFocusedSection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.focusedSection {
	case SectionInput:
		return m.updateInput(msg)
	case SectionAlgorithms:
		return m.updateAlgorithms(msg)
	case SectionResults:
		return m.updateResults(msg)
	}
	return m, nil
}

func (m DashboardModel) saveResult() (tea.Model, tea.Cmd) {
	// TODO: Implement save result to file
	return m, nil
}

// View renders the dashboard.
func (m DashboardModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	// If help overlay is shown, render it on top
	if m.helpOverlay {
		return m.renderHelpOverlay()
	}

	return m.renderDashboard()
}

// checkConsistency verifies that all successful results match.
func checkConsistency(results []orchestration.CalculationResult) bool {
	var firstResult string
	for _, r := range results {
		if r.Err == nil && r.Result != nil {
			if firstResult == "" {
				firstResult = r.Result.String()
			} else if r.Result.String() != firstResult {
				return false
			}
		}
	}
	return true
}
