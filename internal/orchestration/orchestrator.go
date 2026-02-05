package orchestration

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/agbru/fibcalc/internal/config"
	apperrors "github.com/agbru/fibcalc/internal/errors"
	"github.com/agbru/fibcalc/internal/fibonacci"
)

// CalculationResult encapsulates the outcome of a single Fibonacci calculation.
// It serves as a standardized container for results from different algorithms,
// facilitating comparison and reporting.
type CalculationResult struct {
	// Name is the identifier of the algorithm used (e.g., "Fast Doubling").
	Name string
	// Result is the computed Fibonacci number. It is nil if an error occurred.
	Result *big.Int
	// Duration is the time taken to complete the calculation.
	Duration time.Duration
	// Err contains any error that occurred during the calculation.
	Err error
}

// ProgressBufferMultiplier defines the buffer size multiplier for the progress
// channel. A larger buffer reduces the likelihood of blocking calculation
// goroutines when the UI is slow to consume updates.
const ProgressBufferMultiplier = 5

// ExecuteCalculations orchestrates the concurrent execution of one or more
// Fibonacci calculations.
//
// It manages the lifecycle of calculation goroutines, collects their results,
// and coordinates the display of progress updates. This function is the core of
// the application's concurrency model.
//
// Parameters:
//   - ctx: The context for managing cancellation and deadlines.
//   - calculators: A slice of calculators to execute.
//   - cfg: The application configuration (N, thresholds, etc.).
//   - progressReporter: The progress reporter for displaying updates (use NullProgressReporter for quiet mode).
//   - out: The io.Writer for displaying progress updates.
//
// Returns:
//   - []CalculationResult: A slice containing the results of each calculation.
func ExecuteCalculations(ctx context.Context, calculators []fibonacci.Calculator, cfg config.AppConfig, progressReporter ProgressReporter, out io.Writer) []CalculationResult {
	g, ctx := errgroup.WithContext(ctx)
	results := make([]CalculationResult, len(calculators))
	progressChan := make(chan fibonacci.ProgressUpdate, len(calculators)*ProgressBufferMultiplier)

	var displayWg sync.WaitGroup
	displayWg.Add(1)
	go progressReporter.DisplayProgress(&displayWg, progressChan, len(calculators), out)

	for i, calc := range calculators {
		idx, calculator := i, calc
		g.Go(func() error {
			startTime := time.Now()
			res, err := calculator.Calculate(ctx, progressChan, idx, cfg.N, cfg.ToCalculationOptions())
			results[idx] = CalculationResult{
				Name: calculator.Name(), Result: res, Duration: time.Since(startTime), Err: err,
			}
			return nil
		})
	}

	g.Wait()
	close(progressChan)
	displayWg.Wait()

	return results
}

// AnalyzeComparisonResults processes the results from multiple algorithms and
// generates a summary report.
//
// It sorts the results by execution time, validates consistency across
// successful calculations, and displays a comparative table. It handles the
// logic for determining global success or failure based on the individual
// outcomes.
//
// Parameters:
//   - results: The slice of calculation results to analyze.
//   - cfg: The application configuration.
//   - presenter: The result presenter for display formatting.
//   - out: The io.Writer for the summary report.
//
// Returns:
//   - int: An exit code indicating success (0) or the type of failure.
func AnalyzeComparisonResults(results []CalculationResult, cfg config.AppConfig, presenter ResultPresenter, out io.Writer) int {
	sort.Slice(results, func(i, j int) bool {
		if (results[i].Err == nil) != (results[j].Err == nil) {
			return results[i].Err == nil
		}
		return results[i].Duration < results[j].Duration
	})

	var firstValidResult *CalculationResult
	var firstError error
	successCount := 0

	for i := range results {
		if results[i].Err != nil {
			if firstError == nil {
				firstError = results[i].Err
			}
		} else {
			successCount++
			if firstValidResult == nil {
				firstValidResult = &results[i]
			}
		}
	}

	// Present the comparison table
	presenter.PresentComparisonTable(results, out)

	if successCount == 0 {
		fmt.Fprintf(out, "\nGlobal Status: Failure. No algorithm could complete the calculation.\n")
		return presenter.HandleError(firstError, 0, out)
	}

	mismatch := false
	for _, res := range results {
		if res.Err == nil && res.Result.Cmp(firstValidResult.Result) != 0 {
			mismatch = true
			break
		}
	}
	if mismatch {
		fmt.Fprintf(out, "\nGlobal Status: CRITICAL ERROR! An inconsistency was detected between the results of the algorithms.")
		return apperrors.ExitErrorMismatch
	}

	fmt.Fprintf(out, "\nGlobal Status: Success. All valid results are consistent.\n")
	presenter.PresentResult(*firstValidResult, cfg.N, cfg.Verbose, cfg.Details, cfg.ShowValue, out)
	return apperrors.ExitSuccess
}
