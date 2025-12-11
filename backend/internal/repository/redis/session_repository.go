package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SessionRepository defines the interface for session management in Redis.
type SessionRepository interface {
	Set(ctx context.Context, sessionHash string, data SessionData, ttl time.Duration) error
	Get(ctx context.Context, sessionHash string) (*SessionData, error)
	Delete(ctx context.Context, sessionHash string) error
	Exists(ctx context.Context, sessionHash string) (bool, error)
	ExtendTTL(ctx context.Context, sessionHash string, ttl time.Duration) error
	GetTTL(ctx context.Context, sessionHash string) (time.Duration, error)
}

// SessionData represents the session data stored in Redis.
type SessionData struct {
	AccountID   int64  `json:"account_id"`
	AccountType string `json:"account_type"`
	CreatedAt   int64  `json:"created_at"`
	LastUsedAt  int64  `json:"last_used_at"`
}

type sessionRepository struct {
	client *redis.Client
}

// NewSessionRepository creates a new session repository.
func NewSessionRepository(client *redis.Client) SessionRepository {
	return &sessionRepository{client: client}
}

func (r *sessionRepository) getKey(sessionHash string) string {
	return fmt.Sprintf("session:%s", sessionHash)
}

// Set stores session data in Redis with the specified TTL.
func (r *sessionRepository) Set(ctx context.Context, sessionHash string, data SessionData, ttl time.Duration) error {
	key := r.getKey(sessionHash)

	// Serialize to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Store to Redis
	if err := r.client.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set session: %w", err)
	}

	return nil
}

// Get retrieves session data from Redis.
func (r *sessionRepository) Get(ctx context.Context, sessionHash string) (*SessionData, error) {
	key := r.getKey(sessionHash)

	// Read from Redis
	jsonData, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Session does not exist
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Deserialize
	var data SessionData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &data, nil
}

// Delete removes a session from Redis.
func (r *sessionRepository) Delete(ctx context.Context, sessionHash string) error {
	key := r.getKey(sessionHash)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// Exists checks if a session exists in Redis.
func (r *sessionRepository) Exists(ctx context.Context, sessionHash string) (bool, error) {
	key := r.getKey(sessionHash)
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}
	return count > 0, nil
}

// ExtendTTL extends the TTL of a session.
func (r *sessionRepository) ExtendTTL(ctx context.Context, sessionHash string, ttl time.Duration) error {
	key := r.getKey(sessionHash)
	if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("failed to extend session TTL: %w", err)
	}
	return nil
}

// GetTTL retrieves the remaining TTL of a session.
func (r *sessionRepository) GetTTL(ctx context.Context, sessionHash string) (time.Duration, error) {
	key := r.getKey(sessionHash)
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get session TTL: %w", err)
	}
	return ttl, nil
}
