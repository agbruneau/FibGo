package metrics

import "runtime"

// MemorySnapshot holds a point-in-time memory reading.
type MemorySnapshot struct {
	HeapAlloc    uint64 // bytes in use by application
	HeapSys      uint64 // bytes obtained from OS for heap
	Sys          uint64 // total bytes obtained from OS
	NumGC        uint32 // number of completed GC cycles
	PauseTotalNs uint64 // cumulative GC pause time
	HeapObjects  uint64 // number of allocated heap objects
}

// MemoryCollector reads runtime memory statistics.
type MemoryCollector struct{}

// NewMemoryCollector creates a new memory collector.
func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{}
}

// Snapshot reads current memory statistics.
func (mc *MemoryCollector) Snapshot() MemorySnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MemorySnapshot{
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		PauseTotalNs: m.PauseTotalNs,
		HeapObjects:  m.HeapObjects,
	}
}
