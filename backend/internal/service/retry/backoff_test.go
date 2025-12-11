package retry

import (
	"testing"
	"time"
)

func TestExponentialBackoff_Next(t *testing.T) {
	tests := []struct {
		name    string
		policy  RetryPolicy
		attempt int
		wantMin time.Duration
		wantMax time.Duration
	}{
		{
			name: "first attempt",
			policy: RetryPolicy{
				InitialBackoff: 100 * time.Millisecond,
				MaxBackoff:     5 * time.Second,
				Multiplier:     2.0,
				Jitter:         0.1,
			},
			attempt: 1,
			wantMin: 90 * time.Millisecond,
			wantMax: 110 * time.Millisecond,
		},
		{
			name: "second attempt",
			policy: RetryPolicy{
				InitialBackoff: 100 * time.Millisecond,
				MaxBackoff:     5 * time.Second,
				Multiplier:     2.0,
				Jitter:         0.1,
			},
			attempt: 2,
			wantMin: 180 * time.Millisecond,
			wantMax: 220 * time.Millisecond,
		},
		{
			name: "max backoff reached",
			policy: RetryPolicy{
				InitialBackoff: 100 * time.Millisecond,
				MaxBackoff:     500 * time.Millisecond,
				Multiplier:     2.0,
				Jitter:         0.1,
			},
			attempt: 10,
			wantMin: 450 * time.Millisecond,
			wantMax: 550 * time.Millisecond,
		},
		{
			name: "zero jitter",
			policy: RetryPolicy{
				InitialBackoff: 100 * time.Millisecond,
				MaxBackoff:     5 * time.Second,
				Multiplier:     2.0,
				Jitter:         0.0,
			},
			attempt: 1,
			wantMin: 100 * time.Millisecond,
			wantMax: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backoff := NewExponentialBackoff(tt.policy)
			got := backoff.Next(tt.attempt)

			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("Next(%d) = %v, want between %v and %v", tt.attempt, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestExponentialBackoff_Consistency(t *testing.T) {
	policy := RetryPolicy{
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		Multiplier:     2.0,
		Jitter:         0.1,
	}
	backoff := NewExponentialBackoff(policy)

	// Test that multiple calls with same attempt produce values within expected range
	for attempt := 1; attempt <= 5; attempt++ {
		for range 100 {
			duration := backoff.Next(attempt)
			if duration < 0 {
				t.Errorf("Next(%d) returned negative duration: %v", attempt, duration)
			}
			if duration > policy.MaxBackoff {
				t.Errorf("Next(%d) = %v, exceeds MaxBackoff %v", attempt, duration, policy.MaxBackoff)
			}
		}
	}
}
