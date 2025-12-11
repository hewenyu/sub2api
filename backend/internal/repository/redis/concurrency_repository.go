package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ConcurrencyRepository defines the interface for concurrency management using Redis Sorted Sets.
type ConcurrencyRepository interface {
	Acquire(ctx context.Context, accountID int64, requestID string, leaseSeconds int) (bool, error)
	Release(ctx context.Context, accountID int64, requestID string) error
	GetCount(ctx context.Context, accountID int64) (int64, error)
	Cleanup(ctx context.Context, accountID int64) error
}

type concurrencyRepository struct {
	client *redis.Client
}

// NewConcurrencyRepository creates a new concurrency repository.
func NewConcurrencyRepository(client *redis.Client) ConcurrencyRepository {
	return &concurrencyRepository{client: client}
}

func (r *concurrencyRepository) getKey(accountID int64) string {
	return fmt.Sprintf("concurrency:account:%d", accountID)
}

// Acquire acquires a concurrency slot by adding to a Sorted Set.
// Score is the expiry timestamp.
func (r *concurrencyRepository) Acquire(ctx context.Context, accountID int64, requestID string, leaseSeconds int) (bool, error) {
	key := r.getKey(accountID)
	expiryTimestamp := time.Now().Add(time.Duration(leaseSeconds) * time.Second).Unix()

	// Cleanup expired records first
	if err := r.Cleanup(ctx, accountID); err != nil {
		return false, err
	}

	// Add to Sorted Set
	if err := r.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(expiryTimestamp),
		Member: requestID,
	}).Err(); err != nil {
		return false, fmt.Errorf("failed to acquire concurrency slot: %w", err)
	}

	// Set key expiry to prevent memory leak
	r.client.Expire(ctx, key, time.Duration(leaseSeconds+3600)*time.Second)

	return true, nil
}

// Release releases a concurrency slot by removing from Sorted Set.
func (r *concurrencyRepository) Release(ctx context.Context, accountID int64, requestID string) error {
	key := r.getKey(accountID)

	if err := r.client.ZRem(ctx, key, requestID).Err(); err != nil {
		return fmt.Errorf("failed to release concurrency slot: %w", err)
	}

	return nil
}

// GetCount returns the current concurrency count for an account.
// Automatically cleans up expired records.
func (r *concurrencyRepository) GetCount(ctx context.Context, accountID int64) (int64, error) {
	key := r.getKey(accountID)
	now := time.Now().Unix()

	// Cleanup expired records first (best effort, ignore errors)
	_ = r.Cleanup(ctx, accountID)

	// Count unexpired records (score >= now)
	count, err := r.client.ZCount(ctx, key, fmt.Sprintf("%d", now), "+inf").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get concurrency count: %w", err)
	}

	return count, nil
}

// Cleanup removes expired concurrency records.
func (r *concurrencyRepository) Cleanup(ctx context.Context, accountID int64) error {
	key := r.getKey(accountID)
	now := time.Now().Unix()

	// Remove records with score < now (exclusive upper bound)
	if err := r.client.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("(%d", now)).Err(); err != nil {
		return fmt.Errorf("failed to cleanup expired concurrency records: %w", err)
	}

	return nil
}
