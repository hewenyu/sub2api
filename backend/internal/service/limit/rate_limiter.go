package limit

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
)

// WindowInfo contains rate limit information for a specific window.
type WindowInfo struct {
	Window    string
	Allowed   bool
	Remaining int64
	ResetAt   time.Time
}

// RateLimiter manages rate limiting for API keys using sliding window algorithm.
type RateLimiter interface {
	CheckLimit(ctx context.Context, apiKeyID int64, window string, maxRequests int64) (allowed bool, current int64, err error)
	IncrementCounter(ctx context.Context, apiKeyID int64, window string, ttl time.Duration) error
	CheckWindow(ctx context.Context, key string, limit int64, windowSeconds int) (allowed bool, remaining int64, resetAt time.Time, err error)
	ResetAPIKey(ctx context.Context, apiKeyID int64) error
}

type rateLimiter struct {
	rateLimitRepo redis.RateLimitRepository
	logger        *zap.Logger
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(
	rateLimitRepo redis.RateLimitRepository,
	logger *zap.Logger,
) RateLimiter {
	return &rateLimiter{
		rateLimitRepo: rateLimitRepo,
		logger:        logger,
	}
}

// CheckLimit checks if the rate limit has been exceeded.
func (l *rateLimiter) CheckLimit(ctx context.Context, apiKeyID int64, window string, maxRequests int64) (bool, int64, error) {
	if maxRequests <= 0 {
		return true, 0, nil
	}

	current, err := l.rateLimitRepo.GetCount(ctx, apiKeyID, window)
	if err != nil {
		return false, 0, fmt.Errorf("failed to get rate limit count: %w", err)
	}

	if current >= maxRequests {
		l.logger.Warn("Rate limit exceeded",
			zap.Int64("api_key_id", apiKeyID),
			zap.String("window", window),
			zap.Int64("current", current),
			zap.Int64("max", maxRequests),
		)
		return false, current, nil
	}

	return true, current, nil
}

// IncrementCounter increments the rate limit counter.
func (l *rateLimiter) IncrementCounter(ctx context.Context, apiKeyID int64, window string, ttl time.Duration) error {
	_, err := l.rateLimitRepo.Increment(ctx, apiKeyID, window, ttl)
	if err != nil {
		return fmt.Errorf("failed to increment rate limit counter: %w", err)
	}
	return nil
}

// CheckWindow checks rate limit for a single window using sliding window algorithm.
func (l *rateLimiter) CheckWindow(ctx context.Context, key string, limit int64, windowSeconds int) (bool, int64, time.Time, error) {
	if limit <= 0 {
		return true, 0, time.Time{}, nil
	}

	allowed, remaining, resetAt, err := l.rateLimitRepo.CheckAndIncrement(ctx, key, limit, windowSeconds)
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to check window: %w", err)
	}

	if !allowed {
		l.logger.Warn("Rate limit exceeded",
			zap.String("key", key),
			zap.Int64("limit", limit),
			zap.Int("window_seconds", windowSeconds),
		)
	}

	return allowed, remaining, resetAt, nil
}

// ResetAPIKey resets all rate limit windows for an API key.
func (l *rateLimiter) ResetAPIKey(ctx context.Context, apiKeyID int64) error {
	windows := []string{"60", "3600", "86400"}
	for _, window := range windows {
		key := fmt.Sprintf("ratelimit:apikey:%d:%s", apiKeyID, window)
		if err := l.rateLimitRepo.Reset(ctx, key); err != nil {
			return fmt.Errorf("failed to reset window %s: %w", window, err)
		}
	}
	return nil
}

// GetWindow calculates the current window ID based on window size in seconds.
func GetWindow(windowSeconds int) string {
	now := time.Now().Unix()
	windowID := now / int64(windowSeconds)
	return fmt.Sprintf("%d", windowID)
}
