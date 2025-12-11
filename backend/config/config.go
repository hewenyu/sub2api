package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Security  SecurityConfig  `mapstructure:"security"`
	Claude    ClaudeConfig    `mapstructure:"claude"`
	Codex     CodexConfig     `mapstructure:"codex"`
	Limits    LimitsConfig    `mapstructure:"limits"`
	Billing   BillingConfig   `mapstructure:"billing"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Retry     RetryConfig     `mapstructure:"retry"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host" validate:"required"`
	Port         int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	Mode         string        `mapstructure:"mode" validate:"required,oneof=debug release test"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required,min=1s"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required,min=1s"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host" validate:"required"`
	Port            int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	User            string `mapstructure:"user" validate:"required"`
	Password        string `mapstructure:"password" validate:"required"`
	Database        string `mapstructure:"database" validate:"required"`
	SSLMode         string `mapstructure:"ssl_mode" validate:"required,oneof=disable require verify-ca verify-full"`
	MaxOpenConns    int    `mapstructure:"max_open_conns" validate:"min=1,max=1000"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns" validate:"min=1,max=100"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime string `mapstructure:"conn_max_idle_time" validate:"required"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db" validate:"min=0,max=15"`
	PoolSize int    `mapstructure:"pool_size" validate:"min=1,max=1000"`
}

type SecurityConfig struct {
	JWTSecret       string        `mapstructure:"jwt_secret" validate:"required,min=32"`
	EncryptionKey   string        `mapstructure:"encryption_key" validate:"required,len=32"`
	TokenExpiration time.Duration `mapstructure:"token_expiration" validate:"required,min=1h"`
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
	DefaultConcurrentRequests int `mapstructure:"default_concurrent_requests" validate:"min=1,max=1000"`
	DefaultRateLimitPerMinute int `mapstructure:"default_rate_limit_per_minute" validate:"min=1"`
	DefaultRateLimitPerHour   int `mapstructure:"default_rate_limit_per_hour" validate:"min=1"`
	DefaultRateLimitPerDay    int `mapstructure:"default_rate_limit_per_day" validate:"min=1"`
}

type BillingConfig struct {
	PricingURL     string        `mapstructure:"pricing_url"`
	UpdateInterval time.Duration `mapstructure:"update_interval"`
}

type LoggingConfig struct {
	Level            string            `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	Format           string            `mapstructure:"format" validate:"required,oneof=json console"`
	Output           string            `mapstructure:"output"`
	OutputPath       string            `mapstructure:"output_path" validate:"required"`
	ErrorOutputPath  string            `mapstructure:"error_output_path" validate:"required"`
	EnableCaller     bool              `mapstructure:"enable_caller"`
	EnableStacktrace bool              `mapstructure:"enable_stacktrace"`
	ComponentLevels  map[string]string `mapstructure:"component_levels"`
	LogPayloads      bool              `mapstructure:"log_payloads"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

type SchedulerConfig struct {
	Strategy   string        `mapstructure:"strategy" validate:"required,oneof=priority round_robin weighted consistent_hash health_aware"`
	SessionTTL time.Duration `mapstructure:"session_ttl" validate:"required,min=1m"`
}

type RetryConfig struct {
	MaxAttempts    int           `mapstructure:"max_attempts"`
	InitialBackoff time.Duration `mapstructure:"initial_backoff"`
	MaxBackoff     time.Duration `mapstructure:"max_backoff"`
	Multiplier     float64       `mapstructure:"multiplier"`
	Jitter         float64       `mapstructure:"jitter"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			Mode:         "release",
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    10,
			ConnMaxLifetime: "5m",
			ConnMaxIdleTime: "10m",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			DB:       0,
			PoolSize: 100,
		},
		Security: SecurityConfig{
			TokenExpiration: 24 * time.Hour,
		},
		Logging: LoggingConfig{
			Level:           "info",
			Format:          "json",
			OutputPath:      "stdout",
			ErrorOutputPath: "stderr",
		},
		Scheduler: SchedulerConfig{
			Strategy:   "priority",
			SessionTTL: 1 * time.Hour,
		},
		Limits: LimitsConfig{
			DefaultConcurrentRequests: 10,
			DefaultRateLimitPerMinute: 60,
			DefaultRateLimitPerHour:   3600,
			DefaultRateLimitPerDay:    86400,
		},
		Retry: RetryConfig{
			MaxAttempts:    3,
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     5 * time.Second,
			Multiplier:     2.0,
			Jitter:         0.1,
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("RELAY")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
