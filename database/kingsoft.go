package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"ksogit.kingsoft.net/chat/lib/xmysql"
	xmysqlv2 "ksogit.kingsoft.net/chat/lib/xmysql/v2"
	"ksogit.kingsoft.net/kgo/mysql"
)

type KingsoftDB struct {
	db      mysql.DBAdapter
	dialect Dialect
}

func NewKingsoftDB() *KingsoftDB {
	return &KingsoftDB{}
}

// Connect 建立MySQL连接
func (m *KingsoftDB) Connect(dsn string) error {
	dbConfig, err := xmysql.GetDatabaseFromDSN(dsn)
	if err != nil {
		return err
	}
	db, err := xmysqlv2.NewDBBuilder(dbConfig, &xmysqlv2.ServiceInfo{}).WithNameSuffix("master").Build(context.Background())
	if err != nil {
		return err
	}
	m.db = db
	m.dialect = NewMySQLDialect(db)
	return nil
}

// Close 关闭连接
func (m *KingsoftDB) Close() error {
	if m.db != nil {
		m.db.Close()
	}
	return nil
}

// GetTables 获取所有表名
func (m *KingsoftDB) GetTables() ([]string, error) {
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
func (m *KingsoftDB) GetTableSchema(tableName string) (string, error) {
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
func (m *KingsoftDB) GetTableColumns(tableName string) ([]ColumnInfo, error) {
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

// ExecuteQuery 执行查询
func (m *KingsoftDB) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	var rows []map[string]interface{}
	if err := m.db.Query(&rows, query); err != nil {
		return nil, fmt.Errorf("执行查询失败: %w", err)
	}
	for i := range rows {
		for k, v := range rows[i] {
			if b, ok := v.([]byte); ok {
				rows[i][k] = string(b)
			}
		}
	}
	return rows, nil
}

// ExecuteUpdate 执行更新
func (m *KingsoftDB) ExecuteUpdate(query string) (int64, error) {
	res, err := m.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行更新失败: %w", err)
	}
	return res.RowsAffected, nil
}

// ExecuteDelete 执行删除
func (m *KingsoftDB) ExecuteDelete(query string) (int64, error) {
	res, err := m.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行删除失败: %w", err)
	}
	return res.RowsAffected, nil
}

// ExecuteInsert 执行插入
func (m *KingsoftDB) ExecuteInsert(query string) (int64, error) {
	res, err := m.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行插入失败: %w", err)
	}
	return res.RowsAffected, nil
}

// GetTableData 获取表数据（分页）
func (m *KingsoftDB) GetTableData(tableName string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	total, err := m.db.QueryInt64(countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT %d OFFSET %d", tableName, pageSize, offset)
	var rows []map[string]interface{}
	if err := m.db.Query(&rows, query); err != nil {
		return nil, 0, fmt.Errorf("查询数据失败: %w", err)
	}
	for i := range rows {
		for k, v := range rows[i] {
			if b, ok := v.([]byte); ok {
				rows[i][k] = string(b)
			} else if v == nil {
				rows[i][k] = nil
			}
		}
	}
	return rows, total, nil
}
