package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SemaphoreRepository interface {
	Acquire(ctx context.Context, key string, requestID string, limit int, ttl int) (bool, error)
	Release(ctx context.Context, key string, requestID string) error
	GetCount(ctx context.Context, key string) (int, error)
}

type semaphoreRepository struct {
	client        *redis.Client
	scriptManager *LuaScriptManager
}

func NewSemaphoreRepository(client *redis.Client) (SemaphoreRepository, error) {
	scriptManager := NewLuaScriptManager(client)
	if err := scriptManager.LoadScripts(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to load Lua scripts: %w", err)
	}

	return &semaphoreRepository{
		client:        client,
		scriptManager: scriptManager,
	}, nil
}

func (r *semaphoreRepository) Acquire(ctx context.Context, key string, requestID string, limit int, ttl int) (bool, error) {
	now := time.Now().Unix()

	result, err := r.scriptManager.EvalSHA(ctx, "acquire",
		[]string{key},
		limit, now, ttl, requestID,
	)

	if err != nil {
		return false, fmt.Errorf("acquire semaphore failed: %w", err)
	}

	acquired := result.(int64) == 1
	return acquired, nil
}

func (r *semaphoreRepository) Release(ctx context.Context, key string, requestID string) error {
	_, err := r.scriptManager.EvalSHA(ctx, "release",
		[]string{key},
		requestID,
	)

	if err != nil {
		return fmt.Errorf("release semaphore failed: %w", err)
	}

	return nil
}

func (r *semaphoreRepository) GetCount(ctx context.Context, key string) (int, error) {
	now := time.Now().Unix()

	result, err := r.scriptManager.EvalSHA(ctx, "count",
		[]string{key},
		now,
	)

	if err != nil {
		return 0, fmt.Errorf("get semaphore count failed: %w", err)
	}

	count := int(result.(int64))
	return count, nil
}
