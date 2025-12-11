package redis

import (
	"math"
	"time"
)

type HealthMetrics struct {
	AccountID           int64     `json:"account_id"`
	SuccessCount        int64     `json:"success_count"`
	FailureCount        int64     `json:"failure_count"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
	LastCheckAt         time.Time `json:"last_check_at"`
	LastSuccessAt       time.Time `json:"last_success_at"`
	LastFailureAt       time.Time `json:"last_failure_at"`
	HealthScore         float64   `json:"health_score"`
	QuarantineUntil     time.Time `json:"quarantine_until"`
	QuarantineCount     int       `json:"quarantine_count"`
}

func CalculateHealthScore(metrics *HealthMetrics) float64 {
	if metrics.SuccessCount == 0 && metrics.FailureCount == 0 {
		return 1.0
	}

	total := float64(metrics.SuccessCount + metrics.FailureCount)
	baseScore := float64(metrics.SuccessCount) / total

	penalty := 0.0
	if !metrics.LastFailureAt.IsZero() {
		hoursSinceFailure := time.Since(metrics.LastFailureAt).Hours()
		if hoursSinceFailure < 24 {
			penalty = 0.2 * (1.0 - hoursSinceFailure/24.0)
		}
	}

	consecutivePenalty := math.Min(float64(metrics.ConsecutiveFailures)*0.15, 0.5)

	score := baseScore - penalty - consecutivePenalty
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

func GetQuarantineDuration(quarantineCount int) time.Duration {
	switch quarantineCount {
	case 0:
		return 5 * time.Minute
	case 1:
		return 15 * time.Minute
	default:
		return 60 * time.Minute
	}
}

func ShouldQuarantine(metrics *HealthMetrics) bool {
	if metrics.ConsecutiveFailures >= 3 {
		return true
	}
	if metrics.HealthScore < 0.3 && metrics.FailureCount > 0 {
		return true
	}
	return false
}
