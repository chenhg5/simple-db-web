package handlers

import (
	"crypto/rand"
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
	templates         *template.Template
	sessionStorage    SessionStorage                // 会话存储接口
	sessions          map[string]*ConnectionSession // 运行时会话缓存（包含实际连接）
	sessionsMutex     sync.RWMutex
	customDatabases   map[string]DatabaseFactory // 自定义数据库类型
	customDbMutex     sync.RWMutex
	customProxies     map[string]ProxyFactory // 自定义代理类型
	customProxyMutex  sync.RWMutex
	builtinTypes      map[string]string // 内置数据库类型及其显示名称
	customScript      string            // 自定义JavaScript脚本，会在页面加载后执行
	customScriptMutex sync.RWMutex      // 保护customScript的读写锁
	validators        []SQLValidator    // SQL校验器列表
	validatorsMutex   sync.RWMutex      // 保护validators的读写锁
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
		"postgres":   "PostgreSQL",
		"postgresql": "PostgreSQL",
		"sqlite":     "SQLite",
		"clickhouse": "ClickHouse",
		"dameng":     "达梦",
		"openguass":  "OpenGauss",
		"vastbase":   "Vastbase",
		"kingbase":   "人大金仓",
		"oceandb":    "OceanDB",
	}

	server := &Server{
		templates:       tmpl,
		sessionStorage:  NewMemorySessionStorage(), // 默认使用内存存储
		sessions:        make(map[string]*ConnectionSession),
		customDatabases: make(map[string]DatabaseFactory),
		customProxies:   make(map[string]ProxyFactory),
		builtinTypes:    builtinTypes,
		validators:      make([]SQLValidator, 0),
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
			// 使用类型名作为显示名称，或者可以扩展为支持自定义显示名称
			types = append(types, DatabaseTypeInfo{
				Type:        dbType,
				DisplayName: dbType, // 可以后续扩展支持自定义显示名称
			})
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"types":   types,
	})
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

// writeJSONError 写入JSON格式的错误响应
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
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
		case "dameng":
			db = database.NewBaseMysqlBasedDB("dameng")
		case "openguass":
			db = database.NewBaseMysqlBasedDB("openguass")
		case "vastbase":
			db = database.NewBaseMysqlBasedDB("vastbase")
		case "kingbase":
			db = database.NewBaseMysqlBasedDB("kingbase")
		case "oceandb":
			db = database.NewBaseMysqlBasedDB("oceandb")
		case "sqlite":
			db = database.NewSQLite3()
		case "postgres", "postgresql":
			db = database.NewPostgreSQL()
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
		if data.DbType == "mysql" || data.DbType == "dameng" || data.DbType == "openguass" ||
			data.DbType == "vastbase" || data.DbType == "kingbase" || data.DbType == "oceandb" {
			// MySQL及其兼容数据库使用代理包装器
			db = NewProxyDatabaseWrapper(db, proxy)
		} else {
			// 其他数据库类型暂不支持代理，记录警告并关闭代理
			log.Printf("警告: 数据库类型 %s 暂不支持代理连接，将尝试直接连接", data.DbType)
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
			log.Printf("警告: 更新会话到持久化存储失败: %v", err)
			// 不返回错误，因为内存中的会话已经更新
		}
	}

	return nil
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
		writeJSONError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var info database.ConnectionInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("解析请求失败: %v", err))
		return
	}

	// 生成连接ID
	connectionID, err := generateConnectionID()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("生成连接ID失败: %v", err))
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
		case "dameng":
			db = database.NewBaseMysqlBasedDB("dameng")
		case "openguass":
			db = database.NewBaseMysqlBasedDB("openguass")
		case "vastbase":
			db = database.NewBaseMysqlBasedDB("vastbase")
		case "kingbase":
			db = database.NewBaseMysqlBasedDB("kingbase")
		case "oceandb":
			db = database.NewBaseMysqlBasedDB("oceandb")
		case "sqlite":
			db = database.NewSQLite3()
		case "postgres", "postgresql":
			db = database.NewPostgreSQL()
		default:
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("不支持的数据库类型: %s", info.Type))
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
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("不支持的代理类型: %s", info.Proxy.Type))
			return
		}

		// 构建代理配置JSON
		proxyConfigJSON, err := json.Marshal(info.Proxy)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("构建代理配置失败: %v", err))
			return
		}

		// 创建代理
		proxy, err = proxyFactory(string(proxyConfigJSON))
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("建立代理连接失败: %v", err))
			return
		}
		defer func() {
			// 注意：代理连接会在会话关闭时关闭，这里不立即关闭
		}()
	}

	// 构建DSN
	var dsn string
	if info.Type == "clickhouse" {
		dsn = database.BuildClickHouseDSN(info)
	} else {
		dsn = database.BuildDSN(info)
	}

	// 如果有代理，使用代理包装器
	if proxy != nil {
		// 目前只支持MySQL及其兼容数据库的代理连接
		// 其他数据库类型的代理支持需要进一步实现
		if info.Type == "mysql" || info.Type == "dameng" || info.Type == "openguass" ||
			info.Type == "vastbase" || info.Type == "kingbase" || info.Type == "oceandb" {
			// MySQL及其兼容数据库使用代理包装器
			db = NewProxyDatabaseWrapper(db, proxy)
		} else {
			// 其他数据库类型暂不支持代理，记录警告
			log.Printf("警告: 数据库类型 %s 暂不支持代理连接，将尝试直接连接", info.Type)
			proxy.Close()
			proxy = nil
		}
	}

	if err := db.Connect(dsn); err != nil {
		if proxy != nil {
			proxy.Close()
		}
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("连接失败: %v", err))
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
		log.Printf("警告: 保存会话到持久化存储失败: %v", err)
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
		writeJSONError(w, http.StatusBadRequest, "缺少连接ID")
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	tables, err := session.db.GetTables()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("获取表列表失败: %v", err))
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
		writeJSONError(w, http.StatusBadRequest, "缺少连接ID")
		return
	}

	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		writeJSONError(w, http.StatusBadRequest, "缺少表名参数")
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.updateSession(connectionID, func(s *ConnectionSession) {
		s.currentTable = tableName
	})
	schema, err := session.db.GetTableSchema(tableName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("获取表结构失败: %v", err))
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
		writeJSONError(w, http.StatusBadRequest, "缺少连接ID")
		return
	}

	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		writeJSONError(w, http.StatusBadRequest, "缺少表名参数")
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 确保数据库已选择（从持久化存储获取的会话应该已经切换了数据库，但为了安全再次检查）
	if session.currentDatabase == "" {
		writeJSONError(w, http.StatusBadRequest, "请先选择数据库")
		return
	}

	columns, err := session.db.GetTableColumns(tableName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("获取列信息失败: %v", err))
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
		writeJSONError(w, http.StatusBadRequest, "缺少连接ID")
		return
	}

	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		writeJSONError(w, http.StatusBadRequest, "缺少表名参数")
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
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.updateSession(connectionID, func(s *ConnectionSession) {
		s.currentTable = tableName
	})

	data, total, err := session.db.GetTableData(tableName, page, pageSize)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("获取数据失败: %v", err))
		return
	}
	columns, err := session.db.GetTableColumns(tableName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("获取列信息失败: %v", err))
		return
	}

	// 检查是否为 ClickHouse（不支持分页）
	isClickHouse := session.dbType == "clickhouse"

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"data":    data,
			"columns": columns,
		},
		"total":        total,
		"page":         page,
		"pageSize":     pageSize,
		"isClickHouse": isClickHouse, // 标识是否为 ClickHouse
	})
}

// ExecuteQuery 执行SQL查询
func (s *Server) ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, "缺少连接ID")
		return
	}

	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("解析请求失败: %v", err))
		return
	}

	if req.Query == "" {
		writeJSONError(w, http.StatusBadRequest, "SQL查询不能为空")
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
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
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("SQL校验失败: %v", err))
		return
	}

	// 判断SQL类型（兼容旧代码）
	queryUpperPrefix := fmt.Sprintf("%.6s", req.Query)
	if queryType == "SELECT" || queryUpperPrefix == "SELECT" || queryUpperPrefix == "select" {
		results, err := session.db.ExecuteQuery(req.Query)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("执行查询失败: %v", err))
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    results,
		})
	} else if queryType == "UPDATE" || queryUpperPrefix == "UPDATE" || queryUpperPrefix == "update" {
		affected, err := session.db.ExecuteUpdate(req.Query)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("执行更新失败: %v", err))
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"affected": affected,
		})
	} else if queryType == "DELETE" || queryUpperPrefix == "DELETE" || queryUpperPrefix == "delete" {
		affected, err := session.db.ExecuteDelete(req.Query)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("执行删除失败: %v", err))
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"affected": affected,
		})
	} else if queryType == "INSERT" || queryUpperPrefix == "INSERT" || queryUpperPrefix == "insert" {
		affected, err := session.db.ExecuteInsert(req.Query)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("执行插入失败: %v", err))
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"affected": affected,
		})
	} else {
		writeJSONError(w, http.StatusBadRequest, "不支持的SQL类型")
	}
}

// UpdateRow 更新行数据
func (s *Server) UpdateRow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, "缺少连接ID")
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// ClickHouse 不支持 UPDATE 操作
	if session.dbType == "clickhouse" {
		writeJSONError(w, http.StatusBadRequest, "ClickHouse 不支持 UPDATE 操作")
		return
	}

	var req struct {
		Table string                 `json:"table"`
		Data  map[string]interface{} `json:"data"`
		Where map[string]interface{} `json:"where"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("解析请求失败: %v", err))
		return
	}

	if req.Table == "" {
		writeJSONError(w, http.StatusBadRequest, "表名不能为空")
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
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("更新失败: %v", err))
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
		writeJSONError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, "缺少连接ID")
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// ClickHouse 不支持 DELETE 操作
	if session.dbType == "clickhouse" {
		writeJSONError(w, http.StatusBadRequest, "ClickHouse 不支持 DELETE 操作")
		return
	}

	var req struct {
		Table string                 `json:"table"`
		Where map[string]interface{} `json:"where"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("解析请求失败: %v", err))
		return
	}

	if req.Table == "" {
		writeJSONError(w, http.StatusBadRequest, "表名不能为空")
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
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("删除失败: %v", err))
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
		writeJSONError(w, http.StatusBadRequest, "缺少连接ID")
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	databases, err := session.db.GetDatabases()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("获取数据库列表失败: %v", err))
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
		writeJSONError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, "缺少连接ID")
		return
	}

	var req struct {
		Database string `json:"database"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("解析请求失败: %v", err))
		return
	}

	if req.Database == "" {
		writeJSONError(w, http.StatusBadRequest, "数据库名不能为空")
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := session.db.SwitchDatabase(req.Database); err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("切换数据库失败: %v", err))
		return
	}

	s.updateSession(connectionID, func(s *ConnectionSession) {
		s.currentDatabase = req.Database
		s.currentTable = "" // 切换数据库时清空当前表
	})

	// 切换数据库后重新加载表列表
	tables, err := session.db.GetTables()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("获取表列表失败: %v", err))
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
		writeJSONError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, "缺少连接ID")
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
			log.Printf("警告: 从持久化存储删除会话失败: %v", err)
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
	router.HandleFunc("/api/table/export", s.ExportTableDataToExcel)
	router.POST("/api/query", s.ExecuteQuery)
	router.POST("/api/query/export", s.ExportQueryResultsToExcel)
	router.POST("/api/row/update", s.UpdateRow)
	router.POST("/api/row/delete", s.DeleteRow)

	// 静态文件 - 使用 embed.FS
	router.StaticFS("/static/", staticFS)

	// 获取数据库类型列表
	router.HandleFunc("/api/database/types", s.GetDatabaseTypes)
}

// Start 启动服务器
func (s *Server) Start(addr string) error {
	log.Printf("服务器启动在 %s", addr)
	return http.ListenAndServe(addr, nil)
}
