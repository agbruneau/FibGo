package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/agbru/fibcalc/internal/cli"
	"github.com/agbru/fibcalc/internal/fibonacci"
)

var keys = newKeyMap()

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case CalcDoneMsg:
		if msg.CalcIndex >= 0 && msg.CalcIndex < len(m.algos) {
			a := &m.algos[msg.CalcIndex]
			a.Done = true
			a.Progress = 1.0
			a.Duration = msg.Duration
			a.Result = msg.Result
			a.Err = msg.Err
			if msg.Result != nil {
				a.BitLen = msg.Result.BitLen()
				a.Digits = len(msg.Result.String())
			}
			m.doneCount++

			if m.doneCount >= len(m.algos) {
				m.state = StateResults
				m.statusMsg = "Calculation complete"
				if m.cancelFunc != nil {
					m.cancelFunc()
					m.cancelFunc = nil
				}
			} else {
				m.statusMsg = fmt.Sprintf("%d/%d completed", m.doneCount, len(m.algos))
			}
		}
		return m, nil

	case tickMsg:
		if m.state == StateRunning {
			m.drainProgressChan()
			return m, tickCmd()
		}
		return m, nil
	}

	// Delegate to text inputs
	if m.state == StateConfig {
		return m.delegateToInput(msg)
	}

	return m, nil
}

// drainProgressChan reads all pending progress updates from the shared channel.
func (m *Model) drainProgressChan() {
	if m.progressChan == nil {
		return
	}
	for {
		select {
		case u := <-m.progressChan:
			if u.CalculatorIndex >= 0 && u.CalculatorIndex < len(m.algos) {
				m.algos[u.CalculatorIndex].Progress = u.Value
			}
		default:
			return
		}
	}
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateConfig:
		return m.handleConfigKey(msg)
	case StateRunning:
		return m.handleRunningKey(msg)
	case StateResults:
		return m.handleResultsKey(msg)
	}
	return m, nil
}

func (m Model) handleConfigKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Start):
		return m.startCalculation()

	case key.Matches(msg, keys.Tab):
		m.focus = (m.focus + 1) % focusCount
		m.updateFocus()
		return m, nil

	case key.Matches(msg, keys.Left):
		if m.focus == FocusAlgo {
			if m.algoIndex > 0 {
				m.algoIndex--
			}
			return m, nil
		}
		return m.delegateToInput(msg)

	case key.Matches(msg, keys.Right):
		if m.focus == FocusAlgo {
			if m.algoIndex < len(m.algoChoices)-1 {
				m.algoIndex++
			}
			return m, nil
		}
		return m.delegateToInput(msg)

	default:
		return m.delegateToInput(msg)
	}
}

func (m Model) handleRunningKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Quit) {
		if m.cancelFunc != nil {
			m.cancelFunc()
			m.cancelFunc = nil
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleResultsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, keys.Reset):
		m.state = StateConfig
		m.algos = nil
		m.doneCount = 0
		m.statusMsg = "Ready"
		m.focus = FocusN
		m.updateFocus()
		return m, nil
	}
	return m, nil
}

func (m *Model) updateFocus() {
	m.nInput.Blur()
	m.timeoutInput.Blur()
	switch m.focus {
	case FocusN:
		m.nInput.Focus()
	case FocusTimeout:
		m.timeoutInput.Focus()
	}
}

func (m Model) delegateToInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focus {
	case FocusN:
		m.nInput, cmd = m.nInput.Update(msg)
	case FocusTimeout:
		m.timeoutInput, cmd = m.timeoutInput.Update(msg)
	}
	return m, cmd
}

func (m Model) startCalculation() (tea.Model, tea.Cmd) {
	// Parse N
	nStr := strings.TrimSpace(m.nInput.Value())
	n, err := strconv.ParseUint(nStr, 10, 64)
	if err != nil || n == 0 {
		m.statusMsg = "Invalid N: enter a positive integer"
		return m, nil
	}

	// Parse timeout
	tStr := strings.TrimSpace(m.timeoutInput.Value())
	timeout, err := time.ParseDuration(tStr)
	if err != nil {
		m.statusMsg = "Invalid timeout: use format like 5m, 30s, 1h"
		return m, nil
	}

	// Get calculators
	cfg := cli.ExecutionConfig{
		N:            n,
		Algo:         m.algoChoices[m.algoIndex],
		Timeout:      timeout,
		Threshold:    fibonacci.DefaultParallelThreshold,
		FFTThreshold: fibonacci.DefaultFFTThreshold,
	}
	calculators := cli.GetCalculatorsToRun(cfg, m.factory)
	if len(calculators) == 0 {
		m.statusMsg = "No calculators available for this algorithm"
		return m, nil
	}

	// Initialize state
	m.algos = make([]AlgoProgress, len(calculators))
	for i, c := range calculators {
		m.algos[i] = AlgoProgress{Name: c.Name()}
	}
	m.doneCount = 0
	m.state = StateRunning
	m.startTime = time.Now()
	m.statusMsg = "Calculating..."

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	m.cancelFunc = cancel

	// Shared progress channel — all calculators write here, tick drains it
	m.progressChan = make(chan fibonacci.ProgressUpdate, len(calculators)*100)

	opts := fibonacci.Options{
		ParallelThreshold: cfg.Threshold,
		FFTThreshold:      cfg.FFTThreshold,
	}

	// Build batch: tick + one Cmd per calculator
	cmds := make([]tea.Cmd, 0, len(calculators)+1)
	cmds = append(cmds, tickCmd())
	for i, calc := range calculators {
		cmds = append(cmds, runSingleCalc(calc, i, n, opts, ctx, m.progressChan))
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	var sections []string
	sections = append(sections, m.renderHeader())
	sections = append(sections, m.renderConfig())
	sections = append(sections, m.renderProgress())
	sections = append(sections, m.renderResults())
	sections = append(sections, m.renderStatusBar())

	return strings.Join(sections, "\n")
}
