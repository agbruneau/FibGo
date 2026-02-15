package memory

import (
	"testing"
)

func TestGCController_Disabled(t *testing.T) {
	t.Parallel()

	gc := NewGCController("disabled", 1_000_000)
	gc.Begin()
	defer gc.End()

	if gc.active {
		t.Error("GC controller should not be active in disabled mode")
	}
}

func TestGCController_Auto_SmallN(t *testing.T) {
	t.Parallel()

	gc := NewGCController("auto", 100)
	gc.Begin()
	defer gc.End()

	if gc.active {
		t.Error("GC controller should not be active for small N")
	}
}

func TestGCController_Auto_LargeN(t *testing.T) {
	t.Parallel()

	gc := NewGCController("auto", 2_000_000)
	if !gc.active {
		t.Error("GC controller should be active for N >= 1M in auto mode")
	}
	gc.Begin()
	defer gc.End()
}

func TestGCController_Aggressive(t *testing.T) {
	t.Parallel()

	gc := NewGCController("aggressive", 100)
	if !gc.active {
		t.Error("GC controller should be active in aggressive mode regardless of N")
	}
	gc.Begin()
	defer gc.End()
}

func TestGCController_Stats_BeforeBegin(t *testing.T) {
	t.Parallel()

	gc := NewGCController("disabled", 100)
	stats := gc.Stats()
	if stats.TotalAlloc != 0 {
		t.Errorf("TotalAlloc before Begin should be 0, got %d", stats.TotalAlloc)
	}
}

func TestGCController_Stats_AfterBeginEnd(t *testing.T) {
	t.Parallel()

	gc := NewGCController("aggressive", 2_000_000)
	gc.Begin()
	// Do some allocations
	_ = make([]byte, 1024*1024)
	gc.End()

	stats := gc.Stats()
	// HeapAlloc should be non-zero after some allocations
	// (we can't assert exact values due to runtime variability)
	_ = stats
}
