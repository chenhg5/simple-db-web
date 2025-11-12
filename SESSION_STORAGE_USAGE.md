# 会话存储使用指南

本项目提供了灵活的会话存储接口，允许外部项目实现自定义的持久化存储（如Redis、MySQL等），以支持多实例部署。

## 为什么需要持久化存储？

在单实例部署中，会话数据存储在内存中即可。但在多实例部署（如Kubernetes多Pod、负载均衡）场景下，不同请求可能被转发到不同的实例，内存存储会导致会话丢失。使用持久化存储（如Redis、MySQL）可以解决这个问题。

## 架构设计

### SessionStorage 接口

```go
type SessionStorage interface {
    // Get 获取会话数据
    Get(connectionID string) (*SessionData, error)
    
    // Set 保存会话数据
    Set(connectionID string, data *SessionData, ttl time.Duration) error
    
    // Delete 删除会话数据
    Delete(connectionID string) error
    
    // Close 关闭存储连接
    Close() error
}
```

### SessionData 结构

```go
type SessionData struct {
    ConnectionInfo  database.ConnectionInfo // 连接信息（用于重建连接）
    DSN            string                  // DSN连接字符串
    DbType         string                  // 数据库类型
    CurrentDatabase string                 // 当前数据库
    CurrentTable    string                 // 当前表
    CreatedAt       time.Time              // 创建时间
}
```

**注意**：`SessionData` 只存储可序列化的连接信息，不包含实际的数据库连接对象。当需要时，系统会根据 `SessionData` 自动重建数据库连接。

## 使用方式

### 方法一：使用默认内存存储（单实例）

默认情况下，使用内存存储，适用于单实例部署：

```go
server, err := handlers.NewServer()
// 无需额外配置，默认使用内存存储
```

### 方法二：使用Redis存储（多实例推荐）

适用于多实例部署，通过Redis共享会话数据：

```go
package main

import (
    "github.com/chenhg5/simple-db-web/handlers"
    "github.com/redis/go-redis/v9"
)

// 实现RedisSessionStorage（参考examples/redis_session_storage.go）
type RedisSessionStorage struct {
    client *redis.Client
    // ... 其他字段
}

func (r *RedisSessionStorage) Get(connectionID string) (*handlers.SessionData, error) {
    // 实现Get方法
}

func (r *RedisSessionStorage) Set(connectionID string, data *handlers.SessionData, ttl time.Duration) error {
    // 实现Set方法
}

func (r *RedisSessionStorage) Delete(connectionID string) error {
    // 实现Delete方法
}

func (r *RedisSessionStorage) Close() error {
    // 实现Close方法
}

func main() {
    // 创建Redis存储
    redisStorage := NewRedisSessionStorage("localhost:6379", "", 0)
    
    // 创建服务器
    server, err := handlers.NewServer()
    if err != nil {
        panic(err)
    }
    
    // 设置Redis存储
    server.SetSessionStorage(redisStorage)
    
    // 注册路由并启动
    router := handlers.NewGinRouter(nil)
    server.RegisterRoutes(router)
    router.Engine().Run(":8080")
}
```

### 方法三：使用MySQL存储

适用于多实例部署，通过MySQL共享会话数据：

```go
package main

import (
    "github.com/chenhg5/simple-db-web/handlers"
)

// 实现MySQLSessionStorage（参考examples/mysql_session_storage.go）
type MySQLSessionStorage struct {
    db *sql.DB
    // ... 其他字段
}

func main() {
    // 创建MySQL存储
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
    
    // 注册路由并启动
    router := handlers.NewGinRouter(nil)
    server.RegisterRoutes(router)
    router.Engine().Run(":8080")
}
```

## 实现示例

### Redis实现示例

完整实现请参考 `examples/redis_session_storage.go`：

```go
type RedisSessionStorage struct {
    client    *redis.Client
    ctx       context.Context
    keyPrefix string
}

func (r *RedisSessionStorage) Get(connectionID string) (*handlers.SessionData, error) {
    key := r.keyPrefix + connectionID
    val, err := r.client.Get(r.ctx, key).Result()
    if err == redis.Nil {
        return nil, fmt.Errorf("会话不存在")
    }
    // ... 解析并返回
}

func (r *RedisSessionStorage) Set(connectionID string, data *handlers.SessionData, ttl time.Duration) error {
    key := r.keyPrefix + connectionID
    val, err := json.Marshal(data)
    // ... 保存到Redis
    return r.client.Set(r.ctx, key, val, ttl).Err()
}
```

### MySQL实现示例

完整实现请参考 `examples/mysql_session_storage.go`：

```go
type MySQLSessionStorage struct {
    db        *sql.DB
    tableName string
}

func (m *MySQLSessionStorage) Get(connectionID string) (*handlers.SessionData, error) {
    query := `SELECT session_data FROM simpledb_sessions WHERE connection_id = ?`
    var sessionDataJSON string
    err := m.db.QueryRow(query, connectionID).Scan(&sessionDataJSON)
    // ... 解析并返回
}

func (m *MySQLSessionStorage) Set(connectionID string, data *handlers.SessionData, ttl time.Duration) error {
    sessionDataJSON, err := json.Marshal(data)
    // ... 保存到MySQL
    query := `INSERT INTO simpledb_sessions ... ON DUPLICATE KEY UPDATE ...`
    return m.db.Exec(query, ...).Err()
}
```

## 工作原理

1. **连接时**：
   - 创建数据库连接
   - 将连接信息（DSN、类型等）保存到 `SessionData`
   - 保存到持久化存储（如Redis、MySQL）
   - 同时在内存中缓存 `ConnectionSession`（包含实际连接对象）

2. **请求时**：
   - 先检查内存缓存
   - 如果缓存中没有，从持久化存储读取 `SessionData`
   - 根据 `SessionData` 重建数据库连接
   - 保存到内存缓存

3. **更新时**：
   - 更新内存中的会话
   - 同步更新持久化存储

4. **断开时**：
   - 关闭数据库连接
   - 从内存缓存删除
   - 从持久化存储删除

## 注意事项

1. **TTL（过期时间）**：
   - 默认TTL为24小时
   - Redis实现支持TTL
   - MySQL实现需要手动清理过期会话（可以使用定时任务）

2. **线程安全**：
   - 所有存储实现都应该是线程安全的
   - 内存存储使用 `sync.RWMutex` 保护

3. **性能考虑**：
   - 内存缓存用于提高性能，避免每次都重建连接
   - 持久化存储用于跨实例共享

4. **连接重建**：
   - 系统会自动根据 `SessionData` 重建数据库连接
   - 如果连接失败，会返回错误

5. **密码安全**：
   - `SessionData` 包含连接密码，请确保持久化存储的安全性
   - 建议使用加密存储或安全的传输通道

## 最佳实践

1. **单实例部署**：使用默认内存存储
2. **多实例部署**：使用Redis或MySQL存储
3. **高可用**：使用Redis Cluster或MySQL主从
4. **性能优化**：合理设置TTL，定期清理过期会话
5. **安全**：加密存储敏感信息，使用HTTPS传输

## 更多示例

查看 `examples/` 目录获取完整的实现示例：
- `redis_session_storage.go` - Redis实现
- `mysql_session_storage.go` - MySQL实现

