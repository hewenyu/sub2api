package billing

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

// PricingService manages model pricing data.
type PricingService interface {
	LoadPricing() error
	GetPricing(model string) (ModelPricing, error)
	UpdatePricing() error
}

type pricingService struct {
	pricingData map[string]ModelPricing
	mu          sync.RWMutex
	pricingFile string
	logger      *zap.Logger
}

// ModelPricing represents pricing information for a specific model.
type ModelPricing struct {
	Model       string  `json:"model"`
	Input       float64 `json:"input"`                  // Price per million input tokens (USD)
	Output      float64 `json:"output"`                 // Price per million output tokens (USD)
	CacheCreate float64 `json:"cache_create,omitempty"` // Price per million cache creation tokens (USD)
	CacheRead   float64 `json:"cache_read,omitempty"`   // Price per million cache read tokens (USD)
}

// NewPricingService creates a new pricing service.
func NewPricingService(pricingFile string, logger *zap.Logger) PricingService {
	return &pricingService{
		pricingData: make(map[string]ModelPricing),
		pricingFile: pricingFile,
		logger:      logger,
	}
}

// RawModelPricing represents the raw pricing structure from the JSON file.
// The JSON uses per-token costs, which we convert to per-million-token costs.
type RawModelPricing struct {
	InputCostPerToken           float64 `json:"input_cost_per_token"`
	OutputCostPerToken          float64 `json:"output_cost_per_token"`
	CacheReadInputTokenCost     float64 `json:"cache_read_input_token_cost"`
	CacheCreationInputTokenCost float64 `json:"cache_creation_input_token_cost"`
}

// LoadPricing loads pricing data from the JSON file.
// Expects object format from model_prices_and_context_window.json.
func (s *pricingService) LoadPricing() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read pricing file
	data, err := os.ReadFile(s.pricingFile)
	if err != nil {
		return fmt.Errorf("failed to read pricing file: %w", err)
	}

	// Unmarshal as object (model_prices_and_context_window.json format)
	var rawPricingMap map[string]RawModelPricing
	if err := json.Unmarshal(data, &rawPricingMap); err != nil {
		return fmt.Errorf("failed to unmarshal pricing data: %w", err)
	}

	// Convert raw pricing to ModelPricing
	newPricingData := make(map[string]ModelPricing)
	for model, rawPricing := range rawPricingMap {
		// Convert per-token costs to per-million-token costs
		pricing := ModelPricing{
			Model:       model,
			Input:       rawPricing.InputCostPerToken * 1_000_000,           // Convert to per million tokens
			Output:      rawPricing.OutputCostPerToken * 1_000_000,          // Convert to per million tokens
			CacheRead:   rawPricing.CacheReadInputTokenCost * 1_000_000,     // Convert to per million tokens
			CacheCreate: rawPricing.CacheCreationInputTokenCost * 1_000_000, // Convert to per million tokens
		}
		newPricingData[model] = pricing
	}

	s.pricingData = newPricingData

	s.logger.Info("Pricing data loaded successfully",
		zap.Int("models_count", len(s.pricingData)),
		zap.String("pricing_file", s.pricingFile),
	)

	return nil
}

// GetPricing retrieves pricing information for a specific model.
func (s *pricingService) GetPricing(model string) (ModelPricing, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pricing, exists := s.pricingData[model]
	if !exists {
		return ModelPricing{}, fmt.Errorf("pricing not found for model: %s", model)
	}

	return pricing, nil
}

// UpdatePricing reloads pricing data from the file.
func (s *pricingService) UpdatePricing() error {
	return s.LoadPricing()
}
