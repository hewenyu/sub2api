package oauth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// IDTokenClaims represents the structure of an OpenAI ID token.
// OpenAI's ID tokens follow the OpenID Connect standard with some custom claims.
type IDTokenClaims struct {
	// Standard OpenID Connect claims
	Email         string      `json:"email"`
	EmailVerified bool        `json:"email_verified"`
	Sub           string      `json:"sub"` // Subject (user ID)
	Iss           string      `json:"iss"` // Issuer
	Aud           interface{} `json:"aud"` // Audience - can be string or []string
	Exp           int64       `json:"exp"` // Expiration time
	Iat           int64       `json:"iat"` // Issued at

	// OpenAI-specific claims
	SubscriptionLevel string `json:"subscription_level,omitempty"` // e.g., "chatgptplusplan"

	// OpenAI auth context (https://api.openai.com/auth)
	AuthContext *OpenAIAuthContext `json:"https://api.openai.com/auth,omitempty"`
}

// GetAudience returns the audience as a string, handling both string and []string formats.
// OpenAI ID tokens may return aud as either a single string or an array of strings.
// If it's an array, returns the first element. Returns empty string if nil or empty.
func (c *IDTokenClaims) GetAudience() string {
	if c.Aud == nil {
		return ""
	}

	switch v := c.Aud.(type) {
	case string:
		return v
	case []interface{}:
		if len(v) > 0 {
			if str, ok := v[0].(string); ok {
				return str
			}
		}
	case []string:
		if len(v) > 0 {
			return v[0]
		}
	}

	return ""
}

// OpenAIAuthContext represents OpenAI-specific auth claims nested under
// the "https://api.openai.com/auth" namespace in the ID token.
//
// This mirrors the Node.js implementation which reads:
//
//	const authClaims = payload['https://api.openai.com/auth'] || {}
//	const accountId = authClaims.chatgpt_account_id || ''
//	const chatgptUserId = authClaims.chatgpt_user_id || authClaims.user_id || ''
//	const planType = authClaims.chatgpt_plan_type || ''
//	const organizations = authClaims.organizations || []
type OpenAIAuthContext struct {
	// ChatGPT account and user identifiers
	ChatGPTAccountID string `json:"chatgpt_account_id,omitempty"`
	ChatGPTUserID    string `json:"chatgpt_user_id,omitempty"`
	UserID           string `json:"user_id,omitempty"`

	// Plan / subscription information
	ChatGPTPlanType string `json:"chatgpt_plan_type,omitempty"` // e.g., "chatgptplusplan"

	// Organizations associated with the user
	Organizations []OpenAIOrganization `json:"organizations,omitempty"`
}

// OpenAIOrganization represents a single organization entry in the
// OpenAI auth context.
type OpenAIOrganization struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Role      string `json:"role,omitempty"`
	Title     string `json:"title,omitempty"`
	IsDefault bool   `json:"is_default,omitempty"`
}

// ParseIDToken parses a JWT ID token without signature verification.
//
// SECURITY NOTE: This function does NOT verify the JWT signature because:
// 1. The ID token is received directly from OpenAI's token endpoint over HTTPS
// 2. We trust the transport layer security (TLS) to ensure authenticity
// 3. This simplifies implementation by avoiding JWKS key fetching and validation
//
// JWT format: header.payload.signature (base64url-encoded, no padding)
//
// This approach is acceptable for server-to-server communication where:
// - The token comes directly from the authorization server over HTTPS
// - We never receive tokens from untrusted sources (e.g., clients)
// - The main goal is to extract user information, not to verify token authenticity
func ParseIDToken(idToken string) (*IDTokenClaims, error) {
	// JWT format: header.payload.signature
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	// Decode payload (part 2) - base64url encoding without padding
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	// Parse JSON payload into claims
	var claims IDTokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWT claims: %w", err)
	}

	return &claims, nil
}

// ExtractUserInfo extracts email and subscription level from an ID token.
// This is a convenience function that wraps ParseIDToken and returns
// the most commonly needed user information.
//
// Returns:
//   - email: User's email address
//   - subscription: Subscription level (e.g., "chatgptplusplan"), empty if not present
//   - error: Parsing error if token is invalid
func ExtractUserInfo(idToken string) (email string, subscription string, err error) {
	if idToken == "" {
		return "", "", fmt.Errorf("ID token is empty")
	}

	claims, err := ParseIDToken(idToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse ID token: %w", err)
	}

	return claims.Email, claims.SubscriptionLevel, nil
}
