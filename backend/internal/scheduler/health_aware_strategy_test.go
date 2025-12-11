package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

func TestHealthAwareStrategy_Select_ByHealthScore(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy := NewHealthAwareStrategy(concurrencyRepo)
	ctx := context.Background()

	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth", LastUsedAt: &now, RateLimitedUntil: nil},
		{ID: 2, Priority: 100, AccountType: "openai-responses", LastUsedAt: &oneHourAgo, RateLimitedUntil: nil},
		{ID: 3, Priority: 50, AccountType: "openai-oauth", LastUsedAt: nil, RateLimitedUntil: nil},
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	selected, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)
	assert.NotNil(t, selected)
}

func TestHealthAwareStrategy_Select_AvoidRateLimited(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy := NewHealthAwareStrategy(concurrencyRepo)
	ctx := context.Background()

	futureTime := time.Now().Add(1 * time.Hour)

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth", RateLimitedUntil: &futureTime},
		{ID: 2, Priority: 50, AccountType: "openai-responses", RateLimitedUntil: nil},
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	selected, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), selected.ID)
}

func TestHealthAwareStrategy_Select_PreferLowConcurrency(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy := NewHealthAwareStrategy(concurrencyRepo)
	ctx := context.Background()

	_, _ = concurrencyRepo.Acquire(ctx, 1, "req-1", 300)
	_, _ = concurrencyRepo.Acquire(ctx, 1, "req-2", 300)

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth"},
		{ID: 2, Priority: 100, AccountType: "openai-responses"},
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	selected, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), selected.ID)
}

func TestHealthAwareStrategy_Select_EmptyCandidates(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy := NewHealthAwareStrategy(concurrencyRepo)
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

func TestHealthAwareStrategy_Select_SingleCandidate(t *testing.T) {
	concurrencyRepo, _ := setupTestConcurrencyRepo(t)
	strategy := NewHealthAwareStrategy(concurrencyRepo)
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
