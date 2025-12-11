package repository

import (
	"context"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

// ProxyConfigFilters defines filters for listing proxy configurations.
type ProxyConfigFilters struct {
	Enabled   *bool
	IsDefault *bool
	Protocol  *string
}

// ProxyConfigRepository defines the interface for proxy configuration operations.
type ProxyConfigRepository interface {
	// Create creates a new proxy configuration.
	Create(ctx context.Context, proxyConfig *model.ProxyConfig) error

	// GetByID retrieves a proxy configuration by ID.
	GetByID(ctx context.Context, id int64) (*model.ProxyConfig, error)

	// GetByName retrieves a proxy configuration by name.
	GetByName(ctx context.Context, name string) (*model.ProxyConfig, error)

	// GetDefault retrieves the default proxy configuration (must be enabled).
	GetDefault(ctx context.Context) (*model.ProxyConfig, error)

	// List retrieves proxy configurations with filters and pagination.
	// Results are sorted by: is_default DESC, created_at DESC
	List(ctx context.Context, filters ProxyConfigFilters, page, pageSize int) ([]*model.ProxyConfig, int64, error)

	// Update updates a proxy configuration.
	Update(ctx context.Context, id int64, updates map[string]any) error

	// Delete soft-deletes a proxy configuration.
	Delete(ctx context.Context, id int64) error

	// SetDefault sets a proxy as the default (transactional operation).
	// This will unset any existing default proxy and set the new one.
	// The proxy must be enabled to be set as default.
	SetDefault(ctx context.Context, id int64) error

	// CountByName counts proxy configurations with the given name (excluding deleted).
	// Used for checking uniqueness, can optionally exclude a specific ID.
	CountByName(ctx context.Context, name string, excludeID int64) (int64, error)
}
