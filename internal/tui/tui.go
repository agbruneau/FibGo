// Package tui provides a Terminal User Interface using Bubbletea.
// It implements the ProgressReporter and ResultPresenter interfaces
// from the orchestration package, providing an interactive TUI mode
// for the Fibonacci calculator.
package tui

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/agbru/fibcalc/internal/config"
	"github.com/agbru/fibcalc/internal/fibonacci"
)

// Run starts the TUI application with the given configuration.
// It returns an exit code (0 for success, non-zero for errors).
func Run(cfg config.AppConfig, calculatorMap map[string]fibonacci.Calculator) int {
	// Convert map to slice for internal use
	calculators := make([]fibonacci.Calculator, 0, len(calculatorMap))
	for _, c := range calculatorMap {
		calculators = append(calculators, c)
	}

	// In TUI mode, disable detailed calculation by default
	cfg.Details = false

	// Use the new HTOP-style dashboard model
	model := NewDashboardModel(cfg, calculators)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		return 1
	}

	return 0
}

// RunWithOutput starts the TUI application with custom output writers.
// This is useful for testing or redirecting output.
func RunWithOutput(cfg config.AppConfig, calculatorMap map[string]fibonacci.Calculator, out io.Writer) int {
	// Convert map to slice for internal use
	calculators := make([]fibonacci.Calculator, 0, len(calculatorMap))
	for _, c := range calculatorMap {
		calculators = append(calculators, c)
	}

	// In TUI mode, disable detailed calculation by default
	cfg.Details = false

	// Use the new HTOP-style dashboard model
	model := NewDashboardModel(cfg, calculators)

	opts := []tea.ProgramOption{
		tea.WithAltScreen(),
	}

	// Only add output option if it's not stdout
	if out != os.Stdout {
		opts = append(opts, tea.WithOutput(out))
	}

	p := tea.NewProgram(model, opts...)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		return 1
	}

	return 0
}
