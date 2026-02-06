//go:build !amd64

// This file provides portable fallback implementations of the exported vector
// arithmetic functions for non-amd64 architectures. On amd64, these functions
// are defined in arith_amd64.go with optional AVX2 dispatch.

package bigfft

import "math/big"

// AddVV computes z = x + y element-wise and returns the carry.
// This is the portable fallback that delegates to the go:linkname function.
func AddVV(z, x, y []big.Word) big.Word {
	if len(z) == 0 {
		return 0
	}
	return addVV(z, x, y)
}

// SubVV computes z = x - y element-wise and returns the borrow.
// This is the portable fallback that delegates to the go:linkname function.
func SubVV(z, x, y []big.Word) big.Word {
	if len(z) == 0 {
		return 0
	}
	return subVV(z, x, y)
}

// AddMulVVW computes z += x * y where y is a single word.
// This is the portable fallback that delegates to the go:linkname function.
func AddMulVVW(z, x []big.Word, y big.Word) big.Word {
	if len(z) == 0 {
		return 0
	}
	return addMulVVW(z, x, y)
}
