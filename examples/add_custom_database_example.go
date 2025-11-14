//go:build ignore
// +build ignore

package main

import (
	"log"

	"github.com/chenhg5/simple-db-web/database"
	"github.com/chenhg5/simple-db-web/handlers"
)

// 这是一个添加自定义数据库类型的示例
func main() {
	// 创建服务器实例
	server, err := handlers.NewServer()
	if err != nil {
		log.Fatalf("创建服务器失败: %v", err)
	}

	// 添加自定义数据库类型
	// 示例1：添加一个自定义的 MySQL 变体（使用默认显示名称，即类型名）
	server.AddDatabase("custom_mysql", func() database.Database {
		// 这里可以返回任何实现了 database.Database 接口的类型
		// 例如：可以包装 MySQL 并添加自定义逻辑
		return database.NewMySQL()
	})

	// 示例1.1：添加自定义数据库类型并指定显示名称
	server.AddDatabaseWithDisplayName("custom_mysql_v2", "自定义MySQL变体", func() database.Database {
		return database.NewMySQL()
	})

	// 使用 AddDatabaseWithDisplayName 添加自定义数据库类型并指定显示名称
	server.AddDatabaseWithDisplayName("mysql_based_dameng", "达梦", func() database.Database {
		return database.NewBaseMysqlBasedDB("dameng")
	})
	server.AddDatabaseWithDisplayName("mysql_based_openguass", "OpenGauss", func() database.Database {
		return database.NewBaseMysqlBasedDB("openguass")
	})
	server.AddDatabaseWithDisplayName("mysql_based_vastbase", "Vastbase", func() database.Database {
		return database.NewBaseMysqlBasedDB("vastbase")
	})
	server.AddDatabaseWithDisplayName("mysql_based_kingbase", "人大金仓", func() database.Database {
		return database.NewBaseMysqlBasedDB("kingbase")
	})

	presetConns := []database.ConnectionInfo{
        {
            Name:     "测试环境PostgreSQL",
            Type:     "postgresql",
            Host:     "test-db.example.com",
            Port:     "5432",
            User:     "testuser",
            Password: "testpass",
            Database: "testdb",
        },
    }
    server.SetPresetConnections(presetConns)

	// 示例2：添加另一个自定义数据库类型
	// 注意：需要实现 database.Database 接口的所有方法
	server.AddDatabase("my_custom_db", func() database.Database {
		// 返回自定义的数据库实现
		// 实际使用时需要完整实现 database.Database 接口
		return database.NewMySQL() // 这里仅作示例，实际应返回自定义实现
	})

	// 注册路由
	server.SetupRoutes()

	// 启动服务器
	addr := ":8080"
	log.Printf("服务器启动在 %s", addr)
	log.Printf("已注册自定义数据库类型: custom_mysql, my_custom_db")
	if err := server.Start(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
