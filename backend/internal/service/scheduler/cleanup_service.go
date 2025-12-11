package scheduler

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
)

// CleanupService handles periodic cleanup of expired rate limits and overload flags.
// This is similar to the rateLimitCleanupService.js in the Node.js implementation.
type CleanupService interface {
	Start(ctx context.Context)
	Stop()
}

type cleanupService struct {
	codexAccountRepo repository.CodexAccountRepository
	logger           *zap.Logger
	ticker           *time.Ticker
	stopChan         chan struct{}
	interval         time.Duration
}

// NewCleanupService creates a new cleanup service.
func NewCleanupService(
	codexAccountRepo repository.CodexAccountRepository,
	logger *zap.Logger,
	interval time.Duration,
) CleanupService {
	if interval == 0 {
		interval = 5 * time.Minute // default to 5 minutes, same as Node.js
	}

	return &cleanupService{
		codexAccountRepo: codexAccountRepo,
		logger:           logger.Named("cleanup_service"),
		interval:         interval,
		stopChan:         make(chan struct{}),
	}
}

// Start begins the periodic cleanup process.
func (s *cleanupService) Start(ctx context.Context) {
	s.ticker = time.NewTicker(s.interval)

	s.logger.Info("Cleanup service started",
		zap.Duration("interval", s.interval),
	)

	// Run cleanup immediately on start
	go s.cleanup(ctx)

	// Then run periodically
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.cleanup(ctx)
			case <-s.stopChan:
				s.logger.Info("Cleanup service stopped")
				return
			case <-ctx.Done():
				s.logger.Info("Cleanup service context cancelled")
				return
			}
		}
	}()
}

// Stop stops the cleanup service.
func (s *cleanupService) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopChan)
}

// cleanup performs the actual cleanup of expired rate limits and overload flags.
func (s *cleanupService) cleanup(ctx context.Context) {
	s.logger.Debug("Running rate limit cleanup")

	now := time.Now().UTC()

	// Find all accounts (we need to check all of them for expired flags)
	// Using a large limit to get all accounts
	accounts, _, err := s.codexAccountRepo.List(ctx, repository.CodexAccountFilters{}, 0, 10000)
	if err != nil {
		s.logger.Error("Failed to fetch accounts for cleanup",
			zap.Error(err),
		)
		return
	}

	clearedCount := 0
	for _, account := range accounts {
		needsUpdate := false
		updates := make(map[string]any)

		// Check if rate_limited_until has expired
		if account.RateLimitedUntil != nil && account.RateLimitedUntil.Before(now) {
			updates["rate_limited_until"] = nil
			updates["rate_limit_status"] = nil
			needsUpdate = true

			email := ""
			if account.Email != nil {
				email = *account.Email
			}
			s.logger.Info("Clearing expired rate limit",
				zap.Int64("account_id", account.ID),
				zap.String("email", email),
				zap.Time("was_limited_until", *account.RateLimitedUntil),
			)
		}

		// Check if overload_until has expired
		if account.OverloadUntil != nil && account.OverloadUntil.Before(now) {
			updates["overload_until"] = nil
			needsUpdate = true

			email := ""
			if account.Email != nil {
				email = *account.Email
			}
			s.logger.Info("Clearing expired overload flag",
				zap.Int64("account_id", account.ID),
				zap.String("email", email),
				zap.Time("was_overloaded_until", *account.OverloadUntil),
			)
		}

		// Update the account if needed
		if needsUpdate {
			if err := s.codexAccountRepo.UpdateFields(ctx, account.ID, updates); err != nil {
				s.logger.Error("Failed to clear expired flags",
					zap.Error(err),
					zap.Int64("account_id", account.ID),
				)
				continue
			}
			clearedCount++
		}
	}

	if clearedCount > 0 {
		s.logger.Info("Cleanup cycle completed",
			zap.Int("cleared_accounts", clearedCount),
		)
	} else {
		s.logger.Debug("Cleanup cycle completed, no accounts needed clearing")
	}
}
