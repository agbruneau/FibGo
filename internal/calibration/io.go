package calibration

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/agbru/fibcalc/internal/cli"
	"github.com/agbru/fibcalc/internal/config"
	"github.com/agbru/fibcalc/internal/ui"
)

// printCalibrationResults formats and prints the calibration results table.
func printCalibrationResults(out io.Writer, results []calibrationResult, bestThreshold int) {
	fmt.Fprintf(out, "\n--- Calibration Summary ---\n")
	tw := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	fmt.Fprintf(tw, "  %sThreshold%s    │ %sExecution Time%s\n", ui.ColorUnderline(), ui.ColorReset(), ui.ColorUnderline(), ui.ColorReset())
	fmt.Fprintf(tw, "  %s┼%s\n", strings.Repeat("─", 14), strings.Repeat("─", 25))
	for _, res := range results {
		thresholdLabel := fmt.Sprintf("%d bits", res.Threshold)
		if res.Threshold == 0 {
			thresholdLabel = "Sequential"
		}
		durationStr := fmt.Sprintf("%sN/A%s", ui.ColorRed(), ui.ColorReset())
		if res.Err == nil {
			durationStr = cli.FormatExecutionDuration(res.Duration)
			if res.Duration == 0 {
				durationStr = "< 1µs"
			}
		}
		highlight := ""
		if res.Threshold == bestThreshold && res.Err == nil {
			highlight = fmt.Sprintf(" %s(Optimal)%s", ui.ColorGreen(), ui.ColorReset())
		}
		fmt.Fprintf(tw, "  %s%-12s%s │ %s%s%s%s\n", ui.ColorCyan(), thresholdLabel, ui.ColorReset(), ui.ColorYellow(), durationStr, ui.ColorReset(), highlight)
	}
	tw.Flush()
}

// printCalibrationOutput prints the calibration results.
//
// Parameters:
//   - cfg: The updated configuration with calibration results.
//   - out: The writer for output.
func printCalibrationOutput(cfg config.AppConfig, out io.Writer) {
	fmt.Fprintf(out, "%sAuto-calibration%s: parallelism=%s%d%s bits, FFT=%s%d%s bits, Strassen=%s%d%s bits\n",
		ui.ColorGreen(), ui.ColorReset(),
		ui.ColorYellow(), cfg.Threshold, ui.ColorReset(),
		ui.ColorYellow(), cfg.FFTThreshold, ui.ColorReset(),
		ui.ColorYellow(), cfg.StrassenThreshold, ui.ColorReset())
}
