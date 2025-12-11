package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimitRepository defines the interface for rate limiting using Redis.
type RateLimitRepository interface {
	Increment(ctx context.Context, apiKeyID int64, window string, ttl time.Duration) (int64, error)
	GetCount(ctx context.Context, apiKeyID int64, window string) (int64, error)
	GetTTL(ctx context.Context, apiKeyID int64, window string) (time.Duration, error)
	Delete(ctx context.Context, apiKeyID int64, window string) error
}

type rateLimitRepository struct {
	client *redis.Client
}

// NewRateLimitRepository creates a new rate limit repository.
func NewRateLimitRepository(client *redis.Client) RateLimitRepository {
	return &rateLimitRepository{client: client}
}

func (r *rateLimitRepository) getKey(apiKeyID int64, window string) string {
	return fmt.Sprintf("rate_limit:%d:%s", apiKeyID, window)
}

// Increment increments the counter and sets TTL if it's the first increment.
func (r *rateLimitRepository) Increment(ctx context.Context, apiKeyID int64, window string, ttl time.Duration) (int64, error) {
	key := r.getKey(apiKeyID, window)

	// Increment counter
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment rate limit counter: %w", err)
	}

	// If first increment, set TTL
	if count == 1 {
		r.client.Expire(ctx, key, ttl)
	}

	return count, nil
}

// GetCount returns the current count for a window.
func (r *rateLimitRepository) GetCount(ctx context.Context, apiKeyID int64, window string) (int64, error) {
	key := r.getKey(apiKeyID, window)

	count, err := r.client.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get rate limit count: %w", err)
	}

	return count, nil
}

// GetTTL returns the TTL for a window.
func (r *rateLimitRepository) GetTTL(ctx context.Context, apiKeyID int64, window string) (time.Duration, error) {
	key := r.getKey(apiKeyID, window)
	return r.client.TTL(ctx, key).Result()
}

// Delete deletes the counter for a window.
func (r *rateLimitRepository) Delete(ctx context.Context, apiKeyID int64, window string) error {
	key := r.getKey(apiKeyID, window)
	return r.client.Del(ctx, key).Err()
}
