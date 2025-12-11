package limit

import (
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

// ModelRestrictor manages model restrictions for API keys.
type ModelRestrictor interface {
	IsModelAllowed(apiKey *model.APIKey, modelName string) bool
}

type modelRestrictor struct {
	logger *zap.Logger
}

// NewModelRestrictor creates a new model restrictor.
func NewModelRestrictor(logger *zap.Logger) ModelRestrictor {
	return &modelRestrictor{logger: logger}
}

// IsModelAllowed checks if a model is allowed for the API key.
func (r *modelRestrictor) IsModelAllowed(apiKey *model.APIKey, modelName string) bool {
	if !apiKey.EnableModelRestriction {
		return true
	}

	// Check if model is in restricted list
	for _, restricted := range apiKey.RestrictedModels {
		if restricted == modelName {
			r.logger.Warn("Model restricted",
				zap.Int64("api_key_id", apiKey.ID),
				zap.String("model", modelName),
			)
			return false
		}
	}

	return true
}
