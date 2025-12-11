package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrencyRepository_Acquire(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewConcurrencyRepository(client)
	ctx := context.Background()

	accountID := int64(100)
	requestID := "request-123"
	leaseSeconds := 300

	acquired, err := repo.Acquire(ctx, accountID, requestID, leaseSeconds)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Verify concurrency count increased
	count, err := repo.GetCount(ctx, accountID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestConcurrencyRepository_Acquire_Multiple(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewConcurrencyRepository(client)
	ctx := context.Background()

	accountID := int64(200)
	leaseSeconds := 300

	// Acquire 5 slots
	for i := 0; i < 5; i++ {
		requestID := fmt.Sprintf("request-%d", i)
		acquired, err := repo.Acquire(ctx, accountID, requestID, leaseSeconds)
		require.NoError(t, err)
		assert.True(t, acquired)
	}

	// Verify count
	count, err := repo.GetCount(ctx, accountID)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestConcurrencyRepository_Release(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewConcurrencyRepository(client)
	ctx := context.Background()

	accountID := int64(300)
	requestID := "request-to-release"
	leaseSeconds := 300

	// Acquire
	acquired, err := repo.Acquire(ctx, accountID, requestID, leaseSeconds)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Release
	err = repo.Release(ctx, accountID, requestID)
	require.NoError(t, err)

	// Verify count is 0
	count, err := repo.GetCount(ctx, accountID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestConcurrencyRepository_GetCount_Empty(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewConcurrencyRepository(client)
	ctx := context.Background()

	count, err := repo.GetCount(ctx, int64(999))
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestConcurrencyRepository_Cleanup_ExpiredRecords(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewConcurrencyRepository(client)
	ctx := context.Background()

	accountID := int64(400)

	// Acquire with short lease (1 second)
	acquired, err := repo.Acquire(ctx, accountID, "expired-request", 1)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Acquire with long lease (300 seconds)
	acquired, err = repo.Acquire(ctx, accountID, "valid-request", 300)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Count should be 2
	count, err := repo.GetCount(ctx, accountID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Wait for short lease to expire
	time.Sleep(2 * time.Second)

	// Cleanup expired records
	err = repo.Cleanup(ctx, accountID)
	require.NoError(t, err)

	// Count should be 1 (only valid-request remains)
	count, err = repo.GetCount(ctx, accountID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestConcurrencyRepository_GetCount_AutoCleanup(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewConcurrencyRepository(client)
	ctx := context.Background()

	accountID := int64(500)

	// Acquire with short lease (1 second)
	acquired, err := repo.Acquire(ctx, accountID, "short-lease", 1)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// GetCount should auto-cleanup and return 0
	count, err := repo.GetCount(ctx, accountID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestConcurrencyRepository_DifferentAccounts(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewConcurrencyRepository(client)
	ctx := context.Background()

	// Acquire for account 1
	acquired, err := repo.Acquire(ctx, 1, "req-1", 300)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Acquire for account 2
	acquired, err = repo.Acquire(ctx, 2, "req-2", 300)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Verify counts are isolated
	count1, err := repo.GetCount(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count1)

	count2, err := repo.GetCount(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count2)
}

func TestConcurrencyRepository_DuplicateRequestID(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo := NewConcurrencyRepository(client)
	ctx := context.Background()

	accountID := int64(600)
	requestID := "duplicate-request"

	// Acquire first time
	acquired, err := repo.Acquire(ctx, accountID, requestID, 300)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Acquire again with same request ID (should update, not add)
	acquired, err = repo.Acquire(ctx, accountID, requestID, 300)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Count should still be 1
	count, err := repo.GetCount(ctx, accountID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
