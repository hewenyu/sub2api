package retry

import (
	"math"
	"math/rand"
	"time"
)

// BackoffStrategy calculates backoff duration for retry attempts
type BackoffStrategy interface {
	Next(attempt int) time.Duration
}

// RetryPolicy defines retry behavior configuration
type RetryPolicy struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
	Jitter         float64
}

// DefaultRetryPolicy returns production-ready retry policy
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		Multiplier:     2.0,
		Jitter:         0.1,
	}
}

type exponentialBackoff struct {
	policy RetryPolicy
}

// NewExponentialBackoff creates exponential backoff strategy with jitter
func NewExponentialBackoff(policy RetryPolicy) BackoffStrategy {
	return &exponentialBackoff{policy: policy}
}

// Next calculates next backoff duration with exponential growth and jitter
func (e *exponentialBackoff) Next(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	// Calculate exponential backoff: initial * multiplier^(attempt-1)
	backoff := float64(e.policy.InitialBackoff) * math.Pow(e.policy.Multiplier, float64(attempt-1))

	// Cap at max backoff
	if backoff > float64(e.policy.MaxBackoff) {
		backoff = float64(e.policy.MaxBackoff)
	}

	// Apply jitter: backoff * (1 +/- jitter)
	if e.policy.Jitter > 0 {
		jitterRange := backoff * e.policy.Jitter
		jitter := (rand.Float64()*2 - 1) * jitterRange
		backoff += jitter
	}

	// Ensure non-negative
	if backoff < 0 {
		backoff = 0
	}

	return time.Duration(backoff)
}
