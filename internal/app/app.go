package app

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/agbru/fibcalc/internal/cli"
	"github.com/agbru/fibcalc/internal/config"
	apperrors "github.com/agbru/fibcalc/internal/errors"
	"github.com/agbru/fibcalc/internal/fibonacci"
	"github.com/agbru/fibcalc/internal/orchestration"
	"github.com/agbru/fibcalc/internal/ui"
	"github.com/rs/zerolog"
)

// Application represents the fibcalc application instance.
// It encapsulates the configuration and provides methods to run
// the application in various modes (CLI, REPL).
type Application struct {
	// Config holds the parsed application configuration.
	Config config.AppConfig
	// Factory provides access to the Fibonacci calculator implementations.
	// Uses the interface type for better testability and dependency injection.
	Factory fibonacci.CalculatorFactory
	// ErrWriter is the writer for error output (typically os.Stderr).
	ErrWriter io.Writer
}

// New creates a new Application instance by parsing command-line arguments.
// It validates the configuration and returns an error if parsing or validation fails.
//
// Parameters:
//   - args: The command-line arguments (typically os.Args).
//   - errWriter: The writer for error output.
//
// Returns:
//   - *Application: A new application instance.
//   - error: An error if configuration parsing or validation fails.
func New(args []string, errWriter io.Writer) (*Application, error) {
	factory := fibonacci.GlobalFactory()
	availableAlgos := factory.List()

	// args[0] is program name, args[1:] are the actual arguments
	programName := "fibcalc"
	var cmdArgs []string
	if len(args) > 0 {
		programName = args[0]
		cmdArgs = args[1:]
	}

	cfg, err := config.ParseConfig(programName, cmdArgs, errWriter, availableAlgos)
	if err != nil {
		return nil, err
	}

	return &Application{
		Config:    cfg,
		Factory:   factory,
		ErrWriter: errWriter,
	}, nil
}

// Run executes the application based on the configured mode.
// It dispatches to the appropriate handler (completion, REPL, or CLI).
//
// Parameters:
//   - ctx: The context for managing cancellation and timeouts.
//   - out: The writer for standard output.
//
// Returns:
//   - int: An exit code (0 for success, non-zero for errors).
func (a *Application) Run(ctx context.Context, out io.Writer) int {
	// Handle completion script generation
	if a.Config.Completion != "" {
		return a.runCompletion(out)
	}

	// Disable trace-level logging by default to avoid polluting CLI output.
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Initialize CLI theme (respects --no-color flag and NO_COLOR env var)
	ui.InitTheme(a.Config.NoColor)

	// Interactive REPL mode
	if a.Config.Interactive {
		return a.runREPL()
	}

	// Standard CLI calculation mode
	return a.runCalculate(ctx, out)
}

// runCompletion generates shell completion scripts.
func (a *Application) runCompletion(out io.Writer) int {
	availableAlgos := a.Factory.List()
	if err := cli.GenerateCompletion(out, a.Config.Completion, availableAlgos); err != nil {
		fmt.Fprintf(a.ErrWriter, "Error generating completion: %v\n", err)
		return apperrors.ExitErrorConfig
	}
	return apperrors.ExitSuccess
}

// runREPL starts the interactive REPL mode.
func (a *Application) runREPL() int {
	repl := cli.NewREPL(a.Factory.GetAll(), cli.REPLConfig{
		DefaultAlgo:  a.Config.Algo,
		Timeout:      a.Config.Timeout,
		Threshold:    a.Config.Threshold,
		FFTThreshold: a.Config.FFTThreshold,
		HexOutput:    a.Config.HexOutput,
	})
	repl.Start()
	return apperrors.ExitSuccess
}

// runCalculate orchestrates the execution of the CLI calculation command.
func (a *Application) runCalculate(ctx context.Context, out io.Writer) int {
	// Setup lifecycle (timeout + signals)
	ctx, cancelTimeout := context.WithTimeout(ctx, a.Config.Timeout)
	defer cancelTimeout()
	ctx, stopSignals := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	// Get calculators to run
	calculatorsToRun := cli.GetCalculatorsToRun(a.Config, a.Factory)

	// Skip verbose output in quiet mode
	if !a.Config.JSONOutput && !a.Config.Quiet {
		cli.PrintExecutionConfig(a.Config, out)
		cli.PrintExecutionMode(calculatorsToRun, out)
	}

	// Choose progress reporter based on quiet mode
	var progressReporter orchestration.ProgressReporter
	progressOut := out
	if a.Config.Quiet {
		progressOut = io.Discard
		progressReporter = orchestration.NullProgressReporter{}
	} else {
		progressReporter = cli.CLIProgressReporter{}
	}

	// Execute calculations
	results := orchestration.ExecuteCalculations(ctx, calculatorsToRun, a.Config, progressReporter, progressOut)

	// Handle JSON output
	if a.Config.JSONOutput {
		return printJSONResults(results, out)
	}

	// Build output config for the CLI options
	outputCfg := cli.OutputConfig{
		OutputFile: a.Config.OutputFile,
		HexOutput:  a.Config.HexOutput,
		Quiet:      a.Config.Quiet,
		Verbose:    a.Config.Verbose,
		Concise:    a.Config.Concise,
	}

	return a.analyzeResultsWithOutput(results, outputCfg, out)
}

func (a *Application) analyzeResultsWithOutput(results []orchestration.CalculationResult, outputCfg cli.OutputConfig, out io.Writer) int {
	bestResult := findBestResult(results)

	// Handle quiet mode for single result
	if outputCfg.Quiet && bestResult != nil {
		cli.DisplayQuietResult(out, bestResult.Result, a.Config.N, bestResult.Duration, outputCfg.HexOutput)

		// Save to file if requested
		if err := a.saveResultIfNeeded(bestResult, outputCfg); err != nil {
			return apperrors.ExitErrorGeneric
		}

		return apperrors.ExitSuccess
	}

	// Use standard analysis for non-quiet mode
	exitCode := orchestration.AnalyzeComparisonResults(results, a.Config, cli.CLIResultPresenter{}, out)

	// Handle file output and hex display for non-quiet mode
	if bestResult != nil && exitCode == apperrors.ExitSuccess {
		// Display hex format if requested
		a.displayHexIfNeeded(bestResult, outputCfg, out)

		// Save to file if requested
		if err := a.saveResultIfNeeded(bestResult, outputCfg); err != nil {
			return apperrors.ExitErrorGeneric
		}
		if outputCfg.OutputFile != "" {
			fmt.Fprintf(out, "\n%s✓ Result saved to: %s%s%s\n",
				ui.ColorGreen(), ui.ColorCyan(), outputCfg.OutputFile, ui.ColorReset())
		}
	}

	return exitCode
}

// IsHelpError checks if the error is a help flag error (--help was used).
// This is useful for determining if the application should exit with success
// after displaying help text.
//
// Parameters:
//   - err: The error to check.
//
// Returns:
//   - bool: True if the error indicates help was requested.
func IsHelpError(err error) bool {
	return errors.Is(err, flag.ErrHelp)
}

func findBestResult(results []orchestration.CalculationResult) *orchestration.CalculationResult {
	var bestResult *orchestration.CalculationResult
	for i := range results {
		if results[i].Err == nil {
			if bestResult == nil || results[i].Duration < bestResult.Duration {
				bestResult = &results[i]
			}
		}
	}
	return bestResult
}

func (a *Application) saveResultIfNeeded(res *orchestration.CalculationResult, cfg cli.OutputConfig) error {
	if cfg.OutputFile == "" {
		return nil
	}
	if err := cli.WriteResultToFile(res.Result, a.Config.N, res.Duration, res.Name, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving result: %v\n", err)
		return err
	}
	return nil
}

func (a *Application) displayHexIfNeeded(res *orchestration.CalculationResult, cfg cli.OutputConfig, out io.Writer) {
	if !cfg.HexOutput {
		return
	}
	cli.DisplayHexResult(out, res.Result, a.Config.N, a.Config.Verbose)
}

// jsonResult represents a single calculation result in JSON format.
type jsonResult struct {
	Algorithm string `json:"algorithm"`
	Duration  string `json:"duration"`
	Result    string `json:"result,omitempty"`
	Error     string `json:"error,omitempty"`
}

// printJSONResults formats the calculation results as a JSON array and writes
// them to the output. This is useful for programmatic consumption of the results.
func printJSONResults(results []orchestration.CalculationResult, out io.Writer) int {
	output := make([]jsonResult, len(results))
	for i, res := range results {
		jr := jsonResult{
			Algorithm: res.Name,
			Duration:  res.Duration.String(),
		}
		if res.Err != nil {
			jr.Error = res.Err.Error()
		} else {
			jr.Result = res.Result.String()
		}
		output[i] = jr
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		return apperrors.ExitErrorGeneric
	}
	return apperrors.ExitSuccess
}
