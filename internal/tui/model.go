package tui

import (
	"context"
	"io"
	"runtime"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/agbru/fibcalc/internal/config"
	apperrors "github.com/agbru/fibcalc/internal/errors"
	"github.com/agbru/fibcalc/internal/fibonacci"
	"github.com/agbru/fibcalc/internal/metrics"
	"github.com/agbru/fibcalc/internal/orchestration"
	"github.com/agbru/fibcalc/internal/sysmon"
)

// Model is the root bubbletea model for the TUI dashboard.
type Model struct {
	header  HeaderModel
	logs    LogsModel
	metrics MetricsModel
	chart   ChartModel
	footer  FooterModel

	keymap KeyMap

	parentCtx   context.Context
	ctx         context.Context
	cancel      context.CancelFunc
	config      config.AppConfig
	calculators []fibonacci.Calculator
	generation  uint64

	ref *programRef

	width    int
	height   int
	paused   bool
	done     bool
	exitCode int
}

// NewModel creates a new TUI model.
func NewModel(parentCtx context.Context, calculators []fibonacci.Calculator, cfg config.AppConfig, version string) Model {
	algoNames := make([]string, len(calculators))
	for i, c := range calculators {
		algoNames[i] = c.Name()
	}

	ctx, cancel := context.WithCancel(parentCtx)

	logs := NewLogsModel(algoNames)
	logs.AddExecutionConfig(cfg)

	return Model{
		header:      NewHeaderModel(version),
		logs:        logs,
		metrics:     NewMetricsModel(),
		chart:       NewChartModel(),
		footer:      NewFooterModel(),
		keymap:      DefaultKeyMap(),
		parentCtx:   parentCtx,
		ctx:         ctx,
		cancel:      cancel,
		config:      cfg,
		calculators: calculators,
		ref:         &programRef{},
		exitCode:    apperrors.ExitSuccess,
	}
}

// Init returns the initial commands.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		startCalculationCmd(m.ref, m.ctx, m.calculators, m.config, m.generation),
		watchContextCmd(m.ctx, m.generation),
	)
}

// Update handles all incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layoutPanels()
		return m, nil

	case ProgressMsg:
		if !m.paused {
			m.logs.AddProgressEntry(msg)
			m.chart.AddDataPoint(msg.Value, msg.AverageProgress, msg.ETA)
			m.metrics.UpdateProgress(msg.AverageProgress)
			// Refresh live indicators from progress data
			elapsed := time.Since(m.header.startTime)
			m.metrics.UpdateIndicators(metrics.ComputeLive(m.config.N, msg.AverageProgress, elapsed))
		}
		return m, nil

	case ProgressDoneMsg:
		return m, nil

	case ComparisonResultsMsg:
		m.logs.AddResults(msg.Results)
		return m, nil

	case FinalResultMsg:
		m.logs.AddFinalResult(msg)
		// Compute indicators asynchronously to avoid blocking the UI
		if msg.Result.Result != nil {
			return m, computeIndicatorsCmd(msg)
		}
		return m, nil

	case IndicatorsMsg:
		m.metrics.UpdateIndicators(msg.Indicators)
		return m, nil

	case ErrorMsg:
		m.logs.AddError(msg)
		m.footer.SetError(true)
		m.done = true
		m.header.SetDone()
		m.footer.SetDone(true)
		return m, nil

	case TickMsg:
		if m.done {
			return m, nil
		}
		if !m.paused {
			return m, tea.Batch(sampleMemStatsCmd(), sampleSysStatsCmd(), tickCmd())
		}
		return m, tickCmd()

	case MemStatsMsg:
		m.metrics.UpdateMemStats(msg)
		return m, nil

	case SysStatsMsg:
		m.chart.UpdateSysStats(msg.CPUPercent, msg.MemPercent)
		return m, nil

	case CalculationCompleteMsg:
		if msg.Generation != m.generation {
			return m, nil // stale message from previous calculation
		}
		m.done = true
		m.exitCode = msg.ExitCode
		m.header.SetDone()
		m.chart.SetDone(time.Since(m.header.startTime))
		m.footer.SetDone(true)
		return m, nil

	case ContextCancelledMsg:
		if msg.Generation != m.generation {
			return m, nil // stale message from previous calculation
		}
		m.done = true
		m.header.SetDone()
		m.footer.SetDone(true)
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Quit):
		if m.cancel != nil {
			m.cancel()
		}
		return m, tea.Quit

	case key.Matches(msg, m.keymap.Pause):
		m.paused = !m.paused
		m.footer.SetPaused(m.paused)
		return m, nil

	case key.Matches(msg, m.keymap.Reset):
		// Cancel the current calculation
		if m.cancel != nil {
			m.cancel()
		}

		// Create a new context for the restarted calculation
		m.generation++
		ctx, cancel := context.WithCancel(m.parentCtx)
		m.ctx = ctx
		m.cancel = cancel

		// Reset all UI components
		m.header.Reset()
		m.logs.Reset()
		m.chart.Reset()
		m.metrics = NewMetricsModel()
		m.metrics.SetSize(m.metricsWidth(), m.metricsHeight())
		m.footer.SetDone(false)
		m.footer.SetError(false)
		m.footer.SetPaused(false)
		m.done = false
		m.paused = false
		m.exitCode = apperrors.ExitSuccess

		// Restart calculation and watchers
		return m, tea.Batch(
			tickCmd(),
			startCalculationCmd(m.ref, m.ctx, m.calculators, m.config, m.generation),
			watchContextCmd(m.ctx, m.generation),
		)

	case key.Matches(msg, m.keymap.Up), key.Matches(msg, m.keymap.Down),
		key.Matches(msg, m.keymap.PageUp), key.Matches(msg, m.keymap.PageDown):
		m.logs.Update(msg)
		return m, nil
	}

	return m, nil
}

// View renders the entire dashboard.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	header := m.header.View()
	footer := m.footer.View()

	metrics := m.metrics.View()
	chart := m.chart.View()

	// Right column: metrics on top, chart on bottom
	rightCol := lipgloss.JoinVertical(lipgloss.Left, metrics, chart)

	// Render logs panel to match the right column's actual height
	logs := m.logs.renderToHeight(lipgloss.Height(rightCol))

	// Main body: logs on left, right column on right
	body := lipgloss.JoinHorizontal(lipgloss.Top, logs, rightCol)

	// Full layout: header + body + footer
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

// Layout constants for the TUI dashboard.
const (
	headerHeight   = 1
	footerHeight   = 1
	minBodyHeight  = 4
	metricsFixedH  = 7 // compact: top line + 1 data row + borders; expands to ~9 with indicators
)

func (m *Model) layoutPanels() {
	bodyHeight := m.height - headerHeight - footerHeight
	if bodyHeight < minBodyHeight {
		bodyHeight = minBodyHeight
	}

	logsWidth := m.width * 60 / 100
	rightWidth := m.width - logsWidth

	metricsH := metricsFixedH
	if metricsH > bodyHeight/2 {
		metricsH = bodyHeight / 2
	}
	chartH := bodyHeight - metricsH

	m.header.SetWidth(m.width)
	m.footer.SetWidth(m.width)
	m.logs.SetSize(logsWidth, bodyHeight)
	m.metrics.SetSize(rightWidth, metricsH)
	m.chart.SetSize(rightWidth, chartH)
}

func (m Model) metricsWidth() int {
	return m.width - m.width*60/100
}

func (m Model) metricsHeight() int {
	bodyHeight := m.height - headerHeight - footerHeight
	if bodyHeight < minBodyHeight {
		bodyHeight = minBodyHeight
	}
	metricsH := metricsFixedH
	if metricsH > bodyHeight/2 {
		metricsH = bodyHeight / 2
	}
	return metricsH
}

// Run is the public entry point for the TUI mode.
// It creates the bubbletea program, runs it, and returns the exit code.
func Run(ctx context.Context, calculators []fibonacci.Calculator, cfg config.AppConfig, version string) int {
	model := NewModel(ctx, calculators, cfg, version)
	defer model.cancel()

	p := tea.NewProgram(model, tea.WithAltScreen())
	// Inject the program reference before running so bridge goroutines can Send.
	model.ref.program = p

	finalModel, err := p.Run()
	if err != nil {
		return apperrors.ExitErrorGeneric
	}

	if m, ok := finalModel.(Model); ok {
		m.cancel()
		return m.exitCode
	}
	return apperrors.ExitSuccess
}

// startCalculationCmd returns a tea.Cmd that launches the orchestration.
func startCalculationCmd(ref *programRef, ctx context.Context, calculators []fibonacci.Calculator, cfg config.AppConfig, gen uint64) tea.Cmd {
	return func() tea.Msg {
		progressReporter := &TUIProgressReporter{ref: ref}
		presenter := &TUIResultPresenter{ref: ref}

		results := orchestration.ExecuteCalculations(ctx, calculators, cfg, progressReporter, io.Discard)
		exitCode := orchestration.AnalyzeComparisonResults(results, cfg, presenter, io.Discard)

		return CalculationCompleteMsg{ExitCode: exitCode, Generation: gen}
	}
}

// tickCmd returns a command that sends a TickMsg after 500ms.
func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// sampleMemStatsCmd reads runtime memory stats and returns a MemStatsMsg.
func sampleMemStatsCmd() tea.Cmd {
	return func() tea.Msg {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		return MemStatsMsg{
			Alloc:        ms.Alloc,
			NumGC:        ms.NumGC,
			NumGoroutine: runtime.NumGoroutine(),
		}
	}
}

// sampleSysStatsCmd reads system-wide CPU and memory stats and returns a SysStatsMsg.
func sampleSysStatsCmd() tea.Cmd {
	return func() tea.Msg {
		s := sysmon.Sample()
		return SysStatsMsg{
			CPUPercent: s.CPUPercent,
			MemPercent: s.MemPercent,
		}
	}
}

// computeIndicatorsCmd returns a tea.Cmd that computes post-calculation
// indicators asynchronously, ensuring no impact on the UI thread.
func computeIndicatorsCmd(msg FinalResultMsg) tea.Cmd {
	return func() tea.Msg {
		ind := metrics.Compute(msg.Result.Result, msg.N, msg.Result.Duration)
		return IndicatorsMsg{Indicators: ind}
	}
}

// watchContextCmd waits for context cancellation and sends a message.
func watchContextCmd(ctx context.Context, gen uint64) tea.Cmd {
	return func() tea.Msg {
		<-ctx.Done()
		return ContextCancelledMsg{Err: ctx.Err(), Generation: gen}
	}
}
