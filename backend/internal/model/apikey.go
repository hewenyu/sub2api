package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// StringArray is a custom type for storing string arrays in the database.
type StringArray []string

// Scan implements the sql.Scanner interface.
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			*a = []string{}
			return nil
		}
		bytes = []byte(str)
	}

	if len(bytes) == 0 {
		*a = []string{}
		return nil
	}

	return json.Unmarshal(bytes, a)
}

// Value implements the driver.Valuer interface.
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

// APIKey represents an API key for accessing the relay service.
type APIKey struct {
	ID                      int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	KeyHash                 string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	KeyPrefix               string         `gorm:"type:varchar(20);not null" json:"key_prefix"`
	Name                    string         `gorm:"type:varchar(255);not null" json:"name"`
	IsActive                bool           `gorm:"type:boolean;default:true;not null" json:"is_active"`
	ExpiresAt               *time.Time     `gorm:"type:timestamp" json:"expires_at"`
	MaxConcurrentRequests   int            `gorm:"type:int;default:5;not null" json:"max_concurrent_requests"`
	BoundCodexAccountID     *int64         `gorm:"type:bigint" json:"bound_codex_account_id,omitempty"`
	RateLimitPerMinute      int            `gorm:"type:int;default:60;not null" json:"rate_limit_per_minute"`
	RateLimitPerHour        int            `gorm:"type:int;default:3600;not null" json:"rate_limit_per_hour"`
	RateLimitPerDay         int            `gorm:"type:int;default:86400;not null" json:"rate_limit_per_day"`
	DailyCostLimit          float64        `gorm:"type:decimal(20,6);default:0;not null" json:"daily_cost_limit"`
	WeeklyCostLimit         float64        `gorm:"type:decimal(20,6);default:0;not null" json:"weekly_cost_limit"`
	MonthlyCostLimit        float64        `gorm:"type:decimal(20,6);default:0;not null" json:"monthly_cost_limit"`
	TotalCostLimit          float64        `gorm:"type:decimal(20,6);default:0;not null" json:"total_cost_limit"`
	EnableModelRestriction  bool           `gorm:"type:boolean;default:false;not null" json:"enable_model_restriction"`
	RestrictedModels        StringArray    `gorm:"type:jsonb;default:'[]'" json:"restricted_models"`
	EnableClientRestriction bool           `gorm:"type:boolean;default:false;not null" json:"enable_client_restriction"`
	AllowedClients          StringArray    `gorm:"type:jsonb;default:'[]'" json:"allowed_clients"`
	TotalRequests           int64          `gorm:"type:bigint;default:0;not null" json:"total_requests"`
	TotalTokens             int64          `gorm:"type:bigint;default:0;not null" json:"total_tokens"`
	TotalCost               float64        `gorm:"type:decimal(20,6);default:0;not null" json:"total_cost"`
	CreatedAt               time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt               time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt               gorm.DeletedAt `gorm:"type:timestamp;index" json:"-"`
}

// TableName specifies the table name for the APIKey model.
func (APIKey) TableName() string {
	return "api_keys"
}
