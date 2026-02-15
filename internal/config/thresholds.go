package config

import "runtime"

// Threshold resolution chain (highest priority first):
//   1. CLI flags (--threshold, --fft-threshold, --strassen-threshold)
//   2. Environment variables (FIBCALC_THRESHOLD, etc.)
//   3. Cached calibration profile (~/.fibcalc_calibration.json)
//   4. Adaptive hardware estimation (this file)
//   5. Static defaults in fibonacci/constants.go

// ApplyAdaptiveThresholds adjusts the configuration thresholds based on
// hardware characteristics (CPU cores, architecture) when default values
// are detected. This provides automatic performance optimization without
// requiring explicit calibration.
//
// The function only modifies thresholds that are set to their zero default,
// preserving any user-specified overrides via command-line flags.
func ApplyAdaptiveThresholds(cfg AppConfig) AppConfig {
	if cfg.Threshold == 0 {
		cfg.Threshold = EstimateOptimalParallelThreshold()
	}
	if cfg.FFTThreshold == 0 {
		cfg.FFTThreshold = EstimateOptimalFFTThreshold()
	}
	if cfg.StrassenThreshold == 0 {
		cfg.StrassenThreshold = EstimateOptimalStrassenThreshold()
	}
	return cfg
}

// EstimateOptimalParallelThreshold provides a heuristic estimate of the optimal
// parallel threshold without running benchmarks.
// This can be used as a fallback or starting point.
func EstimateOptimalParallelThreshold() int {
	numCPU := runtime.NumCPU()

	switch {
	case numCPU == 1:
		return 0 // No parallelism
	case numCPU <= 2:
		return 8192 // High threshold - parallelism overhead is significant
	case numCPU <= 4:
		return 4096 // Default
	case numCPU <= 8:
		return 2048 // Can use more parallelism
	case numCPU <= 16:
		return 1024 // Many cores available
	default:
		return 512 // High core count - aggressive parallelism
	}
}

// EstimateOptimalFFTThreshold provides a heuristic estimate of the optimal
// FFT threshold without running benchmarks.
func EstimateOptimalFFTThreshold() int {
	wordSize := 32 << (^uint(0) >> 63)

	if wordSize == 64 {
		return 500000 // 500K bits on 64-bit (optimal for modern CPUs with large L3 caches)
	}
	return 250000 // 250K bits on 32-bit (lower due to smaller word size)
}

// EstimateOptimalStrassenThreshold provides a heuristic estimate of the optimal
// Strassen threshold without running benchmarks.
func EstimateOptimalStrassenThreshold() int {
	numCPU := runtime.NumCPU()

	if numCPU >= 4 {
		return 256 // With parallelism, lower threshold
	}
	return 3072 // Default from constants
}
