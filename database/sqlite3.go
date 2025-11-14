package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// SQLite3 实现Database接口
type SQLite3 struct {
	db *sql.DB
}

// NewSQLite3 创建SQLite3实例
func NewSQLite3() *SQLite3 {
	return &SQLite3{}
}

// Connect 建立SQLite3连接
func (s *SQLite3) Connect(dsn string) error {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	s.db = db
	return nil
}

// Close 关闭连接
func (s *SQLite3) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetTypeName 获取数据库类型名称
func (s *SQLite3) GetTypeName() string {
	return "sqlite"
}

// GetDisplayName 获取数据库显示名称
func (s *SQLite3) GetDisplayName() string {
	return "SQLite3"
}

// GetTables 获取所有表名
func (s *SQLite3) GetTables() ([]string, error) {
	rows, err := s.db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return nil, fmt.Errorf("failed to query table list: %w", err)
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
func (s *SQLite3) GetTableSchema(tableName string) (string, error) {
	rows, err := s.db.Query(fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' AND name='%s'", tableName))
	if err != nil {
		return "", fmt.Errorf("failed to query table schema: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", fmt.Errorf("table %s does not exist", tableName)
	}

	var createTable string
	if err := rows.Scan(&createTable); err != nil {
		return "", err
	}

	return createTable, nil
}

// GetTableColumns 获取表的列信息
func (s *SQLite3) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info(`%s`)", tableName)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query column information: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var cid int
		var notnull, pk int
		var defaultVal sql.NullString

		if err := rows.Scan(&cid, &col.Name, &col.Type, &notnull, &defaultVal, &pk); err != nil {
			return nil, err
		}

		col.Nullable = (notnull == 0)
		if pk == 1 {
			col.Key = "PRI"
		}
		if defaultVal.Valid {
			col.DefaultValue = defaultVal.String
		}

		columns = append(columns, col)
	}
	return columns, rows.Err()
}

// ExecuteQuery 执行查询
func (s *SQLite3) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
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
func (s *SQLite3) ExecuteUpdate(query string) (int64, error) {
	result, err := s.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute update: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteDelete 执行删除
func (s *SQLite3) ExecuteDelete(query string) (int64, error) {
	result, err := s.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute delete: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteInsert 执行插入
func (s *SQLite3) ExecuteInsert(query string) (int64, error) {
	result, err := s.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute insert: %w", err)
	}
	return result.RowsAffected()
}

// GetTableData 获取表数据（分页）
func (s *SQLite3) GetTableData(tableName string, page, pageSize int, filters *FilterGroup) ([]map[string]interface{}, int64, error) {
	// 构建 WHERE 子句
	whereClause, whereArgs, err := BuildWhereClause("sqlite", tableName, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build where clause: %w", err)
	}

	// 获取总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	if whereClause != "" {
		countQuery += " WHERE " + whereClause
	}
	if err := s.db.QueryRow(countQuery, whereArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to query total count: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM `%s`", tableName)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)

	rows, err := s.db.Query(query, whereArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query data: %w", err)
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
func (s *SQLite3) GetTableDataByID(tableName string, primaryKey string, lastId interface{}, pageSize int, direction string, filters *FilterGroup) ([]map[string]interface{}, int64, interface{}, error) {
	// 构建 WHERE 子句
	whereClause, whereArgs, err := BuildWhereClause("sqlite", tableName, filters)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to build where clause: %w", err)
	}

	// 获取总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	if whereClause != "" {
		countQuery += " WHERE " + whereClause
	}
	if err := s.db.QueryRow(countQuery, whereArgs...).Scan(&total); err != nil {
		return nil, 0, nil, fmt.Errorf("failed to query total count: %w", err)
	}

	// 构建基于ID的查询
	var query string
	var rows *sql.Rows
	var queryArgs []interface{}
	
	// 合并过滤条件和ID条件
	idCondition := ""
	if direction == "prev" {
		// 上一页：使用 WHERE id < lastId ORDER BY id DESC，然后反转结果
		if lastId == nil {
			return nil, 0, nil, fmt.Errorf("lastId is required for previous page")
		}
		idCondition = fmt.Sprintf("`%s` < ?", primaryKey)
		queryArgs = append(whereArgs, lastId)
	} else {
		// 下一页或第一页
		if lastId == nil {
			// 第一页：不需要ID条件
			idCondition = ""
			queryArgs = whereArgs
		} else {
			// 后续页：使用 WHERE id > lastId
			idCondition = fmt.Sprintf("`%s` > ?", primaryKey)
			queryArgs = append(whereArgs, lastId)
		}
	}

	// 合并所有WHERE条件
	allConditions := []string{}
	if whereClause != "" {
		allConditions = append(allConditions, whereClause)
	}
	if idCondition != "" {
		allConditions = append(allConditions, idCondition)
	}

	query = fmt.Sprintf("SELECT * FROM `%s`", tableName)
	if len(allConditions) > 0 {
		query += " WHERE " + strings.Join(allConditions, " AND ")
	}
	
	if direction == "prev" {
		query += fmt.Sprintf(" ORDER BY `%s` DESC LIMIT %d", primaryKey, pageSize)
	} else {
		query += fmt.Sprintf(" ORDER BY `%s` ASC LIMIT %d", primaryKey, pageSize)
	}
	
	rows, err = s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to query data: %w", err)
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
		
		// 记录第一个和最后一个ID
		if idVal, ok := row[primaryKey]; ok {
			if firstId == nil {
				firstId = idVal
			}
			nextId = idVal
		}
	}
	
	// 如果是上一页，需要反转结果（因为查询时用了DESC）
	if direction == "prev" {
		// 反转结果数组
		for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
			results[i], results[j] = results[j], results[i]
		}
		// 上一页的nextId应该是当前页的第一个ID（用于继续向前翻页）
		nextId = firstId
	}

	return results, total, nextId, rows.Err()
}

// GetPageIdByPageNumber 根据页码计算该页的起始ID（用于页码跳转）
func (s *SQLite3) GetPageIdByPageNumber(tableName string, primaryKey string, page, pageSize int) (interface{}, error) {
	if page <= 1 {
		return nil, nil // 第一页没有lastId
	}
	
	// 计算需要跳过的记录数
	offset := (page - 1) * pageSize - 1 // 减1是因为我们要获取上一页的最后一个ID
	
	// 查询第offset条记录的ID
	query := fmt.Sprintf("SELECT `%s` FROM `%s` ORDER BY `%s` ASC LIMIT 1 OFFSET %d", primaryKey, tableName, primaryKey, offset)
	
	var id interface{}
	err := s.db.QueryRow(query).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果查询不到，说明页码超出范围，返回nil
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query page ID: %w", err)
	}
	
	return id, nil
}

// GetDatabases SQLite3不支持多数据库，返回空列表
func (s *SQLite3) GetDatabases() ([]string, error) {
	return []string{}, nil
}

// SwitchDatabase SQLite3不支持切换数据库
func (s *SQLite3) SwitchDatabase(databaseName string) error {
	return fmt.Errorf("SQLite3 does not support switching databases")
}

// BuildSQLite3DSN 根据连接信息构建SQLite3 DSN
func BuildSQLite3DSN(info ConnectionInfo) string {
	if info.DSN != "" {
		return info.DSN
	}

	// SQLite3 DSN格式: file:path/to/database.db
	// 如果提供了Database字段，使用它作为文件路径
	// 否则使用Host字段作为文件路径
	var dsn string
	if info.Database != "" {
		dsn = info.Database
	} else if info.Host != "" {
		dsn = info.Host
	} else {
		dsn = ":memory:" // 默认使用内存数据库
	}

	// SQLite3支持查询参数
	params := []string{}
	if info.User != "" {
		params = append(params, fmt.Sprintf("_auth_user=%s", info.User))
	}
	if info.Password != "" {
		params = append(params, fmt.Sprintf("_auth_pass=%s", info.Password))
	}

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}
