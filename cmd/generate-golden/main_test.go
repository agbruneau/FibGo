package main

import (
	"math/big"
	"testing"
)

// TestFibBig tests the oracle Fibonacci function with known values.
func TestFibBig(t *testing.T) {
	tests := []struct {
		name     string
		n        uint64
		expected string
	}{
		{"F(0) base case", 0, "0"},
		{"F(1) base case", 1, "1"},
		{"F(2) first non-trivial", 2, "1"},
		{"F(3)", 3, "2"},
		{"F(4)", 4, "3"},
		{"F(5)", 5, "5"},
		{"F(10)", 10, "55"},
		{"F(20)", 20, "6765"},
		{"F(50)", 50, "12586269025"},
		{"F(92) max uint64", 92, "7540113804746346429"},
		{"F(93) overflows uint64", 93, "12200160415121876738"},
		{"F(94) requires big.Int", 94, "19740274219868223167"},
		{"F(100)", 100, "354224848179261915075"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fibBig(tt.n)
			if result.String() != tt.expected {
				t.Errorf("fibBig(%d) = %s, want %s", tt.n, result.String(), tt.expected)
			}
		})
	}
}

// TestFibBig_Properties tests mathematical properties of Fibonacci numbers.
func TestFibBig_Properties(t *testing.T) {
	t.Run("F(n) + F(n+1) = F(n+2)", func(t *testing.T) {
		for n := uint64(0); n < 50; n++ {
			fn := fibBig(n)
			fn1 := fibBig(n + 1)
			fn2 := fibBig(n + 2)

			sum := new(big.Int).Add(fn, fn1)
			if sum.Cmp(fn2) != 0 {
				t.Errorf("F(%d) + F(%d) = %s, but F(%d) = %s",
					n, n+1, sum.String(), n+2, fn2.String())
			}
		}
	})

	t.Run("F(n) is monotonically increasing for n >= 1", func(t *testing.T) {
		prev := fibBig(1)
		for n := uint64(2); n <= 100; n++ {
			curr := fibBig(n)
			if curr.Cmp(prev) < 0 {
				t.Errorf("F(%d) = %s < F(%d) = %s, should be increasing",
					n, curr.String(), n-1, prev.String())
			}
			prev = curr
		}
	})
}

// TestFibBig_LargeValues tests larger Fibonacci numbers.
func TestFibBig_LargeValues(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large value tests in short mode")
	}

	tests := []struct {
		name     string
		n        uint64
		expected string
	}{
		{
			"F(1000)",
			1000,
			"43466557686937456435688527675040625802564660517371780402481729089536555417949051890403879840079255169295922593080322634775209689623239873322471161642996440906533187938298969649928516003704476137795166849228875",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fibBig(tt.n)
			if result.String() != tt.expected {
				t.Errorf("fibBig(%d) result mismatch\ngot:  %s\nwant: %s",
					tt.n, result.String(), tt.expected)
			}
		})
	}
}
