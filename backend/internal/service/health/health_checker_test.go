package health

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	redisrepo "github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
)

func setupHealthCheckerTest(t *testing.T) (*healthChecker, *miniredis.Miniredis, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	healthRepo := redisrepo.NewHealthRepository(redisClient)

	logger := zap.NewNop()
	checker := &healthChecker{
		accountRepo:   nil,
		healthRepo:    healthRepo,
		checkInterval: 100 * time.Millisecond,
		idleThreshold: 5 * time.Minute,
		logger:        logger,
		stopCh:        make(chan struct{}),
	}

	cleanupFunc := func() {
		mr.Close()
	}

	return checker, mr, cleanupFunc
}

func TestHealthChecker_ReportResult_Success(t *testing.T) {
	checker, _, cleanup := setupHealthCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	err := checker.ReportResult(ctx, 1, true)
	require.NoError(t, err)

	metrics, err := checker.healthRepo.GetMetrics(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), metrics.SuccessCount)
	assert.Equal(t, int64(0), metrics.FailureCount)
}

func TestHealthChecker_ReportResult_Failure(t *testing.T) {
	checker, _, cleanup := setupHealthCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	err := checker.ReportResult(ctx, 1, false)
	require.NoError(t, err)

	metrics, err := checker.healthRepo.GetMetrics(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(0), metrics.SuccessCount)
	assert.Equal(t, int64(1), metrics.FailureCount)
}

func TestHealthChecker_ReportResult_QuarantineAfterFailures(t *testing.T) {
	checker, _, cleanup := setupHealthCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	require.NoError(t, checker.ReportResult(ctx, 1, false))
	require.NoError(t, checker.ReportResult(ctx, 1, false))
	require.NoError(t, checker.ReportResult(ctx, 1, false))

	quarantined, err := checker.healthRepo.IsQuarantined(ctx, 1)
	require.NoError(t, err)
	assert.True(t, quarantined)
}

func TestHealthChecker_StartStop(t *testing.T) {
	checker, _, cleanup := setupHealthCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	checker.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	checker.Stop()
}
