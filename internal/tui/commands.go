package tui

import (
	"context"
	"io"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/agbru/fibcalc/internal/config"
	"github.com/agbru/fibcalc/internal/fibonacci"
	"github.com/agbru/fibcalc/internal/orchestration"
)

// listenForProgress creates a command that listens for progress updates.
// It returns one message per progress update from the channel.
func listenForProgress(ch <-chan fibonacci.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		update, ok := <-ch
		if !ok {
			return ProgressDoneMsg{}
		}
		return ProgressMsg{Update: update}
	}
}

// tickCmd creates a command that sends a tick message after a delay.
func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return TickMsg{Time: t}
	})
}

// runCalculation runs a single calculation and returns the result.
func runCalculation(ctx context.Context, calc fibonacci.Calculator, n uint64, opts fibonacci.Options, progressChan chan<- fibonacci.ProgressUpdate, calcIndex int) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		result, err := calc.Calculate(ctx, progressChan, calcIndex, n, opts)
		duration := time.Since(start)

		close(progressChan)

		if err != nil {
			return ErrorMsg{Err: err}
		}

		return CalculationResultMsg{
			Result: orchestration.CalculationResult{
				Name:     calc.Name(),
				Result:   result,
				Duration: duration,
				Err:      nil,
			},
			N:        n,
			Duration: duration,
		}
	}
}

// runComparison runs all calculators and returns comparison results.
func runComparison(ctx context.Context, calculators []fibonacci.Calculator, cfg config.AppConfig, progressChan chan<- fibonacci.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		// Use a custom progress reporter that forwards to our channel
		reporter := &channelProgressReporter{ch: progressChan}

		results := orchestration.ExecuteCalculations(ctx, calculators, cfg, reporter, nil)

		return ComparisonResultsMsg{
			Results: results,
			N:       cfg.N,
		}
	}
}

// channelProgressReporter forwards progress updates to a channel.
type channelProgressReporter struct {
	ch chan<- fibonacci.ProgressUpdate
}

// DisplayProgress implements orchestration.ProgressReporter.
func (r *channelProgressReporter) DisplayProgress(wg *sync.WaitGroup, progressChan <-chan fibonacci.ProgressUpdate, _ int, _ io.Writer) {
	defer wg.Done()
	for update := range progressChan {
		// Forward to our TUI channel (non-blocking to avoid deadlock)
		select {
		case r.ch <- update:
		default:
			// Drop update if channel is full
		}
	}
}
