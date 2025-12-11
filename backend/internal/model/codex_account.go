package model

import (
	"time"

	"gorm.io/gorm"
)

// CodexAccount represents a Codex API account.
// Supports two account types:
// - openai-oauth: OAuth-authenticated accounts with access/refresh tokens
// - openai-responses: Manually created accounts with API keys
type CodexAccount struct {
	ID          int64   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string  `gorm:"type:varchar(255);not null" json:"name"`
	AccountType string  `gorm:"type:varchar(50);default:'openai-responses'" json:"account_type"`
	Email       *string `gorm:"type:varchar(255)" json:"email,omitempty"`

	// Encrypted fields (AES-256-CBC, format: {base64(iv)}:{base64(ciphertext)})
	APIKey       *string `gorm:"type:text;uniqueIndex" json:"-"` // For openai-responses type
	AccessToken  *string `gorm:"type:text" json:"-"`             // For openai-oauth type
	RefreshToken *string `gorm:"type:text" json:"-"`             // For openai-oauth type

	// OAuth metadata
	ExpiresAt *time.Time `gorm:"type:timestamp" json:"expires_at,omitempty"`
	Scopes    *string    `gorm:"type:text" json:"scopes,omitempty"`

	// API configuration
	BaseAPI         string  `gorm:"type:varchar(255);default:'https://api.openai.com/v1'" json:"base_api"`
	CustomUserAgent *string `gorm:"type:varchar(255)" json:"custom_user_agent,omitempty"`

	// Subscription information
	SubscriptionLevel     *string    `gorm:"type:varchar(50)" json:"subscription_level,omitempty"`
	SubscriptionExpiresAt *time.Time `gorm:"type:timestamp" json:"subscription_expires_at,omitempty"`

	// OpenAI-specific identifiers (from ID token)
	ChatGPTAccountID  *string `gorm:"column:chatgpt_account_id;type:varchar(255)" json:"chatgpt_account_id,omitempty"`
	ChatGPTUserID     *string `gorm:"column:chatgpt_user_id;type:varchar(255)" json:"chatgpt_user_id,omitempty"`
	OrganizationID    *string `gorm:"column:organization_id;type:varchar(255)" json:"organization_id,omitempty"`
	OrganizationRole  *string `gorm:"column:organization_role;type:varchar(100)" json:"organization_role,omitempty"`
	OrganizationTitle *string `gorm:"column:organization_title;type:varchar(255)" json:"organization_title,omitempty"`

	// Quota management
	DailyQuota     float64    `gorm:"type:decimal(10,2);default:0" json:"daily_quota"`
	DailyUsage     float64    `gorm:"type:decimal(10,2);default:0" json:"daily_usage"`
	LastResetDate  *time.Time `gorm:"type:timestamp" json:"last_reset_date,omitempty"`
	QuotaResetTime string     `gorm:"type:varchar(10);default:'00:00'" json:"quota_reset_time"`

	// Rate limiting status
	RateLimitedUntil *time.Time `gorm:"type:timestamp" json:"rate_limited_until,omitempty"`
	RateLimitStatus  *string    `gorm:"type:varchar(50)" json:"rate_limit_status,omitempty"`
	RateLimitResetAt *time.Time `gorm:"type:timestamp" json:"rate_limit_reset_at,omitempty"`

	// Scheduling configuration
	IsActive           bool `gorm:"type:boolean;default:true;not null" json:"is_active"`
	Schedulable        bool `gorm:"type:boolean;default:true;not null" json:"schedulable"`
	Priority           int  `gorm:"type:int;default:100" json:"priority"`
	ConcurrentRequests int  `gorm:"type:int;default:0;not null" json:"concurrent_requests"`

	// Proxy configuration (references proxy_configs.name)
	ProxyName *string `gorm:"type:varchar(100)" json:"proxy_name,omitempty"`

	// Legacy fields (kept for backward compatibility)
	OverloadUntil *time.Time `gorm:"type:timestamp" json:"overload_until,omitempty"`
	TotalRequests int64      `gorm:"type:bigint;default:0;not null" json:"total_requests"`
	TotalTokens   int64      `gorm:"type:bigint;default:0;not null" json:"total_tokens"`

	// Timestamps
	CreatedAt  time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	LastUsedAt *time.Time     `gorm:"type:timestamp" json:"last_used_at,omitempty"`
	DeletedAt  gorm.DeletedAt `gorm:"type:timestamp;index" json:"-"`
}

// TableName specifies the table name for the CodexAccount model.
func (CodexAccount) TableName() string {
	return "codex_accounts"
}
