package database

// Database 定义数据库操作的通用接口
type Database interface {
	// Connect 建立数据库连接
	// dsn: 数据库连接字符串，格式如 "user:password@tcp(host:port)/dbname"
	Connect(dsn string) error

	// Close 关闭数据库连接
	Close() error

	// GetTables 获取数据库中所有表的名称
	GetTables() ([]string, error)

	// GetTableSchema 获取表的结构信息
	// 返回格式化的表结构字符串
	GetTableSchema(tableName string) (string, error)

	// ExecuteQuery 执行查询SQL，返回结果集
	// 返回格式为 []map[string]interface{}，每个map代表一行数据
	ExecuteQuery(query string) ([]map[string]interface{}, error)

	// ExecuteUpdate 执行更新SQL
	// 返回受影响的行数
	ExecuteUpdate(query string) (int64, error)

	// ExecuteDelete 执行删除SQL
	// 返回受影响的行数
	ExecuteDelete(query string) (int64, error)

	// ExecuteInsert 执行插入SQL
	// 返回受影响的行数
	ExecuteInsert(query string) (int64, error)

	// GetTableData 获取表的数据（分页）
	// tableName: 表名
	// page: 页码（从1开始）
	// pageSize: 每页大小
	GetTableData(tableName string, page, pageSize int) ([]map[string]interface{}, int64, error)

	// GetTableColumns 获取表的列信息
	GetTableColumns(tableName string) ([]ColumnInfo, error)

	// GetDatabases 获取所有数据库名称列表
	GetDatabases() ([]string, error)

	// SwitchDatabase 切换当前使用的数据库
	SwitchDatabase(databaseName string) error
}

// ColumnInfo 列信息
type ColumnInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"default_value"`
	Key          string `json:"key"` // PRI, UNI, MUL等
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	Type     string `json:"type"` // mysql, postgresql等
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	DSN      string `json:"dsn"` // 如果提供DSN，则优先使用
}
