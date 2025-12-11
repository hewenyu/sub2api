package relay

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/billing"
)

// handleStreamResponse handles a streaming SSE response.
func (s *codexRelayService) handleStreamResponse(
	upstream *http.Response,
	c *gin.Context,
	apiKey *model.APIKey,
	acct *model.CodexAccount,
	reqBody *CodexRequest,
	requestID string,
) error {
	s.logger.Info("Starting streaming response",
		zap.String("request_id", requestID),
		zap.String("content_type", upstream.Header.Get("Content-Type")),
	)

	// Set SSE response headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Get flusher
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		s.logger.Error("Streaming not supported by response writer",
			zap.String("request_id", requestID),
		)
		return fmt.Errorf("streaming not supported")
	}

	// Track usage data and chunks
	var usageData billing.UsageData
	chunkCount := 0
	var lastChunkSample string
	scanner := bufio.NewScanner(upstream.Body)

	for scanner.Scan() {
		line := scanner.Text()
		chunkCount++

		// Forward line to client
		if _, err := c.Writer.Write([]byte(line + "\n")); err != nil {
			// Treat client disconnects/cancellations as normal termination.
			if errors.Is(err, context.Canceled) || c.Request.Context().Err() == context.Canceled {
				s.logger.Info("Client canceled streaming response",
					zap.String("request_id", requestID),
					zap.Int("chunk_count", chunkCount),
				)
				return context.Canceled
			}

			s.logger.Error("Failed to write to client",
				zap.Error(err),
				zap.String("request_id", requestID),
				zap.Int("chunk_count", chunkCount),
			)
			return err
		}
		flusher.Flush()

		// Log first few chunks at DEBUG level for debugging
		if s.logPayloads && chunkCount <= 3 {
			s.logger.Debug("Streaming response chunk",
				zap.String("request_id", requestID),
				zap.Int("chunk_number", chunkCount),
				zap.String("chunk_data", line),
			)
		}

		// Keep sample of last chunk for logging
		lastChunkSample = line

		// Parse SSE event to extract usage
		if jsonStr, found := strings.CutPrefix(line, "data: "); found {
			if jsonStr == "[DONE]" {
				s.logger.Debug("Received stream completion marker",
					zap.String("request_id", requestID),
					zap.Int("total_chunks", chunkCount),
				)
				continue
			}

			// Try to parse as JSON to extract usage
			var event map[string]any
			if err := json.Unmarshal([]byte(jsonStr), &event); err == nil {
				// ChatGPT Codex Responses API sends usage in different shapes depending
				// on the event type. We support both:
				//  1) event["usage"] = { ... }
				//  2) event["response"].usage = { ... } (for response.completed)
				var usage map[string]any

				if u, ok := event["usage"].(map[string]any); ok {
					usage = u
				} else if resp, ok := event["response"].(map[string]any); ok {
					if u, ok2 := resp["usage"].(map[string]any); ok2 {
						usage = u
					}
				}

				if usage != nil {
					// Prefer standard OpenAI field names when present.
					totalInput := 0
					if inputTokens, ok := usage["input_tokens"].(float64); ok {
						totalInput = int(inputTokens)
					} else if promptTokens, ok := usage["prompt_tokens"].(float64); ok {
						// Fallback to legacy prompt_tokens naming
						totalInput = int(promptTokens)
					}

					outputTokens := 0
					if v, ok := usage["output_tokens"].(float64); ok {
						outputTokens = int(v)
					} else if completionTokens, ok := usage["completion_tokens"].(float64); ok {
						outputTokens = int(completionTokens)
					}

					// Extract cache read tokens from input_tokens_details.cached_tokens
					// or top-level cache_read_input_tokens if present.
					cacheReadTokens := 0
					if details, ok := usage["input_tokens_details"].(map[string]any); ok {
						if v, ok2 := details["cached_tokens"].(float64); ok2 {
							cacheReadTokens = int(v)
						}
					}
					if v, ok := usage["cache_read_input_tokens"].(float64); ok && cacheReadTokens == 0 {
						cacheReadTokens = int(v)
					}

					// Extract cache creation tokens from details / top-level / cache_creation object.
					cacheCreateTokens := 0
					if details, ok := usage["input_tokens_details"].(map[string]any); ok {
						if v, ok2 := details["cache_creation_input_tokens"].(float64); ok2 {
							cacheCreateTokens = int(v)
						} else if v, ok2 := details["cache_creation_tokens"].(float64); ok2 {
							cacheCreateTokens = int(v)
						}
					}
					if v, ok := usage["cache_creation_input_tokens"].(float64); ok && cacheCreateTokens == 0 {
						cacheCreateTokens = int(v)
					} else if v, ok := usage["cache_creation_tokens"].(float64); ok && cacheCreateTokens == 0 {
						cacheCreateTokens = int(v)
					}
					if cc, ok := usage["cache_creation"].(map[string]any); ok {
						if v, ok2 := cc["ephemeral_5m_input_tokens"].(float64); ok2 {
							cacheCreateTokens += int(v)
						}
						if v, ok2 := cc["ephemeral_1h_input_tokens"].(float64); ok2 {
							cacheCreateTokens += int(v)
						}
					}

					// Actual input tokens are total input minus cached read tokens.
					actualInputTokens := totalInput
					if cacheReadTokens > 0 && totalInput > 0 {
						actualInputTokens = totalInput - cacheReadTokens
						if actualInputTokens < 0 {
							actualInputTokens = 0
						}
					}

					totalTokens := 0
					if v, ok := usage["total_tokens"].(float64); ok {
						totalTokens = int(v)
					} else if totalInput > 0 || outputTokens > 0 || cacheCreateTokens > 0 {
						totalTokens = totalInput + outputTokens + cacheCreateTokens
					}

					usageData.InputTokens = actualInputTokens
					usageData.OutputTokens = outputTokens
					usageData.TotalTokens = totalTokens
					usageData.CacheCreateTokens = cacheCreateTokens
					usageData.CacheReadTokens = cacheReadTokens

					s.logger.Debug("Extracted usage data from stream",
						zap.String("request_id", requestID),
						zap.Int("input_tokens", usageData.InputTokens),
						zap.Int("output_tokens", usageData.OutputTokens),
						zap.Int("total_tokens", usageData.TotalTokens),
						zap.Int("cache_create_tokens", usageData.CacheCreateTokens),
						zap.Int("cache_read_tokens", usageData.CacheReadTokens),
					)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		if errors.Is(err, context.Canceled) || c.Request.Context().Err() == context.Canceled {
			s.logger.Info("Client canceled streaming while reading",
				zap.String("request_id", requestID),
				zap.Int("chunks_processed", chunkCount),
			)
			return context.Canceled
		}

		s.logger.Error("Error reading stream",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.Int("chunks_processed", chunkCount),
		)
		return err
	}

	s.logger.Info("Streaming response completed",
		zap.String("request_id", requestID),
		zap.Int("total_chunks", chunkCount),
		zap.Int("total_tokens", usageData.TotalTokens),
	)

	if s.logPayloads {
		s.logger.Debug("Last chunk sample",
			zap.String("request_id", requestID),
			zap.String("last_chunk", lastChunkSample),
		)
	}

	// Record usage if available
	if usageData.TotalTokens > 0 {
		s.recordUsage(c.Request.Context(), apiKey, acct, reqBody, &usageData, requestID)
	} else {
		s.logger.Warn("No usage data captured from stream",
			zap.String("request_id", requestID),
		)
	}

	return nil
}

// handleNonStreamResponse handles a non-streaming JSON response.
func (s *codexRelayService) handleNonStreamResponse(
	upstream *http.Response,
	c *gin.Context,
	apiKey *model.APIKey,
	acct *model.CodexAccount,
	reqBody *CodexRequest,
	requestID string,
) error {
	s.logger.Info("Processing non-streaming response",
		zap.String("request_id", requestID),
		zap.Int("status_code", upstream.StatusCode),
		zap.String("content_type", upstream.Header.Get("Content-Type")),
	)

	// Read response body
	bodyBytes, err := io.ReadAll(upstream.Body)
	if err != nil {
		s.logger.Error("Failed to read response body",
			zap.Error(err),
			zap.String("request_id", requestID),
		)
		return fmt.Errorf("failed to read response: %w", err)
	}

	if s.logPayloads {
		s.logger.Debug("Response body received",
			zap.String("request_id", requestID),
			zap.Int("body_size", len(bodyBytes)),
			zap.String("body_preview", string(bodyBytes[:min(len(bodyBytes), 200)])),
		)
	} else {
		s.logger.Debug("Response body received",
			zap.String("request_id", requestID),
			zap.Int("body_size", len(bodyBytes)),
		)
	}

	// Attempt to parse as CodexResponse for usage extraction and logging
	var resp CodexResponse
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		s.logger.Warn("Failed to unmarshal response as CodexResponse; forwarding raw body",
			zap.Error(err),
			zap.String("request_id", requestID),
		)
		// Forward raw JSON response
		c.Data(upstream.StatusCode, upstream.Header.Get("Content-Type"), bodyBytes)
		return nil
	}

	// Log parsed response summary
	if s.logPayloads {
		if respJSON, err := json.Marshal(resp); err == nil {
			s.logger.Debug("Parsed CodexResponse from upstream",
				zap.String("request_id", requestID),
				zap.Int("choices_count", len(resp.Choices)),
				zap.String("response_json", string(respJSON)),
				zap.String("response_id", resp.ID),
				zap.String("model", resp.Model),
			)
		}
	} else {
		s.logger.Debug("Parsed CodexResponse from upstream",
			zap.String("request_id", requestID),
			zap.Int("choices_count", len(resp.Choices)),
			zap.String("response_id", resp.ID),
			zap.String("model", resp.Model),
		)
	}

	// Extract usage
	if resp.Usage != nil {
		// Support both classic chat.completions usage shape and newer
		// Responses API style usage with detailed caching information.
		totalInput := resp.Usage.InputTokens
		if totalInput == 0 {
			totalInput = resp.Usage.PromptTokens
		}

		outputTokens := resp.Usage.OutputTokens
		if outputTokens == 0 {
			outputTokens = resp.Usage.CompletionTokens
		}

		// Extract cache read tokens: prefer explicit cache_read_input_tokens,
		// then fall back to cached_tokens inside *tokens_details.
		cacheReadTokens := resp.Usage.CacheReadInputTokens
		if cacheReadTokens == 0 {
			if details := resp.Usage.InputTokensDetails; details != nil {
				cacheReadTokens = details.CachedTokens
			} else if details := resp.Usage.PromptTokensDetails; details != nil {
				cacheReadTokens = details.CachedTokens
			}
		}

		// Extract cache creation tokens from either top-level or details,
		// and also include any ephemeral cache creation tokens.
		cacheCreateTokens := resp.Usage.CacheCreationInputTokens
		if details := resp.Usage.InputTokensDetails; details != nil {
			if cacheCreateTokens == 0 && details.CacheCreationInputTokens > 0 {
				cacheCreateTokens = details.CacheCreationInputTokens
			}
			if cacheCreateTokens == 0 && details.CacheCreationTokens > 0 {
				cacheCreateTokens = details.CacheCreationTokens
			}
		}
		if cc := resp.Usage.CacheCreation; cc != nil {
			cacheCreateTokens += cc.Ephemeral5mInputTokens + cc.Ephemeral1hInputTokens
		}

		// Actual input tokens are total input minus cached read tokens.
		actualInputTokens := totalInput
		if cacheReadTokens > 0 && totalInput > 0 {
			actualInputTokens = totalInput - cacheReadTokens
			if actualInputTokens < 0 {
				actualInputTokens = 0
			}
		}

		totalTokens := resp.Usage.TotalTokens
		if totalTokens == 0 {
			// Fallback: approximate total tokens as provider-visible tokens.
			totalTokens = totalInput + outputTokens + cacheCreateTokens
		}

		usageData := &billing.UsageData{
			InputTokens:       actualInputTokens,
			OutputTokens:      outputTokens,
			TotalTokens:       totalTokens,
			CacheCreateTokens: cacheCreateTokens,
			CacheReadTokens:   cacheReadTokens,
		}

		s.logger.Info("Token usage from response",
			zap.String("request_id", requestID),
			zap.Int("input_tokens", usageData.InputTokens),
			zap.Int("output_tokens", usageData.OutputTokens),
			zap.Int("total_tokens", usageData.TotalTokens),
			zap.Int("cache_create_tokens", usageData.CacheCreateTokens),
			zap.Int("cache_read_tokens", usageData.CacheReadTokens),
		)

		s.recordUsage(c.Request.Context(), apiKey, acct, reqBody, usageData, requestID)
	} else {
		s.logger.Warn("No usage data in response",
			zap.String("request_id", requestID),
		)
	}

	// Forward response
	c.JSON(upstream.StatusCode, resp)
	return nil
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// recordUsage records usage data to the database.
func (s *codexRelayService) recordUsage(
	ctx context.Context,
	apiKey *model.APIKey,
	acct *model.CodexAccount,
	reqBody *CodexRequest,
	usageData *billing.UsageData,
	requestID string,
) {
	conversationID := ""
	if reqBody.ConversationID != nil {
		conversationID = *reqBody.ConversationID
	}

	if err := s.usageCollector.CollectUsage(
		ctx,
		apiKey.ID,
		acct.ID,
		acct.AccountType,
		usageData,
		reqBody.Model,
		requestID,
		conversationID,
	); err != nil {
		s.logger.Error("Failed to collect usage",
			zap.Error(err),
			zap.String("request_id", requestID),
		)
	}
}
