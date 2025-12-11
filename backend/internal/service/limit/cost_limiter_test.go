package limit

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
)

// MockUsageRepository is a mock for UsageRepository.
type MockUsageRepository struct {
	mock.Mock
}

func (m *MockUsageRepository) Create(ctx context.Context, usage *model.Usage) error { //nolint:errcheck
	args := m.Called(ctx, usage)
	return args.Error(0)
}

func (m *MockUsageRepository) GetByID(ctx context.Context, id int64) (*model.Usage, error) { //nolint:errcheck
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Usage), args.Error(1)
}

func (m *MockUsageRepository) List(ctx context.Context, filters repository.UsageFilters, offset, limit int) ([]*model.Usage, error) { //nolint:errcheck
	args := m.Called(ctx, filters, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Usage), args.Error(1)
}

func (m *MockUsageRepository) Aggregate(ctx context.Context, filters repository.UsageFilters) (*model.UsageAggregate, error) { //nolint:errcheck
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UsageAggregate), args.Error(1)
}

func (m *MockUsageRepository) Delete(ctx context.Context, id int64) error { //nolint:errcheck
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCostLimiter_CheckDailyLimit(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockUsageRepository)
	limiter := NewCostLimiter(mockRepo, logger)
	ctx := context.Background()

	t.Run("no limit when limit is 0", func(t *testing.T) {
		allowed, current, err := limiter.CheckDailyLimit(ctx, 1, 0)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, float64(0), current)
	})

	t.Run("allowed when under limit", func(t *testing.T) {
		mockRepo.On("Aggregate", ctx, mock.MatchedBy(func(filters repository.UsageFilters) bool {
			return filters.APIKeyID != nil && *filters.APIKeyID == 1
		})).Return(&model.UsageAggregate{TotalCost: 5.0}, nil).Once()

		allowed, current, err := limiter.CheckDailyLimit(ctx, 1, 10.0)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, 5.0, current)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not allowed when at limit", func(t *testing.T) {
		mockRepo.On("Aggregate", ctx, mock.MatchedBy(func(filters repository.UsageFilters) bool {
			return filters.APIKeyID != nil && *filters.APIKeyID == 1
		})).Return(&model.UsageAggregate{TotalCost: 10.0}, nil).Once()

		allowed, current, err := limiter.CheckDailyLimit(ctx, 1, 10.0)
		assert.NoError(t, err)
		assert.False(t, allowed)
		assert.Equal(t, 10.0, current)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not allowed when over limit", func(t *testing.T) {
		mockRepo.On("Aggregate", ctx, mock.MatchedBy(func(filters repository.UsageFilters) bool {
			return filters.APIKeyID != nil && *filters.APIKeyID == 1
		})).Return(&model.UsageAggregate{TotalCost: 15.0}, nil).Once()

		allowed, current, err := limiter.CheckDailyLimit(ctx, 1, 10.0)
		assert.NoError(t, err)
		assert.False(t, allowed)
		assert.Equal(t, 15.0, current)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error when aggregate fails", func(t *testing.T) {
		mockRepo.On("Aggregate", ctx, mock.MatchedBy(func(filters repository.UsageFilters) bool {
			return filters.APIKeyID != nil && *filters.APIKeyID == 1
		})).Return((*model.UsageAggregate)(nil), errors.New("db error")).Once()

		allowed, current, err := limiter.CheckDailyLimit(ctx, 1, 10.0)
		assert.Error(t, err)
		assert.False(t, allowed)
		assert.Equal(t, float64(0), current)
		mockRepo.AssertExpectations(t)
	})
}

func TestCostLimiter_CheckWeeklyLimit(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockUsageRepository)
	limiter := NewCostLimiter(mockRepo, logger)
	ctx := context.Background()

	t.Run("no limit when limit is 0", func(t *testing.T) {
		allowed, current, err := limiter.CheckWeeklyLimit(ctx, 1, 0)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, float64(0), current)
	})

	t.Run("allowed when under limit", func(t *testing.T) {
		mockRepo.On("Aggregate", ctx, mock.MatchedBy(func(filters repository.UsageFilters) bool {
			return filters.APIKeyID != nil && *filters.APIKeyID == 1
		})).Return(&model.UsageAggregate{TotalCost: 50.0}, nil).Once()

		allowed, current, err := limiter.CheckWeeklyLimit(ctx, 1, 100.0)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, 50.0, current)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not allowed when over limit", func(t *testing.T) {
		mockRepo.On("Aggregate", ctx, mock.MatchedBy(func(filters repository.UsageFilters) bool {
			return filters.APIKeyID != nil && *filters.APIKeyID == 1
		})).Return(&model.UsageAggregate{TotalCost: 150.0}, nil).Once()

		allowed, current, err := limiter.CheckWeeklyLimit(ctx, 1, 100.0)
		assert.NoError(t, err)
		assert.False(t, allowed)
		assert.Equal(t, 150.0, current)
		mockRepo.AssertExpectations(t)
	})
}

func TestCostLimiter_CheckMonthlyLimit(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockUsageRepository)
	limiter := NewCostLimiter(mockRepo, logger)
	ctx := context.Background()

	t.Run("no limit when limit is 0", func(t *testing.T) {
		allowed, current, err := limiter.CheckMonthlyLimit(ctx, 1, 0)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, float64(0), current)
	})

	t.Run("allowed when under limit", func(t *testing.T) {
		mockRepo.On("Aggregate", ctx, mock.MatchedBy(func(filters repository.UsageFilters) bool {
			return filters.APIKeyID != nil && *filters.APIKeyID == 1
		})).Return(&model.UsageAggregate{TotalCost: 200.0}, nil).Once()

		allowed, current, err := limiter.CheckMonthlyLimit(ctx, 1, 500.0)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, 200.0, current)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not allowed when over limit", func(t *testing.T) {
		mockRepo.On("Aggregate", ctx, mock.MatchedBy(func(filters repository.UsageFilters) bool {
			return filters.APIKeyID != nil && *filters.APIKeyID == 1
		})).Return(&model.UsageAggregate{TotalCost: 600.0}, nil).Once()

		allowed, current, err := limiter.CheckMonthlyLimit(ctx, 1, 500.0)
		assert.NoError(t, err)
		assert.False(t, allowed)
		assert.Equal(t, 600.0, current)
		mockRepo.AssertExpectations(t)
	})
}

func TestCostLimiter_CheckTotalLimit(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockUsageRepository)
	limiter := NewCostLimiter(mockRepo, logger)
	ctx := context.Background()

	t.Run("no limit when limit is 0", func(t *testing.T) {
		allowed, current, err := limiter.CheckTotalLimit(ctx, 1, 0)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, float64(0), current)
	})

	t.Run("allowed when under limit", func(t *testing.T) {
		mockRepo.On("Aggregate", ctx, mock.MatchedBy(func(filters repository.UsageFilters) bool {
			return filters.APIKeyID != nil && *filters.APIKeyID == 1
		})).Return(&model.UsageAggregate{TotalCost: 800.0}, nil).Once()

		allowed, current, err := limiter.CheckTotalLimit(ctx, 1, 1000.0)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, 800.0, current)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not allowed when over limit", func(t *testing.T) {
		mockRepo.On("Aggregate", ctx, mock.MatchedBy(func(filters repository.UsageFilters) bool {
			return filters.APIKeyID != nil && *filters.APIKeyID == 1
		})).Return(&model.UsageAggregate{TotalCost: 1200.0}, nil).Once()

		allowed, current, err := limiter.CheckTotalLimit(ctx, 1, 1000.0)
		assert.NoError(t, err)
		assert.False(t, allowed)
		assert.Equal(t, 1200.0, current)
		mockRepo.AssertExpectations(t)
	})
}
