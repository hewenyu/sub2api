package circuitbreaker

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/backend/internal/metrics"
)

var (
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrTooManyRequests = errors.New("too many requests in half-open state")
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker wraps HTTP client calls with circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           State
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	openUntil       time.Time

	// Configuration
	name             string
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	halfOpenMaxReqs  int
	halfOpenReqs     int
}

// New creates a new circuit breaker with default settings
func New() *CircuitBreaker {
	return NewWithName("default")
}

// NewWithName creates a new circuit breaker with a specific name for metrics
func NewWithName(name string) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:             name,
		state:            StateClosed,
		failureThreshold: 5,
		successThreshold: 2,
		timeout:          30 * time.Second,
		halfOpenMaxReqs:  3,
	}
	cb.updateStateMetric()
	return cb
}

// Do executes the HTTP request with circuit breaker protection
func (cb *CircuitBreaker) Do(client *http.Client, req *http.Request) (*http.Response, error) {
	if err := cb.beforeRequest(); err != nil {
		return nil, err
	}

	resp, err := client.Do(req)

	cb.afterRequest(err, resp)

	return resp, err
}

func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateOpen:
		if time.Now().After(cb.openUntil) {
			oldState := cb.state
			cb.state = StateHalfOpen
			cb.halfOpenReqs = 0
			cb.successCount = 0
			cb.failureCount = 0
			cb.recordStateTransition(oldState, cb.state)
		} else {
			metrics.CircuitBreakerRejectedTotal.WithLabelValues(cb.name).Inc()
			return ErrCircuitOpen
		}
	case StateHalfOpen:
		if cb.halfOpenReqs >= cb.halfOpenMaxReqs {
			metrics.CircuitBreakerRejectedTotal.WithLabelValues(cb.name).Inc()
			return ErrTooManyRequests
		}
		cb.halfOpenReqs++
	}

	return nil
}

func (cb *CircuitBreaker) afterRequest(err error, resp *http.Response) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	failed := err != nil || (resp != nil && resp.StatusCode >= 500)

	switch cb.state {
	case StateClosed:
		if failed {
			cb.failureCount++
			cb.lastFailureTime = time.Now()
			if cb.failureCount >= cb.failureThreshold {
				oldState := cb.state
				cb.state = StateOpen
				cb.openUntil = time.Now().Add(cb.timeout)
				cb.recordStateTransition(oldState, cb.state)
			}
		} else {
			cb.failureCount = 0
		}

	case StateHalfOpen:
		if failed {
			oldState := cb.state
			cb.state = StateOpen
			cb.openUntil = time.Now().Add(cb.timeout)
			cb.halfOpenReqs = 0
			cb.recordStateTransition(oldState, cb.state)
		} else {
			cb.successCount++
			if cb.successCount >= cb.successThreshold {
				oldState := cb.state
				cb.state = StateClosed
				cb.failureCount = 0
				cb.successCount = 0
				cb.halfOpenReqs = 0
				cb.recordStateTransition(oldState, cb.state)
			}
		}
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	oldState := cb.state
	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.halfOpenReqs = 0
	if oldState != StateClosed {
		cb.recordStateTransition(oldState, cb.state)
	}
}

func (cb *CircuitBreaker) updateStateMetric() {
	var stateValue float64
	switch cb.state {
	case StateClosed:
		stateValue = 0
	case StateOpen:
		stateValue = 1
	case StateHalfOpen:
		stateValue = 2
	}
	metrics.CircuitBreakerState.WithLabelValues(cb.name).Set(stateValue)
}

func (cb *CircuitBreaker) recordStateTransition(from, to State) {
	fromStr := stateToString(from)
	toStr := stateToString(to)
	metrics.CircuitBreakerTransitionsTotal.WithLabelValues(cb.name, fromStr, toStr).Inc()
	cb.updateStateMetric()
}

func stateToString(s State) string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half_open"
	default:
		return "unknown"
	}
}
