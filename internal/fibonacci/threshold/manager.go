// This file implements dynamic threshold adjustment during calculation.

package threshold

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// ─────────────────────────────────────────────────────────────────────────────
// Dynamic Threshold Configuration
// ─────────────────────────────────────────────────────────────────────────────

const (
	// DynamicAdjustmentInterval is the number of iterations between threshold checks.
	DynamicAdjustmentInterval = 5

	// MinMetricsForAdjustment is the minimum number of metrics needed before adjusting.
	MinMetricsForAdjustment = 3

	// MaxMetricsHistory is the maximum number of metrics to keep for analysis.
	MaxMetricsHistory = 20

	// FFTSpeedupThreshold is the minimum speedup ratio to switch to FFT.
	// If FFT is expected to be at least this much faster, switch to it.
	FFTSpeedupThreshold = 1.2

	// ParallelSpeedupThreshold is the minimum speedup to enable parallelism.
	ParallelSpeedupThreshold = 1.1

	// HysteresisMargin prevents oscillating between modes.
	// Threshold must change by at least this factor to trigger adjustment.
	HysteresisMargin = 0.15
)

// DynamicThresholdManager adjusts FFT and parallel thresholds during calculation
// based on observed performance metrics.
type DynamicThresholdManager struct {
	mu     sync.RWMutex
	logger zerolog.Logger

	// Current thresholds (can be adjusted during calculation)
	currentFFTThreshold      int
	currentParallelThreshold int

	// Original thresholds (for comparison and bounds)
	originalFFTThreshold      int
	originalParallelThreshold int

	// Collected metrics - implemented as a Ring Buffer for O(1) ops
	metrics      [MaxMetricsHistory]IterationMetric
	metricsCount int // Total metrics collected (ever)
	metricsHead  int // Index of the next slot to write to

	// Running sums for fast average calculation (O(1))
	// We track separate sums for different categories to avoid iterating.
	// Note: These sums are approximate as they cover the window, but simplifying updates is key.
	// Actually, accurate running sums require removing the overwritten value.
	// Given MaxMetricsHistory is small (20), iterating is fast enough, but let's optimize slightly.
	// We will stick to iterating the small ring buffer for simplicity in categorization (FFT vs Non-FFT),
	// but the structure itself is now a fixed array to avoid allocation.

	// Adjustment state
	iterationCount     int
	adjustmentInterval int
	lastAdjustment     time.Time
}

// ─────────────────────────────────────────────────────────────────────────────
// Constructor and Configuration
// ─────────────────────────────────────────────────────────────────────────────

// NewDynamicThresholdManager creates a new manager with the given initial thresholds.
func NewDynamicThresholdManager(fftThreshold, parallelThreshold int) *DynamicThresholdManager {
	return &DynamicThresholdManager{
		logger:                    zerolog.Nop(),
		currentFFTThreshold:       fftThreshold,
		currentParallelThreshold:  parallelThreshold,
		originalFFTThreshold:      fftThreshold,
		originalParallelThreshold: parallelThreshold,
		adjustmentInterval:        DynamicAdjustmentInterval,
	}
}

// NewDynamicThresholdManagerFromConfig creates a manager from configuration.
func NewDynamicThresholdManagerFromConfig(cfg DynamicThresholdConfig) *DynamicThresholdManager {
	if !cfg.Enabled {
		return nil
	}

	interval := cfg.AdjustmentInterval
	if interval <= 0 {
		interval = DynamicAdjustmentInterval
	}

	return &DynamicThresholdManager{
		logger:                    zerolog.Nop(),
		currentFFTThreshold:       cfg.InitialFFTThreshold,
		currentParallelThreshold:  cfg.InitialParallelThreshold,
		originalFFTThreshold:      cfg.InitialFFTThreshold,
		originalParallelThreshold: cfg.InitialParallelThreshold,
		adjustmentInterval:        interval,
	}
}

// SetLogger configures the logger for threshold adjustment events.
func (m *DynamicThresholdManager) SetLogger(l zerolog.Logger) {
	m.logger = l
}

// ─────────────────────────────────────────────────────────────────────────────
// Metric Recording
// ─────────────────────────────────────────────────────────────────────────────

// RecordIteration records timing data for a completed iteration.
// This should be called after each doubling step in the algorithm.
func (m *DynamicThresholdManager) RecordIteration(bitLen int, duration time.Duration, usedFFT, usedParallel bool) {
	metric := IterationMetric{
		BitLen:       bitLen,
		Duration:     duration,
		UsedFFT:      usedFFT,
		UsedParallel: usedParallel,
	}

	// Write to ring buffer (no mutex needed: called from single goroutine in the doubling loop)
	m.metrics[m.metricsHead] = metric
	m.metricsHead = (m.metricsHead + 1) % MaxMetricsHistory
	m.metricsCount++
	m.iterationCount++
}

// ─────────────────────────────────────────────────────────────────────────────
// Threshold Access
// ─────────────────────────────────────────────────────────────────────────────

// GetThresholds returns the current FFT and parallel thresholds.
func (m *DynamicThresholdManager) GetThresholds() (fft, parallel int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentFFTThreshold, m.currentParallelThreshold
}

// GetFFTThreshold returns the current FFT threshold.
func (m *DynamicThresholdManager) GetFFTThreshold() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentFFTThreshold
}

// GetParallelThreshold returns the current parallel threshold.
func (m *DynamicThresholdManager) GetParallelThreshold() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentParallelThreshold
}

// ─────────────────────────────────────────────────────────────────────────────
// Adjustment Logic
// ─────────────────────────────────────────────────────────────────────────────

// ShouldAdjust checks if thresholds should be adjusted based on collected metrics.
// Returns the new thresholds and whether an adjustment was made.
// No mutex needed: called from single goroutine in the doubling loop.
func (m *DynamicThresholdManager) ShouldAdjust() (newFFT, newParallel int, adjusted bool) {
	// Check if we should evaluate adjustments
	if m.iterationCount%m.adjustmentInterval != 0 {
		return m.currentFFTThreshold, m.currentParallelThreshold, false
	}

	if m.metricsCount < MinMetricsForAdjustment {
		return m.currentFFTThreshold, m.currentParallelThreshold, false
	}

	// Analyze recent metrics to determine if adjustments are beneficial
	newFFT = m.analyzeFFTThreshold()
	newParallel = m.analyzeParallelThreshold()

	// Check if changes are significant enough (hysteresis)
	fftChanged := m.significantChange(m.currentFFTThreshold, newFFT)
	parallelChanged := m.significantChange(m.currentParallelThreshold, newParallel)

	if fftChanged || parallelChanged {
		oldFFT := m.currentFFTThreshold
		oldParallel := m.currentParallelThreshold
		if fftChanged {
			m.currentFFTThreshold = newFFT
		}
		if parallelChanged {
			m.currentParallelThreshold = newParallel
		}
		m.lastAdjustment = time.Now()
		m.logger.Debug().
			Int("iteration", m.iterationCount).
			Bool("fft_changed", fftChanged).
			Int("fft_old", oldFFT).
			Int("fft_new", m.currentFFTThreshold).
			Bool("parallel_changed", parallelChanged).
			Int("parallel_old", oldParallel).
			Int("parallel_new", m.currentParallelThreshold).
			Msg("thresholds adjusted")
		return m.currentFFTThreshold, m.currentParallelThreshold, true
	}

	return m.currentFFTThreshold, m.currentParallelThreshold, false
}

// getActiveMetrics returns a slice of valid metrics from the ring buffer.
func (m *DynamicThresholdManager) getActiveMetrics() []IterationMetric {
	count := m.metricsCount
	if count > MaxMetricsHistory {
		count = MaxMetricsHistory
	}

	// Create a temporary slice to make analysis easier without complex ring buffer arithmetic
	// Since MaxMetricsHistory is small (20), this copy is cheap and simplifies logic.
	result := make([]IterationMetric, count)

	if m.metricsCount <= MaxMetricsHistory {
		copy(result, m.metrics[:count])
	} else {
		// Buffer wrapped around. Order: [Head...End] + [0...Head-1]
		// Actually, since we just need the values and not strict temporal order for averages,
		// we can just copy the whole array.
		// However, to be strictly correct if we needed order:
		// tailLen := MaxMetricsHistory - m.metricsHead
		// copy(result, m.metrics[m.metricsHead:])
		// copy(result[tailLen:], m.metrics[:m.metricsHead])
		// For averages, order doesn't matter.
		copy(result, m.metrics[:])
	}
	return result
}

// thresholdAnalysisParams captures the per-threshold configuration differences
// used by analyzeThreshold to avoid duplicated analysis logic.
type thresholdAnalysisParams struct {
	// predicate selects which metrics belong to the "optimized" mode (FFT or parallel).
	predicate func(IterationMetric) bool
	// speedupThreshold is the minimum ratio to consider the optimized mode faster.
	speedupThreshold float64
	// lowerNumerator/raiseNumerator over 10: adjustment factors when lowering/raising.
	lowerNumerator int
	raiseNumerator int
	// minThreshold is the floor for the threshold value.
	minThreshold int
	// maxCapMultiplier is the multiplier against the original threshold for the ceiling.
	maxCapMultiplier int
	// currentThreshold and originalThreshold are the values being analyzed.
	currentThreshold  int
	originalThreshold int
}

// filterMetricsByMode partitions metrics into two groups based on the predicate.
// Returns (matching, non-matching) slices.
func filterMetricsByMode(metrics []IterationMetric, predicate func(IterationMetric) bool) (matching, nonMatching []IterationMetric) {
	matching = make([]IterationMetric, 0, len(metrics))
	nonMatching = make([]IterationMetric, 0, len(metrics))
	for _, metric := range metrics {
		if predicate(metric) {
			matching = append(matching, metric)
		} else {
			nonMatching = append(nonMatching, metric)
		}
	}
	return matching, nonMatching
}

// calculateSpeedupRatio returns the speedup ratio of baseline over optimized.
// Returns 0 if either average is non-positive.
func calculateSpeedupRatio(avgOptimized, avgBaseline float64) float64 {
	if avgOptimized <= 0 || avgBaseline <= 0 {
		return 0
	}
	return avgBaseline / avgOptimized
}

// analyzeThreshold is the common analysis logic for both FFT and parallel thresholds.
// It partitions metrics, computes a speedup ratio, and returns an adjusted threshold.
func (m *DynamicThresholdManager) analyzeThreshold(params thresholdAnalysisParams) int {
	metrics := m.getActiveMetrics()
	if len(metrics) == 0 {
		return params.currentThreshold
	}

	optimized, baseline := filterMetricsByMode(metrics, params.predicate)
	if len(optimized) == 0 || len(baseline) == 0 {
		return params.currentThreshold
	}

	ratio := calculateSpeedupRatio(m.avgTimePerBit(optimized), m.avgTimePerBit(baseline))
	if ratio == 0 {
		return params.currentThreshold
	}

	return m.applyThresholdAdjustment(ratio, params)
}

// applyThresholdAdjustment applies the lower/raise logic based on the speedup ratio.
func (m *DynamicThresholdManager) applyThresholdAdjustment(ratio float64, params thresholdAnalysisParams) int {
	if ratio > params.speedupThreshold {
		// Optimized mode is faster, lower threshold
		newThreshold := params.currentThreshold * params.lowerNumerator / 10
		if newThreshold < params.minThreshold {
			newThreshold = params.minThreshold
		}
		return newThreshold
	}
	if ratio < 1.0/params.speedupThreshold {
		// Optimized mode is slower, raise threshold
		newThreshold := params.currentThreshold * params.raiseNumerator / 10
		maxCap := params.originalThreshold * params.maxCapMultiplier
		if newThreshold > maxCap {
			newThreshold = maxCap
		}
		return newThreshold
	}
	return params.currentThreshold
}

// analyzeFFTThreshold analyzes metrics to determine optimal FFT threshold.
func (m *DynamicThresholdManager) analyzeFFTThreshold() int {
	return m.analyzeThreshold(thresholdAnalysisParams{
		predicate:         func(metric IterationMetric) bool { return metric.UsedFFT },
		speedupThreshold:  FFTSpeedupThreshold,
		lowerNumerator:    9,
		raiseNumerator:    11,
		minThreshold:      100000,
		maxCapMultiplier:  2,
		currentThreshold:  m.currentFFTThreshold,
		originalThreshold: m.originalFFTThreshold,
	})
}

// analyzeParallelThreshold analyzes metrics to determine optimal parallel threshold.
func (m *DynamicThresholdManager) analyzeParallelThreshold() int {
	return m.analyzeThreshold(thresholdAnalysisParams{
		predicate:         func(metric IterationMetric) bool { return metric.UsedParallel },
		speedupThreshold:  ParallelSpeedupThreshold,
		lowerNumerator:    8,
		raiseNumerator:    12,
		minThreshold:      1024,
		maxCapMultiplier:  4,
		currentThreshold:  m.currentParallelThreshold,
		originalThreshold: m.originalParallelThreshold,
	})
}

// avgTimePerBit calculates average time per bit across metrics.
func (m *DynamicThresholdManager) avgTimePerBit(metrics []IterationMetric) float64 {
	if len(metrics) == 0 {
		return 0
	}

	var totalTime time.Duration
	var totalBits int64
	for _, metric := range metrics {
		totalTime += metric.Duration
		totalBits += int64(metric.BitLen)
	}

	if totalBits == 0 {
		return 0
	}

	return float64(totalTime.Nanoseconds()) / float64(totalBits)
}

// significantChange checks if a threshold change is significant enough to apply.
func (m *DynamicThresholdManager) significantChange(oldVal, newVal int) bool {
	if oldVal == 0 {
		return newVal != 0
	}
	change := float64(newVal-oldVal) / float64(oldVal)
	if change < 0 {
		change = -change
	}
	return change > HysteresisMargin
}

// ─────────────────────────────────────────────────────────────────────────────
// Statistics and Reporting
// ─────────────────────────────────────────────────────────────────────────────

// GetStats returns current statistics about the manager.
func (m *DynamicThresholdManager) GetStats() ThresholdStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := m.metricsCount
	if count > MaxMetricsHistory {
		count = MaxMetricsHistory
	}

	return ThresholdStats{
		CurrentFFT:          m.currentFFTThreshold,
		CurrentParallel:     m.currentParallelThreshold,
		OriginalFFT:         m.originalFFTThreshold,
		OriginalParallel:    m.originalParallelThreshold,
		MetricsCollected:    count,
		IterationsProcessed: m.iterationCount,
	}
}

// Reset clears all collected metrics and restores original thresholds.
func (m *DynamicThresholdManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentFFTThreshold = m.originalFFTThreshold
	m.currentParallelThreshold = m.originalParallelThreshold
	// Ring buffer reset is simple
	m.metricsCount = 0
	m.metricsHead = 0
	m.iterationCount = 0
}
