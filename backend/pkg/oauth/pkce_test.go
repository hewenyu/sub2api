package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateCodeVerifier tests code verifier generation
func TestGenerateCodeVerifier(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Generate valid code verifier"},
		{name: "Generate unique code verifiers"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier, err := GenerateCodeVerifier()
			require.NoError(t, err)
			assert.NotEmpty(t, verifier)

			// Verify hex encoding (128 characters from 64 bytes)
			assert.Equal(t, 128, len(verifier), "Hex-encoded verifier should be 128 characters")

			// Verify length constraints (RFC 7636)
			assert.GreaterOrEqual(t, len(verifier), 43, "Verifier must be at least 43 characters")
			assert.LessOrEqual(t, len(verifier), 128, "Verifier must be at most 128 characters")

			// Verify it's valid according to RFC 7636
			err = ValidateCodeVerifier(verifier)
			assert.NoError(t, err)
		})
	}
}

// TestGenerateCodeVerifier_Uniqueness verifies that generated verifiers are unique
func TestGenerateCodeVerifier_Uniqueness(t *testing.T) {
	verifiers := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		verifier, err := GenerateCodeVerifier()
		require.NoError(t, err)

		// Verify uniqueness
		assert.False(t, verifiers[verifier], "Generated duplicate verifier")
		verifiers[verifier] = true
	}
}

// TestGenerateCodeChallenge tests code challenge generation
func TestGenerateCodeChallenge(t *testing.T) {
	tests := []struct {
		name          string
		verifier      string
		wantChallenge string
	}{
		{
			name:          "Hex-encoded verifier (128 chars)",
			verifier:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			wantChallenge: "syDoWXjbBRNAA6KRTuvd2NO4cmgY8uLGeeGJjHIVYqk",
		},
		{
			name:          "Generate challenge for custom hex verifier",
			verifier:      "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			wantChallenge: "i7zoXAt0q5M5ecthxK9rtAlRC2il04XhqkE1tbDNcYQ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challenge := GenerateCodeChallenge(tt.verifier)
			assert.Equal(t, tt.wantChallenge, challenge)
			assert.NotEmpty(t, challenge)

			// Verify base64url encoding (no padding)
			assert.NotContains(t, challenge, "=")
		})
	}
}

// TestGenerateCodeChallenge_Deterministic verifies that the same verifier produces the same challenge
func TestGenerateCodeChallenge_Deterministic(t *testing.T) {
	verifier := "test_verifier_for_deterministic_check_1234567890"

	challenge1 := GenerateCodeChallenge(verifier)
	challenge2 := GenerateCodeChallenge(verifier)

	assert.Equal(t, challenge1, challenge2, "Same verifier should produce same challenge")
}

// TestValidateCodeVerifier tests code verifier validation
func TestValidateCodeVerifier(t *testing.T) {
	tests := []struct {
		name      string
		verifier  string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Valid verifier (RFC 7636 test vector)",
			verifier:  "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			wantError: false,
		},
		{
			name:      "Valid verifier (minimum length)",
			verifier:  "1234567890123456789012345678901234567890123",
			wantError: false,
		},
		{
			name:      "Valid verifier with allowed characters",
			verifier:  "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~",
			wantError: false,
		},
		{
			name:      "Too short (42 characters)",
			verifier:  "123456789012345678901234567890123456789012",
			wantError: true,
			errorMsg:  "code verifier length must be between 43 and 128 characters",
		},
		{
			name:      "Too long (129 characters)",
			verifier:  "12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901",
			wantError: true,
			errorMsg:  "code verifier length must be between 43 and 128 characters",
		},
		{
			name:      "Invalid character (space)",
			verifier:  "123456789012345678901234567890123456789012 3",
			wantError: true,
			errorMsg:  "code verifier contains invalid character",
		},
		{
			name:      "Invalid character (plus)",
			verifier:  "1234567890123456789012345678901234567890123+",
			wantError: true,
			errorMsg:  "code verifier contains invalid character",
		},
		{
			name:      "Invalid character (slash)",
			verifier:  "1234567890123456789012345678901234567890123/",
			wantError: true,
			errorMsg:  "code verifier contains invalid character",
		},
		{
			name:      "Invalid character (equals)",
			verifier:  "1234567890123456789012345678901234567890123=",
			wantError: true,
			errorMsg:  "code verifier contains invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCodeVerifier(tt.verifier)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPKCE_FullFlow tests the complete PKCE flow
func TestPKCE_FullFlow(t *testing.T) {
	// Step 1: Generate code verifier
	verifier, err := GenerateCodeVerifier()
	require.NoError(t, err)

	// Step 2: Validate verifier
	err = ValidateCodeVerifier(verifier)
	require.NoError(t, err)

	// Step 3: Generate code challenge
	challenge := GenerateCodeChallenge(verifier)
	require.NotEmpty(t, challenge)

	// Step 4: Verify challenge is different from verifier
	assert.NotEqual(t, verifier, challenge, "Challenge should be different from verifier")

	// Step 5: Verify challenge is reproducible
	challenge2 := GenerateCodeChallenge(verifier)
	assert.Equal(t, challenge, challenge2, "Challenge should be deterministic")
}
