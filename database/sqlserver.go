package database

import (
	"database/sql"
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
)

// SQLServer 实现Database接口
type SQLServer struct {
	db *sql.DB
}

// NewSQLServer 创建SQLServer实例
func NewSQLServer() *SQLServer {
	return &SQLServer{}
}

// Connect 建立SQLServer连接
func (s *SQLServer) Connect(dsn string) error {
	db, err := sql.Open("sqlserver", dsn)
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
func (s *SQLServer) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetTypeName 获取数据库类型名称
func (s *SQLServer) GetTypeName() string {
	return "sqlserver"
}

// GetDisplayName 获取数据库显示名称
func (s *SQLServer) GetDisplayName() string {
	return "SQL Server"
}

// GetTables 获取所有表名
func (s *SQLServer) GetTables() ([]string, error) {
	query := `SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE' ORDER BY TABLE_NAME`
	rows, err := s.db.Query(query)
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
func (s *SQLServer) GetTableSchema(tableName string) (string, error) {
	query := fmt.Sprintf(`
		SELECT 
			'CREATE TABLE [' + TABLE_SCHEMA + '].[' + TABLE_NAME + '] (' +
			STUFF((
				SELECT ', [' + COLUMN_NAME + '] ' + 
				DATA_TYPE + 
				CASE 
					WHEN CHARACTER_MAXIMUM_LENGTH IS NOT NULL THEN '(' + CAST(CHARACTER_MAXIMUM_LENGTH AS VARCHAR) + ')'
					WHEN NUMERIC_PRECISION IS NOT NULL AND NUMERIC_SCALE IS NOT NULL THEN '(' + CAST(NUMERIC_PRECISION AS VARCHAR) + ',' + CAST(NUMERIC_SCALE AS VARCHAR) + ')'
					ELSE ''
				END +
				CASE 
					WHEN IS_NULLABLE = 'NO' THEN ' NOT NULL'
					ELSE ''
				END +
				CASE 
					WHEN COLUMN_DEFAULT IS NOT NULL THEN ' DEFAULT ' + COLUMN_DEFAULT
					ELSE ''
				END
				FROM INFORMATION_SCHEMA.COLUMNS
				WHERE TABLE_NAME = '%s' AND TABLE_SCHEMA = (SELECT TABLE_SCHEMA FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = '%s')
				ORDER BY ORDINAL_POSITION
				FOR XML PATH('')
			), 1, 1, '') + ');' as create_table
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_NAME = '%s'
	`, tableName, tableName, tableName)

	rows, err := s.db.Query(query)
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
func (s *SQLServer) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf(`
		SELECT 
			COLUMN_NAME,
			DATA_TYPE + 
			CASE 
				WHEN CHARACTER_MAXIMUM_LENGTH IS NOT NULL THEN '(' + CAST(CHARACTER_MAXIMUM_LENGTH AS VARCHAR) + ')'
				WHEN NUMERIC_PRECISION IS NOT NULL AND NUMERIC_SCALE IS NOT NULL THEN '(' + CAST(NUMERIC_PRECISION AS VARCHAR) + ',' + CAST(NUMERIC_SCALE AS VARCHAR) + ')'
				ELSE ''
			END as DATA_TYPE,
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
		WHERE TABLE_NAME = '%s'
		ORDER BY ORDINAL_POSITION
	`, tableName, tableName)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query column information: %w", err)
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
func (s *SQLServer) ExecuteQuery(query string) ([]map[string]interface{}, error) {
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
func (s *SQLServer) ExecuteUpdate(query string) (int64, error) {
	result, err := s.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute update: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteDelete 执行删除
func (s *SQLServer) ExecuteDelete(query string) (int64, error) {
	result, err := s.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute delete: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteInsert 执行插入
func (s *SQLServer) ExecuteInsert(query string) (int64, error) {
	result, err := s.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute insert: %w", err)
	}
	return result.RowsAffected()
}

// GetTableData 获取表数据（分页）
func (s *SQLServer) GetTableData(tableName string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	// 获取总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM [%s]", tableName)
	if err := s.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to query total count: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT * FROM [%s]
		ORDER BY (SELECT NULL)
		OFFSET %d ROWS
		FETCH NEXT %d ROWS ONLY
	`, tableName, offset, pageSize)

	rows, err := s.db.Query(query)
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
func (s *SQLServer) GetTableDataByID(tableName string, primaryKey string, lastId interface{}, pageSize int, direction string) ([]map[string]interface{}, int64, interface{}, error) {
	// 获取总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM [%s]", tableName)
	if err := s.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, nil, fmt.Errorf("查询总数失败: %w", err)
	}

	// 构建基于ID的查询
	var query string
	var rows *sql.Rows
	var err error

	if direction == "prev" {
		if lastId == nil {
			return nil, 0, nil, fmt.Errorf("lastId is required for previous page")
		}
		query = fmt.Sprintf(`
			SELECT * FROM (
				SELECT TOP %d * FROM [%s] WHERE [%s] < ? ORDER BY [%s] DESC
			) AS t ORDER BY [%s] ASC
		`, pageSize, tableName, primaryKey, primaryKey, primaryKey)
		rows, err = s.db.Query(query, lastId)
	} else {
		if lastId == nil {
			query = fmt.Sprintf(`
				SELECT TOP %d * FROM [%s] ORDER BY [%s] ASC
			`, pageSize, tableName, primaryKey)
			rows, err = s.db.Query(query)
		} else {
			query = fmt.Sprintf(`
				SELECT TOP %d * FROM [%s] WHERE [%s] > ? ORDER BY [%s] ASC
			`, pageSize, tableName, primaryKey, primaryKey)
			rows, err = s.db.Query(query, lastId)
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

		if idVal, ok := row[primaryKey]; ok {
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
func (s *SQLServer) GetPageIdByPageNumber(tableName string, primaryKey string, page, pageSize int) (interface{}, error) {
	if page <= 1 {
		return nil, nil
	}

	offset := (page - 1) * pageSize - 1
	query := fmt.Sprintf(`
		SELECT [%s] FROM [%s]
		ORDER BY [%s] ASC
		OFFSET %d ROWS
		FETCH NEXT 1 ROWS ONLY
	`, primaryKey, tableName, primaryKey, offset)

	var id interface{}
	err := s.db.QueryRow(query).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query page ID: %w", err)
	}

	return id, nil
}

// GetDatabases 获取所有数据库名称
func (s *SQLServer) GetDatabases() ([]string, error) {
	query := `SELECT name FROM sys.databases WHERE name NOT IN ('master', 'tempdb', 'model', 'msdb') ORDER BY name`
	rows, err := s.db.Query(query)
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
func (s *SQLServer) SwitchDatabase(databaseName string) error {
	_, err := s.db.Exec(fmt.Sprintf("USE [%s]", databaseName))
	return err
}

// BuildSQLServerDSN 根据连接信息构建SQLServer DSN
func BuildSQLServerDSN(info ConnectionInfo) string {
	if info.DSN != "" {
		return info.DSN
	}

	// SQL Server DSN格式: server=host;user id=user;password=password;port=port;database=db
	var dsn string
	if info.Database != "" {
		dsn = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s",
			info.Host,
			info.User,
			info.Password,
			info.Port,
			info.Database,
		)
	} else {
		dsn = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s",
			info.Host,
			info.User,
			info.Password,
			info.Port,
		)
	}

	// 添加加密选项（可选）
	dsn += ";encrypt=disable"

	return dsn
}

