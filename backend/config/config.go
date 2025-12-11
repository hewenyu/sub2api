package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Security SecurityConfig `mapstructure:"security"`
	Claude   ClaudeConfig   `mapstructure:"claude"`
	Codex    CodexConfig    `mapstructure:"codex"`
	Limits   LimitsConfig   `mapstructure:"limits"`
	Billing  BillingConfig  `mapstructure:"billing"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime string `mapstructure:"conn_max_idle_time"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type SecurityConfig struct {
	JWTSecret       string        `mapstructure:"jwt_secret"`
	EncryptionKey   string        `mapstructure:"encryption_key"`
	TokenExpiration time.Duration `mapstructure:"token_expiration"`
}

type ClaudeConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURI  string `mapstructure:"redirect_uri"`
	BaseAPI      string `mapstructure:"base_api"`
	AuthURL      string `mapstructure:"auth_url"`
	TokenURL     string `mapstructure:"token_url"`
}

type CodexConfig struct {
	BaseAPI string           `mapstructure:"base_api"`
	OAuth   CodexOAuthConfig `mapstructure:"oauth"`
}

type CodexOAuthConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	AuthURL      string `mapstructure:"auth_url"`
	TokenURL     string `mapstructure:"token_url"`
	Scopes       string `mapstructure:"scopes"`
}


type LimitsConfig struct {
	DefaultConcurrentRequests int `mapstructure:"default_concurrent_requests"`
	DefaultRateLimitPerMinute int `mapstructure:"default_rate_limit_per_minute"`
	DefaultRateLimitPerHour   int `mapstructure:"default_rate_limit_per_hour"`
	DefaultRateLimitPerDay    int `mapstructure:"default_rate_limit_per_day"`
}

type BillingConfig struct {
	PricingURL     string        `mapstructure:"pricing_url"`
	UpdateInterval time.Duration `mapstructure:"update_interval"`
}

type LoggingConfig struct {
	Level           string `mapstructure:"level"`
	Format          string `mapstructure:"format"`
	OutputPath      string `mapstructure:"output_path"`
	ErrorOutputPath string `mapstructure:"error_output_path"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

// LoadConfig loads configuration from the specified YAML file.
// It supports environment variable overrides with the RELAY_ prefix.
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// Automatic environment variable binding
	viper.AutomaticEnv()
	viper.SetEnvPrefix("RELAY")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// validateConfig validates that all required configuration fields are set correctly.
func validateConfig(cfg *Config) error {
	if cfg.Server.Port == 0 {
		return fmt.Errorf("server.port is required")
	}
	if cfg.Security.JWTSecret == "" {
		return fmt.Errorf("security.jwt_secret is required")
	}
	if cfg.Security.EncryptionKey == "" {
		return fmt.Errorf("security.encryption_key is required")
	}
	if len(cfg.Security.EncryptionKey) != 32 {
		return fmt.Errorf("security.encryption_key must be 32 characters")
	}
	return nil
}
