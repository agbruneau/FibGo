package tui

import (
	"io"
	"sync"
	"time"

	"github.com/agbru/fibcalc/internal/fibonacci"
	"github.com/agbru/fibcalc/internal/orchestration"
)

// ProgressReporter implements orchestration.ProgressReporter for the TUI.
// It forwards progress updates to a Bubbletea program via messages.
type ProgressReporter struct {
	updateChan chan<- fibonacci.ProgressUpdate
}

// NewProgressReporter creates a new TUI progress reporter.
func NewProgressReporter(updateChan chan<- fibonacci.ProgressUpdate) *ProgressReporter {
	return &ProgressReporter{
		updateChan: updateChan,
	}
}

// DisplayProgress implements orchestration.ProgressReporter.
// It forwards progress updates from the calculation channel to the TUI.
func (r *ProgressReporter) DisplayProgress(wg *sync.WaitGroup, progressChan <-chan fibonacci.ProgressUpdate, _ int, _ io.Writer) {
	defer wg.Done()

	for update := range progressChan {
		// Forward to TUI channel (non-blocking to avoid deadlock)
		select {
		case r.updateChan <- update:
		default:
			// Drop update if channel is full
		}
	}
}

// ResultPresenter implements orchestration.ResultPresenter for the TUI.
// Note: In TUI mode, results are handled via messages rather than direct output.
// This implementation provides the interface but the actual presentation
// is handled by the TUI views.
type ResultPresenter struct{}

// NewResultPresenter creates a new TUI result presenter.
func NewResultPresenter() *ResultPresenter {
	return &ResultPresenter{}
}

// PresentComparisonTable implements orchestration.ResultPresenter.
// In TUI mode, this is a no-op as comparison is shown via the TUI view.
func (p *ResultPresenter) PresentComparisonTable(_ []orchestration.CalculationResult, _ io.Writer) {
	// No-op in TUI mode - handled by viewComparison
}

// PresentResult implements orchestration.ResultPresenter.
// In TUI mode, this is a no-op as results are shown via the TUI view.
func (p *ResultPresenter) PresentResult(_ orchestration.CalculationResult, _ uint64, _, _, _ bool, _ io.Writer) {
	// No-op in TUI mode - handled by viewResults
}

// FormatDuration formats a duration for display.
func (p *ResultPresenter) FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return d.Round(time.Microsecond).String()
	}
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return d.Round(time.Millisecond).String()
}

// HandleError handles calculation errors.
// In TUI mode, errors are displayed via the TUI error state.
func (p *ResultPresenter) HandleError(_ error, _ time.Duration, _ io.Writer) int {
	// Return generic error code - actual error handling in TUI view
	return 1
}

// Verify interface compliance
var (
	_ orchestration.ProgressReporter = (*ProgressReporter)(nil)
	_ orchestration.ResultPresenter  = (*ResultPresenter)(nil)
)
