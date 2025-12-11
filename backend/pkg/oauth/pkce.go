package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const (
	// codeVerifierLength is the length in bytes for the code verifier
	// RFC 7636 allows 43-128 characters, we use 64 bytes which produces 128 characters after hex encoding
	codeVerifierLength = 64
)

// GenerateCodeVerifier generates a cryptographically random code verifier for PKCE.
// RFC 7636: code_verifier = high-entropy cryptographic random STRING using the
// unreserved characters [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
// with a minimum length of 43 characters and a maximum length of 128 characters.
//
// This function generates 64 random bytes and encodes them using hex encoding
// to match Node.js implementation, resulting in a 128-character string.
func GenerateCodeVerifier() (string, error) {
	// Generate cryptographically secure random bytes
	bytes := make([]byte, codeVerifierLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Hex encoding to match Node.js implementation (128 chars)
	codeVerifier := hex.EncodeToString(bytes)

	return codeVerifier, nil
}

// GenerateCodeChallenge computes the SHA-256 hash of the code verifier
// and returns it as a base64url-encoded string.
// RFC 7636: code_challenge = BASE64URL(SHA256(ASCII(code_verifier)))
//
// This is the S256 (SHA-256) challenge method required by OpenAI's OAuth implementation.
func GenerateCodeChallenge(codeVerifier string) string {
	// Compute SHA-256 hash of the code verifier
	hash := sha256.Sum256([]byte(codeVerifier))

	// Base64 URL encoding without padding
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return codeChallenge
}

// ValidateCodeVerifier checks if a code verifier meets RFC 7636 requirements:
// - Length must be between 43 and 128 characters
// - Must only contain unreserved characters: [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
//
// This validation is useful for testing and debugging PKCE implementations.
func ValidateCodeVerifier(codeVerifier string) error {
	// Check length constraints
	if len(codeVerifier) < 43 || len(codeVerifier) > 128 {
		return fmt.Errorf("code verifier length must be between 43 and 128 characters, got %d", len(codeVerifier))
	}

	// Check allowed characters: [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
	for _, c := range codeVerifier {
		if !isUnreservedCharacter(c) {
			return fmt.Errorf("code verifier contains invalid character: %c", c)
		}
	}

	return nil
}

// isUnreservedCharacter checks if a character is an unreserved character as defined by RFC 7636.
// Unreserved characters: [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
func isUnreservedCharacter(c rune) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '.' || c == '_' || c == '~'
}
