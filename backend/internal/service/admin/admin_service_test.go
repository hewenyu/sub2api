package admin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository/postgres"
	"github.com/Wei-Shaw/sub2api/backend/pkg/crypto"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.Admin{})
	require.NoError(t, err)

	return db
}

func TestAdminService_CreateAdmin(t *testing.T) {
	db := setupTestDB(t)
	adminRepo := postgres.NewAdminRepository(db)
	logger := zap.NewNop()
	service := NewAdminService(adminRepo, "test-secret-key", 24*time.Hour, logger)

	ctx := context.Background()

	admin, err := service.CreateAdmin(ctx, "testadmin", "password123")
	require.NoError(t, err)
	assert.NotNil(t, admin)
	assert.Equal(t, "testadmin", admin.Username)
	assert.NotEmpty(t, admin.PasswordHash)
	assert.True(t, admin.IsActive)

	// Verify password was hashed
	err = crypto.BcryptCompare(admin.PasswordHash, "password123")
	assert.NoError(t, err)

	// Try to create duplicate admin
	_, err = service.CreateAdmin(ctx, "testadmin", "password456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestAdminService_Login(t *testing.T) {
	db := setupTestDB(t)
	adminRepo := postgres.NewAdminRepository(db)
	logger := zap.NewNop()
	service := NewAdminService(adminRepo, "test-secret-key-minimum-32chars!", 24*time.Hour, logger)

	ctx := context.Background()

	// Create test admin
	_, err := service.CreateAdmin(ctx, "testadmin", "password123")
	require.NoError(t, err)

	t.Run("Successful login", func(t *testing.T) {
		token, admin, err := service.Login(ctx, "testadmin", "password123")
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotNil(t, admin)
		assert.Equal(t, "testadmin", admin.Username)
	})

	t.Run("Invalid username", func(t *testing.T) {
		_, _, err := service.Login(ctx, "nonexistent", "password123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("Invalid password", func(t *testing.T) {
		_, _, err := service.Login(ctx, "testadmin", "wrongpassword")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
	})
}

func TestAdminService_ValidateToken(t *testing.T) {
	db := setupTestDB(t)
	adminRepo := postgres.NewAdminRepository(db)
	logger := zap.NewNop()
	jwtSecret := "test-secret-key-minimum-32chars!"
	service := NewAdminService(adminRepo, jwtSecret, 24*time.Hour, logger)

	ctx := context.Background()

	// Create and login admin
	admin, err := service.CreateAdmin(ctx, "testadmin", "password123")
	require.NoError(t, err)

	token, _, err := service.Login(ctx, "testadmin", "password123")
	require.NoError(t, err)

	t.Run("Valid token", func(t *testing.T) {
		adminID, err := service.ValidateToken(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, admin.ID, adminID)
	})

	t.Run("Invalid token", func(t *testing.T) {
		_, err := service.ValidateToken(ctx, "invalid.token.here")
		assert.Error(t, err)
	})

	t.Run("Expired token", func(t *testing.T) {
		// Create service with very short expiration
		shortService := NewAdminService(adminRepo, jwtSecret, 1*time.Millisecond, logger)
		shortToken, _, err := shortService.Login(ctx, "testadmin", "password123")
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		_, err = shortService.ValidateToken(ctx, shortToken)
		assert.Error(t, err)
	})
}

func TestAdminService_ChangePassword(t *testing.T) {
	db := setupTestDB(t)
	adminRepo := postgres.NewAdminRepository(db)
	logger := zap.NewNop()
	service := NewAdminService(adminRepo, "test-secret-key", 24*time.Hour, logger)

	ctx := context.Background()

	// Create test admin
	admin, err := service.CreateAdmin(ctx, "testadmin", "oldpassword")
	require.NoError(t, err)

	t.Run("Successful password change", func(t *testing.T) {
		err := service.ChangePassword(ctx, admin.ID, "oldpassword", "newpassword")
		require.NoError(t, err)

		// Verify new password works
		_, _, err = service.Login(ctx, "testadmin", "newpassword")
		assert.NoError(t, err)

		// Verify old password doesn't work
		_, _, err = service.Login(ctx, "testadmin", "oldpassword")
		assert.Error(t, err)
	})

	t.Run("Invalid old password", func(t *testing.T) {
		err := service.ChangePassword(ctx, admin.ID, "wrongoldpassword", "newpassword2")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid old password")
	})

	t.Run("Non-existent admin", func(t *testing.T) {
		err := service.ChangePassword(ctx, 99999, "oldpassword", "newpassword")
		assert.Error(t, err)
	})
}
