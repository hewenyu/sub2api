package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

// MockRateLimiter is a mock for RateLimiter.
type MockRateLimiter struct {
	mock.Mock
}

func (m *MockRateLimiter) CheckLimit(ctx context.Context, apiKeyID int64, window string, maxRequests int64) (bool, int64, error) { //nolint:errcheck
	args := m.Called(ctx, apiKeyID, window, maxRequests)
	return args.Bool(0), args.Get(1).(int64), args.Error(2)
}

func (m *MockRateLimiter) IncrementCounter(ctx context.Context, apiKeyID int64, window string, ttl time.Duration) error { //nolint:errcheck
	args := m.Called(ctx, apiKeyID, window, ttl)
	return args.Error(0)
}

func (m *MockRateLimiter) CheckWindow(ctx context.Context, key string, limit int64, windowSeconds int) (bool, int64, time.Time, error) { //nolint:errcheck
	args := m.Called(ctx, key, limit, windowSeconds)
	return args.Bool(0), args.Get(1).(int64), args.Get(2).(time.Time), args.Error(3)
}

func (m *MockRateLimiter) ResetAPIKey(ctx context.Context, apiKeyID int64) error { //nolint:errcheck
	args := m.Called(ctx, apiKeyID)
	return args.Error(0)
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	t.Run("no API key in context", func(t *testing.T) {
		mockLimiter := new(MockRateLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		middleware := RateLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("no limit configured", func(t *testing.T) {
		mockLimiter := new(MockRateLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set(ContextKeyAPIKey, &model.APIKey{
			ID:                 1,
			RateLimitPerMinute: 0,
		})

		middleware := RateLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("allowed when under limit", func(t *testing.T) {
		mockLimiter := new(MockRateLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:                 1,
			RateLimitPerMinute: 60,
			RateLimitPerHour:   1000,
			RateLimitPerDay:    10000,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		resetAt := time.Now().Add(60 * time.Second)
		mockLimiter.On("CheckWindow", mock.Anything, "ratelimit:apikey:1:60", int64(60), 60).Return(true, int64(10), resetAt, nil)
		mockLimiter.On("CheckWindow", mock.Anything, "ratelimit:apikey:1:3600", int64(1000), 3600).Return(true, int64(100), resetAt, nil)
		mockLimiter.On("CheckWindow", mock.Anything, "ratelimit:apikey:1:86400", int64(10000), 86400).Return(true, int64(1000), resetAt, nil)

		middleware := RateLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "60", w.Header().Get("X-RateLimit-Limit-Minute"))
		assert.Equal(t, "10", w.Header().Get("X-RateLimit-Remaining-Minute"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset-Minute"))
		mockLimiter.AssertExpectations(t)
	})

	t.Run("blocked when minute limit exceeded", func(t *testing.T) {
		mockLimiter := new(MockRateLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:                 1,
			RateLimitPerMinute: 60,
			RateLimitPerHour:   1000,
			RateLimitPerDay:    10000,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		resetAt := time.Now().Add(60 * time.Second)
		mockLimiter.On("CheckWindow", mock.Anything, "ratelimit:apikey:1:60", int64(60), 60).Return(false, int64(0), resetAt, nil)

		middleware := RateLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.True(t, c.IsAborted())
		mockLimiter.AssertExpectations(t)
	})

	t.Run("blocked when hour limit exceeded", func(t *testing.T) {
		mockLimiter := new(MockRateLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:                 1,
			RateLimitPerMinute: 60,
			RateLimitPerHour:   1000,
			RateLimitPerDay:    10000,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		resetAt := time.Now().Add(60 * time.Second)
		mockLimiter.On("CheckWindow", mock.Anything, "ratelimit:apikey:1:60", int64(60), 60).Return(true, int64(10), resetAt, nil)
		mockLimiter.On("CheckWindow", mock.Anything, "ratelimit:apikey:1:3600", int64(1000), 3600).Return(false, int64(0), resetAt, nil)

		middleware := RateLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.True(t, c.IsAborted())
		mockLimiter.AssertExpectations(t)
	})

	t.Run("continues on check window error", func(t *testing.T) {
		mockLimiter := new(MockRateLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:                 1,
			RateLimitPerMinute: 60,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockLimiter.On("CheckWindow", mock.Anything, "ratelimit:apikey:1:60", int64(60), 60).Return(false, int64(0), time.Time{}, errors.New("redis error"))

		middleware := RateLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockLimiter.AssertExpectations(t)
	})
}
