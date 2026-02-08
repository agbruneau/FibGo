package bigfft

import (
	"math/big"
	"testing"
)

// FuzzFFTMulVsBigInt verifies that the public Mul(x, y) function matches
// new(big.Int).Mul(x, y) for arbitrary inputs. This exercises the full FFT
// multiplication pipeline including polynomial decomposition, forward/inverse
// transforms, and result reconstruction.
func FuzzFFTMulVsBigInt(f *testing.F) {
	// Seeds at various byte lengths to hit different FFT thresholds
	for _, size := range []int{8, 64, 256, 1024, 4096} {
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

		// Expected via standard library
		expected := new(big.Int).Mul(x, y)

		// Via FFT
		result, err := Mul(x, y)
		if err != nil {
			t.Fatalf("Mul returned error: %v", err)
		}

		if result.Cmp(expected) != 0 {
			t.Errorf("FFT Mul mismatch for %d-byte * %d-byte inputs", half, len(data)-half)
		}
	})
}

// FuzzFFTSqrVsBigInt verifies that the public Sqr(x) function matches
// new(big.Int).Mul(x, x) for arbitrary inputs. Squaring is optimized to
// perform only a single forward FFT transform, so this test validates that
// the optimization produces correct results.
func FuzzFFTSqrVsBigInt(f *testing.F) {
	for _, size := range []int{8, 64, 256, 1024, 4096} {
		data := make([]byte, size)
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 1 {
			return
		}

		x := new(big.Int).SetBytes(data)

		// Expected via standard library
		expected := new(big.Int).Mul(x, x)

		// Via FFT Sqr
		result, err := Sqr(x)
		if err != nil {
			t.Fatalf("Sqr returned error: %v", err)
		}

		if result.Cmp(expected) != 0 {
			t.Errorf("FFT Sqr mismatch for %d-byte input", len(data))
		}
	})
}

// FuzzValueSizeAdequacy verifies that valueSize(k, m, extra) returns a value
// large enough to hold the polynomial coefficients of the FFT product.
// The mathematical requirement is: n * _W >= 2*m*_W + k, where n is the
// returned value size.
func FuzzValueSizeAdequacy(f *testing.F) {
	f.Add(uint8(2), uint16(10), uint8(2))
	f.Add(uint8(5), uint16(100), uint8(2))
	f.Add(uint8(10), uint16(1000), uint8(2))

	f.Fuzz(func(t *testing.T, kByte uint8, mShort uint16, extraByte uint8) {
		k := uint(kByte%20) + 1    // k in [1, 20]
		m := int(mShort) + 1       // m in [1, 65536]
		extra := uint(extraByte%3) + 1 // extra in [1, 3]
		if extra > k {
			return
		}

		n := valueSize(k, m, extra)

		// Mathematical requirement: n * _W >= 2*m*_W + k
		required := 2*m*_W + int(k)
		actual := n * _W
		if actual < required {
			t.Errorf("valueSize(%d, %d, %d) = %d, but %d*_W=%d < required %d",
				k, m, extra, n, n, actual, required)
		}

		// The implementation rounds bits to a multiple of K = max(1<<(k-extra), _W),
		// so n (in words) must be a multiple of K/_W.
		K := 1 << (k - extra)
		if K < _W {
			K = _W
		}
		wordDivisor := K / _W
		if wordDivisor > 0 && n%wordDivisor != 0 {
			t.Errorf("valueSize(%d, %d, %d) = %d, not a multiple of %d words",
				k, m, extra, n, wordDivisor)
		}
	})
}
