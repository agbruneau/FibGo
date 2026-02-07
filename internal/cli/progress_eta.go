package cli

import (
	"time"

	"github.com/agbru/fibcalc/internal/format"
)

// ProgressWithETA is a type alias for format.ProgressWithETA.
// It is kept here for backward compatibility within the CLI package.
type ProgressWithETA = format.ProgressWithETA

// NewProgressWithETA delegates to format.NewProgressWithETA.
func NewProgressWithETA(numCalculators int) *ProgressWithETA {
	return format.NewProgressWithETA(numCalculators)
}

// FormatETA delegates to format.FormatETA.
func FormatETA(eta time.Duration) string {
	return format.FormatETA(eta)
}

// FormatProgressBarWithETA delegates to format.FormatProgressBarWithETA.
func FormatProgressBarWithETA(progress float64, eta time.Duration, width int) string {
	return format.FormatProgressBarWithETA(progress, eta, width)
}
