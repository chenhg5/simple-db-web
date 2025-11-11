package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chenhg5/simple-db-web/database"
)

// ConnectionSession 连接会话信息
type ConnectionSession struct {
	db              database.Database
	currentDatabase string
	currentTable    string
	createdAt       time.Time
}

// DatabaseFactory 数据库工厂函数类型
type DatabaseFactory func() database.Database

// DatabaseTypeInfo 数据库类型信息
type DatabaseTypeInfo struct {
	Type        string `json:"type"`         // 数据库类型标识
	DisplayName string `json:"display_name"` // 显示名称
}

// Server 服务器结构
type Server struct {
	templates       *template.Template
	sessions        map[string]*ConnectionSession
	sessionsMutex   sync.RWMutex
	customDatabases map[string]DatabaseFactory // 自定义数据库类型
	customDbMutex   sync.RWMutex
	builtinTypes    map[string]string // 内置数据库类型及其显示名称
}

// NewServer 创建新的服务器实例
func NewServer() (*Server, error) {
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("加载模板失败: %w", err)
	}

	// 初始化内置数据库类型
	builtinTypes := map[string]string{
		"mysql":      "MySQL",
		"postgres":   "PostgreSQL",
		"postgresql": "PostgreSQL",
		"sqlite":     "SQLite",
		"dameng":     "达梦",
		"openguass":  "OpenGauss",
		"vastbase":   "Vastbase",
		"kingbase":   "人大金仓",
		"oceandb":    "OceanDB",
	}

	return &Server{
		templates:       tmpl,
		sessions:        make(map[string]*ConnectionSession),
		customDatabases: make(map[string]DatabaseFactory),
		builtinTypes:    builtinTypes,
	}, nil
}

// AddDatabase 添加自定义数据库类型
// name: 数据库类型标识（如 "custom_db"）
// factory: 创建数据库实例的工厂函数
func (s *Server) AddDatabase(name string, factory DatabaseFactory) {
	s.customDbMutex.Lock()
	defer s.customDbMutex.Unlock()
	s.customDatabases[name] = factory
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

// getSession 根据连接ID获取会话
func (s *Server) getSession(connectionID string) (*ConnectionSession, error) {
	s.sessionsMutex.RLock()
	defer s.sessionsMutex.RUnlock()

	session, exists := s.sessions[connectionID]
	if !exists {
		return nil, fmt.Errorf("连接不存在或已断开")
	}
	return session, nil
}

// Home 首页
func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
	if err := s.templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Connect 连接数据库
func (s *Server) Connect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var info database.ConnectionInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
		return
	}

	// 生成连接ID
	connectionID, err := generateConnectionID()
	if err != nil {
		http.Error(w, fmt.Sprintf("生成连接ID失败: %v", err), http.StatusInternalServerError)
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
			http.Error(w, fmt.Sprintf("不支持的数据库类型: %s", info.Type), http.StatusBadRequest)
			return
		}
	}

	// 构建DSN
	dsn := database.BuildDSN(info)
	if err := db.Connect(dsn); err != nil {
		http.Error(w, fmt.Sprintf("连接失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 创建会话
	session := &ConnectionSession{
		db:              db,
		currentDatabase: "",
		currentTable:    "",
		createdAt:       time.Now(),
	}

	// 保存会话
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
		http.Error(w, "缺少连接ID", http.StatusBadRequest)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tables, err := session.db.GetTables()
	if err != nil {
		http.Error(w, fmt.Sprintf("获取表列表失败: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "缺少连接ID", http.StatusBadRequest)
		return
	}

	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		http.Error(w, "缺少表名参数", http.StatusBadRequest)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.sessionsMutex.Lock()
	session.currentTable = tableName
	s.sessionsMutex.Unlock()

	schema, err := session.db.GetTableSchema(tableName)
	if err != nil {
		http.Error(w, fmt.Sprintf("获取表结构失败: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "缺少连接ID", http.StatusBadRequest)
		return
	}

	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		http.Error(w, "缺少表名参数", http.StatusBadRequest)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	columns, err := session.db.GetTableColumns(tableName)
	if err != nil {
		http.Error(w, fmt.Sprintf("获取列信息失败: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "缺少连接ID", http.StatusBadRequest)
		return
	}

	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		http.Error(w, "缺少表名参数", http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.sessionsMutex.Lock()
	session.currentTable = tableName
	s.sessionsMutex.Unlock()

	data, total, err := session.db.GetTableData(tableName, page, pageSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("获取数据失败: %v", err), http.StatusInternalServerError)
		return
	}
	columns, err := session.db.GetTableColumns(tableName)
	if err != nil {
		http.Error(w, fmt.Sprintf("获取列信息失败: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"data":    data,
			"columns": columns,
		},
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// ExecuteQuery 执行SQL查询
func (s *Server) ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		http.Error(w, "缺少连接ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "SQL查询不能为空", http.StatusBadRequest)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 判断SQL类型
	queryUpper := fmt.Sprintf("%.6s", req.Query)
	if queryUpper == "SELECT" || queryUpper == "select" {
		results, err := session.db.ExecuteQuery(req.Query)
		if err != nil {
			http.Error(w, fmt.Sprintf("执行查询失败: %v", err), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    results,
		})
	} else if queryUpper == "UPDATE" || queryUpper == "update" {
		affected, err := session.db.ExecuteUpdate(req.Query)
		if err != nil {
			http.Error(w, fmt.Sprintf("执行更新失败: %v", err), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"affected": affected,
		})
	} else if queryUpper == "DELETE" || queryUpper == "delete" {
		affected, err := session.db.ExecuteDelete(req.Query)
		if err != nil {
			http.Error(w, fmt.Sprintf("执行删除失败: %v", err), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"affected": affected,
		})
	} else if queryUpper == "INSERT" || queryUpper == "insert" {
		affected, err := session.db.ExecuteInsert(req.Query)
		if err != nil {
			http.Error(w, fmt.Sprintf("执行插入失败: %v", err), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"affected": affected,
		})
	} else {
		http.Error(w, "不支持的SQL类型", http.StatusBadRequest)
	}
}

// UpdateRow 更新行数据
func (s *Server) UpdateRow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		http.Error(w, "缺少连接ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Table string                 `json:"table"`
		Data  map[string]interface{} `json:"data"`
		Where map[string]interface{} `json:"where"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
		return
	}

	if req.Table == "" {
		http.Error(w, "表名不能为空", http.StatusBadRequest)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("更新失败: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		http.Error(w, "缺少连接ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Table string                 `json:"table"`
		Where map[string]interface{} `json:"where"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
		return
	}

	if req.Table == "" {
		http.Error(w, "表名不能为空", http.StatusBadRequest)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("删除失败: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "缺少连接ID", http.StatusBadRequest)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	databases, err := session.db.GetDatabases()
	if err != nil {
		http.Error(w, fmt.Sprintf("获取数据库列表失败: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		http.Error(w, "缺少连接ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Database string `json:"database"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
		return
	}

	if req.Database == "" {
		http.Error(w, "数据库名不能为空", http.StatusBadRequest)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := session.db.SwitchDatabase(req.Database); err != nil {
		http.Error(w, fmt.Sprintf("切换数据库失败: %v", err), http.StatusInternalServerError)
		return
	}

	s.sessionsMutex.Lock()
	session.currentDatabase = req.Database
	session.currentTable = "" // 切换数据库时清空当前表
	s.sessionsMutex.Unlock()

	// 切换数据库后重新加载表列表
	tables, err := session.db.GetTables()
	if err != nil {
		http.Error(w, fmt.Sprintf("获取表列表失败: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		http.Error(w, "缺少连接ID", http.StatusBadRequest)
		return
	}

	s.sessionsMutex.Lock()
	defer s.sessionsMutex.Unlock()

	session, exists := s.sessions[connectionID]
	if exists {
		if session.db != nil {
			session.db.Close()
		}
		delete(s.sessions, connectionID)
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
	s.sessionsMutex.RUnlock()

	// 获取数据库列表
	databases, err := session.db.GetDatabases()
	response := map[string]interface{}{
		"connected": true,
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
	router.POST("/api/query", s.ExecuteQuery)
	router.POST("/api/row/update", s.UpdateRow)
	router.POST("/api/row/delete", s.DeleteRow)

	// 静态文件
	router.Static("/static/", "static")

	// 获取数据库类型列表
	router.HandleFunc("/api/database/types", s.GetDatabaseTypes)
}

// Start 启动服务器
func (s *Server) Start(addr string) error {
	log.Printf("服务器启动在 %s", addr)
	return http.ListenAndServe(addr, nil)
}
