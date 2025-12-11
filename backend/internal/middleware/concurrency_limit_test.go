package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

// MockConcurrencyTracker is a mock for ConcurrencyTracker.
type MockConcurrencyTracker struct {
	mock.Mock
}

func (m *MockConcurrencyTracker) Acquire(ctx context.Context, apiKeyID int64, requestID string, leaseSeconds int, limit int) (bool, error) { //nolint:errcheck
	args := m.Called(ctx, apiKeyID, requestID, leaseSeconds, limit)
	return args.Bool(0), args.Error(1)
}

func (m *MockConcurrencyTracker) Release(ctx context.Context, apiKeyID int64, requestID string) error { //nolint:errcheck
	args := m.Called(ctx, apiKeyID, requestID)
	return args.Error(0)
}

func (m *MockConcurrencyTracker) GetCurrentCount(ctx context.Context, apiKeyID int64) (int64, error) { //nolint:errcheck
	args := m.Called(ctx, apiKeyID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockConcurrencyTracker) Cleanup(ctx context.Context, apiKeyID int64) error { //nolint:errcheck
	args := m.Called(ctx, apiKeyID)
	return args.Error(0)
}

func TestConcurrencyLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	t.Run("no API key in context", func(t *testing.T) {
		mockTracker := new(MockConcurrencyTracker)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		middleware := ConcurrencyLimitMiddleware(mockTracker, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("no limit configured", func(t *testing.T) {
		mockTracker := new(MockConcurrencyTracker)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set(ContextKeyAPIKey, &model.APIKey{
			ID:                    1,
			MaxConcurrentRequests: 0,
		})

		middleware := ConcurrencyLimitMiddleware(mockTracker, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("allowed when under limit", func(t *testing.T) {
		mockTracker := new(MockConcurrencyTracker)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:                    1,
			MaxConcurrentRequests: 5,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockTracker.On("Acquire", mock.Anything, int64(1), mock.AnythingOfType("string"), 300, 5).Return(true, nil)
		mockTracker.On("Release", mock.Anything, int64(1), mock.AnythingOfType("string")).Return(nil)

		middleware := ConcurrencyLimitMiddleware(mockTracker, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockTracker.AssertExpectations(t)
	})

	t.Run("blocked when acquire returns false at limit", func(t *testing.T) {
		mockTracker := new(MockConcurrencyTracker)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:                    1,
			MaxConcurrentRequests: 5,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockTracker.On("Acquire", mock.Anything, int64(1), mock.AnythingOfType("string"), 300, 5).Return(false, nil)

		middleware := ConcurrencyLimitMiddleware(mockTracker, logger)
		middleware(c)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.True(t, c.IsAborted())
		mockTracker.AssertExpectations(t)
	})

	t.Run("blocked when acquire returns false over limit", func(t *testing.T) {
		mockTracker := new(MockConcurrencyTracker)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:                    1,
			MaxConcurrentRequests: 5,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockTracker.On("Acquire", mock.Anything, int64(1), mock.AnythingOfType("string"), 300, 5).Return(false, nil)

		middleware := ConcurrencyLimitMiddleware(mockTracker, logger)
		middleware(c)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.True(t, c.IsAborted())
		mockTracker.AssertExpectations(t)
	})

	t.Run("continues on acquire error", func(t *testing.T) {
		mockTracker := new(MockConcurrencyTracker)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:                    1,
			MaxConcurrentRequests: 5,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockTracker.On("Acquire", mock.Anything, int64(1), mock.AnythingOfType("string"), 300, 5).Return(false, errors.New("redis error"))

		middleware := ConcurrencyLimitMiddleware(mockTracker, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockTracker.AssertExpectations(t)
	})

	t.Run("blocked when acquire fails without error", func(t *testing.T) {
		mockTracker := new(MockConcurrencyTracker)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:                    1,
			MaxConcurrentRequests: 5,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockTracker.On("Acquire", mock.Anything, int64(1), mock.AnythingOfType("string"), 300, 5).Return(false, nil)

		middleware := ConcurrencyLimitMiddleware(mockTracker, logger)
		middleware(c)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.True(t, c.IsAborted())
		mockTracker.AssertExpectations(t)
	})
}
