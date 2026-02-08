package fibonacci

import (
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestSemaphoreSaturation saturates the task semaphore with 3x its capacity
// and verifies that all goroutines complete without deadlocking.
func TestSemaphoreSaturation(t *testing.T) {
	sem := getTaskSemaphore()
	numWorkers := cap(sem) * 3 // 3x the semaphore capacity

	var wg sync.WaitGroup
	var completed atomic.Int64

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			time.Sleep(time.Millisecond) // simulate work
			<-sem                    // release
			completed.Add(1)
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if completed.Load() != int64(numWorkers) {
			t.Errorf("expected %d completions, got %d", numWorkers, completed.Load())
		}
	case <-time.After(30 * time.Second):
		t.Fatalf("DEADLOCK: only %d of %d workers completed", completed.Load(), numWorkers)
	}
}

// TestExecuteTasksSaturation tests executeTasks under high load by running
// many squaring tasks in parallel through the semaphore-limited executor.
func TestExecuteTasksSaturation(t *testing.T) {
	// Create many squaring tasks
	numTasks := 100
	results := make([]*big.Int, numTasks)

	tasks := make([]squaringTask, numTasks)
	for i := range tasks {
		results[i] = new(big.Int)
		x := big.NewInt(int64(i + 2))
		tasks[i] = squaringTask{
			dest:         &results[i],
			x:            x,
			fftThreshold: 0, // no FFT
		}
	}

	err := executeTasks[squaringTask, *squaringTask](tasks, true)
	if err != nil {
		t.Fatalf("executeTasks failed: %v", err)
	}

	// Verify results
	for i := range results {
		expected := new(big.Int).Mul(big.NewInt(int64(i+2)), big.NewInt(int64(i+2)))
		if results[i].Cmp(expected) != 0 {
			t.Errorf("task %d: got %s, want %s", i, results[i], expected)
		}
	}
}
