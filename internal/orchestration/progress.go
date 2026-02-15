package orchestration

import (
	"time"

	"github.com/agbru/fibcalc/internal/format"
	"github.com/agbru/fibcalc/internal/progress"
)

// ProgressAggregator manages multi-calculator progress aggregation.
// It wraps format.ProgressWithETA and provides a higher-level API
// for consuming progress updates from a channel. Both CLI and TUI
// use this to avoid duplicating the aggregation setup and update logic.
type ProgressAggregator struct {
	state          *format.ProgressWithETA
	numCalculators int
}

// NewProgressAggregator creates a new aggregator for the given number
// of calculators. Returns nil if numCalculators <= 0.
func NewProgressAggregator(numCalculators int) *ProgressAggregator {
	if numCalculators <= 0 {
		return nil
	}
	return &ProgressAggregator{
		state:          format.NewProgressWithETA(numCalculators),
		numCalculators: numCalculators,
	}
}

// AggregatedProgress holds the result of processing a single progress update.
type AggregatedProgress struct {
	// CalculatorIndex is the index of the calculator that sent the update.
	CalculatorIndex int
	// Value is the raw progress value from the update (0.0 to 1.0).
	Value float64
	// AverageProgress is the aggregated average across all calculators.
	AverageProgress float64
	// ETA is the estimated time remaining based on smoothed progress rate.
	ETA time.Duration
}

// Update processes a single progress update and returns the aggregated result.
func (a *ProgressAggregator) Update(update progress.ProgressUpdate) AggregatedProgress {
	avgProgress, eta := a.state.UpdateWithETA(update.CalculatorIndex, update.Value)
	return AggregatedProgress{
		CalculatorIndex: update.CalculatorIndex,
		Value:           update.Value,
		AverageProgress: avgProgress,
		ETA:             eta,
	}
}

// CalculateAverage returns the current average progress without updating.
// Useful for periodic refresh between updates (e.g., CLI ticker).
func (a *ProgressAggregator) CalculateAverage() float64 {
	return a.state.CalculateAverage()
}

// GetETA returns the current ETA estimate without updating.
// Useful for periodic refresh between updates (e.g., CLI ticker).
func (a *ProgressAggregator) GetETA() time.Duration {
	return a.state.GetETA()
}

// NumCalculators returns the number of calculators being tracked.
func (a *ProgressAggregator) NumCalculators() int {
	return a.numCalculators
}

// IsMultiCalculator returns true if tracking more than one calculator.
func (a *ProgressAggregator) IsMultiCalculator() bool {
	return a.numCalculators > 1
}

// DrainChannel reads all updates from the channel without processing.
// Use this when numCalculators <= 0 and updates should be discarded.
func DrainChannel(progressChan <-chan progress.ProgressUpdate) {
	for range progressChan {
	}
}
