package dialect

import "gorm.io/gorm"

// Dialect 数据库方言接口，用于抽象不同数据库之间的差异
type Dialect interface {
	// Name 返回数据库方言名称
	Name() string

	// GetDialector 返回GORM数据库方言实现
	GetDialector(dsn string) gorm.Dialector

	// CaseInsensitiveLike 返回不区分大小写的LIKE操作符
	// PostgreSQL: "ILIKE"
	// SQLite: "LIKE" (默认不区分大小写，仅ASCII)
	CaseInsensitiveLike() string

	// JSONExtract 返回JSON字段值提取的SQL表达式
	// PostgreSQL: column->>'key'
	// SQLite: json_extract(column, '$.key')
	JSONExtract(column, key string) string

	// SupportsArrayType 返回是否支持数组数据类型
	SupportsArrayType() bool
}
