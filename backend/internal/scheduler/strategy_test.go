package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	redisRepo "github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
)

func setupTestConcurrencyRepo(t *testing.T) (redisRepo.ConcurrencyRepository, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return redisRepo.NewConcurrencyRepository(client), mr
}

func TestPriorityStrategy_Select_ByPriority(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy := NewPriorityStrategy(concurrencyRepo)
	ctx := context.Background()

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 50, AccountType: "openai-oauth"},
		{ID: 2, Priority: 100, AccountType: "openai-oauth"}, // Highest priority
		{ID: 3, Priority: 75, AccountType: "openai-responses"},
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	selected, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), selected.ID) // Highest priority wins
}

func TestPriorityStrategy_Select_ByConcurrency(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy := NewPriorityStrategy(concurrencyRepo)
	ctx := context.Background()

	// Set up different concurrency levels
	_, _ = concurrencyRepo.Acquire(ctx, 1, "req-1", 300)
	_, _ = concurrencyRepo.Acquire(ctx, 1, "req-2", 300) // 2 concurrent
	_, _ = concurrencyRepo.Acquire(ctx, 2, "req-3", 300) // 1 concurrent

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth"},     // Same priority, higher concurrency
		{ID: 2, Priority: 100, AccountType: "openai-responses"}, // Same priority, lower concurrency
		{ID: 3, Priority: 100, AccountType: "openai-oauth"},     // Same priority, no concurrency
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	selected, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), selected.ID) // Lowest concurrency (0) wins
}

func TestPriorityStrategy_Select_ByLastUsedAt(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy := NewPriorityStrategy(concurrencyRepo)
	ctx := context.Background()

	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth", LastUsedAt: &now},
		{ID: 2, Priority: 100, AccountType: "openai-responses", LastUsedAt: &twoHoursAgo}, // Least recently used
		{ID: 3, Priority: 100, AccountType: "openai-oauth", LastUsedAt: &oneHourAgo},
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	selected, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), selected.ID) // Least recently used wins
}

func TestPriorityStrategy_Select_NeverUsedFirst(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy := NewPriorityStrategy(concurrencyRepo)
	ctx := context.Background()

	now := time.Now()

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth", LastUsedAt: &now},
		{ID: 2, Priority: 100, AccountType: "openai-responses", LastUsedAt: nil}, // Never used
		{ID: 3, Priority: 100, AccountType: "openai-oauth", LastUsedAt: nil},     // Never used
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	selected, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)
	// Should select one of the never-used accounts (ID 2 or 3)
	assert.True(t, selected.ID == 2 || selected.ID == 3)
}

func TestPriorityStrategy_Select_MultiLevelSorting(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy := NewPriorityStrategy(concurrencyRepo)
	ctx := context.Background()

	// Set up concurrency
	_, _ = concurrencyRepo.Acquire(ctx, 3, "req-1", 300)
	_, _ = concurrencyRepo.Acquire(ctx, 5, "req-2", 300)
	_, _ = concurrencyRepo.Acquire(ctx, 5, "req-3", 300)

	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 50, AccountType: "openai-oauth", LastUsedAt: nil},
		{ID: 2, Priority: 100, AccountType: "openai-responses", LastUsedAt: &now},
		{ID: 3, Priority: 100, AccountType: "openai-oauth", LastUsedAt: &oneHourAgo}, // High priority, low concurrency, older
		{ID: 4, Priority: 75, AccountType: "openai-responses", LastUsedAt: nil},
		{ID: 5, Priority: 100, AccountType: "openai-oauth", LastUsedAt: nil}, // High priority, high concurrency
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	selected, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)
	// Should select ID 5: Priority 100, Concurrency 2, LastUsedAt nil (never used)
	// Among priority 100: ID 2 has concurrency 0, ID 3 has concurrency 1, ID 5 has concurrency 2
	// ID 2 should win (highest priority, lowest concurrency)
	assert.Equal(t, int64(2), selected.ID)
}

func TestPriorityStrategy_Select_EmptyCandidates(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy := NewPriorityStrategy(concurrencyRepo)
	ctx := context.Background()

	candidates := []*model.CodexAccount{}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	_, err := strategy.Select(ctx, candidates, selectionCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no available accounts")
}

func TestPriorityStrategy_Select_SingleCandidate(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy := NewPriorityStrategy(concurrencyRepo)
	ctx := context.Background()

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth"},
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	selected, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), selected.ID)
}

func TestRoundRobinStrategy_Select(t *testing.T) {
	strategy := NewRoundRobinStrategy()
	ctx := context.Background()

	candidates := []*model.CodexAccount{
		{ID: 1, AccountType: "openai-oauth"},
		{ID: 2, AccountType: "openai-responses"},
		{ID: 3, AccountType: "openai-oauth"},
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	// Should cycle through accounts
	selected1, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), selected1.ID)

	selected2, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), selected2.ID)

	selected3, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), selected3.ID)

	// Should wrap around
	selected4, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), selected4.ID)
}

func TestNewStrategy_Priority(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy, err := NewStrategy(StrategyTypePriority, concurrencyRepo)
	require.NoError(t, err)
	assert.NotNil(t, strategy)
	assert.IsType(t, &PriorityStrategy{}, strategy)
}

func TestNewStrategy_RoundRobin(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy, err := NewStrategy(StrategyTypeRoundRobin, concurrencyRepo)
	require.NoError(t, err)
	assert.NotNil(t, strategy)
	assert.IsType(t, &RoundRobinStrategy{}, strategy)
}

func TestNewStrategy_Weighted(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy, err := NewStrategy(StrategyTypeWeighted, concurrencyRepo)
	require.NoError(t, err)
	assert.NotNil(t, strategy)
	assert.IsType(t, &WeightedRoundRobinStrategy{}, strategy)
}

func TestNewStrategy_HealthAware(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy, err := NewStrategy(StrategyTypeHealthAware, concurrencyRepo)
	require.NoError(t, err)
	assert.NotNil(t, strategy)
	assert.IsType(t, &HealthAwareStrategy{}, strategy)
}

func TestNewStrategy_ConsistentHash(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy, err := NewStrategy(StrategyTypeConsistentHash, concurrencyRepo)
	require.NoError(t, err)
	assert.NotNil(t, strategy)
	assert.IsType(t, &ConsistentHashStrategy{}, strategy)
}

func TestNewStrategy_Unknown(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy, err := NewStrategy("unknown", concurrencyRepo)
	assert.Error(t, err)
	assert.Nil(t, strategy)
	assert.Contains(t, err.Error(), "unknown strategy type")
}
