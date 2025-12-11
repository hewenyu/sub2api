package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
)

func TestUsageRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUsageRepository(db)
	ctx := context.Background()

	usage := &model.Usage{
		APIKeyID:     1,
		Type:         model.UsageTypeClaude,
		AccountID:    1,
		Model:        "claude-3-5-sonnet-20241022",
		InputTokens:  100,
		OutputTokens: 200,
		TotalTokens:  300,
		Cost:         0.15,
		StatusCode:   200,
	}

	err := repo.Create(ctx, usage)
	require.NoError(t, err)
	assert.NotZero(t, usage.ID)
}

func TestUsageRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUsageRepository(db)
	ctx := context.Background()

	usage := &model.Usage{
		APIKeyID:     1,
		Type:         model.UsageTypeClaude,
		AccountID:    1,
		Model:        "claude-3-5-sonnet-20241022",
		InputTokens:  100,
		OutputTokens: 200,
		TotalTokens:  300,
		Cost:         0.15,
		StatusCode:   200,
	}
	require.NoError(t, repo.Create(ctx, usage))

	found, err := repo.GetByID(ctx, usage.ID)
	require.NoError(t, err)
	assert.Equal(t, usage.Model, found.Model)
	assert.Equal(t, usage.TotalTokens, found.TotalTokens)
}

func TestUsageRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUsageRepository(db)
	ctx := context.Background()

	// Create test usage records
	for i := 1; i <= 5; i++ {
		usage := &model.Usage{
			APIKeyID:     int64(i),
			Type:         model.UsageTypeClaude,
			AccountID:    1,
			Model:        "claude-3-5-sonnet-20241022",
			InputTokens:  100,
			OutputTokens: 200,
			TotalTokens:  300,
			Cost:         0.15,
			StatusCode:   200,
		}
		require.NoError(t, repo.Create(ctx, usage))
	}

	// Test list all
	usages, err := repo.List(ctx, repository.UsageFilters{}, 0, 10)
	require.NoError(t, err)
	assert.Len(t, usages, 5)

	// Test filter by API key
	apiKeyID := int64(1)
	usages, err = repo.List(ctx, repository.UsageFilters{APIKeyID: &apiKeyID}, 0, 10)
	require.NoError(t, err)
	assert.Len(t, usages, 1)

	// Test pagination
	usages, err = repo.List(ctx, repository.UsageFilters{}, 0, 3)
	require.NoError(t, err)
	assert.Len(t, usages, 3)
}

func TestUsageRepository_List_FilterByType(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUsageRepository(db)
	ctx := context.Background()

	// Create Claude usage
	claude := &model.Usage{
		APIKeyID:     1,
		Type:         model.UsageTypeClaude,
		AccountID:    1,
		Model:        "claude-3-5-sonnet-20241022",
		InputTokens:  100,
		OutputTokens: 200,
		TotalTokens:  300,
		Cost:         0.15,
		StatusCode:   200,
	}
	require.NoError(t, repo.Create(ctx, claude))

	// Create Codex usage
	codex := &model.Usage{
		APIKeyID:     1,
		Type:         model.UsageTypeCodex,
		AccountID:    2,
		Model:        "codex-model",
		InputTokens:  50,
		OutputTokens: 100,
		TotalTokens:  150,
		Cost:         0.08,
		StatusCode:   200,
	}
	require.NoError(t, repo.Create(ctx, codex))

	// Filter by Claude type
	apiKeyID := int64(1)
	claudeType := model.UsageTypeClaude
	usages, err := repo.List(ctx, repository.UsageFilters{APIKeyID: &apiKeyID, UsageType: &claudeType}, 0, 10)
	require.NoError(t, err)
	assert.Len(t, usages, 1)
	assert.Equal(t, model.UsageTypeClaude, usages[0].Type)

	// Filter by Codex type
	codexType := model.UsageTypeCodex
	usages, err = repo.List(ctx, repository.UsageFilters{APIKeyID: &apiKeyID, UsageType: &codexType}, 0, 10)
	require.NoError(t, err)
	assert.Len(t, usages, 1)
	assert.Equal(t, model.UsageTypeCodex, usages[0].Type)
}

func TestUsageRepository_Aggregate(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUsageRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()

	// Create usage records
	for i := 1; i <= 5; i++ {
		usage := &model.Usage{
			APIKeyID:     1,
			Type:         model.UsageTypeClaude,
			AccountID:    1,
			Model:        "claude-3-5-sonnet-20241022",
			InputTokens:  100,
			OutputTokens: 200,
			TotalTokens:  300,
			Cost:         0.15,
			StatusCode:   200,
		}
		require.NoError(t, repo.Create(ctx, usage))
	}

	// Aggregate all
	apiKeyID := int64(1)
	usageType := model.UsageTypeClaude
	aggregate, err := repo.Aggregate(ctx, repository.UsageFilters{APIKeyID: &apiKeyID, UsageType: &usageType})
	require.NoError(t, err)
	assert.Equal(t, int64(5), aggregate.TotalRequests)
	assert.Equal(t, int64(1500), aggregate.TotalTokens)
	assert.InDelta(t, 0.75, aggregate.TotalCost, 0.01)

	// Aggregate with time range
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(1 * time.Hour)
	aggregate, err = repo.Aggregate(ctx, repository.UsageFilters{APIKeyID: &apiKeyID, UsageType: &usageType, StartDate: &startTime, EndDate: &endTime})
	require.NoError(t, err)
	assert.Equal(t, int64(5), aggregate.TotalRequests)
}

func TestUsageRepository_Aggregate_EmptyResult(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUsageRepository(db)
	ctx := context.Background()

	apiKeyID := int64(999)
	usageType := model.UsageTypeClaude
	aggregate, err := repo.Aggregate(ctx, repository.UsageFilters{APIKeyID: &apiKeyID, UsageType: &usageType})
	require.NoError(t, err)
	assert.Equal(t, int64(0), aggregate.TotalRequests)
	assert.Equal(t, int64(0), aggregate.TotalTokens)
	assert.Equal(t, 0.0, aggregate.TotalCost)
}

func TestUsageRepository_Aggregate_MultipleAPIKeys(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUsageRepository(db)
	ctx := context.Background()

	// Create usage for API key 1
	for i := 1; i <= 3; i++ {
		usage := &model.Usage{
			APIKeyID:     1,
			Type:         model.UsageTypeClaude,
			AccountID:    1,
			Model:        "claude-3-5-sonnet-20241022",
			InputTokens:  100,
			OutputTokens: 200,
			TotalTokens:  300,
			Cost:         0.15,
			StatusCode:   200,
		}
		require.NoError(t, repo.Create(ctx, usage))
	}

	// Create usage for API key 2
	for i := 1; i <= 2; i++ {
		usage := &model.Usage{
			APIKeyID:     2,
			Type:         model.UsageTypeClaude,
			AccountID:    1,
			Model:        "claude-3-5-sonnet-20241022",
			InputTokens:  100,
			OutputTokens: 200,
			TotalTokens:  300,
			Cost:         0.15,
			StatusCode:   200,
		}
		require.NoError(t, repo.Create(ctx, usage))
	}

	// Aggregate for API key 1
	apiKeyID1 := int64(1)
	usageType := model.UsageTypeClaude
	aggregate, err := repo.Aggregate(ctx, repository.UsageFilters{APIKeyID: &apiKeyID1, UsageType: &usageType})
	require.NoError(t, err)
	assert.Equal(t, int64(3), aggregate.TotalRequests)

	// Aggregate for API key 2
	apiKeyID2 := int64(2)
	aggregate, err = repo.Aggregate(ctx, repository.UsageFilters{APIKeyID: &apiKeyID2, UsageType: &usageType})
	require.NoError(t, err)
	assert.Equal(t, int64(2), aggregate.TotalRequests)
}

func TestUsageRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUsageRepository(db)
	ctx := context.Background()

	usage := &model.Usage{
		APIKeyID:     1,
		Type:         model.UsageTypeClaude,
		AccountID:    1,
		Model:        "claude-3-5-sonnet-20241022",
		InputTokens:  100,
		OutputTokens: 200,
		TotalTokens:  300,
		Cost:         0.15,
		StatusCode:   200,
	}
	require.NoError(t, repo.Create(ctx, usage))

	err := repo.Delete(ctx, usage.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, usage.ID)
	require.Error(t, err)
}
