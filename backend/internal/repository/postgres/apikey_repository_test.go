package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/pkg/crypto"
)

func TestAPIKeyRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	apiKey := &model.APIKey{
		KeyHash:               crypto.HashAPIKey("test-key-123"),
		KeyPrefix:             "cr_1234567",
		Name:                  "Test API Key",
		IsActive:              true,
		MaxConcurrentRequests: 5,
		RateLimitPerMinute:    60,
		RateLimitPerHour:      3600,
		RateLimitPerDay:       86400,
	}

	err := repo.Create(ctx, apiKey)
	require.NoError(t, err)
	assert.NotZero(t, apiKey.ID)
}

func TestAPIKeyRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	apiKey := &model.APIKey{
		KeyHash:   crypto.HashAPIKey("test-key-123"),
		KeyPrefix: "cr_1234567",
		Name:      "Test API Key",
		IsActive:  true,
	}
	require.NoError(t, repo.Create(ctx, apiKey))

	found, err := repo.GetByID(ctx, apiKey.ID)
	require.NoError(t, err)
	assert.Equal(t, apiKey.KeyHash, found.KeyHash)
	assert.Equal(t, apiKey.Name, found.Name)
}

func TestAPIKeyRepository_GetByHash(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	keyHash := crypto.HashAPIKey("test-key-123")
	apiKey := &model.APIKey{
		KeyHash:   keyHash,
		KeyPrefix: "cr_1234567",
		Name:      "Test API Key",
		IsActive:  true,
	}
	require.NoError(t, repo.Create(ctx, apiKey))

	found, err := repo.GetByHash(ctx, keyHash)
	require.NoError(t, err)
	assert.Equal(t, apiKey.ID, found.ID)
	assert.Equal(t, apiKey.Name, found.Name)
}

func TestAPIKeyRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	// Create test API keys
	for i := 1; i <= 5; i++ {
		apiKey := &model.APIKey{
			KeyHash:   crypto.HashAPIKey("key-" + string(rune('0'+i))),
			KeyPrefix: "cr_123456" + string(rune('0'+i)),
			Name:      "API Key " + string(rune('0'+i)),
			IsActive:  true,
		}
		require.NoError(t, repo.Create(ctx, apiKey))
	}

	// Test pagination
	keys, err := repo.List(ctx, 0, 3)
	require.NoError(t, err)
	assert.Len(t, keys, 3)
}

func TestAPIKeyRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	apiKey := &model.APIKey{
		KeyHash:   crypto.HashAPIKey("test-key-123"),
		KeyPrefix: "cr_1234567",
		Name:      "Test API Key",
		IsActive:  true,
	}
	require.NoError(t, repo.Create(ctx, apiKey))

	updates := map[string]interface{}{
		"name":      "Updated API Key",
		"is_active": false,
	}
	err := repo.Update(ctx, apiKey.ID, updates)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, apiKey.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated API Key", updated.Name)
	assert.False(t, updated.IsActive)
}

func TestAPIKeyRepository_UpdateStats(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	apiKey := &model.APIKey{
		KeyHash:       crypto.HashAPIKey("test-key-123"),
		KeyPrefix:     "cr_1234567",
		Name:          "Test API Key",
		IsActive:      true,
		TotalRequests: 10,
		TotalTokens:   100,
		TotalCost:     1.5,
	}
	require.NoError(t, repo.Create(ctx, apiKey))

	// Update stats atomically
	err := repo.UpdateStats(ctx, apiKey.ID, 5, 50, 0.75)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, apiKey.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(15), updated.TotalRequests)
	assert.Equal(t, int64(150), updated.TotalTokens)
	assert.InDelta(t, 2.25, updated.TotalCost, 0.01)
}

func TestAPIKeyRepository_UpdateStats_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	err := repo.UpdateStats(ctx, 999, 1, 10, 0.1)
	require.Error(t, err)
}

func TestAPIKeyRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	apiKey := &model.APIKey{
		KeyHash:   crypto.HashAPIKey("test-key-123"),
		KeyPrefix: "cr_1234567",
		Name:      "Test API Key",
		IsActive:  true,
	}
	require.NoError(t, repo.Create(ctx, apiKey))

	err := repo.Delete(ctx, apiKey.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, apiKey.ID)
	require.Error(t, err)
}

func TestAPIKeyRepository_ExpiresAt(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	apiKey := &model.APIKey{
		KeyHash:   crypto.HashAPIKey("test-key-123"),
		KeyPrefix: "cr_1234567",
		Name:      "Test API Key",
		IsActive:  true,
		ExpiresAt: &expiresAt,
	}
	require.NoError(t, repo.Create(ctx, apiKey))

	found, err := repo.GetByID(ctx, apiKey.ID)
	require.NoError(t, err)
	assert.NotNil(t, found.ExpiresAt)
	assert.WithinDuration(t, expiresAt, *found.ExpiresAt, time.Second)
}

func TestAPIKeyRepository_GetByHash_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	_, err := repo.GetByHash(ctx, "nonexistent-hash")
	require.Error(t, err)
}

func TestAPIKeyRepository_List_EmptyResult(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	keys, err := repo.List(ctx, 0, 10)
	require.NoError(t, err)
	assert.Empty(t, keys)
}
