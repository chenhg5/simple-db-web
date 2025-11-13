# 适配器模式使用指南

本项目使用适配器模式，支持将 handlers 接入到不同的 Web 框架中。

## 架构设计

### 核心接口

```go
type Router interface {
    GET(path string, handler http.HandlerFunc)
    POST(path string, handler http.HandlerFunc)
    Static(path, dir string)
    HandleFunc(path string, handler http.HandlerFunc)
}
```

### 适配器实现

- `StandardRouter` - 标准库 net/http（默认，无需额外依赖）
- `GinRouter` - Gin 框架适配器（需要安装 `github.com/gin-gonic/gin`）
- `EchoRouter` - Echo 框架适配器（需要安装 `github.com/labstack/echo/v4`）

## 使用方法

### 1. 标准库（默认，无需修改）

```go
package main

import (
    "github.com/chenhg5/simple-db-web/handlers"
    "log"
)

func main() {
    server, err := handlers.NewServer()
    if err != nil {
        log.Fatalf("创建服务器失败: %v", err)
    }

    // 使用原有的 SetupRoutes 方法（向后兼容）
    server.SetupRoutes()

    if err := server.Start(":8080"); err != nil {
        log.Fatalf("启动服务器失败: %v", err)
    }
}
```

### 2. Gin 框架

#### 安装依赖

```bash
go get -u github.com/gin-gonic/gin
```

#### 使用示例

```go
package main

import (
    "github.com/chenhg5/simple-db-web/handlers"
    "log"
)

func main() {
    server, err := handlers.NewServer()
    if err != nil {
        log.Fatalf("创建服务器失败: %v", err)
    }

    // 创建 Gin 适配器
    ginRouter := handlers.NewGinRouter(nil) // nil 表示使用 gin.Default()

    // 注册路由
    server.RegisterRoutes(ginRouter)

    // 启动服务器
    if err := ginRouter.Engine().Run(":8080"); err != nil {
        log.Fatalf("启动服务器失败: %v", err)
    }
}
```

#### 使用自定义 Gin 引擎（添加中间件）

```go
package main

import (
    "github.com/chenhg5/simple-db-web/handlers"
    "github.com/gin-gonic/gin"
    "log"
)

func main() {
    server, err := handlers.NewServer()
    if err != nil {
        log.Fatalf("创建服务器失败: %v", err)
    }

    // 创建自定义 Gin 引擎
    engine := gin.New()
    engine.Use(gin.Logger())
    engine.Use(gin.Recovery())
    
    // 添加自定义中间件（如 CORS）
    engine.Use(func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, X-Connection-ID")
        c.Next()
    })

    // 创建适配器
    ginRouter := handlers.NewGinRouter(engine)

    // 注册路由
    server.RegisterRoutes(ginRouter)

    // 启动服务器
    if err := engine.Run(":8080"); err != nil {
        log.Fatalf("启动服务器失败: %v", err)
    }
}
```

### 3. Echo 框架

#### 安装依赖

```bash
go get -u github.com/labstack/echo/v4
```

#### 使用示例

```go
package main

import (
    "github.com/chenhg5/simple-db-web/handlers"
    "log"
)

func main() {
    server, err := handlers.NewServer()
    if err != nil {
        log.Fatalf("创建服务器失败: %v", err)
    }

    // 创建 Echo 适配器
    echoRouter := handlers.NewEchoRouter(nil)

    // 注册路由
    server.RegisterRoutes(echoRouter)

    // 启动服务器
    if err := echoRouter.Echo().Start(":8080"); err != nil {
        log.Fatalf("启动服务器失败: %v", err)
    }
}
```

## 扩展支持其他框架

要实现其他框架的适配器，只需要实现 `Router` 接口。

### 示例：为 Fiber 框架创建适配器

1. 创建 `handlers/adapter_fiber.go`：

```go
package handlers

import (
    "github.com/gofiber/fiber/v2"
    "net/http"
)

type FiberRouter struct {
    app *fiber.App
}

func NewFiberRouter(app *fiber.App) *FiberRouter {
    if app == nil {
        app = fiber.New()
    }
    return &FiberRouter{app: app}
}

func (r *FiberRouter) GET(path string, handler http.HandlerFunc) {
    r.app.Get(path, func(c *fiber.Ctx) error {
        // 将 fiber.Ctx 转换为 http.Request/ResponseWriter
        // 这里需要实现转换逻辑
        return nil
    })
}

func (r *FiberRouter) POST(path string, handler http.HandlerFunc) {
    r.app.Post(path, func(c *fiber.Ctx) error {
        // 实现转换逻辑
        return nil
    })
}

func (r *FiberRouter) Static(path, dir string) {
    r.app.Static(path, dir)
}

func (r *FiberRouter) HandleFunc(path string, handler http.HandlerFunc) {
    r.app.All(path, func(c *fiber.Ctx) error {
        // 实现转换逻辑
        return nil
    })
}

func (r *FiberRouter) App() *fiber.App {
    return r.app
}
```

2. 使用：

```go
fiberRouter := handlers.NewFiberRouter(nil)
server.RegisterRoutes(fiberRouter)
fiberRouter.App().Listen(":8080")
```

## API 路由列表

所有路由通过 `RegisterRoutes` 方法注册：

- `GET /` - 首页
- `POST /api/connect` - 连接数据库
- `POST /api/disconnect` - 断开连接
- `GET /api/status` - 获取连接状态
- `GET /api/databases` - 获取数据库列表
- `POST /api/database/switch` - 切换数据库
- `GET /api/tables` - 获取表列表
- `GET /api/table/schema` - 获取表结构
- `GET /api/table/columns` - 获取表列信息
- `GET /api/table/data` - 获取表数据
- `POST /api/query` - 执行 SQL 查询
- `POST /api/row/update` - 更新行数据
- `POST /api/row/delete` - 删除行数据
- `GET /static/*` - 静态文件

### 5. 使用路由前缀

如果需要为所有路由添加前缀（例如 `/v1`、`/api/v1` 等），可以使用 `NewPrefixRouter` 包装器：

```go
package main

import (
    "github.com/chenhg5/simple-db-web/handlers"
    "log"
)

func main() {
    server, err := handlers.NewServer()
    if err != nil {
        log.Fatalf("创建服务器失败: %v", err)
    }

    // 创建基础适配器
    ginRouter := handlers.NewGinRouter(nil)

    // 使用前缀包装器，所有路由将添加 /v1 前缀
    // 例如: /api/connect -> /v1/api/connect
    prefixedRouter := handlers.NewPrefixRouter(ginRouter, "/v1")

    // 注册路由
    server.RegisterRoutes(prefixedRouter)

    // 启动服务器
    ginRouter.Engine().Run(":8080")
}
```

**前缀规则**：
- 前缀会自动添加前导 `/`（如果没有）
- 前缀末尾的 `/` 会被移除（除非是根路径 `/`）
- 示例：`v1` -> `/v1`，`/v1/` -> `/v1`

## 注意事项

1. **依赖管理**：Gin 和 Echo 适配器是可选的，只有使用对应框架时才需要安装依赖
2. **向后兼容**：原有的 `SetupRoutes()` 方法仍然可用，内部使用 `StandardRouter`
3. **连接 ID**：所有 API 通过请求头 `X-Connection-ID` 传递连接标识
4. **静态文件**：默认路径为 `static/` 目录
5. **路由前缀**：使用 `NewPrefixRouter` 可以为任何适配器添加前缀支持

## 添加自定义数据库类型

可以通过 `AddDatabase` 方法动态注册自定义数据库类型：

```go
package main

import (
    "github.com/chenhg5/simple-db-web/database"
    "github.com/chenhg5/simple-db-web/handlers"
    "log"
)

func main() {
    server, err := handlers.NewServer()
    if err != nil {
        log.Fatalf("创建服务器失败: %v", err)
    }

    // 添加自定义数据库类型
    server.AddDatabase("custom_db", func() database.Database {
        // 返回实现了 database.Database 接口的实例
        return &MyCustomDatabase{}
    })

    server.SetupRoutes()
    server.Start(":8080")
}
```

**注意事项**：
- 自定义数据库类型必须实现 `database.Database` 接口的所有方法
- 数据库类型标识（name）应该是唯一的
- 前端会自动从 `/api/database/types` 获取所有可用的数据库类型（包括内置和自定义的）

## 优势

1. **解耦**：业务逻辑与 Web 框架解耦
2. **灵活**：可以轻松切换或同时支持多个框架
3. **可扩展**：
   - 支持动态添加自定义数据库类型
   - 添加新框架支持只需实现 `Router` 接口
4. **向后兼容**：不影响现有代码

