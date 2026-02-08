package parallel

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// TestErrorCollectorHighContention verifies that ErrorCollector correctly
// captures exactly one error under extreme contention from 1000 concurrent
// goroutines, repeated 100 times to increase confidence.
func TestErrorCollectorHighContention(t *testing.T) {
	for round := 0; round < 100; round++ {
		var ec ErrorCollector
		var wg sync.WaitGroup
		numGoroutines := 1000

		// Barrier to start all goroutines simultaneously
		barrier := make(chan struct{})

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				<-barrier // wait for start signal
				ec.SetError(fmt.Errorf("error from goroutine %d", id))
			}(i)
		}

		close(barrier) // start all goroutines
		wg.Wait()

		err := ec.Err()
		if err == nil {
			t.Fatalf("round %d: expected an error, got nil", round)
		}

		// Verify it's one of the goroutine errors
		if !strings.HasPrefix(err.Error(), "error from goroutine ") {
			t.Errorf("round %d: unexpected error format: %v", round, err)
		}
	}
}

// TestErrorCollectorNilIgnored verifies that nil errors are correctly ignored
// even when set concurrently alongside real errors.
func TestErrorCollectorNilIgnored(t *testing.T) {
	var ec ErrorCollector
	var wg sync.WaitGroup

	// 500 goroutines setting nil, 500 setting real errors
	wg.Add(1000)
	barrier := make(chan struct{})

	for i := 0; i < 500; i++ {
		go func() {
			defer wg.Done()
			<-barrier
			ec.SetError(nil)
		}()
	}
	for i := 0; i < 500; i++ {
		go func(id int) {
			defer wg.Done()
			<-barrier
			ec.SetError(fmt.Errorf("real error %d", id))
		}(i)
	}

	close(barrier)
	wg.Wait()

	err := ec.Err()
	if err == nil {
		t.Fatal("expected a real error, got nil")
	}
	if !strings.HasPrefix(err.Error(), "real error ") {
		t.Errorf("unexpected error: %v", err)
	}
}
