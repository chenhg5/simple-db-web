package handlers

import (
	"dbweb/database"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Server 服务器结构
type Server struct {
	templates *template.Template
	db        database.Database
	dbMutex   sync.RWMutex
}

// NewServer 创建新的服务器实例
func NewServer() (*Server, error) {
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("加载模板失败: %w", err)
	}

	return &Server{
		templates: tmpl,
	}, nil
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

	s.dbMutex.Lock()
	defer s.dbMutex.Unlock()

	// 关闭旧连接
	if s.db != nil {
		s.db.Close()
	}

	// 创建新连接
	var db database.Database
	switch info.Type {
	case "mysql":
		db = database.NewMySQL()
	case "dameng":
		db = database.NewKingsoftDB("dameng")
	default:
		http.Error(w, "不支持的数据库类型", http.StatusBadRequest)
		return
	}

	// 构建DSN
	dsn := database.BuildDSN(info)
	if err := db.Connect(dsn); err != nil {
		http.Error(w, fmt.Sprintf("连接失败: %v", err), http.StatusInternalServerError)
		return
	}

	s.db = db

	// 获取数据库列表
	databases, err := db.GetDatabases()
	if err != nil {
		// 如果获取数据库列表失败,仍然返回连接成功,但数据库列表为空
		databases = []string{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "连接成功",
		"databases": databases,
	})
}

// GetTables 获取表列表
func (s *Server) GetTables(w http.ResponseWriter, r *http.Request) {
	s.dbMutex.RLock()
	db := s.db
	s.dbMutex.RUnlock()

	if db == nil {
		http.Error(w, "未连接数据库", http.StatusBadRequest)
		return
	}

	tables, err := db.GetTables()
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
	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		http.Error(w, "缺少表名参数", http.StatusBadRequest)
		return
	}

	s.dbMutex.RLock()
	db := s.db
	s.dbMutex.RUnlock()

	if db == nil {
		http.Error(w, "未连接数据库", http.StatusBadRequest)
		return
	}

	schema, err := db.GetTableSchema(tableName)
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
	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		http.Error(w, "缺少表名参数", http.StatusBadRequest)
		return
	}

	s.dbMutex.RLock()
	db := s.db
	s.dbMutex.RUnlock()

	if db == nil {
		http.Error(w, "未连接数据库", http.StatusBadRequest)
		return
	}

	columns, err := db.GetTableColumns(tableName)
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

	s.dbMutex.RLock()
	db := s.db
	s.dbMutex.RUnlock()

	if db == nil {
		http.Error(w, "未连接数据库", http.StatusBadRequest)
		return
	}

	data, total, err := db.GetTableData(tableName, page, pageSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("获取数据失败: %v", err), http.StatusInternalServerError)
		return
	}
	columns, err := db.GetTableColumns(tableName)
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

	s.dbMutex.RLock()
	db := s.db
	s.dbMutex.RUnlock()

	if db == nil {
		http.Error(w, "未连接数据库", http.StatusBadRequest)
		return
	}

	// 判断SQL类型
	queryUpper := fmt.Sprintf("%.6s", req.Query)
	if queryUpper == "SELECT" || queryUpper == "select" {
		results, err := db.ExecuteQuery(req.Query)
		if err != nil {
			http.Error(w, fmt.Sprintf("执行查询失败: %v", err), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    results,
		})
	} else if queryUpper == "UPDATE" || queryUpper == "update" {
		affected, err := db.ExecuteUpdate(req.Query)
		if err != nil {
			http.Error(w, fmt.Sprintf("执行更新失败: %v", err), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"affected": affected,
		})
	} else if queryUpper == "DELETE" || queryUpper == "delete" {
		affected, err := db.ExecuteDelete(req.Query)
		if err != nil {
			http.Error(w, fmt.Sprintf("执行删除失败: %v", err), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"affected": affected,
		})
	} else if queryUpper == "INSERT" || queryUpper == "insert" {
		affected, err := db.ExecuteInsert(req.Query)
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

	s.dbMutex.RLock()
	db := s.db
	s.dbMutex.RUnlock()

	if db == nil {
		http.Error(w, "未连接数据库", http.StatusBadRequest)
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

	affected, err := db.ExecuteUpdate(query)
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

	s.dbMutex.RLock()
	db := s.db
	s.dbMutex.RUnlock()

	if db == nil {
		http.Error(w, "未连接数据库", http.StatusBadRequest)
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

	affected, err := db.ExecuteDelete(query)
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
	s.dbMutex.RLock()
	db := s.db
	s.dbMutex.RUnlock()

	if db == nil {
		http.Error(w, "未连接数据库", http.StatusBadRequest)
		return
	}

	databases, err := db.GetDatabases()
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

	s.dbMutex.RLock()
	db := s.db
	s.dbMutex.RUnlock()

	if db == nil {
		http.Error(w, "未连接数据库", http.StatusBadRequest)
		return
	}

	if err := db.SwitchDatabase(req.Database); err != nil {
		http.Error(w, fmt.Sprintf("切换数据库失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 切换数据库后重新加载表列表
	tables, err := db.GetTables()
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

	s.dbMutex.Lock()
	defer s.dbMutex.Unlock()

	if s.db != nil {
		s.db.Close()
		s.db = nil
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "已断开连接",
	})
}

// SetupRoutes 设置路由
func (s *Server) SetupRoutes() {
	http.HandleFunc("/", s.Home)
	http.HandleFunc("/api/connect", s.Connect)
	http.HandleFunc("/api/disconnect", s.Disconnect)
	http.HandleFunc("/api/databases", s.GetDatabases)
	http.HandleFunc("/api/database/switch", s.SwitchDatabase)
	http.HandleFunc("/api/tables", s.GetTables)
	http.HandleFunc("/api/table/schema", s.GetTableSchema)
	http.HandleFunc("/api/table/columns", s.GetTableColumns)
	http.HandleFunc("/api/table/data", s.GetTableData)
	http.HandleFunc("/api/query", s.ExecuteQuery)
	http.HandleFunc("/api/row/update", s.UpdateRow)
	http.HandleFunc("/api/row/delete", s.DeleteRow)

	// 静态文件
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}

// Start 启动服务器
func (s *Server) Start(addr string) error {
	log.Printf("服务器启动在 %s", addr)
	return http.ListenAndServe(addr, nil)
}
