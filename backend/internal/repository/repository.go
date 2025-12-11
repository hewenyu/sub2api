package repository

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

// AdminRepository defines the interface for admin operations.
type AdminRepository interface {
	Create(ctx context.Context, admin *model.Admin) error
	GetByID(ctx context.Context, id int64) (*model.Admin, error)
	GetByUsername(ctx context.Context, username string) (*model.Admin, error)
	List(ctx context.Context, offset, limit int) ([]*model.Admin, error)
	Update(ctx context.Context, id int64, updates map[string]any) error
	Delete(ctx context.Context, id int64) error
}

// APIKeyRepository defines the interface for API key operations.
type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *model.APIKey) error
	GetByID(ctx context.Context, id int64) (*model.APIKey, error)
	GetByHash(ctx context.Context, keyHash string) (*model.APIKey, error)
	List(ctx context.Context, offset, limit int) ([]*model.APIKey, error)
	Update(ctx context.Context, id int64, updates map[string]any) error
	UpdateStats(ctx context.Context, id int64, requests, tokens int64, cost float64) error
	Delete(ctx context.Context, id int64) error
}

// ClaudeAccountRepository defines the interface for Claude account operations.
type ClaudeAccountRepository interface {
	Create(ctx context.Context, account *model.ClaudeAccount) error
	GetByID(ctx context.Context, id int64) (*model.ClaudeAccount, error)
	GetByEmail(ctx context.Context, email string) (*model.ClaudeAccount, error)
	List(ctx context.Context, offset, limit int) ([]*model.ClaudeAccount, error)
	GetSchedulable(ctx context.Context, model string) ([]*model.ClaudeAccount, error)
	Update(ctx context.Context, account *model.ClaudeAccount) error
	UpdateConcurrentRequests(ctx context.Context, id int64, delta int) error
	Delete(ctx context.Context, id int64) error
}

// CodexAccountRepository defines the interface for Codex account operations.
type CodexAccountRepository interface {
	Create(ctx context.Context, account *model.CodexAccount) error
	GetByID(ctx context.Context, id int64) (*model.CodexAccount, error)
	GetByEmail(ctx context.Context, email string) (*model.CodexAccount, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*model.CodexAccount, error)
	List(ctx context.Context, filters CodexAccountFilters, offset, limit int) ([]*model.CodexAccount, int64, error)
	GetSchedulable(ctx context.Context) ([]*model.CodexAccount, error)
	Update(ctx context.Context, account *model.CodexAccount) error
	UpdateFields(ctx context.Context, id int64, updates map[string]any) error
	UpdateConcurrentRequests(ctx context.Context, id int64, delta int) error
	Delete(ctx context.Context, id int64) error
}

// CodexAccountFilters defines filters for listing Codex accounts.
type CodexAccountFilters struct {
	AccountType *string
	IsActive    *bool
	Schedulable *bool
	Email       *string
}

// UsageFilters defines filters for usage queries.
type UsageFilters struct {
	APIKeyID  *int64
	UsageType *model.UsageType
	StartDate *time.Time
	EndDate   *time.Time
}

// UsageRepository defines the interface for usage operations.
type UsageRepository interface {
	Create(ctx context.Context, usage *model.Usage) error
	GetByID(ctx context.Context, id int64) (*model.Usage, error)
	List(ctx context.Context, filters UsageFilters, offset, limit int) ([]*model.Usage, error)
	Aggregate(ctx context.Context, filters UsageFilters) (*model.UsageAggregate, error)
	Delete(ctx context.Context, id int64) error
}
