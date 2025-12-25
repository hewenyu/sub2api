package infrastructure

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/infrastructure/dialect"
	"github.com/Wei-Shaw/sub2api/internal/model"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	// 初始化时区（在数据库连接之前，确保时区设置正确）
	if err := timezone.Init(cfg.Timezone); err != nil {
		return nil, err
	}

	// 获取数据库方言
	dbDialect, err := dialect.GetDialect(cfg.Database.Type)
	if err != nil {
		return nil, fmt.Errorf("get database dialect failed: %w", err)
	}

	// GORM配置
	gormConfig := &gorm.Config{}
	if cfg.Server.Mode == "debug" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	// 构建DSN
	dsn := cfg.Database.DSNWithTimezone(cfg.Timezone)

	// SQLite特定: 确保数据目录存在
	if dbDialect.Name() == "sqlite" {
		dir := filepath.Dir(dsn)
		// 移除查询参数（如 ?_fk=1）
		if idx := filepath.Base(dsn); idx != "." {
			dir = filepath.Dir(dsn[:len(dsn)-len("?_fk=1")])
		}
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("create sqlite data directory failed: %w", err)
			}
		}
	}

	// 连接数据库
	db, err := gorm.Open(dbDialect.GetDialector(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("connect to %s database failed: %w", dbDialect.Name(), err)
	}

	// SQLite特定配置
	if dbDialect.Name() == "sqlite" {
		// 启用WAL模式（Write-Ahead Logging），提升并发性能
		if err := db.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
			return nil, fmt.Errorf("enable WAL mode failed: %w", err)
		}
		// 启用外键约束（SQLite默认禁用）
		if err := db.Exec("PRAGMA foreign_keys=ON").Error; err != nil {
			return nil, fmt.Errorf("enable foreign keys failed: %w", err)
		}
	}

	// 自动迁移（始终执行，确保数据库结构与代码同步）
	// GORM 的 AutoMigrate 只会添加新字段，不会删除或修改已有字段，是安全的
	if err := model.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("auto migrate failed: %w", err)
	}

	return db, nil
}
