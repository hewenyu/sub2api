package billing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPricingService_LoadPricing(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful load", func(t *testing.T) {
		// Create temp pricing file
		tempDir := t.TempDir()
		pricingFile := filepath.Join(tempDir, "pricing.json")

		// Object format: per-token costs are converted to per-million-token costs
		// 30.0 per million = 0.00003 per token
		// 60.0 per million = 0.00006 per token
		content := `{
			"gpt-4": {
				"input_cost_per_token": 0.00003,
				"output_cost_per_token": 0.00006,
				"cache_read_input_token_cost": 0.0000015,
				"cache_creation_input_token_cost": 0.0000075
			},
			"gpt-3.5-turbo": {
				"input_cost_per_token": 0.0000005,
				"output_cost_per_token": 0.0000015,
				"cache_read_input_token_cost": 0.0,
				"cache_creation_input_token_cost": 0.0
			}
		}`
		err := os.WriteFile(pricingFile, []byte(content), 0644)
		require.NoError(t, err)

		service := NewPricingService(pricingFile, logger)
		err = service.LoadPricing()
		require.NoError(t, err)

		// Verify pricing loaded correctly
		pricing, err := service.GetPricing("gpt-4")
		require.NoError(t, err)
		assert.Equal(t, "gpt-4", pricing.Model)
		assert.Equal(t, 30.0, pricing.Input)
		assert.Equal(t, 60.0, pricing.Output)
		assert.Equal(t, 1.5, pricing.CacheRead)
		assert.Equal(t, 7.5, pricing.CacheCreate)

		pricing, err = service.GetPricing("gpt-3.5-turbo")
		require.NoError(t, err)
		assert.Equal(t, "gpt-3.5-turbo", pricing.Model)
		assert.Equal(t, 0.5, pricing.Input)
		assert.Equal(t, 1.5, pricing.Output)
		assert.Equal(t, 0.0, pricing.CacheRead)
		assert.Equal(t, 0.0, pricing.CacheCreate)
	})

	t.Run("file not found", func(t *testing.T) {
		service := NewPricingService("/nonexistent/pricing.json", logger)
		err := service.LoadPricing()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read pricing file")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		tempDir := t.TempDir()
		pricingFile := filepath.Join(tempDir, "pricing.json")

		err := os.WriteFile(pricingFile, []byte("invalid json"), 0644)
		require.NoError(t, err)

		service := NewPricingService(pricingFile, logger)
		err = service.LoadPricing()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal pricing data")
	})

	t.Run("empty pricing file", func(t *testing.T) {
		tempDir := t.TempDir()
		pricingFile := filepath.Join(tempDir, "pricing.json")

		err := os.WriteFile(pricingFile, []byte("{}"), 0644)
		require.NoError(t, err)

		service := NewPricingService(pricingFile, logger)
		err = service.LoadPricing()
		require.NoError(t, err)

		_, err = service.GetPricing("gpt-4")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pricing not found")
	})
}

func TestPricingService_GetPricing(t *testing.T) {
	logger := zap.NewNop()

	// Setup
	tempDir := t.TempDir()
	pricingFile := filepath.Join(tempDir, "pricing.json")

	content := `{
		"gpt-4": {
			"input_cost_per_token": 0.00003,
			"output_cost_per_token": 0.00006,
			"cache_read_input_token_cost": 0.0,
			"cache_creation_input_token_cost": 0.0
		},
		"gpt-4o": {
			"input_cost_per_token": 0.000005,
			"output_cost_per_token": 0.000015,
			"cache_read_input_token_cost": 0.0,
			"cache_creation_input_token_cost": 0.0
		}
	}`
	err := os.WriteFile(pricingFile, []byte(content), 0644)
	require.NoError(t, err)

	service := NewPricingService(pricingFile, logger)
	err = service.LoadPricing()
	require.NoError(t, err)

	t.Run("get existing model", func(t *testing.T) {
		pricing, err := service.GetPricing("gpt-4")
		require.NoError(t, err)
		assert.Equal(t, "gpt-4", pricing.Model)
		assert.Equal(t, 30.0, pricing.Input)
		assert.Equal(t, 60.0, pricing.Output)
	})

	t.Run("get non-existent model", func(t *testing.T) {
		_, err := service.GetPricing("non-existent-model")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pricing not found for model: non-existent-model")
	})
}

func TestPricingService_UpdatePricing(t *testing.T) {
	logger := zap.NewNop()

	tempDir := t.TempDir()
	pricingFile := filepath.Join(tempDir, "pricing.json")

	// Initial content
	content := `{
		"gpt-4": {
			"input_cost_per_token": 0.00003,
			"output_cost_per_token": 0.00006,
			"cache_read_input_token_cost": 0.0,
			"cache_creation_input_token_cost": 0.0
		}
	}`
	err := os.WriteFile(pricingFile, []byte(content), 0644)
	require.NoError(t, err)

	service := NewPricingService(pricingFile, logger)
	err = service.LoadPricing()
	require.NoError(t, err)

	pricing, err := service.GetPricing("gpt-4")
	require.NoError(t, err)
	assert.Equal(t, 30.0, pricing.Input)

	// Update file content
	newContent := `{
		"gpt-4": {
			"input_cost_per_token": 0.000035,
			"output_cost_per_token": 0.00007,
			"cache_read_input_token_cost": 0.0,
			"cache_creation_input_token_cost": 0.0
		}
	}`
	err = os.WriteFile(pricingFile, []byte(newContent), 0644)
	require.NoError(t, err)

	// Update pricing
	err = service.UpdatePricing()
	require.NoError(t, err)

	// Verify new pricing
	pricing, err = service.GetPricing("gpt-4")
	require.NoError(t, err)
	assert.Equal(t, 35.0, pricing.Input)
	assert.Equal(t, 70.0, pricing.Output)
}

func TestPricingService_ThreadSafety(t *testing.T) {
	logger := zap.NewNop()

	tempDir := t.TempDir()
	pricingFile := filepath.Join(tempDir, "pricing.json")

	content := `{
		"gpt-4": {
			"input_cost_per_token": 0.00003,
			"output_cost_per_token": 0.00006,
			"cache_read_input_token_cost": 0.0,
			"cache_creation_input_token_cost": 0.0
		}
	}`
	err := os.WriteFile(pricingFile, []byte(content), 0644)
	require.NoError(t, err)

	service := NewPricingService(pricingFile, logger)
	err = service.LoadPricing()
	require.NoError(t, err)

	// Test concurrent reads and writes
	done := make(chan bool)

	// Concurrent readers
	for range 10 {
		go func() {
			for range 100 {
				_, _ = service.GetPricing("gpt-4")
			}
			done <- true
		}()
	}

	// Concurrent updater
	go func() {
		for range 10 {
			_ = service.UpdatePricing()
		}
		done <- true
	}()

	// Wait for all goroutines
	for range 11 {
		<-done
	}
}
