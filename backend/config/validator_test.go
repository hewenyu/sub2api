package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidator_Validate(t *testing.T) {
	validator := NewValidator()

	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{
				Host:         "0.0.0.0",
				Port:         8080,
				Mode:         "release",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
			},
			Database: DatabaseConfig{
				Host:            "localhost",
				Port:            5432,
				User:            "test",
				Password:        "test",
				Database:        "test",
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
				PoolSize: 10,
			},
			Security: SecurityConfig{
				JWTSecret:       "this-is-a-very-long-jwt-secret-key-for-testing",
				EncryptionKey:   "12345678901234567890123456789012",
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
		}

		err := validator.Validate(cfg)
		assert.NoError(t, err)
	})

	t.Run("invalid port", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Server.Port = 0
		cfg.Security.JWTSecret = "this-is-a-very-long-jwt-secret-key-for-testing"
		cfg.Security.EncryptionKey = "12345678901234567890123456789012"
		cfg.Database.User = "test"
		cfg.Database.Password = "test"
		cfg.Database.Database = "test"

		err := validator.Validate(cfg)
		assert.Error(t, err)
	})

	t.Run("invalid mode", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Server.Mode = "invalid"
		cfg.Security.JWTSecret = "this-is-a-very-long-jwt-secret-key-for-testing"
		cfg.Security.EncryptionKey = "12345678901234567890123456789012"
		cfg.Database.User = "test"
		cfg.Database.Password = "test"
		cfg.Database.Database = "test"

		err := validator.Validate(cfg)
		assert.Error(t, err)
	})

	t.Run("invalid encryption key length", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Security.JWTSecret = "this-is-a-very-long-jwt-secret-key-for-testing"
		cfg.Security.EncryptionKey = "short"
		cfg.Database.User = "test"
		cfg.Database.Password = "test"
		cfg.Database.Database = "test"

		err := validator.Validate(cfg)
		assert.Error(t, err)
	})

	t.Run("max_idle_conns exceeds max_open_conns", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Database.MaxOpenConns = 10
		cfg.Database.MaxIdleConns = 20
		cfg.Security.JWTSecret = "this-is-a-very-long-jwt-secret-key-for-testing"
		cfg.Security.EncryptionKey = "12345678901234567890123456789012"
		cfg.Database.User = "test"
		cfg.Database.Password = "test"
		cfg.Database.Database = "test"

		err := validator.Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max_idle_conns cannot exceed max_open_conns")
	})

	t.Run("invalid log level", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Logging.Level = "invalid"
		cfg.Security.JWTSecret = "this-is-a-very-long-jwt-secret-key-for-testing"
		cfg.Security.EncryptionKey = "12345678901234567890123456789012"
		cfg.Database.User = "test"
		cfg.Database.Password = "test"
		cfg.Database.Database = "test"

		err := validator.Validate(cfg)
		assert.Error(t, err)
	})

	t.Run("invalid scheduler strategy", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Scheduler.Strategy = "invalid"
		cfg.Security.JWTSecret = "this-is-a-very-long-jwt-secret-key-for-testing"
		cfg.Security.EncryptionKey = "12345678901234567890123456789012"
		cfg.Database.User = "test"
		cfg.Database.Password = "test"
		cfg.Database.Database = "test"

		err := validator.Validate(cfg)
		assert.Error(t, err)
	})
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "release", cfg.Server.Mode)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "priority", cfg.Scheduler.Strategy)
	assert.Equal(t, 1*time.Hour, cfg.Scheduler.SessionTTL)
}
