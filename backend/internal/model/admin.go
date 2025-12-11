package model

import (
	"time"

	"gorm.io/gorm"
)

// Admin represents an administrator user.
type Admin struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex" json:"email"`
	IsActive     bool           `gorm:"type:boolean;default:true;not null" json:"is_active"`
	CreatedAt    time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"type:timestamp;index" json:"-"`
}

// TableName specifies the table name for the Admin model.
func (Admin) TableName() string {
	return "admins"
}
