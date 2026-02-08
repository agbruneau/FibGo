package orchestration

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/agbru/fibcalc/internal/config"
	"github.com/agbru/fibcalc/internal/fibonacci"
)

// mockCalculator simulates various calculator behaviors for deadlock testing.
type mockCalculator struct {
	name     string
	behavior string // "instant", "slow", "error", "progress_flood"
	delay    time.Duration
}

func (m *mockCalculator) Calculate(ctx context.Context, progressChan chan<- fibonacci.ProgressUpdate, calcIndex int, n uint64, opts fibonacci.Options) (*big.Int, error) {
	switch m.behavior {
	case "instant":
		return big.NewInt(1), nil
	case "slow":
		for i := 0; i < 100; i++ {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case progressChan <- fibonacci.ProgressUpdate{CalculatorIndex: calcIndex, Value: float64(i) / 100.0}:
			default: // non-blocking
			}
			time.Sleep(m.delay)
		}
		return big.NewInt(1), nil
	case "error":
		return nil, fmt.Errorf("simulated error")
	case "progress_flood":
		// Flood the progress channel
		for i := 0; i < 10000; i++ {
			select {
			case progressChan <- fibonacci.ProgressUpdate{CalculatorIndex: calcIndex, Value: float64(i) / 10000.0}:
			default:
			}
		}
		return big.NewInt(1), nil
	}
	return big.NewInt(1), nil
}

func (m *mockCalculator) Name() string { return m.name }

// mockProgressReporter that just drains the channel.
type mockProgressReporter struct{}

func (m *mockProgressReporter) DisplayProgress(wg *sync.WaitGroup, progressChan <-chan fibonacci.ProgressUpdate, numCalcs int, out io.Writer) {
	defer wg.Done()
	for range progressChan {
	} // drain until closed
}

// TestOrchestrationNoDeadlock_MixedBehaviors verifies that ExecuteCalculations
// completes without deadlocking under various calculator behavior combinations.
func TestOrchestrationNoDeadlock_MixedBehaviors(t *testing.T) {
	testCases := []struct {
		name        string
		calculators []fibonacci.Calculator
	}{
		{
			name: "all_instant",
			calculators: []fibonacci.Calculator{
				&mockCalculator{name: "c1", behavior: "instant"},
				&mockCalculator{name: "c2", behavior: "instant"},
				&mockCalculator{name: "c3", behavior: "instant"},
			},
		},
		{
			name: "mixed_instant_and_slow",
			calculators: []fibonacci.Calculator{
				&mockCalculator{name: "fast", behavior: "instant"},
				&mockCalculator{name: "slow", behavior: "slow", delay: time.Millisecond},
			},
		},
		{
			name: "mixed_with_errors",
			calculators: []fibonacci.Calculator{
				&mockCalculator{name: "ok", behavior: "instant"},
				&mockCalculator{name: "err", behavior: "error"},
			},
		},
		{
			name: "progress_flood",
			calculators: []fibonacci.Calculator{
				&mockCalculator{name: "flood1", behavior: "progress_flood"},
				&mockCalculator{name: "flood2", behavior: "progress_flood"},
			},
		},
		{
			name: "single_calculator",
			calculators: []fibonacci.Calculator{
				&mockCalculator{name: "solo", behavior: "instant"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cfg := config.AppConfig{N: 100}
			reporter := &mockProgressReporter{}

			done := make(chan struct{})
			go func() {
				defer close(done)
				ExecuteCalculations(ctx, tc.calculators, cfg, reporter, io.Discard)
			}()

			select {
			case <-done:
				// Success - no deadlock
			case <-time.After(10 * time.Second):
				t.Fatal("DEADLOCK: ExecuteCalculations did not complete within timeout")
			}
		})
	}
}

// TestOrchestrationNoDeadlock_ContextCancellation verifies that cancelling
// the context during execution does not cause a deadlock.
func TestOrchestrationNoDeadlock_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calcs := []fibonacci.Calculator{
		&mockCalculator{name: "slow1", behavior: "slow", delay: 100 * time.Millisecond},
		&mockCalculator{name: "slow2", behavior: "slow", delay: 100 * time.Millisecond},
	}

	cfg := config.AppConfig{N: 100}
	reporter := &mockProgressReporter{}

	done := make(chan struct{})
	go func() {
		defer close(done)
		ExecuteCalculations(ctx, calcs, cfg, reporter, io.Discard)
	}()

	// Cancel after a short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("DEADLOCK after context cancellation")
	}
}
