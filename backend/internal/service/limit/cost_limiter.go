package limit

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
)

// CostLimiter manages cost-based limits for API keys.
type CostLimiter interface {
	CheckDailyLimit(ctx context.Context, apiKeyID int64, limit float64) (allowed bool, current float64, err error)
	CheckWeeklyLimit(ctx context.Context, apiKeyID int64, limit float64) (allowed bool, current float64, err error)
	CheckMonthlyLimit(ctx context.Context, apiKeyID int64, limit float64) (allowed bool, current float64, err error)
	CheckTotalLimit(ctx context.Context, apiKeyID int64, limit float64) (allowed bool, current float64, err error)
}

type costLimiter struct {
	usageRepo repository.UsageRepository
	logger    *zap.Logger
}

// NewCostLimiter creates a new cost limiter.
func NewCostLimiter(
	usageRepo repository.UsageRepository,
	logger *zap.Logger,
) CostLimiter {
	return &costLimiter{
		usageRepo: usageRepo,
		logger:    logger,
	}
}

// CheckDailyLimit checks the daily cost limit.
func (l *costLimiter) CheckDailyLimit(ctx context.Context, apiKeyID int64, limit float64) (bool, float64, error) {
	if limit <= 0 {
		return true, 0, nil
	}

	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endDate := startDate.Add(24 * time.Hour)

	filters := repository.UsageFilters{
		APIKeyID:  &apiKeyID,
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	aggregate, err := l.usageRepo.Aggregate(ctx, filters)
	if err != nil {
		return false, 0, fmt.Errorf("failed to aggregate daily cost: %w", err)
	}

	current := aggregate.TotalCost

	if current >= limit {
		l.logger.Warn("Daily cost limit exceeded",
			zap.Int64("api_key_id", apiKeyID),
			zap.Float64("current", current),
			zap.Float64("limit", limit),
		)
		return false, current, nil
	}

	return true, current, nil
}

// CheckWeeklyLimit checks the weekly cost limit.
func (l *costLimiter) CheckWeeklyLimit(ctx context.Context, apiKeyID int64, limit float64) (bool, float64, error) {
	if limit <= 0 {
		return true, 0, nil
	}

	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	daysFromMonday := weekday - 1
	startDate := now.AddDate(0, 0, -daysFromMonday)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endDate := startDate.Add(7 * 24 * time.Hour)

	filters := repository.UsageFilters{
		APIKeyID:  &apiKeyID,
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	aggregate, err := l.usageRepo.Aggregate(ctx, filters)
	if err != nil {
		return false, 0, fmt.Errorf("failed to aggregate weekly cost: %w", err)
	}

	current := aggregate.TotalCost

	if current >= limit {
		l.logger.Warn("Weekly cost limit exceeded",
			zap.Int64("api_key_id", apiKeyID),
			zap.Float64("current", current),
			zap.Float64("limit", limit),
		)
		return false, current, nil
	}

	return true, current, nil
}

// CheckMonthlyLimit checks the monthly cost limit.
func (l *costLimiter) CheckMonthlyLimit(ctx context.Context, apiKeyID int64, limit float64) (bool, float64, error) {
	if limit <= 0 {
		return true, 0, nil
	}

	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endDate := startDate.AddDate(0, 1, 0)

	filters := repository.UsageFilters{
		APIKeyID:  &apiKeyID,
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	aggregate, err := l.usageRepo.Aggregate(ctx, filters)
	if err != nil {
		return false, 0, fmt.Errorf("failed to aggregate monthly cost: %w", err)
	}

	current := aggregate.TotalCost

	if current >= limit {
		l.logger.Warn("Monthly cost limit exceeded",
			zap.Int64("api_key_id", apiKeyID),
			zap.Float64("current", current),
			zap.Float64("limit", limit),
		)
		return false, current, nil
	}

	return true, current, nil
}

// CheckTotalLimit checks the total cost limit.
func (l *costLimiter) CheckTotalLimit(ctx context.Context, apiKeyID int64, limit float64) (bool, float64, error) {
	if limit <= 0 {
		return true, 0, nil
	}

	filters := repository.UsageFilters{
		APIKeyID: &apiKeyID,
	}

	aggregate, err := l.usageRepo.Aggregate(ctx, filters)
	if err != nil {
		return false, 0, fmt.Errorf("failed to aggregate total cost: %w", err)
	}

	current := aggregate.TotalCost

	if current >= limit {
		l.logger.Warn("Total cost limit exceeded",
			zap.Int64("api_key_id", apiKeyID),
			zap.Float64("current", current),
			zap.Float64("limit", limit),
		)
		return false, current, nil
	}

	return true, current, nil
}
