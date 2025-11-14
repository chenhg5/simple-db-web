package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// H2 实现Database接口
// 注意：H2 是 Java 数据库，Go 中直接支持需要 JDBC 桥接
// 这里提供一个基础实现框架，实际使用时可能需要通过 JDBC 桥接或使用其他方式
type H2 struct {
	db *sql.DB
}

// NewH2 创建H2实例
func NewH2() *H2 {
	return &H2{}
}

// Connect 建立H2连接
// H2 DSN格式: jdbc:h2:tcp://host:port/database 或 jdbc:h2:file:/path/to/database
// 注意：实际使用时可能需要 JDBC 桥接库
func (h *H2) Connect(dsn string) error {
	// H2 在 Go 中直接支持较困难，因为它是 Java 数据库
	// 这里提供一个占位实现
	// 实际使用时可能需要通过 JDBC 桥接（如 go-java）或使用其他方式
	return fmt.Errorf("H2 数据库在 Go 中需要 JDBC 桥接支持，当前版本暂不支持直接连接")
}

// Close 关闭连接
func (h *H2) Close() error {
	if h.db != nil {
		return h.db.Close()
	}
	return nil
}

// GetTables 获取所有表名
func (h *H2) GetTables() ([]string, error) {
	if h.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	query := `SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'PUBLIC' ORDER BY TABLE_NAME`
	rows, err := h.db.Query(query)
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
func (h *H2) GetTableSchema(tableName string) (string, error) {
	if h.db == nil {
		return "", fmt.Errorf("数据库未连接")
	}
	query := fmt.Sprintf(`
		SELECT 'CREATE TABLE ' || TABLE_NAME || ' (' ||
			GROUP_CONCAT(
				COLUMN_NAME || ' ' || 
				TYPE_NAME || 
				CASE 
					WHEN CHARACTER_MAXIMUM_LENGTH IS NOT NULL THEN '(' || CHARACTER_MAXIMUM_LENGTH || ')'
					ELSE ''
				END ||
				CASE 
					WHEN IS_NULLABLE = 'NO' THEN ' NOT NULL'
					ELSE ''
				END ||
				CASE 
					WHEN COLUMN_DEFAULT IS NOT NULL THEN ' DEFAULT ' || COLUMN_DEFAULT
					ELSE ''
				END
				ORDER BY ORDINAL_POSITION
				SEPARATOR ', '
			) || ');' as create_table
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_NAME = '%s' AND TABLE_SCHEMA = 'PUBLIC'
		GROUP BY TABLE_NAME
	`, strings.ToUpper(tableName))

	rows, err := h.db.Query(query)
	if err != nil {
		return "", fmt.Errorf("查询表结构失败: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", fmt.Errorf("表 %s 不存在", tableName)
	}

	var createTable string
	if err := rows.Scan(&createTable); err != nil {
		return "", err
	}

	return createTable, nil
}

// GetTableColumns 获取表的列信息
func (h *H2) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	if h.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	query := fmt.Sprintf(`
		SELECT 
			COLUMN_NAME,
			TYPE_NAME || 
			CASE 
				WHEN CHARACTER_MAXIMUM_LENGTH IS NOT NULL THEN '(' || CHARACTER_MAXIMUM_LENGTH || ')'
				ELSE ''
			END as TYPE_NAME,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			CASE 
				WHEN EXISTS (
					SELECT 1 FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
					INNER JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu 
						ON tc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
					WHERE tc.TABLE_NAME = '%s' 
						AND kcu.COLUMN_NAME = c.COLUMN_NAME 
						AND tc.CONSTRAINT_TYPE = 'PRIMARY KEY'
				) THEN 'PRI'
				ELSE ''
			END as KEY_TYPE
		FROM INFORMATION_SCHEMA.COLUMNS c
		WHERE TABLE_NAME = '%s' AND TABLE_SCHEMA = 'PUBLIC'
		ORDER BY ORDINAL_POSITION
	`, strings.ToUpper(tableName), strings.ToUpper(tableName))

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询列信息失败: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var nullable string
		var defaultVal sql.NullString

		if err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultVal, &col.Key); err != nil {
			return nil, err
		}

		col.Nullable = (nullable == "YES")
		if defaultVal.Valid {
			col.DefaultValue = defaultVal.String
		}

		columns = append(columns, col)
	}
	return columns, rows.Err()
}

// ExecuteQuery 执行查询
func (h *H2) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	if h.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	rows, err := h.db.Query(query)
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
func (h *H2) ExecuteUpdate(query string) (int64, error) {
	if h.db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}
	result, err := h.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行更新失败: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteDelete 执行删除
func (h *H2) ExecuteDelete(query string) (int64, error) {
	if h.db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}
	result, err := h.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行删除失败: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteInsert 执行插入
func (h *H2) ExecuteInsert(query string) (int64, error) {
	if h.db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}
	result, err := h.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行插入失败: %w", err)
	}
	return result.RowsAffected()
}

// GetTableData 获取表数据（分页）
func (h *H2) GetTableData(tableName string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	if h.db == nil {
		return nil, 0, fmt.Errorf("数据库未连接")
	}
	// 获取总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", strings.ToUpper(tableName))
	if err := h.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", strings.ToUpper(tableName), pageSize, offset)

	rows, err := h.db.Query(query)
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

// GetTableDataByID 基于主键ID获取表数据（高性能分页）
func (h *H2) GetTableDataByID(tableName string, primaryKey string, lastId interface{}, pageSize int, direction string) ([]map[string]interface{}, int64, interface{}, error) {
	if h.db == nil {
		return nil, 0, nil, fmt.Errorf("数据库未连接")
	}
	// 获取总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", strings.ToUpper(tableName))
	if err := h.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, nil, fmt.Errorf("查询总数失败: %w", err)
	}

	// 构建基于ID的查询
	var query string
	var rows *sql.Rows
	var err error

	if direction == "prev" {
		if lastId == nil {
			return nil, 0, nil, fmt.Errorf("上一页需要提供lastId")
		}
		query = fmt.Sprintf(`
			SELECT * FROM (
				SELECT * FROM %s WHERE %s < ? ORDER BY %s DESC LIMIT %d
			) ORDER BY %s ASC
		`, strings.ToUpper(tableName), strings.ToUpper(primaryKey), strings.ToUpper(primaryKey), pageSize, strings.ToUpper(primaryKey))
		rows, err = h.db.Query(query, lastId)
	} else {
		if lastId == nil {
			query = fmt.Sprintf(`
				SELECT * FROM %s ORDER BY %s ASC LIMIT %d
			`, strings.ToUpper(tableName), strings.ToUpper(primaryKey), pageSize)
			rows, err = h.db.Query(query)
		} else {
			query = fmt.Sprintf(`
				SELECT * FROM %s WHERE %s > ? ORDER BY %s ASC LIMIT %d
			`, strings.ToUpper(tableName), strings.ToUpper(primaryKey), strings.ToUpper(primaryKey), pageSize)
			rows, err = h.db.Query(query, lastId)
		}
	}

	if err != nil {
		return nil, 0, nil, fmt.Errorf("查询数据失败: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, 0, nil, err
	}

	var results = make([]map[string]interface{}, 0)
	var nextId interface{} = nil
	var firstId interface{} = nil

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, 0, nil, err
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

		if idVal, ok := row[strings.ToUpper(primaryKey)]; ok {
			if firstId == nil {
				firstId = idVal
			}
			nextId = idVal
		}
	}

	if direction == "prev" {
		for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
			results[i], results[j] = results[j], results[i]
		}
		nextId = firstId
	}

	return results, total, nextId, rows.Err()
}

// GetPageIdByPageNumber 根据页码计算该页的起始ID（用于页码跳转）
func (h *H2) GetPageIdByPageNumber(tableName string, primaryKey string, page, pageSize int) (interface{}, error) {
	if h.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	if page <= 1 {
		return nil, nil
	}

	offset := (page - 1) * pageSize - 1
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		ORDER BY %s ASC
		LIMIT 1 OFFSET %d
	`, strings.ToUpper(primaryKey), strings.ToUpper(tableName), strings.ToUpper(primaryKey), offset)

	var id interface{}
	err := h.db.QueryRow(query).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询页码ID失败: %w", err)
	}

	return id, nil
}

// GetDatabases 获取所有数据库名称（H2 通常只有一个数据库）
func (h *H2) GetDatabases() ([]string, error) {
	if h.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	// H2 通常只有一个数据库，返回空列表或当前数据库名
	return []string{}, nil
}

// SwitchDatabase 切换当前使用的数据库（H2 不支持多数据库）
func (h *H2) SwitchDatabase(databaseName string) error {
	return fmt.Errorf("H2不支持切换数据库")
}

// BuildH2DSN 根据连接信息构建H2 DSN
func BuildH2DSN(info ConnectionInfo) string {
	if info.DSN != "" {
		return info.DSN
	}

	// H2 DSN格式: jdbc:h2:tcp://host:port/database 或 jdbc:h2:file:/path/to/database
	// 注意：实际使用时可能需要 JDBC 桥接
	var dsn string
	if info.Database != "" {
		if info.Host != "" {
			// TCP 模式
			dsn = fmt.Sprintf("jdbc:h2:tcp://%s:%s/%s",
				info.Host,
				info.Port,
				info.Database,
			)
		} else {
			// 文件模式
			dsn = fmt.Sprintf("jdbc:h2:file:%s", info.Database)
		}
	} else {
		// 内存模式
		dsn = "jdbc:h2:mem:"
	}

	// 添加用户名和密码（如果提供）
	if info.User != "" {
		dsn += ";USER=" + info.User
	}
	if info.Password != "" {
		dsn += ";PASSWORD=" + info.Password
	}

	return dsn
}

