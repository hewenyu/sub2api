package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/config"
)

// RedisClient wraps the Redis client.
type RedisClient struct {
	*redis.Client
}

// NewRedisClient creates a new Redis client.
func NewRedisClient(cfg *config.RedisConfig, logger *zap.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connected successfully",
		zap.String("addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
		zap.Int("db", cfg.DB),
	)

	return &RedisClient{Client: client}, nil
}

// Close closes the Redis connection.
func (r *RedisClient) Close() error {
	return r.Client.Close()
}
