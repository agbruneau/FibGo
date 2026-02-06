// Display functions for progress reporting and result presentation.

package cli

import (
	"fmt"
	"io"
	"math/big"
	"sync"
	"time"

	"github.com/agbru/fibcalc/internal/fibonacci"
	"github.com/agbru/fibcalc/internal/ui"
	"github.com/briandowns/spinner"
)

// DisplayProgress manages the asynchronous display of a spinner and progress bar.
// It is designed to run in a dedicated goroutine and orchestrates the UI updates
// for the duration of the calculations.
//
// The function's responsibilities include:
//   - Receiving progress updates from a channel.
//   - Aggregating these updates to calculate the average progress.
//   - Calculating and displaying the estimated time remaining (ETA).
//   - Periodically refreshing the spinner and progress bar.
//   - Gracefully shutting down when the progress channel is closed.
//
// Parameters:
//   - wg: A WaitGroup to signal when the display routine is complete.
//   - progressChan: The channel receiving progress updates.
//   - numCalculators: The number of calculators contributing to the progress.
//   - out: The io.Writer to which the progress bar is rendered.
func DisplayProgress(wg *sync.WaitGroup, progressChan <-chan fibonacci.ProgressUpdate, numCalculators int, out io.Writer) {
	defer wg.Done()
	if numCalculators <= 0 {
		for range progressChan { // Drain the channel
		}
		return
	}

	state := NewProgressWithETA(numCalculators)
	s := newSpinner(spinner.WithWriter(out))
	s.Start()
	spinnerStopped := false
	defer func() {
		if !spinnerStopped {
			s.Stop()
		}
	}()

	ticker := time.NewTicker(ProgressRefreshRate)
	defer ticker.Stop()

	for {
		select {
		case update, ok := <-progressChan:
			if !ok {
				// Stop the spinner first to free the line
				if !spinnerStopped {
					s.Stop()
					spinnerStopped = true
				}

				// Display final 100% progress permanently by printing directly to output
				bar := progressBar(1.0, ProgressBarWidth)
				label := "Progress"
				if numCalculators > 1 {
					label = "Avg progress"
				}
				// Print the final progress line with a newline so it persists
				fmt.Fprintf(out, "%s: %6.2f%% [%s] ETA: %s\n", label, 100.0, bar, "< 1s")
				return
			}
			state.UpdateWithETA(update.CalculatorIndex, update.Value)
		case <-ticker.C:
			avgProgress := state.CalculateAverage()
			eta := state.GetETA()
			bar := progressBar(avgProgress, ProgressBarWidth)
			label := "Progress"
			if numCalculators > 1 {
				label = "Avg progress"
			}
			etaStr := FormatETA(eta)
			s.UpdateSuffix(fmt.Sprintf(" %s: %6.2f%% [%s] ETA: %s", label, avgProgress*100, bar, etaStr))
		}
	}
}

// displayResultHeader prints the binary size of the result.
//
// Parameters:
//   - out: The io.Writer for the output.
//   - bitLen: The number of bits in the result.
func displayResultHeader(out io.Writer, bitLen int) {
	fmt.Fprintf(out, "Result binary size: %s%s%s bits.\n",
		ui.ColorCyan(), formatNumberString(fmt.Sprintf("%d", bitLen)), ui.ColorReset())
}

// displayDetailedAnalysis prints detailed execution metrics including
// calculation time, number of digits, and scientific notation for large numbers.
//
// Parameters:
//   - out: The io.Writer for the output.
//   - result: The calculation result.
//   - duration: The time taken for the calculation.
func displayDetailedAnalysis(out io.Writer, result *big.Int, duration time.Duration) {
	fmt.Fprintf(out, "\n%s--- Detailed result analysis ---%s\n", ui.ColorBold(), ui.ColorReset())

	durationStr := FormatExecutionDuration(duration)
	if duration == 0 {
		durationStr = "< 1µs"
	}
	fmt.Fprintf(out, "Calculation time        : %s%s%s\n", ui.ColorGreen(), durationStr, ui.ColorReset())

	resultStr := result.String()
	numDigits := len(resultStr)
	fmt.Fprintf(out, "Number of digits      : %s%s%s\n",
		ui.ColorCyan(), formatNumberString(fmt.Sprintf("%d", numDigits)), ui.ColorReset())

	if numDigits > 6 {
		f := new(big.Float).SetInt(result)
		fmt.Fprintf(out, "Scientific notation    : %s%.6e%s\n", ui.ColorCyan(), f, ui.ColorReset())
	}
}

// displayCalculatedValue prints the Fibonacci value, truncating if necessary.
//
// Parameters:
//   - out: The io.Writer for the output.
//   - result: The calculation result.
//   - n: The index of the Fibonacci number calculated.
//   - verbose: If true, prints the full number regardless of size.
func displayCalculatedValue(out io.Writer, result *big.Int, n uint64, verbose bool) {
	resultStr := result.String()
	numDigits := len(resultStr)

	fmt.Fprintf(out, "\n%s--- Calculated value ---%s\n", ui.ColorBold(), ui.ColorReset())

	if verbose {
		fmt.Fprintf(out, "F(%s%d%s) =\n%s%s%s\n",
			ui.ColorMagenta(), n, ui.ColorReset(),
			ui.ColorGreen(), formatNumberString(resultStr), ui.ColorReset())
		return
	}

	if numDigits > TruncationLimit {
		fmt.Fprintf(out, "F(%s%d%s) (truncated) = %s%s...%s%s\n",
			ui.ColorMagenta(), n, ui.ColorReset(),
			ui.ColorGreen(), resultStr[:DisplayEdges], resultStr[numDigits-DisplayEdges:], ui.ColorReset())
		fmt.Fprintf(out, "(Tip: use the %s-v%s or %s--verbose%s option to display the full value)\n",
			ui.ColorYellow(), ui.ColorReset(), ui.ColorYellow(), ui.ColorReset())
		return
	}

	fmt.Fprintf(out, "F(%s%d%s) = %s%s%s\n",
		ui.ColorMagenta(), n, ui.ColorReset(),
		ui.ColorGreen(), formatNumberString(resultStr), ui.ColorReset())
}

// DisplayResult formats and prints the final calculation result.
// It provides different levels of detail based on the verbose and details flags,
// including metadata like binary size, number of digits, and scientific
// notation. For very large numbers, it truncates the output unless verbose is
// true.
//
// Parameters:
//   - result: The calculation result.
//   - n: The index of the Fibonacci number calculated.
//   - duration: The time taken for the calculation.
//   - verbose: If true, prints the full number regardless of size.
//   - details: If true, prints detailed execution metrics.
//   - showValue: If true, displays the calculated value section (disabled by default).
//   - out: The io.Writer for the output.
func DisplayResult(result *big.Int, n uint64, duration time.Duration, verbose, details, showValue bool, out io.Writer) {
	displayResultHeader(out, result.BitLen())

	if details {
		displayDetailedAnalysis(out, result, duration)
	}

	if showValue {
		displayCalculatedValue(out, result, n, verbose)
	}
}
