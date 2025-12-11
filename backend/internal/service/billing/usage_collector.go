package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
)

// UsageCollector collects and stores usage data with cost calculation.
type UsageCollector interface {
	CollectUsage(ctx context.Context, apiKeyID, accountID int64, accountType string, usage *UsageData, model, requestID, conversationID string) error
	GetDailyCost(ctx context.Context, apiKeyID int64, date time.Time) (float64, error)
	GetWeeklyCost(ctx context.Context, apiKeyID int64, startDate time.Time) (float64, error)
	GetMonthlyCost(ctx context.Context, apiKeyID int64, year int, month time.Month) (float64, error)
	GetAggregate(ctx context.Context, apiKeyID int64, startTime, endTime time.Time) (*model.UsageAggregate, error)
}

type usageCollector struct {
	usageRepo repository.UsageRepository
	costCalc  CostCalculator
	logger    *zap.Logger
}

// NewUsageCollector creates a new usage collector.
func NewUsageCollector(
	usageRepo repository.UsageRepository,
	costCalc CostCalculator,
	logger *zap.Logger,
) UsageCollector {
	return &usageCollector{
		usageRepo: usageRepo,
		costCalc:  costCalc,
		logger:    logger,
	}
}

// CollectUsage records usage data with cost calculation.
func (u *usageCollector) CollectUsage(
	ctx context.Context,
	apiKeyID, accountID int64,
	accountType string,
	usage *UsageData,
	modelName, requestID, conversationID string,
) error {
	if usage == nil {
		return fmt.Errorf("usage data cannot be nil")
	}

	// Calculate cost
	costInfo, err := u.costCalc.Calculate(usage, modelName)
	if err != nil {
		u.logger.Warn("Failed to calculate cost, using zero cost",
			zap.Error(err),
			zap.String("model", modelName),
		)
		// Use zero cost if calculation fails
		costInfo = CostInfo{
			InputCost:  0,
			OutputCost: 0,
			TotalCost:  0,
		}
	}

	// Build request metadata JSON. This keeps jsonb columns valid even if
	// we don't have rich metadata yet.
	requestMeta := map[string]any{
		"request_id":      requestID,
		"conversation_id": conversationID,
		"model":           modelName,
		"account_type":    accountType,
	}

	requestMetaJSON, err := json.Marshal(requestMeta)
	if err != nil {
		u.logger.Warn("Failed to marshal request metadata, using empty object",
			zap.Error(err),
			zap.String("model", modelName),
		)
		requestMetaJSON = []byte("{}")
	}

	// Create usage record
	usageRecord := &model.Usage{
		APIKeyID:                 apiKeyID,
		Type:                     model.UsageTypeCodex,
		AccountID:                accountID,
		Model:                    modelName,
		InputTokens:              int64(usage.InputTokens),
		OutputTokens:             int64(usage.OutputTokens),
		TotalTokens:              int64(usage.TotalTokens),
		CacheCreationInputTokens: int64(usage.CacheCreateTokens),
		CacheReadInputTokens:     int64(usage.CacheReadTokens),
		Cost:                     costInfo.TotalCost,
		StatusCode:               200, // Default to success
		RequestMetadata:          string(requestMetaJSON),
		ResponseMetadata:         "{}", // Placeholder; can be enriched later
		CreatedAt:                time.Now(),
	}

	// Save to database
	if err := u.usageRepo.Create(ctx, usageRecord); err != nil {
		return fmt.Errorf("failed to create usage record: %w", err)
	}

	u.logger.Info("Usage collected successfully",
		zap.Int64("api_key_id", apiKeyID),
		zap.Int64("account_id", accountID),
		zap.String("account_type", accountType),
		zap.String("model", modelName),
		zap.Int("input_tokens", usage.InputTokens),
		zap.Int("output_tokens", usage.OutputTokens),
		zap.Float64("input_cost", costInfo.InputCost),
		zap.Float64("output_cost", costInfo.OutputCost),
		zap.Float64("total_cost", costInfo.TotalCost),
	)

	return nil
}

// GetDailyCost retrieves the total cost for a specific day.
func (u *usageCollector) GetDailyCost(ctx context.Context, apiKeyID int64, date time.Time) (float64, error) {
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endDate := startDate.Add(24 * time.Hour)

	usageType := model.UsageTypeCodex
	filters := repository.UsageFilters{
		APIKeyID:  &apiKeyID,
		UsageType: &usageType,
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	aggregate, err := u.usageRepo.Aggregate(ctx, filters)
	if err != nil {
		return 0, fmt.Errorf("failed to get daily cost: %w", err)
	}

	return aggregate.TotalCost, nil
}

// GetWeeklyCost retrieves the total cost for a week starting from the given date.
func (u *usageCollector) GetWeeklyCost(ctx context.Context, apiKeyID int64, startDate time.Time) (float64, error) {
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endDate := start.Add(7 * 24 * time.Hour)

	usageType := model.UsageTypeCodex
	filters := repository.UsageFilters{
		APIKeyID:  &apiKeyID,
		UsageType: &usageType,
		StartDate: &start,
		EndDate:   &endDate,
	}
	aggregate, err := u.usageRepo.Aggregate(ctx, filters)
	if err != nil {
		return 0, fmt.Errorf("failed to get weekly cost: %w", err)
	}

	return aggregate.TotalCost, nil
}

// GetMonthlyCost retrieves the total cost for a specific month.
func (u *usageCollector) GetMonthlyCost(ctx context.Context, apiKeyID int64, year int, month time.Month) (float64, error) {
	location := time.Now().Location()
	startDate := time.Date(year, month, 1, 0, 0, 0, 0, location)
	endDate := startDate.AddDate(0, 1, 0) // First day of next month

	usageType := model.UsageTypeCodex
	filters := repository.UsageFilters{
		APIKeyID:  &apiKeyID,
		UsageType: &usageType,
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	aggregate, err := u.usageRepo.Aggregate(ctx, filters)
	if err != nil {
		return 0, fmt.Errorf("failed to get monthly cost: %w", err)
	}

	return aggregate.TotalCost, nil
}

// GetAggregate retrieves aggregated usage statistics for a time range.
func (u *usageCollector) GetAggregate(ctx context.Context, apiKeyID int64, startTime, endTime time.Time) (*model.UsageAggregate, error) {
	usageType := model.UsageTypeCodex
	filters := repository.UsageFilters{
		APIKeyID:  &apiKeyID,
		UsageType: &usageType,
		StartDate: &startTime,
		EndDate:   &endTime,
	}
	aggregate, err := u.usageRepo.Aggregate(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregate: %w", err)
	}

	return aggregate, nil
}
