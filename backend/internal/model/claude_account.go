package model

import (
	"time"

	"gorm.io/gorm"
)

// ClaudeAccount represents a Claude API account.
type ClaudeAccount struct {
	ID                 int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Email              string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	AccessToken        string         `gorm:"type:text;not null" json:"-"`
	RefreshToken       string         `gorm:"type:text;not null" json:"-"`
	ExpiresAt          time.Time      `gorm:"type:timestamp;not null" json:"expires_at"`
	IsActive           bool           `gorm:"type:boolean;default:true;not null" json:"is_active"`
	IsSchedulable      bool           `gorm:"type:boolean;default:true;not null" json:"is_schedulable"`
	ConcurrentRequests int            `gorm:"type:int;default:0;not null" json:"concurrent_requests"`
	RateLimitedUntil   *time.Time     `gorm:"type:timestamp" json:"rate_limited_until"`
	OverloadUntil      *time.Time     `gorm:"type:timestamp" json:"overload_until"`
	TotalRequests      int64          `gorm:"type:bigint;default:0;not null" json:"total_requests"`
	TotalTokens        int64          `gorm:"type:bigint;default:0;not null" json:"total_tokens"`
	Features           string         `gorm:"type:jsonb" json:"features"`
	CreatedAt          time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"type:timestamp;index" json:"-"`
}

// TableName specifies the table name for the ClaudeAccount model.
func (ClaudeAccount) TableName() string {
	return "claude_accounts"
}
