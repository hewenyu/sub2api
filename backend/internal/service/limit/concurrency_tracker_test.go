package limit

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockConcurrencyRepository is a mock for ConcurrencyRepository.
type MockConcurrencyRepository struct {
	mock.Mock
}

func (m *MockConcurrencyRepository) Acquire(ctx context.Context, accountID int64, requestID string, leaseSeconds int) (bool, error) { //nolint:errcheck
	args := m.Called(ctx, accountID, requestID, leaseSeconds)
	return args.Bool(0), args.Error(1)
}

func (m *MockConcurrencyRepository) Release(ctx context.Context, accountID int64, requestID string) error { //nolint:errcheck
	args := m.Called(ctx, accountID, requestID)
	return args.Error(0)
}

func (m *MockConcurrencyRepository) GetCount(ctx context.Context, accountID int64) (int64, error) { //nolint:errcheck
	args := m.Called(ctx, accountID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockConcurrencyRepository) Cleanup(ctx context.Context, accountID int64) error { //nolint:errcheck
	args := m.Called(ctx, accountID)
	return args.Error(0)
}

func TestConcurrencyTracker_Acquire(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockConcurrencyRepository)
	tracker := NewConcurrencyTracker(mockRepo, logger)
	ctx := context.Background()

	t.Run("successful acquire", func(t *testing.T) {
		mockRepo.On("Acquire", ctx, int64(1), "req1", 300).Return(true, nil).Once()

		acquired, err := tracker.Acquire(ctx, 1, "req1", 300)
		assert.NoError(t, err)
		assert.True(t, acquired)
		mockRepo.AssertExpectations(t)
	})

	t.Run("acquire error", func(t *testing.T) {
		mockRepo.On("Acquire", ctx, int64(2), "req2", 300).Return(false, errors.New("redis error")).Once()

		acquired, err := tracker.Acquire(ctx, 2, "req2", 300)
		assert.Error(t, err)
		assert.False(t, acquired)
		mockRepo.AssertExpectations(t)
	})
}

func TestConcurrencyTracker_Release(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockConcurrencyRepository)
	tracker := NewConcurrencyTracker(mockRepo, logger)
	ctx := context.Background()

	t.Run("successful release", func(t *testing.T) {
		mockRepo.On("Release", ctx, int64(1), "req1").Return(nil).Once()

		err := tracker.Release(ctx, 1, "req1")
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("release error", func(t *testing.T) {
		mockRepo.On("Release", ctx, int64(2), "req2").Return(errors.New("redis error")).Once()

		err := tracker.Release(ctx, 2, "req2")
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestConcurrencyTracker_GetCurrentCount(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockConcurrencyRepository)
	tracker := NewConcurrencyTracker(mockRepo, logger)
	ctx := context.Background()

	t.Run("successful get count", func(t *testing.T) {
		mockRepo.On("GetCount", ctx, int64(1)).Return(int64(5), nil).Once()

		count, err := tracker.GetCurrentCount(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get count error", func(t *testing.T) {
		mockRepo.On("GetCount", ctx, int64(2)).Return(int64(0), errors.New("redis error")).Once()

		count, err := tracker.GetCurrentCount(ctx, 2)
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
		mockRepo.AssertExpectations(t)
	})
}

func TestConcurrencyTracker_Cleanup(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockConcurrencyRepository)
	tracker := NewConcurrencyTracker(mockRepo, logger)
	ctx := context.Background()

	t.Run("successful cleanup", func(t *testing.T) {
		mockRepo.On("Cleanup", ctx, int64(1)).Return(nil).Once()

		err := tracker.Cleanup(ctx, 1)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("cleanup error", func(t *testing.T) {
		mockRepo.On("Cleanup", ctx, int64(2)).Return(errors.New("redis error")).Once()

		err := tracker.Cleanup(ctx, 2)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}
