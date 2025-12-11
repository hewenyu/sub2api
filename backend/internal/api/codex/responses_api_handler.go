package codex

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/middleware"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/relay"
)

// codexDefaultInstructions is the default system prompt used when a request
// does not already contain Codex CLI-style instructions. It mirrors the
// behavior of the original Node.js implementation, guiding the model to act
// as the Codex CLI coding assistant.

// ResponsesAPIHandler handles OpenAI Responses API endpoint.
type ResponsesAPIHandler struct {
	relaySvc relay.CodexRelayService
	logger   *zap.Logger
}

// NewResponsesAPIHandler creates a new Responses API handler.
func NewResponsesAPIHandler(relaySvc relay.CodexRelayService, logger *zap.Logger) *ResponsesAPIHandler {
	return &ResponsesAPIHandler{
		relaySvc: relaySvc,
		logger:   logger,
	}
}

// HandleResponsesAPI handles POST /openai/responses and /openai/v1/responses requests.
func (h *ResponsesAPIHandler) HandleResponsesAPI(c *gin.Context) {
	h.logger.Info("Received Responses API request",
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

	// Read raw body so we can both:
	// 1) Parse into a typed ResponsesRequest for internal use, and
	// 2) Preserve all original fields (including unknown ones like tools[*].name)
	//    when forwarding to upstream. Using a generic map for the latter ensures
	//    we don't accidentally drop fields during struct round-tripping.
	rawBodyBytes, err := c.GetRawData()
	if err != nil {
		h.logger.Error("Failed to read request body",
			zap.Error(err),
			zap.Int64("api_key_id", apiKey.ID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]any{
				"message": "Failed to read request body: " + err.Error(),
				"type":    "invalid_request_error",
				"code":    "invalid_request",
			},
		})
		return
	}

	// Parse request body as ResponsesRequest (typed struct for internal logic)
	var reqBody relay.ResponsesRequest
	if unmarshalErr := json.Unmarshal(rawBodyBytes, &reqBody); unmarshalErr != nil {
		h.logger.Warn("Invalid request body",
			zap.Error(unmarshalErr),
			zap.Int64("api_key_id", apiKey.ID),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]any{
				"message": "Invalid request body: " + unmarshalErr.Error(),
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
	)

	// Also parse into a generic map to preserve all original fields
	// (including tools[*].name, custom metadata, etc.) when forwarding
	// to upstream.
	var rawMap map[string]any
	if mapErr := json.Unmarshal(rawBodyBytes, &rawMap); mapErr != nil {
		h.logger.Warn("Failed to parse request body into generic map; falling back to struct-only forwarding",
			zap.Error(mapErr),
			zap.Int64("api_key_id", apiKey.ID),
		)
		// rawMap will remain nil; we'll fall back to marshaling reqBody later.
	}

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

	// Validate input is provided
	if reqBody.Input == nil {
		h.logger.Warn("Missing input in request",
			zap.Int64("api_key_id", apiKey.ID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]any{
				"message": "Input is required",
				"type":    "invalid_request_error",
				"code":    "invalid_request",
			},
		})
		return
	}

	// Detect Codex CLI client by User-Agent. For official Codex CLI traffic,
	// we avoid modifying instructions at all and let the upstream service
	// apply its own system prompt and validation rules.
	userAgent := c.Request.UserAgent()
	isCodexCLIClient := strings.Contains(userAgent, "codex_cli_rs/")

	// Normalize model name for Codex compatibility:
	// If the model starts with "gpt-5-" (e.g. "gpt-5-2025-08-07") and is not
	// the special "gpt-5-codex", normalize it to "gpt-5" as in the original
	// Node.js implementation. This applies to all callers.
	if strings.HasPrefix(reqBody.Model, "gpt-5-") && reqBody.Model != "gpt-5-codex" {
		h.logger.Info("Normalizing model for Codex API",
			zap.Int64("api_key_id", apiKey.ID),
			zap.String("original_model", reqBody.Model),
			zap.String("normalized_model", "gpt-5"),
		)
		reqBody.Model = "gpt-5"
		if rawMap != nil {
			rawMap["model"] = "gpt-5"
		}
	}

	if isCodexCLIClient {
		h.logger.Info("Codex CLI client detected by User-Agent, not modifying instructions",
			zap.Int64("api_key_id", apiKey.ID),
			zap.String("user_agent", userAgent),
		)
	} else {
		// For non-CLI clients, detect whether this is already a Codex CLI-style
		// request based on the instructions prefix. If not, apply the Codex CLI
		// adaptation by clearing certain tuning fields and injecting the
		// default instructions.
		isCodexCLI := false
		if reqBody.Instructions != nil {
			instr := *reqBody.Instructions
			if strings.HasPrefix(instr, "You are a coding agent running in the Codex CLI") ||
				strings.HasPrefix(instr, "You are Codex") {
				isCodexCLI = true
			}
		}

		if !isCodexCLI {
			// Clear optional tuning fields so the Codex CLI defaults apply.
			reqBody.Temperature = nil
			reqBody.TopP = nil
			reqBody.MaxOutputTokens = nil

			if rawMap != nil {
				// Remove or null out optional tuning fields in the raw map as well
				delete(rawMap, "temperature")
				delete(rawMap, "top_p")
				delete(rawMap, "max_output_tokens")
			}

			// Override instructions with the Codex CLI system prompt.
			instr := codexDefaultInstructions
			reqBody.Instructions = &instr
			if rawMap != nil {
				rawMap["instructions"] = instr
			}

			h.logger.Info("Non-Codex CLI request detected, applying Codex CLI adaptation",
				zap.Int64("api_key_id", apiKey.ID),
				zap.String("path", c.Request.URL.Path),
			)
		} else {
			h.logger.Info("Codex CLI-style instructions detected in non-CLI request, forwarding as-is",
				zap.Int64("api_key_id", apiKey.ID),
			)
		}
	}

	// Convert Responses API request to internal CodexRequest format
	codexReq, err := h.convertToCodexRequest(&reqBody)
	if err != nil {
		h.logger.Error("Failed to convert request",
			zap.Error(err),
			zap.Int64("api_key_id", apiKey.ID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]any{
				"message": "Failed to convert request: " + err.Error(),
				"type":    "invalid_request_error",
				"code":    "invalid_request",
			},
		})
		return
	}

	// Re-marshal the original ResponsesRequest body so we can forward it
	// to upstream endpoints that expect the official Responses API schema
	// (e.g. ChatGPT Codex Responses API and OpenAI v1/responses).
	//
	// IMPORTANT: For Codex CLI traffic, we must preserve all original tool
	// definitions and unknown fields exactly as sent by the CLI, otherwise
	// upstream may reject the request (e.g. "Missing required parameter:
	// 'tools[0].name'"). Using the generic rawMap (when available) ensures
	// we don't drop fields that are not modeled in ResponsesRequest.
	var rawBody []byte
	if rawMap != nil {
		rawBody, err = json.Marshal(rawMap)
		if err != nil {
			h.logger.Error("Failed to marshal raw request map",
				zap.Error(err),
				zap.Int64("api_key_id", apiKey.ID),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": map[string]any{
					"message": "Failed to marshal request body",
					"type":    "api_error",
					"code":    "internal_error",
				},
			})
			return
		}
	} else {
		rawBody, err = json.Marshal(reqBody)
		if err != nil {
			h.logger.Error("Failed to marshal request struct",
				zap.Error(err),
				zap.Int64("api_key_id", apiKey.ID),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": map[string]any{
					"message": "Failed to marshal request body",
					"type":    "api_error",
					"code":    "internal_error",
				},
			})
			return
		}
	}

	// Extract request path (remove /openai prefix)
	requestPath := c.Request.URL.Path
	if idx := len("/openai"); len(requestPath) >= idx {
		requestPath = requestPath[idx:]
	}

	h.logger.Info("Delegating to relay service",
		zap.Int64("api_key_id", apiKey.ID),
		zap.String("request_path", requestPath),
		zap.String("model", codexReq.Model),
	)

	// Handle request through relay service
	if err := h.relaySvc.HandleRequest(c, apiKey, codexReq, rawBody, requestPath); err != nil {
		if errors.Is(err, context.Canceled) {
			h.logger.Info("Request canceled by client",
				zap.Int64("api_key_id", apiKey.ID),
				zap.String("model", codexReq.Model),
				zap.String("path", c.Request.URL.Path),
			)
			return
		}

		h.logger.Error("Failed to handle request",
			zap.Error(err),
			zap.Int64("api_key_id", apiKey.ID),
			zap.String("model", codexReq.Model),
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
		} else if strings.Contains(err.Error(), "credentials refreshed") {
			// Credentials were refreshed in the background but this request failed.
			// Communicate a temporary service issue to the client instead of a generic 500.
			statusCode = http.StatusServiceUnavailable
			errorType = "service_unavailable"
			errorCode = "auth_refresh_in_progress"
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
		zap.String("model", codexReq.Model),
	)
}

// convertToCodexRequest converts a ResponsesRequest to a CodexRequest.
func (h *ResponsesAPIHandler) convertToCodexRequest(req *relay.ResponsesRequest) (*relay.CodexRequest, error) {
	codexReq := &relay.CodexRequest{
		Model:  req.Model,
		Stream: req.Stream,
	}

	// Convert temperature if provided
	if req.Temperature != nil {
		codexReq.Temperature = req.Temperature
	}

	// Convert max_output_tokens to max_tokens if provided
	if req.MaxOutputTokens != nil {
		codexReq.MaxTokens = req.MaxOutputTokens
	}

	// Convert input to messages format
	switch input := req.Input.(type) {
	case string:
		// Simple string input - create a user message
		codexReq.Messages = []relay.Message{
			{
				Role:    "user",
				Content: input,
			},
		}
		// Add instructions as system message if provided
		if req.Instructions != nil && *req.Instructions != "" {
			codexReq.Messages = append([]relay.Message{
				{
					Role:    "system",
					Content: *req.Instructions,
				},
			}, codexReq.Messages...)
		}

	case []any:
		// Array of message objects
		messages := make([]relay.Message, 0, len(input))
		for _, msg := range input {
			msgMap, ok := msg.(map[string]any)
			if !ok {
				h.logger.Warn("Invalid message format in input array")
				continue
			}

			role, _ := msgMap["role"].(string)
			content, _ := msgMap["content"].(string)

			if role != "" && content != "" {
				messages = append(messages, relay.Message{
					Role:    role,
					Content: content,
				})
			}
		}
		codexReq.Messages = messages

		// Add instructions as system message if provided and not already present
		if req.Instructions != nil && *req.Instructions != "" {
			hasSystem := false
			for _, msg := range messages {
				if msg.Role == "system" {
					hasSystem = true
					break
				}
			}
			if !hasSystem {
				codexReq.Messages = append([]relay.Message{
					{
						Role:    "system",
						Content: *req.Instructions,
					},
				}, codexReq.Messages...)
			}
		}

	default:
		// Try to unmarshal as JSON to handle complex structures
		inputJSON, err := json.Marshal(req.Input)
		if err != nil {
			return nil, err
		}

		// Try as array of messages first
		var messages []relay.Message
		if err := json.Unmarshal(inputJSON, &messages); err == nil {
			codexReq.Messages = messages
		} else {
			// Fallback: convert to string
			codexReq.Messages = []relay.Message{
				{
					Role:    "user",
					Content: string(inputJSON),
				},
			}
		}
	}

	return codexReq, nil
}
