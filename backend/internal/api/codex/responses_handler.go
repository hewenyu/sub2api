package codex

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/middleware"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/relay"
)

// ResponsesHandler handles Codex API responses endpoint.
type ResponsesHandler struct {
	relaySvc relay.CodexRelayService
	logger   *zap.Logger
}

// NewResponsesHandler creates a new responses handler.
func NewResponsesHandler(relaySvc relay.CodexRelayService, logger *zap.Logger) *ResponsesHandler {
	return &ResponsesHandler{
		relaySvc: relaySvc,
		logger:   logger,
	}
}

// HandleResponses handles POST /openai/chat/completions requests.
func (h *ResponsesHandler) HandleResponses(c *gin.Context) {
	h.logger.Info("Received Codex API request",
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("remote_addr", c.ClientIP()),
	)

	// Get authenticated API key from middleware
	apiKey := middleware.GetAPIKey(c)
	if apiKey == nil {
		h.logger.Warn("Unauthorized request - no API key found",
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_addr", c.ClientIP()),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": map[string]any{
				"message": "Unauthorized",
				"type":    "invalid_request_error",
				"code":    "invalid_api_key",
			},
		})
		return
	}

	h.logger.Debug("API key authenticated",
		zap.Int64("api_key_id", apiKey.ID),
		zap.String("api_key_name", apiKey.Name),
	)

	// Parse request body
	var reqBody relay.CodexRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		h.logger.Warn("Invalid request body",
			zap.Error(err),
			zap.Int64("api_key_id", apiKey.ID),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]any{
				"message": "Invalid request body: " + err.Error(),
				"type":    "invalid_request_error",
				"code":    "invalid_request",
			},
		})
		return
	}

	h.logger.Debug("Request body parsed successfully",
		zap.Int64("api_key_id", apiKey.ID),
		zap.String("model", reqBody.Model),
		zap.Bool("stream", reqBody.Stream),
		zap.Int("message_count", len(reqBody.Messages)),
	)

	// Validate model is provided
	if reqBody.Model == "" {
		h.logger.Warn("Missing model in request",
			zap.Int64("api_key_id", apiKey.ID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]any{
				"message": "Model is required",
				"type":    "invalid_request_error",
				"code":    "invalid_request",
			},
		})
		return
	}

	// Validate messages are provided
	if len(reqBody.Messages) == 0 {
		h.logger.Warn("Missing messages in request",
			zap.Int64("api_key_id", apiKey.ID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]any{
				"message": "Messages are required",
				"type":    "invalid_request_error",
				"code":    "invalid_request",
			},
		})
		return
	}

	// Extract request path (remove /openai prefix)
	requestPath := c.Request.URL.Path
	if idx := len("/openai"); len(requestPath) >= idx {
		requestPath = requestPath[idx:]
	}

	h.logger.Info("Delegating to relay service",
		zap.Int64("api_key_id", apiKey.ID),
		zap.String("request_path", requestPath),
		zap.String("model", reqBody.Model),
	)

	// Handle request through relay service (no raw body needed for
	// chat/completions-style Codex requests).
	if err := h.relaySvc.HandleRequest(c, apiKey, &reqBody, nil, requestPath); err != nil {
		if errors.Is(err, context.Canceled) {
			h.logger.Info("Request canceled by client",
				zap.Int64("api_key_id", apiKey.ID),
				zap.String("model", reqBody.Model),
				zap.String("path", c.Request.URL.Path),
			)
			return
		}

		h.logger.Error("Failed to handle request",
			zap.Error(err),
			zap.Int64("api_key_id", apiKey.ID),
			zap.String("model", reqBody.Model),
			zap.String("path", c.Request.URL.Path),
		)

		// Map error to appropriate HTTP status
		statusCode := http.StatusInternalServerError
		errorType := "api_error"
		errorCode := "internal_error"

		if strings.Contains(err.Error(), "no available accounts") || strings.Contains(err.Error(), "account is at capacity") {
			statusCode = http.StatusServiceUnavailable
			errorType = "service_unavailable"
			errorCode = "no_available_accounts"
		} else if strings.Contains(err.Error(), "rate limit exceeded") {
			statusCode = http.StatusTooManyRequests
			errorType = "rate_limit_exceeded"
			errorCode = "rate_limit"
		} else if strings.Contains(err.Error(), "invalid credentials") {
			statusCode = http.StatusUnauthorized
			errorType = "authentication_error"
			errorCode = "invalid_credentials"
		}

		c.JSON(statusCode, gin.H{
			"error": map[string]any{
				"message": err.Error(),
				"type":    errorType,
				"code":    errorCode,
			},
		})
		return
	}

	h.logger.Info("Request handled successfully",
		zap.Int64("api_key_id", apiKey.ID),
		zap.String("model", reqBody.Model),
	)
}
