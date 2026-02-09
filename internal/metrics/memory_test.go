package metrics

import "testing"

func TestMemoryCollector_Snapshot(t *testing.T) {
	t.Parallel()

	mc := NewMemoryCollector()
	snap := mc.Snapshot()

	if snap.HeapAlloc == 0 {
		t.Error("HeapAlloc should be > 0")
	}
	if snap.Sys == 0 {
		t.Error("Sys should be > 0")
	}
}

func TestMemoryCollector_Delta(t *testing.T) {
	t.Parallel()

	mc := NewMemoryCollector()
	before := mc.Snapshot()

	// Allocate some memory
	_ = make([]byte, 1024*1024) // 1 MB

	after := mc.Snapshot()

	// Sys should not decrease between snapshots
	if after.Sys < before.Sys {
		t.Error("Sys should not decrease between snapshots")
	}
}
