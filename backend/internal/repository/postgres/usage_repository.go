package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
)

type usageRepository struct {
	db *gorm.DB
}

// NewUsageRepository creates a new usage repository.
func NewUsageRepository(db *gorm.DB) repository.UsageRepository {
	return &usageRepository{db: db}
}

func (r *usageRepository) Create(ctx context.Context, usage *model.Usage) error {
	if err := r.db.WithContext(ctx).Create(usage).Error; err != nil {
		return fmt.Errorf("failed to create usage record: %w", err)
	}
	return nil
}

func (r *usageRepository) GetByID(ctx context.Context, id int64) (*model.Usage, error) {
	var usage model.Usage
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("usage record not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get usage record by id: %w", err)
	}
	return &usage, nil
}

// List returns usage records with filters.
func (r *usageRepository) List(ctx context.Context, filters repository.UsageFilters, offset, limit int) ([]*model.Usage, error) {
	var usages []*model.Usage
	query := r.db.WithContext(ctx)

	query = r.applyFilters(query, filters)
	query = query.Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&usages).Error; err != nil {
		return nil, fmt.Errorf("failed to list usage records: %w", err)
	}
	return usages, nil
}

// Aggregate calculates aggregate statistics for usage records.
func (r *usageRepository) Aggregate(ctx context.Context, filters repository.UsageFilters) (*model.UsageAggregate, error) {
	var aggregate model.UsageAggregate

	query := r.db.WithContext(ctx).Model(&model.Usage{})
	query = r.applyFilters(query, filters)

	// Calculate aggregate statistics
	if err := query.
		Select("COUNT(*) as total_requests, COALESCE(SUM(total_tokens), 0) as total_tokens, COALESCE(SUM(cost), 0) as total_cost").
		Scan(&aggregate).Error; err != nil {
		return nil, fmt.Errorf("failed to aggregate usage records: %w", err)
	}

	return &aggregate, nil
}

// applyFilters applies filters to a query.
func (r *usageRepository) applyFilters(query *gorm.DB, filters repository.UsageFilters) *gorm.DB {
	if filters.APIKeyID != nil {
		query = query.Where("api_key_id = ?", *filters.APIKeyID)
	}

	if filters.UsageType != nil {
		query = query.Where("type = ?", *filters.UsageType)
	}

	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}

	if filters.EndDate != nil {
		query = query.Where("created_at < ?", *filters.EndDate)
	}

	return query
}

func (r *usageRepository) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&model.Usage{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete usage record: %w", err)
	}
	return nil
}
