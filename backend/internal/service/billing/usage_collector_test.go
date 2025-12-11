package billing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
)

// Mock UsageRepository
type MockUsageRepository struct {
	mock.Mock
}

func (m *MockUsageRepository) Create(ctx context.Context, usage *model.Usage) error {
	args := m.Called(ctx, usage)
	return args.Error(0)
}

func (m *MockUsageRepository) GetByID(ctx context.Context, id int64) (*model.Usage, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Usage), args.Error(1)
}

func (m *MockUsageRepository) List(ctx context.Context, filters repository.UsageFilters, offset, limit int) ([]*model.Usage, error) {
	args := m.Called(ctx, filters, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Usage), args.Error(1)
}

func (m *MockUsageRepository) Aggregate(ctx context.Context, filters repository.UsageFilters) (*model.UsageAggregate, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UsageAggregate), args.Error(1)
}

func (m *MockUsageRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Mock CostCalculator
type MockCostCalculator struct {
	mock.Mock
}

func (m *MockCostCalculator) Calculate(usage *UsageData, model string) (CostInfo, error) {
	args := m.Called(usage, model)
	return args.Get(0).(CostInfo), args.Error(1)
}

func (m *MockCostCalculator) CalculateTotalCost(inputTokens, outputTokens int, model string) (float64, error) {
	args := m.Called(inputTokens, outputTokens, model)
	return args.Get(0).(float64), args.Error(1)
}

func TestUsageCollector_CollectUsage(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("successful collection", func(t *testing.T) {
		mockRepo := new(MockUsageRepository)
		mockCalc := new(MockCostCalculator)

		usage := &UsageData{
			InputTokens:  1000,
			OutputTokens: 500,
			TotalTokens:  1500,
		}

		costInfo := CostInfo{
			InputCost:  0.03,
			OutputCost: 0.03,
			TotalCost:  0.06,
		}

		mockCalc.On("Calculate", usage, "gpt-4").Return(costInfo, nil)
		mockRepo.On("Create", ctx, mock.MatchedBy(func(u *model.Usage) bool {
			return u.APIKeyID == 1 &&
				u.AccountID == 2 &&
				u.Model == "gpt-4" &&
				u.InputTokens == 1000 &&
				u.OutputTokens == 500 &&
				u.TotalTokens == 1500 &&
				u.Cost == 0.06
		})).Return(nil)

		collector := NewUsageCollector(mockRepo, mockCalc, logger)

		err := collector.CollectUsage(ctx, 1, 2, "openai-responses", usage, "gpt-4", "", "")
		require.NoError(t, err)

		mockCalc.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("nil usage data", func(t *testing.T) {
		mockRepo := new(MockUsageRepository)
		mockCalc := new(MockCostCalculator)

		collector := NewUsageCollector(mockRepo, mockCalc, logger)

		err := collector.CollectUsage(ctx, 1, 2, "openai-responses", nil, "gpt-4", "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage data cannot be nil")
	})

	t.Run("cost calculation fails but continues with zero cost", func(t *testing.T) {
		mockRepo := new(MockUsageRepository)
		mockCalc := new(MockCostCalculator)

		usage := &UsageData{
			InputTokens:  1000,
			OutputTokens: 500,
			TotalTokens:  1500,
		}

		mockCalc.On("Calculate", usage, "unknown-model").Return(CostInfo{}, assert.AnError)
		mockRepo.On("Create", ctx, mock.MatchedBy(func(u *model.Usage) bool {
			return u.Cost == 0
		})).Return(nil)

		collector := NewUsageCollector(mockRepo, mockCalc, logger)

		err := collector.CollectUsage(ctx, 1, 2, "openai-responses", usage, "unknown-model", "", "")
		require.NoError(t, err)

		mockCalc.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository create fails", func(t *testing.T) {
		mockRepo := new(MockUsageRepository)
		mockCalc := new(MockCostCalculator)

		usage := &UsageData{
			InputTokens:  1000,
			OutputTokens: 500,
			TotalTokens:  1500,
		}

		costInfo := CostInfo{
			InputCost:  0.03,
			OutputCost: 0.03,
			TotalCost:  0.06,
		}

		mockCalc.On("Calculate", usage, "gpt-4").Return(costInfo, nil)
		mockRepo.On("Create", ctx, mock.Anything).Return(assert.AnError)

		collector := NewUsageCollector(mockRepo, mockCalc, logger)

		err := collector.CollectUsage(ctx, 1, 2, "openai-responses", usage, "gpt-4", "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create usage record")

		mockCalc.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestUsageCollector_GetDailyCost(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("get daily cost", func(t *testing.T) {
		mockRepo := new(MockUsageRepository)
		mockCalc := new(MockCostCalculator)

		date := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		expectedStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		expectedEnd := expectedStart.Add(24 * time.Hour)

		aggregate := &model.UsageAggregate{
			TotalRequests: 10,
			TotalTokens:   15000,
			TotalCost:     1.25,
		}

		apiKeyID := int64(1)
		usageType := model.UsageTypeCodex
		mockRepo.On("Aggregate", ctx, mock.MatchedBy(func(f repository.UsageFilters) bool {
			return f.APIKeyID != nil && *f.APIKeyID == apiKeyID &&
				f.UsageType != nil && *f.UsageType == usageType &&
				f.StartDate != nil && f.StartDate.Equal(expectedStart) &&
				f.EndDate != nil && f.EndDate.Equal(expectedEnd)
		})).Return(aggregate, nil)

		collector := NewUsageCollector(mockRepo, mockCalc, logger)

		cost, err := collector.GetDailyCost(ctx, 1, date)
		require.NoError(t, err)
		assert.Equal(t, 1.25, cost)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockUsageRepository)
		mockCalc := new(MockCostCalculator)

		date := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

		mockRepo.On("Aggregate", ctx, mock.Anything).
			Return((*model.UsageAggregate)(nil), assert.AnError)

		collector := NewUsageCollector(mockRepo, mockCalc, logger)

		_, err := collector.GetDailyCost(ctx, 1, date)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get daily cost")

		mockRepo.AssertExpectations(t)
	})
}

func TestUsageCollector_GetWeeklyCost(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("get weekly cost", func(t *testing.T) {
		mockRepo := new(MockUsageRepository)
		mockCalc := new(MockCostCalculator)

		startDate := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		expectedStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		expectedEnd := expectedStart.Add(7 * 24 * time.Hour)

		aggregate := &model.UsageAggregate{
			TotalRequests: 50,
			TotalTokens:   75000,
			TotalCost:     5.75,
		}

		apiKeyID := int64(1)
		usageType := model.UsageTypeCodex
		mockRepo.On("Aggregate", ctx, mock.MatchedBy(func(f repository.UsageFilters) bool {
			return f.APIKeyID != nil && *f.APIKeyID == apiKeyID &&
				f.UsageType != nil && *f.UsageType == usageType &&
				f.StartDate != nil && f.StartDate.Equal(expectedStart) &&
				f.EndDate != nil && f.EndDate.Equal(expectedEnd)
		})).Return(aggregate, nil)

		collector := NewUsageCollector(mockRepo, mockCalc, logger)

		cost, err := collector.GetWeeklyCost(ctx, 1, startDate)
		require.NoError(t, err)
		assert.Equal(t, 5.75, cost)

		mockRepo.AssertExpectations(t)
	})
}

func TestUsageCollector_GetMonthlyCost(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("get monthly cost for January 2024", func(t *testing.T) {
		mockRepo := new(MockUsageRepository)
		mockCalc := new(MockCostCalculator)

		location := time.Now().Location()
		expectedStart := time.Date(2024, 1, 1, 0, 0, 0, 0, location)
		expectedEnd := time.Date(2024, 2, 1, 0, 0, 0, 0, location)

		aggregate := &model.UsageAggregate{
			TotalRequests: 200,
			TotalTokens:   300000,
			TotalCost:     25.50,
		}

		apiKeyID := int64(1)
		usageType := model.UsageTypeCodex
		mockRepo.On("Aggregate", ctx, mock.MatchedBy(func(f repository.UsageFilters) bool {
			return f.APIKeyID != nil && *f.APIKeyID == apiKeyID &&
				f.UsageType != nil && *f.UsageType == usageType &&
				f.StartDate != nil && f.StartDate.Equal(expectedStart) &&
				f.EndDate != nil && f.EndDate.Equal(expectedEnd)
		})).Return(aggregate, nil)

		collector := NewUsageCollector(mockRepo, mockCalc, logger)

		cost, err := collector.GetMonthlyCost(ctx, 1, 2024, time.January)
		require.NoError(t, err)
		assert.Equal(t, 25.50, cost)

		mockRepo.AssertExpectations(t)
	})
}

func TestUsageCollector_GetAggregate(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("get aggregate for time range", func(t *testing.T) {
		mockRepo := new(MockUsageRepository)
		mockCalc := new(MockCostCalculator)

		startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endTime := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		expectedAggregate := &model.UsageAggregate{
			TotalRequests: 100,
			TotalTokens:   150000,
			TotalCost:     12.75,
		}

		apiKeyID := int64(1)
		usageType := model.UsageTypeCodex
		mockRepo.On("Aggregate", ctx, mock.MatchedBy(func(f repository.UsageFilters) bool {
			return f.APIKeyID != nil && *f.APIKeyID == apiKeyID &&
				f.UsageType != nil && *f.UsageType == usageType &&
				f.StartDate != nil && f.StartDate.Equal(startTime) &&
				f.EndDate != nil && f.EndDate.Equal(endTime)
		})).Return(expectedAggregate, nil)

		collector := NewUsageCollector(mockRepo, mockCalc, logger)

		aggregate, err := collector.GetAggregate(ctx, 1, startTime, endTime)
		require.NoError(t, err)
		assert.Equal(t, expectedAggregate.TotalRequests, aggregate.TotalRequests)
		assert.Equal(t, expectedAggregate.TotalTokens, aggregate.TotalTokens)
		assert.Equal(t, expectedAggregate.TotalCost, aggregate.TotalCost)

		mockRepo.AssertExpectations(t)
	})
}
