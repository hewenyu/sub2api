package dialect

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SQLiteDialect SQLite数据库方言实现
type SQLiteDialect struct{}

// NewSQLiteDialect 创建SQLite方言实例
func NewSQLiteDialect() *SQLiteDialect {
	return &SQLiteDialect{}
}

// Name 返回数据库方言名称
func (d *SQLiteDialect) Name() string {
	return "sqlite"
}

// GetDialector 返回SQLite的GORM方言实现
func (d *SQLiteDialect) GetDialector(dsn string) gorm.Dialector {
	return sqlite.Open(dsn)
}

// CaseInsensitiveLike 返回SQLite的不区分大小写LIKE操作符
// SQLite的LIKE操作符默认对ASCII字符不区分大小写
func (d *SQLiteDialect) CaseInsensitiveLike() string {
	return "LIKE"
}

// JSONExtract 返回SQLite的JSON字段提取表达式
// 使用 json_extract() 函数提取JSON字段的值
func (d *SQLiteDialect) JSONExtract(column, key string) string {
	return fmt.Sprintf("json_extract(%s, '$.%s')", column, key)
}

// SupportsArrayType 返回SQLite是否支持数组类型
// SQLite不支持原生数组类型
func (d *SQLiteDialect) SupportsArrayType() bool {
	return false
}
