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

	dsn      string
	dbConfig *DBConfig
}

// NewMySQL 创建MySQL实例
func NewMySQL() *MySQL {
	return &MySQL{}
}

// Connect 建立MySQL连接
func (m *MySQL) Connect(dsn string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	oldDB := m.db
	m.db = db
	if oldDB != nil {
		oldDB.Close()
	}
	m.dsn = dsn
	dbConfig, err := GetDBConfigFromDSN(dsn)
	if err != nil {
		return err
	}
	m.dbConfig = dbConfig
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
func (m *MySQL) GetTableSchema(tableName string) (string, error) {
	rows, err := m.db.Query(fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`", m.dbConfig.Database, tableName))
	if err != nil {
		return "", fmt.Errorf("failed to query table schema: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", fmt.Errorf("table %s does not exist", tableName)
	}

	var table, createTable string
	if err := rows.Scan(&table, &createTable); err != nil {
		return "", err
	}

	return createTable, nil
}

// GetTableColumns 获取表的列信息
func (m *MySQL) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf("DESCRIBE `%s`.`%s`", m.dbConfig.Database, tableName)
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query column information: %w", err)
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
func (m *MySQL) ExecuteUpdate(query string) (int64, error) {
	result, err := m.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute update: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteDelete 执行删除
func (m *MySQL) ExecuteDelete(query string) (int64, error) {
	result, err := m.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute delete: %w", err)
	}
	return result.RowsAffected()
}

// ExecuteInsert 执行插入
func (m *MySQL) ExecuteInsert(query string) (int64, error) {
	result, err := m.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute insert: %w", err)
	}
	return result.RowsAffected()
}

// GetTableData 获取表数据（分页）
func (m *MySQL) GetTableData(tableName string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	// 获取总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	if err := m.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to query total count: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT %d OFFSET %d", tableName, pageSize, offset)

	rows, err := m.db.Query(query)
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
func (m *MySQL) GetTableDataByID(tableName string, primaryKey string, lastId interface{}, pageSize int, direction string) ([]map[string]interface{}, int64, interface{}, error) {
	// 获取总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	if err := m.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, nil, fmt.Errorf("查询总数失败: %w", err)
	}

	// 构建基于ID的查询
	var query string
	var rows *sql.Rows
	var err error
	
	if direction == "prev" {
		// 上一页：使用 WHERE id < lastId ORDER BY id DESC，然后反转结果
		if lastId == nil {
			return nil, 0, nil, fmt.Errorf("lastId is required for previous page")
		}
		query = fmt.Sprintf("SELECT * FROM `%s` WHERE `%s` < ? ORDER BY `%s` DESC LIMIT %d", tableName, primaryKey, primaryKey, pageSize)
		rows, err = m.db.Query(query, lastId)
	} else {
		// 下一页或第一页
		if lastId == nil {
			// 第一页：直接按ID排序取前pageSize条
			query = fmt.Sprintf("SELECT * FROM `%s` ORDER BY `%s` ASC LIMIT %d", tableName, primaryKey, pageSize)
			rows, err = m.db.Query(query)
		} else {
			// 后续页：使用 WHERE id > lastId
			query = fmt.Sprintf("SELECT * FROM `%s` WHERE `%s` > ? ORDER BY `%s` ASC LIMIT %d", tableName, primaryKey, primaryKey, pageSize)
			rows, err = m.db.Query(query, lastId)
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
func (m *MySQL) GetPageIdByPageNumber(tableName string, primaryKey string, page, pageSize int) (interface{}, error) {
	if page <= 1 {
		return nil, nil // 第一页没有lastId
	}
	
	// 计算需要跳过的记录数
	offset := (page - 1) * pageSize - 1 // 减1是因为我们要获取上一页的最后一个ID
	
	// 查询第offset条记录的ID
	query := fmt.Sprintf("SELECT `%s` FROM `%s` ORDER BY `%s` ASC LIMIT 1 OFFSET %d", primaryKey, tableName, primaryKey, offset)
	
	var id interface{}
	err := m.db.QueryRow(query).Scan(&id)
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
func (m *MySQL) GetDatabases() ([]string, error) {
	rows, err := m.db.Query("SHOW DATABASES")
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
	dbConfig, err := GetDBConfigFromDSN(m.dsn)
	if err != nil {
		return err
	}
	dbConfig.Database = databaseName
	m.dsn = dbConfig.BuildDSN()
	return m.Connect(m.dsn)
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
