package scheduler

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

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
}

type schedulerService struct {
	codexAccountRepo repository.CodexAccountRepository
	sessionRepo      redis.SessionRepository
	concurrencyRepo  redis.ConcurrencyRepository
	strategy         scheduler.SelectionStrategy
	sessionTTL       time.Duration
	logger           *zap.Logger
}

// NewSchedulerService creates a new scheduler service.
func NewSchedulerService(
	codexAccountRepo repository.CodexAccountRepository,
	sessionRepo redis.SessionRepository,
	concurrencyRepo redis.ConcurrencyRepository,
	strategy scheduler.SelectionStrategy,
	sessionTTL time.Duration,
	logger *zap.Logger,
) SchedulerService {
	return &schedulerService{
		codexAccountRepo: codexAccountRepo,
		sessionRepo:      sessionRepo,
		concurrencyRepo:  concurrencyRepo,
		strategy:         strategy,
		sessionTTL:       sessionTTL,
		logger:           logger,
	}
}

// SelectCodexAccount selects a Codex account for the given API key and session.
// Selection logic:
//  1. Check if API key has a bound account (skip scheduling)
//  2. Check sticky session (return cached account)
//  3. Select from shared pool using selection strategy
func (s *schedulerService) SelectCodexAccount(ctx context.Context, apiKey *model.APIKey, sessionHash string) (int64, string, error) {
	// 1. Check for bound account
	if apiKey.BoundCodexAccountID != nil && *apiKey.BoundCodexAccountID > 0 {
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
	candidates, err := s.codexAccountRepo.GetSchedulable(ctx)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get schedulable codex accounts: %w", err)
	}

	if len(candidates) == 0 {
		return 0, "", fmt.Errorf("no available codex accounts")
	}

	// Use selection strategy
	selectionCtx := scheduler.SelectionContext{
		APIKey:      apiKey,
		SessionHash: sessionHash,
	}

	selected, err := s.strategy.Select(ctx, candidates, selectionCtx)
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
	_, err := s.concurrencyRepo.Acquire(ctx, accountID, requestID, leaseSeconds)
	if err != nil {
		return fmt.Errorf("failed to acquire concurrency slot: %w", err)
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
