package scheduler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

func TestWeightedRoundRobinStrategy_Select_DistributionByWeight(t *testing.T) {
	strategy := NewWeightedRoundRobinStrategy()
	ctx := context.Background()

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth"},
		{ID: 2, Priority: 50, AccountType: "openai-responses"},
		{ID: 3, Priority: 25, AccountType: "openai-oauth"},
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	selections := make(map[int64]int)
	totalSelections := 175

	for range totalSelections {
		selected, err := strategy.Select(ctx, candidates, selectionCtx)
		require.NoError(t, err)
		selections[selected.ID]++
	}

	assert.Equal(t, 100, selections[1])
	assert.Equal(t, 50, selections[2])
	assert.Equal(t, 25, selections[3])
}

func TestWeightedRoundRobinStrategy_Select_EqualWeights(t *testing.T) {
	strategy := NewWeightedRoundRobinStrategy()
	ctx := context.Background()

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth"},
		{ID: 2, Priority: 100, AccountType: "openai-responses"},
		{ID: 3, Priority: 100, AccountType: "openai-oauth"},
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	selections := make(map[int64]int)
	for range 300 {
		selected, err := strategy.Select(ctx, candidates, selectionCtx)
		require.NoError(t, err)
		selections[selected.ID]++
	}

	assert.Equal(t, 100, selections[1])
	assert.Equal(t, 100, selections[2])
	assert.Equal(t, 100, selections[3])
}

func TestWeightedRoundRobinStrategy_Select_SingleCandidate(t *testing.T) {
	strategy := NewWeightedRoundRobinStrategy()
	ctx := context.Background()

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 100, AccountType: "openai-oauth"},
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	for range 10 {
		selected, err := strategy.Select(ctx, candidates, selectionCtx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), selected.ID)
	}
}

func TestWeightedRoundRobinStrategy_Select_EmptyCandidates(t *testing.T) {
	strategy := NewWeightedRoundRobinStrategy()
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

func TestWeightedRoundRobinStrategy_Select_ZeroWeights(t *testing.T) {
	strategy := NewWeightedRoundRobinStrategy()
	ctx := context.Background()

	candidates := []*model.CodexAccount{
		{ID: 1, Priority: 0, AccountType: "openai-oauth"},
		{ID: 2, Priority: 0, AccountType: "openai-responses"},
	}

	selectionCtx := SelectionContext{
		APIKey:      &model.APIKey{ID: 1},
		SessionHash: "",
	}

	selections := make(map[int64]int)
	for range 10 {
		selected, err := strategy.Select(ctx, candidates, selectionCtx)
		require.NoError(t, err)
		selections[selected.ID]++
	}

	assert.Equal(t, 5, selections[1])
	assert.Equal(t, 5, selections[2])
}
