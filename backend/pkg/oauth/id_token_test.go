package oauth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestIDToken creates a test JWT ID token with the given claims
func createTestIDToken(claims map[string]interface{}) string {
	// Create header (not parsed, but needed for JWT format)
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Create payload
	payloadJSON, _ := json.Marshal(claims)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature (dummy, not verified)
	signature := base64.RawURLEncoding.EncodeToString([]byte("dummy_signature"))

	return headerB64 + "." + payloadB64 + "." + signature
}

// TestParseIDToken tests ID token parsing
func TestParseIDToken(t *testing.T) {
	tests := []struct {
		name          string
		claims        map[string]interface{}
		wantEmail     string
		wantSub       string
		wantError     bool
		errorContains string
	}{
		{
			name: "Valid OpenAI ID token",
			claims: map[string]interface{}{
				"email":          "user@example.com",
				"email_verified": true,
				"sub":            "user_123456",
				"iss":            "https://auth.openai.com",
				"aud":            "app_EMoamEEZ73f0CkXaXp7hrann",
				"exp":            1735689600,
				"iat":            1735686000,
			},
			wantEmail: "user@example.com",
			wantSub:   "user_123456",
			wantError: false,
		},
		{
			name: "ID token with subscription level",
			claims: map[string]interface{}{
				"email":              "premium@example.com",
				"email_verified":     true,
				"subscription_level": "chatgptplusplan",
				"sub":                "user_789",
				"iss":                "https://auth.openai.com",
				"aud":                "app_EMoamEEZ73f0CkXaXp7hrann",
				"exp":                1735689600,
				"iat":                1735686000,
			},
			wantEmail: "premium@example.com",
			wantSub:   "user_789",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idToken := createTestIDToken(tt.claims)

			claims, err := ParseIDToken(idToken)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantEmail, claims.Email)
				assert.Equal(t, tt.wantSub, claims.Sub)

				// Check subscription level if present
				if subLevel, ok := tt.claims["subscription_level"].(string); ok {
					assert.Equal(t, subLevel, claims.SubscriptionLevel)
				}
			}
		})
	}
}

// TestParseIDToken_InvalidFormat tests error handling for invalid JWT formats
func TestParseIDToken_InvalidFormat(t *testing.T) {
	tests := []struct {
		name          string
		idToken       string
		errorContains string
	}{
		{
			name:          "Empty token",
			idToken:       "",
			errorContains: "invalid JWT format",
		},
		{
			name:          "Only one part",
			idToken:       "header",
			errorContains: "invalid JWT format",
		},
		{
			name:          "Only two parts",
			idToken:       "header.payload",
			errorContains: "invalid JWT format",
		},
		{
			name:          "Four parts",
			idToken:       "header.payload.signature.extra",
			errorContains: "invalid JWT format",
		},
		{
			name:          "Invalid base64 in payload",
			idToken:       "eyJhbGciOiJSUzI1NiJ9.invalid!!!base64.signature",
			errorContains: "failed to decode JWT payload",
		},
		{
			name:          "Invalid JSON in payload",
			idToken:       "eyJhbGciOiJSUzI1NiJ9." + base64.RawURLEncoding.EncodeToString([]byte("{invalid json")) + ".signature",
			errorContains: "failed to unmarshal JWT claims",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseIDToken(tt.idToken)

			assert.Error(t, err)
			assert.Nil(t, claims)
			assert.Contains(t, err.Error(), tt.errorContains)
		})
	}
}

// TestExtractUserInfo tests user info extraction from ID token
func TestExtractUserInfo(t *testing.T) {
	tests := []struct {
		name             string
		claims           map[string]interface{}
		wantEmail        string
		wantSubscription string
		wantError        bool
		errorContains    string
	}{
		{
			name: "Extract email and subscription",
			claims: map[string]interface{}{
				"email":              "test@example.com",
				"subscription_level": "chatgptplusplan",
				"sub":                "user_123",
			},
			wantEmail:        "test@example.com",
			wantSubscription: "chatgptplusplan",
			wantError:        false,
		},
		{
			name: "Extract email only (no subscription)",
			claims: map[string]interface{}{
				"email": "free@example.com",
				"sub":   "user_456",
			},
			wantEmail:        "free@example.com",
			wantSubscription: "",
			wantError:        false,
		},
		{
			name: "Empty email",
			claims: map[string]interface{}{
				"email": "",
				"sub":   "user_789",
			},
			wantEmail:        "",
			wantSubscription: "",
			wantError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idToken := createTestIDToken(tt.claims)

			email, subscription, err := ExtractUserInfo(idToken)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantEmail, email)
				assert.Equal(t, tt.wantSubscription, subscription)
			}
		})
	}
}

// TestExtractUserInfo_EmptyToken tests error handling for empty token
func TestExtractUserInfo_EmptyToken(t *testing.T) {
	email, subscription, err := ExtractUserInfo("")

	assert.Error(t, err)
	assert.Empty(t, email)
	assert.Empty(t, subscription)
	assert.Contains(t, err.Error(), "ID token is empty")
}

// TestExtractUserInfo_InvalidToken tests error handling for invalid token
func TestExtractUserInfo_InvalidToken(t *testing.T) {
	email, subscription, err := ExtractUserInfo("invalid.token")

	assert.Error(t, err)
	assert.Empty(t, email)
	assert.Empty(t, subscription)
	assert.Contains(t, err.Error(), "failed to parse ID token")
}

// TestIDToken_RealWorldFormat tests parsing of a realistic ID token format
func TestIDToken_RealWorldFormat(t *testing.T) {
	// Create a realistic OpenAI ID token structure
	claims := map[string]interface{}{
		"email":          "john.doe@company.com",
		"email_verified": true,
		"sub":            "user-abc123def456",
		"iss":            "https://auth.openai.com",
		"aud":            "app_EMoamEEZ73f0CkXaXp7hrann",
		"exp":            1735689600,
		"iat":            1735686000,
		"https://api.openai.com/auth": map[string]interface{}{
			"chatgpt_plan_type":  "chatgptplusplan",
			"chatgpt_account_id": "acct_123",
			"chatgpt_user_id":    "user_456",
			"organizations": []map[string]interface{}{
				{
					"id":         "org_abc",
					"name":       "Test Org",
					"role":       "owner",
					"title":      "Engineer",
					"is_default": true,
				},
			},
		},
	}

	idToken := createTestIDToken(claims)

	// Verify token structure
	parts := strings.Split(idToken, ".")
	require.Len(t, parts, 3, "JWT should have 3 parts")

	// Parse token
	parsedClaims, err := ParseIDToken(idToken)
	require.NoError(t, err)
	assert.Equal(t, "john.doe@company.com", parsedClaims.Email)
	assert.True(t, parsedClaims.EmailVerified)
	assert.Equal(t, "user-abc123def456", parsedClaims.Sub)
	assert.Equal(t, "https://auth.openai.com", parsedClaims.Iss)

	// Verify OpenAI auth context parsing
	require.NotNil(t, parsedClaims.AuthContext)
	assert.Equal(t, "chatgptplusplan", parsedClaims.AuthContext.ChatGPTPlanType)
	assert.Equal(t, "acct_123", parsedClaims.AuthContext.ChatGPTAccountID)
	assert.Equal(t, "user_456", parsedClaims.AuthContext.ChatGPTUserID)
	require.Len(t, parsedClaims.AuthContext.Organizations, 1)
	org := parsedClaims.AuthContext.Organizations[0]
	assert.Equal(t, "org_abc", org.ID)
	assert.Equal(t, "Test Org", org.Name)
	assert.Equal(t, "owner", org.Role)
	assert.Equal(t, "Engineer", org.Title)
	assert.True(t, org.IsDefault)

	// Extract user info
	email, _, err := ExtractUserInfo(idToken)
	require.NoError(t, err)
	assert.Equal(t, "john.doe@company.com", email)
	// Note: subscription_level is at top level in our test, but real OpenAI might nest it
	// Our parser handles both cases
}

// TestIDToken_NoSignatureVerification tests that we don't verify signatures
func TestIDToken_NoSignatureVerification(t *testing.T) {
	claims := map[string]interface{}{
		"email": "test@example.com",
		"sub":   "user_123",
	}

	// Create token with dummy signature
	idToken := createTestIDToken(claims)

	// Should parse successfully even with invalid signature
	// because we trust HTTPS transport security
	parsedClaims, err := ParseIDToken(idToken)
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", parsedClaims.Email)

	// Also verify ExtractUserInfo works
	email, _, err := ExtractUserInfo(idToken)
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", email)
}
