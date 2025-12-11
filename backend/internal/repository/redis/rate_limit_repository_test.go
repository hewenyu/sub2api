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

func setupRateLimitTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return mr, client
}

func TestRateLimitRepository_Increment(t *testing.T) {
	mr, client := setupRateLimitTestRedis(t)
	defer func() { mr.Close() }()

	repo := NewRateLimitRepository(client)
	ctx := context.Background()

	// Test first increment
	count, err := repo.Increment(ctx, 1, "window1", 60*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Test second increment
	count, err = repo.Increment(ctx, 1, "window1", 60*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Verify TTL is set
	ttl, err := repo.GetTTL(ctx, 1, "window1")
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
	assert.LessOrEqual(t, ttl, 60*time.Second)
}

func TestRateLimitRepository_GetCount(t *testing.T) {
	mr, client := setupRateLimitTestRedis(t)
	defer func() { mr.Close() }()

	repo := NewRateLimitRepository(client)
	ctx := context.Background()

	// Test non-existent key
	count, err := repo.GetCount(ctx, 1, "window1")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Add some counts
	_, err = repo.Increment(ctx, 1, "window1", 60*time.Second)
	require.NoError(t, err)
	_, err = repo.Increment(ctx, 1, "window1", 60*time.Second)
	require.NoError(t, err)

	// Test existing key
	count, err = repo.GetCount(ctx, 1, "window1")
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestRateLimitRepository_Delete(t *testing.T) {
	mr, client := setupRateLimitTestRedis(t)
	defer func() { mr.Close() }()

	repo := NewRateLimitRepository(client)
	ctx := context.Background()

	// Create a counter
	_, err := repo.Increment(ctx, 1, "window1", 60*time.Second)
	require.NoError(t, err)

	// Verify it exists
	count, err := repo.GetCount(ctx, 1, "window1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete it
	err = repo.Delete(ctx, 1, "window1")
	require.NoError(t, err)

	// Verify it's gone
	count, err = repo.GetCount(ctx, 1, "window1")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRateLimitRepository_MultipleAPIKeys(t *testing.T) {
	mr, client := setupRateLimitTestRedis(t)
	defer func() { mr.Close() }()

	repo := NewRateLimitRepository(client)
	ctx := context.Background()

	// Increment for different API keys
	count1, err := repo.Increment(ctx, 1, "window1", 60*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count1)

	count2, err := repo.Increment(ctx, 2, "window1", 60*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count2)

	count1, err = repo.Increment(ctx, 1, "window1", 60*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count1)

	// Verify counts are separate
	finalCount1, err := repo.GetCount(ctx, 1, "window1")
	require.NoError(t, err)
	assert.Equal(t, int64(2), finalCount1)

	finalCount2, err := repo.GetCount(ctx, 2, "window1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), finalCount2)
}

func TestRateLimitRepository_MultipleWindows(t *testing.T) {
	mr, client := setupRateLimitTestRedis(t)
	defer func() { mr.Close() }()

	repo := NewRateLimitRepository(client)
	ctx := context.Background()

	// Increment for different windows
	count1, err := repo.Increment(ctx, 1, "window1", 60*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count1)

	count2, err := repo.Increment(ctx, 1, "window2", 60*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count2)

	count1, err = repo.Increment(ctx, 1, "window1", 60*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count1)

	// Verify counts are separate
	finalCount1, err := repo.GetCount(ctx, 1, "window1")
	require.NoError(t, err)
	assert.Equal(t, int64(2), finalCount1)

	finalCount2, err := repo.GetCount(ctx, 1, "window2")
	require.NoError(t, err)
	assert.Equal(t, int64(1), finalCount2)
}
