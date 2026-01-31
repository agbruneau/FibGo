// Package tui provides a Terminal User Interface using Bubbletea.
package tui

import (
	"math/big"
	"time"

	"github.com/agbru/fibcalc/internal/fibonacci"
	"github.com/agbru/fibcalc/internal/orchestration"
)

// View represents the current view/screen in the TUI.
type View int

const (
	ViewHome View = iota
	ViewCalculator
	ViewProgress
	ViewResults
	ViewComparison
	ViewSettings
	ViewHelp
)

// ProgressMsg carries a progress update from a calculator.
type ProgressMsg struct {
	Update fibonacci.ProgressUpdate
}

// ProgressDoneMsg signals that all progress updates have been received.
type ProgressDoneMsg struct{}

// CalculationStartMsg signals the start of a calculation.
type CalculationStartMsg struct {
	N              uint64
	Algorithm      string
	NumCalculators int
}

// CalculationResultMsg carries the result of a single calculation.
type CalculationResultMsg struct {
	Result   orchestration.CalculationResult
	N        uint64
	Duration time.Duration
}

// ComparisonResultsMsg carries all results from a comparison run.
type ComparisonResultsMsg struct {
	Results []orchestration.CalculationResult
	N       uint64
}

// ErrorMsg carries an error that occurred during operation.
type ErrorMsg struct {
	Err error
}

// ThemeChangedMsg signals that the theme has been changed.
type ThemeChangedMsg struct {
	ThemeName string
}

// NavigateMsg requests navigation to a specific view.
type NavigateMsg struct {
	To View
}

// CancelMsg requests cancellation of the current operation.
type CancelMsg struct{}

// TickMsg is sent periodically for animations and updates.
type TickMsg struct {
	Time time.Time
}

// WindowSizeMsg carries terminal window size information.
type WindowSizeMsg struct {
	Width  int
	Height int
}

// InputSubmittedMsg carries user input submission.
type InputSubmittedMsg struct {
	Value string
}

// AlgorithmSelectedMsg carries the selected algorithm.
type AlgorithmSelectedMsg struct {
	Algorithm string
}

// ResultSavedMsg signals that a result was saved to file.
type ResultSavedMsg struct {
	FilePath string
}

// ResultActionMsg carries a result action request.
type ResultActionMsg struct {
	Action ResultAction
}

// ResultAction represents actions that can be taken on a result.
type ResultAction int

const (
	ActionSaveToFile ResultAction = iota
	ActionShowHex
	ActionCopyToClipboard
	ActionNewCalculation
)

// SingleResultState holds the state for a single calculation result view.
type SingleResultState struct {
	Result    *big.Int
	N         uint64
	Duration  time.Duration
	Algorithm string
}
