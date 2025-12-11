package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// handleUpstreamError handles errors from upstream API.
func (s *codexRelayService) handleUpstreamError(
	ctx context.Context,
	err error,
	resp *http.Response,
	accountID int64,
	sessionHash string,
) error {
	if resp == nil {
		return err
	}

	// For client errors (4xx) that are not handled in dedicated branches
	// below, log the upstream response body to aid debugging (HTTP 400 is
	// especially helpful when integrating with the ChatGPT Codex API).
	if resp.StatusCode >= 400 && resp.StatusCode < 500 &&
		resp.StatusCode != http.StatusTooManyRequests && // 429 handled separately
		resp.StatusCode != http.StatusUnauthorized { // 401 has its own logic

		if resp.Body != nil {
			bodyBytes, readErr := io.ReadAll(resp.Body)
			if readErr == nil {
				bodyStr := string(bodyBytes)
				const maxLogBody = 2000
				if len(bodyStr) > maxLogBody {
					bodyStr = bodyStr[:maxLogBody] + "..."
				}

				s.logger.Error("Upstream client error response body",
					zap.Int("status_code", resp.StatusCode),
					zap.Int64("account_id", accountID),
					zap.String("body", bodyStr),
				)

				// Restore body so it can be read again if needed.
				resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			} else {
				s.logger.Error("Failed to read upstream client error body",
					zap.Error(readErr),
					zap.Int("status_code", resp.StatusCode),
					zap.Int64("account_id", accountID),
				)
			}
		}
	}

	switch resp.StatusCode {
	case http.StatusTooManyRequests: // 429 Rate Limit
		// Parse rate limit duration from response (similar to Node.js implementation)
		rateLimitDuration := 5 * time.Minute // default fallback

		// First, check Retry-After header (HTTP standard)
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, parseErr := strconv.Atoi(retryAfter); parseErr == nil && seconds > 0 {
				rateLimitDuration = time.Duration(seconds) * time.Second
				s.logger.Info("Using Retry-After header for rate limit duration",
					zap.Int("seconds", seconds),
					zap.Int64("account_id", accountID),
				)
			}
		}

		// Second, check response body for resets_in_seconds (Claude API specific)
		if resp.Body != nil {
			bodyBytes, readErr := io.ReadAll(resp.Body)
			if readErr == nil {
				// Try to parse error response
				var errorResp struct {
					Error struct {
						Type            string `json:"type"`
						Message         string `json:"message"`
						ResetsInSeconds int    `json:"resets_in_seconds"`
					} `json:"error"`
				}
				if json.Unmarshal(bodyBytes, &errorResp) == nil && errorResp.Error.ResetsInSeconds > 0 {
					rateLimitDuration = time.Duration(errorResp.Error.ResetsInSeconds) * time.Second
					s.logger.Info("Using resets_in_seconds from response body",
						zap.Int("seconds", errorResp.Error.ResetsInSeconds),
						zap.Int64("account_id", accountID),
					)
				}
				// Restore body for potential further reading
				resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		s.logger.Warn("Account rate limited",
			zap.Int64("account_id", accountID),
			zap.Int("status_code", resp.StatusCode),
			zap.Duration("duration", rateLimitDuration),
		)

		// Mark account as unavailable with dynamic duration
		if markErr := s.schedulerSvc.MarkAccountUnavailable(ctx, accountID, "rate_limited", rateLimitDuration); markErr != nil {
			s.logger.Error("Failed to mark account as unavailable",
				zap.Int64("account_id", accountID),
				zap.Error(markErr),
			)
		}

		// Clear sticky session
		if sessionHash != "" {
			if clearErr := s.schedulerSvc.ClearSessionMapping(ctx, sessionHash); clearErr != nil {
				s.logger.Warn("Failed to clear session mapping",
					zap.String("session_hash", sessionHash),
					zap.Error(clearErr),
				)
			}
		}

		return fmt.Errorf("rate limit exceeded")

	case http.StatusUnauthorized: // 401 Unauthorized
		s.logger.Warn("Account unauthorized, checking if token refresh is possible",
			zap.Int64("account_id", accountID),
			zap.Int("status_code", resp.StatusCode),
		)

		// Try to refresh token for OAuth accounts before marking unavailable.
		// However, only refresh when the token is close to expiry. If the token
		// is still valid for a long time, a 401 likely indicates other issues
		// (revoked credentials, permission changes, etc.), and we treat it as
		// invalid credentials instead of attempting an unnecessary refresh.
		acct, getErr := s.accountSvc.GetAccount(ctx, accountID)
		if getErr == nil && acct.AccountType == "openai-oauth" && acct.RefreshToken != nil {
			shouldRefresh := false
			if acct.ExpiresAt != nil {
				timeUntilExpiry := time.Until(*acct.ExpiresAt)
				if timeUntilExpiry <= oauthRefreshThreshold {
					shouldRefresh = true
					s.logger.Info("OAuth token near expiry, attempting refresh after 401",
						zap.Int64("account_id", accountID),
						zap.Duration("time_until_expiry", timeUntilExpiry),
					)
				} else {
					s.logger.Info("Skipping OAuth token refresh after 401; expiry is far in the future",
						zap.Int64("account_id", accountID),
						zap.Duration("time_until_expiry", timeUntilExpiry),
					)
				}
			} else {
				// No expiry information; fall back to allowing a refresh attempt.
				shouldRefresh = true
				s.logger.Info("No ExpiresAt set for OAuth token; allowing refresh after 401",
					zap.Int64("account_id", accountID),
				)
			}

			if shouldRefresh {
				// Use a separate background context for token refresh to ensure it completes
				// even if the client cancels the request
				refreshCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				refreshErr := s.accountSvc.RefreshToken(refreshCtx, accountID)
				if refreshErr == nil {
					s.logger.Info("Successfully refreshed OAuth token, account remains available",
						zap.Int64("account_id", accountID),
					)
					// Do NOT mark account as unavailable after successful refresh.
					// The account should remain available immediately for subsequent requests.
					// Signal to the caller that a retry with refreshed credentials is possible.
					return newRetryableError("credentials refreshed, retry available")
				}

				s.logger.Error("Failed to refresh OAuth token",
					zap.Error(refreshErr),
					zap.Int64("account_id", accountID),
				)
			}
		}

		// Token refresh not available or failed, mark as unauthorized for longer period
		s.logger.Error("Account credentials invalid and cannot be refreshed",
			zap.Int64("account_id", accountID),
		)
		if markErr := s.schedulerSvc.MarkAccountUnavailable(ctx, accountID, "unauthorized", 24*time.Hour); markErr != nil {
			s.logger.Error("Failed to mark account as unavailable",
				zap.Int64("account_id", accountID),
				zap.Error(markErr),
			)
		}

		return fmt.Errorf("invalid credentials")

	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable: // 500, 502, 503
		s.logger.Warn("Upstream server error",
			zap.Int64("account_id", accountID),
			zap.Int("status_code", resp.StatusCode),
		)

		return fmt.Errorf("upstream server error: %d", resp.StatusCode)

	default:
		return err
	}
}
