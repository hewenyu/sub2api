package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	BcryptCost   = 10
	APIKeyLength = 64
	APIKeyPrefix = "cr_"
)

// SHA256Hash computes SHA-256 hash of input string
func SHA256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// BcryptHash generates bcrypt hash of password
func BcryptHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// BcryptCompare verifies password against bcrypt hash
func BcryptCompare(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("password mismatch")
	}
	return nil
}

// GenerateAPIKey generates a new API key with "cr_" prefix
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, APIKeyLength/2) // hex encoding doubles length
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return APIKeyPrefix + hex.EncodeToString(bytes), nil
}

// HashAPIKey computes SHA-256 hash of API key for storage
func HashAPIKey(apiKey string) string {
	return SHA256Hash(apiKey)
}

// GetAPIKeyPrefix extracts first 10 characters for display
func GetAPIKeyPrefix(apiKey string) string {
	if len(apiKey) < 10 {
		return apiKey
	}
	return apiKey[:10]
}
