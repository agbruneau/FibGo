package orchestration

import (
	"io"
	"sync"
	"time"

	"github.com/agbru/fibcalc/internal/fibonacci"
)

// ProgressReporter defines the interface for displaying calculation progress.
// This interface decouples the orchestration layer from the presentation layer,
// following Clean Architecture principles where business logic should not
// depend on UI concerns.
//
// Implementations handle the visual representation of progress (spinners,
// progress bars, etc.) while the orchestration layer focuses on coordinating
// the calculations.
type ProgressReporter interface {
	// DisplayProgress starts displaying progress updates from the channel.
	// It should be called in a separate goroutine and will run until the
	// progressChan is closed.
	//
	// Parameters:
	//   - wg: A WaitGroup to signal when display is complete.
	//   - progressChan: Channel receiving progress updates from calculators.
	//   - numCalculators: The number of concurrent calculators being tracked.
	//   - out: The writer for progress output.
	DisplayProgress(wg *sync.WaitGroup, progressChan <-chan fibonacci.ProgressUpdate, numCalculators int, out io.Writer)
}

// ProgressReporterFunc is a function adapter that implements ProgressReporter.
// This allows passing a function directly where a ProgressReporter is expected.
type ProgressReporterFunc func(wg *sync.WaitGroup, progressChan <-chan fibonacci.ProgressUpdate, numCalculators int, out io.Writer)

// DisplayProgress calls the underlying function.
func (f ProgressReporterFunc) DisplayProgress(wg *sync.WaitGroup, progressChan <-chan fibonacci.ProgressUpdate, numCalculators int, out io.Writer) {
	f(wg, progressChan, numCalculators, out)
}

// NullProgressReporter is a no-op implementation of ProgressReporter.
// It drains the progress channel without displaying anything.
// Useful for quiet mode or testing.
type NullProgressReporter struct{}

// DisplayProgress drains the channel without output.
func (NullProgressReporter) DisplayProgress(wg *sync.WaitGroup, progressChan <-chan fibonacci.ProgressUpdate, _ int, _ io.Writer) {
	defer wg.Done()
	for range progressChan {
		// Drain channel silently
	}
}

// ResultPresenter defines the interface for presenting calculation results.
// This interface decouples the orchestration layer from presentation concerns,
// allowing different output formats (CLI, JSON, etc.) without modifying
// the orchestration logic.
type ResultPresenter interface {
	// PresentComparisonTable displays the comparison summary table with
	// algorithm names, durations, and status.
	//
	// Parameters:
	//   - results: The sorted calculation results to display.
	//   - out: The writer for output.
	PresentComparisonTable(results []CalculationResult, out io.Writer)

	// PresentResult displays the final calculation result.
	//
	// Parameters:
	//   - result: The calculation result to display.
	//   - n: The Fibonacci index calculated.
	//   - verbose: Whether to display full result.
	//   - details: Whether to display detailed metrics.
	//   - showValue: Whether to display the calculated value section.
	//   - out: The writer for output.
	PresentResult(result CalculationResult, n uint64, verbose, details, showValue bool, out io.Writer)

	// FormatDuration formats a duration for display.
	//
	// Parameters:
	//   - d: The duration to format.
	//
	// Returns:
	//   - string: The formatted duration string.
	FormatDuration(d time.Duration) string

	// HandleError handles calculation errors and returns an appropriate exit code.
	//
	// Parameters:
	//   - err: The error that occurred.
	//   - duration: The duration of the calculation attempt.
	//   - out: The writer for error output.
	//
	// Returns:
	//   - int: The exit code for the error.
	HandleError(err error, duration time.Duration, out io.Writer) int
}
