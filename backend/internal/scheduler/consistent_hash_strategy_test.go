package scheduler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

func TestConsistentHashStrategy_Select_SessionAffinity(t *testing.T) {
	strategy := NewConsistentHashStrategy()
	ctx := context.Background()

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth"},
		{ID: 2, Priority: 100, AccountType: "openai-responses"},
		{ID: 3, Priority: 100, AccountType: "openai-oauth"},
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "session-123",
	}

	selected1, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)

	for range 10 {
		selected, err := strategy.Select(ctx, candidates, selectionCtx)
		require.NoError(t, err)
		assert.Equal(t, selected1.ID, selected.ID)
	}
}

func TestConsistentHashStrategy_Select_DifferentSessions(t *testing.T) {
	strategy := NewConsistentHashStrategy()
	ctx := context.Background()

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth"},
		{ID: 2, Priority: 100, AccountType: "openai-responses"},
		{ID: 3, Priority: 100, AccountType: "openai-oauth"},
	}

	selectionCtx1 := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "session-1",
	}

	selectionCtx2 := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "session-2",
	}

	selected1, err := strategy.Select(ctx, candidates, selectionCtx1)
	require.NoError(t, err)

	selected2, err := strategy.Select(ctx, candidates, selectionCtx2)
	require.NoError(t, err)

	assert.NotNil(t, selected1)
	assert.NotNil(t, selected2)
}

func TestConsistentHashStrategy_Select_EmptySessionHash(t *testing.T) {
	strategy := NewConsistentHashStrategy()
	ctx := context.Background()

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
	assert.NotNil(t, selected)
}

func TestConsistentHashStrategy_Select_SingleCandidate(t *testing.T) {
	strategy := NewConsistentHashStrategy()
	ctx := context.Background()

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth"},
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "session-123",
	}

	selected, err := strategy.Select(ctx, candidates, selectionCtx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), selected.ID)
}

func TestConsistentHashStrategy_Select_EmptyCandidates(t *testing.T) {
	strategy := NewConsistentHashStrategy()
	ctx := context.Background()

	candidates := []*model.CodexAccount{}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "session-123",
	}

	_, err := strategy.Select(ctx, candidates, selectionCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no available accounts")
}

func TestConsistentHashStrategy_Select_Distribution(t *testing.T) {
	strategy := NewConsistentHashStrategy()
	ctx := context.Background()

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth"},
		{ID: 2, Priority: 100, AccountType: "openai-responses"},
		{ID: 3, Priority: 100, AccountType: "openai-oauth"},
	}

	selections := make(map[int64]int)

	for i := range 100 {
		selectionCtx := SelectionContext{
			APIKey:      &model.APIKey{ID: 1},
			SessionHash: string(rune(i)),
		}

		selected, err := strategy.Select(ctx, candidates, selectionCtx)
		require.NoError(t, err)
		selections[selected.ID]++
	}

	assert.True(t, selections[1] > 0)
	assert.True(t, selections[2] > 0)
	assert.True(t, selections[3] > 0)
}
