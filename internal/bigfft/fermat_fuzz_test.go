package bigfft

import (
	"encoding/binary"
	"math/big"
	"slices"
	"testing"
)

// --- Helper functions ---

// makeFermatFromBytes interprets raw bytes as a fermat number of size n+1 words.
func makeFermatFromBytes(data []byte, n int) fermat {
	z := make(fermat, n+1)
	for i := 0; i < n+1 && i*8 < len(data); i++ {
		end := (i + 1) * 8
		if end > len(data) {
			end = len(data)
		}
		var word uint64
		for j := i * 8; j < end; j++ {
			word |= uint64(data[j]) << (uint(j-i*8) * 8)
		}
		z[i] = big.Word(word)
	}
	return z
}

// fermatToBigInt converts a fermat number to a *big.Int reduced modulo 2^(n*_W)+1.
// The value is z[0..n-1] as an integer plus z[n] * 2^(n*_W), all mod 2^(n*_W)+1.
func fermatToBigInt(z fermat, n int) *big.Int {
	result := new(big.Int)
	words := make([]big.Word, n)
	copy(words, z[:n])
	result.SetBits(words)
	if z[n] != 0 {
		high := new(big.Int).Lsh(new(big.Int).SetUint64(uint64(z[n])), uint(n*_W))
		result.Add(result, high)
	}
	// Reduce modulo 2^(n*_W)+1
	modulus := fermatModulus(n)
	result.Mod(result, modulus)
	return result
}

// fermatModulus returns 2^(n*_W) + 1.
func fermatModulus(n int) *big.Int {
	return new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), uint(n*_W)), big.NewInt(1))
}

// fermatEqual compares two fermat numbers for equality in the Fermat ring mod 2^(n*_W)+1.
func fermatEqual(a, b fermat, n int) bool {
	aBig := fermatToBigInt(a, n)
	bBig := fermatToBigInt(b, n)
	return aBig.Cmp(bBig) == 0
}

// --- Fuzz Tests ---

// FuzzFermatNormIdempotent verifies that norm(norm(x)) == norm(x).
// Normalization should be idempotent: applying it twice produces the same result
// as applying it once.
func FuzzFermatNormIdempotent(f *testing.F) {
	// Seeds at various sizes
	for _, size := range []int{1, 2, 5, 10, 20, 30, 50} {
		data := make([]byte, size*8+8) // size words + overflow word, 8 bytes each
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 16 {
			return // need at least 2 words
		}

		// Interpret data as a fermat number
		n := (len(data) / 8) - 1 // n data words, 1 overflow
		if n < 1 {
			return
		}

		z := make(fermat, n+1)
		for i := range z {
			if i*8+8 <= len(data) {
				z[i] = big.Word(binary.LittleEndian.Uint64(data[i*8:]))
			}
		}

		// Normalize once
		z.norm()
		// Save state
		after1 := make(fermat, len(z))
		copy(after1, z)

		// Normalize again
		z.norm()

		// Must be identical
		if !slices.Equal([]big.Word(z), []big.Word(after1)) {
			t.Errorf("norm is not idempotent: after first norm=%v, after second=%v", after1, z)
		}
	})
}

// FuzzFermatAddSubInverse verifies that Sub(Add(x, y), y) == x in the Fermat ring.
// Addition and subtraction should be inverse operations modulo 2^(n*_W)+1.
func FuzzFermatAddSubInverse(f *testing.F) {
	for _, size := range []int{1, 5, 10, 25, 35} {
		data := make([]byte, 2*(size*8+8))
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 32 {
			return
		}

		halfLen := len(data) / 2
		n := (halfLen / 8) - 1
		if n < 1 {
			return
		}

		// Create x and y
		x := makeFermatFromBytes(data[:halfLen], n)
		y := makeFermatFromBytes(data[halfLen:], n)
		x.norm()
		y.norm()

		// Save original x
		origX := make(fermat, len(x))
		copy(origX, x)

		// z = Add(x, y)
		z := make(fermat, n+1)
		z.Add(x, y)

		// w = Sub(z, y)
		w := make(fermat, n+1)
		w.Sub(z, y)

		// w should equal original x
		if !fermatEqual(w, origX, n) {
			t.Errorf("Sub(Add(x,y),y) != x for n=%d", n)
		}
	})
}

// FuzzFermatMulCommutativity verifies that Mul(x, y) == Mul(y, x).
// Multiplication should be commutative in the Fermat ring.
// Seeds span the smallMulThreshold (30) to exercise both basicMul and big.Int paths.
func FuzzFermatMulCommutativity(f *testing.F) {
	for _, size := range []int{2, 10, 25, 29, 30, 31, 35, 50} {
		data := make([]byte, 2*(size*8+8))
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 32 {
			return
		}

		halfLen := len(data) / 2
		n := (halfLen / 8) - 1
		if n < 1 || n > 100 {
			return // cap size to keep tests fast
		}

		x := makeFermatFromBytes(data[:halfLen], n)
		y := makeFermatFromBytes(data[halfLen:], n)
		x.norm()
		y.norm()

		// xy = Mul(x, y)
		bufXY := make(fermat, 2*n+2)
		xy := bufXY.Mul(x, y)

		// yx = Mul(y, x)
		bufYX := make(fermat, 2*n+2)
		yx := bufYX.Mul(y, x)

		if !fermatEqual(xy, yx, n) {
			t.Errorf("Mul not commutative for n=%d", n)
		}
	})
}

// FuzzFermatMulVsBigInt verifies fermat.Mul(x,y) == (x*y) mod (2^(n*_W)+1)
// computed via big.Int. This is the highest-value fuzz test: it checks
// correctness of the entire multiplication pipeline against an independent
// reference implementation.
func FuzzFermatMulVsBigInt(f *testing.F) {
	for _, size := range []int{2, 10, 25, 29, 30, 31, 35, 50} {
		data := make([]byte, 2*(size*8+8))
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 32 {
			return
		}

		halfLen := len(data) / 2
		n := (halfLen / 8) - 1
		if n < 1 || n > 100 {
			return
		}

		x := makeFermatFromBytes(data[:halfLen], n)
		y := makeFermatFromBytes(data[halfLen:], n)
		x.norm()
		y.norm()

		// Compute via fermat
		buf := make(fermat, 2*n+2)
		result := buf.Mul(x, y)

		// Compute via big.Int
		xBig := fermatToBigInt(x, n)
		yBig := fermatToBigInt(y, n)
		modulus := fermatModulus(n)
		expected := new(big.Int).Mul(xBig, yBig)
		expected.Mod(expected, modulus)

		// Convert result to big.Int
		resultBig := fermatToBigInt(result, n)

		if resultBig.Cmp(expected) != 0 {
			t.Errorf("Mul mismatch for n=%d: fermat=%s, bigint=%s", n, resultBig.String(), expected.String())
		}
	})
}

// FuzzFermatSqrVsMul verifies that Sqr(x) == Mul(x, x).
// The Sqr optimization (basicSqr for small n, same-pointer big.Int for large n)
// must produce identical results to the general Mul path.
func FuzzFermatSqrVsMul(f *testing.F) {
	for _, size := range []int{2, 10, 25, 29, 30, 31, 35, 50} {
		data := make([]byte, size*8+8)
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 16 {
			return
		}

		n := (len(data) / 8) - 1
		if n < 1 || n > 100 {
			return
		}

		x := makeFermatFromBytes(data, n)
		x.norm()

		// Compute via Sqr
		bufSqr := make(fermat, 8*n)
		resSqr := bufSqr.Sqr(x)

		// Compute via Mul(x, x)
		bufMul := make(fermat, 8*n)
		resMul := bufMul.Mul(x, x)

		if !fermatEqual(resSqr, resMul, n) {
			t.Errorf("Sqr(x) != Mul(x,x) for n=%d", n)
		}
	})
}

// FuzzFermatShiftModular verifies that Shift(x, k) == x * 2^k mod (2^(n*_W)+1)
// computed via big.Int. This validates the modular shift operation against a
// reference implementation.
func FuzzFermatShiftModular(f *testing.F) {
	for _, size := range []int{2, 5, 10, 20, 30} {
		data := make([]byte, size*8+8)
		// Add seeds with various shift amounts
		for _, shift := range []int16{0, 1, 7, 64, -1, -64} {
			seed := make([]byte, len(data)+2)
			copy(seed, data)
			binary.LittleEndian.PutUint16(seed[len(data):], uint16(shift))
			f.Add(seed)
		}
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 18 {
			return // need at least 2 words + 2 bytes for shift
		}

		// Last 2 bytes encode the shift amount
		shiftBytes := data[len(data)-2:]
		dataWords := data[:len(data)-2]

		k := int(int16(binary.LittleEndian.Uint16(shiftBytes)))

		n := (len(dataWords) / 8) - 1
		if n < 1 || n > 100 {
			return
		}

		x := makeFermatFromBytes(dataWords, n)
		x.norm()

		// Compute via fermat.Shift
		z := make(fermat, n+1)
		z.Shift(x, k)

		// Compute via big.Int: x * 2^k mod (2^(n*_W)+1)
		xBig := fermatToBigInt(x, n)
		modulus := fermatModulus(n)

		// Normalize k into [0, 2*n*_W)
		period := 2 * n * _W
		kNorm := k % period
		if kNorm < 0 {
			kNorm += period
		}

		var expected *big.Int
		if kNorm < n*_W {
			// x * 2^kNorm mod modulus
			shifted := new(big.Int).Lsh(xBig, uint(kNorm))
			expected = shifted.Mod(shifted, modulus)
		} else {
			// Shifting by n*_W is negation mod (2^(n*_W)+1),
			// so shift by kNorm = shift by (kNorm - n*_W) then negate.
			rem := kNorm - n*_W
			shifted := new(big.Int).Lsh(xBig, uint(rem))
			shifted.Mod(shifted, modulus)
			// Negate: modulus - shifted (unless shifted is 0)
			if shifted.Sign() == 0 {
				expected = shifted
			} else {
				expected = new(big.Int).Sub(modulus, shifted)
			}
		}

		// Convert result to big.Int
		resultBig := fermatToBigInt(z, n)

		if resultBig.Cmp(expected) != 0 {
			t.Errorf("Shift mismatch for n=%d, k=%d: fermat=%s, bigint=%s",
				n, k, resultBig.String(), expected.String())
		}
	})
}
