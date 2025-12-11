//go:build wireinject
// +build wireinject

package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/config"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository/postgres"
	redisrepo "github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
	"github.com/Wei-Shaw/sub2api/backend/internal/scheduler"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/account"
	adminservice "github.com/Wei-Shaw/sub2api/backend/internal/service/admin"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/billing"
	healthsvc "github.com/Wei-Shaw/sub2api/backend/internal/service/health"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/limit"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/proxy"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/relay"
	schedulersvc "github.com/Wei-Shaw/sub2api/backend/internal/service/scheduler"
	"github.com/Wei-Shaw/sub2api/backend/pkg/database"
)

type Services struct {
	AdminService        adminservice.AdminService
	ProxyClientManager  proxy.ProxyClientManager
	ProxyConfigService  proxy.ProxyConfigService
	CodexAccountService account.CodexAccountService
	PricingService      billing.PricingService
	CostCalculator      billing.CostCalculator
	UsageCollector      billing.UsageCollector
	CostLimiter         limit.CostLimiter
	ConcurrencyTracker  limit.ConcurrencyTracker
	RateLimiter         limit.RateLimiter
	SchedulerService    schedulersvc.SchedulerService
	CleanupService      schedulersvc.CleanupService
	HealthChecker       healthsvc.HealthChecker
	CodexRelayService   relay.CodexRelayService

	AdminRepo        repository.AdminRepository
	CodexAccountRepo repository.CodexAccountRepository
	APIKeyRepo       repository.APIKeyRepository
}

func provideGormDB(db *database.DB) *gorm.DB {
	return db.DB
}

func provideRedisClient(client *redis.Client) *redis.Client {
	return client
}

func provideOAuth2Config(cfg *config.Config) *oauth2.Config {
	scopesStr := cfg.Codex.OAuth.Scopes
	var scopes []string
	if scopesStr != "" {
		scopes = strings.Split(strings.TrimSpace(scopesStr), " ")
	}
	return &oauth2.Config{
		ClientID:     cfg.Codex.OAuth.ClientID,
		ClientSecret: cfg.Codex.OAuth.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   cfg.Codex.OAuth.AuthURL,
			TokenURL:  cfg.Codex.OAuth.TokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: scopes,
	}
}

func provideDefaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

func provideSchedulerStrategy(cfg *config.Config, repo redisrepo.ConcurrencyRepository) (scheduler.SelectionStrategy, error) {
	strategyType := scheduler.StrategyType(cfg.Scheduler.Strategy)
	if strategyType == "" {
		strategyType = scheduler.StrategyTypePriority
	}
	return scheduler.NewStrategy(strategyType, repo)
}

func provideAccountConcurrencyRepo(client *redis.Client) redisrepo.ConcurrencyRepository {
	return redisrepo.NewAccountConcurrencyRepository(client)
}

func provideAdminService(
	adminRepo repository.AdminRepository,
	cfg *config.Config,
	log *zap.Logger,
) adminservice.AdminService {
	return adminservice.NewAdminService(
		adminRepo,
		cfg.Security.JWTSecret,
		cfg.Security.TokenExpiration,
		log,
	)
}

func provideProxyClientManager(
	proxyRepo repository.ProxyConfigRepository,
	cfg *config.Config,
	log *zap.Logger,
) proxy.ProxyClientManager {
	return proxy.NewProxyClientManager(
		proxyRepo,
		cfg.Security.EncryptionKey,
		log,
	)
}

func provideProxyConfigService(
	proxyRepo repository.ProxyConfigRepository,
	proxyClientManager proxy.ProxyClientManager,
	cfg *config.Config,
	log *zap.Logger,
) proxy.ProxyConfigService {
	return proxy.NewProxyConfigService(
		proxyRepo,
		cfg.Security.EncryptionKey,
		log,
		proxyClientManager,
	)
}

func provideCodexAccountService(
	codexAccountRepo repository.CodexAccountRepository,
	oauthStateRepo redisrepo.OAuthStateRepository,
	oauthConfig *oauth2.Config,
	defaultHTTPClient *http.Client,
	proxyClientManager proxy.ProxyClientManager,
	cfg *config.Config,
	log *zap.Logger,
) account.CodexAccountService {
	return account.NewCodexAccountService(
		codexAccountRepo,
		oauthStateRepo,
		cfg.Security.EncryptionKey,
		oauthConfig,
		defaultHTTPClient,
		proxyClientManager,
		log,
	)
}

func providePricingService(log *zap.Logger) billing.PricingService {
	return billing.NewPricingService("config/model_prices_and_context_window.json", log)
}

func provideSchedulerService(
	codexAccountRepo repository.CodexAccountRepository,
	sessionRepo redisrepo.SessionRepository,
	concurrencyRepo redisrepo.ConcurrencyRepository,
	healthRepo redisrepo.HealthRepository,
	strategy scheduler.SelectionStrategy,
	cfg *config.Config,
	log *zap.Logger,
) schedulersvc.SchedulerService {
	sessionTTL := cfg.Scheduler.SessionTTL
	if sessionTTL == 0 {
		sessionTTL = 24 * time.Hour
	}
	maxAccountConcurrency := cfg.Limits.DefaultConcurrentRequests
	return schedulersvc.NewSchedulerService(
		codexAccountRepo,
		sessionRepo,
		concurrencyRepo,
		healthRepo,
		strategy,
		sessionTTL,
		maxAccountConcurrency,
		log,
	)
}

func provideCleanupService(
	codexAccountRepo repository.CodexAccountRepository,
	log *zap.Logger,
) schedulersvc.CleanupService {
	return schedulersvc.NewCleanupService(
		codexAccountRepo,
		log,
		5*time.Minute,
	)
}

func provideHealthChecker(
	codexAccountRepo repository.CodexAccountRepository,
	healthRepo redisrepo.HealthRepository,
	log *zap.Logger,
) healthsvc.HealthChecker {
	return healthsvc.NewHealthChecker(
		codexAccountRepo,
		healthRepo,
		2*time.Minute,
		log,
	)
}

func provideCodexRelayService(
	schedulerService schedulersvc.SchedulerService,
	codexAccountService account.CodexAccountService,
	usageCollector billing.UsageCollector,
	proxyClientManager proxy.ProxyClientManager,
	cfg *config.Config,
	log *zap.Logger,
) relay.CodexRelayService {
	return relay.NewCodexRelayService(
		schedulerService,
		codexAccountService,
		usageCollector,
		proxyClientManager,
		log,
		cfg.Logging.LogPayloads,
	)
}

var repositorySet = wire.NewSet(
	postgres.NewCodexAccountRepository,
	postgres.NewAPIKeyRepository,
	postgres.NewAdminRepository,
	postgres.NewUsageRepository,
	postgres.NewProxyConfigRepository,
	redisrepo.NewSessionRepository,
	provideAccountConcurrencyRepo,
	redisrepo.NewOAuthStateRepository,
	redisrepo.NewHealthRepository,
	redisrepo.NewRateLimitRepository,
)

func provideConcurrencyTracker(client *redis.Client, log *zap.Logger) limit.ConcurrencyTracker {
	apiKeyConcurrencyRepo := redisrepo.NewAPIKeyConcurrencyRepository(client)
	return limit.NewConcurrencyTracker(apiKeyConcurrencyRepo, log)
}

var serviceSet = wire.NewSet(
	provideAdminService,
	provideProxyClientManager,
	provideProxyConfigService,
	provideCodexAccountService,
	providePricingService,
	billing.NewCostCalculator,
	billing.NewUsageCollector,
	limit.NewCostLimiter,
	provideSchedulerService,
	provideCleanupService,
	provideHealthChecker,
	provideCodexRelayService,
	provideConcurrencyTracker,
	limit.NewRateLimiter,
	wire.Struct(new(Services), "*"),
)

func InitializeServices(
	cfg *config.Config,
	log *zap.Logger,
	db *database.DB,
	redisClient *redis.Client,
) (*Services, error) {
	wire.Build(
		provideGormDB,
		provideOAuth2Config,
		provideDefaultHTTPClient,
		provideSchedulerStrategy,
		repositorySet,
		serviceSet,
	)
	return nil, nil
}
