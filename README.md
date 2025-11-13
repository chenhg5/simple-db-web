# SimpleDBWeb - Database Management Tool

A modern database management web tool implemented with Go and Go Template, supporting multiple database types.

## Features

- ✅ Establish database connections (supports DSN or form input)
- ✅ List all tables in the database
- ✅ Display table structure (supports one-click copy)
- ✅ Query table data (with pagination)
- ✅ Edit row data in tables
- ✅ Delete row data in tables
- ✅ Execute SQL queries and display results
- ✅ Modular design, easy to extend support for other databases
- ✅ Modern UI with yellow theme (similar to Beekeeper Studio)
- ✅ Support for multi-instance deployment (via custom session storage)
- ✅ Adapter pattern, supports integration with Gin, Echo, and other web frameworks
- ✅ Embedded resources, supports importing via go mod

## Supported Databases

- MySQL
- PostgreSQL
- SQLite
- ClickHouse
- Dameng (达梦)
- OpenGauss
- Vastbase
- Kingbase (人大金仓)
- OceanDB

## Quick Start

### Build

```bash
go build -o dbweb
```

### Run

```bash
./dbweb
```

The server will start at `http://localhost:8080`.

### Use as a Library

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

    // Use standard library
    server.SetupRoutes()
    server.Start(":8080")

    // Or use Gin framework
    // router := handlers.NewGinRouter(nil)
    // server.RegisterRoutes(router)
    // router.Engine().Run(":8080")
}
```

## Usage

1. **Connect to Database**
   - Select database type
   - Choose connection method:
     - DSN connection string: Enter full DSN directly, e.g., `user:password@tcp(host:port)/database`
     - Form input: Enter host, port, username, password separately
   - Click "Connect" button

2. **Select Database**
   - After successful connection, select the database to operate on

3. **View Table List**
   - After selecting a database, all tables will be displayed on the left
   - Click a table name to view its data and structure

4. **View Table Data**
   - View table data in the "Data" tab
   - Supports pagination
   - Can edit or delete row data
   - Action column is fixed on the right for easy operation

5. **View Table Structure**
   - View CREATE TABLE statement in the "Structure" tab
   - Supports one-click copy of table structure

6. **Execute SQL Queries**
   - Enter SQL statements in the "SQL Query" tab
   - Supports SELECT, UPDATE, DELETE, INSERT operations
   - Click "Execute Query" button to run

## Project Structure

```
dbweb/
├── main.go              # Main program entry
├── database/            # Database interface and implementations
│   ├── interface.go     # Database interface definition
│   ├── mysql.go         # MySQL implementation
│   ├── postgresql.go    # PostgreSQL implementation
│   ├── sqlite3.go       # SQLite implementation
│   ├── clickhouse.go    # ClickHouse implementation
│   └── mysql_based*.go  # MySQL-compatible database implementations
├── handlers/            # HTTP handlers
│   ├── handlers.go      # Routes and handlers
│   ├── adapter*.go      # Adapter implementations (Gin, Echo, etc.)
│   ├── embed.go         # Resource file embedding
│   ├── templates/       # HTML templates
│   │   └── index.html
│   └── static/          # Static resources
│       ├── style.css
│       └── app.js
├── examples/            # Usage examples
│   ├── gin_example.go   # Gin framework example
│   ├── echo_example.go  # Echo framework example
│   └── ...
├── docs/                # Documentation
│   ├── zh/              # Chinese documentation
│   └── en/              # English documentation
└── README.md            # Documentation
```

## Extending Functionality

### Extend Support for Other Databases

To add support for other databases:

1. Create a new implementation file in the `database/` directory
2. Implement the `Database` interface
3. Add the new database type in `handlers/handlers.go`'s `NewServer` function

Or use the `AddDatabase` method to register dynamically:

```go
server.AddDatabase("custom_db", func() database.Database {
    return &MyCustomDatabase{}
})
```

### Integrate with Other Web Frameworks

Supports Gin, Echo, and other frameworks. See [Adapter Usage Guide](docs/en/ADAPTER_USAGE.md) for details.

### Custom Session Storage

Supports persistent storage like Redis, MySQL for multi-instance deployment. See [Session Storage Usage Guide](docs/en/SESSION_STORAGE_USAGE.md) for details.

### Custom JavaScript Logic

Supports injecting custom JavaScript, such as adding authentication tokens. See [Custom JS Usage Guide](docs/en/CUSTOM_JS_USAGE.md) for details.

## Tech Stack

- **Backend**: Go 1.16+
- **Database Drivers**: 
  - MySQL: `github.com/go-sql-driver/mysql`
  - PostgreSQL: `github.com/lib/pq`
  - SQLite: `github.com/mattn/go-sqlite3`
  - ClickHouse: `github.com/ClickHouse/clickhouse-go/v2`
- **Frontend**: Vanilla JavaScript + CSS
- **Templates**: Go Template
- **Resource Embedding**: Go 1.16+ embed

## Documentation

- [Adapter Usage Guide](docs/en/ADAPTER_USAGE.md) - How to integrate with Gin, Echo, and other frameworks
- [Session Storage Usage Guide](docs/en/SESSION_STORAGE_USAGE.md) - How to implement multi-instance deployment
- [Custom JS Usage Guide](docs/en/CUSTOM_JS_USAGE.md) - How to add custom JavaScript logic
- [Embed Usage Guide](docs/en/EMBED_USAGE.md) - Resource file embedding guide

## Notes

- Edit and delete operations are based on primary keys (PRI)
- String values in SQL queries are escaped, but parameterized queries are recommended in production
- For multi-instance deployment, Redis or MySQL is recommended as session storage

## License

MIT

