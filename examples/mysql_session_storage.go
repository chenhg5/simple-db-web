//go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chenhg5/simple-db-web/handlers"
	_ "github.com/go-sql-driver/mysql"
)

// MySQLSessionStorage MySQL会话存储实现
// 适用于多实例部署，通过MySQL共享会话数据
type MySQLSessionStorage struct {
	db        *sql.DB
	tableName string // 表名，默认为 "simpledb_sessions"
}

// NewMySQLSessionStorage 创建MySQL会话存储
// dsn: MySQL连接字符串，如 "user:password@tcp(localhost:3306)/dbname"
func NewMySQLSessionStorage(dsn string) (*MySQLSessionStorage, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	storage := &MySQLSessionStorage{
		db:        db,
		tableName: "simpledb_sessions",
	}

	// 创建表（如果不存在）
	if err := storage.createTable(); err != nil {
		return nil, fmt.Errorf("创建表失败: %w", err)
	}

	return storage, nil
}

// createTable 创建会话表
func (m *MySQLSessionStorage) createTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			connection_id VARCHAR(64) PRIMARY KEY,
			session_data TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NULL,
			INDEX idx_expires_at (expires_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`, m.tableName)

	_, err := m.db.Exec(query)
	return err
}

// Get 获取会话数据
func (m *MySQLSessionStorage) Get(connectionID string) (*handlers.SessionData, error) {
	query := fmt.Sprintf(`
		SELECT session_data, expires_at 
		FROM %s 
		WHERE connection_id = ? 
		AND (expires_at IS NULL OR expires_at > NOW())
	`, m.tableName)

	var sessionDataJSON string
	var expiresAt sql.NullTime
	err := m.db.QueryRow(query, connectionID).Scan(&sessionDataJSON, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("会话不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询会话失败: %w", err)
	}

	var data handlers.SessionData
	if err := json.Unmarshal([]byte(sessionDataJSON), &data); err != nil {
		return nil, fmt.Errorf("解析会话数据失败: %w", err)
	}

	return &data, nil
}

// Set 保存会话数据
func (m *MySQLSessionStorage) Set(connectionID string, data *handlers.SessionData, ttl time.Duration) error {
	sessionDataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化会话数据失败: %w", err)
	}

	var expiresAt interface{}
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	} else {
		expiresAt = nil
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (connection_id, session_data, expires_at)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			session_data = VALUES(session_data),
			expires_at = VALUES(expires_at)
	`, m.tableName)

	_, err = m.db.Exec(query, connectionID, sessionDataJSON, expiresAt)
	if err != nil {
		return fmt.Errorf("保存会话失败: %w", err)
	}

	return nil
}

// Delete 删除会话数据
func (m *MySQLSessionStorage) Delete(connectionID string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE connection_id = ?", m.tableName)
	_, err := m.db.Exec(query, connectionID)
	if err != nil {
		return fmt.Errorf("删除会话失败: %w", err)
	}
	return nil
}

// Close 关闭数据库连接
func (m *MySQLSessionStorage) Close() error {
	return m.db.Close()
}

// SetTableName 设置表名
func (m *MySQLSessionStorage) SetTableName(tableName string) {
	m.tableName = tableName
}

// CleanupExpiredSessions 清理过期会话（可以定期调用）
func (m *MySQLSessionStorage) CleanupExpiredSessions() error {
	query := fmt.Sprintf("DELETE FROM %s WHERE expires_at IS NOT NULL AND expires_at < NOW()", m.tableName)
	_, err := m.db.Exec(query)
	return err
}

// 使用示例
func main() {
	// 创建MySQL会话存储
	mysqlStorage, err := NewMySQLSessionStorage("root:password@tcp(localhost:3306)/simpledb")
	if err != nil {
		panic(err)
	}
	defer mysqlStorage.Close()

	// 创建服务器
	server, err := handlers.NewServer()
	if err != nil {
		panic(err)
	}

	// 设置MySQL存储
	server.SetSessionStorage(mysqlStorage)

	// 可选：定期清理过期会话
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := mysqlStorage.CleanupExpiredSessions(); err != nil {
				fmt.Printf("清理过期会话失败: %v\n", err)
			}
		}
	}()

	// 注册路由并启动服务器
	router := handlers.NewGinRouter(nil)
	server.RegisterRoutes(router)
	router.Engine().Run(":8080")
}

