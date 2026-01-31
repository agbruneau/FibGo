// Package cli provides output utilities for exporting calculation results.
//
// # Naming Conventions
//
// Functions in this package follow consistent naming patterns based on their behavior:
//
//   - Display* functions write formatted output to an [io.Writer].
//     They handle presentation logic and colorization.
//     Examples: [DisplayResult], [DisplayQuietResult], [DisplayProgress].
//
//   - Format* functions return a formatted string without performing I/O.
//     They are pure functions suitable for composition.
//     Examples: [FormatQuietResult], [FormatExecutionDuration].
//
//   - Write* functions write data to files on the filesystem.
//     They handle file creation, directory setup, and error handling.
//     Examples: [WriteResultToFile].
package cli

import (
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/agbru/fibcalc/internal/ui"
)

// OutputConfig holds configuration for result output.
type OutputConfig struct {
	// OutputFile is the path to save the result (empty for no file output).
	OutputFile string
	// HexOutput displays the result in hexadecimal format.
	HexOutput bool
	// Quiet mode suppresses verbose output.
	Quiet bool
	// Verbose shows the full result value.
	Verbose bool
	// Concise enables the calculated value display when true (disabled by default).
	Concise bool
}

// WriteResultToFile writes a calculation result to a file.
//
// Parameters:
//   - result: The calculated Fibonacci number.
//   - n: The index of the Fibonacci number.
//   - duration: The calculation duration.
//   - algo: The algorithm name used.
//   - config: Output configuration.
//
// Returns:
//   - error: An error if the file cannot be written.
func WriteResultToFile(result *big.Int, n uint64, duration time.Duration, algo string, config OutputConfig) error {
	if config.OutputFile == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(config.OutputFile)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	file, err := os.Create(config.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "# Fibonacci Calculation Result\n")
	fmt.Fprintf(file, "# Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "# Algorithm: %s\n", algo)
	fmt.Fprintf(file, "# Duration: %s\n", duration)
	fmt.Fprintf(file, "# N: %d\n", n)
	fmt.Fprintf(file, "# Bits: %d\n", result.BitLen())
	fmt.Fprintf(file, "# Digits: %d\n", len(result.String()))
	fmt.Fprintf(file, "\n")

	// Write result
	if config.HexOutput {
		fmt.Fprintf(file, "F(%d) [hex] =\n0x%s\n", n, result.Text(16))
	} else {
		fmt.Fprintf(file, "F(%d) =\n%s\n", n, result.String())
	}

	return nil
}

// FormatQuietResult formats a result for quiet mode output.
// Returns a single-line result suitable for scripting.
//
// Parameters:
//   - result: The calculated Fibonacci number.
//   - n: The index.
//   - duration: The calculation duration.
//   - hexOutput: Whether to format as hexadecimal.
//
// Returns:
//   - string: The formatted result string.
func FormatQuietResult(result *big.Int, n uint64, duration time.Duration, hexOutput bool) string {
	if hexOutput {
		return fmt.Sprintf("0x%s", result.Text(16))
	}
	return result.String()
}

// DisplayQuietResult outputs a result in quiet mode (minimal output).
//
// Parameters:
//   - out: The output writer.
//   - result: The calculated Fibonacci number.
//   - n: The index.
//   - duration: The calculation duration.
//   - hexOutput: Whether to format as hexadecimal.
func DisplayQuietResult(out io.Writer, result *big.Int, n uint64, duration time.Duration, hexOutput bool) {
	fmt.Fprintln(out, FormatQuietResult(result, n, duration, hexOutput))
}

// DisplayHexResult displays a result in hexadecimal format with optional truncation.
// Uses consistent formatting with color highlighting and truncation based on TruncationLimit.
//
// Parameters:
//   - out: The output writer.
//   - result: The calculated Fibonacci number.
//   - n: The index of the Fibonacci number.
//   - verbose: If true, displays the full hex string without truncation.
func DisplayHexResult(out io.Writer, result *big.Int, n uint64, verbose bool) {
	fmt.Fprintf(out, "\n%s--- Hexadecimal Format ---%s\n", ui.ColorBold(), ui.ColorReset())
	hexStr := result.Text(16)
	if len(hexStr) > TruncationLimit && !verbose {
		fmt.Fprintf(out, "F(%s%d%s) [hex] = %s0x%s...%s%s\n",
			ui.ColorMagenta(), n, ui.ColorReset(),
			ui.ColorGreen(), hexStr[:HexDisplayEdges], hexStr[len(hexStr)-HexDisplayEdges:], ui.ColorReset())
	} else {
		fmt.Fprintf(out, "F(%s%d%s) [hex] = %s0x%s%s\n",
			ui.ColorMagenta(), n, ui.ColorReset(),
			ui.ColorGreen(), hexStr, ui.ColorReset())
	}
}

// DisplayResultWithConfig displays a result with the given output configuration.
// This is a unified function that handles all output modes.
//
// Parameters:
//   - out: The output writer.
//   - result: The calculated Fibonacci number.
//   - n: The index.
//   - duration: The calculation duration.
//   - algo: The algorithm name.
//   - config: Output configuration.
//
// Returns:
//   - error: An error if file output fails.
func DisplayResultWithConfig(out io.Writer, result *big.Int, n uint64, duration time.Duration, algo string, config OutputConfig) error {
	// Handle quiet mode
	if config.Quiet {
		DisplayQuietResult(out, result, n, duration, config.HexOutput)
	} else {
		// Use standard display
		DisplayResult(result, n, duration, config.Verbose, true, config.Concise, out)

		// Show hex format if requested
		if config.HexOutput {
			DisplayHexResult(out, result, n, config.Verbose)
		}
	}

	// Save to file if requested
	if config.OutputFile != "" {
		if err := WriteResultToFile(result, n, duration, algo, config); err != nil {
			return err
		}
		if !config.Quiet {
			fmt.Fprintf(out, "\n%s✓ Result saved to: %s%s%s\n",
				ui.ColorGreen(), ui.ColorCyan(), config.OutputFile, ui.ColorReset())
		}
	}

	return nil
}
