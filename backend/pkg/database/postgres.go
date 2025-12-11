package database

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Wei-Shaw/sub2api/backend/config"
)

// DB represents the database connection.
type DB struct {
	*gorm.DB
}

// NewPostgresDB initializes a new PostgreSQL database connection with connection pooling.
func NewPostgresDB(cfg *config.DatabaseConfig, log *zap.Logger) (*DB, error) {
	// Build DSN
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)

	// Configure GORM logger to use zap
	var gormLogger logger.Interface
	if log != nil {
		// Use custom zap logger with 200ms slow query threshold
		gormLogger = NewGormLogger(log, 200*time.Millisecond)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	// Parse durations
	if cfg.ConnMaxLifetime != "" {
		duration, err := time.ParseDuration(cfg.ConnMaxLifetime)
		if err != nil {
			return nil, fmt.Errorf("invalid conn_max_lifetime: %w", err)
		}
		sqlDB.SetConnMaxLifetime(duration)
	}

	if cfg.ConnMaxIdleTime != "" {
		duration, err := time.ParseDuration(cfg.ConnMaxIdleTime)
		if err != nil {
			return nil, fmt.Errorf("invalid conn_max_idle_time: %w", err)
		}
		sqlDB.SetConnMaxIdleTime(duration)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if log != nil {
		log.Info("Database connection established",
			zap.String("host", cfg.Host),
			zap.Int("port", cfg.Port),
			zap.String("database", cfg.Database),
			zap.Int("max_open_conns", cfg.MaxOpenConns),
			zap.Int("max_idle_conns", cfg.MaxIdleConns),
		)
	}

	return &DB{DB: db}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.Close()
}

// HealthCheck performs a health check on the database connection.
func (db *DB) HealthCheck(ctx context.Context) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.PingContext(ctx)
}
