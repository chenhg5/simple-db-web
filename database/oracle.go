package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/sijms/go-ora/v2"
)

// Oracle 实现Database接口
type Oracle struct {
	db *sql.DB
}

// NewOracle 创建Oracle实例
func NewOracle() *Oracle {
	return &Oracle{}
}

// Connect 建立Oracle连接
func (o *Oracle) Connect(dsn string) error {
	db, err := sql.Open("oracle", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	o.db = db
	return nil
}

// Close 关闭连接
func (o *Oracle) Close() error {
	if o.db != nil {
		return o.db.Close()
	}
	return nil
}

// GetTables 获取所有表名
func (o *Oracle) GetTables() ([]string, error) {
	query := `SELECT table_name FROM user_tables ORDER BY table_name`
	rows, err := o.db.Query(query)
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
func (o *Oracle) GetTableSchema(tableName string) (string, error) {
	query := fmt.Sprintf(`
		SELECT DBMS_METADATA.GET_DDL('TABLE', '%s') FROM DUAL
	`, strings.ToUpper(tableName))

	rows, err := o.db.Query(query)
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
func (o *Oracle) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf(`
		SELECT 
			column_name,
			data_type || 
			CASE 
				WHEN data_precision IS NOT NULL AND data_scale IS NOT NULL THEN '(' || data_precision || ',' || data_scale || ')'
				WHEN data_length IS NOT NULL AND data_type IN ('CHAR', 'VARCHAR2', 'NCHAR', 'NVARCHAR2') THEN '(' || data_length || ')'
				ELSE ''
			END as data_type,
			nullable,
			data_default,
			CASE 
				WHEN constraint_type = 'P' THEN 'PRI'
				ELSE ''
			END as key_type
		FROM user_tab_columns
		LEFT JOIN (
			SELECT cols.column_name, cons.constraint_type
			FROM user_constraints cons
			INNER JOIN user_cons_columns cols ON cons.constraint_name = cols.constraint_name
			WHERE cons.table_name = UPPER('%s') AND cons.constraint_type = 'P'
		) pk ON user_tab_columns.column_name = pk.column_name
		WHERE table_name = UPPER('%s')
		ORDER BY column_id
	`, tableName, tableName)

	rows, err := o.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询列信息失败: %w", err)
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

		col.Nullable = (nullable == "Y")
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
func (o *Oracle) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	rows, err := o.db.Query(query)
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
func (o *Oracle) ExecuteUpdate(query string) (int64, error) {
	result, err := o.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行更新失败: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteDelete 执行删除
func (o *Oracle) ExecuteDelete(query string) (int64, error) {
	result, err := o.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行删除失败: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteInsert 执行插入
func (o *Oracle) ExecuteInsert(query string) (int64, error) {
	result, err := o.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行插入失败: %w", err)
	}
	return result.RowsAffected()
}

// GetTableData 获取表数据（分页）
func (o *Oracle) GetTableData(tableName string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	// 获取总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", strings.ToUpper(tableName))
	if err := o.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 获取分页数据（Oracle 12c+ 使用 FETCH FIRST/OFFSET，旧版本使用 ROWNUM）
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT * FROM "%s"
		OFFSET %d ROWS FETCH NEXT %d ROWS ONLY
	`, strings.ToUpper(tableName), offset, pageSize)

	rows, err := o.db.Query(query)
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
			// 跳过 ROWNUM 列
			if col == "RNUM" {
				continue
			}
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
func (o *Oracle) GetTableDataByID(tableName string, primaryKey string, lastId interface{}, pageSize int, direction string) ([]map[string]interface{}, int64, interface{}, error) {
	// 获取总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", strings.ToUpper(tableName))
	if err := o.db.QueryRow(countQuery).Scan(&total); err != nil {
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
				SELECT * FROM "%s" WHERE "%s" < ? ORDER BY "%s" DESC
			) WHERE ROWNUM <= %d ORDER BY "%s" ASC
		`, strings.ToUpper(tableName), strings.ToUpper(primaryKey), strings.ToUpper(primaryKey), pageSize, strings.ToUpper(primaryKey))
		rows, err = o.db.Query(query, lastId)
	} else {
		if lastId == nil {
			query = fmt.Sprintf(`
				SELECT * FROM "%s" ORDER BY "%s" ASC FETCH FIRST %d ROWS ONLY
			`, strings.ToUpper(tableName), strings.ToUpper(primaryKey), pageSize)
			rows, err = o.db.Query(query)
		} else {
			query = fmt.Sprintf(`
				SELECT * FROM "%s" WHERE "%s" > ? ORDER BY "%s" ASC FETCH FIRST %d ROWS ONLY
			`, strings.ToUpper(tableName), strings.ToUpper(primaryKey), strings.ToUpper(primaryKey), pageSize)
			rows, err = o.db.Query(query, lastId)
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
func (o *Oracle) GetPageIdByPageNumber(tableName string, primaryKey string, page, pageSize int) (interface{}, error) {
	if page <= 1 {
		return nil, nil
	}

	offset := (page - 1) * pageSize - 1
	query := fmt.Sprintf(`
		SELECT "%s" FROM "%s"
		ORDER BY "%s" ASC
		OFFSET %d ROWS FETCH NEXT 1 ROWS ONLY
	`, strings.ToUpper(primaryKey), strings.ToUpper(tableName), strings.ToUpper(primaryKey), offset)

	var id interface{}
	err := o.db.QueryRow(query).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询页码ID失败: %w", err)
	}

	return id, nil
}

// GetDatabases 获取所有数据库名称（Oracle使用schema概念）
func (o *Oracle) GetDatabases() ([]string, error) {
	// Oracle 使用 schema 概念，这里返回当前用户可访问的 schema
	query := `SELECT username FROM all_users WHERE username NOT IN ('SYS', 'SYSTEM') ORDER BY username`
	rows, err := o.db.Query(query)
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
		databases = append(databases, dbName)
	}
	return databases, rows.Err()
}

// SwitchDatabase 切换当前使用的数据库（Oracle使用schema）
func (o *Oracle) SwitchDatabase(databaseName string) error {
	// Oracle 切换 schema 需要重新连接
	// 这里返回错误，提示需要重新连接
	return fmt.Errorf("Oracle不支持动态切换schema，请重新连接")
}

// BuildOracleDSN 根据连接信息构建Oracle DSN
func BuildOracleDSN(info ConnectionInfo) string {
	if info.DSN != "" {
		return info.DSN
	}

	// Oracle DSN格式: oracle://user:password@host:port/service_name
	// 或者: oracle://user:password@host:port/sid
	var dsn string
	if info.Database != "" {
		dsn = fmt.Sprintf("oracle://%s:%s@%s:%s/%s",
			info.User,
			info.Password,
			info.Host,
			info.Port,
			info.Database,
		)
	} else {
		dsn = fmt.Sprintf("oracle://%s:%s@%s:%s",
			info.User,
			info.Password,
			info.Host,
			info.Port,
		)
	}

	return dsn
}

