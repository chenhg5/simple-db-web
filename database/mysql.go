package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// MySQL 实现Database接口
type MySQL struct {
	db *sql.DB
}

// NewMySQL 创建MySQL实例
func NewMySQL() *MySQL {
	return &MySQL{}
}

// Connect 建立MySQL连接
func (m *MySQL) Connect(dsn string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	oldDB := m.db
	m.db = db
	if oldDB != nil {
		oldDB.Close()
	}
	return nil
}

// Close 关闭连接
func (m *MySQL) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// GetTables 获取所有表名
func (m *MySQL) GetTables() ([]string, error) {
	rows, err := m.db.Query("SHOW TABLES")
	if err != nil {
		return nil, fmt.Errorf("查询表列表失败: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}
	return tables, rows.Err()
}

// GetTableSchema 获取表结构
func (m *MySQL) GetTableSchema(tableName string) (string, error) {
	rows, err := m.db.Query(fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName))
	if err != nil {
		return "", fmt.Errorf("查询表结构失败: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", fmt.Errorf("表 %s 不存在", tableName)
	}

	var table, createTable string
	if err := rows.Scan(&table, &createTable); err != nil {
		return "", err
	}

	return createTable, nil
}

// GetTableColumns 获取表的列信息
func (m *MySQL) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf("DESCRIBE `%s`", tableName)
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询列信息失败: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var null, key, extra string
		var defaultVal sql.NullString

		if err := rows.Scan(&col.Name, &col.Type, &null, &key, &defaultVal, &extra); err != nil {
			return nil, err
		}

		col.Nullable = (null == "YES")
		col.Key = key
		if defaultVal.Valid {
			col.DefaultValue = defaultVal.String
		}

		columns = append(columns, col)
	}
	return columns, rows.Err()
}

// ExecuteQuery 执行查询
func (m *MySQL) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("执行查询失败: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results = make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return results, rows.Err()
}

// ExecuteUpdate 执行更新
func (m *MySQL) ExecuteUpdate(query string) (int64, error) {
	result, err := m.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行更新失败: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteDelete 执行删除
func (m *MySQL) ExecuteDelete(query string) (int64, error) {
	result, err := m.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行删除失败: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteInsert 执行插入
func (m *MySQL) ExecuteInsert(query string) (int64, error) {
	result, err := m.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行插入失败: %w", err)
	}
	return result.RowsAffected()
}

// GetTableData 获取表数据（分页）
func (m *MySQL) GetTableData(tableName string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	// 获取总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	if err := m.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT %d OFFSET %d", tableName, pageSize, offset)

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, 0, fmt.Errorf("查询数据失败: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, 0, err
	}

	var results = make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, 0, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else if val == nil {
				row[col] = nil
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return results, total, rows.Err()
}

// GetDatabases 获取所有数据库名称
func (m *MySQL) GetDatabases() ([]string, error) {
	rows, err := m.db.Query("SHOW DATABASES")
	if err != nil {
		return nil, fmt.Errorf("查询数据库列表失败: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		// 过滤掉系统数据库
		if dbName != "information_schema" && dbName != "performance_schema" &&
			dbName != "mysql" && dbName != "sys" {
			databases = append(databases, dbName)
		}
	}
	return databases, rows.Err()
}

// SwitchDatabase 切换当前使用的数据库
func (m *MySQL) SwitchDatabase(databaseName string) error {
	_, err := m.db.Exec(fmt.Sprintf("USE `%s`", databaseName))
	if err != nil {
		return fmt.Errorf("切换数据库失败: %w", err)
	}
	return nil
}

// BuildDSN 根据连接信息构建DSN
func BuildDSN(info ConnectionInfo) string {
	if info.DSN != "" {
		return info.DSN
	}

	// MySQL DSN格式: user:password@tcp(host:port)/database
	// 如果Database为空,则不指定数据库,用于先连接到服务器
	var dsn string
	if info.Database != "" {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			info.User,
			info.Password,
			info.Host,
			info.Port,
			info.Database,
		)
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/",
			info.User,
			info.Password,
			info.Host,
			info.Port,
		)
	}

	// 添加参数
	params := []string{
		"charset=utf8mb4",
		"parseTime=True",
		"loc=Local",
	}

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}
