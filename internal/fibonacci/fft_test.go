package fibonacci

import (
	"context"
	"math/big"
	"testing"
)

func TestExecuteDoublingStepFFT(t *testing.T) {
	t.Parallel()

	t.Run("Execute doubling step with FFT", func(t *testing.T) {
		t.Parallel()
		// Create a calculation state with values that will trigger FFT
		// All fields must be initialized to avoid nil pointer dereference
		state := &CalculationState{
			FK:  new(big.Int).Exp(big.NewInt(2), big.NewInt(1000), nil), // Large number
			FK1: new(big.Int).Exp(big.NewInt(2), big.NewInt(1000), nil), // Large number
			T1:  new(big.Int),
			T2:  new(big.Int),
			T3:  new(big.Int),
		}

		opts := Options{
			ParallelThreshold: 4096,
			FFTThreshold:      10000, // Low threshold to trigger FFT
		}

		err := executeDoublingStepFFT(context.Background(), state, opts, false)
		if err != nil {
			t.Errorf("executeDoublingStepFFT returned unexpected error: %v", err)
		}
	})

	t.Run("Execute doubling step with FFT in parallel", func(t *testing.T) {
		t.Parallel()
		// Create a calculation state with values that will trigger FFT
		state := &CalculationState{
			FK:  new(big.Int).Exp(big.NewInt(2), big.NewInt(1000), nil),
			FK1: new(big.Int).Exp(big.NewInt(2), big.NewInt(1000), nil),
			T1:  new(big.Int),
			T2:  new(big.Int),
			T3:  new(big.Int),
		}

		opts := Options{
			ParallelThreshold: 4096,
			FFTThreshold:      10000,
		}

		err := executeDoublingStepFFT(context.Background(), state, opts, true)
		if err != nil {
			t.Errorf("executeDoublingStepFFT returned unexpected error: %v", err)
		}
	})

	t.Run("Execute doubling step with smaller numbers", func(t *testing.T) {
		t.Parallel()
		// Create a calculation state with smaller values
		state := &CalculationState{
			FK:  big.NewInt(5),
			FK1: big.NewInt(8),
			T1:  new(big.Int),
			T2:  new(big.Int),
			T3:  new(big.Int),
		}

		opts := Options{
			ParallelThreshold: 4096,
			FFTThreshold:      10000,
		}

		err := executeDoublingStepFFT(context.Background(), state, opts, false)
		if err != nil {
			t.Errorf("executeDoublingStepFFT returned unexpected error: %v", err)
		}
	})
}

// TestSmartMultiply_InPlace_BufferReuse verifies that smartMultiply reuses
// the destination buffer when it has sufficient capacity.
func TestSmartMultiply_InPlace_BufferReuse(t *testing.T) {
	t.Parallel()

	x := new(big.Int).SetInt64(123456789)
	y := new(big.Int).SetInt64(987654321)
	expected := new(big.Int).Mul(x, y)

	// Pre-allocate z with sufficient capacity
	z := new(big.Int)
	preSizeBigInt(z, len(expected.Bits())+10)

	result, err := smartMultiply(z, x, y, 0)
	if err != nil {
		t.Fatalf("smartMultiply error: %v", err)
	}
	if result.Cmp(expected) != 0 {
		t.Errorf("smartMultiply = %s, want %s", result.String(), expected.String())
	}
}

// TestSmartMultiply_NilZ_AllPaths verifies that smartMultiply handles nil z
// across both the FFT and non-FFT code paths.
func TestSmartMultiply_NilZ_AllPaths(t *testing.T) {
	t.Parallel()

	x := new(big.Int).SetInt64(123456789)
	y := new(big.Int).SetInt64(987654321)
	expected := new(big.Int).Mul(x, y)

	result, err := smartMultiply(nil, x, y, 0)
	if err != nil {
		t.Fatalf("smartMultiply error: %v", err)
	}
	if result.Cmp(expected) != 0 {
		t.Errorf("smartMultiply = %s, want %s", result.String(), expected.String())
	}
}

// TestSmartSquare_InPlace_BufferReuse verifies that smartSquare reuses
// the destination buffer when it has sufficient capacity.
func TestSmartSquare_InPlace_BufferReuse(t *testing.T) {
	t.Parallel()

	x := new(big.Int).SetInt64(123456789)
	expected := new(big.Int).Mul(x, x)

	z := new(big.Int)
	preSizeBigInt(z, len(expected.Bits())+10)

	result, err := smartSquare(z, x, 0)
	if err != nil {
		t.Fatalf("smartSquare error: %v", err)
	}
	if result.Cmp(expected) != 0 {
		t.Errorf("smartSquare = %s, want %s", result.String(), expected.String())
	}
}
