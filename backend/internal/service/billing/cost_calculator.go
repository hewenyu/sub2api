package billing

import (
	"fmt"

	"go.uber.org/zap"
)

// CostCalculator calculates costs based on token usage and pricing.
type CostCalculator interface {
	Calculate(usage *UsageData, model string) (CostInfo, error)
	CalculateTotalCost(inputTokens, outputTokens int, model string) (float64, error)
}

type costCalculator struct {
	pricingSvc PricingService
	logger     *zap.Logger
}

// UsageData represents token usage information.
// CacheCreateTokens / CacheReadTokens are optional and are used to
// account for prompt caching where supported by the upstream API.
type UsageData struct {
	InputTokens       int
	OutputTokens      int
	TotalTokens       int
	CacheCreateTokens int
	CacheReadTokens   int
}

// CostInfo represents calculated cost information.
type CostInfo struct {
	InputCost       float64 `json:"input_cost"`
	OutputCost      float64 `json:"output_cost"`
	CacheCreateCost float64 `json:"cache_create_cost"`
	CacheReadCost   float64 `json:"cache_read_cost"`
	TotalCost       float64 `json:"total_cost"`
}

// NewCostCalculator creates a new cost calculator.
func NewCostCalculator(pricingSvc PricingService, logger *zap.Logger) CostCalculator {
	return &costCalculator{
		pricingSvc: pricingSvc,
		logger:     logger,
	}
}

// Calculate computes the cost for the given usage and model.
// Cost calculation: (tokens / 1,000,000) * price_per_million
func (c *costCalculator) Calculate(usage *UsageData, model string) (CostInfo, error) {
	if usage == nil {
		return CostInfo{}, fmt.Errorf("usage data cannot be nil")
	}

	pricing, err := c.pricingSvc.GetPricing(model)
	if err != nil {
		return CostInfo{}, fmt.Errorf("failed to get pricing for model %s: %w", model, err)
	}

	// Calculate costs (price is per million tokens)
	inputCost := float64(usage.InputTokens) / 1_000_000.0 * pricing.Input
	outputCost := float64(usage.OutputTokens) / 1_000_000.0 * pricing.Output

	// Cache creation tokens are typically billed at the same rate as input
	// unless an explicit cache creation price is provided.
	cacheCreatePrice := pricing.CacheCreate
	if cacheCreatePrice == 0 {
		cacheCreatePrice = pricing.Input
	}
	cacheCreateCost := float64(usage.CacheCreateTokens) / 1_000_000.0 * cacheCreatePrice

	// Cache read tokens may have a discounted price. If not configured,
	// fall back to the normal input price so that cached reads are still
	// billed instead of being treated as free.
	cacheReadPrice := pricing.CacheRead
	if cacheReadPrice == 0 {
		cacheReadPrice = pricing.Input
	}
	cacheReadCost := float64(usage.CacheReadTokens) / 1_000_000.0 * cacheReadPrice

	totalCost := inputCost + outputCost + cacheCreateCost + cacheReadCost

	c.logger.Debug("Cost calculated",
		zap.String("model", model),
		zap.Int("input_tokens", usage.InputTokens),
		zap.Int("output_tokens", usage.OutputTokens),
		zap.Int("cache_create_tokens", usage.CacheCreateTokens),
		zap.Int("cache_read_tokens", usage.CacheReadTokens),
		zap.Float64("input_cost", inputCost),
		zap.Float64("output_cost", outputCost),
		zap.Float64("cache_create_cost", cacheCreateCost),
		zap.Float64("cache_read_cost", cacheReadCost),
		zap.Float64("total_cost", totalCost),
	)

	return CostInfo{
		InputCost:       inputCost,
		OutputCost:      outputCost,
		CacheCreateCost: cacheCreateCost,
		CacheReadCost:   cacheReadCost,
		TotalCost:       totalCost,
	}, nil
}

// CalculateTotalCost is a convenience method to calculate total cost directly.
func (c *costCalculator) CalculateTotalCost(inputTokens, outputTokens int, model string) (float64, error) {
	usage := &UsageData{
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  inputTokens + outputTokens,
	}

	costInfo, err := c.Calculate(usage, model)
	if err != nil {
		return 0, err
	}

	return costInfo.TotalCost, nil
}
