//go:build amd64

package bigfft

import "math/big"

// AddVV computes z = x + y element-wise and returns the carry.
// Delegates to math/big's internal addVV via go:linkname (declared in arith_decl.go).
func AddVV(z, x, y []big.Word) big.Word {
	if len(z) == 0 {
		return 0
	}
	return addVV(z, x, y)
}

// SubVV computes z = x - y element-wise and returns the borrow.
// Delegates to math/big's internal subVV via go:linkname (declared in arith_decl.go).
func SubVV(z, x, y []big.Word) big.Word {
	if len(z) == 0 {
		return 0
	}
	return subVV(z, x, y)
}

// AddMulVVW computes z += x * y where y is a single word.
// Delegates to math/big's internal addMulVVW via go:linkname (declared in arith_decl.go).
func AddMulVVW(z, x []big.Word, y big.Word) big.Word {
	if len(z) == 0 {
		return 0
	}
	return addMulVVW(z, x, y)
}
