// This file provides type aliases for progress types that were moved to
// the internal/progress package. These aliases maintain backward compatibility
// so that consumers of the fibonacci package can continue to reference these
// types without changing their imports.

package fibonacci

import "github.com/agbru/fibcalc/internal/progress"

// Type aliases for types moved to internal/progress.
type (
	// ProgressUpdate is a type alias for progress.ProgressUpdate.
	ProgressUpdate = progress.ProgressUpdate

	// ProgressCallback is a type alias for progress.ProgressCallback.
	ProgressCallback = progress.ProgressCallback

	// ProgressObserver is a type alias for progress.ProgressObserver.
	ProgressObserver = progress.ProgressObserver

	// ProgressSubject is a type alias for progress.ProgressSubject.
	ProgressSubject = progress.ProgressSubject

	// ChannelObserver is a type alias for progress.ChannelObserver.
	ChannelObserver = progress.ChannelObserver

	// LoggingObserver is a type alias for progress.LoggingObserver.
	LoggingObserver = progress.LoggingObserver

	// NoOpObserver is a type alias for progress.NoOpObserver.
	NoOpObserver = progress.NoOpObserver
)

// Re-exported constructors and functions from internal/progress.
var (
	// NewProgressSubject creates a new progress subject.
	NewProgressSubject = progress.NewProgressSubject

	// NewChannelObserver creates a new channel observer.
	NewChannelObserver = progress.NewChannelObserver

	// NewLoggingObserver creates a new logging observer.
	NewLoggingObserver = progress.NewLoggingObserver

	// NewNoOpObserver creates a new no-op observer.
	NewNoOpObserver = progress.NewNoOpObserver

	// CalcTotalWork calculates the total work for O(log n) algorithms.
	CalcTotalWork = progress.CalcTotalWork

	// PrecomputePowers4 pre-calculates powers of 4.
	PrecomputePowers4 = progress.PrecomputePowers4

	// ReportStepProgress handles harmonized progress reporting.
	ReportStepProgress = progress.ReportStepProgress
)
