# Embed Usage Guide

This project uses Go 1.16+ `embed` feature to embed templates and static files into the binary, ensuring it works correctly when imported via `go mod`.

## Implementation

### 1. File Structure

```
handlers/
├── embed.go          # embed definition file
├── templates/        # template files directory
│   └── index.html
└── static/          # static files directory
    ├── app.js
    └── style.css
```

### 2. Embed Definition

Defined in `handlers/embed.go`:

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

### 3. Usage

#### Using Within handlers Package

- **Template parsing**: `template.ParseFS(templatesFS, "templates/*.html")`
- **Static file serving**: Registered via `router.StaticFS("/static/", staticFS)`

#### Importing in Other Projects

When other projects import this package via `go mod`:

```go
import (
	"github.com/chenhg5/simple-db-web/handlers"
)

func main() {
	// Use directly, templates and static files are already embedded
	server, err := handlers.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	
	// Register routes
	router := handlers.NewStandardRouter()
	server.RegisterRoutes(router)
	
	// Start server
	server.Start(":8080")
}
```

## Advantages

1. **Single File Deployment**: All resources are compiled into the binary, no additional file directories needed
2. **Module Support**: When imported via `go mod`, resource files are automatically included
3. **Version Consistency**: Resource files are bound to code version, avoiding version mismatch issues
4. **Simplified Deployment**: Only one binary file is needed to run

## Notes

1. **Path Issues**: Embed paths are relative to the `.go` file containing the `//go:embed` directive
2. **File Size**: Embedded files will increase the binary file size
3. **Development Modifications**: After modifying template or static files, you need to recompile for changes to take effect

## Compatibility

- Supports Go 1.16+
- Supports all adapters (StandardRouter, GinRouter, EchoRouter)
- Backward compatible: If embed is not needed, you can continue using the `Static(path, dir)` method

