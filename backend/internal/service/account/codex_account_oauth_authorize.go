package account

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	redisrepo "github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
	"github.com/Wei-Shaw/sub2api/backend/pkg/crypto"
	"github.com/Wei-Shaw/sub2api/backend/pkg/oauth"
)

// GenerateAuthURL generates an OAuth authorization URL with PKCE and state storage.
// This method implements:
// 1. CSRF protection via random state parameter
// 2. PKCE (RFC 7636) for authorization code interception prevention
// 3. State storage in Redis for validation during callback
func (s *codexAccountService) GenerateAuthURL(ctx context.Context, callbackPort int) (string, string, string, error) {
	// Generate random state (CSRF token)
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", "", fmt.Errorf("failed to generate state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	// Generate PKCE code verifier (RFC 7636)
	codeVerifier, err := oauth.GenerateCodeVerifier()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate PKCE code verifier: %w", err)
	}

	// Compute PKCE code challenge (SHA-256)
	codeChallenge := oauth.GenerateCodeChallenge(codeVerifier)

	// For Codex OAuth we must use the official Codex CLI redirect URI
	// to match the registered configuration on auth.openai.com.
	// This value is fixed and must NOT be changed unless you register
	// your own OAuth client with a different redirect URI.
	const codexRedirectURI = "http://localhost:1455/auth/callback"
	callbackURL := codexRedirectURI

	// Store state + PKCE verifier in Redis (10 min TTL)
	stateData := redisrepo.OAuthStateData{
		CodeVerifier: codeVerifier,
		CallbackURL:  callbackURL,
		CreatedAt:    time.Now().Unix(),
	}
	if err := s.oauthStateRepo.StoreState(ctx, state, stateData, 10*time.Minute); err != nil {
		return "", "", "", fmt.Errorf("failed to store OAuth state: %w", err)
	}

	// Manually construct the authorization URL to match Node.js parameter ordering
	// OpenAI's OAuth implementation is strict about:
	// 1. Parameter order (alphabetical sorting causes failures)
	// 2. Scope encoding: MUST use '+' for spaces (application/x-www-form-urlencoded format)
	//    NOT '%20' (standard URL encoding) - this is critical for OpenAI OAuth
	//
	// Parameter order must match: response_type -> client_id -> redirect_uri -> scope ->
	// code_challenge -> code_challenge_method -> state -> id_token_add_organizations -> codex_cli_simplified_flow
	//
	// IMPORTANT: scope uses '+' directly instead of url.QueryEscape to match
	// Node.js URLSearchParams behavior (application/x-www-form-urlencoded)
	scopeValue := strings.Join(s.oauthConfig.Scopes, "+") // Use '+' for spaces (form-urlencoded)

	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&code_challenge=%s&code_challenge_method=S256&state=%s&id_token_add_organizations=true&codex_cli_simplified_flow=true",
		s.oauthConfig.Endpoint.AuthURL,
		url.QueryEscape(s.oauthConfig.ClientID),
		url.QueryEscape(callbackURL),
		scopeValue, // Don't escape! Already formatted with '+' for form-urlencoded
		url.QueryEscape(codeChallenge),
		url.QueryEscape(state),
	)

	s.logger.Info("Generated OAuth auth URL with PKCE (manual construction for parameter order)",
		zap.String("state", state),
		zap.String("callback_url", callbackURL),
		zap.String("code_challenge", codeChallenge),
	)

	return authURL, callbackURL, state, nil
}

// VerifyAuth verifies OAuth authorization and creates an account.
// This method implements:
// 1. State validation (CSRF protection)
// 2. PKCE verification (RFC 7636)
// 3. ID token parsing for user info
func (s *codexAccountService) VerifyAuth(ctx context.Context, code, state string, accountData CreateCodexAccountRequest) (*model.CodexAccount, error) {
	// Input validation: ensure code and state are not empty
	if code == "" {
		return nil, errors.New("authorization code is required")
	}
	if state == "" {
		return nil, errors.New("state parameter is required")
	}

	// Validate account data
	if accountData.Name == "" {
		return nil, errors.New("account name is required")
	}
	if accountData.AccountType == "" {
		return nil, errors.New("account_type is required and must be 'openai-oauth' or 'openai-responses'")
	}
	if accountData.AccountType != "openai-oauth" && accountData.AccountType != "openai-responses" {
		return nil, fmt.Errorf("invalid account_type '%s', must be 'openai-oauth' or 'openai-responses'", accountData.AccountType)
	}

	s.logger.Info("Verifying OAuth authorization",
		zap.String("account_name", accountData.Name),
		zap.String("account_type", accountData.AccountType),
		zap.Int("code_length", len(code)),
		zap.Int("state_length", len(state)),
	)

	// Consume state from Redis (atomic get + delete)
	// This ensures the state can only be used once (replay attack prevention)
	stateData, err := s.oauthStateRepo.ConsumeState(ctx, state)
	if err != nil {
		s.logger.Warn("Failed to retrieve OAuth state from Redis",
			zap.Error(err),
			zap.String("state", state),
		)
		return nil, fmt.Errorf("failed to retrieve OAuth state: %w", err)
	}
	if stateData == nil {
		s.logger.Warn("OAuth state not found or already consumed",
			zap.String("state", state),
		)
		return nil, errors.New("invalid or expired OAuth state (CSRF protection)")
	}

	// Validate state age (max 10 minutes)
	stateAge := time.Now().Unix() - stateData.CreatedAt
	if stateAge > 600 {
		s.logger.Warn("OAuth state expired",
			zap.Int64("state_age_seconds", stateAge),
			zap.String("state", state),
		)
		return nil, errors.New("OAuth state expired")
	}

	s.logger.Info("OAuth state validated successfully",
		zap.Int64("state_age_seconds", stateAge),
		zap.String("callback_url", stateData.CallbackURL),
	)

	// Exchange authorization code for tokens with PKCE verifier
	accessToken, refreshToken, expiresAt, idToken, err := s.exchangeCodeForTokensWithPKCE(
		ctx,
		code,
		stateData.CodeVerifier,
		stateData.CallbackURL,
		accountData.ProxyName,
	)
	if err != nil {
		s.logger.Error("Failed to exchange authorization code for tokens",
			zap.Error(err),
			zap.String("account_name", accountData.Name),
		)
		return nil, fmt.Errorf("failed to exchange code for tokens: %w", err)
	}

	// Validate token response
	if accessToken == "" {
		s.logger.Error("Received empty access token from OAuth provider")
		return nil, errors.New("OAuth provider returned empty access token")
	}

	s.logger.Info("Successfully exchanged authorization code for tokens",
		zap.Bool("has_refresh_token", refreshToken != ""),
		zap.Bool("has_id_token", idToken != ""),
		zap.Time("expires_at", expiresAt),
	)

	// Parse ID token to extract user and account info
	var (
		email             string
		subscriptionLevel string
		chatgptAccountID  *string
		chatgptUserID     *string
		orgID             *string
		orgRole           *string
		orgTitle          *string
	)

	claims, err := oauth.ParseIDToken(idToken)
	if err != nil {
		s.logger.Warn("Failed to parse ID token, continuing with empty user info",
			zap.Error(err),
		)
	} else {
		email = claims.Email
		subscriptionLevel = claims.SubscriptionLevel

		if claims.AuthContext != nil {
			if claims.AuthContext.ChatGPTAccountID != "" {
				chatgptAccountID = &claims.AuthContext.ChatGPTAccountID
			}

			// Prefer chatgpt_user_id, fallback to user_id
			if claims.AuthContext.ChatGPTUserID != "" {
				chatgptUserID = &claims.AuthContext.ChatGPTUserID
			} else if claims.AuthContext.UserID != "" {
				chatgptUserID = &claims.AuthContext.UserID
			}

			// Extract default organization if present
			if len(claims.AuthContext.Organizations) > 0 {
				var defaultOrg *oauth.OpenAIOrganization
				for i := range claims.AuthContext.Organizations {
					if claims.AuthContext.Organizations[i].IsDefault {
						defaultOrg = &claims.AuthContext.Organizations[i]
						break
					}
				}
				if defaultOrg == nil {
					defaultOrg = &claims.AuthContext.Organizations[0]
				}

				if defaultOrg.ID != "" {
					orgID = &defaultOrg.ID
				}
				if defaultOrg.Role != "" {
					orgRole = &defaultOrg.Role
				}
				if defaultOrg.Title != "" {
					orgTitle = &defaultOrg.Title
				}
			}
		}
	}

	// Encrypt tokens for secure storage
	s.logger.Info("Encrypting OAuth tokens for secure storage")
	encryptedAccessToken, err := crypto.AES256Encrypt(accessToken, s.encryptionKey)
	if err != nil {
		s.logger.Error("Failed to encrypt access token", zap.Error(err))
		return nil, fmt.Errorf("failed to encrypt access token: %w", err)
	}

	var encryptedRefreshToken string
	if refreshToken != "" {
		encryptedRefreshToken, err = crypto.AES256Encrypt(refreshToken, s.encryptionKey)
		if err != nil {
			s.logger.Error("Failed to encrypt refresh token", zap.Error(err))
			return nil, fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
	} else {
		s.logger.Warn("No refresh token received - token refresh will not be available")
	}

	// Create account with validated data
	// Force account type to openai-oauth for OAuth flow
	account := &model.CodexAccount{
		Name:        accountData.Name,
		AccountType: "openai-oauth",
		IsActive:    true,
	}

	// Set tokens
	account.AccessToken = &encryptedAccessToken
	if encryptedRefreshToken != "" {
		account.RefreshToken = &encryptedRefreshToken
	}
	account.ExpiresAt = &expiresAt

	// Set email if available
	if email != "" {
		account.Email = &email
	}

	// Set API configuration with defaults
	if accountData.BaseAPI != "" {
		account.BaseAPI = accountData.BaseAPI
	} else {
		account.BaseAPI = "https://api.openai.com/v1"
	}

	// Set quota configuration with defaults
	account.DailyQuota = accountData.DailyQuota
	if accountData.QuotaResetTime != "" {
		account.QuotaResetTime = accountData.QuotaResetTime
	} else {
		account.QuotaResetTime = "00:00"
	}

	// Set scheduling configuration with defaults
	if accountData.Priority > 0 {
		account.Priority = accountData.Priority
	} else {
		account.Priority = 100
	}
	// OAuth accounts are schedulable by default
	// Since Schedulable is a bool (not *bool), it defaults to false in JSON
	// We need to explicitly set it to true for OAuth accounts
	account.Schedulable = true

	if subscriptionLevel != "" {
		account.SubscriptionLevel = &subscriptionLevel
	}

	// Attach OpenAI-specific identifiers if available
	if chatgptAccountID != nil {
		account.ChatGPTAccountID = chatgptAccountID
	}
	if chatgptUserID != nil {
		account.ChatGPTUserID = chatgptUserID
	}
	if orgID != nil {
		account.OrganizationID = orgID
	}
	if orgRole != nil {
		account.OrganizationRole = orgRole
	}
	if orgTitle != nil {
		account.OrganizationTitle = orgTitle
	}

	if accountData.CustomUserAgent != nil {
		account.CustomUserAgent = accountData.CustomUserAgent
	}

	// Set proxy by name if provided
	if accountData.ProxyName != nil && *accountData.ProxyName != "" {
		account.ProxyName = accountData.ProxyName
		s.logger.Info("Proxy name added to account",
			zap.String("proxy_name", *accountData.ProxyName),
		)
	}

	// Persist account to database
	s.logger.Info("Persisting OAuth account to database",
		zap.String("account_name", account.Name),
		zap.String("account_type", account.AccountType),
	)

	if err := s.repo.Create(ctx, account); err != nil {
		s.logger.Error("Failed to persist account to database",
			zap.Error(err),
			zap.String("account_name", account.Name),
		)
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	s.logger.Info("Codex account created via OAuth successfully",
		zap.Int64("account_id", account.ID),
		zap.String("account_name", account.Name),
		zap.String("email", email),
		zap.String("subscription_level", subscriptionLevel),
		zap.Bool("has_refresh_token", account.RefreshToken != nil),
	)

	return account, nil
}
