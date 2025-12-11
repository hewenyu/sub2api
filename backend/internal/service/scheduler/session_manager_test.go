package scheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionManager_CreateSessionHash(t *testing.T) {
	manager := NewSessionManager()

	apiKeyID := int64(123)
	conversationID := "conv-abc-123"

	hash := manager.CreateSessionHash(apiKeyID, conversationID)

	// Should generate a non-empty hash
	assert.NotEmpty(t, hash)

	// Should be deterministic (same inputs = same hash)
	hash2 := manager.CreateSessionHash(apiKeyID, conversationID)
	assert.Equal(t, hash, hash2)

	// Different inputs should produce different hashes
	hash3 := manager.CreateSessionHash(apiKeyID, "different-conv")
	assert.NotEqual(t, hash, hash3)

	hash4 := manager.CreateSessionHash(456, conversationID)
	assert.NotEqual(t, hash, hash4)
}

func TestSessionManager_CreateSessionHash_EmptyConversationID(t *testing.T) {
	manager := NewSessionManager()

	hash := manager.CreateSessionHash(123, "")

	// Should return empty string for empty conversation ID
	assert.Empty(t, hash)
}

func TestSessionManager_CreateSessionHash_Length(t *testing.T) {
	manager := NewSessionManager()

	hash := manager.CreateSessionHash(123, "conv-123")

	// SHA256 produces 64-character hex string
	assert.Len(t, hash, 64)
}

func TestSessionManager_CreateSessionHash_Deterministic(t *testing.T) {
	manager := NewSessionManager()

	apiKeyID := int64(999)
	conversationID := "conv-test-deterministic"

	// Generate multiple times
	hashes := make([]string, 10)
	for i := 0; i < 10; i++ {
		hashes[i] = manager.CreateSessionHash(apiKeyID, conversationID)
	}

	// All should be identical
	for i := 1; i < len(hashes); i++ {
		assert.Equal(t, hashes[0], hashes[i])
	}
}

func TestSessionManager_CreateSessionHash_UniquePerAPIKey(t *testing.T) {
	manager := NewSessionManager()

	conversationID := "same-conv"

	hash1 := manager.CreateSessionHash(1, conversationID)
	hash2 := manager.CreateSessionHash(2, conversationID)
	hash3 := manager.CreateSessionHash(3, conversationID)

	// All should be different
	assert.NotEqual(t, hash1, hash2)
	assert.NotEqual(t, hash2, hash3)
	assert.NotEqual(t, hash1, hash3)
}

func TestSessionManager_CreateSessionHash_UniquePerConversation(t *testing.T) {
	manager := NewSessionManager()

	apiKeyID := int64(100)

	hash1 := manager.CreateSessionHash(apiKeyID, "conv-1")
	hash2 := manager.CreateSessionHash(apiKeyID, "conv-2")
	hash3 := manager.CreateSessionHash(apiKeyID, "conv-3")

	// All should be different
	assert.NotEqual(t, hash1, hash2)
	assert.NotEqual(t, hash2, hash3)
	assert.NotEqual(t, hash1, hash3)
}
