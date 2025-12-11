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
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockLimiter.On("CheckLimit", mock.Anything, int64(1), mock.AnythingOfType("string"), int64(60)).Return(true, int64(10), nil)
		mockLimiter.On("IncrementCounter", mock.Anything, int64(1), mock.AnythingOfType("string"), 60*time.Second).Return(nil)

		middleware := RateLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockLimiter.AssertExpectations(t)
	})

	t.Run("blocked when at limit", func(t *testing.T) {
		mockLimiter := new(MockRateLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:                 1,
			RateLimitPerMinute: 60,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockLimiter.On("CheckLimit", mock.Anything, int64(1), mock.AnythingOfType("string"), int64(60)).Return(false, int64(60), nil)

		middleware := RateLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.True(t, c.IsAborted())
		mockLimiter.AssertExpectations(t)
	})

	t.Run("blocked when over limit", func(t *testing.T) {
		mockLimiter := new(MockRateLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:                 1,
			RateLimitPerMinute: 60,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockLimiter.On("CheckLimit", mock.Anything, int64(1), mock.AnythingOfType("string"), int64(60)).Return(false, int64(100), nil)

		middleware := RateLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.True(t, c.IsAborted())
		mockLimiter.AssertExpectations(t)
	})

	t.Run("continues on check limit error", func(t *testing.T) {
		mockLimiter := new(MockRateLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:                 1,
			RateLimitPerMinute: 60,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockLimiter.On("CheckLimit", mock.Anything, int64(1), mock.AnythingOfType("string"), int64(60)).Return(false, int64(0), errors.New("redis error"))

		middleware := RateLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockLimiter.AssertExpectations(t)
	})

	t.Run("continues on increment error", func(t *testing.T) {
		mockLimiter := new(MockRateLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:                 1,
			RateLimitPerMinute: 60,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockLimiter.On("CheckLimit", mock.Anything, int64(1), mock.AnythingOfType("string"), int64(60)).Return(true, int64(10), nil)
		mockLimiter.On("IncrementCounter", mock.Anything, int64(1), mock.AnythingOfType("string"), 60*time.Second).Return(errors.New("redis error"))

		middleware := RateLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockLimiter.AssertExpectations(t)
	})
}
