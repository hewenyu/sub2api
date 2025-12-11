package shutdown

import (
	"context"
	"testing"
	"time"
)

func TestRequestTracker_StartEnd(t *testing.T) {
	tracker := NewRequestTracker()

	if tracker.ActiveCount() != 0 {
		t.Errorf("Expected initial active count to be 0, got %d", tracker.ActiveCount())
	}

	tracker.Start()
	if tracker.ActiveCount() != 1 {
		t.Errorf("Expected active count to be 1 after Start, got %d", tracker.ActiveCount())
	}

	tracker.Start()
	if tracker.ActiveCount() != 2 {
		t.Errorf("Expected active count to be 2 after second Start, got %d", tracker.ActiveCount())
	}

	tracker.End()
	if tracker.ActiveCount() != 1 {
		t.Errorf("Expected active count to be 1 after End, got %d", tracker.ActiveCount())
	}

	tracker.End()
	if tracker.ActiveCount() != 0 {
		t.Errorf("Expected active count to be 0 after second End, got %d", tracker.ActiveCount())
	}
}

func TestRequestTracker_ShuttingDown(t *testing.T) {
	tracker := NewRequestTracker()

	if tracker.IsShuttingDown() {
		t.Error("Expected IsShuttingDown to be false initially")
	}

	tracker.BeginShutdown()

	if !tracker.IsShuttingDown() {
		t.Error("Expected IsShuttingDown to be true after BeginShutdown")
	}
}

func TestRequestTracker_Wait(t *testing.T) {
	tracker := NewRequestTracker()

	tracker.Start()
	tracker.Start()

	done := make(chan struct{})
	go func() {
		ctx := context.Background()
		if err := tracker.Wait(ctx); err != nil {
			t.Errorf("Wait returned error: %v", err)
		}
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)

	tracker.End()
	tracker.End()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Wait did not complete in time")
	}
}

func TestRequestTracker_WaitTimeout(t *testing.T) {
	tracker := NewRequestTracker()

	tracker.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := tracker.Wait(ctx)
	if err == nil {
		t.Error("Expected Wait to return timeout error")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}

	tracker.End()
}

func TestRequestTracker_Concurrent(t *testing.T) {
	tracker := NewRequestTracker()

	const numRequests = 100

	for i := 0; i < numRequests; i++ {
		go func() {
			tracker.Start()
			time.Sleep(10 * time.Millisecond)
			tracker.End()
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := tracker.Wait(ctx); err != nil {
		t.Errorf("Wait returned error: %v", err)
	}

	if tracker.ActiveCount() != 0 {
		t.Errorf("Expected active count to be 0 after all requests complete, got %d", tracker.ActiveCount())
	}
}
