// Pool pre-warming for adaptive buffer pre-allocation based on calculation size.

package bigfft

import (
	"math/big"
	"sync/atomic"
)

// ─────────────────────────────────────────────────────────────────────────────
// Pool Pre-warming
// ─────────────────────────────────────────────────────────────────────────────

// PreWarmPools pre-allocates buffers in the pools based on estimated memory
// needs for calculating F(n). This reduces allocation overhead during the
// calculation by ensuring pools have ready-to-use buffers.
//
// The function estimates the required buffer sizes and pre-allocates an
// adaptive number of buffers in each relevant pool size class based on n:
//   - N < 100,000: 2 buffers (minimal overhead)
//   - 100,000 ≤ N < 1,000,000: 4 buffers
//   - 1,000,000 ≤ N < 10,000,000: 5 buffers
//   - N ≥ 10,000,000: 6 buffers (maximum for large calculations)
//
// This adaptive approach provides better performance for large calculations
// by reducing allocations during the computation.
//
// Parameters:
//   - n: The Fibonacci index to calculate (used for estimation).
func PreWarmPools(n uint64) {
	est := EstimateMemoryNeeds(n)

	// Determine the number of buffers based on calculation size
	numBuffers := 2 // Default for small calculations
	if n >= 10_000_000 {
		numBuffers = 6
	} else if n >= 1_000_000 {
		numBuffers = 5
	} else if n >= 100_000 {
		numBuffers = 4
	}

	// Pre-warm word slice pools
	wordIdx := getWordSlicePoolIndex(est.MaxWordSliceSize)
	if wordIdx >= 0 {
		for i := 0; i < numBuffers; i++ {
			buf := make([]big.Word, wordSliceSizes[wordIdx])
			wordSlicePools[wordIdx].Put(buf)
		}
	}

	// Pre-warm fermat pools
	fermatIdx := getFermatPoolIndex(est.MaxFermatSize)
	if fermatIdx >= 0 {
		for i := 0; i < numBuffers; i++ {
			buf := make(fermat, fermatSizes[fermatIdx])
			fermatPools[fermatIdx].Put(buf)
		}
	}

	// Pre-warm nat slice pools
	natIdx := getNatSlicePoolIndex(est.MaxNatSliceSize)
	if natIdx >= 0 {
		for i := 0; i < numBuffers; i++ {
			buf := make([]nat, natSliceSizes[natIdx])
			natSlicePools[natIdx].Put(buf)
		}
	}

	// Pre-warm fermat slice pools
	fermatSliceIdx := getFermatSlicePoolIndex(est.MaxFermatSliceSize)
	if fermatSliceIdx >= 0 {
		for i := 0; i < numBuffers; i++ {
			buf := make([]fermat, fermatSliceSizes[fermatSliceIdx])
			fermatSlicePools[fermatSliceIdx].Put(buf)
		}
	}
}

// poolsWarmed tracks whether pools have been pre-warmed.
// Using sync/atomic for lock-free, thread-safe initialization.
var poolsWarmed atomic.Bool

// EnsurePoolsWarmed ensures that pools are pre-warmed exactly once.
// This is more efficient than calling PreWarmPools on every calculation,
// as it uses atomic compare-and-swap to guarantee single initialization.
//
// The function is safe to call concurrently from multiple goroutines.
// Only the first call will actually pre-warm the pools; subsequent calls
// return immediately.
//
// Parameters:
//   - maxN: The maximum Fibonacci index expected (used for estimation).
func EnsurePoolsWarmed(maxN uint64) {
	if poolsWarmed.CompareAndSwap(false, true) {
		PreWarmPools(maxN)
	}
}
