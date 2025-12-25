package repository

import (
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/infrastructure/dialect"
	"gorm.io/gorm"
)

// DBHelper 数据库方言感知的辅助函数
// 用于抽象不同数据库之间的SQL差异
type DBHelper struct {
	dialect dialect.Dialect
}

// NewDBHelper 创建数据库助手实例
// 从GORM的Dialector中自动识别数据库类型
func NewDBHelper(db *gorm.DB) *DBHelper {
	dialectName := db.Dialector.Name()
	d, err := dialect.GetDialect(dialectName)
	if err != nil {
		// 如果无法识别方言，默认使用PostgreSQL
		d, _ = dialect.GetDialect("postgres")
	}
	return &DBHelper{dialect: d}
}

// CaseInsensitiveLike 返回不区分大小写的LIKE查询条件
// 使用示例: db.Where(helper.CaseInsensitiveLike("email"), "%test%")
func (h *DBHelper) CaseInsensitiveLike(column string) string {
	op := h.dialect.CaseInsensitiveLike()
	return fmt.Sprintf("%s %s ?", column, op)
}

// CaseInsensitiveLikeMultiple 返回多列不区分大小写的LIKE查询条件（OR连接）
// 使用示例: db.Where(helper.CaseInsensitiveLikeMultiple("email", "username"), "%test%", "%test%")
func (h *DBHelper) CaseInsensitiveLikeMultiple(columns ...string) string {
	if len(columns) == 0 {
		return ""
	}

	op := h.dialect.CaseInsensitiveLike()
	var conditions string
	for i, col := range columns {
		if i > 0 {
			conditions += " OR "
		}
		conditions += fmt.Sprintf("%s %s ?", col, op)
	}
	return conditions
}

// JSONExtract 返回JSON字段值提取的WHERE子句
// 使用示例: db.Where(helper.JSONExtract("extra", "crs_account_id") + " = ?", "value")
func (h *DBHelper) JSONExtract(column, key string) string {
	return h.dialect.JSONExtract(column, key)
}

// SupportsArrayType 返回当前数据库是否支持原生数组类型
func (h *DBHelper) SupportsArrayType() bool {
	return h.dialect.SupportsArrayType()
}

// GetDialectName 返回当前数据库方言名称
func (h *DBHelper) GetDialectName() string {
	return h.dialect.Name()
}
