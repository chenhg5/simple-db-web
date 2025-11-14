package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/chenhg5/simple-db-web/database"
)

// SessionData 会话数据（可序列化）
// 用于持久化存储，不包含实际的数据库连接对象
type SessionData struct {
	ConnectionInfo  database.ConnectionInfo `json:"connection_info"`  // 连接信息（用于重建连接）
	DSN             string                  `json:"dsn"`              // DSN连接字符串
	DbType          string                  `json:"db_type"`          // 数据库类型
	CurrentDatabase string                  `json:"current_database"` // 当前数据库
	CurrentTable    string                  `json:"current_table"`    // 当前表
	CreatedAt       time.Time               `json:"created_at"`       // 创建时间
}

// ConnectionSession 连接会话信息（运行时对象）
// 包含实际的数据库连接对象
type ConnectionSession struct {
	db              database.Database
	dbType          string // 数据库类型，用于前端判断
	currentDatabase string
	currentTable    string
	createdAt       time.Time
	sessionData     *SessionData // 保存原始数据，用于持久化
}

// SessionStorage 会话存储接口
// 允许外部项目实现自定义的持久化存储（如Redis、MySQL等）
type SessionStorage interface {
	// Get 获取会话数据
	// connectionID: 连接ID
	// 返回会话数据，如果不存在返回nil和error
	Get(connectionID string) (*SessionData, error)

	// Set 保存会话数据
	// connectionID: 连接ID
	// data: 会话数据
	// ttl: 过期时间（秒），0表示不过期
	Set(connectionID string, data *SessionData, ttl time.Duration) error

	// Delete 删除会话数据
	// connectionID: 连接ID
	Delete(connectionID string) error

	// Close 关闭存储连接（如果需要）
	Close() error
}

// Proxy 代理接口
// 允许外部项目实现自定义的代理协议（如SSH、SOCKS5等）
type Proxy interface {
	// Dial 建立代理连接
	// network: 网络类型，如 "tcp"
	// address: 目标地址，格式如 "host:port"
	// 返回代理后的连接
	Dial(network, address string) (net.Conn, error)

	// Close 关闭代理连接
	Close() error
}

// ProxyFactory 代理工厂函数类型
// config: 代理配置的JSON字符串
type ProxyFactory func(config string) (Proxy, error)

// DatabaseFactory 数据库工厂函数类型
type DatabaseFactory func() database.Database

// Logger 日志接口
// 允许外部项目实现自定义的日志记录器
type Logger interface {
	// Debug 记录调试日志
	Debug(ctx context.Context, format string, args ...interface{})
	// Info 记录信息日志
	Info(ctx context.Context, format string, args ...interface{})
	// Warn 记录警告日志
	Warn(ctx context.Context, format string, args ...interface{})
	// Error 记录错误日志
	Error(ctx context.Context, format string, args ...interface{})
}

// DefaultLogger 默认日志实现（使用标准库log包）
type DefaultLogger struct{}

// Debug 实现Logger接口
func (l *DefaultLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

// Info 实现Logger接口
func (l *DefaultLogger) Info(ctx context.Context, format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

// Warn 实现Logger接口
func (l *DefaultLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

// Error 实现Logger接口
func (l *DefaultLogger) Error(ctx context.Context, format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

// SQLValidator SQL校验器接口
// 允许外部项目实现自定义的SQL校验规则
type SQLValidator interface {
	// Validate 校验SQL语句
	// query: SQL查询语句
	// queryType: SQL类型（SELECT, UPDATE, DELETE, INSERT等）
	// 返回错误信息，如果校验通过返回nil
	Validate(query string, queryType string) error

	// Name 返回校验器名称（用于日志和错误提示）
	Name() string
}

// SQLValidatorFunc SQL校验器函数类型（简化版本）
// 可以直接使用函数作为校验器
type SQLValidatorFunc func(query string, queryType string) error

// Validate 实现SQLValidator接口
func (f SQLValidatorFunc) Validate(query string, queryType string) error {
	return f(query, queryType)
}

// Name 返回校验器名称
func (f SQLValidatorFunc) Name() string {
	return "CustomValidator"
}

// DatabaseTypeInfo 数据库类型信息
type DatabaseTypeInfo struct {
	Type        string `json:"type"`         // 数据库类型标识
	DisplayName string `json:"display_name"` // 显示名称
}

// MemorySessionStorage 内存会话存储（默认实现）
// 适用于单实例部署，多实例部署请使用Redis等外部存储
type MemorySessionStorage struct {
	sessions map[string]*SessionData
	mutex    sync.RWMutex
}

// NewMemorySessionStorage 创建内存会话存储
func NewMemorySessionStorage() *MemorySessionStorage {
	return &MemorySessionStorage{
		sessions: make(map[string]*SessionData),
	}
}

// Get 获取会话数据
func (m *MemorySessionStorage) Get(connectionID string) (*SessionData, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	session, exists := m.sessions[connectionID]
	if !exists {
		return nil, fmt.Errorf("会话不存在")
	}
	// 返回副本，避免并发修改
	sessionCopy := *session
	return &sessionCopy, nil
}

// Set 保存会话数据
func (m *MemorySessionStorage) Set(connectionID string, data *SessionData, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// 创建副本保存
	dataCopy := *data
	m.sessions[connectionID] = &dataCopy
	// 注意：内存存储不支持TTL，如果需要TTL请使用Redis等外部存储
	return nil
}

// Delete 删除会话数据
func (m *MemorySessionStorage) Delete(connectionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.sessions, connectionID)
	return nil
}

// Close 关闭存储连接（内存存储无需关闭）
func (m *MemorySessionStorage) Close() error {
	return nil
}

// Server 服务器结构
type Server struct {
	templates            *template.Template
	sessionStorage       SessionStorage                // 会话存储接口
	sessions             map[string]*ConnectionSession // 运行时会话缓存（包含实际连接）
	sessionsMutex        sync.RWMutex
	customDatabases      map[string]DatabaseFactory // 自定义数据库类型
	customDbDisplayNames map[string]string          // 自定义数据库类型的显示名称
	customDbMutex        sync.RWMutex
	customProxies        map[string]ProxyFactory // 自定义代理类型
	customProxyMutex     sync.RWMutex
	builtinTypes         map[string]string // 内置数据库类型及其显示名称
	customScript         string            // 自定义JavaScript脚本，会在页面加载后执行
	customScriptMutex    sync.RWMutex      // 保护customScript的读写锁
	validators           []SQLValidator    // SQL校验器列表
	validatorsMutex      sync.RWMutex      // 保护validators的读写锁
	logger               Logger            // 日志记录器
	loggerMutex          sync.RWMutex      // 保护logger的读写锁
	presetConnections    []database.ConnectionInfo // 预设连接列表
	presetConnectionsMutex sync.RWMutex    // 保护presetConnections的读写锁
}

// NewServer 创建新的服务器实例
// 使用 embed 嵌入的模板和静态文件，支持通过 go mod 引入
func NewServer() (*Server, error) {
	// 从 embed 文件系统解析模板
	// templatesFS 使用 all:templates，所以路径是 templates/*.html
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("加载模板失败: %w", err)
	}

	// 初始化内置数据库类型
	builtinTypes := map[string]string{
		"mysql":      "MySQL",
		"postgresql": "PostgreSQL",
		"sqlite":     "SQLite",
		"clickhouse": "ClickHouse",
		"oceandb":    "OceanBase",
		"oracle":     "Oracle",
		"sqlserver":  "SQL Server",
		"mssql":      "SQL Server",
		"mongodb":    "MongoDB",
	}

	server := &Server{
		templates:            tmpl,
		sessionStorage:       NewMemorySessionStorage(), // 默认使用内存存储
		sessions:             make(map[string]*ConnectionSession),
		customDatabases:      make(map[string]DatabaseFactory),
		customDbDisplayNames: make(map[string]string),
		customProxies:        make(map[string]ProxyFactory),
		builtinTypes:         builtinTypes,
		validators:           make([]SQLValidator, 0),
		logger:               &DefaultLogger{}, // 默认使用标准库log
		presetConnections:    make([]database.ConnectionInfo, 0),
	}

	// 注册默认的SSH代理
	server.AddProxy("ssh", NewSSHProxy)

	// 注册默认的SQL校验器
	server.AddValidator(NewRequireLimitValidator())
	server.AddValidator(NewNoDropTableValidator())
	server.AddValidator(NewNoTruncateValidator())

	return server, nil
}

// SetSessionStorage 设置自定义会话存储
// 允许外部项目使用Redis、MySQL等外部存储
// 示例：
//
//	redisStorage := NewRedisSessionStorage(redisClient)
//	server.SetSessionStorage(redisStorage)
func (s *Server) SetSessionStorage(storage SessionStorage) {
	s.sessionsMutex.Lock()
	defer s.sessionsMutex.Unlock()
	// 关闭旧的存储
	if s.sessionStorage != nil {
		s.sessionStorage.Close()
	}
	s.sessionStorage = storage
}

// AddDatabase 添加自定义数据库类型
// name: 数据库类型标识（如 "custom_db"）
// factory: 创建数据库实例的工厂函数
func (s *Server) AddDatabase(name string, factory DatabaseFactory) {
	s.customDbMutex.Lock()
	defer s.customDbMutex.Unlock()
	s.customDatabases[name] = factory
}

// AddDatabaseWithDisplayName 添加自定义数据库类型并指定显示名称
// name: 数据库类型标识（如 "custom_db"）
// displayName: 显示名称（如 "自定义数据库"）
// factory: 创建数据库实例的工厂函数
// 示例：
//
//	server.AddDatabaseWithDisplayName("dameng", "达梦数据库", func() database.Database {
//	    return database.NewBaseMysqlBasedDB("dameng")
//	})
func (s *Server) AddDatabaseWithDisplayName(name string, displayName string, factory DatabaseFactory) {
	s.customDbMutex.Lock()
	defer s.customDbMutex.Unlock()
	s.customDatabases[name] = factory
	s.customDbDisplayNames[name] = displayName
}

// AddProxy 添加自定义代理类型
// name: 代理类型标识（如 "socks5"）
// factory: 创建代理实例的工厂函数
func (s *Server) AddProxy(name string, factory ProxyFactory) {
	s.customProxyMutex.Lock()
	defer s.customProxyMutex.Unlock()
	s.customProxies[name] = factory
}

// AddValidator 添加SQL校验器
// 允许外部项目注册自定义的SQL校验规则
// 示例：
//
//	server.AddValidator(MyCustomValidator{})
//	// 或使用函数
//	server.AddValidator(SQLValidatorFunc(func(query, queryType string) error {
//	    if strings.Contains(query, "DROP") {
//	        return fmt.Errorf("不允许执行DROP语句")
//	    }
//	    return nil
//	}))
func (s *Server) AddValidator(validator SQLValidator) {
	s.validatorsMutex.Lock()
	defer s.validatorsMutex.Unlock()
	s.validators = append(s.validators, validator)
}

// validateSQL 执行所有注册的SQL校验器
func (s *Server) validateSQL(query string, queryType string) error {
	s.validatorsMutex.RLock()
	defer s.validatorsMutex.RUnlock()

	for _, validator := range s.validators {
		if err := validator.Validate(query, queryType); err != nil {
			// 如果错误消息是错误代码（以 "error." 开头），直接返回
			// 否则包装错误消息
			errMsg := err.Error()
			if strings.HasPrefix(errMsg, "error.") {
				// 检查是否包含参数（格式：error.xxx: param）
				if strings.Contains(errMsg, ": ") {
					parts := strings.SplitN(errMsg, ": ", 2)
					if len(parts) == 2 {
						// 返回错误代码和参数
						return fmt.Errorf("%s: %s", parts[0], parts[1])
					}
				}
				return err
			}
			return fmt.Errorf("[%s] %v", validator.Name(), err)
		}
	}
	return nil
}

// GetDatabaseTypes 获取所有可用的数据库类型列表
func (s *Server) GetDatabaseTypes(w http.ResponseWriter, r *http.Request) {
	s.customDbMutex.RLock()
	defer s.customDbMutex.RUnlock()

	types := make([]DatabaseTypeInfo, 0)

	// 添加内置类型
	for dbType, displayName := range s.builtinTypes {
		types = append(types, DatabaseTypeInfo{
			Type:        dbType,
			DisplayName: displayName,
		})
	}

	// 添加自定义类型
	for dbType := range s.customDatabases {
		// 如果自定义类型不在内置类型中，添加它
		if _, exists := s.builtinTypes[dbType]; !exists {
			// 优先使用自定义显示名称，如果没有则使用类型名
			displayName := dbType
			if customDisplayName, hasCustomName := s.customDbDisplayNames[dbType]; hasCustomName {
				displayName = customDisplayName
			}
			types = append(types, DatabaseTypeInfo{
				Type:        dbType,
				DisplayName: displayName,
			})
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"types":   types,
	})
}

// decryptPassword 解密密码（Base64解码，与前端encryptPassword对应）
func decryptPassword(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	// Base64解码
	decoded, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	// 转换为UTF-8字符串
	if !utf8.Valid(decoded) {
		return "", fmt.Errorf("解码后的数据不是有效的UTF-8字符串")
	}

	return string(decoded), nil
}

// generateConnectionID 生成唯一的连接ID
func generateConnectionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// getConnectionID 从请求中获取连接ID
func getConnectionID(r *http.Request) string {
	// 优先从请求头获取
	connID := r.Header.Get("X-Connection-ID")
	if connID != "" {
		return connID
	}
	// 兼容：从查询参数获取
	return r.URL.Query().Get("connectionId")
}

// 错误代码常量
const (
	ErrCodeMethodNotAllowed           = "error.methodNotAllowed"
	ErrCodeMissingConnectionID        = "error.missingConnectionID"
	ErrCodeMissingTableName           = "error.missingTableName"
	ErrCodeMissingDatabaseName        = "error.missingDatabaseName"
	ErrCodeEmptySQLQuery              = "error.emptySQLQuery"
	ErrCodeUnsupportedSQLType         = "error.unsupportedSQLType"
	ErrCodeUnsupportedDatabaseType    = "error.unsupportedDatabaseType"
	ErrCodeUnsupportedProxyType       = "error.unsupportedProxyType"
	ErrCodeParseRequestFailed         = "error.parseRequestFailed"
	ErrCodeGenerateConnectionIDFailed = "error.generateConnectionIDFailed"
	ErrCodeBuildProxyConfigFailed     = "error.buildProxyConfigFailed"
	ErrCodeEstablishProxyFailed       = "error.establishProxyFailed"
	ErrCodeConnectionFailed           = "error.connectionFailed"
	ErrCodeGetTablesFailed            = "error.getTablesFailed"
	ErrCodeGetTableSchemaFailed       = "error.getTableSchemaFailed"
	ErrCodeGetTableColumnsFailed      = "error.getTableColumnsFailed"
	ErrCodeGetTableDataFailed         = "error.getTableDataFailed"
	ErrCodeGetPageIDFailed            = "error.getPageIDFailed"
	ErrCodeSQLValidationFailed        = "error.sqlValidationFailed"
	ErrCodeExecuteQueryFailed         = "error.executeQueryFailed"
	ErrCodeExecuteUpdateFailed        = "error.executeUpdateFailed"
	ErrCodeExecuteDeleteFailed        = "error.executeDeleteFailed"
	ErrCodeExecuteInsertFailed        = "error.executeInsertFailed"
	ErrCodeUpdateFailed               = "error.updateFailed"
	ErrCodeDeleteFailed               = "error.deleteFailed"
	ErrCodeGetDatabasesFailed         = "error.getDatabasesFailed"
	ErrCodeSwitchDatabaseFailed       = "error.switchDatabaseFailed"
	ErrCodeNoSinglePrimaryKey         = "error.noSinglePrimaryKey"
	ErrCodePrimaryKeyNotInteger       = "error.primaryKeyNotInteger"
	ErrCodeSelectDatabaseFirst        = "error.selectDatabaseFirst"
	ErrCodeTableNameEmpty             = "error.tableNameEmpty"
	ErrCodeClickHouseNoUpdate         = "error.clickHouseNoUpdate"
	ErrCodeClickHouseNoDelete         = "error.clickHouseNoDelete"
	ErrCodeConnectionNotExists        = "error.connectionNotExists"
	ErrCodeCreateExcelSheetFailed     = "error.createExcelSheetFailed"
	ErrCodeExportExcelFailed          = "error.exportExcelFailed"
	ErrCodeOnlySelectQueryAllowed     = "error.onlySelectQueryAllowed"
	ErrCodeQueryResultEmpty           = "error.queryResultEmpty"
	ErrCodeRequireLimit               = "error.requireLimit"
	ErrCodeNoDropTable                = "error.noDropTable"
	ErrCodeNoTruncate                 = "error.noTruncate"
	ErrCodeNoTruncateTable            = "error.noTruncateTable"
	ErrCodeNoDropDatabase             = "error.noDropDatabase"
	ErrCodeQueryTooLong               = "error.queryTooLong"
)

// writeJSONError 写入JSON格式的错误响应
// 支持错误代码和参数化消息
func writeJSONError(w http.ResponseWriter, statusCode int, errorCode string, params ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"success":   false,
		"errorCode": errorCode,
	}

	// 如果有参数，构建参数化消息（用于向后兼容）
	if len(params) > 0 {
		// 将error类型转换为字符串，避免JSON序列化为空对象
		serializedParams := make([]interface{}, len(params))
		for i, p := range params {
			if err, ok := p.(error); ok {
				// 如果是error类型，转换为字符串
				serializedParams[i] = err.Error()
			} else {
				serializedParams[i] = p
			}
		}

		// 构建参数化消息，用于日志或调试
		message := errorCode
		if len(serializedParams) == 1 {
			message = fmt.Sprintf("%s: %v", errorCode, serializedParams[0])
		} else if len(serializedParams) > 1 {
			message = fmt.Sprintf("%s: %v", errorCode, serializedParams)
		}
		response["message"] = message
		response["params"] = serializedParams
	} else {
		response["message"] = errorCode
	}

	json.NewEncoder(w).Encode(response)
}

// createDatabaseFromSessionData 根据SessionData重建数据库连接
func (s *Server) createDatabaseFromSessionData(data *SessionData) (database.Database, error) {
	var db database.Database

	// 先检查是否为自定义数据库类型
	s.customDbMutex.RLock()
	factory, isCustom := s.customDatabases[data.DbType]
	s.customDbMutex.RUnlock()

	if isCustom {
		// 使用自定义数据库工厂函数
		db = factory()
	} else {
		// 使用内置数据库类型
		switch data.DbType {
		case "mysql":
			db = database.NewMySQL()
		case "clickhouse":
			db = database.NewClickHouse()
		case "oceandb":
			db = database.NewBaseMysqlBasedDB("oceandb")
		case "sqlite":
			db = database.NewSQLite3()
		case "postgres", "postgresql":
			db = database.NewPostgreSQL()
		case "oracle":
			db = database.NewOracle()
		case "sqlserver", "mssql":
			db = database.NewSQLServer()
		case "mongodb":
			db = database.NewMongoDB()
		default:
			return nil, fmt.Errorf("不支持的数据库类型: %s", data.DbType)
		}
	}

	// 如果有代理配置，先建立代理连接
	var proxy Proxy
	if data.ConnectionInfo.Proxy != nil && data.ConnectionInfo.Proxy.Type != "" {
		s.customProxyMutex.RLock()
		proxyFactory, exists := s.customProxies[data.ConnectionInfo.Proxy.Type]
		s.customProxyMutex.RUnlock()

		if exists {
			proxyConfigJSON, err := json.Marshal(data.ConnectionInfo.Proxy)
			if err == nil {
				proxy, _ = proxyFactory(string(proxyConfigJSON))
			}
		}
	}

	// 如果有代理，使用代理包装器
	if proxy != nil {
		// 目前只支持MySQL及其兼容数据库的代理连接
		if data.DbType == "mysql" || strings.HasPrefix(data.DbType, "mysql_based_") || data.DbType == "oceandb" {
			// MySQL及其兼容数据库使用代理包装器
			db = NewProxyDatabaseWrapper(db, proxy)
		} else {
			// 其他数据库类型暂不支持代理，记录警告并关闭代理
			s.getLogger().Warn(context.Background(), "Database type %s does not support proxy connection, attempting direct connection", data.DbType)
			proxy.Close()
			proxy = nil
		}
	}

	if err := db.Connect(data.DSN); err != nil {
		if proxy != nil {
			proxy.Close()
		}
		return nil, fmt.Errorf("重建连接失败: %w", err)
	}

	// 如果之前选择了数据库，切换回去
	if data.CurrentDatabase != "" {
		if err := db.SwitchDatabase(data.CurrentDatabase); err != nil {
			// 切换失败返回错误，确保数据库正确切换
			db.Close()
			if proxy != nil {
				proxy.Close()
			}
			return nil, fmt.Errorf("切换数据库失败: %w", err)
		}
	}

	return db, nil
}

// getSession 根据连接ID获取会话
// 如果内存缓存中没有，会尝试从持久化存储重建
// 优化：尽量复用内存中的连接，避免频繁重连
func (s *Server) getSession(connectionID string) (*ConnectionSession, error) {
	// 从持久化存储获取（这是权威数据源）
	sessionData, err := s.sessionStorage.Get(connectionID)
	if err != nil {
		return nil, fmt.Errorf("连接不存在或已断开")
	}

	// 检查内存缓存
	s.sessionsMutex.RLock()
	session, exists := s.sessions[connectionID]
	s.sessionsMutex.RUnlock()

	// 如果内存中存在且数据库一致，直接返回（最佳情况：无需任何操作）
	if exists && session != nil && session.db != nil && session.currentDatabase == sessionData.CurrentDatabase {
		// 确保sessionData引用是最新的
		session.sessionData = sessionData
		// 更新当前表（可能在其他请求中被更新）
		if session.currentTable != sessionData.CurrentTable {
			session.currentTable = sessionData.CurrentTable
		}
		return session, nil
	}

	// 如果内存中存在连接但数据库不一致，尝试在现有连接上切换数据库
	// 注意：对于MySQL等数据库，SwitchDatabase实际上是重连，但至少我们复用了连接对象
	if exists && session != nil && session.db != nil && session.currentDatabase != sessionData.CurrentDatabase {
		// 需要切换数据库
		if sessionData.CurrentDatabase != "" {
			// 尝试在现有连接上切换数据库
			if err := session.db.SwitchDatabase(sessionData.CurrentDatabase); err != nil {
				// 切换失败，关闭旧连接并重建
				session.db.Close()
				session.db = nil
			} else {
				// 切换成功，更新会话信息
				s.sessionsMutex.Lock()
				session.currentDatabase = sessionData.CurrentDatabase
				session.currentTable = sessionData.CurrentTable
				session.sessionData = sessionData
				s.sessionsMutex.Unlock()
				return session, nil
			}
		} else {
			// 持久化存储中没有数据库，但内存中有，这种情况不应该发生
			// 为了安全，关闭旧连接并重建
			session.db.Close()
			session.db = nil
		}
	}

	// 如果内存中没有连接或连接已关闭，需要重建
	// 重建数据库连接（使用持久化存储中的最新数据）
	db, err := s.createDatabaseFromSessionData(sessionData)
	if err != nil {
		return nil, fmt.Errorf("重建连接失败: %w", err)
	}

	// 创建或更新会话对象
	if exists && session != nil {
		// 更新现有会话
		s.sessionsMutex.Lock()
		session.db = db
		session.currentDatabase = sessionData.CurrentDatabase
		session.currentTable = sessionData.CurrentTable
		session.sessionData = sessionData
		s.sessionsMutex.Unlock()
	} else {
		// 创建新会话
		session = &ConnectionSession{
			db:              db,
			dbType:          sessionData.DbType,
			currentDatabase: sessionData.CurrentDatabase,
			currentTable:    sessionData.CurrentTable,
			createdAt:       sessionData.CreatedAt,
			sessionData:     sessionData,
		}

		// 保存到内存缓存
		s.sessionsMutex.Lock()
		s.sessions[connectionID] = session
		s.sessionsMutex.Unlock()
	}

	return session, nil
}

// updateSession 更新会话并同步到持久化存储
func (s *Server) updateSession(connectionID string, updateFn func(*ConnectionSession)) error {
	session, err := s.getSession(connectionID)
	if err != nil {
		return err
	}

	// 更新内存中的会话
	s.sessionsMutex.Lock()
	updateFn(session)
	// 同步更新sessionData
	if session.sessionData != nil {
		session.sessionData.CurrentDatabase = session.currentDatabase
		session.sessionData.CurrentTable = session.currentTable
	}
	s.sessionsMutex.Unlock()

	// 同步到持久化存储
	if session.sessionData != nil {
		ttl := 24 * time.Hour
		if err := s.sessionStorage.Set(connectionID, session.sessionData, ttl); err != nil {
			s.getLogger().Warn(context.Background(), "Failed to update session to persistent storage: %v", err)
			// 不返回错误，因为内存中的会话已经更新
		}
	}

	return nil
}

// SetLogger 设置自定义日志记录器
// 允许外部项目使用自定义的日志库（如zap、logrus等）
// 示例：
//
//	server.SetLogger(MyCustomLogger{})
func (s *Server) SetLogger(logger Logger) {
	s.loggerMutex.Lock()
	defer s.loggerMutex.Unlock()
	if logger == nil {
		s.logger = &DefaultLogger{}
	} else {
		s.logger = logger
	}
}

// getLogger 获取日志记录器（线程安全）
func (s *Server) getLogger() Logger {
	s.loggerMutex.RLock()
	defer s.loggerMutex.RUnlock()
	if s.logger == nil {
		return &DefaultLogger{}
	}
	return s.logger
}

// SetCustomScript 设置自定义JavaScript脚本
// 这个脚本会在页面加载后执行，可以用于配置请求拦截器等
// 示例：
//
//	server.SetCustomScript(`
//	  window.SimpleDB.config.requestInterceptor = function(url, options) {
//	    const token = getCookie('auth_token');
//	    if (token) {
//	      options.headers = options.headers || {};
//	      options.headers['Authorization'] = 'Bearer ' + token;
//	    }
//	    return { url, options };
//	  };
//	`)
func (s *Server) SetCustomScript(script string) {
	s.customScriptMutex.Lock()
	defer s.customScriptMutex.Unlock()
	s.customScript = script
}

// GetCustomScript 获取自定义脚本
func (s *Server) GetCustomScript() string {
	s.customScriptMutex.RLock()
	defer s.customScriptMutex.RUnlock()
	return s.customScript
}

// SetPresetConnections 设置预设连接列表
// 允许外部项目在启动时预设一些已保存的连接
// 这些连接会通过 API 提供给前端，前端会自动保存到本地并统一显示
// 示例：
//
//	presetConns := []database.ConnectionInfo{
//	    {
//	        Type:     "mysql",
//	        Host:     "localhost",
//	        Port:     "3306",
//	        User:     "root",
//	        Password: "password",
//	        Database: "testdb",
//	    },
//	}
//	server.SetPresetConnections(presetConns)
func (s *Server) SetPresetConnections(connections []database.ConnectionInfo) {
	s.presetConnectionsMutex.Lock()
	defer s.presetConnectionsMutex.Unlock()
	// 创建副本，避免外部修改
	s.presetConnections = make([]database.ConnectionInfo, len(connections))
	for i, conn := range connections {
		s.presetConnections[i] = conn
	}
}

// GetPresetConnections 获取预设连接列表（线程安全）
func (s *Server) GetPresetConnections() []database.ConnectionInfo {
	s.presetConnectionsMutex.RLock()
	defer s.presetConnectionsMutex.RUnlock()
	// 返回副本，避免外部修改
	result := make([]database.ConnectionInfo, len(s.presetConnections))
	for i, conn := range s.presetConnections {
		result[i] = conn
	}
	return result
}

// GetPresetConnectionsAPI 获取预设连接列表的API端点
func (s *Server) GetPresetConnectionsAPI(w http.ResponseWriter, r *http.Request) {
	presetConns := s.GetPresetConnections()
	
	// 转换为前端需要的格式（与保存的连接格式一致）
	connections := make([]map[string]interface{}, 0, len(presetConns))
	for _, conn := range presetConns {
		connMap := map[string]interface{}{
			"type":     conn.Type,
			"name":     conn.Name,
			"host":     conn.Host,
			"port":     conn.Port,
			"user":     conn.User,
			"password": conn.Password, // 注意：密码需要前端加密后保存
			"database": conn.Database,
			"dsn":      conn.DSN,
			"proxy":    conn.Proxy,
			"preset":   true, // 标记为预设连接
		}
		connections = append(connections, connMap)
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"connections": connections,
	})
}

// Home 首页
func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
	s.customScriptMutex.RLock()
	customScript := s.customScript
	s.customScriptMutex.RUnlock()

	data := map[string]interface{}{
		"CustomScript": template.JS(customScript), // 使用template.JS避免转义
	}

	if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Connect 连接数据库
func (s *Server) Connect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, ErrCodeMethodNotAllowed)
		return
	}

	var info database.ConnectionInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeParseRequestFailed, err)
		return
	}

	// 解密数据库密码（如果提供了）
	if info.Password != "" {
		decryptedPassword, err := decryptPassword(info.Password)
		if err != nil {
			s.getLogger().Error(r.Context(), "Failed to decrypt database password: %v", err)
			writeJSONError(w, http.StatusBadRequest, ErrCodeParseRequestFailed, fmt.Errorf("密码解密失败"))
			return
		}
		info.Password = decryptedPassword
	}

	// 解密代理密码和私钥（如果提供了代理配置）
	if info.Proxy != nil {
		// 解密代理密码
		if info.Proxy.Password != "" {
			decryptedProxyPassword, err := decryptPassword(info.Proxy.Password)
			if err != nil {
				s.getLogger().Error(r.Context(), "Failed to decrypt proxy password: %v", err)
				writeJSONError(w, http.StatusBadRequest, ErrCodeParseRequestFailed, fmt.Errorf("代理密码解密失败"))
				return
			}
			info.Proxy.Password = decryptedProxyPassword
		}

		// 解密私钥（如果存在）
		if info.Proxy.Config != "" {
			var config map[string]interface{}
			if err := json.Unmarshal([]byte(info.Proxy.Config), &config); err == nil {
				if keyData, ok := config["key_data"].(string); ok && keyData != "" {
					decryptedKeyData, err := decryptPassword(keyData)
					if err != nil {
						s.getLogger().Error(r.Context(), "Failed to decrypt SSH private key: %v", err)
						writeJSONError(w, http.StatusBadRequest, ErrCodeParseRequestFailed, fmt.Errorf("SSH私钥解密失败"))
						return
					}
					config["key_data"] = decryptedKeyData
					configJSON, _ := json.Marshal(config)
					info.Proxy.Config = string(configJSON)
				}
			}
		}
	}

	// 生成连接ID
	connectionID, err := generateConnectionID()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeGenerateConnectionIDFailed, err)
		return
	}

	// 创建新连接
	var db database.Database

	// 先检查是否为自定义数据库类型
	s.customDbMutex.RLock()
	factory, isCustom := s.customDatabases[info.Type]
	s.customDbMutex.RUnlock()

	if isCustom {
		// 使用自定义数据库工厂函数
		db = factory()
	} else {
		// 使用内置数据库类型
		switch info.Type {
		case "mysql":
			db = database.NewMySQL()
		case "clickhouse":
			db = database.NewClickHouse()
		case "oceandb":
			db = database.NewBaseMysqlBasedDB("oceandb")
		case "sqlite":
			db = database.NewSQLite3()
		case "postgres", "postgresql":
			db = database.NewPostgreSQL()
		case "oracle":
			db = database.NewOracle()
		case "sqlserver", "mssql":
			db = database.NewSQLServer()
		case "mongodb":
			db = database.NewMongoDB()
		default:
			writeJSONError(w, http.StatusBadRequest, ErrCodeUnsupportedDatabaseType, info.Type)
			return
		}
	}

	// 如果有代理配置，先建立代理连接
	var proxy Proxy
	if info.Proxy != nil && info.Proxy.Type != "" {
		// 获取代理工厂
		s.customProxyMutex.RLock()
		proxyFactory, exists := s.customProxies[info.Proxy.Type]
		s.customProxyMutex.RUnlock()

		if !exists {
			writeJSONError(w, http.StatusBadRequest, ErrCodeUnsupportedProxyType, info.Proxy.Type)
			return
		}

		// 构建代理配置JSON
		proxyConfigJSON, err := json.Marshal(info.Proxy)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, ErrCodeBuildProxyConfigFailed, err)
			return
		}

		// 创建代理
		proxy, err = proxyFactory(string(proxyConfigJSON))
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, ErrCodeEstablishProxyFailed, err)
			return
		}
		defer func() {
			// 注意：代理连接会在会话关闭时关闭，这里不立即关闭
		}()
	}

	// 构建DSN
	var dsn string
	switch info.Type {
	case "clickhouse":
		dsn = database.BuildClickHouseDSN(info)
	case "sqlite":
		dsn = database.BuildSQLite3DSN(info)
	case "oracle":
		dsn = database.BuildOracleDSN(info)
	case "sqlserver", "mssql":
		dsn = database.BuildSQLServerDSN(info)
	case "mongodb":
		dsn = database.BuildMongoDBDSN(info)
	default:
		dsn = database.BuildDSN(info)
	}

	// 如果有代理，使用代理包装器
	if proxy != nil {
		// 目前只支持MySQL及其兼容数据库的代理连接
		// 其他数据库类型的代理支持需要进一步实现
		if info.Type == "mysql" || strings.HasPrefix(info.Type, "mysql_based_") || info.Type == "oceandb" {
			// MySQL及其兼容数据库使用代理包装器
			db = NewProxyDatabaseWrapper(db, proxy)
		} else {
			// 其他数据库类型暂不支持代理，记录警告
			s.getLogger().Warn(r.Context(), "Database type %s does not support proxy connection, attempting direct connection", info.Type)
			proxy.Close()
			proxy = nil
		}
	}

	if err := db.Connect(dsn); err != nil {
		if proxy != nil {
			proxy.Close()
		}
		writeJSONError(w, http.StatusInternalServerError, ErrCodeConnectionFailed, err)
		return
	}

	// 创建会话数据（用于持久化）
	sessionData := &SessionData{
		ConnectionInfo:  info,
		DSN:             dsn,
		DbType:          info.Type,
		CurrentDatabase: "",
		CurrentTable:    "",
		CreatedAt:       time.Now(),
	}

	// 保存到持久化存储（默认TTL为24小时）
	ttl := 24 * time.Hour
	if err := s.sessionStorage.Set(connectionID, sessionData, ttl); err != nil {
		s.getLogger().Warn(r.Context(), "Failed to save session to persistent storage: %v", err)
		// 继续执行，不中断连接流程
	}

	// 创建运行时会话对象
	session := &ConnectionSession{
		db:              db,
		dbType:          info.Type,
		currentDatabase: "",
		currentTable:    "",
		createdAt:       time.Now(),
		sessionData:     sessionData,
	}

	// 保存到内存缓存
	s.sessionsMutex.Lock()
	s.sessions[connectionID] = session
	s.sessionsMutex.Unlock()

	// 获取数据库列表
	databases, err := db.GetDatabases()
	if err != nil {
		// 如果获取数据库列表失败,仍然返回连接成功,但数据库列表为空
		databases = []string{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"message":      "连接成功",
		"databases":    databases,
		"connectionId": connectionID,
	})
}

// GetTables 获取表列表
func (s *Server) GetTables(w http.ResponseWriter, r *http.Request) {
	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingConnectionID)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeConnectionNotExists, err)
		return
	}

	tables, err := session.db.GetTables()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeGetTablesFailed, err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tables":  tables,
	})
}

// GetTableSchema 获取表结构
func (s *Server) GetTableSchema(w http.ResponseWriter, r *http.Request) {
	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingConnectionID)
		return
	}

	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingTableName)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeConnectionNotExists, err)
		return
	}

	s.updateSession(connectionID, func(s *ConnectionSession) {
		s.currentTable = tableName
	})
	schema, err := session.db.GetTableSchema(tableName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeGetTableSchemaFailed, err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"schema":  schema,
	})
}

// GetTableColumns 获取表列信息
func (s *Server) GetTableColumns(w http.ResponseWriter, r *http.Request) {
	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingConnectionID)
		return
	}

	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingTableName)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeConnectionNotExists, err)
		return
	}

	// 确保数据库已选择（从持久化存储获取的会话应该已经切换了数据库，但为了安全再次检查）
	// SQLite3、H2 没有数据库概念，MongoDB 在连接时已选择数据库，跳过数据库检查
	if session.currentDatabase == "" && session.dbType != "sqlite" && session.dbType != "h2" && session.dbType != "mongodb" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeSelectDatabaseFirst)
		return
	}

	columns, err := session.db.GetTableColumns(tableName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeGetTableColumnsFailed, err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"columns": columns,
	})
}

// GetTableData 获取表数据
func (s *Server) GetTableData(w http.ResponseWriter, r *http.Request) {
	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingConnectionID)
		return
	}

	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingTableName)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 50
	}

	// 获取lastId参数（用于基于ID的分页）
	lastIdStr := r.URL.Query().Get("lastId")
	var lastId interface{} = nil
	if lastIdStr != "" {
		// 尝试解析为整数
		if idInt, err := strconv.ParseInt(lastIdStr, 10, 64); err == nil {
			lastId = idInt
		} else {
			// 如果不是整数，保持为字符串（用于UUID等）
			lastId = lastIdStr
		}
	}

	// 获取direction参数（用于基于ID的分页方向）
	direction := r.URL.Query().Get("direction")
	if direction == "" {
		direction = "next" // 默认为下一页
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeConnectionNotExists, err)
		return
	}

	s.updateSession(connectionID, func(s *ConnectionSession) {
		s.currentTable = tableName
	})

	// 先获取列信息，检查是否有单个整数主键
	columns, err := session.db.GetTableColumns(tableName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeGetTableColumnsFailed, err)
		return
	}

	// 检查是否有单个整数主键ID
	var primaryKeyColumn *database.ColumnInfo = nil
	primaryKeyCount := 0
	for i := range columns {
		if columns[i].Key == "PRI" {
			primaryKeyCount++
			primaryKeyColumn = &columns[i]
		}
	}

	// 判断是否可以使用基于ID的分页
	useIdBasedPagination := false
	var primaryKeyName string = ""
	if primaryKeyCount == 1 && primaryKeyColumn != nil {
		// 检查主键类型是否为整数类型
		typeLower := strings.ToLower(primaryKeyColumn.Type)
		if strings.Contains(typeLower, "int") || strings.Contains(typeLower, "serial") ||
			strings.Contains(typeLower, "bigint") || strings.Contains(typeLower, "smallint") ||
			strings.Contains(typeLower, "tinyint") || strings.Contains(typeLower, "mediumint") {
			useIdBasedPagination = true
			primaryKeyName = primaryKeyColumn.Name
		}
	}

	var data []map[string]interface{}
	var total int64
	var nextId interface{} = nil

	if useIdBasedPagination {
		// 使用基于ID的分页
		// direction: "next"表示下一页（id > lastId），"prev"表示上一页（id < lastId）
		data, total, nextId, err = session.db.GetTableDataByID(tableName, primaryKeyName, lastId, pageSize, direction)
		if err != nil {
			// 如果基于ID的分页失败，回退到传统分页
			data, total, err = session.db.GetTableData(tableName, page, pageSize)
			useIdBasedPagination = false
		}
	} else {
		// 使用传统OFFSET/LIMIT分页
		data, total, err = session.db.GetTableData(tableName, page, pageSize)
	}

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeGetTableDataFailed, err)
		return
	}

	// 如果使用基于ID的分页，从结果中提取ID
	if useIdBasedPagination && nextId == nil && len(data) > 0 {
		if direction == "prev" {
			// 上一页：nextId应该是当前页的第一个ID（用于继续向前翻页）
			if idVal, ok := data[0][primaryKeyName]; ok {
				nextId = idVal
			}
		} else {
			// 下一页：nextId应该是当前页的最后一个ID
			if idVal, ok := data[len(data)-1][primaryKeyName]; ok {
				nextId = idVal
			}
		}
	}

	// 检查是否还有下一页（基于ID分页时，如果返回的数据少于pageSize，说明没有下一页了）
	hasNextPage := true
	if useIdBasedPagination {
		hasNextPage = len(data) >= pageSize && nextId != nil
	}

	// 检查是否为 ClickHouse（不支持分页）
	isClickHouse := session.dbType == "clickhouse"

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"data":    data,
			"columns": columns,
		},
		"total":        total,
		"page":         page,
		"pageSize":     pageSize,
		"isClickHouse": isClickHouse,
	}

	// 如果使用基于ID的分页，添加相关信息
	if useIdBasedPagination {
		response["useIdPagination"] = true
		response["primaryKey"] = primaryKeyName
		if nextId != nil {
			response["nextId"] = nextId
		}
		response["hasNextPage"] = hasNextPage
		// 如果是上一页，还需要返回当前页的第一个ID（用于继续向前翻页）
		if direction == "prev" && len(data) > 0 {
			if firstId, ok := data[0][primaryKeyName]; ok {
				response["firstId"] = firstId
			}
		}
	}

	json.NewEncoder(w).Encode(response)
}

// GetPageId 根据页码获取该页的起始ID（用于基于ID分页的页码跳转）
func (s *Server) GetPageId(w http.ResponseWriter, r *http.Request) {
	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingConnectionID)
		return
	}

	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingTableName)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 50
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeConnectionNotExists, err)
		return
	}

	// 获取列信息，检查是否有单个整数主键
	columns, err := session.db.GetTableColumns(tableName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeGetTableColumnsFailed, err)
		return
	}

	// 检查是否有单个整数主键ID
	var primaryKeyColumn *database.ColumnInfo = nil
	primaryKeyCount := 0
	for i := range columns {
		if columns[i].Key == "PRI" {
			primaryKeyCount++
			primaryKeyColumn = &columns[i]
		}
	}

	// 判断是否可以使用基于ID的分页
	if primaryKeyCount != 1 || primaryKeyColumn == nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeNoSinglePrimaryKey)
		return
	}

	// 检查主键类型是否为整数类型
	typeLower := strings.ToLower(primaryKeyColumn.Type)
	if !strings.Contains(typeLower, "int") && !strings.Contains(typeLower, "serial") &&
		!strings.Contains(typeLower, "bigint") && !strings.Contains(typeLower, "smallint") &&
		!strings.Contains(typeLower, "tinyint") && !strings.Contains(typeLower, "mediumint") {
		writeJSONError(w, http.StatusBadRequest, ErrCodePrimaryKeyNotInteger)
		return
	}

	// 获取指定页码的ID
	pageId, err := session.db.GetPageIdByPageNumber(tableName, primaryKeyColumn.Name, page, pageSize)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeGetPageIDFailed, err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"pageId":  pageId,
		"page":    page,
	})
}

// ExecuteQuery 执行SQL查询
func (s *Server) ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, ErrCodeMethodNotAllowed)
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingConnectionID)
		return
	}

	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeParseRequestFailed, err)
		return
	}

	if req.Query == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeEmptySQLQuery)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeConnectionNotExists, err)
		return
	}

	// 判断SQL类型
	queryUpper := strings.ToUpper(strings.TrimSpace(req.Query))
	queryType := ""
	if len(queryUpper) >= 6 {
		queryType = queryUpper[:6]
	}

	// 执行SQL校验
	if err := s.validateSQL(req.Query, queryType); err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeSQLValidationFailed, err)
		return
	}

	// 判断SQL类型（兼容旧代码）
	queryUpperPrefix := fmt.Sprintf("%.6s", req.Query)
	if queryType == "SELECT" || queryUpperPrefix == "SELECT" || queryUpperPrefix == "select" {
		results, err := session.db.ExecuteQuery(req.Query)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, ErrCodeExecuteQueryFailed, err)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    results,
		})
	} else if queryType == "UPDATE" || queryUpperPrefix == "UPDATE" || queryUpperPrefix == "update" {
		affected, err := session.db.ExecuteUpdate(req.Query)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, ErrCodeExecuteUpdateFailed, err)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"affected": affected,
		})
	} else if queryType == "DELETE" || queryUpperPrefix == "DELETE" || queryUpperPrefix == "delete" {
		affected, err := session.db.ExecuteDelete(req.Query)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, ErrCodeExecuteDeleteFailed, err)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"affected": affected,
		})
	} else if queryType == "INSERT" || queryUpperPrefix == "INSERT" || queryUpperPrefix == "insert" {
		affected, err := session.db.ExecuteInsert(req.Query)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, ErrCodeExecuteInsertFailed, err)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"affected": affected,
		})
	} else {
		writeJSONError(w, http.StatusBadRequest, ErrCodeUnsupportedSQLType)
	}
}

// UpdateRow 更新行数据
func (s *Server) UpdateRow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, ErrCodeMethodNotAllowed)
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingConnectionID)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeConnectionNotExists, err)
		return
	}

	// ClickHouse 不支持 UPDATE 操作
	if session.dbType == "clickhouse" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeClickHouseNoUpdate)
		return
	}

	var req struct {
		Table string                 `json:"table"`
		Data  map[string]interface{} `json:"data"`
		Where map[string]interface{} `json:"where"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeParseRequestFailed, err)
		return
	}

	if req.Table == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeTableNameEmpty)
		return
	}

	// 构建UPDATE SQL
	query := fmt.Sprintf("UPDATE `%s` SET ", req.Table)
	first := true
	for k, v := range req.Data {
		if !first {
			query += ", "
		}
		if v == nil {
			query += fmt.Sprintf("`%s` = NULL", k)
		} else {
			// 转义单引号防止SQL注入
			valStr := fmt.Sprintf("%v", v)
			valStr = strings.ReplaceAll(valStr, "'", "''")
			query += fmt.Sprintf("`%s` = '%s'", k, valStr)
		}
		first = false
	}

	query += " WHERE "
	first = true
	for k, v := range req.Where {
		if !first {
			query += " AND "
		}
		if v == nil {
			query += fmt.Sprintf("`%s` IS NULL", k)
		} else {
			// 转义单引号防止SQL注入
			valStr := fmt.Sprintf("%v", v)
			valStr = strings.ReplaceAll(valStr, "'", "''")
			query += fmt.Sprintf("`%s` = '%s'", k, valStr)
		}
		first = false
	}

	affected, err := session.db.ExecuteUpdate(query)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeUpdateFailed, err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"affected": affected,
	})
}

// DeleteRow 删除行数据
func (s *Server) DeleteRow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, ErrCodeMethodNotAllowed)
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingConnectionID)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeConnectionNotExists, err)
		return
	}

	// ClickHouse 不支持 DELETE 操作
	if session.dbType == "clickhouse" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeClickHouseNoDelete)
		return
	}

	var req struct {
		Table string                 `json:"table"`
		Where map[string]interface{} `json:"where"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeParseRequestFailed, err)
		return
	}

	if req.Table == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeTableNameEmpty)
		return
	}

	// 构建DELETE SQL
	query := fmt.Sprintf("DELETE FROM `%s` WHERE ", req.Table)
	first := true
	for k, v := range req.Where {
		if !first {
			query += " AND "
		}
		if v == nil {
			query += fmt.Sprintf("`%s` IS NULL", k)
		} else {
			// 转义单引号防止SQL注入
			valStr := fmt.Sprintf("%v", v)
			valStr = strings.ReplaceAll(valStr, "'", "''")
			query += fmt.Sprintf("`%s` = '%s'", k, valStr)
		}
		first = false
	}

	affected, err := session.db.ExecuteDelete(query)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeDeleteFailed, err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"affected": affected,
	})
}

// GetDatabases 获取数据库列表
func (s *Server) GetDatabases(w http.ResponseWriter, r *http.Request) {
	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingConnectionID)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeConnectionNotExists, err)
		return
	}

	databases, err := session.db.GetDatabases()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeGetDatabasesFailed, err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"databases": databases,
	})
}

// SwitchDatabase 切换数据库
func (s *Server) SwitchDatabase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, ErrCodeMethodNotAllowed)
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingConnectionID)
		return
	}

	var req struct {
		Database string `json:"database"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeParseRequestFailed, err)
		return
	}

	if req.Database == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingDatabaseName)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeConnectionNotExists, err)
		return
	}

	if err := session.db.SwitchDatabase(req.Database); err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeSwitchDatabaseFailed, err)
		return
	}

	s.updateSession(connectionID, func(s *ConnectionSession) {
		s.currentDatabase = req.Database
		s.currentTable = "" // 切换数据库时清空当前表
	})

	// 切换数据库后重新加载表列表
	tables, err := session.db.GetTables()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeGetTablesFailed, err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "切换数据库成功",
		"tables":  tables,
	})
}

// Disconnect 断开连接
func (s *Server) Disconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, ErrCodeMethodNotAllowed)
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingConnectionID)
		return
	}

	s.sessionsMutex.Lock()
	session, exists := s.sessions[connectionID]
	if exists {
		if session.db != nil {
			session.db.Close()
		}
		delete(s.sessions, connectionID)
	}
	s.sessionsMutex.Unlock()

	// 从持久化存储删除
	if exists {
		if err := s.sessionStorage.Delete(connectionID); err != nil {
			s.getLogger().Warn(r.Context(), "Failed to delete session from persistent storage: %v", err)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "已断开连接",
	})
}

// GetStatus 获取连接状态
func (s *Server) GetStatus(w http.ResponseWriter, r *http.Request) {
	connectionID := getConnectionID(r)
	if connectionID == "" {
		// 如果没有连接ID，返回未连接状态
		json.NewEncoder(w).Encode(map[string]interface{}{
			"connected": false,
		})
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"connected": false,
		})
		return
	}

	s.sessionsMutex.RLock()
	currentDatabase := session.currentDatabase
	currentTable := session.currentTable
	dbType := session.dbType
	s.sessionsMutex.RUnlock()

	// 获取数据库列表
	databases, err := session.db.GetDatabases()
	response := map[string]interface{}{
		"connected": true,
		"dbType":    dbType,
	}
	if err == nil {
		response["databases"] = databases
	}
	response["currentDatabase"] = currentDatabase
	response["currentTable"] = currentTable

	json.NewEncoder(w).Encode(response)
}

// SetupRoutes 设置路由（使用标准库，保持向后兼容）
func (s *Server) SetupRoutes() {
	router := NewStandardRouter()
	s.RegisterRoutes(router)
}

// RegisterRoutes 注册路由到指定的路由适配器
// 这个方法支持适配不同的 Web 框架（Gin、Echo、Fiber 等）
// 如果 router 设置了前缀（通过 SetPrefix 或 NewPrefixRouter），所有路由会自动添加前缀
func (s *Server) RegisterRoutes(router Router) {
	// 首页
	router.HandleFunc("/", s.Home)

	// API 路由
	router.POST("/api/connect", s.Connect)
	router.POST("/api/disconnect", s.Disconnect)
	router.HandleFunc("/api/status", s.GetStatus)
	router.HandleFunc("/api/databases", s.GetDatabases)
	router.POST("/api/database/switch", s.SwitchDatabase)
	router.HandleFunc("/api/tables", s.GetTables)
	router.HandleFunc("/api/table/schema", s.GetTableSchema)
	router.HandleFunc("/api/table/columns", s.GetTableColumns)
	router.HandleFunc("/api/table/data", s.GetTableData)
	router.HandleFunc("/api/table/page-id", s.GetPageId)
	router.HandleFunc("/api/table/export", s.ExportTableDataToExcel)
	router.POST("/api/query", s.ExecuteQuery)
	router.POST("/api/query/export", s.ExportQueryResultsToExcel)
	router.POST("/api/row/update", s.UpdateRow)
	router.POST("/api/row/delete", s.DeleteRow)

	// 静态文件 - 使用 embed.FS
	router.StaticFS("/static/", staticFS)

	// 获取数据库类型列表
	router.HandleFunc("/api/database/types", s.GetDatabaseTypes)
	
	// 获取预设连接列表
	router.HandleFunc("/api/preset-connections", s.GetPresetConnectionsAPI)
}

// Start 启动服务器
func (s *Server) Start(addr string) error {
	s.getLogger().Info(context.Background(), "Server starting on %s", addr)
	return http.ListenAndServe(addr, nil)
}
