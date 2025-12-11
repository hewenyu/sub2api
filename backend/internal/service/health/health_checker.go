package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
)

type HealthChecker interface {
	Start(ctx context.Context)
	Stop()
	ReportResult(ctx context.Context, accountID int64, success bool) error
}

type healthChecker struct {
	accountRepo   repository.CodexAccountRepository
	healthRepo    redis.HealthRepository
	checkInterval time.Duration
	idleThreshold time.Duration
	logger        *zap.Logger
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

func NewHealthChecker(
	accountRepo repository.CodexAccountRepository,
	healthRepo redis.HealthRepository,
	checkInterval time.Duration,
	logger *zap.Logger,
) HealthChecker {
	return &healthChecker{
		accountRepo:   accountRepo,
		healthRepo:    healthRepo,
		checkInterval: checkInterval,
		idleThreshold: 5 * time.Minute,
		logger:        logger,
		stopCh:        make(chan struct{}),
	}
}

func (h *healthChecker) Start(ctx context.Context) {
	h.wg.Add(1)
	go h.run(ctx)
}

func (h *healthChecker) Stop() {
	close(h.stopCh)
	h.wg.Wait()
}

func (h *healthChecker) run(ctx context.Context) {
	defer h.wg.Done()

	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	h.logger.Info("Health checker started", zap.Duration("interval", h.checkInterval))

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("Health checker stopped due to context cancellation")
			return
		case <-h.stopCh:
			h.logger.Info("Health checker stopped")
			return
		case <-ticker.C:
			h.performHealthChecks(ctx)
		}
	}
}

func (h *healthChecker) performHealthChecks(ctx context.Context) {
	accounts, err := h.accountRepo.GetSchedulable(ctx)
	if err != nil {
		h.logger.Error("Failed to get schedulable accounts", zap.Error(err))
		return
	}

	for _, account := range accounts {
		if err := h.checkAccount(ctx, account); err != nil {
			h.logger.Warn("Health check failed",
				zap.Int64("account_id", account.ID),
				zap.Error(err),
			)
		}
	}
}

func (h *healthChecker) checkAccount(ctx context.Context, account *model.CodexAccount) error {
	lastCheck, err := h.healthRepo.GetLastCheckTime(ctx, account.ID)
	if err != nil {
		return fmt.Errorf("failed to get last check time: %w", err)
	}

	if !lastCheck.IsZero() && time.Since(lastCheck) < h.idleThreshold {
		return nil
	}

	quarantined, err := h.healthRepo.IsQuarantined(ctx, account.ID)
	if err != nil {
		return fmt.Errorf("failed to check quarantine status: %w", err)
	}
	if quarantined {
		return nil
	}

	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	success := h.performCheck(checkCtx, account)

	if updateErr := h.healthRepo.UpdateMetrics(ctx, account.ID, success); updateErr != nil {
		return fmt.Errorf("failed to update metrics: %w", updateErr)
	}

	metrics, err := h.healthRepo.GetMetrics(ctx, account.ID)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	if redis.ShouldQuarantine(metrics) {
		duration := redis.GetQuarantineDuration(metrics.QuarantineCount)
		if err := h.healthRepo.SetQuarantine(ctx, account.ID, duration); err != nil {
			return fmt.Errorf("failed to set quarantine: %w", err)
		}

		h.logger.Warn("Account quarantined",
			zap.Int64("account_id", account.ID),
			zap.Duration("duration", duration),
			zap.Float64("health_score", metrics.HealthScore),
			zap.Int("consecutive_failures", metrics.ConsecutiveFailures),
		)
	}

	return nil
}

func (h *healthChecker) performCheck(ctx context.Context, account *model.CodexAccount) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", account.BaseAPI+"/models", nil)
	if err != nil {
		return false
	}

	if account.AccountType == "openai-responses" && account.APIKey != nil {
		req.Header.Set("Authorization", "Bearer "+*account.APIKey)
	} else if account.AccountType == "openai-oauth" && account.AccessToken != nil {
		req.Header.Set("Authorization", "Bearer "+*account.AccessToken)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}

func (h *healthChecker) ReportResult(ctx context.Context, accountID int64, success bool) error {
	if err := h.healthRepo.UpdateMetrics(ctx, accountID, success); err != nil {
		return fmt.Errorf("failed to report result: %w", err)
	}

	if !success {
		metrics, err := h.healthRepo.GetMetrics(ctx, accountID)
		if err != nil {
			return fmt.Errorf("failed to get metrics: %w", err)
		}

		if redis.ShouldQuarantine(metrics) {
			duration := redis.GetQuarantineDuration(metrics.QuarantineCount)
			if err := h.healthRepo.SetQuarantine(ctx, accountID, duration); err != nil {
				return fmt.Errorf("failed to set quarantine: %w", err)
			}

			h.logger.Warn("Account quarantined after request failure",
				zap.Int64("account_id", accountID),
				zap.Duration("duration", duration),
			)
		}
	}

	return nil
}
