package fibonacci

import (
	"context"
	"fmt"
	"math/big"
	"math/bits"
	"reflect"
	"testing"
)

// allDistinct checks that all provided *big.Int pointers point to distinct
// memory locations. This is critical for the Fast Doubling algorithm's
// correctness: if any two pointers alias the same object, in-place
// arithmetic operations would corrupt intermediate results.
func allDistinct(ptrs ...*big.Int) bool {
	seen := make(map[uintptr]bool, len(ptrs))
	for _, p := range ptrs {
		addr := reflect.ValueOf(p).Pointer()
		if seen[addr] {
			return false
		}
		seen[addr] = true
	}
	return true
}

// fibIterative computes the n-th Fibonacci number using simple iterative
// addition. This serves as the ground-truth oracle for verifying that
// the Fast Doubling algorithm produces correct results after all pointer
// swaps and rotations.
func fibIterative(n uint64) *big.Int {
	if n == 0 {
		return big.NewInt(0)
	}
	if n == 1 {
		return big.NewInt(1)
	}
	a, b := big.NewInt(0), big.NewInt(1)
	for i := uint64(2); i <= n; i++ {
		a.Add(a, b)
		a, b = b, a
	}
	return b
}

// TestStatePointerSwapsExhaustive exhaustively verifies that after every
// swap operation in the Fast Doubling loop, all 6 pointers in the
// CalculationState remain distinct (no aliasing).
//
// The algorithm from doubling_framework.go performs these swaps per iteration:
//  1. Capacity swap (conditional): s.T1, s.T4 = s.T4, s.T1
//  2. 5-way rotation: s.FK, s.FK1, s.T2, s.T3, s.T1 = s.T3, s.T1, s.FK, s.FK1, s.T2
//  3. 3-way rotation (if bit is 1): s.FK, s.FK1, s.T4 = s.FK1, s.T4, s.FK
//
// This test simulates just the pointer swap logic (without actual
// multiplication) for all 16-bit numbers (0 through 65535) to cover
// every possible bit pattern. At each iteration, it checks that all 6
// pointers remain distinct.
func TestStatePointerSwapsExhaustive(t *testing.T) {
	t.Parallel()

	for n := uint64(0); n <= 65535; n++ {
		// Create 6 distinct big.Int values
		vals := [6]*big.Int{
			big.NewInt(0), big.NewInt(1), big.NewInt(2),
			big.NewInt(3), big.NewInt(4), big.NewInt(5),
		}
		fk, fk1, t1, t2, t3, t4 := vals[0], vals[1], vals[2], vals[3], vals[4], vals[5]

		numBits := bits.Len64(n)
		for i := numBits - 1; i >= 0; i-- {
			// Capacity swap (simulate: alternate the condition)
			if i%2 == 0 {
				t1, t4 = t4, t1
			}

			// 5-way rotation (always happens after multiplications)
			fk, fk1, t2, t3, t1 = t3, t1, fk, fk1, t2

			// 3-way rotation (if bit is 1)
			if (n>>uint(i))&1 == 1 {
				fk, fk1, t4 = fk1, t4, fk
			}

			// Check all 6 are distinct
			if !allDistinct(fk, fk1, t1, t2, t3, t4) {
				t.Fatalf("Aliasing detected at n=%d, bit=%d", n, i)
			}
		}
	}
}

// TestResultStealingIndependence verifies the zero-copy result stealing
// pattern used at the end of ExecuteDoublingLoop. After the loop, the
// algorithm "steals" FK from the state by saving a reference and replacing
// s.FK with a fresh big.Int. This test ensures:
//  1. The returned result matches the expected Fibonacci value.
//  2. s.FK is a fresh zero-valued big.Int after stealing.
//  3. The result and s.FK are different pointers.
//  4. Releasing the state back to the pool does not corrupt the result.
func TestResultStealingIndependence(t *testing.T) {
	t.Parallel()

	s := AcquireState()

	// Compute F(100) using the real doubling framework
	strategy := &KaratsubaStrategy{}
	framework := NewDoublingFramework(strategy)
	result, err := framework.ExecuteDoublingLoop(context.Background(), func(float64) {}, 100, Options{}, s, false)
	if err != nil {
		t.Fatalf("ExecuteDoublingLoop failed: %v", err)
	}

	// Result should match the iteratively computed F(100)
	expected := fibIterative(100)
	if result.Cmp(expected) != 0 {
		t.Errorf("Result mismatch: got %s, want %s", result.String(), expected.String())
	}

	// s.FK should now be a fresh big.Int (value 0) after the steal
	if s.FK.Sign() != 0 {
		t.Errorf("s.FK should be zero after stealing, got %s", s.FK.String())
	}

	// result and s.FK should be different pointers
	resultAddr := reflect.ValueOf(result).Pointer()
	fkAddr := reflect.ValueOf(s.FK).Pointer()
	if resultAddr == fkAddr {
		t.Error("result and s.FK should be different pointers after stealing")
	}

	// Release state and verify result is unaffected
	ReleaseState(s)
	if result.Cmp(expected) != 0 {
		t.Error("Result should be unaffected by ReleaseState")
	}
}

// TestAliasingWithRealComputation runs the real Fast Doubling algorithm
// for various n values and verifies correctness by comparing against
// the iterative oracle. This serves as an integration-level aliasing
// test: if pointer swaps were incorrect, the computed values would
// diverge from the expected Fibonacci numbers.
func TestAliasingWithRealComputation(t *testing.T) {
	t.Parallel()

	testCases := []uint64{0, 1, 2, 3, 5, 10, 50, 92, 93, 94, 100, 500, 1000, 5000, 10000}
	for _, n := range testCases {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			t.Parallel()

			s := AcquireState()
			defer ReleaseState(s)

			strategy := &KaratsubaStrategy{}
			framework := NewDoublingFramework(strategy)
			result, err := framework.ExecuteDoublingLoop(context.Background(), func(float64) {}, n, Options{}, s, false)

			if n == 0 {
				// For n=0, the loop does 0 iterations; result is FK which is 0
				if err != nil {
					t.Fatalf("Unexpected error for n=0: %v", err)
				}
				if result.Sign() != 0 {
					t.Errorf("F(0) should be 0, got %s", result.String())
				}
				return
			}

			if err != nil {
				t.Fatalf("ExecuteDoublingLoop failed for n=%d: %v", n, err)
			}

			expected := fibIterative(n)
			if result.Cmp(expected) != 0 {
				t.Errorf("F(%d) mismatch: got %s, want %s", n, result.String(), expected.String())
			}
		})
	}
}
