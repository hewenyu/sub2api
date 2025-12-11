package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

func TestAdminRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAdminRepository(db)
	ctx := context.Background()

	admin := &model.Admin{
		Username:     "testadmin",
		PasswordHash: "hashed_password",
		Email:        "admin@test.com",
		IsActive:     true,
	}

	err := repo.Create(ctx, admin)
	require.NoError(t, err)
	assert.NotZero(t, admin.ID)
}

func TestAdminRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAdminRepository(db)
	ctx := context.Background()

	admin := &model.Admin{
		Username:     "testadmin",
		PasswordHash: "hashed_password",
		Email:        "admin@test.com",
		IsActive:     true,
	}
	require.NoError(t, repo.Create(ctx, admin))

	found, err := repo.GetByID(ctx, admin.ID)
	require.NoError(t, err)
	assert.Equal(t, admin.Username, found.Username)
	assert.Equal(t, admin.Email, found.Email)
}

func TestAdminRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAdminRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 999)
	require.Error(t, err)
}

func TestAdminRepository_GetByUsername(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAdminRepository(db)
	ctx := context.Background()

	admin := &model.Admin{
		Username:     "testadmin",
		PasswordHash: "hashed_password",
		Email:        "admin@test.com",
		IsActive:     true,
	}
	require.NoError(t, repo.Create(ctx, admin))

	found, err := repo.GetByUsername(ctx, "testadmin")
	require.NoError(t, err)
	assert.Equal(t, admin.ID, found.ID)
	assert.Equal(t, admin.Email, found.Email)
}

func TestAdminRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAdminRepository(db)
	ctx := context.Background()

	// Create test admins
	for i := 1; i <= 5; i++ {
		admin := &model.Admin{
			Username:     "admin" + string(rune('0'+i)),
			PasswordHash: "hash",
			Email:        "admin" + string(rune('0'+i)) + "@test.com",
			IsActive:     true,
		}
		require.NoError(t, repo.Create(ctx, admin))
	}

	// Test pagination
	admins, err := repo.List(ctx, 0, 3)
	require.NoError(t, err)
	assert.Len(t, admins, 3)

	// Test offset
	admins, err = repo.List(ctx, 2, 3)
	require.NoError(t, err)
	assert.Len(t, admins, 3)
}

func TestAdminRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAdminRepository(db)
	ctx := context.Background()

	admin := &model.Admin{
		Username:     "testadmin",
		PasswordHash: "hashed_password",
		Email:        "admin@test.com",
		IsActive:     true,
	}
	require.NoError(t, repo.Create(ctx, admin))

	updates := map[string]interface{}{
		"email":     "updated@test.com",
		"is_active": false,
	}
	err := repo.Update(ctx, admin.ID, updates)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, admin.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated@test.com", updated.Email)
	assert.False(t, updated.IsActive)
}

func TestAdminRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAdminRepository(db)
	ctx := context.Background()

	admin := &model.Admin{
		Username:     "testadmin",
		PasswordHash: "hashed_password",
		Email:        "admin@test.com",
		IsActive:     true,
	}
	require.NoError(t, repo.Create(ctx, admin))

	err := repo.Delete(ctx, admin.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, admin.ID)
	require.Error(t, err)
}

func TestAdminRepository_GetByUsername_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAdminRepository(db)
	ctx := context.Background()

	_, err := repo.GetByUsername(ctx, "nonexistent")
	require.Error(t, err)
}

func TestAdminRepository_List_EmptyResult(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewAdminRepository(db)
	ctx := context.Background()

	admins, err := repo.List(ctx, 0, 10)
	require.NoError(t, err)
	assert.Empty(t, admins)
}
