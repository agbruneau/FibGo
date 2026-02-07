package cli

import (
	"time"

	"github.com/briandowns/spinner"

	"github.com/agbru/fibcalc/internal/format"
)

// FormatExecutionDuration formats a time.Duration for display.
// It delegates to the format package for the shared implementation.
func FormatExecutionDuration(d time.Duration) string {
	return format.FormatExecutionDuration(d)
}

const (
	// TruncationLimit is the digit threshold from which a result is truncated
	// in standard output to avoid cluttering the terminal.
	TruncationLimit = 100
	// DisplayEdges specifies the number of digits to display at the beginning
	// and end of a truncated number.
	DisplayEdges = 25
	// HexDisplayEdges specifies the number of hex characters to display at the
	// beginning and end of a truncated hexadecimal number.
	HexDisplayEdges = 40
	// ProgressRefreshRate defines the refresh frequency of the progress bar.
	// Optimized to 200ms to reduce updates and improve performance.
	ProgressRefreshRate = 200 * time.Millisecond
	// ProgressBarWidth defines the width in characters of the progress bar.
	ProgressBarWidth = 40
)

// Spinner is an interface that abstracts the behavior of a terminal spinner.
// This allows for the decoupling of the `DisplayProgress` function from a
// specific spinner implementation, facilitating easier testing and maintenance.
// It defines the essential controls for a spinner: starting, stopping, and
// updating its status message.
type Spinner interface {
	// Start begins the spinner animation.
	Start()
	// Stop halts the spinner animation.
	Stop()
	// UpdateSuffix sets the text that is displayed after the spinner.
	//
	// Parameters:
	//   - suffix: The text string to display.
	UpdateSuffix(suffix string)
}

// realSpinner is a wrapper for the `spinner.Spinner` that implements the
// `Spinner` interface. This adapter allows the `spinner` library to be used
// within the application's CLI framework.
type realSpinner struct {
	s *spinner.Spinner
}

// Start begins the spinner animation.
func (rs *realSpinner) Start() {
	rs.s.Start()
}

// Stop halts the spinner animation.
func (rs *realSpinner) Stop() {
	rs.s.Stop()
}

// UpdateSuffix sets the text that is displayed after the spinner.
//
// Parameters:
//   - suffix: The string to display.
func (rs *realSpinner) UpdateSuffix(suffix string) {
	rs.s.Suffix = suffix
}

var newSpinner = func(options ...spinner.Option) Spinner {
	// Using the same interval as ProgressRefreshRate to synchronize
	s := spinner.New(spinner.CharSets[11], ProgressRefreshRate, options...)
	return &realSpinner{s}
}

// ProgressState is a type alias for format.ProgressState.
// It is kept here for backward compatibility within the CLI package.
type ProgressState = format.ProgressState

// NewProgressState delegates to format.NewProgressState.
func NewProgressState(numCalculators int) *ProgressState {
	return format.NewProgressState(numCalculators)
}

// progressBar delegates to format.ProgressBar.
func progressBar(progress float64, length int) string {
	return format.ProgressBar(progress, length)
}
