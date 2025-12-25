package dialect

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresDialect PostgreSQL数据库方言实现
type PostgresDialect struct{}

// NewPostgresDialect 创建PostgreSQL方言实例
func NewPostgresDialect() *PostgresDialect {
	return &PostgresDialect{}
}

// Name 返回数据库方言名称
func (d *PostgresDialect) Name() string {
	return "postgres"
}

// GetDialector 返回PostgreSQL的GORM方言实现
func (d *PostgresDialect) GetDialector(dsn string) gorm.Dialector {
	return postgres.Open(dsn)
}

// CaseInsensitiveLike 返回PostgreSQL的不区分大小写LIKE操作符
func (d *PostgresDialect) CaseInsensitiveLike() string {
	return "ILIKE"
}

// JSONExtract 返回PostgreSQL的JSON字段提取表达式
// 使用 ->> 操作符提取JSON字段的文本值
func (d *PostgresDialect) JSONExtract(column, key string) string {
	return fmt.Sprintf("%s->>'%s'", column, key)
}

// SupportsArrayType 返回PostgreSQL是否支持数组类型
func (d *PostgresDialect) SupportsArrayType() bool {
	return true
}
