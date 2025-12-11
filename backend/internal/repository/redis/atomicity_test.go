package redis

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAtomicity_NoRaceConditions(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo, err := NewSemaphoreRepository(client)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:atomicity:race"
	limit := 50
	ttl := 300
	goroutines := 200

	var wg sync.WaitGroup
	successCount := atomic.Int32{}
	failureCount := atomic.Int32{}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			acquired, acquireErr := repo.Acquire(ctx, key, fmt.Sprintf("req-%d", id), limit, ttl)
			require.NoError(t, acquireErr)
			if acquired {
				successCount.Add(1)
			} else {
				failureCount.Add(1)
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int32(limit), successCount.Load())
	assert.Equal(t, int32(goroutines-limit), failureCount.Load())

	count, err := repo.GetCount(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, limit, count)
}

func TestAtomicity_ZombieLockPrevention(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo, err := NewSemaphoreRepository(client)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:atomicity:zombie"
	limit := 5
	shortTTL := 1

	for i := 0; i < limit; i++ {
		acquired, acquireErr := repo.Acquire(ctx, key, fmt.Sprintf("zombie-%d", i), limit, shortTTL)
		require.NoError(t, acquireErr)
		assert.True(t, acquired)
	}

	count, err := repo.GetCount(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, limit, count)

	time.Sleep(2 * time.Second)

	for i := 0; i < limit; i++ {
		acquired, acquireErr := repo.Acquire(ctx, key, fmt.Sprintf("new-%d", i), limit, 300)
		require.NoError(t, acquireErr)
		assert.True(t, acquired)
	}

	count, err = repo.GetCount(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, limit, count)
}

func TestAtomicity_ConcurrentReleaseAndAcquire(t *testing.T) {
	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo, err := NewSemaphoreRepository(client)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:atomicity:release-acquire"
	limit := 10
	ttl := 300
	iterations := 100

	for i := 0; i < limit; i++ {
		acquired, acquireErr := repo.Acquire(ctx, key, fmt.Sprintf("initial-%d", i), limit, ttl)
		require.NoError(t, acquireErr)
		assert.True(t, acquired)
	}

	var wg sync.WaitGroup
	successCount := atomic.Int32{}

	for i := 0; i < iterations; i++ {
		wg.Add(2)

		go func(id int) {
			defer wg.Done()
			_ = repo.Release(ctx, key, fmt.Sprintf("initial-%d", id%limit))
		}(i)

		go func(id int) {
			defer wg.Done()
			acquired, acquireErr := repo.Acquire(ctx, key, fmt.Sprintf("new-%d", id), limit, ttl)
			if acquireErr == nil && acquired {
				successCount.Add(1)
			}
		}(i)
	}

	wg.Wait()

	count, err := repo.GetCount(ctx, key)
	require.NoError(t, err)
	assert.LessOrEqual(t, count, limit)
}

func TestAtomicity_HighConcurrencyStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	client, _ := setupTestRedisClient(t)
	defer func() { _ = client.Close() }()

	repo, err := NewSemaphoreRepository(client)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:atomicity:stress"
	limit := 100
	ttl := 300
	goroutines := 1000

	var wg sync.WaitGroup
	successCount := atomic.Int32{}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			acquired, acquireErr := repo.Acquire(ctx, key, fmt.Sprintf("stress-%d", id), limit, ttl)
			if acquireErr == nil && acquired {
				successCount.Add(1)
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int32(limit), successCount.Load())

	count, err := repo.GetCount(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, limit, count)
}
