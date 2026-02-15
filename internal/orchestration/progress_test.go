package orchestration

import (
	"testing"

	"github.com/agbru/fibcalc/internal/progress"
)

func TestNewProgressAggregator_Positive(t *testing.T) {
	agg := NewProgressAggregator(3)
	if agg == nil {
		t.Fatal("expected non-nil aggregator for numCalculators=3")
	}
	if agg.NumCalculators() != 3 {
		t.Errorf("expected NumCalculators()=3, got %d", agg.NumCalculators())
	}
	if !agg.IsMultiCalculator() {
		t.Error("expected IsMultiCalculator()=true for 3 calculators")
	}
}

func TestNewProgressAggregator_Single(t *testing.T) {
	agg := NewProgressAggregator(1)
	if agg == nil {
		t.Fatal("expected non-nil aggregator for numCalculators=1")
	}
	if agg.IsMultiCalculator() {
		t.Error("expected IsMultiCalculator()=false for 1 calculator")
	}
}

func TestNewProgressAggregator_Zero(t *testing.T) {
	agg := NewProgressAggregator(0)
	if agg != nil {
		t.Error("expected nil aggregator for numCalculators=0")
	}
}

func TestNewProgressAggregator_Negative(t *testing.T) {
	agg := NewProgressAggregator(-1)
	if agg != nil {
		t.Error("expected nil aggregator for numCalculators=-1")
	}
}

func TestProgressAggregator_Update(t *testing.T) {
	agg := NewProgressAggregator(2)

	ap := agg.Update(progress.ProgressUpdate{CalculatorIndex: 0, Value: 0.5})
	if ap.CalculatorIndex != 0 {
		t.Errorf("expected CalculatorIndex=0, got %d", ap.CalculatorIndex)
	}
	if ap.Value != 0.5 {
		t.Errorf("expected Value=0.5, got %f", ap.Value)
	}
	// Average of [0.5, 0.0] = 0.25
	if ap.AverageProgress != 0.25 {
		t.Errorf("expected AverageProgress=0.25, got %f", ap.AverageProgress)
	}

	ap = agg.Update(progress.ProgressUpdate{CalculatorIndex: 1, Value: 0.5})
	// Average of [0.5, 0.5] = 0.5
	if ap.AverageProgress != 0.5 {
		t.Errorf("expected AverageProgress=0.5, got %f", ap.AverageProgress)
	}
}

func TestProgressAggregator_CalculateAverage(t *testing.T) {
	agg := NewProgressAggregator(2)

	avg := agg.CalculateAverage()
	if avg != 0.0 {
		t.Errorf("expected initial average=0.0, got %f", avg)
	}

	agg.Update(progress.ProgressUpdate{CalculatorIndex: 0, Value: 1.0})
	avg = agg.CalculateAverage()
	if avg != 0.5 {
		t.Errorf("expected average=0.5 after one update, got %f", avg)
	}
}

func TestProgressAggregator_GetETA(t *testing.T) {
	agg := NewProgressAggregator(1)

	// Initially ETA should be 0 (not enough data)
	eta := agg.GetETA()
	if eta != 0 {
		t.Errorf("expected initial ETA=0, got %v", eta)
	}
}

func TestDrainChannel(t *testing.T) {
	ch := make(chan progress.ProgressUpdate, 5)
	ch <- progress.ProgressUpdate{CalculatorIndex: 0, Value: 0.1}
	ch <- progress.ProgressUpdate{CalculatorIndex: 0, Value: 0.2}
	ch <- progress.ProgressUpdate{CalculatorIndex: 0, Value: 0.3}
	close(ch)

	DrainChannel(ch)
	// If we reach here without deadlock, the test passes
}

func TestDrainChannel_Empty(t *testing.T) {
	ch := make(chan progress.ProgressUpdate)
	close(ch)

	DrainChannel(ch)
	// If we reach here without deadlock, the test passes
}
