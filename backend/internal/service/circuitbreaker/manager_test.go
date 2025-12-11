package circuitbreaker

import (
	"testing"
)

func TestManager_GetBreaker(t *testing.T) {
	m := NewManager()

	cb1 := m.GetBreaker("codex", 1)
	cb2 := m.GetBreaker("codex", 1)

	if cb1 != cb2 {
		t.Fatal("expected same circuit breaker instance")
	}

	cb3 := m.GetBreaker("codex", 2)
	if cb1 == cb3 {
		t.Fatal("expected different circuit breaker instances for different accounts")
	}
}

func TestManager_ResetBreaker(t *testing.T) {
	m := NewManager()

	cb := m.GetBreaker("codex", 1)
	cb.failureCount = 5
	cb.state = StateOpen

	m.ResetBreaker("codex", 1)

	if cb.GetState() != StateClosed {
		t.Fatalf("expected state Closed after reset, got %v", cb.GetState())
	}
}
