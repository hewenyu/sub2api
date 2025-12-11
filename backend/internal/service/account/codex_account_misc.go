package account

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/pkg/crypto"
)

// DecryptAPIKey decrypts the API key for an account.
func (s *codexAccountService) DecryptAPIKey(ctx context.Context, account *model.CodexAccount) (string, error) {
	if account.AccountType == "openai-oauth" {
		if account.AccessToken == nil {
			return "", errors.New("no access token available")
		}
		apiKey, err := crypto.AES256Decrypt(*account.AccessToken, s.encryptionKey)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt access token: %w", err)
		}
		return apiKey, nil
	}

	if account.APIKey == nil {
		return "", errors.New("no API key available")
	}

	apiKey, err := crypto.AES256Decrypt(*account.APIKey, s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt API key: %w", err)
	}

	return apiKey, nil
}

// TestAccount tests if an account is valid by calling the OpenAI API.
func (s *codexAccountService) TestAccount(ctx context.Context, id int64) error {
	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Decrypt API key or access token
	apiKey, err := s.DecryptAPIKey(ctx, account)
	if err != nil {
		return fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	// Build API URL
	baseURL := account.BaseAPI
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	apiURL, err := url.JoinPath(baseURL, "models")
	if err != nil {
		return fmt.Errorf("failed to build API URL: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	if account.CustomUserAgent != nil {
		req.Header.Set("User-Agent", *account.CustomUserAgent)
	}

	// Get appropriate HTTP client (with proxy if configured)
	httpClient, err := s.getHTTPClient(ctx, account)
	if err != nil {
		return fmt.Errorf("failed to get HTTP client: %w", err)
	}

	// Send request
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			s.logger.Warn("Failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
