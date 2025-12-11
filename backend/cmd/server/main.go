package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/config"
	"github.com/Wei-Shaw/sub2api/backend/internal/api/admin"
	"github.com/Wei-Shaw/sub2api/backend/internal/api/codex"
	"github.com/Wei-Shaw/sub2api/backend/internal/middleware"
	"github.com/Wei-Shaw/sub2api/backend/internal/shutdown"
	"github.com/Wei-Shaw/sub2api/backend/pkg/database"
	"github.com/Wei-Shaw/sub2api/backend/pkg/logger"
)

var (
	configPath = flag.String("config", "config/config.yaml", "Path to config file")
)

func main() {
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	validator := config.NewValidator()
	if validateErr := validator.Validate(cfg); validateErr != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", validateErr)
		os.Exit(1)
	}

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

	db, err := database.NewPostgresDB(&cfg.Database, log)
	if err != nil {
		log.Fatal("Failed to initialize database", zap.Error(err))
	}
	log.Info("Database initialized successfully")

	redisClient, err := database.NewRedisClient(&cfg.Redis, log)
	if err != nil {
		log.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	log.Info("Redis initialized successfully")

	services, err := InitializeServices(cfg, log, db, redisClient.Client)
	if err != nil {
		log.Fatal("Failed to initialize services", zap.Error(err))
	}

	if err := services.PricingService.LoadPricing(); err != nil {
		log.Fatal("Failed to load pricing data", zap.Error(err))
	}

	ctx := context.Background()
	services.CleanupService.Start(ctx)
	log.Info("Cleanup service started", zap.Duration("interval", 5*time.Minute))

	services.HealthChecker.Start(ctx)
	log.Info("Health checker started", zap.Duration("interval", 2*time.Minute))

	log.Info("All services initialized successfully")

	configWatcher := config.NewWatcher(
		*configPath,
		cfg,
		validator,
		func(newConfig *config.Config) {
			log.Info("Configuration updated",
				zap.String("server_mode", newConfig.Server.Mode),
				zap.String("log_level", newConfig.Logging.Level),
			)
		},
		log,
	)
	configWatcher.Start()
	defer configWatcher.Stop()
	log.Info("Config watcher started", zap.String("config_path", *configPath))

	requestTracker := shutdown.NewRequestTracker()

	gin.DefaultWriter = logger.NewGinWriter(log)
	gin.DefaultErrorWriter = logger.NewGinWriter(log)
	gin.SetMode(cfg.Server.Mode)

	router := gin.New()
	router.Use(middleware.RequestTrackerMiddleware(requestTracker))

	loggingMiddleware := middleware.NewLoggingMiddleware(log)
	router.Use(loggingMiddleware.Handler())
	router.Use(middleware.ZapRecovery(log))

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

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	apiV1 := router.Group("/api/v1")

	loginHandler := admin.NewLoginHandler(services.AdminService, services.AdminRepo, log)
	apiV1.POST("/admin/login", loginHandler.Login)

	adminGroup := apiV1.Group("/admin")
	adminGroup.Use(middleware.AuthenticateAdmin(services.AdminService, log))
	{
		adminGroup.GET("/info", loginHandler.GetInfo)

		codexAccountHandler := admin.NewCodexAccountHandler(services.CodexAccountService, log)
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

		apiKeyHandler := admin.NewAPIKeyHandler(services.APIKeyRepo, log)
		adminGroup.POST("/api-keys", apiKeyHandler.CreateAPIKey)
		adminGroup.GET("/api-keys", apiKeyHandler.ListAPIKeys)
		adminGroup.GET("/api-keys/:id", apiKeyHandler.GetAPIKey)
		adminGroup.PUT("/api-keys/:id", apiKeyHandler.UpdateAPIKey)
		adminGroup.DELETE("/api-keys/:id", apiKeyHandler.DeleteAPIKey)
		adminGroup.PATCH("/api-keys/:id/toggle", apiKeyHandler.ToggleAPIKeyStatus)

		proxyHandler := admin.NewProxyHandler(services.ProxyConfigService, log)
		adminGroup.GET("/proxies", proxyHandler.ListProxies)
		adminGroup.POST("/proxies", proxyHandler.CreateProxy)
		adminGroup.GET("/proxies/names", proxyHandler.GetProxyNames)
		adminGroup.GET("/proxies/:id", proxyHandler.GetProxy)
		adminGroup.PUT("/proxies/:id", proxyHandler.UpdateProxy)
		adminGroup.DELETE("/proxies/:id", proxyHandler.DeleteProxy)
		adminGroup.POST("/proxies/:id/set-default", proxyHandler.SetDefaultProxy)
		adminGroup.POST("/proxies/:id/test", proxyHandler.TestProxy)

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

	codexGroup := router.Group("/openai")
	codexGroup.Use(middleware.AuthenticateAPIKey(services.APIKeyRepo, log))
	codexGroup.Use(middleware.RateLimitMiddleware(services.RateLimiter, log))
	codexGroup.Use(middleware.ConcurrencyLimitMiddleware(services.ConcurrencyTracker, log))
	codexGroup.Use(middleware.CostLimitMiddleware(services.CostLimiter, log))
	{
		chatCompletionsHandler := codex.NewResponsesHandler(services.CodexRelayService, log)
		codexGroup.POST("/chat/completions", chatCompletionsHandler.HandleResponses)
		codexGroup.POST("/v1/chat/completions", chatCompletionsHandler.HandleResponses)

		responsesAPIHandler := codex.NewResponsesAPIHandler(services.CodexRelayService, log)
		codexGroup.POST("/responses", responsesAPIHandler.HandleResponsesAPI)
		codexGroup.POST("/v1/responses", responsesAPIHandler.HandleResponsesAPI)
	}

	log.Info("Routes registered successfully")

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	shutdownTimeout := 30 * time.Second
	if timeout := os.Getenv("SHUTDOWN_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			shutdownTimeout = d
		}
	}

	shutdownManager := shutdown.NewManager(
		server,
		services.CleanupService,
		services.HealthChecker,
		db.DB,
		redisClient.Client,
		requestTracker,
		shutdownTimeout,
		log,
	)

	go func() {
		log.Info("Server started", zap.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	reload := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(reload, syscall.SIGHUP)

	for {
		select {
		case sig := <-quit:
			log.Info("Received shutdown signal", zap.String("signal", sig.String()))
			goto shutdown

		case sig := <-reload:
			configWatcher.ReloadOnSignal(sig)
		}
	}

shutdown:
	log.Info("Initiating graceful shutdown")
	requestTracker.BeginShutdown()

	if err := shutdownManager.Shutdown(context.Background()); err != nil {
		log.Error("Graceful shutdown failed", zap.Error(err))
		os.Exit(1)
	}

	log.Info("Server stopped")
}
