// Package calibration provides performance calibration for the Fibonacci calculator.
// This file implements adaptive threshold generation based on hardware characteristics.
package calibration

import (
	"runtime"
)

// ─────────────────────────────────────────────────────────────────────────────
// Adaptive Parallel Threshold Generation
// ─────────────────────────────────────────────────────────────────────────────

// GenerateParallelThresholds generates a list of parallel thresholds to test
// based on the number of available CPU cores.
//
// The rationale:
// - Single-core: Only test sequential (0) as parallelism has no benefit
// - 2-4 cores: Test lower thresholds as parallelism overhead is relatively high
// - 8+ cores: Include higher thresholds as more parallelism can be beneficial
// - 16+ cores: Add even higher thresholds for very fine-grained parallelism
func GenerateParallelThresholds() []int {
	numCPU := runtime.NumCPU()

	// Base thresholds always tested
	thresholds := []int{0} // Sequential (no parallelism)

	switch {
	case numCPU == 1:
		// Single core: only sequential makes sense
		return thresholds

	case numCPU <= 4:
		// Few cores: test moderate thresholds
		thresholds = append(thresholds, 512, 1024, 2048, 4096)

	case numCPU <= 8:
		// Medium core count: broader range
		thresholds = append(thresholds, 256, 512, 1024, 2048, 4096, 8192)

	case numCPU <= 16:
		// Many cores: include higher thresholds
		thresholds = append(thresholds, 256, 512, 1024, 2048, 4096, 8192, 16384)

	default:
		// High core count (16+): full range including very high thresholds
		thresholds = append(thresholds, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768)
	}

	return thresholds
}

// GenerateQuickParallelThresholds generates a smaller set of thresholds for
// quick auto-calibration at startup.
func GenerateQuickParallelThresholds() []int {
	numCPU := runtime.NumCPU()

	if numCPU == 1 {
		return []int{0}
	}

	// Reduced set for quick calibration
	switch {
	case numCPU <= 4:
		return []int{0, 2048, 4096}
	case numCPU <= 8:
		return []int{0, 2048, 4096, 8192}
	default:
		return []int{0, 2048, 4096, 8192, 16384}
	}
}

// GenerateQuickFFTThresholds generates a smaller set for quick calibration.
func GenerateQuickFFTThresholds() []int {
	return []int{0, 750000, 1000000, 1500000}
}

// ─────────────────────────────────────────────────────────────────────────────
// Adaptive Strassen Threshold Generation
// ─────────────────────────────────────────────────────────────────────────────

// GenerateQuickStrassenThresholds generates a smaller set for quick calibration.
func GenerateQuickStrassenThresholds() []int {
	return []int{192, 256, 384, 512}
}

// ─────────────────────────────────────────────────────────────────────────────
// Threshold Estimation (without benchmarking)
// ─────────────────────────────────────────────────────────────────────────────

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

