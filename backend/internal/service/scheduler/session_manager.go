package scheduler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// SessionManager defines the interface for session hash management.
type SessionManager interface {
	CreateSessionHash(apiKeyID int64, conversationID string) string
}

type sessionManager struct{}

// NewSessionManager creates a new session manager.
func NewSessionManager() SessionManager {
	return &sessionManager{}
}

// CreateSessionHash creates a deterministic session hash from API key ID and conversation ID.
// Format: SHA256(apiKeyID + conversationID)
// Returns empty string if conversationID is empty (no sticky session).
func (m *sessionManager) CreateSessionHash(apiKeyID int64, conversationID string) string {
	if conversationID == "" {
		return "" // No session ID, sticky session disabled
	}

	data := fmt.Sprintf("%d:%s", apiKeyID, conversationID)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
