package crypto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAES256Encrypt(t *testing.T) {
	key := "12345678901234567890123456789012" // 32 bytes
	plaintext := "my_secret_api_key_12345"

	ciphertext, err := AES256Encrypt(plaintext, key)
	require.NoError(t, err)
	assert.NotEmpty(t, ciphertext)
	assert.Contains(t, ciphertext, ":") // Contains IV and ciphertext separator

	// Verify format: {base64}:{base64}
	parts := strings.Split(ciphertext, ":")
	assert.Len(t, parts, 2)
	assert.NotEmpty(t, parts[0]) // IV
	assert.NotEmpty(t, parts[1]) // Ciphertext
}

func TestAES256Encrypt_Decrypt_RoundTrip(t *testing.T) {
	key := "12345678901234567890123456789012" // 32 bytes
	plaintext := "my_secret_api_key_12345"

	// Encrypt
	ciphertext, err := AES256Encrypt(plaintext, key)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := AES256Decrypt(ciphertext, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestAES256Encrypt_DifferentPlaintexts(t *testing.T) {
	key := "12345678901234567890123456789012"
	testCases := []string{
		"",
		"a",
		"short",
		"this is a longer plaintext to test padding",
		"sk-1234567890abcdefghijklmnopqrstuvwxyz",
		strings.Repeat("x", 100),
	}

	for _, plaintext := range testCases {
		t.Run(plaintext, func(t *testing.T) {
			ciphertext, err := AES256Encrypt(plaintext, key)
			require.NoError(t, err)

			decrypted, err := AES256Decrypt(ciphertext, key)
			require.NoError(t, err)
			assert.Equal(t, plaintext, decrypted)
		})
	}
}

func TestAES256Encrypt_InvalidKeyLength(t *testing.T) {
	testCases := []struct {
		name string
		key  string
	}{
		{"empty", ""},
		{"too short", "short_key"},
		{"16 bytes (AES-128)", "1234567890123456"},
		{"24 bytes (AES-192)", "123456789012345678901234"},
		{"too long", "123456789012345678901234567890123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := AES256Encrypt("test", tc.key)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "key must be 32 bytes")
		})
	}
}

func TestAES256Decrypt_InvalidFormat(t *testing.T) {
	key := "12345678901234567890123456789012"

	testCases := []struct {
		name       string
		ciphertext string
		errMsg     string
	}{
		{
			name:       "no separator",
			ciphertext: "invalidformat",
			errMsg:     "invalid ciphertext format",
		},
		{
			name:       "invalid base64 IV",
			ciphertext: "invalid!!!:AAAAAAAAAAAAAAAAAAAAAA==",
			errMsg:     "failed to decode IV",
		},
		{
			name:       "invalid base64 ciphertext",
			ciphertext: "AAAAAAAAAAAAAAAAAAAAAA==:invalid!!!",
			errMsg:     "failed to decode ciphertext",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := AES256Decrypt(tc.ciphertext, key)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

func TestAES256Decrypt_InvalidKeyLength(t *testing.T) {
	_, err := AES256Decrypt("AAAA:BBBB", "short_key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key must be 32 bytes")
}

func TestAES256Decrypt_CiphertextTooShort(t *testing.T) {
	key := "12345678901234567890123456789012"
	// Valid base64 but too short ciphertext
	ciphertext := "AAAAAAAAAAAAAAAAAAAAAA==:AA=="

	_, err := AES256Decrypt(ciphertext, key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ciphertext too short")
}

func TestAES256Decrypt_InvalidBlockSize(t *testing.T) {
	key := "12345678901234567890123456789012"
	// Valid base64 but ciphertext not multiple of block size (17 bytes = not multiple of 16)
	// AQEBAQEBAQEBAQEBAQEBAQE= decodes to 17 bytes
	ciphertext := "AAAAAAAAAAAAAAAAAAAAAA==:AQEBAQEBAQEBAQEBAQEBAQE="

	_, err := AES256Decrypt(ciphertext, key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ciphertext is not a multiple of block size")
}

func TestAES256Encrypt_DifferentIVs(t *testing.T) {
	key := "12345678901234567890123456789012"
	plaintext := "same_plaintext"

	// Encrypt same plaintext twice
	ciphertext1, err := AES256Encrypt(plaintext, key)
	require.NoError(t, err)

	ciphertext2, err := AES256Encrypt(plaintext, key)
	require.NoError(t, err)

	// Ciphertexts should be different (different IVs)
	assert.NotEqual(t, ciphertext1, ciphertext2)

	// Both should decrypt to the same plaintext
	decrypted1, err := AES256Decrypt(ciphertext1, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted1)

	decrypted2, err := AES256Decrypt(ciphertext2, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted2)
}

func TestAES256Decrypt_WrongKey(t *testing.T) {
	key1 := "12345678901234567890123456789012"
	key2 := "abcdefghijklmnopqrstuvwxyz123456"
	plaintext := "secret_data"

	// Encrypt with key1
	ciphertext, err := AES256Encrypt(plaintext, key1)
	require.NoError(t, err)

	// Try to decrypt with key2
	_, err = AES256Decrypt(ciphertext, key2)
	// Should error due to invalid padding or other decryption issues
	assert.Error(t, err)
}
