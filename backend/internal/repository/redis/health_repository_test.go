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

func setupHealthTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

func TestHealthRepository_GetMetrics_NewAccount(t *testing.T) {
	client, mr := setupHealthTestRedis(t)
	defer mr.Close()

	repo := NewHealthRepository(client)
	ctx := context.Background()

	metrics, err := repo.GetMetrics(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), metrics.AccountID)
	assert.Equal(t, 1.0, metrics.HealthScore)
	assert.Equal(t, int64(0), metrics.SuccessCount)
	assert.Equal(t, int64(0), metrics.FailureCount)
}

func TestHealthRepository_UpdateMetrics_Success(t *testing.T) {
	client, mr := setupHealthTestRedis(t)
	defer mr.Close()

	repo := NewHealthRepository(client)
	ctx := context.Background()

	err := repo.UpdateMetrics(ctx, 1, true)
	require.NoError(t, err)

	metrics, err := repo.GetMetrics(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), metrics.SuccessCount)
	assert.Equal(t, int64(0), metrics.FailureCount)
	assert.Equal(t, 0, metrics.ConsecutiveFailures)
	assert.Equal(t, 1.0, metrics.HealthScore)
}

func TestHealthRepository_UpdateMetrics_Failure(t *testing.T) {
	client, mr := setupHealthTestRedis(t)
	defer mr.Close()

	repo := NewHealthRepository(client)
	ctx := context.Background()

	err := repo.UpdateMetrics(ctx, 1, false)
	require.NoError(t, err)

	metrics, err := repo.GetMetrics(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(0), metrics.SuccessCount)
	assert.Equal(t, int64(1), metrics.FailureCount)
	assert.Equal(t, 1, metrics.ConsecutiveFailures)
	assert.True(t, metrics.HealthScore < 1.0)
}

func TestHealthRepository_UpdateMetrics_ConsecutiveFailures(t *testing.T) {
	client, mr := setupHealthTestRedis(t)
	defer mr.Close()

	repo := NewHealthRepository(client)
	ctx := context.Background()

	require.NoError(t, repo.UpdateMetrics(ctx, 1, false))
	require.NoError(t, repo.UpdateMetrics(ctx, 1, false))
	require.NoError(t, repo.UpdateMetrics(ctx, 1, false))

	metrics, err := repo.GetMetrics(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, 3, metrics.ConsecutiveFailures)
	assert.Equal(t, int64(3), metrics.FailureCount)
}

func TestHealthRepository_UpdateMetrics_ResetConsecutiveFailures(t *testing.T) {
	client, mr := setupHealthTestRedis(t)
	defer mr.Close()

	repo := NewHealthRepository(client)
	ctx := context.Background()

	require.NoError(t, repo.UpdateMetrics(ctx, 1, false))
	require.NoError(t, repo.UpdateMetrics(ctx, 1, false))
	require.NoError(t, repo.UpdateMetrics(ctx, 1, true))

	metrics, err := repo.GetMetrics(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, 0, metrics.ConsecutiveFailures)
	assert.Equal(t, int64(1), metrics.SuccessCount)
	assert.Equal(t, int64(2), metrics.FailureCount)
}

func TestHealthRepository_GetHealthScore(t *testing.T) {
	client, mr := setupHealthTestRedis(t)
	defer mr.Close()

	repo := NewHealthRepository(client)
	ctx := context.Background()

	score, err := repo.GetHealthScore(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, 1.0, score)

	require.NoError(t, repo.UpdateMetrics(ctx, 1, true))
	require.NoError(t, repo.UpdateMetrics(ctx, 1, false))

	score, err = repo.GetHealthScore(ctx, 1)
	require.NoError(t, err)
	assert.True(t, score < 1.0 && score > 0.0)
}

func TestHealthRepository_SetQuarantine(t *testing.T) {
	client, mr := setupHealthTestRedis(t)
	defer mr.Close()

	repo := NewHealthRepository(client)
	ctx := context.Background()

	err := repo.SetQuarantine(ctx, 1, 5*time.Minute)
	require.NoError(t, err)

	metrics, err := repo.GetMetrics(ctx, 1)
	require.NoError(t, err)
	assert.True(t, metrics.QuarantineUntil.After(time.Now()))
	assert.Equal(t, 1, metrics.QuarantineCount)
}

func TestHealthRepository_IsQuarantined(t *testing.T) {
	client, mr := setupHealthTestRedis(t)
	defer mr.Close()

	repo := NewHealthRepository(client)
	ctx := context.Background()

	quarantined, err := repo.IsQuarantined(ctx, 1)
	require.NoError(t, err)
	assert.False(t, quarantined)

	err = repo.SetQuarantine(ctx, 1, 5*time.Minute)
	require.NoError(t, err)

	quarantined, err = repo.IsQuarantined(ctx, 1)
	require.NoError(t, err)
	assert.True(t, quarantined)
}

func TestHealthRepository_GetLastCheckTime(t *testing.T) {
	client, mr := setupHealthTestRedis(t)
	defer mr.Close()

	repo := NewHealthRepository(client)
	ctx := context.Background()

	lastCheck, err := repo.GetLastCheckTime(ctx, 1)
	require.NoError(t, err)
	assert.True(t, lastCheck.IsZero())

	require.NoError(t, repo.UpdateMetrics(ctx, 1, true))

	lastCheck, err = repo.GetLastCheckTime(ctx, 1)
	require.NoError(t, err)
	assert.False(t, lastCheck.IsZero())
	assert.True(t, time.Since(lastCheck) < time.Second)
}
