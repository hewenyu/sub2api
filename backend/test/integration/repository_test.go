package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository/postgres"
)

func TestAPIKeyRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := postgres.NewAPIKeyRepository(db)
	ctx := context.Background()

	t.Run("Create and retrieve API key", func(t *testing.T) {
		apiKey, _ := createTestAPIKey(t, db, "Test API Key")

		// Retrieve by ID
		retrieved, err := repo.GetByID(ctx, apiKey.ID)
		require.NoError(t, err)
		assert.Equal(t, apiKey.Name, retrieved.Name)
		assert.Equal(t, apiKey.KeyPrefix, retrieved.KeyPrefix)
	})

	t.Run("List API keys", func(t *testing.T) {
		// Create multiple API keys
		for i := 1; i <= 3; i++ {
			createTestAPIKey(t, db, fmt.Sprintf("API Key %d", i))
		}

		// List all keys
		keys, err := repo.List(ctx, 0, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(keys), 3)
	})

	t.Run("Update API key", func(t *testing.T) {
		apiKey, _ := createTestAPIKey(t, db, "Original Name")

		// Update
		updates := map[string]any{
			"name":             "Updated Name",
			"daily_cost_limit": 100.0,
		}
		err := repo.Update(ctx, apiKey.ID, updates)
		require.NoError(t, err)

		// Verify update
		retrieved, err := repo.GetByID(ctx, apiKey.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.Name)
		assert.Equal(t, 100.0, retrieved.DailyCostLimit)
	})

	t.Run("Delete API key", func(t *testing.T) {
		apiKey, _ := createTestAPIKey(t, db, "To Delete")

		// Delete
		err := repo.Delete(ctx, apiKey.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(ctx, apiKey.ID)
		assert.Error(t, err)
	})
}

func TestCodexAccountRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := postgres.NewCodexAccountRepository(db)
	ctx := context.Background()

	t.Run("Create and retrieve Codex account", func(t *testing.T) {
		account := createTestCodexAccount(t, db, "Test Account", "openai-responses")

		// Retrieve by ID
		retrieved, err := repo.GetByID(ctx, account.ID)
		require.NoError(t, err)
		assert.Equal(t, account.Name, retrieved.Name)
		assert.Equal(t, account.AccountType, retrieved.AccountType)
	})

	t.Run("List schedulable accounts", func(t *testing.T) {
		// Create multiple accounts
		createTestCodexAccount(t, db, "Account 1", "openai-responses")
		createTestCodexAccount(t, db, "Account 2", "openai-responses")

		// List schedulable accounts
		accounts, err := repo.GetSchedulable(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(accounts), 2)

		// Verify all are active and schedulable
		for _, acc := range accounts {
			assert.True(t, acc.IsActive)
			assert.True(t, acc.Schedulable)
		}
	})

	t.Run("Update account fields", func(t *testing.T) {
		account := createTestCodexAccount(t, db, "To Update", "openai-responses")

		// Update fields using UpdateFields method
		futureTime := time.Now().Add(1 * time.Hour)
		updates := map[string]any{
			"is_active":          false,
			"rate_limited_until": futureTime,
		}
		err := repo.UpdateFields(ctx, account.ID, updates)
		require.NoError(t, err)

		// Verify
		retrieved, err := repo.GetByID(ctx, account.ID)
		require.NoError(t, err)
		assert.False(t, retrieved.IsActive)
		assert.NotNil(t, retrieved.RateLimitedUntil)
	})
}

func TestUsageRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := postgres.NewUsageRepository(db)
	ctx := context.Background()

	t.Run("Create and retrieve usage record", func(t *testing.T) {
		apiKey, _ := createTestAPIKey(t, db, "Test API Key")

		usage := &model.Usage{
			APIKeyID:         apiKey.ID,
			Type:             model.UsageTypeCodex,
			AccountID:        1,
			Model:            "gpt-4",
			InputTokens:      1000,
			OutputTokens:     500,
			TotalTokens:      1500,
			Cost:             0.06,
			StatusCode:       200,
			RequestMetadata:  "{}",
			ResponseMetadata: "{}",
			CreatedAt:        time.Now(),
		}

		err := repo.Create(ctx, usage)
		require.NoError(t, err)
		assert.NotZero(t, usage.ID)

		// Retrieve
		retrieved, err := repo.GetByID(ctx, usage.ID)
		require.NoError(t, err)
		assert.Equal(t, usage.APIKeyID, retrieved.APIKeyID)
		assert.Equal(t, usage.Model, retrieved.Model)
		assert.Equal(t, usage.TotalTokens, retrieved.TotalTokens)
	})

	t.Run("Aggregate usage statistics", func(t *testing.T) {
		apiKey, _ := createTestAPIKey(t, db, "Aggregate Test Key")

		// Create multiple usage records
		for i := 0; i < 5; i++ {
			usage := &model.Usage{
				APIKeyID:         apiKey.ID,
				Type:             model.UsageTypeCodex,
				AccountID:        1,
				Model:            "gpt-4",
				InputTokens:      100,
				OutputTokens:     50,
				TotalTokens:      150,
				Cost:             0.01,
				StatusCode:       200,
				RequestMetadata:  "{}",
				ResponseMetadata: "{}",
				CreatedAt:        time.Now(),
			}
			err := repo.Create(ctx, usage)
			require.NoError(t, err)
		}

		// Aggregate
		usageType := model.UsageTypeCodex
		filters := repository.UsageFilters{
			APIKeyID:  &apiKey.ID,
			UsageType: &usageType,
		}

		aggregate, err := repo.Aggregate(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, int64(5), aggregate.TotalRequests)
		assert.Equal(t, int64(750), aggregate.TotalTokens)
		assert.InDelta(t, 0.05, aggregate.TotalCost, 0.001)
	})

	t.Run("List usage with pagination", func(t *testing.T) {
		apiKey, _ := createTestAPIKey(t, db, "Pagination Test Key")

		// Create 10 usage records
		for i := 0; i < 10; i++ {
			usage := &model.Usage{
				APIKeyID:         apiKey.ID,
				Type:             model.UsageTypeCodex,
				AccountID:        1,
				Model:            "gpt-4",
				InputTokens:      100,
				OutputTokens:     50,
				TotalTokens:      150,
				Cost:             0.01,
				StatusCode:       200,
				RequestMetadata:  "{}",
				ResponseMetadata: "{}",
				CreatedAt:        time.Now(),
			}
			err := repo.Create(ctx, usage)
			require.NoError(t, err)
		}

		// List first page
		filters := repository.UsageFilters{
			APIKeyID: &apiKey.ID,
		}
		usages, err := repo.List(ctx, filters, 0, 5)
		require.NoError(t, err)
		assert.Len(t, usages, 5)

		// List second page
		usages, err = repo.List(ctx, filters, 5, 5)
		require.NoError(t, err)
		assert.Len(t, usages, 5)
	})
}
