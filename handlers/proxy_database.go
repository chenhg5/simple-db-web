package handlers

import (
	"fmt"
	"net"
	"sync"

	"github.com/chenhg5/simple-db-web/database"
	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	proxyDialerMutex sync.Mutex
	proxyDialerCount int
)

// ProxyDatabaseWrapper 包装Database接口，支持通过代理连接
type ProxyDatabaseWrapper struct {
	db    database.Database
	proxy Proxy
	netName string // 注册的网络类型名称
}

// NewProxyDatabaseWrapper 创建代理包装的数据库实例
func NewProxyDatabaseWrapper(db database.Database, proxy Proxy) *ProxyDatabaseWrapper {
	proxyDialerMutex.Lock()
	defer proxyDialerMutex.Unlock()
	proxyDialerCount++
	netName := fmt.Sprintf("proxy_%d", proxyDialerCount)
	
	return &ProxyDatabaseWrapper{
		db:      db,
		proxy:   proxy,
		netName: netName,
	}
}

// Connect 通过代理建立数据库连接
func (p *ProxyDatabaseWrapper) Connect(dsn string) error {
	// 解析DSN
	config, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		return fmt.Errorf("解析DSN失败: %w", err)
	}

	// 保存原始地址
	originalAddr := config.Addr

	// 注册自定义网络类型（使用代理）
	mysqlDriver.RegisterDial(p.netName, func(addr string) (net.Conn, error) {
		// 通过代理连接到目标地址
		return p.proxy.Dial("tcp", addr)
	})

	// 修改DSN使用自定义网络类型
	config.Net = p.netName
	config.Addr = originalAddr

	// 重新构建DSN
	newDSN := config.FormatDSN()

	// 使用包装的数据库连接
	return p.db.Connect(newDSN)
}

// Close 关闭代理连接
func (p *ProxyDatabaseWrapper) Close() error {
	var errs []error
	if p.db != nil {
		if err := p.db.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if p.proxy != nil {
		if err := p.proxy.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("关闭连接失败: %v", errs)
	}
	return nil
}

// 实现database.Database接口的所有方法，委托给内部db
func (p *ProxyDatabaseWrapper) GetTables() ([]string, error) {
	return p.db.GetTables()
}

func (p *ProxyDatabaseWrapper) GetTableSchema(tableName string) (string, error) {
	return p.db.GetTableSchema(tableName)
}

func (p *ProxyDatabaseWrapper) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	return p.db.ExecuteQuery(query)
}

func (p *ProxyDatabaseWrapper) ExecuteUpdate(query string) (int64, error) {
	return p.db.ExecuteUpdate(query)
}

func (p *ProxyDatabaseWrapper) ExecuteDelete(query string) (int64, error) {
	return p.db.ExecuteDelete(query)
}

func (p *ProxyDatabaseWrapper) ExecuteInsert(query string) (int64, error) {
	return p.db.ExecuteInsert(query)
}

func (p *ProxyDatabaseWrapper) GetTableData(tableName string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	return p.db.GetTableData(tableName, page, pageSize)
}

func (p *ProxyDatabaseWrapper) GetTableColumns(tableName string) ([]database.ColumnInfo, error) {
	return p.db.GetTableColumns(tableName)
}

func (p *ProxyDatabaseWrapper) GetDatabases() ([]string, error) {
	return p.db.GetDatabases()
}

func (p *ProxyDatabaseWrapper) SwitchDatabase(databaseName string) error {
	return p.db.SwitchDatabase(databaseName)
}

