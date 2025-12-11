package relay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/scheduler"
)

// HandleRequest handles a Codex API request.
func (s *codexRelayService) HandleRequest(c *gin.Context, apiKey *model.APIKey, reqBody *CodexRequest, rawBody []byte, requestPath string) error {
	ctx := c.Request.Context()
	requestID := uuid.New().String()
	startTime := time.Now()

	// Log request summary
	s.logger.Info("Starting Codex relay request",
		zap.String("request_id", requestID),
		zap.Int64("api_key_id", apiKey.ID),
		zap.String("api_key_prefix", apiKey.KeyPrefix),
		zap.String("model", reqBody.Model),
		zap.Bool("stream", reqBody.Stream),
		zap.String("path", requestPath),
		zap.String("method", c.Request.Method),
	)

	// Log detailed request body at DEBUG level
	if s.logPayloads {
		if reqBodyJSON, err := json.Marshal(reqBody); err == nil {
			s.logger.Debug("Request body details",
				zap.String("request_id", requestID),
				zap.String("body", string(reqBodyJSON)),
				zap.Int("message_count", len(reqBody.Messages)),
			)
		}
	}

	// Generate session hash for conversation tracking
	sessionHash := ""
	if reqBody.ConversationID != nil && *reqBody.ConversationID != "" {
		sessionManager := scheduler.NewSessionManager()
		sessionHash = sessionManager.CreateSessionHash(apiKey.ID, *reqBody.ConversationID)
	}

	// Select account
	accountID, _, err := s.schedulerSvc.SelectCodexAccount(ctx, apiKey, sessionHash)
	if err != nil {
		s.logger.Error("Failed to select account",
			zap.Error(err),
			zap.String("request_id", requestID),
		)
		return fmt.Errorf("no available accounts: %w", err)
	}

	// Acquire concurrency slot
	if acquireErr := s.schedulerSvc.AcquireConcurrencySlot(ctx, accountID, requestID, 300); acquireErr != nil {
		s.logger.Error("Failed to acquire concurrency slot",
			zap.Error(acquireErr),
			zap.Int64("account_id", accountID),
		)
		return fmt.Errorf("account is at capacity: %w", acquireErr)
	}
	// Use a background context for deferred cleanup to ensure it completes even if the request is canceled
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if releaseErr := s.schedulerSvc.ReleaseConcurrencySlot(releaseCtx, accountID, requestID); releaseErr != nil {
			s.logger.Error("Failed to release concurrency slot in deferred cleanup",
				zap.Error(releaseErr),
				zap.Int64("account_id", accountID),
				zap.String("request_id", requestID),
			)
		}
	}()

	// Get account
	acct, err := s.accountSvc.GetAccount(ctx, accountID)
	if err != nil {
		s.logger.Error("Failed to get account",
			zap.Error(err),
			zap.Int64("account_id", accountID),
			zap.String("request_id", requestID),
		)
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Log account details
	accountEmail := ""
	if acct.Email != nil {
		accountEmail = *acct.Email
	}
	s.logger.Info("Selected Codex account for request",
		zap.String("request_id", requestID),
		zap.Int64("account_id", acct.ID),
		zap.String("account_type", acct.AccountType),
		zap.String("account_email", accountEmail),
	)

	// Check if OAuth token needs refresh (for openai-oauth accounts)
	if acct.AccountType == "openai-oauth" && acct.ExpiresAt != nil {
		timeUntilExpiry := time.Until(*acct.ExpiresAt)
		if timeUntilExpiry <= oauthRefreshThreshold {
			s.logger.Info("OAuth token within refresh threshold, attempting proactive refresh",
				zap.String("request_id", requestID),
				zap.Int64("account_id", acct.ID),
				zap.Duration("time_until_expiry", timeUntilExpiry),
			)

			// Use a separate background context for token refresh to ensure it completes
			// even if the client cancels the request
			refreshCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Attempt to refresh the token
			if refreshErr := s.accountSvc.RefreshToken(refreshCtx, acct.ID); refreshErr != nil {
				s.logger.Error("Failed to refresh OAuth token",
					zap.Error(refreshErr),
					zap.Int64("account_id", acct.ID),
					zap.String("request_id", requestID),
				)
				// Continue with existing token, it might still work
			} else {
				s.logger.Info("OAuth token refreshed successfully",
					zap.String("request_id", requestID),
					zap.Int64("account_id", acct.ID),
				)
				// Reload account to get fresh credentials
				acct, err = s.accountSvc.GetAccount(ctx, accountID)
				if err != nil {
					s.logger.Error("Failed to reload account after token refresh",
						zap.Error(err),
						zap.Int64("account_id", accountID),
					)
					return fmt.Errorf("failed to reload account: %w", err)
				}
			}
		}
	}

	// Forward request to upstream with at most one automatic retry on successful credential refresh
	var (
		upstreamResp     *http.Response
		apiKeyDecrypted  string
		upstreamDuration time.Duration
	)

	for attempt := 1; attempt <= 2; attempt++ {
		// Decrypt API key / access token for current account state
		apiKeyDecrypted, err = s.accountSvc.DecryptAPIKey(ctx, acct)
		if err != nil {
			s.logger.Error("Failed to decrypt API key",
				zap.Error(err),
				zap.Int64("account_id", accountID),
				zap.String("request_id", requestID),
			)
			return fmt.Errorf("failed to decrypt credentials: %w", err)
		}

		// Forward request to upstream
		upstreamStartTime := time.Now()
		upstreamResp, err = s.forwardRequest(ctx, c, acct, apiKeyDecrypted, reqBody, rawBody, requestPath, requestID)
		upstreamDuration = time.Since(upstreamStartTime)

		if err != nil {
			s.logger.Error("Failed to forward request",
				zap.Error(err),
				zap.String("request_id", requestID),
				zap.Duration("upstream_duration", upstreamDuration),
			)

			handleErr := s.handleUpstreamError(ctx, err, upstreamResp, accountID, sessionHash)
			if upstreamResp != nil {
				_ = upstreamResp.Body.Close() // Ignore error in cleanup path
			}

			if isRetryableError(handleErr) && attempt == 1 {
				s.logger.Info("Retrying upstream request after credential refresh",
					zap.String("request_id", requestID),
					zap.Int("retry_attempt", attempt+1),
				)

				// Reload account with refreshed credentials using a fresh background context.
				// This avoids failures when the original request context has been canceled
				// while the token refresh was in progress.
				reloadCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				acct, err = s.accountSvc.GetAccount(reloadCtx, accountID)
				cancel()
				if err != nil {
					s.logger.Error("Failed to reload account after token refresh",
						zap.Error(err),
						zap.Int64("account_id", accountID),
						zap.String("request_id", requestID),
					)
					return fmt.Errorf("failed to reload account after token refresh: %w", err)
				}
				continue
			}

			return handleErr
		}

		// Ensure we have a non-nil upstream response before dereferencing.
		if upstreamResp == nil {
			s.logger.Error("Received nil upstream response",
				zap.String("request_id", requestID),
				zap.Duration("upstream_duration", upstreamDuration),
			)
			return fmt.Errorf("no successful upstream response after retries")
		}

		// Log upstream response status
		s.logger.Info("Received upstream response",
			zap.String("request_id", requestID),
			zap.Int("status_code", upstreamResp.StatusCode),
			zap.Duration("upstream_duration", upstreamDuration),
		)

		// Check status code
		if upstreamResp.StatusCode != http.StatusOK {
			handleErr := s.handleUpstreamError(
				ctx,
				fmt.Errorf("upstream returned status %d", upstreamResp.StatusCode),
				upstreamResp,
				accountID,
				sessionHash,
			)

			_ = upstreamResp.Body.Close() // Ignore error in retry path

			if isRetryableError(handleErr) && attempt == 1 {
				s.logger.Info("Retrying upstream request after credential refresh",
					zap.String("request_id", requestID),
					zap.Int("retry_attempt", attempt+1),
				)

				// Reload account with refreshed credentials using a fresh background context.
				reloadCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				acct, err = s.accountSvc.GetAccount(reloadCtx, accountID)
				cancel()
				if err != nil {
					s.logger.Error("Failed to reload account after token refresh",
						zap.Error(err),
						zap.Int64("account_id", accountID),
						zap.String("request_id", requestID),
					)
					return fmt.Errorf("failed to reload account after token refresh: %w", err)
				}
				continue
			}

			return handleErr
		}

		// Upstream returned 200 OK; proceed with response handling
		break
	}

	// At this point upstreamResp is guaranteed to be non-nil.
	defer func() {
		_ = upstreamResp.Body.Close() // Ignore error in defer
	}()

	// Handle response based on stream type
	var handleErr error
	if reqBody.Stream {
		handleErr = s.handleStreamResponse(upstreamResp, c, apiKey, acct, reqBody, requestID)
	} else {
		handleErr = s.handleNonStreamResponse(upstreamResp, c, apiKey, acct, reqBody, requestID)
	}

	// Log request completion
	totalDuration := time.Since(startTime)
	if handleErr != nil {
		if errors.Is(handleErr, context.Canceled) {
			s.logger.Info("Codex relay request canceled",
				zap.String("request_id", requestID),
				zap.Duration("total_duration", totalDuration),
			)
		} else {
			s.logger.Error("Codex relay request failed",
				zap.String("request_id", requestID),
				zap.Error(handleErr),
				zap.Duration("total_duration", totalDuration),
			)
		}
	} else {
		s.logger.Info("Codex relay request completed successfully",
			zap.String("request_id", requestID),
			zap.Duration("total_duration", totalDuration),
			zap.Duration("upstream_duration", upstreamDuration),
		)
	}

	return handleErr
}
