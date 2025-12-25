package dialect

import (
	"fmt"
	"strings"
)

// GetDialect 根据数据库类型字符串返回对应的方言实现
// 支持的类型: postgres/postgresql, sqlite/sqlite3
func GetDialect(dbType string) (Dialect, error) {
	switch strings.ToLower(dbType) {
	case "postgres", "postgresql":
		return NewPostgresDialect(), nil
	case "sqlite", "sqlite3":
		return NewSQLiteDialect(), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: postgres, sqlite)", dbType)
	}
}
