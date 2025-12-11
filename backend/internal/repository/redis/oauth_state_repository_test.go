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

// setupTestRedis creates a test Redis client with miniredis
func setupTestOAuthRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

// TestStoreState tests storing OAuth state
func TestStoreState(t *testing.T) {
	client, mr := setupTestOAuthRedis(t)
	defer mr.Close()

	repo := NewOAuthStateRepository(client)
	ctx := context.Background()

	tests := []struct {
		name      string
		state     string
		data      OAuthStateData
		ttl       time.Duration
		wantError bool
	}{
		{
			name:  "Store valid state",
			state: "test_state_123",
			data: OAuthStateData{
				CodeVerifier: "test_verifier_abc",
				CallbackURL:  "http://localhost:8888/callback",
				CreatedAt:    time.Now().Unix(),
			},
			ttl:       10 * time.Minute,
			wantError: false,
		},
		{
			name:  "Store state with long verifier",
			state: "state_456",
			data: OAuthStateData{
				CodeVerifier: "very_long_code_verifier_1234567890_abcdefghijklmnopqrstuvwxyz",
				CallbackURL:  "http://localhost:9999/oauth/callback",
				CreatedAt:    time.Now().Unix(),
			},
			ttl:       5 * time.Minute,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.StoreState(ctx, tt.state, tt.data, tt.ttl)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify data is stored
				exists, err := repo.StateExists(ctx, tt.state)
				require.NoError(t, err)
				assert.True(t, exists)
			}
		})
	}
}

// TestConsumeState tests consuming OAuth state (atomic get + delete)
func TestConsumeState(t *testing.T) {
	client, mr := setupTestOAuthRedis(t)
	defer mr.Close()

	repo := NewOAuthStateRepository(client)
	ctx := context.Background()

	// Store test state
	state := "test_state_consume"
	data := OAuthStateData{
		CodeVerifier: "verifier_xyz",
		CallbackURL:  "http://localhost:8080/callback",
		CreatedAt:    time.Now().Unix(),
	}

	err := repo.StoreState(ctx, state, data, 10*time.Minute)
	require.NoError(t, err)

	// Consume state
	consumedData, err := repo.ConsumeState(ctx, state)
	require.NoError(t, err)
	require.NotNil(t, consumedData)

	// Verify data matches
	assert.Equal(t, data.CodeVerifier, consumedData.CodeVerifier)
	assert.Equal(t, data.CallbackURL, consumedData.CallbackURL)
	assert.Equal(t, data.CreatedAt, consumedData.CreatedAt)

	// Verify state is deleted (consumed)
	exists, err := repo.StateExists(ctx, state)
	require.NoError(t, err)
	assert.False(t, exists, "State should be deleted after consumption")

	// Attempt to consume again should return nil
	consumedAgain, err := repo.ConsumeState(ctx, state)
	require.NoError(t, err)
	assert.Nil(t, consumedAgain, "Consuming again should return nil")
}

// TestConsumeState_NonExistent tests consuming non-existent state
func TestConsumeState_NonExistent(t *testing.T) {
	client, mr := setupTestOAuthRedis(t)
	defer mr.Close()

	repo := NewOAuthStateRepository(client)
	ctx := context.Background()

	// Attempt to consume non-existent state
	data, err := repo.ConsumeState(ctx, "non_existent_state")
	require.NoError(t, err)
	assert.Nil(t, data, "Non-existent state should return nil")
}

// TestStateExists tests checking state existence
func TestStateExists(t *testing.T) {
	client, mr := setupTestOAuthRedis(t)
	defer mr.Close()

	repo := NewOAuthStateRepository(client)
	ctx := context.Background()

	tests := []struct {
		name       string
		state      string
		storeFirst bool
		wantExists bool
	}{
		{
			name:       "Existing state",
			state:      "existing_state",
			storeFirst: true,
			wantExists: true,
		},
		{
			name:       "Non-existing state",
			state:      "non_existing_state",
			storeFirst: false,
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.storeFirst {
				data := OAuthStateData{
					CodeVerifier: "test_verifier",
					CallbackURL:  "http://localhost:8080/callback",
					CreatedAt:    time.Now().Unix(),
				}
				err := repo.StoreState(ctx, tt.state, data, 10*time.Minute)
				require.NoError(t, err)
			}

			exists, err := repo.StateExists(ctx, tt.state)
			require.NoError(t, err)
			assert.Equal(t, tt.wantExists, exists)
		})
	}
}

// TestStateExpiration tests that states expire after TTL
func TestStateExpiration(t *testing.T) {
	client, mr := setupTestOAuthRedis(t)
	defer mr.Close()

	repo := NewOAuthStateRepository(client)
	ctx := context.Background()

	state := "expiring_state"
	data := OAuthStateData{
		CodeVerifier: "test_verifier",
		CallbackURL:  "http://localhost:8080/callback",
		CreatedAt:    time.Now().Unix(),
	}

	// Store with 1 second TTL
	err := repo.StoreState(ctx, state, data, 1*time.Second)
	require.NoError(t, err)

	// Verify it exists immediately
	exists, err := repo.StateExists(ctx, state)
	require.NoError(t, err)
	assert.True(t, exists)

	// Fast forward time in miniredis
	mr.FastForward(2 * time.Second)

	// Verify it no longer exists
	exists, err = repo.StateExists(ctx, state)
	require.NoError(t, err)
	assert.False(t, exists, "State should be expired")
}

// TestConsumeState_Atomicity tests that ConsumeState is atomic
func TestConsumeState_Atomicity(t *testing.T) {
	client, mr := setupTestOAuthRedis(t)
	defer mr.Close()

	repo := NewOAuthStateRepository(client)
	ctx := context.Background()

	state := "atomic_state"
	data := OAuthStateData{
		CodeVerifier: "test_verifier",
		CallbackURL:  "http://localhost:8080/callback",
		CreatedAt:    time.Now().Unix(),
	}

	err := repo.StoreState(ctx, state, data, 10*time.Minute)
	require.NoError(t, err)

	// First consumption should succeed
	consumedData1, err := repo.ConsumeState(ctx, state)
	require.NoError(t, err)
	require.NotNil(t, consumedData1)

	// Second consumption should return nil (state already deleted)
	consumedData2, err := repo.ConsumeState(ctx, state)
	require.NoError(t, err)
	assert.Nil(t, consumedData2, "Second consume should return nil")

	// Verify state no longer exists
	exists, err := repo.StateExists(ctx, state)
	require.NoError(t, err)
	assert.False(t, exists)
}

// TestOAuthStateRepository_FullFlow tests the complete OAuth state lifecycle
func TestOAuthStateRepository_FullFlow(t *testing.T) {
	client, mr := setupTestOAuthRedis(t)
	defer mr.Close()

	repo := NewOAuthStateRepository(client)
	ctx := context.Background()

	// Step 1: Store state
	state := "full_flow_state"
	data := OAuthStateData{
		CodeVerifier: "full_flow_verifier_12345",
		CallbackURL:  "http://localhost:8888/oauth/callback",
		CreatedAt:    time.Now().Unix(),
	}

	err := repo.StoreState(ctx, state, data, 10*time.Minute)
	require.NoError(t, err)

	// Step 2: Check existence
	exists, err := repo.StateExists(ctx, state)
	require.NoError(t, err)
	assert.True(t, exists)

	// Step 3: Consume state (simulating callback)
	consumedData, err := repo.ConsumeState(ctx, state)
	require.NoError(t, err)
	require.NotNil(t, consumedData)
	assert.Equal(t, data.CodeVerifier, consumedData.CodeVerifier)
	assert.Equal(t, data.CallbackURL, consumedData.CallbackURL)

	// Step 4: Verify state is deleted
	exists, err = repo.StateExists(ctx, state)
	require.NoError(t, err)
	assert.False(t, exists, "State should be deleted after consumption")

	// Step 5: Attempt to consume again (replay attack simulation)
	consumedAgain, err := repo.ConsumeState(ctx, state)
	require.NoError(t, err)
	assert.Nil(t, consumedAgain, "Replay should fail")
}

// TestGetKey tests the key formatting
func TestGetKey(t *testing.T) {
	client, mr := setupTestOAuthRedis(t)
	defer mr.Close()

	repo := NewOAuthStateRepository(client).(*oauthStateRepository)

	tests := []struct {
		name    string
		state   string
		wantKey string
	}{
		{
			name:    "Simple state",
			state:   "abc123",
			wantKey: "oauth:codex:state:abc123",
		},
		{
			name:    "State with special characters",
			state:   "state_with-special.chars",
			wantKey: "oauth:codex:state:state_with-special.chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := repo.getKey(tt.state)
			assert.Equal(t, tt.wantKey, key)
		})
	}
}
