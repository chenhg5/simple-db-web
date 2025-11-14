package database

import (
	"database/sql"
	"fmt"
	"strings"

	// DB2 驱动需要系统级别的 DB2 客户端库和 CGO
	// 如果编译失败，请安装 DB2 客户端库或使用其他 DB2 驱动
	// _ "github.com/ibmdb/go_ibm_db"
)

// DB2 实现Database接口
type DB2 struct {
	db *sql.DB
}

// NewDB2 创建DB2实例
func NewDB2() *DB2 {
	return &DB2{}
}

// Connect 建立DB2连接
func (d *DB2) Connect(dsn string) error {
	// DB2 驱动需要系统级别的 DB2 客户端库
	// 当前实现需要安装 DB2 客户端库和配置 CGO
	// 如果编译失败，请安装 DB2 客户端库或使用其他 DB2 驱动
	return fmt.Errorf("DB2 驱动需要系统级别的 DB2 客户端库，请安装 DB2 客户端库后重新编译")
	
	// 取消注释以下代码以使用 go_ibm_db 驱动（需要安装 DB2 客户端库）
	/*
	db, err := sql.Open("go_ibm_db", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	d.db = db
	return nil
	*/
}

// Close 关闭连接
func (d *DB2) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// GetTables 获取所有表名
func (d *DB2) GetTables() ([]string, error) {
	query := `SELECT TABNAME FROM SYSCAT.TABLES WHERE TABSCHEMA = CURRENT SCHEMA AND TYPE = 'T' ORDER BY TABNAME`
	rows, err := d.db.Query(query)
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
func (d *DB2) GetTableSchema(tableName string) (string, error) {
	query := fmt.Sprintf(`
		SELECT TEXT FROM SYSCAT.VIEWS 
		WHERE VIEWNAME = 'CREATE_%s' AND VIEWSCHEMA = CURRENT SCHEMA
		UNION ALL
		SELECT 'CREATE TABLE ' || TABSCHEMA || '.' || TABNAME || ' (' ||
			LISTAGG(COLNAME || ' ' || TYPENAME || 
				CASE 
					WHEN LENGTH > 0 THEN '(' || LENGTH || ')'
					ELSE ''
				END ||
				CASE 
					WHEN NULLS = 'N' THEN ' NOT NULL'
					ELSE ''
				END, ', ') WITHIN GROUP (ORDER BY COLNO) ||
			');' as TEXT
		FROM SYSCAT.COLUMNS
		WHERE TABNAME = '%s' AND TABSCHEMA = CURRENT SCHEMA
		GROUP BY TABSCHEMA, TABNAME
	`, strings.ToUpper(tableName), strings.ToUpper(tableName))

	rows, err := d.db.Query(query)
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
func (d *DB2) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf(`
		SELECT 
			COLNAME,
			TYPENAME || 
			CASE 
				WHEN LENGTH > 0 THEN '(' || LENGTH || ')'
				ELSE ''
			END as TYPENAME,
			NULLS,
			DEFAULT,
			CASE 
				WHEN KEYSEQ IS NOT NULL THEN 'PRI'
				ELSE ''
			END as KEY_TYPE
		FROM SYSCAT.COLUMNS
		LEFT JOIN SYSCAT.KEYCOLUSE ON 
			SYSCAT.COLUMNS.TABSCHEMA = SYSCAT.KEYCOLUSE.TABSCHEMA AND
			SYSCAT.COLUMNS.TABNAME = SYSCAT.KEYCOLUSE.TABNAME AND
			SYSCAT.COLUMNS.COLNAME = SYSCAT.KEYCOLUSE.COLNAME
		WHERE SYSCAT.COLUMNS.TABNAME = '%s' AND SYSCAT.COLUMNS.TABSCHEMA = CURRENT SCHEMA
		ORDER BY COLNO
	`, strings.ToUpper(tableName))

	rows, err := d.db.Query(query)
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

		col.Nullable = (nullable == "Y")
		if defaultVal.Valid {
			col.DefaultValue = defaultVal.String
		}

		columns = append(columns, col)
	}
	return columns, rows.Err()
}

// ExecuteQuery 执行查询
func (d *DB2) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	rows, err := d.db.Query(query)
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
func (d *DB2) ExecuteUpdate(query string) (int64, error) {
	result, err := d.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行更新失败: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteDelete 执行删除
func (d *DB2) ExecuteDelete(query string) (int64, error) {
	result, err := d.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行删除失败: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteInsert 执行插入
func (d *DB2) ExecuteInsert(query string) (int64, error) {
	result, err := d.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行插入失败: %w", err)
	}
	return result.RowsAffected()
}

// GetTableData 获取表数据（分页）
func (d *DB2) GetTableData(tableName string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	// 获取总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", strings.ToUpper(tableName))
	if err := d.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT * FROM %s
		ORDER BY (SELECT NULL)
		FETCH FIRST %d ROWS ONLY
		OFFSET %d ROWS
	`, strings.ToUpper(tableName), pageSize, offset)

	rows, err := d.db.Query(query)
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
func (d *DB2) GetTableDataByID(tableName string, primaryKey string, lastId interface{}, pageSize int, direction string) ([]map[string]interface{}, int64, interface{}, error) {
	// 获取总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", strings.ToUpper(tableName))
	if err := d.db.QueryRow(countQuery).Scan(&total); err != nil {
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
				SELECT * FROM %s WHERE %s < ? ORDER BY %s DESC FETCH FIRST %d ROWS ONLY
			) ORDER BY %s ASC
		`, strings.ToUpper(tableName), strings.ToUpper(primaryKey), strings.ToUpper(primaryKey), pageSize, strings.ToUpper(primaryKey))
		rows, err = d.db.Query(query, lastId)
	} else {
		if lastId == nil {
			query = fmt.Sprintf(`
				SELECT * FROM %s ORDER BY %s ASC FETCH FIRST %d ROWS ONLY
			`, strings.ToUpper(tableName), strings.ToUpper(primaryKey), pageSize)
			rows, err = d.db.Query(query)
		} else {
			query = fmt.Sprintf(`
				SELECT * FROM %s WHERE %s > ? ORDER BY %s ASC FETCH FIRST %d ROWS ONLY
			`, strings.ToUpper(tableName), strings.ToUpper(primaryKey), strings.ToUpper(primaryKey), pageSize)
			rows, err = d.db.Query(query, lastId)
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
func (d *DB2) GetPageIdByPageNumber(tableName string, primaryKey string, page, pageSize int) (interface{}, error) {
	if page <= 1 {
		return nil, nil
	}

	offset := (page - 1) * pageSize - 1
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		ORDER BY %s ASC
		FETCH FIRST 1 ROWS ONLY
		OFFSET %d ROWS
	`, strings.ToUpper(primaryKey), strings.ToUpper(tableName), strings.ToUpper(primaryKey), offset)

	var id interface{}
	err := d.db.QueryRow(query).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询页码ID失败: %w", err)
	}

	return id, nil
}

// GetDatabases 获取所有数据库名称
func (d *DB2) GetDatabases() ([]string, error) {
	// DB2 使用 schema 概念，这里返回当前用户可访问的 schema
	query := `SELECT SCHEMANAME FROM SYSCAT.SCHEMATA WHERE SCHEMANAME NOT LIKE 'SYS%' ORDER BY SCHEMANAME`
	rows, err := d.db.Query(query)
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

// SwitchDatabase 切换当前使用的数据库（DB2使用schema）
func (d *DB2) SwitchDatabase(databaseName string) error {
	_, err := d.db.Exec(fmt.Sprintf("SET SCHEMA %s", strings.ToUpper(databaseName)))
	return err
}

// BuildDB2DSN 根据连接信息构建DB2 DSN
func BuildDB2DSN(info ConnectionInfo) string {
	if info.DSN != "" {
		return info.DSN
	}

	// DB2 DSN格式: HOSTNAME=host;DATABASE=db;PORT=port;UID=user;PWD=password
	var dsn string
	if info.Database != "" {
		dsn = fmt.Sprintf("HOSTNAME=%s;DATABASE=%s;PORT=%s;UID=%s;PWD=%s",
			info.Host,
			info.Database,
			info.Port,
			info.User,
			info.Password,
		)
	} else {
		dsn = fmt.Sprintf("HOSTNAME=%s;PORT=%s;UID=%s;PWD=%s",
			info.Host,
			info.Port,
			info.User,
			info.Password,
		)
	}

	return dsn
}

