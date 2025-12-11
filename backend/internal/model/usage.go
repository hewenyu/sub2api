package model

import (
	"time"

	"gorm.io/gorm"
)

// UsageType represents the type of usage record.
type UsageType string

const (
	UsageTypeClaude UsageType = "claude"
	UsageTypeCodex  UsageType = "codex"
)

// Usage represents a usage record for API requests.
type Usage struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	APIKeyID     int64     `gorm:"type:bigint;not null;index:idx_usage_composite" json:"api_key_id"`
	Type         UsageType `gorm:"type:varchar(50);not null;index:idx_usage_composite" json:"type"`
	AccountID    int64     `gorm:"type:bigint;not null" json:"account_id"`
	Model        string    `gorm:"type:varchar(100);not null" json:"model"`
	InputTokens  int64     `gorm:"type:bigint;default:0;not null" json:"input_tokens"`
	OutputTokens int64     `gorm:"type:bigint;default:0;not null" json:"output_tokens"`
	TotalTokens  int64     `gorm:"type:bigint;default:0;not null" json:"total_tokens"`
	// CacheCreationInputTokens and CacheReadInputTokens store prompt caching
	// usage details for more accurate billing and analytics. They mirror the
	// design described in the PRD's usage_records table.
	CacheCreationInputTokens int64          `gorm:"type:bigint;default:0;not null" json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64          `gorm:"type:bigint;default:0;not null" json:"cache_read_input_tokens"`
	Cost                     float64        `gorm:"type:decimal(20,6);default:0;not null" json:"cost"`
	RequestDuration          int64          `gorm:"type:bigint;default:0;not null" json:"request_duration"`
	StatusCode               int            `gorm:"type:int;not null" json:"status_code"`
	ErrorMessage             string         `gorm:"type:text" json:"error_message"`
	RequestMetadata          string         `gorm:"type:jsonb" json:"request_metadata"`
	ResponseMetadata         string         `gorm:"type:jsonb" json:"response_metadata"`
	CreatedAt                time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;index:idx_usage_composite" json:"created_at"`
	UpdatedAt                time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"type:timestamp;index" json:"-"`
}

// TableName specifies the table name for the Usage model.
func (Usage) TableName() string {
	return "usage"
}

// UsageAggregate represents aggregated usage statistics.
type UsageAggregate struct {
	TotalRequests int64   `json:"total_requests"`
	TotalTokens   int64   `json:"total_tokens"`
	TotalCost     float64 `json:"total_cost"`
}
