package billing

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupPricingService(t *testing.T) PricingService {
	logger := zap.NewNop()
	// Use the real pricing config used by the backend so that
	// cost calculations stay in sync with production configuration.
	//
	// Resolve the project root based on this test file location:
	// backend/internal/service/billing -> backend -> config/...
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	testDir := filepath.Dir(filename)
	projectRoot := filepath.Join(testDir, "..", "..", "..")
	pricingFile := filepath.Join(projectRoot, "config", "model_prices_and_context_window.json")

	service := NewPricingService(pricingFile, logger)
	err := service.LoadPricing()
	require.NoError(t, err)

	return service
}

func TestCostCalculator_Calculate(t *testing.T) {
	logger := zap.NewNop()
	pricingSvc := setupPricingService(t)
	calculator := NewCostCalculator(pricingSvc, logger)

	t.Run("calculate gpt-4 cost", func(t *testing.T) {
		usage := &UsageData{
			InputTokens:  1000,
			OutputTokens: 500,
			TotalTokens:  1500,
		}

		costInfo, err := calculator.Calculate(usage, "gpt-4")
		require.NoError(t, err)

		// Expected: (1000/1000000)*30 + (500/1000000)*60 = 0.03 + 0.03 = 0.06
		expectedInputCost := 0.03
		expectedOutputCost := 0.03
		expectedTotalCost := 0.06

		assert.InDelta(t, expectedInputCost, costInfo.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, costInfo.OutputCost, 0.0001)
		assert.InDelta(t, expectedTotalCost, costInfo.TotalCost, 0.0001)
	})

	t.Run("calculate gpt-3.5-turbo cost", func(t *testing.T) {
		usage := &UsageData{
			InputTokens:  10000,
			OutputTokens: 5000,
			TotalTokens:  15000,
		}

		costInfo, err := calculator.Calculate(usage, "gpt-3.5-turbo")
		require.NoError(t, err)

		// Expected: (10000/1000000)*0.5 + (5000/1000000)*1.5 = 0.005 + 0.0075 = 0.0125
		expectedInputCost := 0.005
		expectedOutputCost := 0.0075
		expectedTotalCost := 0.0125

		assert.InDelta(t, expectedInputCost, costInfo.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, costInfo.OutputCost, 0.0001)
		assert.InDelta(t, expectedTotalCost, costInfo.TotalCost, 0.0001)
	})

	t.Run("calculate with zero tokens", func(t *testing.T) {
		usage := &UsageData{
			InputTokens:  0,
			OutputTokens: 0,
			TotalTokens:  0,
		}

		costInfo, err := calculator.Calculate(usage, "gpt-4")
		require.NoError(t, err)

		assert.Equal(t, 0.0, costInfo.InputCost)
		assert.Equal(t, 0.0, costInfo.OutputCost)
		assert.Equal(t, 0.0, costInfo.TotalCost)
	})

	t.Run("calculate with large token count", func(t *testing.T) {
		usage := &UsageData{
			InputTokens:  1_000_000, // 1 million
			OutputTokens: 2_000_000, // 2 million
			TotalTokens:  3_000_000,
		}

		costInfo, err := calculator.Calculate(usage, "gpt-4o")
		require.NoError(t, err)

		// Pricing from model_prices_and_context_window.json:
		// gpt-4o input_cost_per_token = 2.5e-06 -> 2.5 per million
		// gpt-4o output_cost_per_token = 1e-05 -> 10 per million
		// Expected: (1_000_000/1_000_000)*2.5 + (2_000_000/1_000_000)*10 = 2.5 + 20 = 22.5
		expectedInputCost := 2.5
		expectedOutputCost := 20.0
		expectedTotalCost := 22.5

		assert.InDelta(t, expectedInputCost, costInfo.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, costInfo.OutputCost, 0.0001)
		assert.InDelta(t, expectedTotalCost, costInfo.TotalCost, 0.0001)
	})

	t.Run("calculate gpt-4.1 cost", func(t *testing.T) {
		usage := &UsageData{
			InputTokens:  2000,
			OutputTokens: 1000,
			TotalTokens:  3000,
		}

		costInfo, err := calculator.Calculate(usage, "gpt-4.1")
		require.NoError(t, err)

		// From model_prices_and_context_window.json:
		// gpt-4.1 input_cost_per_token = 2e-06 -> 2 per million
		// gpt-4.1 output_cost_per_token = 8e-06 -> 8 per million
		// Expected: (2000/1_000_000)*2 + (1000/1_000_000)*8 = 0.004 + 0.008 = 0.012
		expectedInputCost := 0.004
		expectedOutputCost := 0.008
		expectedTotalCost := 0.012

		assert.InDelta(t, expectedInputCost, costInfo.InputCost, 0.000001)
		assert.InDelta(t, expectedOutputCost, costInfo.OutputCost, 0.000001)
		assert.InDelta(t, expectedTotalCost, costInfo.TotalCost, 0.000001)
	})

	t.Run("nil usage data", func(t *testing.T) {
		_, err := calculator.Calculate(nil, "gpt-4")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage data cannot be nil")
	})

	t.Run("unknown model", func(t *testing.T) {
		usage := &UsageData{
			InputTokens:  1000,
			OutputTokens: 500,
			TotalTokens:  1500,
		}

		_, err := calculator.Calculate(usage, "unknown-model")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get pricing")
	})
}

func TestCostCalculator_CalculateTotalCost(t *testing.T) {
	logger := zap.NewNop()
	pricingSvc := setupPricingService(t)
	calculator := NewCostCalculator(pricingSvc, logger)

	t.Run("calculate total cost for gpt-4", func(t *testing.T) {
		totalCost, err := calculator.CalculateTotalCost(2000, 1000, "gpt-4")
		require.NoError(t, err)

		// Expected: (2000/1000000)*30 + (1000/1000000)*60 = 0.06 + 0.06 = 0.12
		expectedTotalCost := 0.12

		assert.InDelta(t, expectedTotalCost, totalCost, 0.0001)
	})

	t.Run("calculate total cost for gpt-3.5-turbo", func(t *testing.T) {
		totalCost, err := calculator.CalculateTotalCost(50000, 25000, "gpt-3.5-turbo")
		require.NoError(t, err)

		// Expected: (50000/1000000)*0.5 + (25000/1000000)*1.5 = 0.025 + 0.0375 = 0.0625
		expectedTotalCost := 0.0625

		assert.InDelta(t, expectedTotalCost, totalCost, 0.0001)
	})

	t.Run("unknown model", func(t *testing.T) {
		_, err := calculator.CalculateTotalCost(1000, 500, "unknown-model")
		assert.Error(t, err)
	})
}

func TestCostCalculator_PrecisionTest(t *testing.T) {
	logger := zap.NewNop()
	pricingSvc := setupPricingService(t)
	calculator := NewCostCalculator(pricingSvc, logger)

	t.Run("small token counts precision", func(t *testing.T) {
		usage := &UsageData{
			InputTokens:  1,
			OutputTokens: 1,
			TotalTokens:  2,
		}

		costInfo, err := calculator.Calculate(usage, "gpt-4")
		require.NoError(t, err)

		// Expected: (1/1000000)*30 + (1/1000000)*60 = 0.00003 + 0.00006 = 0.00009
		expectedTotalCost := 0.00009

		assert.InDelta(t, expectedTotalCost, costInfo.TotalCost, 0.000001)
	})

	t.Run("fractional calculations", func(t *testing.T) {
		usage := &UsageData{
			InputTokens:  123,
			OutputTokens: 456,
			TotalTokens:  579,
		}

		costInfo, err := calculator.Calculate(usage, "gpt-4o")
		require.NoError(t, err)

		// Pricing from model_prices_and_context_window.json:
		// gpt-4o input_cost_per_token = 2.5e-06 -> 2.5 per million
		// gpt-4o output_cost_per_token = 1e-05 -> 10 per million
		// Expected: (123/1_000_000)*2.5 + (456/1_000_000)*10
		expectedInputCost := 0.0003075
		expectedOutputCost := 0.00456

		assert.InDelta(t, expectedInputCost, costInfo.InputCost, 0.000001)
		assert.InDelta(t, expectedOutputCost, costInfo.OutputCost, 0.000001)
	})

	t.Run("gpt-4o-mini fractional calculations", func(t *testing.T) {
		usage := &UsageData{
			InputTokens:  321,
			OutputTokens: 654,
			TotalTokens:  975,
		}

		costInfo, err := calculator.Calculate(usage, "gpt-4o-mini")
		require.NoError(t, err)

		// From model_prices_and_context_window.json:
		// gpt-4o-mini input_cost_per_token = 1.5e-07 -> 0.15 per million
		// gpt-4o-mini output_cost_per_token = 6e-07 -> 0.6 per million
		//
		// Expected:
		//  input_cost  = (321 / 1_000_000) * 0.15  = 4.815e-05
		//  output_cost = (654 / 1_000_000) * 0.6   = 0.0003924
		expectedInputCost := 0.00004815
		expectedOutputCost := 0.0003924

		assert.InDelta(t, expectedInputCost, costInfo.InputCost, 1e-9)
		assert.InDelta(t, expectedOutputCost, costInfo.OutputCost, 1e-9)
	})
}
