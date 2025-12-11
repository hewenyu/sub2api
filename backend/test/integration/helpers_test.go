package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/pkg/crypto"
)

const (
	testDBHost     = "localhost"
	testDBPort     = "5432"
	testDBUser     = "relay"
	testDBPassword = "relay123"
	testDBName     = "claude_relay_test"
)

// setupTestDB creates a test database connection and runs migrations
func setupTestDB(t *testing.T) *gorm.DB {
	// Ensure the test database exists. We connect to the default
	// "postgres" database and create the test DB if needed.
	adminDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable TimeZone=UTC",
		testDBHost, testDBPort, testDBUser, testDBPassword)

	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to postgres database for setup")

	var exists bool
	checkErr := adminDB.
		Raw("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = ?)", testDBName).
		Scan(&exists).Error
	require.NoError(t, checkErr, "Failed to check if test database exists")

	if !exists {
		createErr := adminDB.Exec("CREATE DATABASE " + testDBName).Error
		require.NoError(t, createErr, "Failed to create test database")
	}

	sqlAdminDB, err := adminDB.DB()
	require.NoError(t, err, "Failed to get SQL admin DB")
	require.NoError(t, sqlAdminDB.Close(), "Failed to close admin database")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		testDBHost, testDBPort, testDBUser, testDBPassword, testDBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to test database")

	// Auto migrate models
	err = db.AutoMigrate(
		&model.Admin{},
		&model.APIKey{},
		&model.CodexAccount{},
		&model.Usage{},
	)
	require.NoError(t, err, "Failed to run migrations")

	return db
}

// cleanupTestDB drops all tables and closes the database connection
func cleanupTestDB(t *testing.T, db *gorm.DB) {
	// Drop all tables
	err := db.Migrator().DropTable(
		&model.Usage{},
		&model.CodexAccount{},
		&model.APIKey{},
		&model.Admin{},
	)
	require.NoError(t, err, "Failed to drop tables")

	sqlDB, err := db.DB()
	require.NoError(t, err, "Failed to get SQL DB")

	err = sqlDB.Close()
	require.NoError(t, err, "Failed to close database")
}

// createTestAPIKey creates a test API key
func createTestAPIKey(t *testing.T, db *gorm.DB, name string) (*model.APIKey, string) {
	rawAPIKey, err := crypto.GenerateAPIKey()
	require.NoError(t, err, "Failed to generate API key")

	hashedAPIKey := crypto.HashAPIKey(rawAPIKey)
	prefix := crypto.GetAPIKeyPrefix(rawAPIKey)

	apiKey := &model.APIKey{
		Name:                  name,
		KeyHash:               hashedAPIKey,
		KeyPrefix:             prefix,
		IsActive:              true,
		MaxConcurrentRequests: 5,
		RateLimitPerMinute:    100,
		DailyCostLimit:        50.0,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err = db.Create(apiKey).Error
	require.NoError(t, err, "Failed to create API key")

	return apiKey, rawAPIKey
}

// createTestCodexAccount creates a test Codex account
func createTestCodexAccount(t *testing.T, db *gorm.DB, name string, accountType string) *model.CodexAccount {
	encryptedAPIKey, err := crypto.AES256Encrypt("sk-test-key-12345", "12345678901234567890123456789012")
	require.NoError(t, err, "Failed to encrypt API key")

	account := &model.CodexAccount{
		Name:        name,
		AccountType: accountType,
		APIKey:      &encryptedAPIKey,
		BaseAPI:     "https://api.openai.com/v1",
		DailyQuota:  100.0,
		DailyUsage:  0.0,
		Priority:    100,
		Schedulable: true,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = db.Create(account).Error
	require.NoError(t, err, "Failed to create Codex account")

	return account
}
