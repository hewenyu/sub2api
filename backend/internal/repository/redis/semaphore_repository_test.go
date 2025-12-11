package redis

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSemaphoreRepository_Acquire(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo, err := NewSemaphoreRepository(client)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:semaphore:1"
	requestID := "req-1"
	limit := 5
	ttl := 300

	acquired, err := repo.Acquire(ctx, key, requestID, limit, ttl)
	require.NoError(t, err)
	assert.True(t, acquired)

	count, err := repo.GetCount(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSemaphoreRepository_Acquire_LimitEnforcement(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo, err := NewSemaphoreRepository(client)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:semaphore:2"
	limit := 3
	ttl := 300

	for i := 0; i < limit; i++ {
		acquired, acquireErr := repo.Acquire(ctx, key, fmt.Sprintf("req-%d", i), limit, ttl)
		require.NoError(t, acquireErr)
		assert.True(t, acquired)
	}

	acquired, err := repo.Acquire(ctx, key, "req-overflow", limit, ttl)
	require.NoError(t, err)
	assert.False(t, acquired)

	count, err := repo.GetCount(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, limit, count)
}

func TestSemaphoreRepository_Release(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo, err := NewSemaphoreRepository(client)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:semaphore:3"
	requestID := "req-release"
	limit := 5
	ttl := 300

	acquired, err := repo.Acquire(ctx, key, requestID, limit, ttl)
	require.NoError(t, err)
	assert.True(t, acquired)

	err = repo.Release(ctx, key, requestID)
	require.NoError(t, err)

	count, err := repo.GetCount(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestSemaphoreRepository_AutomaticCleanup(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo, err := NewSemaphoreRepository(client)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:semaphore:4"
	limit := 5

	acquired, err := repo.Acquire(ctx, key, "req-expired", limit, 1)
	require.NoError(t, err)
	assert.True(t, acquired)

	acquired, err = repo.Acquire(ctx, key, "req-valid", limit, 300)
	require.NoError(t, err)
	assert.True(t, acquired)

	time.Sleep(2 * time.Second)

	count, err := repo.GetCount(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSemaphoreRepository_ConcurrentAcquire(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo, err := NewSemaphoreRepository(client)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:semaphore:5"
	limit := 10
	ttl := 300
	goroutines := 100

	var wg sync.WaitGroup
	successCount := int32(0)
	var mu sync.Mutex

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			acquired, acquireErr := repo.Acquire(ctx, key, fmt.Sprintf("req-%d", id), limit, ttl)
			if acquireErr == nil && acquired {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int32(limit), successCount)

	count, err := repo.GetCount(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, limit, count)
}

func TestSemaphoreRepository_DuplicateRequestID(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo, err := NewSemaphoreRepository(client)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:semaphore:6"
	requestID := "duplicate-req"
	limit := 5
	ttl := 300

	acquired, err := repo.Acquire(ctx, key, requestID, limit, ttl)
	require.NoError(t, err)
	assert.True(t, acquired)

	acquired, err = repo.Acquire(ctx, key, requestID, limit, ttl)
	require.NoError(t, err)
	assert.True(t, acquired)

	count, err := repo.GetCount(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSemaphoreRepository_ReleaseIdempotency(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo, err := NewSemaphoreRepository(client)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:semaphore:7"
	requestID := "req-idempotent"
	limit := 5
	ttl := 300

	acquired, err := repo.Acquire(ctx, key, requestID, limit, ttl)
	require.NoError(t, err)
	assert.True(t, acquired)

	err = repo.Release(ctx, key, requestID)
	require.NoError(t, err)

	err = repo.Release(ctx, key, requestID)
	require.NoError(t, err)

	count, err := repo.GetCount(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestSemaphoreRepository_GetCount_Empty(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo, err := NewSemaphoreRepository(client)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:semaphore:nonexistent"

	count, err := repo.GetCount(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
