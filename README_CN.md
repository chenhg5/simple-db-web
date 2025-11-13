# SimpleDBWeb - 数据库管理工具

一个使用 Go 和 Go Template 实现的现代化数据库管理 Web 工具，支持多种数据库类型。

## 功能特性

- ✅ 建立数据库连接（支持 DSN 或表单输入）
- ✅ 列出数据库中的所有表
- ✅ 展示数据表的结构（支持一键复制）
- ✅ 支持数据表的查询（分页）
- ✅ 支持对数据表某一行数据的编辑
- ✅ 支持对数据表某一行数据的删除
- ✅ 支持写查询 SQL，展示查询结果
- ✅ 模块化设计，易于扩展支持其他数据库
- ✅ 现代化 UI，黄色主题（类似 Beekeeper Studio）
- ✅ 支持多实例部署（通过自定义会话存储）
- ✅ 适配器模式，支持集成到 Gin、Echo 等 Web 框架
- ✅ 资源文件嵌入，支持通过 go mod 引入

## 支持的数据库

- MySQL
- PostgreSQL
- SQLite
- ClickHouse
- 达梦 (Dameng)
- OpenGauss
- Vastbase
- 人大金仓 (Kingbase)
- OceanDB

## 快速开始

### 编译

```bash
go build -o dbweb
```

### 运行

```bash
./dbweb
```

服务器将在 `http://localhost:8080` 启动。

### 作为库使用

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

    // 使用标准库
    server.SetupRoutes()
    server.Start(":8080")

    // 或使用 Gin 框架
    // router := handlers.NewGinRouter(nil)
    // server.RegisterRoutes(router)
    // router.Engine().Run(":8080")
}
```

## 使用说明

1. **连接数据库**
   - 选择数据库类型
   - 选择连接方式：
     - DSN 连接字符串：直接输入完整的 DSN，如 `user:password@tcp(host:port)/database`
     - 表单输入：分别输入主机、端口、用户名、密码
   - 点击"连接"按钮

2. **选择数据库**
   - 连接成功后，选择要操作的数据库

3. **查看表列表**
   - 选择数据库后，左侧会显示所有数据表
   - 点击表名可以查看表的数据和结构

4. **查看表数据**
   - 在"数据"标签页中查看表的数据
   - 支持分页浏览
   - 可以编辑或删除某一行数据
   - 操作列固定在右侧，方便操作

5. **查看表结构**
   - 在"结构"标签页中查看表的 CREATE TABLE 语句
   - 支持一键复制表结构

6. **执行 SQL 查询**
   - 在"SQL查询"标签页中输入 SQL 语句
   - 支持 SELECT、UPDATE、DELETE、INSERT 等操作
   - 点击"执行查询"按钮执行

## 项目结构

```
dbweb/
├── main.go              # 主程序入口
├── database/            # 数据库接口和实现
│   ├── interface.go     # Database 接口定义
│   ├── mysql.go         # MySQL 实现
│   ├── postgresql.go    # PostgreSQL 实现
│   ├── sqlite3.go       # SQLite 实现
│   ├── clickhouse.go    # ClickHouse 实现
│   └── mysql_based*.go  # MySQL 兼容数据库实现
├── handlers/            # HTTP 处理器
│   ├── handlers.go      # 路由和处理器
│   ├── adapter*.go      # 适配器实现（Gin、Echo等）
│   ├── embed.go         # 资源文件嵌入
│   ├── templates/         # HTML 模板
│   │   └── index.html
│   └── static/          # 静态资源
│       ├── style.css
│       └── app.js
├── examples/            # 使用示例
│   ├── gin_example.go   # Gin 框架示例
│   ├── echo_example.go  # Echo 框架示例
│   └── ...
├── docs/                # 文档
│   ├── zh/              # 中文文档
│   └── en/              # 英文文档
└── README.md            # 说明文档
```

## 扩展功能

### 扩展支持其他数据库

要实现其他数据库的支持，只需：

1. 在 `database/` 目录下创建新的实现文件
2. 实现 `Database` 接口
3. 在 `handlers/handlers.go` 的 `NewServer` 函数中添加新的数据库类型

或者使用 `AddDatabase` 方法动态注册：

```go
server.AddDatabase("custom_db", func() database.Database {
    return &MyCustomDatabase{}
})
```

### 集成到其他 Web 框架

支持 Gin、Echo 等框架，详见 [适配器使用指南](docs/zh/ADAPTER_USAGE.md)。

### 自定义会话存储

支持 Redis、MySQL 等持久化存储，适用于多实例部署，详见 [会话存储使用指南](docs/zh/SESSION_STORAGE_USAGE.md)。

### 自定义 JavaScript 逻辑

支持注入自定义 JavaScript，如添加认证 token，详见 [自定义 JS 使用指南](docs/zh/CUSTOM_JS_USAGE.md)。

## 技术栈

- **后端**: Go 1.16+
- **数据库驱动**: 
  - MySQL: `github.com/go-sql-driver/mysql`
  - PostgreSQL: `github.com/lib/pq`
  - SQLite: `github.com/mattn/go-sqlite3`
  - ClickHouse: `github.com/ClickHouse/clickhouse-go/v2`
- **前端**: 原生 JavaScript + CSS
- **模板**: Go Template
- **资源嵌入**: Go 1.16+ embed

## 文档

- [适配器使用指南](docs/zh/ADAPTER_USAGE.md) - 如何集成到 Gin、Echo 等框架
- [会话存储使用指南](docs/zh/SESSION_STORAGE_USAGE.md) - 如何实现多实例部署
- [自定义 JS 使用指南](docs/zh/CUSTOM_JS_USAGE.md) - 如何添加自定义 JavaScript 逻辑
- [Embed 使用说明](docs/zh/EMBED_USAGE.md) - 资源文件嵌入说明

## 注意事项

- 编辑和删除操作基于主键（PRI）
- SQL 查询中的字符串值已做基本的转义处理，但建议在生产环境中使用参数化查询
- 多实例部署时，建议使用 Redis 或 MySQL 作为会话存储

## 许可证

MIT

