package database

import (
	"database/sql"
	"fmt"
	"strings"

	"ksogit.kingsoft.net/kgo/mysql"
)

type Dialect interface {
	GetTables() ([]string, error)
	GetTableSchema(tableName string) (string, error)
	GetTableColumns(tableName string) ([]ColumnInfo, error)
}

type MySQLDialect struct {
	db mysql.DBAdapter
}

func NewMySQLDialect(db mysql.DBAdapter) *MySQLDialect {
	return &MySQLDialect{db: db}
}

// GetTables 获取所有表名
func (m *MySQLDialect) GetTables() ([]string, error) {
	var rows []map[string]interface{}
	if err := m.db.Query(&rows, "SHOW TABLES"); err != nil {
		return nil, fmt.Errorf("查询表列表失败: %w", err)
	}
	var tables []string
	for _, row := range rows {
		for _, v := range row {
			switch val := v.(type) {
			case []byte:
				tables = append(tables, string(val))
			case string:
				tables = append(tables, val)
			default:
				tables = append(tables, fmt.Sprint(val))
			}
			break
		}
	}
	return tables, nil
}

// GetTableSchema 获取表结构
func (m *MySQLDialect) GetTableSchema(tableName string) (string, error) {
	query := fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName)
	var rows []map[string]interface{}
	if err := m.db.Query(&rows, query); err != nil {
		return "", fmt.Errorf("查询表结构失败: %w", err)
	}
	if len(rows) == 0 {
		return "", fmt.Errorf("表 %s 不存在", tableName)
	}
	for k, v := range rows[0] {
		if strings.EqualFold(k, "Create Table") {
			if b, ok := v.([]byte); ok {
				return string(b), nil
			}
			return fmt.Sprint(v), nil
		}
	}
	for _, v := range rows[0] {
		if b, ok := v.([]byte); ok {
			return string(b), nil
		}
		return fmt.Sprint(v), nil
	}
	return "", fmt.Errorf("未能获取表结构")
}

// GetTableColumns 获取表的列信息
func (m *MySQLDialect) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf("DESCRIBE `%s`", tableName)
	var rows []map[string]interface{}
	if err := m.db.Query(&rows, query); err != nil {
		return nil, fmt.Errorf("查询列信息失败: %w", err)
	}
	var columns []ColumnInfo
	for _, r := range rows {
		col := ColumnInfo{}
		if v, ok := r["Field"]; ok {
			if b, ok2 := v.([]byte); ok2 {
				col.Name = string(b)
			} else {
				col.Name = fmt.Sprint(v)
			}
		}
		if v, ok := r["Type"]; ok {
			if b, ok2 := v.([]byte); ok2 {
				col.Type = string(b)
			} else {
				col.Type = fmt.Sprint(v)
			}
		}
		if v, ok := r["Null"]; ok {
			var s string
			if b, ok2 := v.([]byte); ok2 {
				s = string(b)
			} else {
				s = fmt.Sprint(v)
			}
			col.Nullable = (strings.EqualFold(s, "YES"))
		}
		if v, ok := r["Key"]; ok {
			if b, ok2 := v.([]byte); ok2 {
				col.Key = string(b)
			} else {
				col.Key = fmt.Sprint(v)
			}
		}
		if v, ok := r["Default"]; ok {
			switch dv := v.(type) {
			case nil:
			case []byte:
				col.DefaultValue = string(dv)
			case sql.NullString:
				if dv.Valid {
					col.DefaultValue = dv.String
				}
			default:
				col.DefaultValue = fmt.Sprint(dv)
			}
		}
		columns = append(columns, col)
	}
	return columns, nil
}
