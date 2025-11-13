# Adapter Pattern Usage Guide

This project uses the adapter pattern to support integrating handlers into different web frameworks.

## Architecture Design

### Core Interface

```go
type Router interface {
    GET(path string, handler http.HandlerFunc)
    POST(path string, handler http.HandlerFunc)
    Static(path, dir string)
    HandleFunc(path string, handler http.HandlerFunc)
}
```

### Adapter Implementations

- `StandardRouter` - Standard library net/http (default, no additional dependencies)
- `GinRouter` - Gin framework adapter (requires `github.com/gin-gonic/gin`)
- `EchoRouter` - Echo framework adapter (requires `github.com/labstack/echo/v4`)

## Usage

### 1. Standard Library (Default, No Changes Required)

```go
package main

import (
    "github.com/chenhg5/simple-db-web/handlers"
    "log"
)

func main() {
    server, err := handlers.NewServer()
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }

    // Use the original SetupRoutes method (backward compatible)
    server.SetupRoutes()

    if err := server.Start(":8080"); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

### 2. Gin Framework

#### Install Dependencies

```bash
go get -u github.com/gin-gonic/gin
```

#### Usage Example

```go
package main

import (
    "github.com/chenhg5/simple-db-web/handlers"
    "log"
)

func main() {
    server, err := handlers.NewServer()
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }

    // Create Gin adapter
    ginRouter := handlers.NewGinRouter(nil) // nil means use gin.Default()

    // Register routes
    server.RegisterRoutes(ginRouter)

    // Start server
    if err := ginRouter.Engine().Run(":8080"); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

#### Using Custom Gin Engine (Adding Middleware)

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
        log.Fatalf("Failed to create server: %v", err)
    }

    // Create custom Gin engine
    engine := gin.New()
    engine.Use(gin.Logger())
    engine.Use(gin.Recovery())
    
    // Add custom middleware (e.g., CORS)
    engine.Use(func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, X-Connection-ID")
        c.Next()
    })

    // Create adapter
    ginRouter := handlers.NewGinRouter(engine)

    // Register routes
    server.RegisterRoutes(ginRouter)

    // Start server
    if err := engine.Run(":8080"); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

### 3. Echo Framework

#### Install Dependencies

```bash
go get -u github.com/labstack/echo/v4
```

#### Usage Example

```go
package main

import (
    "github.com/chenhg5/simple-db-web/handlers"
    "log"
)

func main() {
    server, err := handlers.NewServer()
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }

    // Create Echo adapter
    echoRouter := handlers.NewEchoRouter(nil)

    // Register routes
    server.RegisterRoutes(echoRouter)

    // Start server
    if err := echoRouter.Echo().Start(":8080"); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

## Extending Support for Other Frameworks

To implement an adapter for another framework, you only need to implement the `Router` interface.

### Example: Creating an Adapter for Fiber Framework

1. Create `handlers/adapter_fiber.go`:

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
        // Convert fiber.Ctx to http.Request/ResponseWriter
        // Implementation of conversion logic needed here
        return nil
    })
}

func (r *FiberRouter) POST(path string, handler http.HandlerFunc) {
    r.app.Post(path, func(c *fiber.Ctx) error {
        // Implementation of conversion logic needed here
        return nil
    })
}

func (r *FiberRouter) Static(path, dir string) {
    r.app.Static(path, dir)
}

func (r *FiberRouter) HandleFunc(path string, handler http.HandlerFunc) {
    r.app.All(path, func(c *fiber.Ctx) error {
        // Implementation of conversion logic needed here
        return nil
    })
}

func (r *FiberRouter) App() *fiber.App {
    return r.app
}
```

2. Usage:

```go
fiberRouter := handlers.NewFiberRouter(nil)
server.RegisterRoutes(fiberRouter)
fiberRouter.App().Listen(":8080")
```

## API Route List

All routes are registered through the `RegisterRoutes` method:

- `GET /` - Home page
- `POST /api/connect` - Connect to database
- `POST /api/disconnect` - Disconnect
- `GET /api/status` - Get connection status
- `GET /api/databases` - Get database list
- `POST /api/database/switch` - Switch database
- `GET /api/tables` - Get table list
- `GET /api/table/schema` - Get table schema
- `GET /api/table/columns` - Get table column information
- `GET /api/table/data` - Get table data
- `POST /api/query` - Execute SQL query
- `POST /api/row/update` - Update row data
- `POST /api/row/delete` - Delete row data
- `GET /static/*` - Static files
- `GET /api/database/types` - Get database type list

### 5. Using Route Prefix

If you need to add a prefix to all routes (e.g., `/v1`, `/api/v1`), you can use the `NewPrefixRouter` wrapper:

```go
package main

import (
    "github.com/chenhg5/simple-db-web/handlers"
    "log"
)

func main() {
    server, err := handlers.NewServer()
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }

    // Create base adapter
    ginRouter := handlers.NewGinRouter(nil)

    // Use prefix wrapper, all routes will have /v1 prefix
    // Example: /api/connect -> /v1/api/connect
    prefixedRouter := handlers.NewPrefixRouter(ginRouter, "/v1")

    // Register routes
    server.RegisterRoutes(prefixedRouter)

    // Start server
    ginRouter.Engine().Run(":8080")
}
```

**Prefix Rules**:
- Prefix automatically adds leading `/` (if not present)
- Trailing `/` in prefix is removed (unless it's root path `/`)
- Examples: `v1` -> `/v1`, `/v1/` -> `/v1`

## Notes

1. **Dependency Management**: Gin and Echo adapters are optional, only install dependencies when using the corresponding framework
2. **Backward Compatibility**: The original `SetupRoutes()` method is still available, internally using `StandardRouter`
3. **Connection ID**: All APIs pass connection identifier through the `X-Connection-ID` request header
4. **Static Files**: Default path is `static/` directory
5. **Route Prefix**: Use `NewPrefixRouter` to add prefix support to any adapter

## Adding Custom Database Types

You can dynamically register custom database types through the `AddDatabase` method:

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
        log.Fatalf("Failed to create server: %v", err)
    }

    // Add custom database type
    server.AddDatabase("custom_db", func() database.Database {
        // Return an instance that implements the database.Database interface
        return &MyCustomDatabase{}
    })

    server.SetupRoutes()
    server.Start(":8080")
}
```

**Notes**:
- Custom database types must implement all methods of the `database.Database` interface
- Database type identifier (name) should be unique
- Frontend automatically fetches all available database types (including built-in and custom) from `/api/database/types`

## Advantages

1. **Decoupling**: Business logic is decoupled from web frameworks
2. **Flexibility**: Easy to switch or support multiple frameworks simultaneously
3. **Extensibility**:
   - Support for dynamically adding custom database types
   - Adding new framework support only requires implementing the `Router` interface
4. **Backward Compatibility**: Does not affect existing code

