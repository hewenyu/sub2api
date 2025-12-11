package redis

import (
	"testing"
	"time"
)

func TestCalculateHealthScore(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *HealthMetrics
		expected float64
	}{
		{
			name: "no history returns 1.0",
			metrics: &HealthMetrics{
				SuccessCount: 0,
				FailureCount: 0,
			},
			expected: 1.0,
		},
		{
			name: "all successes returns 1.0",
			metrics: &HealthMetrics{
				SuccessCount: 10,
				FailureCount: 0,
			},
			expected: 1.0,
		},
		{
			name: "50% success rate",
			metrics: &HealthMetrics{
				SuccessCount: 5,
				FailureCount: 5,
			},
			expected: 0.5,
		},
		{
			name: "recent failure applies penalty",
			metrics: &HealthMetrics{
				SuccessCount:  8,
				FailureCount:  2,
				LastFailureAt: time.Now().Add(-1 * time.Hour),
			},
			expected: 0.6,
		},
		{
			name: "consecutive failures apply penalty",
			metrics: &HealthMetrics{
				SuccessCount:        7,
				FailureCount:        3,
				ConsecutiveFailures: 2,
			},
			expected: 0.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateHealthScore(tt.metrics)
			if diff := score - tt.expected; diff > 0.1 || diff < -0.1 {
				t.Errorf("CalculateHealthScore() = %v, want ~%v", score, tt.expected)
			}
		})
	}
}

func TestGetQuarantineDuration(t *testing.T) {
	tests := []struct {
		name            string
		quarantineCount int
		expected        time.Duration
	}{
		{
			name:            "first quarantine is 5 minutes",
			quarantineCount: 0,
			expected:        5 * time.Minute,
		},
		{
			name:            "second quarantine is 15 minutes",
			quarantineCount: 1,
			expected:        15 * time.Minute,
		},
		{
			name:            "third quarantine is 60 minutes",
			quarantineCount: 2,
			expected:        60 * time.Minute,
		},
		{
			name:            "subsequent quarantines are 60 minutes",
			quarantineCount: 5,
			expected:        60 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := GetQuarantineDuration(tt.quarantineCount)
			if duration != tt.expected {
				t.Errorf("GetQuarantineDuration() = %v, want %v", duration, tt.expected)
			}
		})
	}
}

func TestShouldQuarantine(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *HealthMetrics
		expected bool
	}{
		{
			name: "3 consecutive failures triggers quarantine",
			metrics: &HealthMetrics{
				ConsecutiveFailures: 3,
				HealthScore:         0.5,
			},
			expected: true,
		},
		{
			name: "health score below 0.3 triggers quarantine",
			metrics: &HealthMetrics{
				ConsecutiveFailures: 1,
				HealthScore:         0.25,
				FailureCount:        1,
			},
			expected: true,
		},
		{
			name: "healthy account not quarantined",
			metrics: &HealthMetrics{
				ConsecutiveFailures: 1,
				HealthScore:         0.8,
			},
			expected: false,
		},
		{
			name: "low score but no failures not quarantined",
			metrics: &HealthMetrics{
				ConsecutiveFailures: 0,
				HealthScore:         0.2,
				FailureCount:        0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldQuarantine(tt.metrics)
			if result != tt.expected {
				t.Errorf("ShouldQuarantine() = %v, want %v", result, tt.expected)
			}
		})
	}
}
