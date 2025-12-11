package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSHA256Hash(t *testing.T) {
	input := "test_api_key_12345"
	hash := SHA256Hash(input)

	assert.NotEmpty(t, hash)
	assert.Equal(t, 64, len(hash)) // SHA-256 produces 64 hex characters

	// Hash should be deterministic
	hash2 := SHA256Hash(input)
	assert.Equal(t, hash, hash2)

	// Different input should produce different hash
	hash3 := SHA256Hash("different_input")
	assert.NotEqual(t, hash, hash3)
}

func TestBcryptHash(t *testing.T) {
	password := "my_secure_password"

	hash, err := BcryptHash(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Verify password
	err = BcryptCompare(hash, password)
	assert.NoError(t, err)

	// Wrong password should fail
	err = BcryptCompare(hash, "wrong_password")
	assert.Error(t, err)
}

func TestBcryptHashDifferentEachTime(t *testing.T) {
	password := "test_password"

	hash1, err := BcryptHash(password)
	require.NoError(t, err)

	hash2, err := BcryptHash(password)
	require.NoError(t, err)

	// Hashes should be different (bcrypt uses random salt)
	assert.NotEqual(t, hash1, hash2)

	// But both should verify the same password
	assert.NoError(t, BcryptCompare(hash1, password))
	assert.NoError(t, BcryptCompare(hash2, password))
}

func TestGenerateAPIKey(t *testing.T) {
	apiKey, err := GenerateAPIKey()
	require.NoError(t, err)
	assert.NotEmpty(t, apiKey)
	assert.True(t, len(apiKey) > 64)
	assert.Equal(t, "cr_", apiKey[:3])

	// Should generate unique keys
	apiKey2, err := GenerateAPIKey()
	require.NoError(t, err)
	assert.NotEqual(t, apiKey, apiKey2)
}

func TestHashAPIKey(t *testing.T) {
	apiKey := "cr_1234567890abcdef"
	hash := HashAPIKey(apiKey)

	assert.NotEmpty(t, hash)
	assert.Equal(t, 64, len(hash))

	// Should be deterministic
	hash2 := HashAPIKey(apiKey)
	assert.Equal(t, hash, hash2)
}

func TestGetAPIKeyPrefix(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected string
	}{
		{
			name:     "Normal API key",
			apiKey:   "cr_1234567890abcdef",
			expected: "cr_1234567",
		},
		{
			name:     "Short API key",
			apiKey:   "cr_123",
			expected: "cr_123",
		},
		{
			name:     "Exactly 10 characters",
			apiKey:   "1234567890",
			expected: "1234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAPIKeyPrefix(tt.apiKey)
			assert.Equal(t, tt.expected, result)
		})
	}
}
