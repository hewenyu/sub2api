package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
)

type proxyConfigRepository struct {
	db *gorm.DB
}

// NewProxyConfigRepository creates a new proxy config repository.
func NewProxyConfigRepository(db *gorm.DB) repository.ProxyConfigRepository {
	return &proxyConfigRepository{db: db}
}

func (r *proxyConfigRepository) Create(ctx context.Context, proxyConfig *model.ProxyConfig) error {
	if err := r.db.WithContext(ctx).Create(proxyConfig).Error; err != nil {
		return fmt.Errorf("failed to create proxy config: %w", err)
	}
	return nil
}

func (r *proxyConfigRepository) GetByID(ctx context.Context, id int64) (*model.ProxyConfig, error) {
	var proxyConfig model.ProxyConfig
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&proxyConfig).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("proxy config not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get proxy config by id: %w", err)
	}
	return &proxyConfig, nil
}

func (r *proxyConfigRepository) GetByName(ctx context.Context, name string) (*model.ProxyConfig, error) {
	var proxyConfig model.ProxyConfig
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&proxyConfig).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("proxy config not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get proxy config by name: %w", err)
	}
	return &proxyConfig, nil
}

func (r *proxyConfigRepository) GetDefault(ctx context.Context) (*model.ProxyConfig, error) {
	var proxyConfig model.ProxyConfig
	if err := r.db.WithContext(ctx).
		Where("is_default = ? AND enabled = ?", true, true).
		First(&proxyConfig).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No default proxy configured – return nil without error as per contract.
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get default proxy config: %w", err)
	}
	return &proxyConfig, nil
}

func (r *proxyConfigRepository) List(ctx context.Context, filters repository.ProxyConfigFilters, page, pageSize int) ([]*model.ProxyConfig, int64, error) {
	var proxyConfigs []*model.ProxyConfig
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ProxyConfig{})

	// Apply filters
	if filters.Enabled != nil {
		query = query.Where("enabled = ?", *filters.Enabled)
	}
	if filters.IsDefault != nil {
		query = query.Where("is_default = ?", *filters.IsDefault)
	}
	if filters.Protocol != nil {
		query = query.Where("protocol = ?", *filters.Protocol)
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count proxy configs: %w", err)
	}

	// Apply sorting: is_default DESC, created_at DESC
	query = query.Order("is_default DESC, created_at DESC")

	// Apply pagination
	if pageSize > 0 {
		query = query.Limit(pageSize)
	}
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		if offset > 0 {
			query = query.Offset(offset)
		}
	}

	if err := query.Find(&proxyConfigs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list proxy configs: %w", err)
	}

	return proxyConfigs, total, nil
}

func (r *proxyConfigRepository) Update(ctx context.Context, id int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&model.ProxyConfig{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update proxy config: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("proxy config not found: id=%d", id)
	}
	return nil
}

func (r *proxyConfigRepository) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&model.ProxyConfig{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete proxy config: %w", err)
	}
	return nil
}

func (r *proxyConfigRepository) SetDefault(ctx context.Context, id int64) error {
	// Use transaction to ensure atomicity
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First, check if the proxy exists and is enabled
		var proxyConfig model.ProxyConfig
		if err := tx.Where("id = ?", id).First(&proxyConfig).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("proxy config not found: %w", err)
			}
			return fmt.Errorf("failed to get proxy config: %w", err)
		}

		if !proxyConfig.Enabled {
			return fmt.Errorf("cannot set disabled proxy as default")
		}

		// Unset any existing default
		if err := tx.Model(&model.ProxyConfig{}).
			Where("is_default = ?", true).
			Update("is_default", false).Error; err != nil {
			return fmt.Errorf("failed to unset existing default: %w", err)
		}

		// Set new default
		if err := tx.Model(&model.ProxyConfig{}).
			Where("id = ?", id).
			Update("is_default", true).Error; err != nil {
			return fmt.Errorf("failed to set new default: %w", err)
		}

		return nil
	})
}

func (r *proxyConfigRepository) CountByName(ctx context.Context, name string, excludeID int64) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.ProxyConfig{}).Where("name = ?", name)

	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count proxy configs by name: %w", err)
	}

	return count, nil
}
