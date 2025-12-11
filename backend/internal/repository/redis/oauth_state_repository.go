package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// OAuthStateRepository defines the interface for OAuth state management in Redis.
// OAuth states are temporary session data used during the authorization code flow
// to prevent CSRF attacks and store PKCE verifiers.
type OAuthStateRepository interface {
	// StoreState stores OAuth state data with a TTL.
	// The state will automatically expire after the specified duration.
	StoreState(ctx context.Context, state string, data OAuthStateData, ttl time.Duration) error

	// ConsumeState atomically retrieves and deletes the OAuth state.
	// This ensures the state can only be used once (replay attack prevention).
	// Returns nil if the state does not exist or has expired.
	ConsumeState(ctx context.Context, state string) (*OAuthStateData, error)

	// StateExists checks if an OAuth state exists in Redis.
	StateExists(ctx context.Context, state string) (bool, error)
}

// OAuthStateData represents the data stored for an OAuth state.
// This data is used to validate the OAuth callback and complete the token exchange.
type OAuthStateData struct {
	// CodeVerifier is the PKCE code verifier (RFC 7636)
	CodeVerifier string `json:"code_verifier"`

	// CallbackURL is the URL where the OAuth callback will be received
	CallbackURL string `json:"callback_url"`

	// CreatedAt is the timestamp when the state was created (Unix seconds)
	CreatedAt int64 `json:"created_at"`
}

type oauthStateRepository struct {
	client *redis.Client
}

// NewOAuthStateRepository creates a new OAuth state repository.
func NewOAuthStateRepository(client *redis.Client) OAuthStateRepository {
	return &oauthStateRepository{client: client}
}

// getKey returns the Redis key for an OAuth state.
// Key pattern: oauth:codex:state:{state}
func (r *oauthStateRepository) getKey(state string) string {
	return fmt.Sprintf("oauth:codex:state:%s", state)
}

// StoreState stores OAuth state data in Redis with the specified TTL.
// The data is serialized to JSON before storage.
//
// Example usage:
//
//	data := OAuthStateData{
//	    CodeVerifier: "abc123...",
//	    CallbackURL: "http://localhost:8888/callback",
//	    CreatedAt: time.Now().Unix(),
//	}
//	err := repo.StoreState(ctx, "state_xyz", data, 10*time.Minute)
func (r *oauthStateRepository) StoreState(ctx context.Context, state string, data OAuthStateData, ttl time.Duration) error {
	key := r.getKey(state)

	// Serialize data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal OAuth state data: %w", err)
	}

	// Store in Redis with TTL
	if err := r.client.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store OAuth state: %w", err)
	}

	return nil
}

// ConsumeState atomically retrieves and deletes an OAuth state from Redis.
// This is implemented using a Lua script to ensure atomicity.
//
// IMPORTANT: This method ensures the state can only be consumed once,
// preventing replay attacks where an attacker tries to reuse a valid state.
//
// Returns:
//   - *OAuthStateData: The state data if found
//   - nil: If the state does not exist or has expired
//   - error: If there's a Redis error (not including "key not found")
//
// Security Note: The atomic GET+DEL operation is critical for CSRF protection.
// Without atomicity, an attacker could potentially:
// 1. Read the state value
// 2. Use it in a forged callback
// 3. Before the legitimate user's callback completes
func (r *oauthStateRepository) ConsumeState(ctx context.Context, state string) (*OAuthStateData, error) {
	key := r.getKey(state)

	// Lua script for atomic GET + DELETE
	// This ensures the state can only be consumed once
	script := `
		local key = KEYS[1]
		local value = redis.call('GET', key)
		if value then
			redis.call('DEL', key)
			return value
		else
			return nil
		end
	`

	// Execute Lua script
	result, err := r.client.Eval(ctx, script, []string{key}).Result()
	if err != nil {
		// redis.Nil means the key doesn't exist, which is not an error
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to consume OAuth state: %w", err)
	}

	// If result is nil, the key didn't exist
	if result == nil {
		return nil, nil
	}

	// Deserialize JSON data
	var data OAuthStateData
	jsonStr, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type from Lua script: %T", result)
	}

	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OAuth state data: %w", err)
	}

	return &data, nil
}

// StateExists checks if an OAuth state exists in Redis.
// This is useful for debugging and testing, but should not be used
// in production code paths (use ConsumeState instead for atomicity).
func (r *oauthStateRepository) StateExists(ctx context.Context, state string) (bool, error) {
	key := r.getKey(state)

	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check OAuth state existence: %w", err)
	}

	return count > 0, nil
}
