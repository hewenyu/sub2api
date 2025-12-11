package limit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

func TestModelRestrictor_IsModelAllowed(t *testing.T) {
	logger := zap.NewNop()
	restrictor := NewModelRestrictor(logger)

	t.Run("allowed when restriction disabled", func(t *testing.T) {
		apiKey := &model.APIKey{
			ID:                     1,
			EnableModelRestriction: false,
			RestrictedModels:       model.StringArray{"claude-3-opus-20240229"},
		}

		allowed := restrictor.IsModelAllowed(apiKey, "claude-3-opus-20240229")
		assert.True(t, allowed)
	})

	t.Run("allowed when model not in restricted list", func(t *testing.T) {
		apiKey := &model.APIKey{
			ID:                     1,
			EnableModelRestriction: true,
			RestrictedModels:       model.StringArray{"claude-3-opus-20240229", "claude-3-sonnet-20240229"},
		}

		allowed := restrictor.IsModelAllowed(apiKey, "claude-3-haiku-20240307")
		assert.True(t, allowed)
	})

	t.Run("not allowed when model in restricted list", func(t *testing.T) {
		apiKey := &model.APIKey{
			ID:                     1,
			EnableModelRestriction: true,
			RestrictedModels:       model.StringArray{"claude-3-opus-20240229", "claude-3-sonnet-20240229"},
		}

		allowed := restrictor.IsModelAllowed(apiKey, "claude-3-opus-20240229")
		assert.False(t, allowed)
	})

	t.Run("allowed when restricted list is empty", func(t *testing.T) {
		apiKey := &model.APIKey{
			ID:                     1,
			EnableModelRestriction: true,
			RestrictedModels:       model.StringArray{},
		}

		allowed := restrictor.IsModelAllowed(apiKey, "claude-3-opus-20240229")
		assert.True(t, allowed)
	})

	t.Run("exact match required", func(t *testing.T) {
		apiKey := &model.APIKey{
			ID:                     1,
			EnableModelRestriction: true,
			RestrictedModels:       model.StringArray{"claude-3-opus"},
		}

		// Should not match partial strings
		allowed := restrictor.IsModelAllowed(apiKey, "claude-3-opus-20240229")
		assert.True(t, allowed)

		// Should match exact string
		allowed = restrictor.IsModelAllowed(apiKey, "claude-3-opus")
		assert.False(t, allowed)
	})
}
