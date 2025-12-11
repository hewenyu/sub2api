package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/Wei-Shaw/sub2api/backend/pkg/crypto"
	"github.com/Wei-Shaw/sub2api/backend/pkg/oauth"
)

// ExchangeCodeForTokens exchanges authorization code for tokens.
// DEPRECATED: Use exchangeCodeForTokensWithPKCE for new OAuth flows.
func (s *codexAccountService) ExchangeCodeForTokens(ctx context.Context, code, callbackURL string) (string, string, time.Time, error) {
	config := s.oauthConfig
	if callbackURL != "" {
		tempConfig := *s.oauthConfig
		tempConfig.RedirectURL = callbackURL
		config = &tempConfig
	}

	// Inject HTTP client into context for OAuth2 library to use
	// This ensures the oauth2 package uses our proxy-configured HTTP client
	ctx = context.WithValue(ctx, oauth2.HTTPClient, s.httpClient)

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to exchange code: %w", err)
	}

	refreshToken := token.RefreshToken
	if refreshToken == "" {
		s.logger.Warn("No refresh token received")
	}

	return token.AccessToken, refreshToken, token.Expiry, nil
}

// exchangeCodeForTokensWithPKCE exchanges authorization code for tokens with PKCE verifier.
// This method includes the code_verifier in the token exchange request (RFC 7636).
// It also extracts the ID token from the OAuth response for user info extraction.
func (s *codexAccountService) exchangeCodeForTokensWithPKCE(
	ctx context.Context,
	code, codeVerifier, callbackURL string,
	proxyName *string,
) (accessToken, refreshToken string, expiresAt time.Time, idToken string, err error) {
	config := *s.oauthConfig
	config.RedirectURL = callbackURL

	// Get HTTP client based on proxy configuration
	httpClient, err := s.getHTTPClientByProxyName(ctx, proxyName)
	if err != nil {
		s.logger.Warn("Failed to get proxy HTTP client for PKCE exchange, using default",
			zap.Error(err),
			zap.Stringp("proxy_name", proxyName),
		)
		if s.proxyClientManager != nil {
			httpClient = s.proxyClientManager.GetDefaultClient(ctx)
		} else {
			httpClient = s.httpClient
		}
	}

	// Inject HTTP client into context for OAuth2 library to use
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	// Exchange code with PKCE verifier
	// The code_verifier is passed as an additional parameter per RFC 7636
	token, err := config.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return "", "", time.Time{}, "", fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	refreshToken = token.RefreshToken
	if refreshToken == "" {
		s.logger.Warn("No refresh token received - offline_access scope may be missing")
	}

	// Extract ID token from OAuth response
	// The ID token is a JWT containing user information
	idTokenRaw, ok := token.Extra("id_token").(string)
	if !ok || idTokenRaw == "" {
		s.logger.Warn("No ID token in OAuth response")
	}

	s.logger.Info("Successfully exchanged authorization code for tokens",
		zap.Bool("has_refresh_token", refreshToken != ""),
		zap.Bool("has_id_token", idTokenRaw != ""),
		zap.Time("expires_at", token.Expiry),
		zap.Stringp("proxy_name", proxyName),
	)

	return token.AccessToken, refreshToken, token.Expiry, idTokenRaw, nil
}

// RefreshToken refreshes the OAuth token for an account.
func (s *codexAccountService) RefreshToken(ctx context.Context, accountID int64) error {
	account, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	if account.AccountType != "openai-oauth" {
		return errors.New("account is not an OAuth account")
	}

	if account.RefreshToken == nil {
		return errors.New("no refresh token available")
	}

	// Decrypt refresh token
	refreshToken, err := crypto.AES256Decrypt(*account.RefreshToken, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	// Get HTTP client based on account's proxy configuration
	httpClient, err := s.getHTTPClientByProxyName(ctx, account.ProxyName)
	if err != nil {
		s.logger.Warn("Failed to get proxy client for token refresh, using default",
			zap.Error(err),
			zap.Int64("account_id", accountID),
			zap.Stringp("proxy_name", account.ProxyName),
		)
		if s.proxyClientManager != nil {
			httpClient = s.proxyClientManager.GetDefaultClient(ctx)
		} else {
			httpClient = s.httpClient
		}
	}

	// Inject HTTP client into context for OAuth2 library to use
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	// Create token source
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	// Initialize updates with common fields that we always want to clear on
	// successful refresh (rate limiting status).
	updates := map[string]any{
		// Clear any rate limiting status on successful token refresh
		// This ensures the account is immediately available for requests.
		"rate_limited_until": nil,
		"rate_limit_status":  nil,
	}

	tokenSource := s.oauthConfig.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Attempt to extract and parse ID token from the refresh response, if present.
	// This allows us to (re)populate ChatGPT account identifiers and organization
	// metadata for existing accounts, mirroring the behavior in VerifyAuth.
	if idTokenRaw, ok := newToken.Extra("id_token").(string); ok && idTokenRaw != "" {
		claims, parseErr := oauth.ParseIDToken(idTokenRaw)
		if parseErr != nil {
			s.logger.Warn("Failed to parse ID token during refresh",
				zap.Error(parseErr),
				zap.Int64("account_id", accountID),
			)
		} else {
			// Update email if available
			if claims.Email != "" {
				updates["email"] = claims.Email
			}

			// Update subscription level (plan type)
			if claims.SubscriptionLevel != "" {
				updates["subscription_level"] = claims.SubscriptionLevel
			}

			// Update OpenAI-specific identifiers and organization metadata
			if claims.AuthContext != nil {
				var chatgptAccountID *string
				var chatgptUserID *string

				if claims.AuthContext.ChatGPTAccountID != "" {
					v := claims.AuthContext.ChatGPTAccountID
					chatgptAccountID = &v
				}

				// Prefer chatgpt_user_id, fallback to user_id
				if claims.AuthContext.ChatGPTUserID != "" {
					v := claims.AuthContext.ChatGPTUserID
					chatgptUserID = &v
				} else if claims.AuthContext.UserID != "" {
					v := claims.AuthContext.UserID
					chatgptUserID = &v
				}

				if chatgptAccountID != nil {
					updates["chatgpt_account_id"] = *chatgptAccountID
				}
				if chatgptUserID != nil {
					updates["chatgpt_user_id"] = *chatgptUserID
				}

				// Select default organization if present, otherwise first organization.
				if len(claims.AuthContext.Organizations) > 0 {
					org := claims.AuthContext.Organizations[0]
					for i := range claims.AuthContext.Organizations {
						if claims.AuthContext.Organizations[i].IsDefault {
							org = claims.AuthContext.Organizations[i]
							break
						}
					}

					if org.ID != "" {
						updates["organization_id"] = org.ID
					}
					if org.Role != "" {
						updates["organization_role"] = org.Role
					}
					if org.Title != "" {
						updates["organization_title"] = org.Title
					}
				}
			}

			s.logger.Info("Updated account metadata from refreshed ID token",
				zap.Int64("account_id", accountID),
				zap.Bool("has_auth_context", claims.AuthContext != nil),
			)
		}
	} else {
		s.logger.Info("No ID token present in refresh response; skipping metadata update",
			zap.Int64("account_id", accountID),
		)
	}

	// Encrypt new tokens
	encryptedAccessToken, err := crypto.AES256Encrypt(newToken.AccessToken, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}

	updates["access_token"] = encryptedAccessToken
	updates["expires_at"] = newToken.Expiry

	if newToken.RefreshToken != "" {
		encryptedRefreshToken, err := crypto.AES256Encrypt(newToken.RefreshToken, s.encryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
		updates["refresh_token"] = encryptedRefreshToken
	}

	if err := s.repo.UpdateFields(ctx, accountID, updates); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	s.logger.Info("Refreshed OAuth token and cleared rate limits",
		zap.Int64("account_id", accountID),
		zap.Stringp("proxy_name", account.ProxyName),
	)

	return nil
}

// GetAccountInfo retrieves account information from OpenAI API or ID token.
// This method attempts to parse the input as an ID token (JWT).
// If parsing fails, it returns empty values (OpenAI doesn't have a /v1/me endpoint).
func (s *codexAccountService) GetAccountInfo(ctx context.Context, apiKey string) (string, string, error) {
	// Try to parse as ID token first
	// The apiKey parameter might actually be an ID token during OAuth flow
	email, subscription, err := oauth.ExtractUserInfo(apiKey)
	if err == nil && email != "" {
		return email, subscription, nil
	}

	// If not an ID token or parsing failed, return empty values
	// OpenAI doesn't provide a standard /v1/me endpoint for API key accounts
	return "", "", nil
}
