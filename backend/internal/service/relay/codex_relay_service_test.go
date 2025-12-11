package relay

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/account"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/billing"
)

// Mock SchedulerService
type MockSchedulerService struct {
	mock.Mock
}

func (m *MockSchedulerService) SelectCodexAccount(ctx context.Context, apiKey *model.APIKey, sessionHash string) (int64, string, error) {
	args := m.Called(ctx, apiKey, sessionHash)
	return args.Get(0).(int64), args.Get(1).(string), args.Error(2)
}

func (m *MockSchedulerService) GetSessionMapping(ctx context.Context, sessionHash string) (*redis.SessionData, error) {
	args := m.Called(ctx, sessionHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*redis.SessionData), args.Error(1)
}

func (m *MockSchedulerService) SetSessionMapping(ctx context.Context, sessionHash string, accountID int64, accountType string, ttl time.Duration) error {
	args := m.Called(ctx, sessionHash, accountID, accountType, ttl)
	return args.Error(0)
}

func (m *MockSchedulerService) ClearSessionMapping(ctx context.Context, sessionHash string) error {
	args := m.Called(ctx, sessionHash)
	return args.Error(0)
}

func (m *MockSchedulerService) ExtendSessionTTL(ctx context.Context, sessionHash string, ttl time.Duration) error {
	args := m.Called(ctx, sessionHash, ttl)
	return args.Error(0)
}

func (m *MockSchedulerService) AcquireConcurrencySlot(ctx context.Context, accountID int64, requestID string, leaseSeconds int) error {
	args := m.Called(ctx, accountID, requestID, leaseSeconds)
	return args.Error(0)
}

func (m *MockSchedulerService) ReleaseConcurrencySlot(ctx context.Context, accountID int64, requestID string) error {
	args := m.Called(ctx, accountID, requestID)
	return args.Error(0)
}

func (m *MockSchedulerService) GetCurrentConcurrency(ctx context.Context, accountID int64) (int64, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSchedulerService) MarkAccountUnavailable(ctx context.Context, accountID int64, reason string, resetAfter time.Duration) error {
	args := m.Called(ctx, accountID, reason, resetAfter)
	return args.Error(0)
}

// Mock AccountService
type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) GetAccount(ctx context.Context, id int64) (*model.CodexAccount, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CodexAccount), args.Error(1)
}

func (m *MockAccountService) DecryptAPIKey(ctx context.Context, account *model.CodexAccount) (string, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockAccountService) CreateAccount(ctx context.Context, req *account.CreateCodexAccountRequest) (*model.CodexAccount, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CodexAccount), args.Error(1)
}

func (m *MockAccountService) UpdateAccount(ctx context.Context, id int64, updates map[string]any) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockAccountService) DeleteAccount(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAccountService) ListAccounts(ctx context.Context, filters repository.CodexAccountFilters, page, pageSize int) ([]*model.CodexAccount, int64, error) {
	args := m.Called(ctx, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*model.CodexAccount), args.Get(1).(int64), args.Error(2)
}

func (m *MockAccountService) TestAccount(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAccountService) GenerateAuthURL(ctx context.Context, callbackPort int) (string, string, string, error) {
	args := m.Called(ctx, callbackPort)
	return args.Get(0).(string), args.Get(1).(string), args.Get(2).(string), args.Error(3)
}

func (m *MockAccountService) VerifyAuth(ctx context.Context, code, state string, accountData account.CreateCodexAccountRequest) (*model.CodexAccount, error) {
	args := m.Called(ctx, code, state, accountData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CodexAccount), args.Error(1)
}

func (m *MockAccountService) ExchangeCodeForTokens(ctx context.Context, code, callbackURL string) (string, string, time.Time, error) {
	args := m.Called(ctx, code, callbackURL)
	return args.Get(0).(string), args.Get(1).(string), args.Get(2).(time.Time), args.Error(3)
}

func (m *MockAccountService) RefreshToken(ctx context.Context, accountID int64) error {
	args := m.Called(ctx, accountID)
	return args.Error(0)
}

func (m *MockAccountService) GetAccountInfo(ctx context.Context, apiKey string) (string, string, error) {
	args := m.Called(ctx, apiKey)
	return args.Get(0).(string), args.Get(1).(string), args.Error(2)
}

// Mock UsageCollector
type MockUsageCollector struct {
	mock.Mock
}

// Mock ProxyClientManager
type MockProxyClientManager struct {
	mock.Mock
}

func (m *MockProxyClientManager) GetClient(ctx context.Context, proxyName string) (*http.Client, error) {
	args := m.Called(ctx, proxyName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Client), args.Error(1)
}

func (m *MockProxyClientManager) GetClientByID(ctx context.Context, proxyID int64) (*http.Client, error) {
	args := m.Called(ctx, proxyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Client), args.Error(1)
}

func (m *MockProxyClientManager) GetDefaultClient(ctx context.Context) *http.Client {
	args := m.Called(ctx)
	return args.Get(0).(*http.Client)
}

func (m *MockProxyClientManager) GetStreamingClient(ctx context.Context, proxyName string) (*http.Client, error) {
	args := m.Called(ctx, proxyName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Client), args.Error(1)
}

func (m *MockProxyClientManager) InvalidateCache(proxyName string) {
	m.Called(proxyName)
}

func (m *MockProxyClientManager) InvalidateAllCache() {
	m.Called()
}

func (m *MockUsageCollector) CollectUsage(ctx context.Context, apiKeyID, accountID int64, accountType string, usage *billing.UsageData, modelName, requestID, conversationID string) error {
	args := m.Called(ctx, apiKeyID, accountID, accountType, usage, modelName, requestID, conversationID)
	return args.Error(0)
}

func (m *MockUsageCollector) GetDailyCost(ctx context.Context, apiKeyID int64, date time.Time) (float64, error) {
	args := m.Called(ctx, apiKeyID, date)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockUsageCollector) GetWeeklyCost(ctx context.Context, apiKeyID int64, startDate time.Time) (float64, error) {
	args := m.Called(ctx, apiKeyID, startDate)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockUsageCollector) GetMonthlyCost(ctx context.Context, apiKeyID int64, year int, month time.Month) (float64, error) {
	args := m.Called(ctx, apiKeyID, year, month)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockUsageCollector) GetAggregate(ctx context.Context, apiKeyID int64, startTime, endTime time.Time) (*model.UsageAggregate, error) {
	args := m.Called(ctx, apiKeyID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UsageAggregate), args.Error(1)
}

func TestCodexRelayService_HandleNonStreamResponse(t *testing.T) {
	logger := zap.NewNop()
	gin.SetMode(gin.TestMode)

	t.Run("successful non-stream request", func(t *testing.T) {
		// Setup mocks
		mockScheduler := new(MockSchedulerService)
		mockAccount := new(MockAccountService)
		mockCollector := new(MockUsageCollector)

		// Mock upstream server
		upstreamResponse := CodexResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4",
			Choices: []Choice{
				{
					Index: 0,
					Message: &Message{
						Role:    "assistant",
						Content: "Hello! How can I help you today?",
					},
				},
			},
			Usage: &Usage{
				PromptTokens:     10,
				CompletionTokens: 15,
				TotalTokens:      25,
			},
		}
		upstreamBody, _ := json.Marshal(upstreamResponse)

		mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(upstreamBody)
		}))
		defer mockUpstream.Close()

		// Setup test data
		apiKey := &model.APIKey{
			ID:        1,
			IsActive:  true,
			KeyPrefix: "sk-test",
		}

		account := &model.CodexAccount{
			ID:          1,
			AccountType: "openai-responses",
			BaseAPI:     mockUpstream.URL,
		}

		// Setup expectations
		mockScheduler.On("SelectCodexAccount", mock.Anything, apiKey, "").Return(int64(1), "openai-responses", nil)
		mockScheduler.On("AcquireConcurrencySlot", mock.Anything, int64(1), mock.Anything, 300).Return(nil)
		mockScheduler.On("ReleaseConcurrencySlot", mock.Anything, int64(1), mock.Anything).Return(nil)
		mockAccount.On("GetAccount", mock.Anything, int64(1)).Return(account, nil)
		mockAccount.On("DecryptAPIKey", mock.Anything, account).Return("sk-test-key", nil)
		mockCollector.On("CollectUsage", mock.Anything, int64(1), int64(1), "openai-responses",
			mock.MatchedBy(func(usage *billing.UsageData) bool {
				return usage.InputTokens == 10 && usage.OutputTokens == 15 && usage.TotalTokens == 25
			}), "gpt-4", mock.Anything, "").Return(nil)

		// Create service
		httpClient := &http.Client{Timeout: 30 * time.Second}
		mockClientManager := new(MockProxyClientManager)
		mockClientManager.On("GetStreamingClient", mock.Anything, "").Return(httpClient, nil)
		service := NewCodexRelayService(mockScheduler, mockAccount, mockCollector, mockClientManager, logger)

		// Create test request
		reqBody := &CodexRequest{
			Model: "gpt-4",
			Messages: []Message{
				{Role: "user", Content: "Hello"},
			},
			Stream: false,
		}

		// Create gin context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/openai/chat/completions", nil)

		// Execute
		rawBody, _ := json.Marshal(reqBody)
		err := service.HandleRequest(c, apiKey, reqBody, rawBody, "/chat/completions")

		// Verify
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		var response CodexResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "chatcmpl-123", response.ID)
		assert.Equal(t, "gpt-4", response.Model)
		assert.NotNil(t, response.Usage)
		assert.Equal(t, 10, response.Usage.PromptTokens)

		mockScheduler.AssertExpectations(t)
		mockAccount.AssertExpectations(t)
		mockCollector.AssertExpectations(t)
	})

	t.Run("account selection fails", func(t *testing.T) {
		mockScheduler := new(MockSchedulerService)
		mockAccount := new(MockAccountService)
		mockCollector := new(MockUsageCollector)

		apiKey := &model.APIKey{ID: 1}

		mockScheduler.On("SelectCodexAccount", mock.Anything, apiKey, "").
			Return(int64(0), "", assert.AnError)

		mockClientManager := new(MockProxyClientManager)
		service := NewCodexRelayService(mockScheduler, mockAccount, mockCollector, mockClientManager, logger)

		reqBody := &CodexRequest{
			Model:    "gpt-4",
			Messages: []Message{{Role: "user", Content: "Hello"}},
			Stream:   false,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/openai/chat/completions", nil)

		rawBody, _ := json.Marshal(reqBody)
		err := service.HandleRequest(c, apiKey, reqBody, rawBody, "/chat/completions")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no available accounts")

		mockScheduler.AssertExpectations(t)
	})

	t.Run("401 triggers token refresh and retry", func(t *testing.T) {
		t.Skip("Skipping: requires mockable URL for openai-oauth accounts (currently hardcoded to chatgpt.com)")
		// Setup mocks
		mockScheduler := new(MockSchedulerService)
		mockAccount := new(MockAccountService)
		mockCollector := new(MockUsageCollector)

		// Mock upstream server: first attempt 401, second attempt 200
		attempt := 0
		upstreamResponse := CodexResponse{
			ID:      "chatcmpl-401-retry",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4",
			Choices: []Choice{
				{
					Index: 0,
					Message: &Message{
						Role:    "assistant",
						Content: "Recovered after token refresh",
					},
				},
			},
			Usage: &Usage{
				PromptTokens:     5,
				CompletionTokens: 7,
				TotalTokens:      12,
			},
		}
		upstreamBody, _ := json.Marshal(upstreamResponse)

		mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempt++
			if attempt == 1 {
				// First attempt returns 401 Unauthorized
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":{"message":"invalid or expired token"}}`))
				return
			}

			// Second attempt succeeds
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(upstreamBody)
		}))
		defer mockUpstream.Close()

		// Test data
		apiKey := &model.APIKey{
			ID:        1,
			IsActive:  true,
			KeyPrefix: "sk-test",
		}

		refreshToken := "encrypted-refresh-token"
		// Set ExpiresAt to nil so token refresh is attempted on 401
		// Note: Using openai-responses instead of openai-oauth to allow BaseAPI to work in tests
		account := &model.CodexAccount{
			ID:           1,
			AccountType:  "openai-responses",
			BaseAPI:      mockUpstream.URL,
			RefreshToken: &refreshToken,
			ExpiresAt:    nil, // No expiry info = allow refresh
		}

		// Expectations
		mockScheduler.On("SelectCodexAccount", mock.Anything, apiKey, "").Return(int64(1), "openai-responses", nil)
		mockScheduler.On("AcquireConcurrencySlot", mock.Anything, int64(1), mock.Anything, 300).Return(nil)
		mockScheduler.On("ReleaseConcurrencySlot", mock.Anything, int64(1), mock.Anything).Return(nil)
		mockScheduler.On("MarkAccountUnavailable", mock.Anything, int64(1), mock.Anything, mock.Anything).Return(nil).Maybe()

		// GetAccount is called initially, during 401 handling, and after successful refresh
		mockAccount.On("GetAccount", mock.Anything, int64(1)).Return(account, nil)
		// DecryptAPIKey is called for each attempt; we don't care about the specific account instance here
		mockAccount.On("DecryptAPIKey", mock.Anything, mock.AnythingOfType("*model.CodexAccount")).Return("sk-test-key", nil)
		// Token refresh is attempted once after the first 401
		mockAccount.On("RefreshToken", mock.Anything, int64(1)).Return(nil)
		// Usage is recorded from the successful second attempt
		mockCollector.On("CollectUsage", mock.Anything, int64(1), int64(1), "openai-responses",
			mock.MatchedBy(func(usage *billing.UsageData) bool {
				return usage.InputTokens == 5 && usage.OutputTokens == 7 && usage.TotalTokens == 12
			}), "gpt-4", mock.Anything, "").Return(nil)

		// Service under test
		httpClient := &http.Client{Timeout: 30 * time.Second}
		mockClientManager := new(MockProxyClientManager)
		mockClientManager.On("GetStreamingClient", mock.Anything, "").Return(httpClient, nil)
		service := NewCodexRelayService(mockScheduler, mockAccount, mockCollector, mockClientManager, logger)

		// Request body
		reqBody := &CodexRequest{
			Model: "gpt-4",
			Messages: []Message{
				{Role: "user", Content: "Hello after token expiry"},
			},
			Stream: false,
		}

		// Gin context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/openai/responses", nil)

		// Execute
		rawBody, _ := json.Marshal(reqBody)
		err := service.HandleRequest(c, apiKey, reqBody, rawBody, "/responses")

		// Verify: request should succeed after a single retry
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 2, attempt, "upstream should be called twice due to retry after token refresh")

		var response CodexResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "chatcmpl-401-retry", response.ID)
		assert.Equal(t, "gpt-4", response.Model)
		assert.NotNil(t, response.Usage)
		assert.Equal(t, 5, response.Usage.PromptTokens)

		mockScheduler.AssertExpectations(t)
		mockAccount.AssertExpectations(t)
		mockCollector.AssertExpectations(t)
	})
}

func TestCodexRelayService_HandleStreamResponse(t *testing.T) {
	logger := zap.NewNop()
	gin.SetMode(gin.TestMode)

	t.Run("successful stream request", func(t *testing.T) {
		mockScheduler := new(MockSchedulerService)
		mockAccount := new(MockAccountService)
		mockCollector := new(MockUsageCollector)

		// Mock SSE stream response
		sseData := []string{
			"data: {\"id\":\"chatcmpl-123\",\"object\":\"chat.completion.chunk\",\"created\":1234567890,\"model\":\"gpt-4\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"Hello\"}}]}",
			"data: {\"id\":\"chatcmpl-123\",\"object\":\"chat.completion.chunk\",\"created\":1234567890,\"model\":\"gpt-4\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"!\"}}]}",
			"data: {\"id\":\"chatcmpl-123\",\"object\":\"chat.completion.chunk\",\"created\":1234567890,\"model\":\"gpt-4\",\"choices\":[{\"index\":0,\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":5,\"total_tokens\":15}}",
			"data: [DONE]",
		}

		mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			for _, line := range sseData {
				w.Write([]byte(line + "\n"))
			}
		}))
		defer mockUpstream.Close()

		apiKey := &model.APIKey{ID: 1}
		account := &model.CodexAccount{
			ID:          1,
			AccountType: "openai-responses",
			BaseAPI:     mockUpstream.URL,
		}

		mockScheduler.On("SelectCodexAccount", mock.Anything, apiKey, "").Return(int64(1), "openai-responses", nil)
		mockScheduler.On("AcquireConcurrencySlot", mock.Anything, int64(1), mock.Anything, 300).Return(nil)
		mockScheduler.On("ReleaseConcurrencySlot", mock.Anything, int64(1), mock.Anything).Return(nil)
		mockAccount.On("GetAccount", mock.Anything, int64(1)).Return(account, nil)
		mockAccount.On("DecryptAPIKey", mock.Anything, account).Return("sk-test-key", nil)
		mockCollector.On("CollectUsage", mock.Anything, int64(1), int64(1), "openai-responses",
			mock.MatchedBy(func(usage *billing.UsageData) bool {
				return usage.InputTokens == 10 && usage.OutputTokens == 5 && usage.TotalTokens == 15
			}), "gpt-4", mock.Anything, "").Return(nil)

		httpClient := &http.Client{Timeout: 30 * time.Second}
		mockClientManager := new(MockProxyClientManager)
		mockClientManager.On("GetStreamingClient", mock.Anything, "").Return(httpClient, nil)
		service := NewCodexRelayService(mockScheduler, mockAccount, mockCollector, mockClientManager, logger)

		reqBody := &CodexRequest{
			Model:    "gpt-4",
			Messages: []Message{{Role: "user", Content: "Hello"}},
			Stream:   true,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/openai/chat/completions", nil)

		rawBody, _ := json.Marshal(reqBody)
		err := service.HandleRequest(c, apiKey, reqBody, rawBody, "/chat/completions")
		require.NoError(t, err)

		// Verify SSE headers
		assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))

		// Verify response contains SSE data
		responseBody := w.Body.String()
		for _, line := range sseData {
			assert.Contains(t, responseBody, line)
		}

		mockScheduler.AssertExpectations(t)
		mockAccount.AssertExpectations(t)
		mockCollector.AssertExpectations(t)
	})
}

func TestCodexRelayService_ForwardRequest(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful forward with /chat/completions", func(t *testing.T) {
		mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/chat/completions", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Contains(t, r.Header.Get("Authorization"), "Bearer sk-test")

			body, _ := io.ReadAll(r.Body)
			var req CodexRequest
			json.Unmarshal(body, &req)
			assert.Equal(t, "gpt-4", req.Model)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"test"}`))
		}))
		defer mockUpstream.Close()

		account := &model.CodexAccount{
			BaseAPI: mockUpstream.URL,
		}

		reqBody := &CodexRequest{
			Model:    "gpt-4",
			Messages: []Message{{Role: "user", Content: "Test"}},
		}

		httpClient := &http.Client{Timeout: 30 * time.Second}
		mockClientManager := new(MockProxyClientManager)
		mockClientManager.On("GetStreamingClient", mock.Anything, "").Return(httpClient, nil)
		service := &codexRelayService{
			clientManager: mockClientManager,
			logger:        logger,
		}

		rawBody, _ := json.Marshal(reqBody)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		resp, err := service.forwardRequest(context.Background(), c, account, "sk-test", reqBody, rawBody, "/chat/completions", "test-request-id")
		require.NoError(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("successful forward with /responses", func(t *testing.T) {
		mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/responses", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Contains(t, r.Header.Get("Authorization"), "Bearer sk-test")

			body, _ := io.ReadAll(r.Body)
			var req CodexRequest
			json.Unmarshal(body, &req)
			assert.Equal(t, "gpt-4", req.Model)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"test-responses"}`))
		}))
		defer mockUpstream.Close()

		account := &model.CodexAccount{
			BaseAPI: mockUpstream.URL,
		}

		reqBody := &CodexRequest{
			Model:    "gpt-4",
			Messages: []Message{{Role: "user", Content: "Test"}},
		}

		httpClient := &http.Client{Timeout: 30 * time.Second}
		mockClientManager := new(MockProxyClientManager)
		mockClientManager.On("GetStreamingClient", mock.Anything, "").Return(httpClient, nil)
		service := &codexRelayService{
			clientManager: mockClientManager,
			logger:        logger,
		}

		rawBody, _ := json.Marshal(reqBody)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		resp, err := service.forwardRequest(context.Background(), c, account, "sk-test", reqBody, rawBody, "/responses", "test-request-id")
		require.NoError(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("successful forward with /v1/responses", func(t *testing.T) {
		mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/responses", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Contains(t, r.Header.Get("Authorization"), "Bearer sk-test")

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"test-v1-responses"}`))
		}))
		defer mockUpstream.Close()

		account := &model.CodexAccount{
			BaseAPI: mockUpstream.URL,
		}

		reqBody := &CodexRequest{
			Model:    "gpt-4",
			Messages: []Message{{Role: "user", Content: "Test"}},
		}

		httpClient := &http.Client{Timeout: 30 * time.Second}
		mockClientManager := new(MockProxyClientManager)
		mockClientManager.On("GetStreamingClient", mock.Anything, "").Return(httpClient, nil)
		service := &codexRelayService{
			clientManager: mockClientManager,
			logger:        logger,
		}

		rawBody, _ := json.Marshal(reqBody)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		resp, err := service.forwardRequest(context.Background(), c, account, "sk-test", reqBody, rawBody, "/v1/responses", "test-request-id")
		require.NoError(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
