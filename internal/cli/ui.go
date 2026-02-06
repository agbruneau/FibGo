//go:generate mockgen -source=ui.go -destination=mocks/mock_ui.go -package=mocks

package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

// FormatExecutionDuration formats a time.Duration for display.
// It shows microseconds for durations less than a millisecond, milliseconds for
// durations less than a second, and the default string representation otherwise.
// This approach provides a more human-readable output for short durations.
//
// Parameters:
//   - d: The duration to format.
//
// Returns:
//   - string: A formatted string representing the duration.
func FormatExecutionDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	} else if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return d.String()
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

// ProgressState encapsulates the aggregated progress of concurrent calculations.
// It maintains the individual progress of each calculator and computes the
// average, which is essential for providing a consolidated progress view when
// multiple algorithms are running in parallel.
type ProgressState struct {
	progresses     []float64
	numCalculators int
}

// NewProgressState creates and initializes a new ProgressState.
// It sets up the internal storage for tracking the progress of a specified
// number of calculators.
//
// Parameters:
//   - numCalculators: The number of calculators to track.
//
// Returns:
//   - *ProgressState: A pointer to the new progress state object.
func NewProgressState(numCalculators int) *ProgressState {
	return &ProgressState{
		progresses:     make([]float64, numCalculators),
		numCalculators: numCalculators,
	}
}

// Update records a new progress value for a specific calculator.
// It is designed to be safe for concurrent use, although in the current
// implementation it is called sequentially. The method ensures that updates are
// only applied for valid calculator indices.
//
// Parameters:
//   - index: The index of the calculator (0 to numCalculators-1).
//   - value: The progress value (0.0 to 1.0).
func (ps *ProgressState) Update(index int, value float64) {
	if index >= 0 && index < len(ps.progresses) {
		ps.progresses[index] = value
	}
}

// CalculateAverage computes the average progress across all tracked calculators.
// This is used to display a single, consolidated progress bar to the user,
// representing the overall progress of the application.
//
// Returns:
//   - float64: The average progress (0.0 to 1.0).
func (ps *ProgressState) CalculateAverage() float64 {
	var totalProgress float64
	for _, p := range ps.progresses {
		totalProgress += p
	}
	if ps.numCalculators == 0 {
		return 0.0
	}
	return totalProgress / float64(ps.numCalculators)
}

// progressBar generates a string representing a textual progress bar.
//
// Parameters:
//   - progress: The normalized progress value (0.0 to 1.0).
//   - length: The total character width of the progress bar.
//
// Returns:
//   - string: A string representation of the progress bar.
func progressBar(progress float64, length int) string {
	if progress > 1.0 {
		progress = 1.0
	}
	if progress < 0.0 {
		progress = 0.0
	}
	count := int(progress * float64(length))
	var builder strings.Builder
	builder.Grow(length)
	for i := 0; i < length; i++ {
		if i < count {
			builder.WriteRune('█')
		} else {
			builder.WriteRune('░')
		}
	}
	return builder.String()
}
