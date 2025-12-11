package shutdown

import (
	"context"
	"sync"
	"sync/atomic"
)

type RequestTracker struct {
	wg           sync.WaitGroup
	activeCount  int64
	shuttingDown atomic.Bool
}

func NewRequestTracker() *RequestTracker {
	return &RequestTracker{}
}

func (t *RequestTracker) Start() {
	t.wg.Add(1)
	atomic.AddInt64(&t.activeCount, 1)
}

func (t *RequestTracker) End() {
	atomic.AddInt64(&t.activeCount, -1)
	t.wg.Done()
}

func (t *RequestTracker) ActiveCount() int {
	return int(atomic.LoadInt64(&t.activeCount))
}

func (t *RequestTracker) IsShuttingDown() bool {
	return t.shuttingDown.Load()
}

func (t *RequestTracker) BeginShutdown() {
	t.shuttingDown.Store(true)
}

func (t *RequestTracker) Wait(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		t.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
