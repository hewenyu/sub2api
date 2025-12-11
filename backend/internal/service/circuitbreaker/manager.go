package circuitbreaker

import (
	"fmt"
	"sync"
)

// Manager manages circuit breakers per account
type Manager struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
}

// NewManager creates a new circuit breaker manager
func NewManager() *Manager {
	return &Manager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetBreaker returns the circuit breaker for an account, creating if needed
func (m *Manager) GetBreaker(accountType string, accountID int64) *CircuitBreaker {
	key := fmt.Sprintf("%s:%d", accountType, accountID)

	m.mu.RLock()
	cb, exists := m.breakers[key]
	m.mu.RUnlock()

	if exists {
		return cb
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if cb, exists := m.breakers[key]; exists {
		return cb
	}

	cb = New()
	m.breakers[key] = cb
	return cb
}

// ResetBreaker resets the circuit breaker for an account
func (m *Manager) ResetBreaker(accountType string, accountID int64) {
	key := fmt.Sprintf("%s:%d", accountType, accountID)

	m.mu.RLock()
	cb, exists := m.breakers[key]
	m.mu.RUnlock()

	if exists {
		cb.Reset()
	}
}
