package fibonacci

import (
	"sync"
	"sync/atomic"
	"testing"
)

// countingObserver tracks the number of Update calls using an atomic counter,
// making it safe for concurrent use.
type countingObserver struct {
	count atomic.Int64
}

func (o *countingObserver) Update(calcIndex int, progress float64) {
	o.count.Add(1)
}

// TestFreezeSnapshotImmutability verifies that after Freeze(), adding new
// observers does NOT affect the frozen callback. The frozen callback should
// only notify observers that were registered at the time of the freeze.
func TestFreezeSnapshotImmutability(t *testing.T) {
	subject := NewProgressSubject()
	obs1 := &countingObserver{}
	subject.Register(obs1)

	// Freeze with 1 observer
	callback := subject.Freeze(0)

	// Add another observer AFTER freeze
	obs2 := &countingObserver{}
	subject.Register(obs2)

	// Invoke frozen callback
	callback(0.5)

	// obs1 should have been notified (was in snapshot)
	if obs1.count.Load() != 1 {
		t.Errorf("obs1 should have count 1, got %d", obs1.count.Load())
	}
	// obs2 should NOT have been notified (added after freeze)
	if obs2.count.Load() != 0 {
		t.Errorf("obs2 should have count 0, got %d", obs2.count.Load())
	}
}

// TestFreezeConcurrentRegister verifies that concurrent Freeze() and Register()
// calls do not cause data races. This test should be run with -race.
func TestFreezeConcurrentRegister(t *testing.T) {
	subject := NewProgressSubject()

	var wg sync.WaitGroup

	// Goroutines registering observers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			obs := &countingObserver{}
			subject.Register(obs)
		}()
	}

	// Goroutines freezing
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cb := subject.Freeze(idx)
			cb(0.5) // invoke the callback
		}(i)
	}

	wg.Wait()
	// If we get here without race detector complaints, the test passes
}

// TestMultipleFrozenCallbacksConcurrent verifies that multiple frozen callbacks
// can be invoked concurrently without data races or lost updates.
func TestMultipleFrozenCallbacksConcurrent(t *testing.T) {
	subject := NewProgressSubject()
	obs := &countingObserver{}
	subject.Register(obs)

	// Create multiple frozen callbacks
	callbacks := make([]ProgressCallback, 10)
	for i := range callbacks {
		callbacks[i] = subject.Freeze(i)
	}

	// Invoke all concurrently
	var wg sync.WaitGroup
	for _, cb := range callbacks {
		wg.Add(1)
		go func(fn ProgressCallback) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				fn(float64(j) / 1000.0)
			}
		}(cb)
	}
	wg.Wait()

	// All invocations should have reached the observer
	expected := int64(10 * 1000)
	if obs.count.Load() != expected {
		t.Errorf("expected %d updates, got %d", expected, obs.count.Load())
	}
}
