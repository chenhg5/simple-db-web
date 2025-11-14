package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

// PostgreSQL 实现Database接口
type PostgreSQL struct {
	db *sql.DB
}

// NewPostgreSQL 创建PostgreSQL实例
func NewPostgreSQL() *PostgreSQL {
	return &PostgreSQL{}
}

// Connect 建立PostgreSQL连接
func (p *PostgreSQL) Connect(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	p.db = db
	return nil
}

// Close 关闭连接
func (p *PostgreSQL) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// GetTypeName 获取数据库类型名称
func (p *PostgreSQL) GetTypeName() string {
	return "postgresql"
}

// GetDisplayName 获取数据库显示名称
func (p *PostgreSQL) GetDisplayName() string {
	return "PostgreSQL"
}

// GetTables 获取所有表名
func (p *PostgreSQL) GetTables() ([]string, error) {
	query := `SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename`
	rows, err := p.db.Query(query)
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
func (p *PostgreSQL) GetTableSchema(tableName string) (string, error) {
	query := fmt.Sprintf(`
		SELECT 
			'CREATE TABLE ' || quote_ident('%s') || ' (' || 
			string_agg(
				quote_ident(column_name) || ' ' || 
				udt_name || 
				CASE 
					WHEN character_maximum_length IS NOT NULL THEN '(' || character_maximum_length || ')'
					ELSE ''
				END ||
				CASE 
					WHEN is_nullable = 'NO' THEN ' NOT NULL'
					ELSE ''
				END ||
				CASE 
					WHEN column_default IS NOT NULL THEN ' DEFAULT ' || column_default
					ELSE ''
				END,
				', '
			) || ');' as create_table
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = '%s'
		GROUP BY table_name;
	`, tableName, tableName)

	rows, err := p.db.Query(query)
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
func (p *PostgreSQL) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf(`
		SELECT 
			column_name,
			udt_name as data_type,
			is_nullable,
			column_default,
			CASE 
				WHEN constraint_type = 'PRIMARY KEY' THEN 'PRI'
				WHEN constraint_type = 'UNIQUE' THEN 'UNI'
				ELSE ''
			END as key_type
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT 
				kcu.column_name,
				tc.constraint_type
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu 
				ON tc.constraint_name = kcu.constraint_name
			WHERE tc.table_schema = 'public' 
				AND tc.table_name = '%s'
		) k ON c.column_name = k.column_name
		WHERE c.table_schema = 'public' AND c.table_name = '%s'
		ORDER BY c.ordinal_position;
	`, tableName, tableName)

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query column information: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var nullable string
		var defaultVal sql.NullString
		var keyType sql.NullString

		if err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultVal, &keyType); err != nil {
			return nil, err
		}

		col.Nullable = (nullable == "YES")
		if keyType.Valid {
			col.Key = keyType.String
		}
		if defaultVal.Valid {
			col.DefaultValue = defaultVal.String
		}

		columns = append(columns, col)
	}
	return columns, rows.Err()
}

// ExecuteQuery 执行查询
func (p *PostgreSQL) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	rows, err := p.db.Query(query)
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
func (p *PostgreSQL) ExecuteUpdate(query string) (int64, error) {
	result, err := p.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute update: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteDelete 执行删除
func (p *PostgreSQL) ExecuteDelete(query string) (int64, error) {
	result, err := p.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute delete: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteInsert 执行插入
func (p *PostgreSQL) ExecuteInsert(query string) (int64, error) {
	result, err := p.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute insert: %w", err)
	}
	return result.RowsAffected()
}

// GetTableData 获取表数据（分页）
func (p *PostgreSQL) GetTableData(tableName string, page, pageSize int, filters *FilterGroup) ([]map[string]interface{}, int64, error) {
	// 构建 WHERE 子句
	whereClause, whereArgs, err := BuildWhereClause("postgresql", tableName, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build where clause: %w", err)
	}

	// 获取总数
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, tableName)
	if whereClause != "" {
		countQuery += " WHERE " + whereClause
	}
	if err := p.db.QueryRow(countQuery, whereArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to query total count: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT * FROM "%s"`, tableName)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	query += fmt.Sprintf(` LIMIT %d OFFSET %d`, pageSize, offset)

	rows, err := p.db.Query(query, whereArgs...)
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
func (p *PostgreSQL) GetTableDataByID(tableName string, primaryKey string, lastId interface{}, pageSize int, direction string, filters *FilterGroup) ([]map[string]interface{}, int64, interface{}, error) {
	// 构建 WHERE 子句
	whereClause, whereArgs, err := BuildWhereClause("postgresql", tableName, filters)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to build where clause: %w", err)
	}

	// 获取总数
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, tableName)
	if whereClause != "" {
		countQuery += " WHERE " + whereClause
	}
	if err := p.db.QueryRow(countQuery, whereArgs...).Scan(&total); err != nil {
		return nil, 0, nil, fmt.Errorf("failed to query total count: %w", err)
	}

	// 构建基于ID的查询
	var query string
	var rows *sql.Rows
	var queryArgs []interface{}
	
	// 合并过滤条件和ID条件
	idCondition := ""
	argIndex := len(whereArgs) + 1 // PostgreSQL 占位符从 $1 开始
	
	if direction == "prev" {
		// 上一页：使用 WHERE id < lastId ORDER BY id DESC，然后反转结果
		if lastId == nil {
			return nil, 0, nil, fmt.Errorf("lastId is required for previous page")
		}
		idCondition = fmt.Sprintf(`"%s" < $%d`, primaryKey, argIndex)
		queryArgs = append(whereArgs, lastId)
	} else {
		// 下一页或第一页
		if lastId == nil {
			// 第一页：不需要ID条件
			idCondition = ""
			queryArgs = whereArgs
		} else {
			// 后续页：使用 WHERE id > lastId
			idCondition = fmt.Sprintf(`"%s" > $%d`, primaryKey, argIndex)
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

	query = fmt.Sprintf(`SELECT * FROM "%s"`, tableName)
	if len(allConditions) > 0 {
		query += " WHERE " + strings.Join(allConditions, " AND ")
	}
	
	if direction == "prev" {
		query += fmt.Sprintf(` ORDER BY "%s" DESC LIMIT %d`, primaryKey, pageSize)
	} else {
		query += fmt.Sprintf(` ORDER BY "%s" ASC LIMIT %d`, primaryKey, pageSize)
	}
	
	rows, err = p.db.Query(query, queryArgs...)
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
func (p *PostgreSQL) GetPageIdByPageNumber(tableName string, primaryKey string, page, pageSize int) (interface{}, error) {
	if page <= 1 {
		return nil, nil // 第一页没有lastId
	}
	
	// 计算需要跳过的记录数
	offset := (page - 1) * pageSize - 1 // 减1是因为我们要获取上一页的最后一个ID
	
	// 查询第offset条记录的ID
	query := fmt.Sprintf(`SELECT "%s" FROM "%s" ORDER BY "%s" ASC LIMIT 1 OFFSET %d`, primaryKey, tableName, primaryKey, offset)
	
	var id interface{}
	err := p.db.QueryRow(query).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果查询不到，说明页码超出范围，返回nil
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query page ID: %w", err)
	}
	
	return id, nil
}

// GetDatabases 获取所有数据库名称
func (p *PostgreSQL) GetDatabases() ([]string, error) {
	query := `SELECT datname FROM pg_database WHERE datistemplate = false AND datname != 'postgres' ORDER BY datname`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query database list: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		databases = append(databases, dbName)
	}
	return databases, rows.Err()
}

// SwitchDatabase 切换当前使用的数据库
func (p *PostgreSQL) SwitchDatabase(databaseName string) error {
	// PostgreSQL需要重新连接才能切换数据库
	// 这里我们关闭当前连接，调用者需要重新连接
	return fmt.Errorf("PostgreSQL requires reconnection to switch database, please reconnect with new connection info")
}

// BuildPostgreSQLDSN 根据连接信息构建PostgreSQL DSN
func BuildPostgreSQLDSN(info ConnectionInfo) string {
	if info.DSN != "" {
		return info.DSN
	}

	// PostgreSQL DSN格式: postgres://user:password@host:port/database?sslmode=disable
	// 或者使用 lib/pq 格式: host=host port=port user=user password=password dbname=database sslmode=disable
	var dsn string
	if info.Database != "" {
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			info.Host,
			info.Port,
			info.User,
			info.Password,
			info.Database,
		)
	} else {
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable",
			info.Host,
			info.Port,
			info.User,
			info.Password,
		)
	}

	return dsn
}
