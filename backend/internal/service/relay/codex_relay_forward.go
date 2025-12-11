package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

// sanitizeAPIKey returns a prefix of the API key for logging purposes.
// Shows first 8 characters followed by "..." for security.
func sanitizeAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:8] + "..."
}

// sanitizeHeaders returns a copy of headers with sensitive values sanitized.
func sanitizeHeaders(headers http.Header) map[string]string {
	sanitized := make(map[string]string)
	for key, values := range headers {
		if len(values) == 0 {
			continue
		}

		// Sanitize authorization headers
		if strings.ToLower(key) == "authorization" {
			if strings.HasPrefix(values[0], "Bearer ") {
				apiKey := strings.TrimPrefix(values[0], "Bearer ")
				sanitized[key] = "Bearer " + sanitizeAPIKey(apiKey)
			} else {
				sanitized[key] = "***"
			}
		} else {
			sanitized[key] = values[0]
		}
	}
	return sanitized
}

// forwardRequest forwards the request to the upstream API.
func (s *codexRelayService) forwardRequest(
	ctx context.Context,
	c *gin.Context,
	acct *model.CodexAccount,
	apiKey string,
	reqBody *CodexRequest,
	rawBody []byte,
	requestPath string,
	requestID string,
) (*http.Response, error) {
	// Decide whether this is a Responses API style request
	// (/responses or /v1/responses) which should preserve the original
	// JSON body (ResponsesRequest) when talking to upstream.
	isResponsesPath := strings.HasSuffix(requestPath, "/responses")

	var (
		bodyBytes   []byte
		err         error
		upstreamURL string
	)

	if isResponsesPath && len(rawBody) > 0 {
		// For Responses API style calls, use the original request body
		// (ResponsesRequest JSON) so that upstream endpoints like
		// ChatGPT Codex Responses API receive the expected schema.
		bodyBytes = rawBody

		// For OAuth accounts, forward to ChatGPT Codex Responses API.
		// For API-key accounts (openai-responses), forward to their
		// configured base_api + /responses (typically OpenAI v1/responses).
		if acct.AccountType == "openai-oauth" {
			upstreamURL = "https://chatgpt.com/backend-api/codex/responses"
		} else {
			upstreamURL = strings.TrimSuffix(acct.BaseAPI, "/") + requestPath
		}
	} else {
		// Serialize internal CodexRequest body for non-Responses endpoints
		// (e.g. /chat/completions).
		bodyBytes, err = json.Marshal(reqBody)
		if err != nil {
			s.logger.Error("Failed to marshal request body",
				zap.Error(err),
				zap.String("request_id", requestID),
			)
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		// Build upstream URL using dynamic request path
		upstreamURL = strings.TrimSuffix(acct.BaseAPI, "/") + requestPath
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", upstreamURL, bytes.NewReader(bodyBytes))
	if err != nil {
		s.logger.Error("Failed to create HTTP request",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("upstream_url", upstreamURL),
		)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Select HTTP client based on account's proxy configuration
	var httpClient *http.Client
	var proxyName string
	if acct.ProxyName != nil && *acct.ProxyName != "" {
		proxyName = *acct.ProxyName
	}

	client, err := s.clientManager.GetStreamingClient(ctx, proxyName)
	if err != nil {
		s.logger.Warn("Failed to get proxy client, falling back to default",
			zap.String("request_id", requestID),
			zap.String("proxy_name", proxyName),
			zap.Error(err))
		httpClient = s.clientManager.GetDefaultClient(ctx)
	} else {
		httpClient = client
		if proxyName != "" {
			s.logger.Info("Using proxy for upstream request",
				zap.String("request_id", requestID),
				zap.String("proxy_name", proxyName))
		}
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	if isResponsesPath && acct.AccountType == "openai-oauth" {
		// ChatGPT Codex Responses API style:
		// - Bearer token is the OpenAI OAuth access token
		// - Use chatgpt-account-id header for account targeting
		// - Host should be chatgpt.com
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Host", "chatgpt.com")

		if reqBody != nil && reqBody.Stream {
			req.Header.Set("Accept", "text/event-stream")
		} else {
			req.Header.Set("Accept", "application/json")
		}

		// Propagate selected diagnostic headers from the original request,
		// matching the Node implementation's behavior.
		if c != nil && c.Request != nil {
			origHeaders := c.Request.Header
			for _, key := range []string{"openai-beta", "version", "session_id"} {
				if val := origHeaders.Get(key); val != "" {
					req.Header.Set(key, val)
				}
			}

			// Prefer custom user agent from account, otherwise propagate client's UA
			if acct.CustomUserAgent != nil {
				req.Header.Set("User-Agent", *acct.CustomUserAgent)
			} else if ua := origHeaders.Get("User-Agent"); ua != "" {
				req.Header.Set("User-Agent", ua)
			}
		}

		// chatgpt-account-id header derived from ID token metadata when available
		chatgptAccountID := ""
		if acct.ChatGPTAccountID != nil && *acct.ChatGPTAccountID != "" {
			chatgptAccountID = *acct.ChatGPTAccountID
		} else if acct.ChatGPTUserID != nil && *acct.ChatGPTUserID != "" {
			chatgptAccountID = *acct.ChatGPTUserID
		} else {
			chatgptAccountID = strconv.FormatInt(acct.ID, 10)
		}
		req.Header.Set("chatgpt-account-id", chatgptAccountID)
	} else {
		// Standard OpenAI-style API (including openai-responses accounts and
		// non-Responses endpoints).
		req.Header.Set("Authorization", "Bearer "+apiKey)

		if acct.CustomUserAgent != nil {
			req.Header.Set("User-Agent", *acct.CustomUserAgent)
		}
	}

	// Log detailed request information at DEBUG level
	s.logger.Debug("Forwarding request to upstream",
		zap.String("request_id", requestID),
		zap.String("method", "POST"),
		zap.String("upstream_url", upstreamURL),
		zap.Any("headers", sanitizeHeaders(req.Header)),
		zap.String("body", string(bodyBytes)),
		zap.Int("body_size", len(bodyBytes)),
	)

	// Send request
	resp, err := httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to send request to upstream",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("upstream_url", upstreamURL),
		)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Log response headers at DEBUG level
	s.logger.Debug("Received response headers from upstream",
		zap.String("request_id", requestID),
		zap.Int("status_code", resp.StatusCode),
		zap.Any("headers", sanitizeHeaders(resp.Header)),
	)

	return resp, nil
}
