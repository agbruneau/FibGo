package fibonacci

import (
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"testing"
)

// TestPoolRoundTripConcurrent verifies that the state pool correctly
// handles 1000 concurrent goroutines each acquiring, validating, and
// releasing a CalculationState. For each acquired state, it checks:
//  1. All 6 pointers (FK, FK1, T1, T2, T3, T4) are non-nil.
//  2. All 6 pointers are distinct (no aliasing between fields).
//  3. FK is initialized to 0 and FK1 to 1 (per Reset semantics).
func TestPoolRoundTripConcurrent(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	errors := make(chan string, 1000)

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := AcquireState()
			defer ReleaseState(s)

			// Verify all 6 pointers non-nil
			if s.FK == nil || s.FK1 == nil || s.T1 == nil || s.T2 == nil || s.T3 == nil || s.T4 == nil {
				errors <- "nil pointer in acquired state"
				return
			}

			// Verify all distinct
			ptrs := []*big.Int{s.FK, s.FK1, s.T1, s.T2, s.T3, s.T4}
			seen := make(map[uintptr]bool)
			for _, p := range ptrs {
				addr := reflect.ValueOf(p).Pointer()
				if seen[addr] {
					errors <- "duplicate pointer in acquired state"
					return
				}
				seen[addr] = true
			}

			// Verify reset values
			if s.FK.Sign() != 0 {
				errors <- fmt.Sprintf("FK not zero: %s", s.FK.String())
				return
			}
			if s.FK1.Cmp(big.NewInt(1)) != 0 {
				errors <- fmt.Sprintf("FK1 not one: %s", s.FK1.String())
				return
			}
		}()
	}

	wg.Wait()
	close(errors)
	for err := range errors {
		t.Fatal(err)
	}
}

// TestPoolOversizedRejection verifies that states containing big.Int
// values exceeding MaxPooledBitLen are not returned to the pool.
// When such a state is released, the pool should discard it so that
// the next AcquireState does not return that specific oversized state.
func TestPoolOversizedRejection(t *testing.T) {
	t.Parallel()

	s := AcquireState()
	// Set T1 to a value exceeding MaxPooledBitLen (100M bits)
	huge := new(big.Int).Lsh(big.NewInt(1), MaxPooledBitLen+1)
	s.T1.Set(huge)

	// Record the pointer of the oversized state before release
	oversizedAddr := reflect.ValueOf(s).Pointer()
	ReleaseState(s)

	// Acquire multiple new states - none should be the oversized one.
	// The pool discards oversized states, so they should not be returned.
	// We acquire several to increase confidence (the pool may have other
	// recycled states queued ahead).
	const attempts = 10
	for i := 0; i < attempts; i++ {
		s2 := AcquireState()
		s2Addr := reflect.ValueOf(s2).Pointer()
		if s2Addr == oversizedAddr {
			// This state was supposed to be discarded. Verify its T1 is not
			// the oversized value (the pool may reuse the struct address via
			// GC, but in that case it would be a fresh allocation).
			if s2.T1.BitLen() > MaxPooledBitLen {
				t.Errorf("Acquired state with oversized T1 (%d bits) that should have been discarded",
					s2.T1.BitLen())
			}
		}
		// Verify the acquired state has valid reset values
		if s2.FK == nil || s2.FK1 == nil {
			t.Fatal("Acquired state has nil FK or FK1")
		}
		if s2.FK.Sign() != 0 {
			t.Errorf("FK not zero in acquired state")
		}
		if s2.FK1.Cmp(big.NewInt(1)) != 0 {
			t.Errorf("FK1 not one in acquired state")
		}
		ReleaseState(s2)
	}
}

// TestPreSizeBigInt verifies the preSizeBigInt utility function which
// pre-allocates capacity for a big.Int's internal word slice without
// changing its value. This is used to avoid repeated reallocation
// during the doubling loop as values grow.
func TestPreSizeBigInt(t *testing.T) {
	t.Parallel()

	t.Run("basic pre-sizing", func(t *testing.T) {
		t.Parallel()
		z := new(big.Int)
		preSizeBigInt(z, 100)
		if cap(z.Bits()) < 100 {
			t.Errorf("cap after preSizeBigInt: got %d, want >= 100", cap(z.Bits()))
		}
		// Value should be 0
		if z.Sign() != 0 {
			t.Error("preSizeBigInt changed the value")
		}
	})

	t.Run("nil safety", func(t *testing.T) {
		t.Parallel()
		// Should not panic
		preSizeBigInt(nil, 100)
	})

	t.Run("zero words", func(t *testing.T) {
		t.Parallel()
		z := new(big.Int)
		preSizeBigInt(z, 0) // should be no-op
	})

	t.Run("negative words", func(t *testing.T) {
		t.Parallel()
		z := new(big.Int)
		preSizeBigInt(z, -1) // should be no-op
	})

	t.Run("already large enough", func(t *testing.T) {
		t.Parallel()
		z := new(big.Int).SetBits(make([]big.Word, 0, 200))
		preSizeBigInt(z, 100) // should be no-op since cap >= 100
		if cap(z.Bits()) < 100 {
			t.Errorf("cap should still be >= 100 after no-op preSizeBigInt, got %d", cap(z.Bits()))
		}
	})
}
