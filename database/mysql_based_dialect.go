package database

import (
	"database/sql"
	"fmt"
	"strings"
)

func GetMysqlBasedDialectByType(dbType string, db *sql.DB) MysqlBasedDialect {
	switch dbType {
	case "dameng":
		return NewMysqlBasedDamengDialect(db)
	case "openguass":
		return NewMysqlBasedOpenguassDialect(db)
	case "vastbase":
		return NewMysqlBasedVastbaseDialect(db)
	case "kingbase":
		return NewMysqlBasedKingbaseDialect(db)
	case "oceandb":
		return NewMysqlBasedOceandbDialect(db)
	}
	return nil
}

type MysqlBasedDialect interface {
	GetDatabases() ([]string, error)
	GetTables() ([]string, error)
	GetTableSchema(tableName string) (string, error)
	GetTableColumns(tableName string) ([]ColumnInfo, error)
}

type BaseMysqlBasedDialect struct {
	db *sql.DB
}

func NewBaseMysqlBasedDialect(db *sql.DB) *BaseMysqlBasedDialect {
	return &BaseMysqlBasedDialect{db: db}
}

func (m *BaseMysqlBasedDialect) GetDatabases() ([]string, error) {
	var databases []string
	rows, err := m.db.Query("SHOW DATABASES")
	if err != nil {
		return nil, fmt.Errorf("failed to query database list: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var database string
		if err := rows.Scan(&database); err != nil {
			return nil, err
		}
		databases = append(databases, database)
	}
	return databases, rows.Err()
}

// GetTables 获取所有表名
func (m *BaseMysqlBasedDialect) GetTables() ([]string, error) {
	var tables []string
	rows, err := m.db.Query("SHOW TABLES")
	if err != nil {
		return nil, fmt.Errorf("failed to query table list: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, rows.Err()
}

// GetTableSchema 获取表结构
func (m *BaseMysqlBasedDialect) GetTableSchema(tableName string) (string, error) {
	query := fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName)
	var schemas []string
	rows, err := m.db.Query(query)
	if err != nil {
		return "", fmt.Errorf("failed to query table schema: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return "", err
		}
		schemas = append(schemas, schema)
	}
	return strings.Join(schemas, "\n"), rows.Err()
}

// GetTableColumns 获取表的列信息
func (m *BaseMysqlBasedDialect) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf("DESCRIBE `%s`", tableName)
	var rows []map[string]interface{}
	scanRows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query column information: %w", err)
	}
	defer scanRows.Close()
	for scanRows.Next() {
		var r map[string]interface{}
		if err := scanRows.Scan(&r); err != nil {
			return nil, err
		}
		rows = append(rows, r)
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
