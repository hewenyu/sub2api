package model

import (
	"time"

	"gorm.io/gorm"
)

// ProxyConfig represents a proxy configuration for API requests.
// Passwords are encrypted using AES-256-CBC (format: {base64(iv)}:{base64(ciphertext)})
type ProxyConfig struct {
	ID       int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Enabled  bool   `gorm:"type:boolean;default:true;not null" json:"enabled"`
	Protocol string `gorm:"type:varchar(10);not null" json:"protocol"` // http, https, socks5
	Host     string `gorm:"type:varchar(255);not null" json:"host"`
	Port     int    `gorm:"type:int;not null" json:"port"`

	// Optional authentication
	Username *string `gorm:"type:varchar(255)" json:"username,omitempty"`
	Password *string `gorm:"type:text" json:"-"` // Encrypted password

	// Default proxy flag
	IsDefault bool `gorm:"type:boolean;default:false;not null" json:"is_default"`

	// Timestamps
	CreatedAt time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamp;index" json:"-"`
}

// TableName specifies the table name for the ProxyConfig model.
func (ProxyConfig) TableName() string {
	return "proxy_configs"
}
