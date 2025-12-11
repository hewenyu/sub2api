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

func TestCodexAccountRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCodexAccountRepository(db)
	ctx := context.Background()

	apiKey := "test-api-key"
	account := &model.CodexAccount{
		APIKey:      &apiKey,
		Name:        "Test Codex Account",
		AccountType: "openai-responses",
		IsActive:    true,
		Schedulable: true,
	}

	err := repo.Create(ctx, account)
	require.NoError(t, err)
	assert.NotZero(t, account.ID)
}

func TestCodexAccountRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCodexAccountRepository(db)
	ctx := context.Background()

	apiKey := "test-api-key"
	account := &model.CodexAccount{
		APIKey:      &apiKey,
		Name:        "Test Codex Account",
		AccountType: "openai-responses",
		IsActive:    true,
		Schedulable: true,
	}
	require.NoError(t, repo.Create(ctx, account))

	found, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	assert.Equal(t, account.Name, found.Name)
}

func TestCodexAccountRepository_GetByAPIKey(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCodexAccountRepository(db)
	ctx := context.Background()

	apiKey := "test-api-key"
	account := &model.CodexAccount{
		APIKey:      &apiKey,
		Name:        "Test Codex Account",
		AccountType: "openai-responses",
		IsActive:    true,
		Schedulable: true,
	}
	require.NoError(t, repo.Create(ctx, account))

	found, err := repo.GetByAPIKey(ctx, apiKey)
	require.NoError(t, err)
	assert.Equal(t, account.ID, found.ID)
}

func TestCodexAccountRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCodexAccountRepository(db)
	ctx := context.Background()

	// Create test accounts
	for i := 1; i <= 5; i++ {
		apiKey := "key-" + string(rune('0'+i))
		account := &model.CodexAccount{
			APIKey:      &apiKey,
			Name:        "Account " + string(rune('0'+i)),
			AccountType: "openai-responses",
			IsActive:    true,
			Schedulable: true,
		}
		require.NoError(t, repo.Create(ctx, account))
	}

	// Test pagination
	accounts, total, err := repo.List(ctx, repository.CodexAccountFilters{}, 0, 3)
	require.NoError(t, err)
	assert.Len(t, accounts, 3)
	assert.Equal(t, int64(5), total)
}

func TestCodexAccountRepository_GetSchedulable(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCodexAccountRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()

	// Create schedulable account
	apiKey1 := "schedulable-key"
	schedulable := &model.CodexAccount{
		APIKey:      &apiKey1,
		Name:        "Schedulable Account",
		AccountType: "openai-responses",
		IsActive:    true,
		Schedulable: true,
	}
	require.NoError(t, repo.Create(ctx, schedulable))

	// Create inactive account
	apiKey2 := "inactive-key"
	inactive := &model.CodexAccount{
		APIKey:      &apiKey2,
		Name:        "Inactive Account",
		AccountType: "openai-responses",
		IsActive:    true,
		Schedulable: true,
	}
	require.NoError(t, repo.Create(ctx, inactive))
	// Update to set IsActive to false (SQLite workaround)
	inactive.IsActive = false
	require.NoError(t, repo.Update(ctx, inactive))

	// Create rate-limited account
	rateLimitedUntil := now.Add(1 * time.Hour)
	apiKey3 := "ratelimited-key"
	rateLimited := &model.CodexAccount{
		APIKey:           &apiKey3,
		Name:             "Rate Limited Account",
		AccountType:      "openai-responses",
		IsActive:         true,
		Schedulable:      false,
		RateLimitedUntil: &rateLimitedUntil,
	}
	require.NoError(t, repo.Create(ctx, rateLimited))

	// Create overloaded account
	overloadUntil := now.Add(1 * time.Hour)
	apiKey4 := "overloaded-key"
	overloaded := &model.CodexAccount{
		APIKey:        &apiKey4,
		Name:          "Overloaded Account",
		AccountType:   "openai-responses",
		IsActive:      true,
		Schedulable:   false,
		OverloadUntil: &overloadUntil,
	}
	require.NoError(t, repo.Create(ctx, overloaded))

	// Get schedulable accounts
	accounts, err := repo.GetSchedulable(ctx)
	require.NoError(t, err)
	assert.Len(t, accounts, 1)
	assert.Equal(t, "Schedulable Account", accounts[0].Name)
}

func TestCodexAccountRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCodexAccountRepository(db)
	ctx := context.Background()

	apiKey := "test-api-key"
	account := &model.CodexAccount{
		APIKey:      &apiKey,
		Name:        "Test Codex Account",
		AccountType: "openai-responses",
		IsActive:    true,
		Schedulable: true,
	}
	require.NoError(t, repo.Create(ctx, account))

	account.Name = "Updated Account"
	account.IsActive = false
	err := repo.Update(ctx, account)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Account", updated.Name)
	assert.False(t, updated.IsActive)
}

func TestCodexAccountRepository_UpdateConcurrentRequests(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCodexAccountRepository(db)
	ctx := context.Background()

	apiKey := "test-api-key"
	account := &model.CodexAccount{
		APIKey:             &apiKey,
		Name:               "Test Codex Account",
		AccountType:        "openai-responses",
		IsActive:           true,
		Schedulable:        true,
		ConcurrentRequests: 0,
	}
	require.NoError(t, repo.Create(ctx, account))

	// Increment concurrent requests
	err := repo.UpdateConcurrentRequests(ctx, account.ID, 2)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, updated.ConcurrentRequests)

	// Decrement concurrent requests
	err = repo.UpdateConcurrentRequests(ctx, account.ID, -1)
	require.NoError(t, err)

	updated, err = repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, updated.ConcurrentRequests)
}

func TestCodexAccountRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCodexAccountRepository(db)
	ctx := context.Background()

	apiKey := "test-api-key"
	account := &model.CodexAccount{
		APIKey:      &apiKey,
		Name:        "Test Codex Account",
		AccountType: "openai-responses",
		IsActive:    true,
		Schedulable: true,
	}
	require.NoError(t, repo.Create(ctx, account))

	err := repo.Delete(ctx, account.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, account.ID)
	require.Error(t, err)
}

func TestCodexAccountRepository_GetByAPIKey_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCodexAccountRepository(db)
	ctx := context.Background()

	_, err := repo.GetByAPIKey(ctx, "nonexistent-key")
	require.Error(t, err)
}

func TestCodexAccountRepository_UpdateConcurrentRequests_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCodexAccountRepository(db)
	ctx := context.Background()

	err := repo.UpdateConcurrentRequests(ctx, 999, 1)
	require.Error(t, err)
}

func TestCodexAccountRepository_List_EmptyResult(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCodexAccountRepository(db)
	ctx := context.Background()

	accounts, total, err := repo.List(ctx, repository.CodexAccountFilters{}, 0, 10)
	require.NoError(t, err)
	assert.Empty(t, accounts)
	assert.Equal(t, int64(0), total)
}
