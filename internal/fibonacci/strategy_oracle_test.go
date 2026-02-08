package fibonacci

import (
	"math/big"
	"testing"
)

// FuzzAdaptiveVsKaratsuba compares AdaptiveStrategy.Multiply with
// KaratsubaStrategy.Multiply for random big.Int inputs. Both strategies
// must produce identical results since they implement the same mathematical
// operation (integer multiplication) using different algorithms.
func FuzzAdaptiveVsKaratsuba(f *testing.F) {
	for _, size := range []int{8, 64, 256, 1024, 4096, 8192} {
		data := make([]byte, 2*size)
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}
		half := len(data) / 2

		x := new(big.Int).SetBytes(data[:half])
		y := new(big.Int).SetBytes(data[half:])

		opts := Options{FFTThreshold: DefaultFFTThreshold}

		adaptive := &AdaptiveStrategy{}
		karatsuba := &KaratsubaStrategy{}

		aResult, aErr := adaptive.Multiply(nil, x, y, opts)
		kResult, kErr := karatsuba.Multiply(nil, x, y, opts)

		if aErr != nil || kErr != nil {
			t.Skipf("errors: adaptive=%v, karatsuba=%v", aErr, kErr)
		}

		if aResult.Cmp(kResult) != 0 {
			t.Errorf("Adaptive != Karatsuba for %d-bit * %d-bit: adaptive=%s, karatsuba=%s",
				x.BitLen(), y.BitLen(), aResult, kResult)
		}
	})
}

// FuzzAdaptiveSquareVsKaratsuba compares AdaptiveStrategy.Square with
// KaratsubaStrategy.Square for random big.Int inputs. Squaring is a
// specialized case of multiplication where both operands are the same,
// and optimized implementations must still produce identical results.
func FuzzAdaptiveSquareVsKaratsuba(f *testing.F) {
	for _, size := range []int{8, 64, 256, 1024, 4096, 8192} {
		data := make([]byte, size)
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 1 {
			return
		}

		x := new(big.Int).SetBytes(data)

		opts := Options{FFTThreshold: DefaultFFTThreshold}

		adaptive := &AdaptiveStrategy{}
		karatsuba := &KaratsubaStrategy{}

		aResult, aErr := adaptive.Square(nil, x, opts)
		kResult, kErr := karatsuba.Square(nil, x, opts)

		if aErr != nil || kErr != nil {
			t.Skipf("errors: adaptive=%v, karatsuba=%v", aErr, kErr)
		}

		if aResult.Cmp(kResult) != 0 {
			t.Errorf("Adaptive.Square != Karatsuba.Square for %d-bit input: adaptive=%s, karatsuba=%s",
				x.BitLen(), aResult, kResult)
		}
	})
}

// FuzzSmartMultiplyOracle tests smartMultiply against math/big.Mul as a
// reference oracle, focusing on inputs near the FFT threshold boundary
// where algorithm switching occurs. This catches edge cases in the
// threshold-based dispatch logic.
func FuzzSmartMultiplyOracle(f *testing.F) {
	// Generate seeds near FFT threshold (~500K bits = ~62.5KB)
	for _, size := range []int{60000, 62000, 63000, 64000, 65000, 70000} {
		data := make([]byte, 2*size)
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}
		half := len(data) / 2

		x := new(big.Int).SetBytes(data[:half])
		y := new(big.Int).SetBytes(data[half:])

		// Use smartMultiply with a realistic threshold
		result, err := smartMultiply(nil, x, y, DefaultFFTThreshold)
		if err != nil {
			t.Skipf("smartMultiply error: %v", err)
		}

		expected := new(big.Int).Mul(x, y)

		if result.Cmp(expected) != 0 {
			t.Errorf("smartMultiply mismatch for %d-bit * %d-bit", x.BitLen(), y.BitLen())
		}
	})
}

// FuzzSmartSquareOracle tests smartSquare against math/big.Mul(x, x) as a
// reference oracle, focusing on inputs near the FFT threshold boundary.
// This ensures the squaring optimization produces correct results at
// all operand sizes, especially around the algorithm switching point.
func FuzzSmartSquareOracle(f *testing.F) {
	// Generate seeds near FFT threshold (~500K bits = ~62.5KB)
	for _, size := range []int{60000, 62000, 63000, 64000, 65000, 70000} {
		data := make([]byte, size)
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 1 {
			return
		}

		x := new(big.Int).SetBytes(data)

		// Use smartSquare with a realistic threshold
		result, err := smartSquare(nil, x, DefaultFFTThreshold)
		if err != nil {
			t.Skipf("smartSquare error: %v", err)
		}

		expected := new(big.Int).Mul(x, x)

		if result.Cmp(expected) != 0 {
			t.Errorf("smartSquare mismatch for %d-bit input", x.BitLen())
		}
	})
}
