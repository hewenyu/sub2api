package account

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	redisrepo "github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
	"github.com/Wei-Shaw/sub2api/backend/pkg/crypto"
)

// MockCodexAccountRepository is a mock implementation of CodexAccountRepository
type MockCodexAccountRepository struct {
	mock.Mock
}

// MockProxyClientManager is a mock implementation of ProxyClientManager
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

func (m *MockCodexAccountRepository) Create(ctx context.Context, account *model.CodexAccount) error {
	args := m.Called(ctx, account)
	if args.Get(0) != nil {
		// Simulate auto-increment ID
		account.ID = 1
	}
	return args.Error(0)
}

func (m *MockCodexAccountRepository) GetByID(ctx context.Context, id int64) (*model.CodexAccount, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CodexAccount), args.Error(1)
}

func (m *MockCodexAccountRepository) GetByEmail(ctx context.Context, email string) (*model.CodexAccount, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CodexAccount), args.Error(1)
}

func (m *MockCodexAccountRepository) GetByAPIKey(ctx context.Context, apiKey string) (*model.CodexAccount, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CodexAccount), args.Error(1)
}

func (m *MockCodexAccountRepository) List(ctx context.Context, filters repository.CodexAccountFilters, offset, limit int) ([]*model.CodexAccount, int64, error) {
	args := m.Called(ctx, filters, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*model.CodexAccount), args.Get(1).(int64), args.Error(2)
}

func (m *MockCodexAccountRepository) GetSchedulable(ctx context.Context) ([]*model.CodexAccount, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.CodexAccount), args.Error(1)
}

func (m *MockCodexAccountRepository) Update(ctx context.Context, account *model.CodexAccount) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockCodexAccountRepository) UpdateFields(ctx context.Context, id int64, updates map[string]any) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockCodexAccountRepository) UpdateConcurrentRequests(ctx context.Context, id int64, delta int) error {
	args := m.Called(ctx, id, delta)
	return args.Error(0)
}

func (m *MockCodexAccountRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTestService() (*codexAccountService, *MockCodexAccountRepository, *miniredis.Miniredis) {
	mockRepo := new(MockCodexAccountRepository)
	logger := zap.NewNop()
	encryptionKey := "12345678901234567890123456789012" // 32 bytes

	// Setup miniredis for OAuth state repository
	mr, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	oauthStateRepo := redisrepo.NewOAuthStateRepository(redisClient)

	oauthConfig := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://auth.example.com/authorize",
			TokenURL: "https://auth.example.com/token",
		},
		Scopes: []string{"openid", "profile", "email"},
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	service := &codexAccountService{
		repo:               mockRepo,
		oauthStateRepo:     oauthStateRepo,
		encryptionKey:      encryptionKey,
		oauthConfig:        oauthConfig,
		httpClient:         httpClient,
		proxyClientManager: nil,
		logger:             logger,
	}

	return service, mockRepo, mr
}

func TestGenerateAuthURL(t *testing.T) {
	service, _, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	authURL, callbackURL, state, err := service.GenerateAuthURL(ctx, 8080)

	require.NoError(t, err)
	assert.NotEmpty(t, authURL)
	// Codex OAuth must use the fixed official redirect URI
	assert.Equal(t, "http://localhost:1455/auth/callback", callbackURL)
	assert.NotEmpty(t, state)

	// Verify state is hex-encoded (64 characters from 32 bytes)
	assert.Equal(t, 64, len(state), "Hex-encoded state should be 64 characters")

	assert.Contains(t, authURL, "client_id=test-client-id")
	assert.Contains(t, authURL, "state=")
	assert.Contains(t, authURL, "redirect_uri=http%3A%2F%2Flocalhost%3A1455%2Fauth%2Fcallback")
	assert.Contains(t, authURL, "code_challenge=")
	assert.Contains(t, authURL, "code_challenge_method=S256")
}

func TestCreateAccount_OpenAIResponses(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	apiKey := "sk-test1234567890"
	req := &CreateCodexAccountRequest{
		Name:           "Test Account",
		AccountType:    "openai-responses",
		APIKey:         &apiKey,
		BaseAPI:        "https://api.openai.com/v1",
		DailyQuota:     100.0,
		QuotaResetTime: "00:00",
		Priority:       100,
		Schedulable:    true,
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.CodexAccount")).Return(nil)

	account, err := service.CreateAccount(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, "Test Account", account.Name)
	assert.Equal(t, "openai-responses", account.AccountType)
	assert.NotNil(t, account.APIKey)
	assert.True(t, account.IsActive)

	// Verify API key was encrypted
	decryptedKey, err := crypto.AES256Decrypt(*account.APIKey, service.encryptionKey)
	require.NoError(t, err)
	assert.Equal(t, apiKey, decryptedKey)

	mockRepo.AssertExpectations(t)
}

func TestCreateAccount_MissingAPIKey(t *testing.T) {
	service, _, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	req := &CreateCodexAccountRequest{
		Name:        "Test Account",
		AccountType: "openai-responses",
		// APIKey is nil
	}

	_, err := service.CreateAccount(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key is required")
}

func TestCreateAccount_WithDefaults(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	apiKey := "sk-test1234567890"
	req := &CreateCodexAccountRequest{
		Name:        "Test Account",
		AccountType: "openai-responses",
		APIKey:      &apiKey,
		// BaseAPI not provided
		// QuotaResetTime not provided
		// Priority not provided
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.CodexAccount")).Return(nil)

	account, err := service.CreateAccount(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, "https://api.openai.com/v1", account.BaseAPI)
	assert.Equal(t, "00:00", account.QuotaResetTime)
	assert.Equal(t, 100, account.Priority)

	mockRepo.AssertExpectations(t)
}

func TestGetAccount(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	expectedAccount := &model.CodexAccount{
		ID:          1,
		Name:        "Test Account",
		AccountType: "openai-responses",
		IsActive:    true,
	}

	mockRepo.On("GetByID", ctx, int64(1)).Return(expectedAccount, nil)

	account, err := service.GetAccount(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, expectedAccount, account)
	mockRepo.AssertExpectations(t)
}

func TestGetAccount_NotFound(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(999)).Return(nil, gorm.ErrRecordNotFound)

	_, err := service.GetAccount(ctx, 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account not found")
	mockRepo.AssertExpectations(t)
}

func TestUpdateAccount(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	updates := map[string]any{
		"name":     "Updated Name",
		"priority": 200,
	}

	mockRepo.On("UpdateFields", ctx, int64(1), updates).Return(nil)

	err := service.UpdateAccount(ctx, 1, updates)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateAccount_EncryptsAPIKey(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	newAPIKey := "sk-newapikey123"
	updates := map[string]any{
		"api_key": newAPIKey,
	}

	mockRepo.On("UpdateFields", ctx, int64(1), mock.MatchedBy(func(u map[string]any) bool {
		encryptedKey, ok := u["api_key"].(string)
		if !ok {
			return false
		}
		// Verify it's encrypted
		decrypted, err := crypto.AES256Decrypt(encryptedKey, service.encryptionKey)
		return err == nil && decrypted == newAPIKey
	})).Return(nil)

	err := service.UpdateAccount(ctx, 1, updates)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteAccount(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	mockRepo.On("Delete", ctx, int64(1)).Return(nil)

	err := service.DeleteAccount(ctx, 1)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestListAccounts(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	expectedAccounts := []*model.CodexAccount{
		{ID: 1, Name: "Account 1"},
		{ID: 2, Name: "Account 2"},
	}

	filters := repository.CodexAccountFilters{}
	mockRepo.On("List", ctx, filters, 0, 20).Return(expectedAccounts, int64(2), nil)

	accounts, total, err := service.ListAccounts(ctx, filters, 1, 20)

	require.NoError(t, err)
	assert.Equal(t, expectedAccounts, accounts)
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestListAccounts_Pagination(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	expectedAccounts := []*model.CodexAccount{
		{ID: 11, Name: "Account 11"},
	}

	filters := repository.CodexAccountFilters{}
	// Page 2 with pageSize 10 = offset 10
	mockRepo.On("List", ctx, filters, 10, 10).Return(expectedAccounts, int64(15), nil)

	accounts, total, err := service.ListAccounts(ctx, filters, 2, 10)

	require.NoError(t, err)
	assert.Equal(t, expectedAccounts, accounts)
	assert.Equal(t, int64(15), total)
	mockRepo.AssertExpectations(t)
}

func TestListAccounts_MaxPageSize(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	filters := repository.CodexAccountFilters{}
	// Request 200, should be capped at 100
	mockRepo.On("List", ctx, filters, 0, 100).Return([]*model.CodexAccount{}, int64(0), nil)

	_, _, err := service.ListAccounts(ctx, filters, 1, 200)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDecryptAPIKey_OpenAIResponses(t *testing.T) {
	service, _, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	originalKey := "sk-test1234567890"
	encryptedKey, err := crypto.AES256Encrypt(originalKey, service.encryptionKey)
	require.NoError(t, err)

	account := &model.CodexAccount{
		AccountType: "openai-responses",
		APIKey:      &encryptedKey,
	}

	decryptedKey, err := service.DecryptAPIKey(ctx, account)

	require.NoError(t, err)
	assert.Equal(t, originalKey, decryptedKey)
}

func TestDecryptAPIKey_OAuth(t *testing.T) {
	service, _, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	originalToken := "access-token-12345"
	encryptedToken, err := crypto.AES256Encrypt(originalToken, service.encryptionKey)
	require.NoError(t, err)

	account := &model.CodexAccount{
		AccountType: "openai-oauth",
		AccessToken: &encryptedToken,
	}

	decryptedToken, err := service.DecryptAPIKey(ctx, account)

	require.NoError(t, err)
	assert.Equal(t, originalToken, decryptedToken)
}

func TestDecryptAPIKey_Missing(t *testing.T) {
	service, _, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	account := &model.CodexAccount{
		AccountType: "openai-responses",
		APIKey:      nil,
	}

	_, err := service.DecryptAPIKey(ctx, account)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no API key available")
}

func TestTestAccount_Success(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	// Create test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/models", r.URL.Path)
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": []}`))
	}))
	defer testServer.Close()

	apiKey := "sk-test1234567890"
	encryptedKey, err := crypto.AES256Encrypt(apiKey, service.encryptionKey)
	require.NoError(t, err)

	account := &model.CodexAccount{
		ID:          1,
		Name:        "Test Account",
		AccountType: "openai-responses",
		BaseAPI:     testServer.URL,
		APIKey:      &encryptedKey,
	}

	mockRepo.On("GetByID", ctx, int64(1)).Return(account, nil)

	err = service.TestAccount(ctx, 1)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTestAccount_Failure(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	// Create test server that returns error
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid api key"}`))
	}))
	defer testServer.Close()

	apiKey := "sk-invalid"
	encryptedKey, err := crypto.AES256Encrypt(apiKey, service.encryptionKey)
	require.NoError(t, err)

	account := &model.CodexAccount{
		ID:          1,
		AccountType: "openai-responses",
		BaseAPI:     testServer.URL,
		APIKey:      &encryptedKey,
	}

	mockRepo.On("GetByID", ctx, int64(1)).Return(account, nil)

	err = service.TestAccount(ctx, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	mockRepo.AssertExpectations(t)
}

func TestRefreshToken(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	// This test is basic since we can't easily mock OAuth2 token refresh
	refreshToken := "refresh-token-123"
	encryptedRefreshToken, err := crypto.AES256Encrypt(refreshToken, service.encryptionKey)
	require.NoError(t, err)

	account := &model.CodexAccount{
		ID:           1,
		AccountType:  "openai-oauth",
		RefreshToken: &encryptedRefreshToken,
	}

	mockRepo.On("GetByID", ctx, int64(1)).Return(account, nil)

	// This will fail because we don't have a real OAuth server
	// but it tests the basic flow
	err = service.RefreshToken(ctx, 1)

	// We expect an error here because the OAuth exchange will fail
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRefreshToken_NotOAuthAccount(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	account := &model.CodexAccount{
		ID:          1,
		AccountType: "openai-responses",
	}

	mockRepo.On("GetByID", ctx, int64(1)).Return(account, nil)

	err := service.RefreshToken(ctx, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not an OAuth account")
	mockRepo.AssertExpectations(t)
}

func TestRefreshToken_NoRefreshToken(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	account := &model.CodexAccount{
		ID:           1,
		AccountType:  "openai-oauth",
		RefreshToken: nil,
	}

	mockRepo.On("GetByID", ctx, int64(1)).Return(account, nil)

	err := service.RefreshToken(ctx, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no refresh token")
	mockRepo.AssertExpectations(t)
}

func TestCreateAccount_WithProxyName(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	apiKey := "sk-test1234567890"
	proxyName := "proxy-example"

	req := &CreateCodexAccountRequest{
		Name:        "Test Account",
		AccountType: "openai-responses",
		APIKey:      &apiKey,
		ProxyName:   &proxyName,
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.CodexAccount")).Return(nil)

	account, err := service.CreateAccount(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, account.ProxyName)
	assert.Equal(t, proxyName, *account.ProxyName)

	mockRepo.AssertExpectations(t)
}

func TestListAccounts_WithFilters(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	accountType := "openai-oauth"
	isActive := true
	filters := repository.CodexAccountFilters{
		AccountType: &accountType,
		IsActive:    &isActive,
	}

	expectedAccounts := []*model.CodexAccount{
		{ID: 1, Name: "OAuth Account 1", AccountType: "openai-oauth", IsActive: true},
	}

	mockRepo.On("List", ctx, filters, 0, 20).Return(expectedAccounts, int64(1), nil)

	accounts, total, err := service.ListAccounts(ctx, filters, 1, 20)

	require.NoError(t, err)
	assert.Equal(t, 1, len(accounts))
	assert.Equal(t, int64(1), total)
	mockRepo.AssertExpectations(t)
}

func TestGetByID_RepositoryError(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("database error"))

	_, err := service.GetAccount(ctx, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get account")
	mockRepo.AssertExpectations(t)
}

func TestRefreshToken_ClearsRateLimits(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	// Setup mock OAuth server that returns a successful token response
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Return a mock OAuth token response
			w.Write([]byte(`{
				"access_token": "new-access-token-123",
				"token_type": "Bearer",
				"expires_in": 3600,
				"refresh_token": "new-refresh-token-456"
			}`))
		}
	}))
	defer testServer.Close()

	// Update service OAuth config to use test server
	service.oauthConfig = &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  testServer.URL + "/authorize",
			TokenURL: testServer.URL + "/token",
		},
	}

	refreshToken := "refresh-token-123"
	encryptedRefreshToken, err := crypto.AES256Encrypt(refreshToken, service.encryptionKey)
	require.NoError(t, err)

	// Account has rate_limited_until set (simulating the bug scenario)
	rateLimitedUntil := time.Now().Add(1 * time.Minute)
	account := &model.CodexAccount{
		ID:               1,
		AccountType:      "openai-oauth",
		RefreshToken:     &encryptedRefreshToken,
		RateLimitedUntil: &rateLimitedUntil,
		RateLimitStatus:  stringPtr("token_refreshed"),
	}

	mockRepo.On("GetByID", ctx, int64(1)).Return(account, nil)

	// Verify that UpdateFields is called with rate limits cleared
	mockRepo.On("UpdateFields", mock.Anything, int64(1), mock.MatchedBy(func(updates map[string]any) bool {
		// Verify the updates contain the correct fields
		_, hasAccessToken := updates["access_token"]
		_, hasExpiresAt := updates["expires_at"]
		_, hasRefreshToken := updates["refresh_token"]
		rateLimitedUntil, hasRateLimitedUntil := updates["rate_limited_until"]
		rateLimitStatus, hasRateLimitStatus := updates["rate_limit_status"]

		// All required fields should be present
		if !hasAccessToken || !hasExpiresAt || !hasRefreshToken {
			return false
		}

		// Rate limit fields should be present and set to nil
		if !hasRateLimitedUntil || rateLimitedUntil != nil {
			return false
		}
		if !hasRateLimitStatus || rateLimitStatus != nil {
			return false
		}

		return true
	})).Return(nil)

	err = service.RefreshToken(ctx, 1)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}

func TestTestAccount_WithProxy(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	mockProxyManager := new(MockProxyClientManager)
	service.proxyClientManager = mockProxyManager

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/models", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": []}`))
	}))
	defer testServer.Close()

	apiKey := "sk-test1234567890"
	encryptedKey, err := crypto.AES256Encrypt(apiKey, service.encryptionKey)
	require.NoError(t, err)

	proxyName := "test-proxy"
	account := &model.CodexAccount{
		ID:          1,
		AccountType: "openai-responses",
		BaseAPI:     testServer.URL,
		APIKey:      &encryptedKey,
		ProxyName:   &proxyName,
	}

	proxyClient := &http.Client{Timeout: 10 * time.Second}

	mockRepo.On("GetByID", ctx, int64(1)).Return(account, nil)
	mockProxyManager.On("GetClient", ctx, proxyName).Return(proxyClient, nil)

	err = service.TestAccount(ctx, 1)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProxyManager.AssertExpectations(t)
}

func TestTestAccount_WithProxyFallback(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	mockProxyManager := new(MockProxyClientManager)
	service.proxyClientManager = mockProxyManager

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": []}`))
	}))
	defer testServer.Close()

	apiKey := "sk-test1234567890"
	encryptedKey, err := crypto.AES256Encrypt(apiKey, service.encryptionKey)
	require.NoError(t, err)

	proxyName := "invalid-proxy"
	account := &model.CodexAccount{
		ID:          1,
		AccountType: "openai-responses",
		BaseAPI:     testServer.URL,
		APIKey:      &encryptedKey,
		ProxyName:   &proxyName,
	}

	mockRepo.On("GetByID", ctx, int64(1)).Return(account, nil)
	mockProxyManager.On("GetClient", ctx, proxyName).Return(nil, errors.New("proxy not found"))

	err = service.TestAccount(ctx, 1)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProxyManager.AssertExpectations(t)
}

func TestTestAccount_WithoutProxy(t *testing.T) {
	service, mockRepo, mr := setupTestService()
	defer mr.Close()
	ctx := context.Background()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": []}`))
	}))
	defer testServer.Close()

	apiKey := "sk-test1234567890"
	encryptedKey, err := crypto.AES256Encrypt(apiKey, service.encryptionKey)
	require.NoError(t, err)

	account := &model.CodexAccount{
		ID:          1,
		AccountType: "openai-responses",
		BaseAPI:     testServer.URL,
		APIKey:      &encryptedKey,
		ProxyName:   nil,
	}

	mockRepo.On("GetByID", ctx, int64(1)).Return(account, nil)

	err = service.TestAccount(ctx, 1)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
