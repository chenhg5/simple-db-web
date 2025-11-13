# Embed 使用说明

本项目使用 Go 1.16+ 的 `embed` 功能将 templates 和 static 文件嵌入到二进制文件中，确保通过 `go mod` 引入时也能正常工作。

## 实现方式

### 1. 文件结构

```
handlers/
├── embed.go          # embed 定义文件
├── templates/        # 模板文件目录
│   └── index.html
└── static/          # 静态文件目录
    ├── app.js
    └── style.css
```

### 2. Embed 定义

在 `handlers/embed.go` 中定义：

```go
package handlers

import (
	"embed"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS
```

### 3. 使用方式

#### 在 handlers 包内使用

- **模板解析**：`template.ParseFS(templatesFS, "templates/*.html")`
- **静态文件服务**：通过 `router.StaticFS("/static/", staticFS)` 注册

#### 在其他项目中引入

当其他项目通过 `go mod` 引入本包时：

```go
import (
	"github.com/chenhg5/simple-db-web/handlers"
)

func main() {
	// 直接使用，templates 和 static 已经嵌入
	server, err := handlers.NewServer()
	if err != nil {
		log.Fatalf("创建服务器失败: %v", err)
	}
	
	// 注册路由
	router := handlers.NewStandardRouter()
	server.RegisterRoutes(router)
	
	// 启动服务器
	server.Start(":8080")
}
```

## 优势

1. **单文件部署**：所有资源都编译到二进制文件中，无需额外的文件目录
2. **模块化支持**：通过 `go mod` 引入时，资源文件自动包含
3. **版本一致性**：资源文件与代码版本绑定，避免版本不匹配问题
4. **部署简化**：只需一个二进制文件即可运行

## 注意事项

1. **路径问题**：embed 的路径是相对于包含 `//go:embed` 指令的 `.go` 文件的
2. **文件大小**：嵌入的文件会增加二进制文件的大小
3. **开发时修改**：修改 templates 或 static 文件后需要重新编译才能生效

## 兼容性

- 支持 Go 1.16+
- 支持所有适配器（StandardRouter、GinRouter、EchoRouter）
- 向后兼容：如果不需要 embed，可以继续使用 `Static(path, dir)` 方法

