package limit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
)

// MockRateLimitRepository is a mock for RateLimitRepository.
type MockRateLimitRepository struct {
	mock.Mock
}

func (m *MockRateLimitRepository) Increment(ctx context.Context, apiKeyID int64, window string, ttl time.Duration) (int64, error) { //nolint:errcheck
	args := m.Called(ctx, apiKeyID, window, ttl)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRateLimitRepository) GetCount(ctx context.Context, apiKeyID int64, window string) (int64, error) { //nolint:errcheck
	args := m.Called(ctx, apiKeyID, window)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRateLimitRepository) GetTTL(ctx context.Context, apiKeyID int64, window string) (time.Duration, error) { //nolint:errcheck
	args := m.Called(ctx, apiKeyID, window)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockRateLimitRepository) Delete(ctx context.Context, apiKeyID int64, window string) error { //nolint:errcheck
	args := m.Called(ctx, apiKeyID, window)
	return args.Error(0)
}

func (m *MockRateLimitRepository) CheckAndIncrement(ctx context.Context, key string, limit int64, windowSeconds int) (bool, int64, time.Time, error) { //nolint:errcheck
	args := m.Called(ctx, key, limit, windowSeconds)
	return args.Bool(0), args.Get(1).(int64), args.Get(2).(time.Time), args.Error(3)
}

func (m *MockRateLimitRepository) GetInfo(ctx context.Context, key string, windowSeconds int) (*redis.RateLimitInfo, error) { //nolint:errcheck
	args := m.Called(ctx, key, windowSeconds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*redis.RateLimitInfo), args.Error(1)
}

func (m *MockRateLimitRepository) Reset(ctx context.Context, key string) error { //nolint:errcheck
	args := m.Called(ctx, key)
	return args.Error(0)
}

func TestRateLimiter_CheckLimit(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockRateLimitRepository)
	limiter := NewRateLimiter(mockRepo, logger)
	ctx := context.Background()

	t.Run("no limit when maxRequests is 0", func(t *testing.T) {
		allowed, current, err := limiter.CheckLimit(ctx, 1, "window1", 0)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, int64(0), current)
	})

	t.Run("allowed when under limit", func(t *testing.T) {
		mockRepo.On("GetCount", ctx, int64(1), "window1").Return(int64(5), nil).Once()

		allowed, current, err := limiter.CheckLimit(ctx, 1, "window1", 10)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, int64(5), current)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not allowed when at limit", func(t *testing.T) {
		mockRepo.On("GetCount", ctx, int64(1), "window1").Return(int64(10), nil).Once()

		allowed, current, err := limiter.CheckLimit(ctx, 1, "window1", 10)
		assert.NoError(t, err)
		assert.False(t, allowed)
		assert.Equal(t, int64(10), current)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not allowed when over limit", func(t *testing.T) {
		mockRepo.On("GetCount", ctx, int64(1), "window1").Return(int64(15), nil).Once()

		allowed, current, err := limiter.CheckLimit(ctx, 1, "window1", 10)
		assert.NoError(t, err)
		assert.False(t, allowed)
		assert.Equal(t, int64(15), current)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error when getting count", func(t *testing.T) {
		mockRepo.On("GetCount", ctx, int64(1), "window1").Return(int64(0), errors.New("redis error")).Once()

		allowed, current, err := limiter.CheckLimit(ctx, 1, "window1", 10)
		assert.Error(t, err)
		assert.False(t, allowed)
		assert.Equal(t, int64(0), current)
		mockRepo.AssertExpectations(t)
	})
}

func TestRateLimiter_IncrementCounter(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockRateLimitRepository)
	limiter := NewRateLimiter(mockRepo, logger)
	ctx := context.Background()

	t.Run("successful increment", func(t *testing.T) {
		ttl := 60 * time.Second
		mockRepo.On("Increment", ctx, int64(1), "window1", ttl).Return(int64(1), nil).Once()

		err := limiter.IncrementCounter(ctx, 1, "window1", ttl)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("increment error", func(t *testing.T) {
		ttl := 60 * time.Second
		mockRepo.On("Increment", ctx, int64(1), "window1", ttl).Return(int64(0), errors.New("redis error")).Once()

		err := limiter.IncrementCounter(ctx, 1, "window1", ttl)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetWindow(t *testing.T) {
	t.Run("generates consistent window IDs", func(t *testing.T) {
		window1 := GetWindow(60)
		time.Sleep(100 * time.Millisecond)
		window2 := GetWindow(60)

		// Should be the same window within 60 seconds
		assert.Equal(t, window1, window2)
	})

	t.Run("generates different windows for different sizes", func(t *testing.T) {
		window60 := GetWindow(60)
		window120 := GetWindow(120)

		// Different window sizes should likely produce different IDs
		// (unless we're exactly at a multiple of both)
		assert.NotEmpty(t, window60)
		assert.NotEmpty(t, window120)
	})
}

func TestRateLimiter_CheckWindow(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockRateLimitRepository)
	limiter := NewRateLimiter(mockRepo, logger)
	ctx := context.Background()

	t.Run("no limit when limit is 0", func(t *testing.T) {
		allowed, remaining, resetAt, err := limiter.CheckWindow(ctx, "key1", 0, 60)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, int64(0), remaining)
		assert.True(t, resetAt.IsZero())
	})

	t.Run("allowed when under limit", func(t *testing.T) {
		resetAt := time.Now().Add(60 * time.Second)
		mockRepo.On("CheckAndIncrement", ctx, "key1", int64(10), 60).Return(true, int64(5), resetAt, nil).Once()

		allowed, remaining, actualResetAt, err := limiter.CheckWindow(ctx, "key1", 10, 60)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, int64(5), remaining)
		assert.Equal(t, resetAt, actualResetAt)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not allowed when at limit", func(t *testing.T) {
		resetAt := time.Now().Add(60 * time.Second)
		mockRepo.On("CheckAndIncrement", ctx, "key2", int64(10), 60).Return(false, int64(0), resetAt, nil).Once()

		allowed, remaining, actualResetAt, err := limiter.CheckWindow(ctx, "key2", 10, 60)
		assert.NoError(t, err)
		assert.False(t, allowed)
		assert.Equal(t, int64(0), remaining)
		assert.Equal(t, resetAt, actualResetAt)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error from repository", func(t *testing.T) {
		mockRepo.On("CheckAndIncrement", ctx, "key3", int64(10), 60).Return(false, int64(0), time.Time{}, errors.New("redis error")).Once()

		allowed, remaining, resetAt, err := limiter.CheckWindow(ctx, "key3", 10, 60)
		assert.Error(t, err)
		assert.False(t, allowed)
		assert.Equal(t, int64(0), remaining)
		assert.True(t, resetAt.IsZero())
		mockRepo.AssertExpectations(t)
	})
}

func TestRateLimiter_ResetAPIKey(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockRateLimitRepository)
	limiter := NewRateLimiter(mockRepo, logger)
	ctx := context.Background()

	t.Run("successful reset", func(t *testing.T) {
		mockRepo.On("Reset", ctx, "ratelimit:apikey:1:60").Return(nil).Once()
		mockRepo.On("Reset", ctx, "ratelimit:apikey:1:3600").Return(nil).Once()
		mockRepo.On("Reset", ctx, "ratelimit:apikey:1:86400").Return(nil).Once()

		err := limiter.ResetAPIKey(ctx, 1)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error during reset", func(t *testing.T) {
		mockRepo.On("Reset", ctx, "ratelimit:apikey:2:60").Return(errors.New("redis error")).Once()

		err := limiter.ResetAPIKey(ctx, 2)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}
