package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedisClient(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	// Create in-memory Redis server
	mr := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

func TestSessionRepository_Set(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewSessionRepository(client)
	ctx := context.Background()

	sessionHash := "test-session-hash-123"
	sessionData := SessionData{
		AccountID:   100,
		AccountType: "openai-oauth",
		CreatedAt:   time.Now().Unix(),
		LastUsedAt:  time.Now().Unix(),
	}
	ttl := 1 * time.Hour

	err := repo.Set(ctx, sessionHash, sessionData, ttl)
	require.NoError(t, err)

	// Verify data was stored
	retrieved, err := repo.Get(ctx, sessionHash)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, sessionData.AccountID, retrieved.AccountID)
	assert.Equal(t, sessionData.AccountType, retrieved.AccountType)
}

func TestSessionRepository_Get_NonExistent(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewSessionRepository(client)
	ctx := context.Background()

	retrieved, err := repo.Get(ctx, "non-existent-session")
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestSessionRepository_Delete(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewSessionRepository(client)
	ctx := context.Background()

	sessionHash := "test-session-to-delete"
	sessionData := SessionData{
		AccountID:   200,
		AccountType: "openai-responses",
		CreatedAt:   time.Now().Unix(),
		LastUsedAt:  time.Now().Unix(),
	}

	// Create session
	err := repo.Set(ctx, sessionHash, sessionData, 1*time.Hour)
	require.NoError(t, err)

	// Delete session
	err = repo.Delete(ctx, sessionHash)
	require.NoError(t, err)

	// Verify it's gone
	retrieved, err := repo.Get(ctx, sessionHash)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestSessionRepository_Exists(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewSessionRepository(client)
	ctx := context.Background()

	sessionHash := "test-session-exists"
	sessionData := SessionData{
		AccountID:   300,
		AccountType: "openai-oauth",
		CreatedAt:   time.Now().Unix(),
		LastUsedAt:  time.Now().Unix(),
	}

	// Before creating
	exists, err := repo.Exists(ctx, sessionHash)
	require.NoError(t, err)
	assert.False(t, exists)

	// After creating
	err = repo.Set(ctx, sessionHash, sessionData, 1*time.Hour)
	require.NoError(t, err)

	exists, err = repo.Exists(ctx, sessionHash)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestSessionRepository_ExtendTTL(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewSessionRepository(client)
	ctx := context.Background()

	sessionHash := "test-session-extend-ttl"
	sessionData := SessionData{
		AccountID:   400,
		AccountType: "openai-responses",
		CreatedAt:   time.Now().Unix(),
		LastUsedAt:  time.Now().Unix(),
	}

	// Create with 10 seconds TTL
	err := repo.Set(ctx, sessionHash, sessionData, 10*time.Second)
	require.NoError(t, err)

	// Extend to 1 hour
	err = repo.ExtendTTL(ctx, sessionHash, 1*time.Hour)
	require.NoError(t, err)

	// Verify TTL is extended
	ttl, err := repo.GetTTL(ctx, sessionHash)
	require.NoError(t, err)
	assert.Greater(t, ttl, 30*time.Minute) // Should be close to 1 hour
}

func TestSessionRepository_GetTTL(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewSessionRepository(client)
	ctx := context.Background()

	sessionHash := "test-session-get-ttl"
	sessionData := SessionData{
		AccountID:   500,
		AccountType: "openai-oauth",
		CreatedAt:   time.Now().Unix(),
		LastUsedAt:  time.Now().Unix(),
	}

	// Create with 30 minutes TTL
	err := repo.Set(ctx, sessionHash, sessionData, 30*time.Minute)
	require.NoError(t, err)

	// Get TTL
	ttl, err := repo.GetTTL(ctx, sessionHash)
	require.NoError(t, err)
	assert.Greater(t, ttl, 25*time.Minute) // Should be close to 30 minutes
	assert.LessOrEqual(t, ttl, 30*time.Minute)
}

func TestSessionRepository_TTLExpiry(t *testing.T) {
	client, mr := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewSessionRepository(client)
	ctx := context.Background()

	sessionHash := "test-session-expiry"
	sessionData := SessionData{
		AccountID:   600,
		AccountType: "openai-responses",
		CreatedAt:   time.Now().Unix(),
		LastUsedAt:  time.Now().Unix(),
	}

	// Create with 2 seconds TTL
	err := repo.Set(ctx, sessionHash, sessionData, 2*time.Second)
	require.NoError(t, err)

	// Verify it exists
	exists, err := repo.Exists(ctx, sessionHash)
	require.NoError(t, err)
	assert.True(t, exists)

	// Fast forward time in miniredis
	mr.FastForward(3 * time.Second)

	// Verify it's gone
	exists, err = repo.Exists(ctx, sessionHash)
	require.NoError(t, err)
	assert.False(t, exists)
}
