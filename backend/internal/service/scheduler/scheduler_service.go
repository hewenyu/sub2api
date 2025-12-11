package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/metrics"
	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
	"github.com/Wei-Shaw/sub2api/backend/internal/scheduler"
)

// SchedulerService defines the interface for the scheduler service.
type SchedulerService interface {
	SelectCodexAccount(ctx context.Context, apiKey *model.APIKey, sessionHash string) (accountID int64, accountType string, error error)
	GetSessionMapping(ctx context.Context, sessionHash string) (*redis.SessionData, error)
	SetSessionMapping(ctx context.Context, sessionHash string, accountID int64, accountType string, ttl time.Duration) error
	ClearSessionMapping(ctx context.Context, sessionHash string) error
	ExtendSessionTTL(ctx context.Context, sessionHash string, ttl time.Duration) error
	AcquireConcurrencySlot(ctx context.Context, accountID int64, requestID string, leaseSeconds int) error
	ReleaseConcurrencySlot(ctx context.Context, accountID int64, requestID string) error
	GetCurrentConcurrency(ctx context.Context, accountID int64) (int64, error)
	MarkAccountUnavailable(ctx context.Context, accountID int64, reason string, resetAfter time.Duration) error
	ReportHealthStatus(ctx context.Context, accountID int64, success bool) error
}

type schedulerService struct {
	codexAccountRepo repository.CodexAccountRepository
	sessionRepo      redis.SessionRepository
	concurrencyRepo  redis.ConcurrencyRepository
	healthRepo       redis.HealthRepository
	strategy         scheduler.SelectionStrategy
	sessionTTL       time.Duration
	maxConcurrency   int
	overloadDuration time.Duration
	logger           *zap.Logger
}

// NewSchedulerService creates a new scheduler service.
func NewSchedulerService(
	codexAccountRepo repository.CodexAccountRepository,
	sessionRepo redis.SessionRepository,
	concurrencyRepo redis.ConcurrencyRepository,
	healthRepo redis.HealthRepository,
	strategy scheduler.SelectionStrategy,
	sessionTTL time.Duration,
	maxConcurrency int,
	logger *zap.Logger,
) SchedulerService {
	if maxConcurrency < 0 {
		maxConcurrency = 0
	}

	const defaultOverloadDuration = 1 * time.Minute

	return &schedulerService{
		codexAccountRepo: codexAccountRepo,
		sessionRepo:      sessionRepo,
		concurrencyRepo:  concurrencyRepo,
		healthRepo:       healthRepo,
		strategy:         strategy,
		sessionTTL:       sessionTTL,
		maxConcurrency:   maxConcurrency,
		overloadDuration: defaultOverloadDuration,
		logger:           logger,
	}
}

// SelectCodexAccount selects a Codex account for the given API key and session.
// Selection logic:
//  1. Check if API key has a bound account (skip scheduling)
//  2. Check sticky session (return cached account)
//  3. Select from shared pool using selection strategy
func (s *schedulerService) SelectCodexAccount(ctx context.Context, apiKey *model.APIKey, sessionHash string) (int64, string, error) {
	start := time.Now()
	strategyName := "default"
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.SchedulerSelectionDuration.WithLabelValues(strategyName).Observe(duration)
	}()

	// 1. Check for bound account
	if apiKey.BoundCodexAccountID != nil && *apiKey.BoundCodexAccountID > 0 {
		metrics.SchedulerSelectionsTotal.WithLabelValues(strategyName, "bound").Inc()
		accountID := *apiKey.BoundCodexAccountID
		account, err := s.codexAccountRepo.GetByID(ctx, accountID)
		if err != nil {
			return 0, "", fmt.Errorf("bound codex account not found: %w", err)
		}

		if !account.IsActive {
			return 0, "", fmt.Errorf("bound codex account is inactive")
		}

		s.logger.Info("Using bound codex account",
			zap.Int64("api_key_id", apiKey.ID),
			zap.Int64("account_id", accountID),
			zap.String("account_type", account.AccountType),
		)

		return accountID, account.AccountType, nil
	}

	// 2. Check sticky session
	if sessionHash != "" {
		sessionData, err := s.sessionRepo.Get(ctx, sessionHash)
		if err == nil && sessionData != nil {
			metrics.SchedulerSelectionsTotal.WithLabelValues(strategyName, "sticky").Inc()
			// Verify account is still valid
			account, err := s.codexAccountRepo.GetByID(ctx, sessionData.AccountID)
			if err == nil && account.IsActive && account.Schedulable {
				// Check if session TTL needs renewal
				ttl, _ := s.sessionRepo.GetTTL(ctx, sessionHash)
				if ttl < 10*time.Minute {
					if extendErr := s.ExtendSessionTTL(ctx, sessionHash, s.sessionTTL); extendErr != nil {
						s.logger.Warn("Failed to extend session TTL",
							zap.String("session_hash", sessionHash),
							zap.Error(extendErr),
						)
					} else {
						s.logger.Debug("Session TTL extended",
							zap.String("session_hash", sessionHash),
							zap.Duration("new_ttl", s.sessionTTL),
						)
					}
				}

				s.logger.Info("Using sticky session",
					zap.Int64("api_key_id", apiKey.ID),
					zap.String("session_hash", sessionHash),
					zap.Int64("account_id", sessionData.AccountID),
					zap.String("account_type", sessionData.AccountType),
				)

				return sessionData.AccountID, sessionData.AccountType, nil
			} else {
				// Account no longer valid, clear session
				if clearErr := s.ClearSessionMapping(ctx, sessionHash); clearErr != nil {
					s.logger.Error("Failed to clear session mapping",
						zap.String("session_hash", sessionHash),
						zap.Error(clearErr),
					)
				}
				s.logger.Warn("Sticky session cleared due to invalid account",
					zap.String("session_hash", sessionHash),
					zap.Int64("account_id", sessionData.AccountID),
				)
			}
		}
	}

	// 3. Select from shared pool
	metrics.SchedulerSelectionsTotal.WithLabelValues(strategyName, "pool").Inc()
	candidates, err := s.codexAccountRepo.GetSchedulable(ctx)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get schedulable codex accounts: %w", err)
	}

	// Filter out quarantined accounts
	var healthyCandidates []*model.CodexAccount
	for _, candidate := range candidates {
		quarantined, checkErr := s.healthRepo.IsQuarantined(ctx, candidate.ID)
		if checkErr != nil {
			s.logger.Warn("Failed to check quarantine status",
				zap.Int64("account_id", candidate.ID),
				zap.Error(checkErr),
			)
			continue
		}
		if !quarantined {
			healthyCandidates = append(healthyCandidates, candidate)
		}
	}

	if len(healthyCandidates) == 0 {
		return 0, "", fmt.Errorf("no available codex accounts")
	}

	// Use selection strategy
	selectionCtx := scheduler.SelectionContext{
		APIKey:      apiKey,
		SessionHash: sessionHash,
	}

	selected, err := s.strategy.Select(ctx, healthyCandidates, selectionCtx)
	if err != nil {
		return 0, "", err
	}

	// Establish session mapping if session hash provided
	if sessionHash != "" {
		if setErr := s.SetSessionMapping(ctx, sessionHash, selected.ID, selected.AccountType, s.sessionTTL); setErr != nil {
			s.logger.Warn("Failed to set session mapping",
				zap.String("session_hash", sessionHash),
				zap.Int64("account_id", selected.ID),
				zap.Error(setErr),
			)
		}
	}

	s.logger.Info("Selected codex account from pool",
		zap.Int64("api_key_id", apiKey.ID),
		zap.Int64("account_id", selected.ID),
		zap.String("account_type", selected.AccountType),
		zap.Int("priority", selected.Priority),
	)

	return selected.ID, selected.AccountType, nil
}

// GetSessionMapping retrieves session mapping data.
func (s *schedulerService) GetSessionMapping(ctx context.Context, sessionHash string) (*redis.SessionData, error) {
	return s.sessionRepo.Get(ctx, sessionHash)
}

// SetSessionMapping creates or updates a session mapping.
func (s *schedulerService) SetSessionMapping(ctx context.Context, sessionHash string, accountID int64, accountType string, ttl time.Duration) error {
	now := time.Now().Unix()
	sessionData := redis.SessionData{
		AccountID:   accountID,
		AccountType: accountType,
		CreatedAt:   now,
		LastUsedAt:  now,
	}

	if err := s.sessionRepo.Set(ctx, sessionHash, sessionData, ttl); err != nil {
		return fmt.Errorf("failed to set session mapping: %w", err)
	}

	s.logger.Debug("Session mapping set",
		zap.String("session_hash", sessionHash),
		zap.Int64("account_id", accountID),
		zap.String("account_type", accountType),
		zap.Duration("ttl", ttl),
	)

	return nil
}

// ClearSessionMapping removes a session mapping.
func (s *schedulerService) ClearSessionMapping(ctx context.Context, sessionHash string) error {
	if err := s.sessionRepo.Delete(ctx, sessionHash); err != nil {
		return fmt.Errorf("failed to clear session mapping: %w", err)
	}

	s.logger.Debug("Session mapping cleared", zap.String("session_hash", sessionHash))
	return nil
}

// ExtendSessionTTL extends the TTL of a session.
func (s *schedulerService) ExtendSessionTTL(ctx context.Context, sessionHash string, ttl time.Duration) error {
	return s.sessionRepo.ExtendTTL(ctx, sessionHash, ttl)
}

// AcquireConcurrencySlot acquires a concurrency slot for a request.
func (s *schedulerService) AcquireConcurrencySlot(ctx context.Context, accountID int64, requestID string, leaseSeconds int) error {
	var (
		acquired bool
		err      error
	)

	if s.maxConcurrency > 0 {
		acquired, err = s.concurrencyRepo.AcquireWithLimit(ctx, accountID, requestID, s.maxConcurrency, leaseSeconds)
	} else {
		acquired, err = s.concurrencyRepo.Acquire(ctx, accountID, requestID, leaseSeconds)
	}

	if err != nil {
		return fmt.Errorf("failed to acquire concurrency slot: %w", err)
	}

	if s.maxConcurrency > 0 && !acquired {
		overloadUntil := time.Now().Add(s.overloadDuration)
		updates := map[string]any{
			"overload_until": overloadUntil,
		}

		if updateErr := s.codexAccountRepo.UpdateFields(ctx, accountID, updates); updateErr != nil {
			s.logger.Warn("Failed to mark account overloaded after concurrency limit reached",
				zap.Int64("account_id", accountID),
				zap.Int("max_concurrent_requests", s.maxConcurrency),
				zap.Error(updateErr),
			)
		} else {
			s.logger.Warn("Account concurrency limit reached, marking as overloaded",
				zap.Int64("account_id", accountID),
				zap.Int("max_concurrent_requests", s.maxConcurrency),
				zap.Time("overload_until", overloadUntil),
			)
		}

		metrics.ConcurrencyLimit.WithLabelValues("account", strconv.FormatInt(accountID, 10)).
			Set(float64(s.maxConcurrency))

		return fmt.Errorf("account concurrency limit exceeded")
	}

	s.logger.Debug("Concurrency slot acquired",
		zap.Int64("account_id", accountID),
		zap.String("request_id", requestID),
		zap.Int("lease_seconds", leaseSeconds),
	)

	return nil
}

// ReleaseConcurrencySlot releases a concurrency slot.
func (s *schedulerService) ReleaseConcurrencySlot(ctx context.Context, accountID int64, requestID string) error {
	if err := s.concurrencyRepo.Release(ctx, accountID, requestID); err != nil {
		s.logger.Warn("Failed to release concurrency slot",
			zap.Error(err),
			zap.Int64("account_id", accountID),
			zap.String("request_id", requestID),
		)
		return err
	}

	s.logger.Debug("Concurrency slot released",
		zap.Int64("account_id", accountID),
		zap.String("request_id", requestID),
	)

	return nil
}

// GetCurrentConcurrency returns the current concurrency count for an account.
func (s *schedulerService) GetCurrentConcurrency(ctx context.Context, accountID int64) (int64, error) {
	return s.concurrencyRepo.GetCount(ctx, accountID)
}

// MarkAccountUnavailable marks an account as temporarily unavailable.
func (s *schedulerService) MarkAccountUnavailable(ctx context.Context, accountID int64, reason string, resetAfter time.Duration) error {
	resetAt := time.Now().Add(resetAfter)

	updates := map[string]any{
		"rate_limited_until": resetAt,
		"rate_limit_status":  reason,
	}

	if err := s.codexAccountRepo.UpdateFields(ctx, accountID, updates); err != nil {
		return fmt.Errorf("failed to mark account unavailable: %w", err)
	}

	s.logger.Warn("Account marked unavailable",
		zap.Int64("account_id", accountID),
		zap.String("reason", reason),
		zap.Time("reset_at", resetAt),
	)

	return nil
}

// ReportHealthStatus reports the health status of an account after a request.
func (s *schedulerService) ReportHealthStatus(ctx context.Context, accountID int64, success bool) error {
	if err := s.healthRepo.UpdateMetrics(ctx, accountID, success); err != nil {
		return fmt.Errorf("failed to update health metrics: %w", err)
	}

	healthMetrics, err := s.healthRepo.GetMetrics(ctx, accountID)
	if err == nil {
		metrics.AccountHealthScore.WithLabelValues(strconv.FormatInt(accountID, 10)).Set(healthMetrics.HealthScore)
	}

	if !success {
		if err != nil {
			return fmt.Errorf("failed to get health metrics: %w", err)
		}

		if redis.ShouldQuarantine(healthMetrics) {
			duration := redis.GetQuarantineDuration(healthMetrics.QuarantineCount)
			if err := s.healthRepo.SetQuarantine(ctx, accountID, duration); err != nil {
				return fmt.Errorf("failed to set quarantine: %w", err)
			}

			metrics.AccountQuarantineTotal.WithLabelValues(strconv.FormatInt(accountID, 10), "health_check_failed").Inc()

			s.logger.Warn("Account quarantined after request failure",
				zap.Int64("account_id", accountID),
				zap.Duration("duration", duration),
				zap.Float64("health_score", healthMetrics.HealthScore),
			)
		}
	}

	return nil
}
