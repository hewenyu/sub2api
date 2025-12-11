package codex

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/relay"
)

// mockCodexRelayService is a mock implementation for testing
type mockCodexRelayService struct {
	handleRequestFunc func(c *gin.Context, apiKey *model.APIKey, reqBody *relay.CodexRequest, rawBody []byte, requestPath string) error
}

func (m *mockCodexRelayService) HandleRequest(c *gin.Context, apiKey *model.APIKey, reqBody *relay.CodexRequest, rawBody []byte, requestPath string) error {
	if m.handleRequestFunc != nil {
		return m.handleRequestFunc(c, apiKey, reqBody, rawBody, requestPath)
	}
	return nil
}

func TestResponsesAPIHandler_ValidStringInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockSvc := &mockCodexRelayService{
		handleRequestFunc: func(c *gin.Context, apiKey *model.APIKey, reqBody *relay.CodexRequest, rawBody []byte, requestPath string) error {
			// Verify the request was converted correctly
			assert.Equal(t, "gpt-4o", reqBody.Model)
			// With instructions, we should have 2 messages: system and user
			assert.Equal(t, 2, len(reqBody.Messages))
			assert.Equal(t, "system", reqBody.Messages[0].Role)
			assert.Equal(t, "You are a coding agent running in the Codex CLI", reqBody.Messages[0].Content)
			assert.Equal(t, "user", reqBody.Messages[1].Role)
			assert.Equal(t, "Hello, world!", reqBody.Messages[1].Content)

			// Send a mock response
			c.JSON(http.StatusOK, gin.H{
				"id":      "resp_123",
				"object":  "response",
				"created": 1234567890,
				"model":   "gpt-4o",
				"output":  "Hello! How can I help you?",
			})
			return nil
		},
	}

	handler := NewResponsesAPIHandler(mockSvc, logger)

	// Create test request with string input
	codexInstructions := "You are a coding agent running in the Codex CLI"
	reqBody := map[string]interface{}{
		"model":        "gpt-4o",
		"input":        "Hello, world!",
		"instructions": codexInstructions,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/openai/responses", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Set API key in context (simulating middleware)
	apiKey := &model.APIKey{
		ID:   1,
		Name: "test-key",
	}
	c.Set("api_key", apiKey)

	handler.HandleResponsesAPI(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResponsesAPIHandler_ValidArrayInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockSvc := &mockCodexRelayService{
		handleRequestFunc: func(c *gin.Context, apiKey *model.APIKey, reqBody *relay.CodexRequest, rawBody []byte, requestPath string) error {
			// Verify the request was converted correctly
			assert.Equal(t, "gpt-4o", reqBody.Model)
			assert.Equal(t, 2, len(reqBody.Messages))
			assert.Equal(t, "system", reqBody.Messages[0].Role)
			assert.Equal(t, "You are a helpful assistant.", reqBody.Messages[0].Content)
			assert.Equal(t, "user", reqBody.Messages[1].Role)
			assert.Equal(t, "What is the weather?", reqBody.Messages[1].Content)

			c.JSON(http.StatusOK, gin.H{
				"id":     "resp_456",
				"object": "response",
				"output": "I don't have access to real-time weather.",
			})
			return nil
		},
	}

	handler := NewResponsesAPIHandler(mockSvc, logger)

	// Create test request with array input
	codexInstructions := "You are a coding agent running in the Codex CLI"
	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"input": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant."},
			{"role": "user", "content": "What is the weather?"},
		},
		"instructions": codexInstructions,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/openai/responses", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	apiKey := &model.APIKey{
		ID:   1,
		Name: "test-key",
	}
	c.Set("api_key", apiKey)

	handler.HandleResponsesAPI(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResponsesAPIHandler_MissingModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockSvc := &mockCodexRelayService{}
	handler := NewResponsesAPIHandler(mockSvc, logger)

	// Create test request without model
	reqBody := map[string]interface{}{
		"input": "Hello, world!",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/openai/responses", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	apiKey := &model.APIKey{
		ID:   1,
		Name: "test-key",
	}
	c.Set("api_key", apiKey)

	handler.HandleResponsesAPI(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	errorObj := response["error"].(map[string]interface{})
	assert.Contains(t, errorObj["message"], "Model")
}

func TestResponsesAPIHandler_MissingInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockSvc := &mockCodexRelayService{}
	handler := NewResponsesAPIHandler(mockSvc, logger)

	// Create test request without input
	reqBody := map[string]interface{}{
		"model": "gpt-4o",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/openai/responses", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	apiKey := &model.APIKey{
		ID:   1,
		Name: "test-key",
	}
	c.Set("api_key", apiKey)

	handler.HandleResponsesAPI(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	errorObj := response["error"].(map[string]interface{})
	assert.Contains(t, errorObj["message"], "Input")
}

func TestResponsesAPIHandler_WithInstructions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockSvc := &mockCodexRelayService{
		handleRequestFunc: func(c *gin.Context, apiKey *model.APIKey, reqBody *relay.CodexRequest, rawBody []byte, requestPath string) error {
			// Verify instructions were added as system message
			assert.Equal(t, 2, len(reqBody.Messages))
			assert.Equal(t, "system", reqBody.Messages[0].Role)
			assert.Equal(t, "You are Codex - Be concise and helpful.", reqBody.Messages[0].Content)
			assert.Equal(t, "user", reqBody.Messages[1].Role)

			c.JSON(http.StatusOK, gin.H{"output": "OK"})
			return nil
		},
	}

	handler := NewResponsesAPIHandler(mockSvc, logger)

	// Use Codex CLI prefix to test the instructions feature
	instructions := "You are Codex - Be concise and helpful."
	reqBody := map[string]interface{}{
		"model":        "gpt-4o",
		"input":        "Tell me a joke",
		"instructions": instructions,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/openai/responses", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	apiKey := &model.APIKey{
		ID:   1,
		Name: "test-key",
	}
	c.Set("api_key", apiKey)

	handler.HandleResponsesAPI(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
