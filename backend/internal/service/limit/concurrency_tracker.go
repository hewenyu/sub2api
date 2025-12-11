package limit

import (
	"context"

	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
)

// ConcurrencyTracker manages concurrent request tracking for API keys.
type ConcurrencyTracker interface {
	Acquire(ctx context.Context, apiKeyID int64, requestID string, leaseSeconds int) (bool, error)
	Release(ctx context.Context, apiKeyID int64, requestID string) error
	GetCurrentCount(ctx context.Context, apiKeyID int64) (int64, error)
	Cleanup(ctx context.Context, apiKeyID int64) error
}

type concurrencyTracker struct {
	concurrencyRepo redis.ConcurrencyRepository
	logger          *zap.Logger
}

// NewConcurrencyTracker creates a new concurrency tracker.
func NewConcurrencyTracker(
	concurrencyRepo redis.ConcurrencyRepository,
	logger *zap.Logger,
) ConcurrencyTracker {
	return &concurrencyTracker{
		concurrencyRepo: concurrencyRepo,
		logger:          logger,
	}
}

// Acquire acquires a concurrency slot for an API key.
func (t *concurrencyTracker) Acquire(ctx context.Context, apiKeyID int64, requestID string, leaseSeconds int) (bool, error) {
	acquired, err := t.concurrencyRepo.Acquire(ctx, apiKeyID, requestID, leaseSeconds)
	if err != nil {
		t.logger.Error("Failed to acquire concurrency slot",
			zap.Int64("api_key_id", apiKeyID),
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return false, err
	}

	if acquired {
		t.logger.Debug("Concurrency slot acquired",
			zap.Int64("api_key_id", apiKeyID),
			zap.String("request_id", requestID),
		)
	}

	return acquired, nil
}

// Release releases a concurrency slot.
func (t *concurrencyTracker) Release(ctx context.Context, apiKeyID int64, requestID string) error {
	err := t.concurrencyRepo.Release(ctx, apiKeyID, requestID)
	if err != nil {
		t.logger.Error("Failed to release concurrency slot",
			zap.Int64("api_key_id", apiKeyID),
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return err
	}

	t.logger.Debug("Concurrency slot released",
		zap.Int64("api_key_id", apiKeyID),
		zap.String("request_id", requestID),
	)

	return nil
}

// GetCurrentCount returns the current concurrency count for an API key.
func (t *concurrencyTracker) GetCurrentCount(ctx context.Context, apiKeyID int64) (int64, error) {
	count, err := t.concurrencyRepo.GetCount(ctx, apiKeyID)
	if err != nil {
		t.logger.Error("Failed to get concurrency count",
			zap.Int64("api_key_id", apiKeyID),
			zap.Error(err),
		)
		return 0, err
	}

	return count, nil
}

// Cleanup removes expired concurrency records.
func (t *concurrencyTracker) Cleanup(ctx context.Context, apiKeyID int64) error {
	err := t.concurrencyRepo.Cleanup(ctx, apiKeyID)
	if err != nil {
		t.logger.Error("Failed to cleanup concurrency records",
			zap.Int64("api_key_id", apiKeyID),
			zap.Error(err),
		)
		return err
	}

	return nil
}
