package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
)

type codexAccountRepository struct {
	db *gorm.DB
}

// NewCodexAccountRepository creates a new Codex account repository.
func NewCodexAccountRepository(db *gorm.DB) repository.CodexAccountRepository {
	return &codexAccountRepository{db: db}
}

func (r *codexAccountRepository) GetByEmail(ctx context.Context, email string) (*model.CodexAccount, error) {
	var account model.CodexAccount
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("codex account not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get codex account by email: %w", err)
	}
	return &account, nil
}

func (r *codexAccountRepository) Create(ctx context.Context, account *model.CodexAccount) error {
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		return fmt.Errorf("failed to create codex account: %w", err)
	}
	return nil
}

func (r *codexAccountRepository) GetByID(ctx context.Context, id int64) (*model.CodexAccount, error) {
	var account model.CodexAccount
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("codex account not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get codex account by id: %w", err)
	}
	return &account, nil
}

func (r *codexAccountRepository) GetByAPIKey(ctx context.Context, apiKey string) (*model.CodexAccount, error) {
	var account model.CodexAccount
	if err := r.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("codex account not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get codex account by api key: %w", err)
	}
	return &account, nil
}

func (r *codexAccountRepository) List(ctx context.Context, filters repository.CodexAccountFilters, offset, limit int) ([]*model.CodexAccount, int64, error) {
	var accounts []*model.CodexAccount
	var total int64

	query := r.db.WithContext(ctx).Model(&model.CodexAccount{})

	// Apply filters
	if filters.AccountType != nil {
		query = query.Where("account_type = ?", *filters.AccountType)
	}
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}
	if filters.Schedulable != nil {
		query = query.Where("schedulable = ?", *filters.Schedulable)
	}
	if filters.Email != nil {
		query = query.Where("email = ?", *filters.Email)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count codex accounts: %w", err)
	}

	// Apply pagination and ordering
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Order("created_at DESC").Find(&accounts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list codex accounts: %w", err)
	}

	return accounts, total, nil
}

// GetSchedulable returns Codex accounts available for scheduling based on filters.
// Filters:
// - is_active = true
// - schedulable = true
// - rate_limited_until is NULL OR < now
// - overload_until is NULL OR < now
func (r *codexAccountRepository) GetSchedulable(ctx context.Context) ([]*model.CodexAccount, error) {
	now := time.Now().UTC()

	var accounts []*model.CodexAccount
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Where("schedulable = ?", true).
		Where("(rate_limited_until IS NULL OR rate_limited_until < ?)", now).
		Where("(overload_until IS NULL OR overload_until < ?)", now).
		Order("concurrent_requests ASC, total_requests ASC").
		Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get schedulable codex accounts: %w", err)
	}

	return accounts, nil
}

func (r *codexAccountRepository) Update(ctx context.Context, account *model.CodexAccount) error {
	if err := r.db.WithContext(ctx).Save(account).Error; err != nil {
		return fmt.Errorf("failed to update codex account: %w", err)
	}
	return nil
}

func (r *codexAccountRepository) UpdateFields(ctx context.Context, id int64, updates map[string]any) error {
	result := r.db.WithContext(ctx).Model(&model.CodexAccount{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update codex account fields: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("codex account not found: id=%d", id)
	}

	return nil
}

// UpdateConcurrentRequests atomically updates the concurrent_requests count.
func (r *codexAccountRepository) UpdateConcurrentRequests(ctx context.Context, id int64, delta int) error {
	result := r.db.WithContext(ctx).Model(&model.CodexAccount{}).
		Where("id = ?", id).
		Update("concurrent_requests", gorm.Expr("concurrent_requests + ?", delta))

	if result.Error != nil {
		return fmt.Errorf("failed to update concurrent requests: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("codex account not found: id=%d", id)
	}

	return nil
}

func (r *codexAccountRepository) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&model.CodexAccount{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete codex account: %w", err)
	}
	return nil
}
