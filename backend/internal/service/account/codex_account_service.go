package account

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	redisrepo "github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/proxy"
)

// CodexAccountService defines the interface for Codex account management.
type CodexAccountService interface {
	// OAuth authorization
	GenerateAuthURL(ctx context.Context, callbackPort int) (authURL, callbackURL, state string, err error)
	VerifyAuth(ctx context.Context, code, state string, accountData CreateCodexAccountRequest) (*model.CodexAccount, error)
	ExchangeCodeForTokens(ctx context.Context, code, callbackURL string) (accessToken, refreshToken string, expiresAt time.Time, err error)

	// Token management
	RefreshToken(ctx context.Context, accountID int64) error
	DecryptAPIKey(ctx context.Context, account *model.CodexAccount) (string, error)

	// Account info
	GetAccountInfo(ctx context.Context, apiKey string) (email string, subscriptionLevel string, err error)

	// CRUD operations
	CreateAccount(ctx context.Context, req *CreateCodexAccountRequest) (*model.CodexAccount, error)
	GetAccount(ctx context.Context, id int64) (*model.CodexAccount, error)
	UpdateAccount(ctx context.Context, id int64, updates map[string]any) error
	DeleteAccount(ctx context.Context, id int64) error
	ListAccounts(ctx context.Context, filters repository.CodexAccountFilters, page, pageSize int) ([]*model.CodexAccount, int64, error)

	// Health check
	TestAccount(ctx context.Context, id int64) error
}

type codexAccountService struct {
	repo               repository.CodexAccountRepository
	oauthStateRepo     redisrepo.OAuthStateRepository
	encryptionKey      string
	oauthConfig        *oauth2.Config
	httpClient         *http.Client
	proxyClientManager proxy.ProxyClientManager
	logger             *zap.Logger
}

// NewCodexAccountService creates a new Codex account service.
func NewCodexAccountService(
	repo repository.CodexAccountRepository,
	oauthStateRepo redisrepo.OAuthStateRepository,
	encryptionKey string,
	oauthConfig *oauth2.Config,
	httpClient *http.Client,
	proxyClientManager proxy.ProxyClientManager,
	logger *zap.Logger,
) CodexAccountService {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &codexAccountService{
		repo:               repo,
		oauthStateRepo:     oauthStateRepo,
		encryptionKey:      encryptionKey,
		oauthConfig:        oauthConfig,
		httpClient:         httpClient,
		proxyClientManager: proxyClientManager,
		logger:             logger,
	}
}

// getHTTPClient returns the appropriate HTTP client for the account.
// If the account has a proxy configured, it returns a proxy-aware client.
// Otherwise, it returns the default client.
func (s *codexAccountService) getHTTPClient(ctx context.Context, account *model.CodexAccount) (*http.Client, error) {
	if s.proxyClientManager == nil || account.ProxyName == nil || *account.ProxyName == "" {
		return s.httpClient, nil
	}

	client, err := s.proxyClientManager.GetClient(ctx, *account.ProxyName)
	if err != nil {
		s.logger.Warn("Failed to get proxy client, falling back to default",
			zap.String("proxy_name", *account.ProxyName),
			zap.Error(err),
		)
		return s.httpClient, nil
	}

	return client, nil
}

// getHTTPClientByProxyName returns the appropriate HTTP client based on proxy name.
// If proxyName is nil or empty, returns the default client (no proxy).
// If proxy retrieval fails, returns an error.
func (s *codexAccountService) getHTTPClientByProxyName(ctx context.Context, proxyName *string) (*http.Client, error) {
	if s.proxyClientManager == nil || proxyName == nil || *proxyName == "" {
		if s.proxyClientManager != nil {
			return s.proxyClientManager.GetDefaultClient(ctx), nil
		}
		return s.httpClient, nil
	}

	client, err := s.proxyClientManager.GetClient(ctx, *proxyName)
	if err != nil {
		return nil, err
	}

	return client, nil
}
