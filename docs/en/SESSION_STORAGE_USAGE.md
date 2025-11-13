# Session Storage Usage Guide

This project provides a flexible session storage interface that allows external projects to implement custom persistent storage (such as Redis, MySQL, etc.) to support multi-instance deployment.

## Why Persistent Storage?

In single-instance deployment, session data can be stored in memory. However, in multi-instance deployment scenarios (such as Kubernetes multi-pod, load balancing), different requests may be forwarded to different instances, causing session loss with in-memory storage. Using persistent storage (such as Redis, MySQL) can solve this problem.

## Architecture Design

### SessionStorage Interface

```go
type SessionStorage interface {
    // Get retrieves session data
    Get(connectionID string) (*SessionData, error)
    
    // Set saves session data
    Set(connectionID string, data *SessionData, ttl time.Duration) error
    
    // Delete removes session data
    Delete(connectionID string) error
    
    // Close closes the storage connection
    Close() error
}
```

### SessionData Structure

```go
type SessionData struct {
    ConnectionInfo  database.ConnectionInfo // Connection info (for rebuilding connection)
    DSN            string                  // DSN connection string
    DbType         string                  // Database type
    CurrentDatabase string                 // Current database
    CurrentTable    string                 // Current table
    CreatedAt       time.Time              // Creation time
}
```

**Note**: `SessionData` only stores serializable connection information, not actual database connection objects. When needed, the system will automatically rebuild database connections based on `SessionData`.

## Usage

### Method 1: Use Default In-Memory Storage (Single Instance)

By default, in-memory storage is used, suitable for single-instance deployment:

```go
server, err := handlers.NewServer()
// No additional configuration needed, uses in-memory storage by default
```

### Method 2: Use Redis Storage (Recommended for Multi-Instance)

Suitable for multi-instance deployment, sharing session data through Redis:

```go
package main

import (
    "github.com/chenhg5/simple-db-web/handlers"
    "github.com/redis/go-redis/v9"
)

// Implement RedisSessionStorage (refer to examples/redis_session_storage.go)
type RedisSessionStorage struct {
    client *redis.Client
    // ... other fields
}

func (r *RedisSessionStorage) Get(connectionID string) (*handlers.SessionData, error) {
    // Implement Get method
}

func (r *RedisSessionStorage) Set(connectionID string, data *handlers.SessionData, ttl time.Duration) error {
    // Implement Set method
}

func (r *RedisSessionStorage) Delete(connectionID string) error {
    // Implement Delete method
}

func (r *RedisSessionStorage) Close() error {
    // Implement Close method
}

func main() {
    // Create Redis storage
    redisStorage := NewRedisSessionStorage("localhost:6379", "", 0)
    
    // Create server
    server, err := handlers.NewServer()
    if err != nil {
        panic(err)
    }
    
    // Set Redis storage
    server.SetSessionStorage(redisStorage)
    
    // Register routes and start
    router := handlers.NewGinRouter(nil)
    server.RegisterRoutes(router)
    router.Engine().Run(":8080")
}
```

### Method 3: Use MySQL Storage

Suitable for multi-instance deployment, sharing session data through MySQL:

```go
package main

import (
    "github.com/chenhg5/simple-db-web/handlers"
)

// Implement MySQLSessionStorage (refer to examples/mysql_session_storage.go)
type MySQLSessionStorage struct {
    db *sql.DB
    // ... other fields
}

func main() {
    // Create MySQL storage
    mysqlStorage, err := NewMySQLSessionStorage("root:password@tcp(localhost:3306)/simpledb")
    if err != nil {
        panic(err)
    }
    defer mysqlStorage.Close()
    
    // Create server
    server, err := handlers.NewServer()
    if err != nil {
        panic(err)
    }
    
    // Set MySQL storage
    server.SetSessionStorage(mysqlStorage)
    
    // Register routes and start
    router := handlers.NewGinRouter(nil)
    server.RegisterRoutes(router)
    router.Engine().Run(":8080")
}
```

## Implementation Examples

### Redis Implementation Example

See `examples/redis_session_storage.go` for complete implementation:

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
        return nil, fmt.Errorf("session not found")
    }
    // ... parse and return
}

func (r *RedisSessionStorage) Set(connectionID string, data *handlers.SessionData, ttl time.Duration) error {
    key := r.keyPrefix + connectionID
    val, err := json.Marshal(data)
    // ... save to Redis
    return r.client.Set(r.ctx, key, val, ttl).Err()
}
```

### MySQL Implementation Example

See `examples/mysql_session_storage.go` for complete implementation:

```go
type MySQLSessionStorage struct {
    db        *sql.DB
    tableName string
}

func (m *MySQLSessionStorage) Get(connectionID string) (*handlers.SessionData, error) {
    query := `SELECT session_data FROM simpledb_sessions WHERE connection_id = ?`
    var sessionDataJSON string
    err := m.db.QueryRow(query, connectionID).Scan(&sessionDataJSON)
    // ... parse and return
}

func (m *MySQLSessionStorage) Set(connectionID string, data *SessionData, ttl time.Duration) error {
    sessionDataJSON, err := json.Marshal(data)
    // ... save to MySQL
    query := `INSERT INTO simpledb_sessions ... ON DUPLICATE KEY UPDATE ...`
    return m.db.Exec(query, ...).Err()
}
```

## How It Works

1. **On Connection**:
   - Create database connection
   - Save connection information (DSN, type, etc.) to `SessionData`
   - Save to persistent storage (e.g., Redis, MySQL)
   - Also cache `ConnectionSession` in memory (contains actual connection object)

2. **On Request**:
   - First check memory cache
   - If not in cache, read `SessionData` from persistent storage
   - Rebuild database connection based on `SessionData`
   - Save to memory cache

3. **On Update**:
   - Update session in memory
   - Synchronously update persistent storage

4. **On Disconnect**:
   - Close database connection
   - Delete from memory cache
   - Delete from persistent storage

## Notes

1. **TTL (Time To Live)**:
   - Default TTL is 24 hours
   - Redis implementation supports TTL
   - MySQL implementation requires manual cleanup of expired sessions (can use scheduled tasks)

2. **Thread Safety**:
   - All storage implementations should be thread-safe
   - In-memory storage uses `sync.RWMutex` for protection

3. **Performance Considerations**:
   - Memory cache is used to improve performance, avoiding rebuilding connections every time
   - Persistent storage is used for cross-instance sharing

4. **Connection Rebuilding**:
   - System automatically rebuilds database connections based on `SessionData`
   - If connection fails, an error will be returned

5. **Password Security**:
   - `SessionData` contains connection passwords, ensure the security of persistent storage
   - Recommend using encrypted storage or secure transmission channels

## Best Practices

1. **Single Instance Deployment**: Use default in-memory storage
2. **Multi-Instance Deployment**: Use Redis or MySQL storage
3. **High Availability**: Use Redis Cluster or MySQL master-slave
4. **Performance Optimization**: Set TTL appropriately, regularly clean up expired sessions
5. **Security**: Encrypt sensitive information, use HTTPS for transmission

## More Examples

See the `examples/` directory for complete implementation examples:
- `redis_session_storage.go` - Redis implementation
- `mysql_session_storage.go` - MySQL implementation

