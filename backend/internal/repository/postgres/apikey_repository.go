package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
)

type apiKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates a new API key repository.
func NewAPIKeyRepository(db *gorm.DB) repository.APIKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) Create(ctx context.Context, apiKey *model.APIKey) error {
	if err := r.db.WithContext(ctx).Create(apiKey).Error; err != nil {
		return fmt.Errorf("failed to create api key: %w", err)
	}
	return nil
}

func (r *apiKeyRepository) GetByID(ctx context.Context, id int64) (*model.APIKey, error) {
	var apiKey model.APIKey
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&apiKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("api key not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get api key by id: %w", err)
	}
	return &apiKey, nil
}

func (r *apiKeyRepository) GetByHash(ctx context.Context, keyHash string) (*model.APIKey, error) {
	var apiKey model.APIKey
	if err := r.db.WithContext(ctx).Where("key_hash = ?", keyHash).First(&apiKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("api key not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get api key by hash: %w", err)
	}
	return &apiKey, nil
}

func (r *apiKeyRepository) List(ctx context.Context, offset, limit int) ([]*model.APIKey, error) {
	var apiKeys []*model.APIKey
	query := r.db.WithContext(ctx).Order("id DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&apiKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to list api keys: %w", err)
	}
	return apiKeys, nil
}

func (r *apiKeyRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.APIKey{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update api key: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("api key not found: id=%d", id)
	}

	return nil
}

// UpdateStats atomically increments the statistics for an API key.
func (r *apiKeyRepository) UpdateStats(ctx context.Context, id int64, requests, tokens int64, cost float64) error {
	result := r.db.WithContext(ctx).Model(&model.APIKey{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"total_requests": gorm.Expr("total_requests + ?", requests),
			"total_tokens":   gorm.Expr("total_tokens + ?", tokens),
			"total_cost":     gorm.Expr("total_cost + ?", cost),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update api key stats: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("api key not found: id=%d", id)
	}

	return nil
}

func (r *apiKeyRepository) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&model.APIKey{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete api key: %w", err)
	}
	return nil
}
