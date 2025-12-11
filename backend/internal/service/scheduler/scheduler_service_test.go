package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	redisRepo "github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
	"github.com/Wei-Shaw/sub2api/backend/internal/scheduler"
)

// Mock repository for CodexAccount
type MockCodexAccountRepository struct {
	mock.Mock
}

func (m *MockCodexAccountRepository) Create(ctx context.Context, account *model.CodexAccount) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockCodexAccountRepository) GetByID(ctx context.Context, id int64) (*model.CodexAccount, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CodexAccount), args.Error(1)
}

func (m *MockCodexAccountRepository) GetByEmail(ctx context.Context, email string) (*model.CodexAccount, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CodexAccount), args.Error(1)
}

func (m *MockCodexAccountRepository) GetByAPIKey(ctx context.Context, apiKey string) (*model.CodexAccount, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CodexAccount), args.Error(1)
}

func (m *MockCodexAccountRepository) List(ctx context.Context, filters repository.CodexAccountFilters, offset, limit int) ([]*model.CodexAccount, int64, error) {
	args := m.Called(ctx, filters, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*model.CodexAccount), args.Get(1).(int64), args.Error(2)
}

func (m *MockCodexAccountRepository) GetSchedulable(ctx context.Context) ([]*model.CodexAccount, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.CodexAccount), args.Error(1)
}

func (m *MockCodexAccountRepository) Update(ctx context.Context, account *model.CodexAccount) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockCodexAccountRepository) UpdateFields(ctx context.Context, id int64, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockCodexAccountRepository) UpdateConcurrentRequests(ctx context.Context, id int64, delta int) error {
	args := m.Called(ctx, id, delta)
	return args.Error(0)
}

func (m *MockCodexAccountRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupSchedulerService(t *testing.T) (SchedulerService, *MockCodexAccountRepository, redisRepo.SessionRepository, redisRepo.ConcurrencyRepository, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	sessionRepo := redisRepo.NewSessionRepository(redisClient)
	concurrencyRepo := redisRepo.NewConcurrencyRepository(redisClient)
	mockCodexRepo := new(MockCodexAccountRepository)

	strategy := scheduler.NewPriorityStrategy(concurrencyRepo)
	logger := zap.NewNop()

	healthRepo := redisRepo.NewHealthRepository(redisClient)

	maxAccountConcurrency := 10

	service := NewSchedulerService(
		mockCodexRepo,
		sessionRepo,
		concurrencyRepo,
		healthRepo,
		strategy,
		1*time.Hour,
		maxAccountConcurrency,
		logger,
	)

	return service, mockCodexRepo, sessionRepo, concurrencyRepo, mr
}

func TestSchedulerService_SelectCodexAccount_BoundAccount(t *testing.T) {
	service, mockRepo, _, _, _ := setupSchedulerService(t)
	ctx := context.Background()

	boundAccountID := int64(100)
	apiKey := &model.APIKey{
		ID:                  1,
		BoundCodexAccountID: &boundAccountID,
	}

	boundAccount := &model.CodexAccount{
		ID:          boundAccountID,
		AccountType: "openai-oauth",
		IsActive:    true,
	}

	mockRepo.On("GetByID", ctx, boundAccountID).Return(boundAccount, nil)

	accountID, accountType, err := service.SelectCodexAccount(ctx, apiKey, "")
	require.NoError(t, err)
	assert.Equal(t, boundAccountID, accountID)
	assert.Equal(t, "openai-oauth", accountType)

	mockRepo.AssertExpectations(t)
}

func TestSchedulerService_SelectCodexAccount_StickySession(t *testing.T) {
	service, mockRepo, sessionRepo, _, _ := setupSchedulerService(t)
	ctx := context.Background()

	apiKey := &model.APIKey{ID: 1}
	sessionHash := "test-session-hash"

	// Set up existing session
	sessionData := redisRepo.SessionData{
		AccountID:   200,
		AccountType: "openai-responses",
		CreatedAt:   time.Now().Unix(),
		LastUsedAt:  time.Now().Unix(),
	}
	_ = sessionRepo.Set(ctx, sessionHash, sessionData, 1*time.Hour)

	stickyAccount := &model.CodexAccount{
		ID:          200,
		AccountType: "openai-responses",
		IsActive:    true,
		Schedulable: true,
	}

	mockRepo.On("GetByID", ctx, int64(200)).Return(stickyAccount, nil)

	accountID, accountType, err := service.SelectCodexAccount(ctx, apiKey, sessionHash)
	require.NoError(t, err)
	assert.Equal(t, int64(200), accountID)
	assert.Equal(t, "openai-responses", accountType)

	mockRepo.AssertExpectations(t)
}

func TestSchedulerService_SelectCodexAccount_StickySession_AutoRenew(t *testing.T) {
	service, mockRepo, sessionRepo, _, _ := setupSchedulerService(t)
	ctx := context.Background()

	apiKey := &model.APIKey{ID: 1}
	sessionHash := "test-session-hash"

	// Set up session with short TTL (5 minutes)
	sessionData := redisRepo.SessionData{
		AccountID:   300,
		AccountType: "openai-oauth",
		CreatedAt:   time.Now().Unix(),
		LastUsedAt:  time.Now().Unix(),
	}
	_ = sessionRepo.Set(ctx, sessionHash, sessionData, 5*time.Minute)

	account := &model.CodexAccount{
		ID:          300,
		AccountType: "openai-oauth",
		IsActive:    true,
		Schedulable: true,
	}

	mockRepo.On("GetByID", ctx, int64(300)).Return(account, nil)

	accountID, accountType, err := service.SelectCodexAccount(ctx, apiKey, sessionHash)
	require.NoError(t, err)
	assert.Equal(t, int64(300), accountID)
	assert.Equal(t, "openai-oauth", accountType)

	// Verify TTL was extended
	ttl, err := sessionRepo.GetTTL(ctx, sessionHash)
	require.NoError(t, err)
	assert.Greater(t, ttl, 30*time.Minute) // Should be close to 1 hour

	mockRepo.AssertExpectations(t)
}

func TestSchedulerService_SelectCodexAccount_SharedPool(t *testing.T) {
	service, mockRepo, _, _, _ := setupSchedulerService(t)
	ctx := context.Background()

	apiKey := &model.APIKey{ID: 1}

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 50, AccountType: "openai-oauth", IsActive: true, Schedulable: true},
		{ID: 2, Priority: 100, AccountType: "openai-responses", IsActive: true, Schedulable: true},
		{ID: 3, Priority: 75, AccountType: "openai-oauth", IsActive: true, Schedulable: true},
	}

	mockRepo.On("GetSchedulable", ctx).Return(candidates, nil)

	accountID, _, err := service.SelectCodexAccount(ctx, apiKey, "")
	require.NoError(t, err)
	assert.Equal(t, int64(2), accountID) // Highest priority

	mockRepo.AssertExpectations(t)
}

func TestSchedulerService_SelectCodexAccount_SharedPool_WithSession(t *testing.T) {
	service, mockRepo, sessionRepo, _, _ := setupSchedulerService(t)
	ctx := context.Background()

	apiKey := &model.APIKey{ID: 1}
	sessionHash := "new-session-hash"

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth", IsActive: true, Schedulable: true},
	}

	mockRepo.On("GetSchedulable", ctx).Return(candidates, nil)

	accountID, _, err := service.SelectCodexAccount(ctx, apiKey, sessionHash)
	require.NoError(t, err)
	assert.Equal(t, int64(1), accountID)

	// Verify session was created
	sessionData, err := sessionRepo.Get(ctx, sessionHash)
	require.NoError(t, err)
	assert.NotNil(t, sessionData)
	assert.Equal(t, int64(1), sessionData.AccountID)
	assert.Equal(t, "openai-oauth", sessionData.AccountType)

	mockRepo.AssertExpectations(t)
}

func TestSchedulerService_SelectCodexAccount_NoAvailableAccounts(t *testing.T) {
	service, mockRepo, _, _, _ := setupSchedulerService(t)
	ctx := context.Background()

	apiKey := &model.APIKey{ID: 1}

	mockRepo.On("GetSchedulable", ctx).Return([]*model.CodexAccount{}, nil)

	_, _, err := service.SelectCodexAccount(ctx, apiKey, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no available codex accounts")

	mockRepo.AssertExpectations(t)
}

func TestSchedulerService_AcquireConcurrencySlot(t *testing.T) {
	service, _, _, concurrencyRepo, _ := setupSchedulerService(t)
	ctx := context.Background()

	err := service.AcquireConcurrencySlot(ctx, 100, "req-123", 300)
	require.NoError(t, err)

	// Verify slot was acquired
	count, err := concurrencyRepo.GetCount(ctx, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestSchedulerService_ReleaseConcurrencySlot(t *testing.T) {
	service, _, _, concurrencyRepo, _ := setupSchedulerService(t)
	ctx := context.Background()

	// Acquire first
	_, _ = concurrencyRepo.Acquire(ctx, 100, "req-123", 300)

	// Release
	err := service.ReleaseConcurrencySlot(ctx, 100, "req-123")
	require.NoError(t, err)

	// Verify slot was released
	count, err := concurrencyRepo.GetCount(ctx, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestSchedulerService_GetCurrentConcurrency(t *testing.T) {
	service, _, _, concurrencyRepo, _ := setupSchedulerService(t)
	ctx := context.Background()

	// Acquire multiple slots
	_, _ = concurrencyRepo.Acquire(ctx, 100, "req-1", 300)
	_, _ = concurrencyRepo.Acquire(ctx, 100, "req-2", 300)
	_, _ = concurrencyRepo.Acquire(ctx, 100, "req-3", 300)

	count, err := service.GetCurrentConcurrency(ctx, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestSchedulerService_AcquireConcurrencySlot_RespectsAccountLimit(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer func() {
		_ = redisClient.Close()
	}()

	sessionRepo := redisRepo.NewSessionRepository(redisClient)
	concurrencyRepo := redisRepo.NewConcurrencyRepository(redisClient)
	mockCodexRepo := new(MockCodexAccountRepository)
	strategy := scheduler.NewPriorityStrategy(concurrencyRepo)
	logger := zap.NewNop()
	healthRepo := redisRepo.NewHealthRepository(redisClient)

	const maxAccountConcurrency = 2

	service := NewSchedulerService(
		mockCodexRepo,
		sessionRepo,
		concurrencyRepo,
		healthRepo,
		strategy,
		1*time.Hour,
		maxAccountConcurrency,
		logger,
	)

	ctx := context.Background()
	accountID := int64(200)

	// Acquire up to the limit successfully.
	require.NoError(t, service.AcquireConcurrencySlot(ctx, accountID, "req-1", 300))
	require.NoError(t, service.AcquireConcurrencySlot(ctx, accountID, "req-2", 300))

	// When the limit is reached, the next acquire should fail and mark overload_until.
	mockCodexRepo.On("UpdateFields", ctx, accountID, mock.MatchedBy(func(updates map[string]any) bool {
		_, hasOverload := updates["overload_until"]
		return hasOverload
	})).Return(nil)

	err := service.AcquireConcurrencySlot(ctx, accountID, "req-3", 300)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "account concurrency limit exceeded")

	// Concurrency count in Redis should not exceed the configured limit.
	count, err := concurrencyRepo.GetCount(ctx, accountID)
	require.NoError(t, err)
	assert.Equal(t, int64(maxAccountConcurrency), count)

	mockCodexRepo.AssertExpectations(t)
}

func TestSchedulerService_MarkAccountUnavailable(t *testing.T) {
	service, mockRepo, _, _, _ := setupSchedulerService(t)
	ctx := context.Background()

	mockRepo.On("UpdateFields", ctx, int64(100), mock.MatchedBy(func(updates map[string]interface{}) bool {
		_, hasRateLimitedUntil := updates["rate_limited_until"]
		status, hasStatus := updates["rate_limit_status"]
		return hasRateLimitedUntil && hasStatus && status == "test reason"
	})).Return(nil)

	err := service.MarkAccountUnavailable(ctx, 100, "test reason", 1*time.Hour)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestSchedulerService_SessionMapping(t *testing.T) {
	service, _, _, _, _ := setupSchedulerService(t)
	ctx := context.Background()

	sessionHash := "test-hash"

	// Set session
	err := service.SetSessionMapping(ctx, sessionHash, 100, "openai-oauth", 1*time.Hour)
	require.NoError(t, err)

	// Get session
	sessionData, err := service.GetSessionMapping(ctx, sessionHash)
	require.NoError(t, err)
	assert.NotNil(t, sessionData)
	assert.Equal(t, int64(100), sessionData.AccountID)
	assert.Equal(t, "openai-oauth", sessionData.AccountType)

	// Clear session
	err = service.ClearSessionMapping(ctx, sessionHash)
	require.NoError(t, err)

	// Verify cleared
	sessionData, err = service.GetSessionMapping(ctx, sessionHash)
	require.NoError(t, err)
	assert.Nil(t, sessionData)
}

func TestSchedulerService_ExtendSessionTTL(t *testing.T) {
	service, _, sessionRepo, _, _ := setupSchedulerService(t)
	ctx := context.Background()

	sessionHash := "test-hash"
	sessionData := redisRepo.SessionData{
		AccountID:   100,
		AccountType: "openai-oauth",
		CreatedAt:   time.Now().Unix(),
		LastUsedAt:  time.Now().Unix(),
	}

	// Set with short TTL
	_ = sessionRepo.Set(ctx, sessionHash, sessionData, 10*time.Second)

	// Extend
	err := service.ExtendSessionTTL(ctx, sessionHash, 1*time.Hour)
	require.NoError(t, err)

	// Verify extended
	ttl, err := sessionRepo.GetTTL(ctx, sessionHash)
	require.NoError(t, err)
	assert.Greater(t, ttl, 30*time.Minute)
}
