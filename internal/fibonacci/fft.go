package fibonacci

import (
	"math/big"

	"github.com/agbru/fibcalc/internal/bigfft"
)

// mulFFT performs the multiplication of two *big.Int instances, x and y.
// It uses an algorithm based on the Fast Fourier Transform (FFT), and returns
// the result as a new *big.Int. This method is particularly efficient for
// multiplying very large numbers, typically offering a time complexity of
// O(N log N), where N is the number of bits in the operands. It serves as a
// high-performance alternative to the standard big.Int.Mul method for numbers
// exceeding a certain size threshold.
//
// Parameters:
//   - x: The first operand.
//   - y: The second operand.
//
// Returns:
//   - *big.Int: The product of x and y.
//   - error: An error if the calculation failed.
func mulFFT(x, y *big.Int) (*big.Int, error) {
	return bigfft.Mul(x, y)
}

// sqrFFT performs optimized squaring of a *big.Int using FFT.
// Squaring is more efficient than general multiplication because
// we only need to transform x once, saving approximately 33% of
// the FFT computation time for large numbers.
//
// Parameters:
//   - x: The operand to square.
//
// Returns:
//   - *big.Int: The result of x * x.
//   - error: An error if the calculation failed.
func sqrFFT(x *big.Int) (*big.Int, error) {
	return bigfft.Sqr(x)
}

func smartMultiply(z, x, y *big.Int, fftThreshold int) (*big.Int, error) {
	bx := x.BitLen()
	by := y.BitLen()

	// Tier 1: FFT Multiplication for very large operands
	if fftThreshold > 0 && bx > fftThreshold && by > fftThreshold {
		return bigfft.MulTo(z, x, y)
	}

	// Tier 2: math/big Multiplication (uses its own optimized Karatsuba internally)
	if z == nil {
		z = new(big.Int)
	}
	return z.Mul(x, y), nil
}

// smartSquare performs optimized squaring, choosing between math/big.Mul and
// FFT (internal/bigfft) based on the operand size.
func smartSquare(z, x *big.Int, fftThreshold int) (*big.Int, error) {
	bx := x.BitLen()

	// Tier 1: FFT Squaring for very large operands
	if fftThreshold > 0 && bx > fftThreshold {
		return bigfft.SqrTo(z, x)
	}

	// Tier 2: math/big Squaring (uses its own optimized Karatsuba internally)
	if z == nil {
		z = new(big.Int)
	}
	return z.Mul(x, x), nil
}

// executeDoublingStepFFT performs the three multiplications of a doubling step
// while minimizing redundant FFT transforms.
// It transforms F_k and F_k1 only once and then performs the calculations.
func executeDoublingStepFFT(s *CalculationState, opts Options, inParallel bool) error {
	// FK1 = F(k) * (2*F(k+1) - F(k))
	// F2k1 = F(k+1)^2 + F(k)^2

	// Determine result size bit length (approx 2 * bitlen(F_k))
	// FK1 is roughly N iterations * 2.
	// For GetFFTParams, we need words.
	fk1Words := len(s.FK1.Bits())
	targetWords := 2*fk1Words + 2
	k, m := bigfft.GetFFTParams(targetWords)

	// Transform operands once
	// Use ValueSize to get the correct coefficient length n in words
	nWords := bigfft.ValueSize(k, m, 2)
	n := nWords

	pFk := bigfft.PolyFromInt(s.FK, k, m)
	fkPoly, err := pFk.Transform(n)
	if err != nil {
		return err
	}

	pFk1 := bigfft.PolyFromInt(s.FK1, k, m)
	fk1Poly, err := pFk1.Transform(n)
	if err != nil {
		return err
	}

	pT4 := bigfft.PolyFromInt(s.T4, k, m)
	t4Poly, err := pT4.Transform(n)
	if err != nil {
		return err
	}

	if inParallel {
		// Parallel execution of pointwise multiplications and inverse transforms.
		//
		// Optimization: No clones needed. PolValues.Mul() and PolValues.Sqr() are
		// read-only on their receiver — they read p.Values[i] as operands to
		// fermat.Mul(buf, x, y) where buf is a separate temporary, so the source
		// PolValues are never modified. Multiple concurrent readers with no writers
		// is safe, eliminating two Clone() calls that previously allocated and
		// copied K*(n+1) words each (e.g., ~hundreds of KB for F(10M)).
		type result struct {
			p   *big.Int
			err error
		}
		resChan := make(chan result, 3)

		go func() {
			v, err := fkPoly.Mul(&t4Poly)
			if err != nil {
				resChan <- result{nil, err}
				return
			}
			p, err := v.InvTransform()
			if err != nil {
				resChan <- result{nil, err}
				return
			}
			p.M = m
			resChan <- result{p.IntToBigInt(s.T3), nil}
		}()

		go func() {
			v, err := fk1Poly.Sqr()
			if err != nil {
				resChan <- result{nil, err}
				return
			}
			p, err := v.InvTransform()
			if err != nil {
				resChan <- result{nil, err}
				return
			}
			p.M = m
			resChan <- result{p.IntToBigInt(s.T1), nil}
		}()

		go func() {
			v, err := fkPoly.Sqr()
			if err != nil {
				resChan <- result{nil, err}
				return
			}
			p, err := v.InvTransform()
			if err != nil {
				resChan <- result{nil, err}
				return
			}
			p.M = m
			resChan <- result{p.IntToBigInt(s.T2), nil}
		}()

		for i := 0; i < 3; i++ {
			res := <-resChan
			if res.err != nil {
				return res.err
			}
		}
		return nil
	}

	// Sequential
	v1, err := fkPoly.Mul(&t4Poly)
	if err != nil {
		return err
	}
	p1, err := v1.InvTransform()
	if err != nil {
		return err
	}
	p1.M = m
	p1.IntToBigInt(s.T3)

	v2, err := fk1Poly.Sqr()
	if err != nil {
		return err
	}
	p2, err := v2.InvTransform()
	if err != nil {
		return err
	}
	p2.M = m
	p2.IntToBigInt(s.T1)

	v3, err := fkPoly.Sqr()
	if err != nil {
		return err
	}
	p3, err := v3.InvTransform()
	if err != nil {
		return err
	}
	p3.M = m
	p3.IntToBigInt(s.T2)

	return nil
}
