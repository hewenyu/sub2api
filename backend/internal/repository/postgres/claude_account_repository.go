package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
)

type claudeAccountRepository struct {
	db *gorm.DB
}

// NewClaudeAccountRepository creates a new Claude account repository.
func NewClaudeAccountRepository(db *gorm.DB) repository.ClaudeAccountRepository {
	return &claudeAccountRepository{db: db}
}

func (r *claudeAccountRepository) Create(ctx context.Context, account *model.ClaudeAccount) error {
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		return fmt.Errorf("failed to create claude account: %w", err)
	}
	return nil
}

func (r *claudeAccountRepository) GetByID(ctx context.Context, id int64) (*model.ClaudeAccount, error) {
	var account model.ClaudeAccount
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("claude account not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get claude account by id: %w", err)
	}
	return &account, nil
}

func (r *claudeAccountRepository) GetByEmail(ctx context.Context, email string) (*model.ClaudeAccount, error) {
	var account model.ClaudeAccount
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("claude account not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get claude account by email: %w", err)
	}
	return &account, nil
}

func (r *claudeAccountRepository) List(ctx context.Context, offset, limit int) ([]*model.ClaudeAccount, error) {
	var accounts []*model.ClaudeAccount
	query := r.db.WithContext(ctx).Order("id DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("failed to list claude accounts: %w", err)
	}
	return accounts, nil
}

// GetSchedulable returns Claude accounts available for scheduling based on filters.
// Filters:
// - is_active = true
// - is_schedulable = true
// - expires_at > now
// - rate_limited_until is NULL OR < now
// - overload_until is NULL OR < now
// - For Opus model: features JSONB contains "claude_max": true
func (r *claudeAccountRepository) GetSchedulable(ctx context.Context, modelName string) ([]*model.ClaudeAccount, error) {
	now := time.Now().UTC()

	query := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Where("is_schedulable = ?", true).
		Where("expires_at > ?", now).
		Where("(rate_limited_until IS NULL OR rate_limited_until < ?)", now).
		Where("(overload_until IS NULL OR overload_until < ?)", now)

	// For Opus model, check if features contains "claude_max": true
	if modelName == "claude-opus-4-20250514" {
		query = query.Where("features->>'claude_max' = ?", "true")
	}

	var accounts []*model.ClaudeAccount
	if err := query.Order("concurrent_requests ASC, total_requests ASC").Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get schedulable claude accounts: %w", err)
	}

	return accounts, nil
}

func (r *claudeAccountRepository) Update(ctx context.Context, account *model.ClaudeAccount) error {
	if err := r.db.WithContext(ctx).Save(account).Error; err != nil {
		return fmt.Errorf("failed to update claude account: %w", err)
	}
	return nil
}

// UpdateConcurrentRequests atomically updates the concurrent_requests count.
func (r *claudeAccountRepository) UpdateConcurrentRequests(ctx context.Context, id int64, delta int) error {
	result := r.db.WithContext(ctx).Model(&model.ClaudeAccount{}).
		Where("id = ?", id).
		Update("concurrent_requests", gorm.Expr("concurrent_requests + ?", delta))

	if result.Error != nil {
		return fmt.Errorf("failed to update concurrent requests: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("claude account not found: id=%d", id)
	}

	return nil
}

func (r *claudeAccountRepository) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&model.ClaudeAccount{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete claude account: %w", err)
	}
	return nil
}
