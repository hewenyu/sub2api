package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/backend/internal/metrics"
	"github.com/redis/go-redis/v9"
)

// ConcurrencyRepository defines the interface for concurrency management using Redis Sorted Sets.
type ConcurrencyRepository interface {
	Acquire(ctx context.Context, accountID int64, requestID string, leaseSeconds int) (bool, error)
	AcquireWithLimit(ctx context.Context, accountID int64, requestID string, limit, leaseSeconds int) (bool, error)
	Release(ctx context.Context, accountID int64, requestID string) error
	GetCount(ctx context.Context, accountID int64) (int64, error)
	Cleanup(ctx context.Context, accountID int64) error
}

type concurrencyRepository struct {
	semaphoreRepo SemaphoreRepository
	keyPrefix     string
	metricsType   string
}

// newConcurrencyRepository is an internal helper to create a typed concurrency repository.
// keyPrefix controls the Redis key namespace, metricsType controls Prometheus label "type".
func newConcurrencyRepository(client *redis.Client, keyPrefix, metricsType string) ConcurrencyRepository {
	semaphoreRepo, err := NewSemaphoreRepository(client)
	if err != nil {
		panic(fmt.Sprintf("failed to create semaphore repository: %v", err))
	}
	return &concurrencyRepository{
		semaphoreRepo: semaphoreRepo,
		keyPrefix:     keyPrefix,
		metricsType:   metricsType,
	}
}

// NewConcurrencyRepository creates a new concurrency repository.
// NOTE: This is kept for backward compatibility and is equivalent to NewAccountConcurrencyRepository.
func NewConcurrencyRepository(client *redis.Client) ConcurrencyRepository {
	return NewAccountConcurrencyRepository(client)
}

// NewAccountConcurrencyRepository creates a concurrency repository scoped to Codex accounts.
// Redis key pattern: concurrency:account:<account_id>
// Metrics "type" label: "account"
func NewAccountConcurrencyRepository(client *redis.Client) ConcurrencyRepository {
	return newConcurrencyRepository(client, "concurrency:account:", "account")
}

// NewAPIKeyConcurrencyRepository creates a concurrency repository scoped to API keys.
// Redis key pattern: concurrency:apikey:<api_key_id>
// Metrics "type" label: "apikey"
func NewAPIKeyConcurrencyRepository(client *redis.Client) ConcurrencyRepository {
	return newConcurrencyRepository(client, "concurrency:apikey:", "apikey")
}

func (r *concurrencyRepository) getKey(accountID int64) string {
	return fmt.Sprintf("%s%d", r.keyPrefix, accountID)
}

func (r *concurrencyRepository) acquireInternal(ctx context.Context, accountID int64, requestID string, limit, leaseSeconds int) (bool, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.SemaphoreAcquireDuration.WithLabelValues(r.metricsType).Observe(duration)
	}()

	key := r.getKey(accountID)
	acquired, err := r.semaphoreRepo.Acquire(ctx, key, requestID, limit, leaseSeconds)

	if err != nil {
		metrics.SemaphoreAcquireTotal.WithLabelValues(r.metricsType, "error").Inc()
		return false, err
	}

	if acquired {
		metrics.SemaphoreAcquireTotal.WithLabelValues(r.metricsType, "success").Inc()
		count, _ := r.GetCount(ctx, accountID)
		metrics.ConcurrencyCurrent.WithLabelValues(r.metricsType, strconv.FormatInt(accountID, 10)).Set(float64(count))
	} else {
		metrics.SemaphoreAcquireTotal.WithLabelValues(r.metricsType, "rejected").Inc()
	}

	return acquired, nil
}

// Acquire acquires a concurrency slot without enforcing a limit at the Redis layer.
// This is primarily used for account-level concurrency tracking where no hard cap is configured.
func (r *concurrencyRepository) Acquire(ctx context.Context, accountID int64, requestID string, leaseSeconds int) (bool, error) {
	// Use a very high limit to preserve previous behavior (no effective cap).
	return r.acquireInternal(ctx, accountID, requestID, 999999, leaseSeconds)
}

// AcquireWithLimit acquires a concurrency slot with an explicit limit enforced atomically in Redis.
// This is intended for API key–level concurrency limiting where MaxConcurrentRequests is known.
func (r *concurrencyRepository) AcquireWithLimit(ctx context.Context, accountID int64, requestID string, limit, leaseSeconds int) (bool, error) {
	return r.acquireInternal(ctx, accountID, requestID, limit, leaseSeconds)
}

// Release releases a concurrency slot atomically using Lua script.
func (r *concurrencyRepository) Release(ctx context.Context, accountID int64, requestID string) error {
	key := r.getKey(accountID)
	err := r.semaphoreRepo.Release(ctx, key, requestID)
	if err == nil {
		count, _ := r.GetCount(ctx, accountID)
		metrics.ConcurrencyCurrent.WithLabelValues(r.metricsType, strconv.FormatInt(accountID, 10)).Set(float64(count))
	}
	return err
}

// GetCount returns the current concurrency count for an account.
// Automatically cleans up expired records atomically.
func (r *concurrencyRepository) GetCount(ctx context.Context, accountID int64) (int64, error) {
	key := r.getKey(accountID)
	count, err := r.semaphoreRepo.GetCount(ctx, key)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}

// Cleanup removes expired concurrency records.
// Note: Cleanup is now automatic in Acquire and GetCount operations.
func (r *concurrencyRepository) Cleanup(ctx context.Context, accountID int64) error {
	_, err := r.GetCount(ctx, accountID)
	return err
}
