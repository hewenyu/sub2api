package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimitInfo contains rate limit information.
type RateLimitInfo struct {
	Current   int64
	Limit     int64
	Remaining int64
	ResetAt   time.Time
}

// RateLimitRepository defines the interface for rate limiting using Redis.
type RateLimitRepository interface {
	Increment(ctx context.Context, apiKeyID int64, window string, ttl time.Duration) (int64, error)
	GetCount(ctx context.Context, apiKeyID int64, window string) (int64, error)
	GetTTL(ctx context.Context, apiKeyID int64, window string) (time.Duration, error)
	Delete(ctx context.Context, apiKeyID int64, window string) error
	CheckAndIncrement(ctx context.Context, key string, limit int64, windowSeconds int) (allowed bool, remaining int64, resetAt time.Time, err error)
	GetInfo(ctx context.Context, key string, windowSeconds int) (*RateLimitInfo, error)
	Reset(ctx context.Context, key string) error
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

const slidingWindowScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local unique_id = ARGV[4]

local min_time = now - window
redis.call('ZREMRANGEBYSCORE', key, 0, min_time)

local current = redis.call('ZCARD', key)

if current < limit then
    redis.call('ZADD', key, now, unique_id)
    redis.call('EXPIRE', key, window)
    return {1, limit - current - 1, now + window}
else
    local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
    local reset_at = tonumber(oldest[2]) + window
    return {0, 0, reset_at}
end
`

// CheckAndIncrement checks rate limit and increments if allowed using sliding window algorithm.
func (r *rateLimitRepository) CheckAndIncrement(ctx context.Context, key string, limit int64, windowSeconds int) (bool, int64, time.Time, error) {
	now := time.Now().Unix()
	uniqueID := fmt.Sprintf("%d-%d", now, time.Now().UnixNano())

	result, err := r.client.Eval(ctx, slidingWindowScript, []string{key}, limit, windowSeconds, now, uniqueID).Result()
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to execute sliding window script: %w", err)
	}

	resultSlice, ok := result.([]any)
	if !ok || len(resultSlice) != 3 {
		return false, 0, time.Time{}, fmt.Errorf("unexpected script result format")
	}

	allowedVal, _ := resultSlice[0].(int64)
	remaining, _ := resultSlice[1].(int64)
	resetAtUnix, _ := resultSlice[2].(int64)
	resetAt := time.Unix(resetAtUnix, 0)

	return allowedVal == 1, remaining, resetAt, nil
}

// GetInfo returns rate limit information without incrementing.
func (r *rateLimitRepository) GetInfo(ctx context.Context, key string, windowSeconds int) (*RateLimitInfo, error) {
	now := time.Now().Unix()
	minTime := now - int64(windowSeconds)

	// Remove expired entries
	err := r.client.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", minTime)).Err()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to remove expired entries: %w", err)
	}

	// Get current count
	current, err := r.client.ZCard(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get current count: %w", err)
	}

	// Get oldest entry for reset time
	var resetAt time.Time
	if current > 0 {
		oldest, err := r.client.ZRangeWithScores(ctx, key, 0, 0).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get oldest entry: %w", err)
		}
		if len(oldest) > 0 {
			resetAt = time.Unix(int64(oldest[0].Score)+int64(windowSeconds), 0)
		}
	} else {
		resetAt = time.Now().Add(time.Duration(windowSeconds) * time.Second)
	}

	return &RateLimitInfo{
		Current:   current,
		Limit:     0,
		Remaining: 0,
		ResetAt:   resetAt,
	}, nil
}

// Reset deletes all rate limit data for a key.
func (r *rateLimitRepository) Reset(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
