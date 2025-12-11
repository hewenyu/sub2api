package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/Wei-Shaw/sub2api/backend/config"
	"github.com/Wei-Shaw/sub2api/backend/internal/api/admin"
	"github.com/Wei-Shaw/sub2api/backend/internal/api/codex"
	"github.com/Wei-Shaw/sub2api/backend/internal/middleware"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository/postgres"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
	"github.com/Wei-Shaw/sub2api/backend/internal/scheduler"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/account"
	adminservice "github.com/Wei-Shaw/sub2api/backend/internal/service/admin"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/billing"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/proxy"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/relay"
	schedulersvc "github.com/Wei-Shaw/sub2api/backend/internal/service/scheduler"
	"github.com/Wei-Shaw/sub2api/backend/pkg/database"
	"github.com/Wei-Shaw/sub2api/backend/pkg/logger"
)

var (
	configPath = flag.String("config", "config/config.yaml", "Path to config file")
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logConfig := logger.LoggingConfig{
		Level:           cfg.Logging.Level,
		Format:          cfg.Logging.Format,
		OutputPath:      cfg.Logging.OutputPath,
		ErrorOutputPath: cfg.Logging.ErrorOutputPath,
	}
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = log.Sync()
	}()

	log.Info("Starting Claude Relay Go server",
		zap.String("version", "1.0.0"),
		zap.String("mode", cfg.Server.Mode),
	)

	// Initialize database
	db, err := database.NewPostgresDB(&cfg.Database, log)
	if err != nil {
		log.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Error("Failed to close database", zap.Error(closeErr))
		}
	}()

	log.Info("Database initialized successfully")

	// Initialize Redis
	redisClient, err := database.NewRedisClient(&cfg.Redis, log)
	if err != nil {
		log.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Error("Failed to close Redis", zap.Error(err))
		}
	}()

	log.Info("Redis initialized successfully")

	// Initialize repositories
	codexAccountRepo := postgres.NewCodexAccountRepository(db.DB)
	apiKeyRepo := postgres.NewAPIKeyRepository(db.DB)
	adminRepo := postgres.NewAdminRepository(db.DB)
	usageRepo := postgres.NewUsageRepository(db.DB)
	proxyRepo := postgres.NewProxyConfigRepository(db.DB)
	sessionRepo := redis.NewSessionRepository(redisClient.Client)
	concurrencyRepo := redis.NewConcurrencyRepository(redisClient.Client)
	oauthStateRepo := redis.NewOAuthStateRepository(redisClient.Client)

	// Initialize OAuth2 config for Codex (OpenAI)
	// Split scopes string by space (e.g., "openid profile email offline_access")
	scopesStr := cfg.Codex.OAuth.Scopes
	var scopes []string
	if scopesStr != "" {
		scopes = strings.Split(strings.TrimSpace(scopesStr), " ")
	}
	codexOAuthConfig := &oauth2.Config{
		ClientID:     cfg.Codex.OAuth.ClientID,
		ClientSecret: cfg.Codex.OAuth.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   cfg.Codex.OAuth.AuthURL,
			TokenURL:  cfg.Codex.OAuth.TokenURL,
			AuthStyle: oauth2.AuthStyleInParams, // CRITICAL: Send client_id in POST body, not as HTTP Basic Auth
			// OpenAI OAuth requires client_id in request body for public clients (no client_secret)
			// This matches the Node.js implementation behavior
		},
		Scopes: scopes,
	}

	// Initialize services
	adminService := adminservice.NewAdminService(
		adminRepo,
		cfg.Security.JWTSecret,
		cfg.Security.TokenExpiration,
		log,
	)

	// Initialize proxy client manager
	proxyClientManager := proxy.NewProxyClientManager(
		proxyRepo,
		cfg.Security.EncryptionKey,
		log,
	)

	// Initialize proxy config service
	proxyConfigService := proxy.NewProxyConfigService(
		proxyRepo,
		cfg.Security.EncryptionKey,
		log,
		proxyClientManager,
	)

	// Create default HTTP client for OAuth operations (when no proxy is configured)
	defaultHTTPClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	codexAccountService := account.NewCodexAccountService(
		codexAccountRepo,
		oauthStateRepo,
		cfg.Security.EncryptionKey,
		codexOAuthConfig,
		defaultHTTPClient,
		proxyClientManager,
		log,
	)

	// Initialize billing services
	pricingFile := "config/model_prices_and_context_window.json"
	pricingService := billing.NewPricingService(pricingFile, log)
	if err := pricingService.LoadPricing(); err != nil {
		log.Fatal("Failed to load pricing data", zap.Error(err))
	}

	costCalculator := billing.NewCostCalculator(pricingService, log)
	usageCollector := billing.NewUsageCollector(usageRepo, costCalculator, log)

	// Initialize scheduler service
	strategy := scheduler.NewPriorityStrategy(concurrencyRepo)
	sessionTTL := 24 * time.Hour
	schedulerService := schedulersvc.NewSchedulerService(
		codexAccountRepo,
		sessionRepo,
		concurrencyRepo,
		strategy,
		sessionTTL,
		log,
	)

	// Initialize and start cleanup service (clears expired rate limits)
	cleanupInterval := 5 * time.Minute // Check every 5 minutes, same as Node.js implementation
	cleanupService := schedulersvc.NewCleanupService(
		codexAccountRepo,
		log,
		cleanupInterval,
	)
	ctx := context.Background()
	cleanupService.Start(ctx)
	log.Info("Cleanup service started",
		zap.Duration("interval", cleanupInterval),
	)

	// Initialize relay service with proxy client manager
	codexRelayService := relay.NewCodexRelayService(
		schedulerService,
		codexAccountService,
		usageCollector,
		proxyClientManager,
		log,
	)

	log.Info("All services initialized successfully")

	// Redirect Gin's default output to our logger
	// Gin outputs debug info (route registration, etc.) to gin.DefaultWriter
	gin.DefaultWriter = logger.NewGinWriter(log)
	gin.DefaultErrorWriter = logger.NewGinWriter(log)

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create Gin engine
	router := gin.New()
	router.Use(middleware.ZapLogger(log))
	router.Use(middleware.ZapRecovery(log))

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	// API routes
	apiV1 := router.Group("/api/v1")

	// Admin authentication routes (public)
	loginHandler := admin.NewLoginHandler(adminService, adminRepo, log)
	apiV1.POST("/admin/login", loginHandler.Login)

	// Admin routes (protected by admin authentication)
	adminGroup := apiV1.Group("/admin")
	adminGroup.Use(middleware.AuthenticateAdmin(adminService, log))
	{
		// Admin info
		adminGroup.GET("/info", loginHandler.GetInfo)

		// Codex account management
		codexAccountHandler := admin.NewCodexAccountHandler(codexAccountService, log)
		adminGroup.POST("/codex-accounts/generate-auth-url", codexAccountHandler.GenerateAuthURL)
		adminGroup.POST("/codex-accounts/verify-auth", codexAccountHandler.VerifyAuth)
		adminGroup.POST("/codex-accounts", codexAccountHandler.CreateAccount)
		adminGroup.GET("/codex-accounts", codexAccountHandler.ListAccounts)
		adminGroup.GET("/codex-accounts/:id", codexAccountHandler.GetAccount)
		adminGroup.PUT("/codex-accounts/:id", codexAccountHandler.UpdateAccount)
		adminGroup.DELETE("/codex-accounts/:id", codexAccountHandler.DeleteAccount)
		adminGroup.POST("/codex-accounts/:id/toggle", codexAccountHandler.ToggleStatus)
		adminGroup.POST("/codex-accounts/:id/test", codexAccountHandler.TestAccount)
		adminGroup.POST("/codex-accounts/:id/refresh-token", codexAccountHandler.RefreshToken)
		adminGroup.POST("/codex-accounts/:id/clear-rate-limit", codexAccountHandler.ClearRateLimit)

		// API Key management
		apiKeyHandler := admin.NewAPIKeyHandler(apiKeyRepo, log)
		adminGroup.POST("/api-keys", apiKeyHandler.CreateAPIKey)
		adminGroup.GET("/api-keys", apiKeyHandler.ListAPIKeys)
		adminGroup.GET("/api-keys/:id", apiKeyHandler.GetAPIKey)
		adminGroup.PUT("/api-keys/:id", apiKeyHandler.UpdateAPIKey)
		adminGroup.DELETE("/api-keys/:id", apiKeyHandler.DeleteAPIKey)
		adminGroup.PATCH("/api-keys/:id/toggle", apiKeyHandler.ToggleAPIKeyStatus)

		// Proxy management
		proxyHandler := admin.NewProxyHandler(proxyConfigService, log)
		adminGroup.GET("/proxies", proxyHandler.ListProxies)
		adminGroup.POST("/proxies", proxyHandler.CreateProxy)
		adminGroup.GET("/proxies/names", proxyHandler.GetProxyNames)
		adminGroup.GET("/proxies/:id", proxyHandler.GetProxy)
		adminGroup.PUT("/proxies/:id", proxyHandler.UpdateProxy)
		adminGroup.DELETE("/proxies/:id", proxyHandler.DeleteProxy)
		adminGroup.POST("/proxies/:id/set-default", proxyHandler.SetDefaultProxy)
		adminGroup.POST("/proxies/:id/test", proxyHandler.TestProxy)

		// Usage Records (placeholder - TODO: Implement usage record handler)
		adminGroup.GET("/usage-records", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"items": []any{},
					"pagination": gin.H{
						"page":      1,
						"page_size": 20,
					},
				},
			})
		})

		// Admin Management (placeholder - TODO: Implement admin management handler)
		adminGroup.GET("/admins", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"items": []any{},
					"pagination": gin.H{
						"page":      1,
						"page_size": 20,
					},
				},
			})
		})
	}

	// Codex relay routes (OpenAI-compatible API, protected by API key)
	codexGroup := router.Group("/openai")
	codexGroup.Use(middleware.AuthenticateAPIKey(apiKeyRepo, log))
	{
		// Chat Completions API handler (original format)
		chatCompletionsHandler := codex.NewResponsesHandler(codexRelayService, log)
		codexGroup.POST("/chat/completions", chatCompletionsHandler.HandleResponses)
		codexGroup.POST("/v1/chat/completions", chatCompletionsHandler.HandleResponses)

		// Responses API handler (new format with 'input' field)
		responsesAPIHandler := codex.NewResponsesAPIHandler(codexRelayService, log)
		codexGroup.POST("/responses", responsesAPIHandler.HandleResponsesAPI)
		codexGroup.POST("/v1/responses", responsesAPIHandler.HandleResponsesAPI)
	}

	log.Info("Routes registered successfully")

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Info("Server started", zap.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Stop cleanup service
	cleanupService.Stop()
	log.Info("Cleanup service stopped")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited")
}
