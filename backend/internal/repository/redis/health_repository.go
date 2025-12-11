package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type HealthRepository interface {
	GetMetrics(ctx context.Context, accountID int64) (*HealthMetrics, error)
	UpdateMetrics(ctx context.Context, accountID int64, success bool) error
	GetHealthScore(ctx context.Context, accountID int64) (float64, error)
	SetQuarantine(ctx context.Context, accountID int64, duration time.Duration) error
	IsQuarantined(ctx context.Context, accountID int64) (bool, error)
	GetLastCheckTime(ctx context.Context, accountID int64) (time.Time, error)
}

type healthRepository struct {
	client *redis.Client
}

func NewHealthRepository(client *redis.Client) HealthRepository {
	return &healthRepository{client: client}
}

func (r *healthRepository) getKey(accountID int64) string {
	return fmt.Sprintf("health:account:%d", accountID)
}

func (r *healthRepository) GetMetrics(ctx context.Context, accountID int64) (*HealthMetrics, error) {
	key := r.getKey(accountID)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return &HealthMetrics{
				AccountID:   accountID,
				HealthScore: 1.0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get health metrics: %w", err)
	}

	var metrics HealthMetrics
	if err := json.Unmarshal([]byte(data), &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal health metrics: %w", err)
	}

	return &metrics, nil
}

func (r *healthRepository) UpdateMetrics(ctx context.Context, accountID int64, success bool) error {
	metrics, err := r.GetMetrics(ctx, accountID)
	if err != nil {
		return err
	}

	now := time.Now()
	metrics.LastCheckAt = now

	if success {
		metrics.SuccessCount++
		metrics.LastSuccessAt = now
		metrics.ConsecutiveFailures = 0
	} else {
		metrics.FailureCount++
		metrics.LastFailureAt = now
		metrics.ConsecutiveFailures++
	}

	metrics.HealthScore = CalculateHealthScore(metrics)

	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal health metrics: %w", err)
	}

	key := r.getKey(accountID)
	if err := r.client.Set(ctx, key, data, 7*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to update health metrics: %w", err)
	}

	return nil
}

func (r *healthRepository) GetHealthScore(ctx context.Context, accountID int64) (float64, error) {
	metrics, err := r.GetMetrics(ctx, accountID)
	if err != nil {
		return 0, err
	}
	return metrics.HealthScore, nil
}

func (r *healthRepository) SetQuarantine(ctx context.Context, accountID int64, duration time.Duration) error {
	metrics, err := r.GetMetrics(ctx, accountID)
	if err != nil {
		return err
	}

	metrics.QuarantineUntil = time.Now().Add(duration)
	metrics.QuarantineCount++

	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal health metrics: %w", err)
	}

	key := r.getKey(accountID)
	if err := r.client.Set(ctx, key, data, 7*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to set quarantine: %w", err)
	}

	return nil
}

func (r *healthRepository) IsQuarantined(ctx context.Context, accountID int64) (bool, error) {
	metrics, err := r.GetMetrics(ctx, accountID)
	if err != nil {
		return false, err
	}

	if metrics.QuarantineUntil.IsZero() {
		return false, nil
	}

	return time.Now().Before(metrics.QuarantineUntil), nil
}

func (r *healthRepository) GetLastCheckTime(ctx context.Context, accountID int64) (time.Time, error) {
	metrics, err := r.GetMetrics(ctx, accountID)
	if err != nil {
		return time.Time{}, err
	}
	return metrics.LastCheckAt, nil
}
