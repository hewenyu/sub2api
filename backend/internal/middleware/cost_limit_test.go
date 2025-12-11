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

// MockCostLimiter is a mock for CostLimiter.
type MockCostLimiter struct {
	mock.Mock
}

func (m *MockCostLimiter) CheckDailyLimit(ctx context.Context, apiKeyID int64, limit float64) (bool, float64, error) { //nolint:errcheck
	args := m.Called(ctx, apiKeyID, limit)
	return args.Bool(0), args.Get(1).(float64), args.Error(2)
}

func (m *MockCostLimiter) CheckWeeklyLimit(ctx context.Context, apiKeyID int64, limit float64) (bool, float64, error) { //nolint:errcheck
	args := m.Called(ctx, apiKeyID, limit)
	return args.Bool(0), args.Get(1).(float64), args.Error(2)
}

func (m *MockCostLimiter) CheckMonthlyLimit(ctx context.Context, apiKeyID int64, limit float64) (bool, float64, error) { //nolint:errcheck
	args := m.Called(ctx, apiKeyID, limit)
	return args.Bool(0), args.Get(1).(float64), args.Error(2)
}

func (m *MockCostLimiter) CheckTotalLimit(ctx context.Context, apiKeyID int64, limit float64) (bool, float64, error) { //nolint:errcheck
	args := m.Called(ctx, apiKeyID, limit)
	return args.Bool(0), args.Get(1).(float64), args.Error(2)
}

func TestCostLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	t.Run("no API key in context", func(t *testing.T) {
		mockLimiter := new(MockCostLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		middleware := CostLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("no limits configured", func(t *testing.T) {
		mockLimiter := new(MockCostLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set(ContextKeyAPIKey, &model.APIKey{
			ID:               1,
			DailyCostLimit:   0,
			WeeklyCostLimit:  0,
			MonthlyCostLimit: 0,
			TotalCostLimit:   0,
		})

		middleware := CostLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("allowed when under daily limit", func(t *testing.T) {
		mockLimiter := new(MockCostLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:             1,
			DailyCostLimit: 10.0,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockLimiter.On("CheckDailyLimit", mock.Anything, int64(1), 10.0).Return(true, 5.0, nil)

		middleware := CostLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockLimiter.AssertExpectations(t)
	})

	t.Run("blocked when daily limit exceeded", func(t *testing.T) {
		mockLimiter := new(MockCostLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:             1,
			DailyCostLimit: 10.0,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockLimiter.On("CheckDailyLimit", mock.Anything, int64(1), 10.0).Return(false, 15.0, nil)

		middleware := CostLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusPaymentRequired, w.Code)
		assert.True(t, c.IsAborted())
		mockLimiter.AssertExpectations(t)
	})

	t.Run("blocked when weekly limit exceeded", func(t *testing.T) {
		mockLimiter := new(MockCostLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:              1,
			WeeklyCostLimit: 50.0,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockLimiter.On("CheckWeeklyLimit", mock.Anything, int64(1), 50.0).Return(false, 60.0, nil)

		middleware := CostLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusPaymentRequired, w.Code)
		assert.True(t, c.IsAborted())
		mockLimiter.AssertExpectations(t)
	})

	t.Run("blocked when monthly limit exceeded", func(t *testing.T) {
		mockLimiter := new(MockCostLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:               1,
			MonthlyCostLimit: 200.0,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockLimiter.On("CheckMonthlyLimit", mock.Anything, int64(1), 200.0).Return(false, 250.0, nil)

		middleware := CostLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusPaymentRequired, w.Code)
		assert.True(t, c.IsAborted())
		mockLimiter.AssertExpectations(t)
	})

	t.Run("blocked when total limit exceeded", func(t *testing.T) {
		mockLimiter := new(MockCostLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:             1,
			TotalCostLimit: 1000.0,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockLimiter.On("CheckTotalLimit", mock.Anything, int64(1), 1000.0).Return(false, 1200.0, nil)

		middleware := CostLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusPaymentRequired, w.Code)
		assert.True(t, c.IsAborted())
		mockLimiter.AssertExpectations(t)
	})

	t.Run("continues on check error", func(t *testing.T) {
		mockLimiter := new(MockCostLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:             1,
			DailyCostLimit: 10.0,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockLimiter.On("CheckDailyLimit", mock.Anything, int64(1), 10.0).Return(false, 0.0, errors.New("db error"))

		middleware := CostLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockLimiter.AssertExpectations(t)
	})

	t.Run("checks multiple limits in order", func(t *testing.T) {
		mockLimiter := new(MockCostLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:               1,
			DailyCostLimit:   10.0,
			WeeklyCostLimit:  50.0,
			MonthlyCostLimit: 200.0,
			TotalCostLimit:   1000.0,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockLimiter.On("CheckDailyLimit", mock.Anything, int64(1), 10.0).Return(true, 5.0, nil)
		mockLimiter.On("CheckWeeklyLimit", mock.Anything, int64(1), 50.0).Return(true, 25.0, nil)
		mockLimiter.On("CheckMonthlyLimit", mock.Anything, int64(1), 200.0).Return(true, 100.0, nil)
		mockLimiter.On("CheckTotalLimit", mock.Anything, int64(1), 1000.0).Return(true, 500.0, nil)

		middleware := CostLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockLimiter.AssertExpectations(t)
	})

	t.Run("stops at first exceeded limit", func(t *testing.T) {
		mockLimiter := new(MockCostLimiter)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:               1,
			DailyCostLimit:   10.0,
			WeeklyCostLimit:  50.0,
			MonthlyCostLimit: 200.0,
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockLimiter.On("CheckDailyLimit", mock.Anything, int64(1), 10.0).Return(true, 5.0, nil)
		mockLimiter.On("CheckWeeklyLimit", mock.Anything, int64(1), 50.0).Return(false, 60.0, nil)

		middleware := CostLimitMiddleware(mockLimiter, logger)
		middleware(c)

		assert.Equal(t, http.StatusPaymentRequired, w.Code)
		assert.True(t, c.IsAborted())
		mockLimiter.AssertExpectations(t)
		mockLimiter.AssertNotCalled(t, "CheckMonthlyLimit")
	})
}
