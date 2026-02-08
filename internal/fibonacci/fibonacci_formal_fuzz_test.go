package fibonacci

import (
	"context"
	"math/big"
	"testing"
)

// FuzzNonNegative verifies that F(n) >= 0 for all n.
// Fibonacci numbers are always non-negative, so any negative result
// indicates a correctness bug in the algorithm.
func FuzzNonNegative(f *testing.F) {
	seeds := []uint64{0, 1, 2, 3, 5, 10, 50, 92, 93, 94, 100, 500, 1000, 5000, 10000}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, n uint64) {
		if n > 50000 {
			n = n % 50000
		}

		factory := NewDefaultFactory()
		calc, err := factory.Get("fast")
		if err != nil {
			t.Fatalf("failed to get calculator: %v", err)
		}

		result, err := calc.Calculate(context.Background(), nil, 0, n, Options{})
		if err != nil {
			t.Skipf("calculation error for n=%d: %v", n, err)
		}
		if result.Sign() < 0 {
			t.Errorf("F(%d) is negative: %s", n, result.String())
		}
	})
}

// FuzzGeneralizedCassini verifies the generalized Cassini identity:
//
//	F(n+1)^2 - F(n)*F(n+2) = (-1)^n
//
// This identity holds for all n >= 1 and provides a strong algebraic
// consistency check on the computed Fibonacci values.
func FuzzGeneralizedCassini(f *testing.F) {
	seeds := []uint64{2, 5, 10, 50, 93, 100, 500, 1000, 5000}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, n uint64) {
		if n < 1 {
			n = 1
		}
		if n > 50000 {
			n = n%49999 + 1
		}

		factory := NewDefaultFactory()
		calc, err := factory.Get("fast")
		if err != nil {
			t.Fatalf("failed to get calculator: %v", err)
		}

		ctx := context.Background()

		fn, err := calc.Calculate(ctx, nil, 0, n, Options{})
		if err != nil {
			t.Skipf("calculation error for F(%d): %v", n, err)
		}
		fn1, err := calc.Calculate(ctx, nil, 0, n+1, Options{})
		if err != nil {
			t.Skipf("calculation error for F(%d): %v", n+1, err)
		}
		fn2, err := calc.Calculate(ctx, nil, 0, n+2, Options{})
		if err != nil {
			t.Skipf("calculation error for F(%d): %v", n+2, err)
		}

		// F(n+1)^2 - F(n)*F(n+2) = (-1)^n
		fn1sq := new(big.Int).Mul(fn1, fn1)
		fnfn2 := new(big.Int).Mul(fn, fn2)
		diff := new(big.Int).Sub(fn1sq, fnfn2)

		expected := big.NewInt(1)
		if n%2 == 1 {
			expected.SetInt64(-1)
		}

		if diff.Cmp(expected) != 0 {
			t.Errorf("Generalized Cassini failed for n=%d: got %s, want %s", n, diff, expected)
		}
	})
}

// FuzzSumIdentity verifies the odd-indexed Fibonacci sum identity:
//
//	F(1) + F(3) + F(5) + ... + F(2n-1) = F(2n)
//
// This provides cumulative error detection: if any single value in the
// sequence is wrong, the sum will fail to match F(2n).
func FuzzSumIdentity(f *testing.F) {
	seeds := []uint64{1, 2, 5, 10, 50, 100, 500}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, n uint64) {
		if n < 1 {
			n = 1
		}
		if n > 5000 {
			n = n%4999 + 1
		}

		factory := NewDefaultFactory()
		calc, err := factory.Get("fast")
		if err != nil {
			t.Fatalf("failed to get calculator: %v", err)
		}

		ctx := context.Background()

		// Compute sum of F(1) + F(3) + ... + F(2n-1)
		sum := new(big.Int)
		for k := uint64(1); k <= 2*n-1; k += 2 {
			fk, err := calc.Calculate(ctx, nil, 0, k, Options{})
			if err != nil {
				t.Skipf("calculation error for F(%d): %v", k, err)
			}
			sum.Add(sum, fk)
		}

		// Should equal F(2n)
		f2n, err := calc.Calculate(ctx, nil, 0, 2*n, Options{})
		if err != nil {
			t.Skipf("calculation error for F(%d): %v", 2*n, err)
		}

		if sum.Cmp(f2n) != 0 {
			t.Errorf("Sum identity failed for n=%d: sum=%s, F(2n)=%s", n, sum, f2n)
		}
	})
}

// FuzzExtendedConsistency compares Fast Doubling vs Matrix Exponentiation
// for a wider range of inputs (up to 200K in full mode, 50K in short mode).
// This extends coverage beyond the existing FuzzFastDoublingConsistency test
// which caps at 50K.
func FuzzExtendedConsistency(f *testing.F) {
	seeds := []uint64{94, 100, 500, 1000, 5000, 10000, 50000}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, n uint64) {
		if n < 94 {
			n = 94
		}
		if testing.Short() {
			if n > 50000 {
				n = n%49907 + 94
			}
		} else {
			if n > 200000 {
				n = n%199907 + 94
			}
		}

		factory := NewDefaultFactory()
		fastCalc, err := factory.Get("fast")
		if err != nil {
			t.Fatalf("failed to get fast calculator: %v", err)
		}
		matrixCalc, err := factory.Get("matrix")
		if err != nil {
			t.Fatalf("failed to get matrix calculator: %v", err)
		}

		ctx := context.Background()

		fastResult, err := fastCalc.Calculate(ctx, nil, 0, n, Options{})
		if err != nil {
			t.Skipf("fast calc error for n=%d: %v", n, err)
		}

		matrixResult, err := matrixCalc.Calculate(ctx, nil, 0, n, Options{})
		if err != nil {
			t.Skipf("matrix calc error for n=%d: %v", n, err)
		}

		if fastResult.Cmp(matrixResult) != 0 {
			t.Errorf("Inconsistency at n=%d: Fast Doubling != Matrix Exponentiation", n)
		}
	})
}
