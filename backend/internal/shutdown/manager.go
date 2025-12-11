package shutdown

import (
	"context"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/internal/service/scheduler"
)

type HealthChecker interface {
	Stop()
}

type Manager struct {
	httpServer      *http.Server
	cleanupService  scheduler.CleanupService
	healthChecker   HealthChecker
	db              *gorm.DB
	redisClient     *redis.Client
	requestTracker  *RequestTracker
	shutdownTimeout time.Duration
	logger          *zap.Logger
}

func NewManager(
	httpServer *http.Server,
	cleanupService scheduler.CleanupService,
	healthChecker HealthChecker,
	db *gorm.DB,
	redisClient *redis.Client,
	requestTracker *RequestTracker,
	shutdownTimeout time.Duration,
	logger *zap.Logger,
) *Manager {
	return &Manager{
		httpServer:      httpServer,
		cleanupService:  cleanupService,
		healthChecker:   healthChecker,
		db:              db,
		redisClient:     redisClient,
		requestTracker:  requestTracker,
		shutdownTimeout: shutdownTimeout,
		logger:          logger,
	}
}

func (m *Manager) Shutdown(ctx context.Context) error {
	m.logger.Info("Starting graceful shutdown")

	shutdownCtx, cancel := context.WithTimeout(ctx, m.shutdownTimeout)
	defer cancel()

	m.logger.Info("Phase 1: Stopping HTTP server")
	if err := m.httpServer.Shutdown(shutdownCtx); err != nil {
		m.logger.Error("HTTP server shutdown failed", zap.Error(err))
	}

	m.logger.Info("Phase 2: Waiting for in-flight requests",
		zap.Int("active_requests", m.requestTracker.ActiveCount()),
	)
	if err := m.requestTracker.Wait(shutdownCtx); err != nil {
		m.logger.Warn("In-flight requests did not complete in time",
			zap.Error(err),
			zap.Int("remaining", m.requestTracker.ActiveCount()),
		)
	}

	m.logger.Info("Phase 3: Stopping background workers")
	if m.cleanupService != nil {
		m.cleanupService.Stop()
	}
	if m.healthChecker != nil {
		m.healthChecker.Stop()
	}

	m.logger.Info("Phase 4: Closing connections")
	if m.redisClient != nil {
		if err := m.redisClient.Close(); err != nil {
			m.logger.Error("Redis close failed", zap.Error(err))
		}
	}

	if m.db != nil {
		sqlDB, err := m.db.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				m.logger.Error("Database close failed", zap.Error(err))
			}
		}
	}

	m.logger.Info("Graceful shutdown completed")
	return nil
}
