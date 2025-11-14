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

	// GetTableDataByID 基于主键ID获取表数据（高性能分页）
	// tableName: 表名
	// primaryKey: 主键列名
	// lastId: 上一页的最后一个ID值（nil表示第一页，用于next方向）或当前页的第一个ID（用于prev方向）
	// pageSize: 每页大小
	// direction: 分页方向，"next"表示下一页（id > lastId），"prev"表示上一页（id < lastId）
	// 返回: 数据列表, 总数, 下一页/上一页的最后一个ID, 错误
	GetTableDataByID(tableName string, primaryKey string, lastId interface{}, pageSize int, direction string) ([]map[string]interface{}, int64, interface{}, error)
	
	// GetPageIdByPageNumber 根据页码计算该页的起始ID（用于页码跳转）
	// tableName: 表名
	// primaryKey: 主键列名
	// page: 目标页码（从1开始）
	// pageSize: 每页大小
	// 返回: 该页的起始ID（即上一页的最后一个ID），错误
	GetPageIdByPageNumber(tableName string, primaryKey string, page, pageSize int) (interface{}, error)

	// GetTableColumns 获取表的列信息
	GetTableColumns(tableName string) ([]ColumnInfo, error)

	// GetDatabases 获取所有数据库名称列表
	GetDatabases() ([]string, error)

	// SwitchDatabase 切换当前使用的数据库
	SwitchDatabase(databaseName string) error

	// GetTypeName 获取数据库类型名称（如 "mysql", "postgresql" 等）
	GetTypeName() string

	// GetDisplayName 获取数据库显示名称（如 "MySQL", "PostgreSQL" 等）
	GetDisplayName() string
}

// ColumnInfo 列信息
type ColumnInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"default_value"`
	Key          string `json:"key"` // PRI, UNI, MUL等
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Type     string `json:"type"`     // 代理类型，如 "ssh", "socks5" 等
	Host     string `json:"host"`     // 代理服务器地址
	Port     string `json:"port"`     // 代理服务器端口
	User     string `json:"user"`     // 代理用户名（如果需要）
	Password string `json:"password"` // 代理密码（如果需要）
	KeyFile  string `json:"key_file"` // SSH密钥文件路径（仅SSH）
	Config   string `json:"config"`   // 其他代理配置（JSON字符串，用于自定义代理）
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	Name     string       `json:"name"`     // 连接名称（可选，用于显示）
	Type     string       `json:"type"`    // mysql, postgresql等
	Host     string       `json:"host"`    // 数据库主机地址
	Port     string       `json:"port"`     // 数据库端口
	User     string       `json:"user"`    // 数据库用户名
	Password string       `json:"password"` // 数据库密码
	Database string       `json:"database"` // 数据库名
	DSN      string       `json:"dsn"`      // 如果提供DSN，则优先使用
	Proxy    *ProxyConfig `json:"proxy"`    // 代理配置（可选）
}
