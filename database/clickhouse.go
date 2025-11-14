package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"sync"

	_ "github.com/ClickHouse/clickhouse-go"
)

// ClickHouse 实现Database接口
type ClickHouse struct {
	db              *sql.DB
	currentDatabase string
	dbMutex         sync.RWMutex // 保护 currentDatabase 的并发访问
}

// NewClickHouse 创建ClickHouse实例
func NewClickHouse() *ClickHouse {
	return &ClickHouse{}
}

// Connect 建立ClickHouse连接
func (c *ClickHouse) Connect(dsn string) error {
	// ClickHouse v1 驱动使用 tcp:// 协议
	// 如果DSN不是以 tcp:// 或 clickhouse:// 开头，则添加 tcp:// 前缀
	if !strings.HasPrefix(dsn, "tcp://") && !strings.HasPrefix(dsn, "clickhouse://") {
		dsn = "tcp://" + dsn
	}

	// 从 DSN 中提取数据库名
	var dbName string
	if strings.HasPrefix(dsn, "tcp://") {
		// 解析 tcp://host:port?username=user&password=pass&database=db
		dsnWithoutProtocol := strings.TrimPrefix(dsn, "tcp://")
		if idx := strings.Index(dsnWithoutProtocol, "?"); idx >= 0 {
			query := dsnWithoutProtocol[idx+1:]
			if parsed, err := url.ParseQuery(query); err == nil {
				if db := parsed.Get("database"); db != "" {
					dbName = db
				}
			}
		}
	} else if strings.HasPrefix(dsn, "clickhouse://") {
		// 解析 clickhouse://user:pass@host:port/database
		if parsed, err := url.Parse(dsn); err == nil {
			if parsed.Path != "" {
				dbName = strings.TrimPrefix(parsed.Path, "/")
			}
		}
	}

	// 如果没有从 DSN 中提取到数据库名，默认为 default
	if dbName == "" {
		dbName = "default"
	}

	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	oldDB := c.db
	c.db = db
	if oldDB != nil {
		oldDB.Close()
	}

	// 存储当前数据库名
	c.dbMutex.Lock()
	c.currentDatabase = dbName
	c.dbMutex.Unlock()

	return nil
}

// Close 关闭连接
func (c *ClickHouse) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// GetTables 获取所有表名
func (c *ClickHouse) GetTables() ([]string, error) {
	rows, err := c.db.Query("SELECT name FROM system.tables WHERE database = currentDatabase()")
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
func (c *ClickHouse) GetTableSchema(tableName string) (string, error) {
	// 获取存储的当前数据库名
	c.dbMutex.RLock()
	currentDB := c.currentDatabase
	c.dbMutex.RUnlock()

	if currentDB == "" {
		return "", fmt.Errorf("当前数据库未设置")
	}

	// ClickHouse 使用 DESCRIBE TABLE 获取表结构
	// 使用 database.table 格式确保查询正确的数据库
	rows, err := c.db.Query(fmt.Sprintf("DESCRIBE TABLE `%s`.`%s`", currentDB, tableName))
	if err != nil {
		return "", fmt.Errorf("查询表结构失败: %w", err)
	}
	defer rows.Close()

	var schema strings.Builder
	schema.WriteString(fmt.Sprintf("CREATE TABLE `%s` (\n", tableName))

	first := true
	for rows.Next() {
		if !first {
			schema.WriteString(",\n")
		}
		var name, typeStr, defaultType, defaultExpr, comment, codec, ttl string
		if err := rows.Scan(&name, &typeStr, &defaultType, &defaultExpr, &comment, &codec, &ttl); err != nil {
			return "", err
		}
		schema.WriteString(fmt.Sprintf("  `%s` %s", name, typeStr))
		if defaultExpr != "" {
			schema.WriteString(fmt.Sprintf(" DEFAULT %s", defaultExpr))
		}
		if comment != "" {
			schema.WriteString(fmt.Sprintf(" COMMENT '%s'", comment))
		}
		if codec != "" {
			schema.WriteString(fmt.Sprintf(" CODEC(%s)", codec))
		}
		if ttl != "" {
			schema.WriteString(fmt.Sprintf(" TTL %s", ttl))
		}
		first = false
	}
	schema.WriteString("\n) ENGINE = MergeTree()")

	return schema.String(), rows.Err()
}

// GetTableColumns 获取表的列信息
func (c *ClickHouse) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	// 获取存储的当前数据库名
	c.dbMutex.RLock()
	currentDB := c.currentDatabase
	c.dbMutex.RUnlock()

	if currentDB == "" {
		return nil, fmt.Errorf("当前数据库未设置")
	}

	// 使用 database.table 格式确保查询正确的数据库
	query := fmt.Sprintf("DESCRIBE TABLE `%s`.`%s`", currentDB, tableName)
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询列信息失败: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var typeStr, defaultType, defaultExpr, comment, codec, ttl string

		if err := rows.Scan(&col.Name, &typeStr, &defaultType, &defaultExpr, &comment, &codec, &ttl); err != nil {
			return nil, err
		}

		col.Type = typeStr
		// ClickHouse 中 Nullable 类型包含在类型字符串中
		col.Nullable = strings.Contains(typeStr, "Nullable")
		// ClickHouse 没有主键概念（MergeTree 引擎有排序键，但不是主键）
		col.Key = ""
		if defaultExpr != "" {
			col.DefaultValue = defaultExpr
		}
		// 注意：ttl 字段暂不存储到 ColumnInfo 中，因为 ColumnInfo 结构中没有对应字段

		columns = append(columns, col)
	}
	return columns, rows.Err()
}

// ExecuteQuery 执行查询
func (c *ClickHouse) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	// 获取存储的当前数据库名，确保连接在正确的数据库上下文中
	c.dbMutex.RLock()
	currentDB := c.currentDatabase
	c.dbMutex.RUnlock()

	// 如果当前数据库已设置，在执行查询前先切换到该数据库
	// 这样可以确保即使用户的 SQL 中没有指定数据库名，也能查询到正确的数据
	if currentDB != "" {
		// 使用 Exec 执行 USE 语句，确保连接在正确的数据库上下文中
		// 注意：虽然连接池可能复用连接，但每次查询前执行 USE 可以确保正确性
		if _, err := c.db.Exec(fmt.Sprintf("USE `%s`", currentDB)); err != nil {
			return nil, fmt.Errorf("切换数据库上下文失败: %w", err)
		}
	}

	rows, err := c.db.Query(query)
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

// ExecuteUpdate 执行更新（ClickHouse 不支持 UPDATE，返回错误）
func (c *ClickHouse) ExecuteUpdate(query string) (int64, error) {
	return 0, fmt.Errorf("ClickHouse 不支持 UPDATE 操作")
}

// ExecuteDelete 执行删除（ClickHouse 不支持 DELETE，返回错误）
func (c *ClickHouse) ExecuteDelete(query string) (int64, error) {
	return 0, fmt.Errorf("ClickHouse 不支持 DELETE 操作")
}

// ExecuteInsert 执行插入
func (c *ClickHouse) ExecuteInsert(query string) (int64, error) {
	result, err := c.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("执行插入失败: %w", err)
	}
	return result.RowsAffected()
}

// GetTableData 获取表数据（ClickHouse 不支持分页，只返回10条数据）
func (c *ClickHouse) GetTableData(tableName string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	// ClickHouse 不支持分页，只返回10条数据
	// 注意：total 返回 -1 表示不支持计数

	// 获取存储的当前数据库名
	c.dbMutex.RLock()
	currentDB := c.currentDatabase
	c.dbMutex.RUnlock()

	if currentDB == "" {
		return nil, 0, fmt.Errorf("当前数据库未设置")
	}

	// 使用 database.table 格式确保查询正确的数据库
	query := fmt.Sprintf("SELECT * FROM `%s`.`%s` LIMIT 10", currentDB, tableName)

	rows, err := c.db.Query(query)
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

	// 返回 -1 表示不支持总数统计
	return results, -1, rows.Err()
}

// GetTableDataByID 基于主键ID获取表数据（ClickHouse不支持，返回错误）
func (c *ClickHouse) GetTableDataByID(tableName string, primaryKey string, lastId interface{}, pageSize int, direction string) ([]map[string]interface{}, int64, interface{}, error) {
	return nil, 0, nil, fmt.Errorf("ClickHouse 不支持基于ID的分页")
}

// GetPageIdByPageNumber 根据页码计算该页的起始ID（ClickHouse不支持，返回错误）
func (c *ClickHouse) GetPageIdByPageNumber(tableName string, primaryKey string, page, pageSize int) (interface{}, error) {
	return nil, fmt.Errorf("ClickHouse 不支持基于ID的分页")
}

// GetDatabases 获取所有数据库名称
func (c *ClickHouse) GetDatabases() ([]string, error) {
	rows, err := c.db.Query("SELECT name FROM system.databases WHERE name NOT IN ('system', 'information_schema', 'INFORMATION_SCHEMA')")
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

// SwitchDatabase 切换当前使用的数据库
func (c *ClickHouse) SwitchDatabase(databaseName string) error {
	_, err := c.db.Exec(fmt.Sprintf("USE `%s`", databaseName))
	if err != nil {
		return fmt.Errorf("切换数据库失败: %w", err)
	}

	// 更新存储的数据库名
	c.dbMutex.Lock()
	c.currentDatabase = databaseName
	c.dbMutex.Unlock()

	return nil
}

// BuildDSN 根据连接信息构建ClickHouse DSN
func BuildClickHouseDSN(info ConnectionInfo) string {
	if info.DSN != "" {
		// 如果提供了DSN，确保使用正确的协议
		if !strings.HasPrefix(info.DSN, "tcp://") && !strings.HasPrefix(info.DSN, "clickhouse://") {
			return "tcp://" + info.DSN
		}
		return info.DSN
	}

	// ClickHouse v1 驱动使用 tcp:// 协议
	// 格式: tcp://host:port?username=user&password=pass&database=db
	var dsn string
	if info.Database != "" {
		dsn = fmt.Sprintf("tcp://%s:%s?username=%s&password=%s&database=%s",
			info.Host,
			info.Port,
			info.User,
			info.Password,
			info.Database,
		)
	} else {
		dsn = fmt.Sprintf("tcp://%s:%s?username=%s&password=%s&database=default",
			info.Host,
			info.Port,
			info.User,
			info.Password,
		)
	}

	return dsn
}
