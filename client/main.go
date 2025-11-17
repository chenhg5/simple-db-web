package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/chenhg5/simple-db-web/database"
	"github.com/chenhg5/simple-db-web/handlers"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

//go:embed static/user-management.js
var userManagementScript string

//go:embed templates/login.html
var loginHTML string

// 语义化版本号
const AppVersion = "v0.0.1"

// 构建信息（通过 ldflags 注入）
var (
	Version   string
	BuildTime string
	Commit    string
	GoVersion string
)

// Config 配置结构
type Config struct {
	Port        string
	Debug       bool
	LogFile     string
	EnableAuth  bool
	RoutePrefix string
	AutoOpen    bool
	DBPath      string
	Connections string // 预设连接 YAML 文件路径
}

// ConnectionsConfig YAML 配置文件结构
type ConnectionsConfig struct {
	Connections []database.ConnectionInfo `yaml:"connections"`
}

func main() {
	// 检查 version 命令
	if len(os.Args) > 1 && os.Args[1] == "version" {
		printVersion()
		os.Exit(0)
	}

	// 解析命令行参数
	config := parseFlags()

	// 初始化日志
	initLogger(config)

	// 初始化数据库
	if config.EnableAuth {
		if err := InitDB(config.DBPath); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer CloseDB()

		// 定期清理过期会话
		go func() {
			ticker := time.NewTicker(1 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				if err := CleanExpiredSessions(); err != nil {
					log.Printf("Failed to clean expired sessions: %v", err)
				}
			}
		}()
	}

	// 创建核心服务器
	server, err := handlers.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// 如果启用认证，设置自定义脚本（注入用户管理UI）
	if config.EnableAuth {
		server.SetCustomScript(userManagementScript)
	}

	// 加载预设连接
	if config.Connections != "" {
		if err := loadPresetConnections(server, config.Connections); err != nil {
			log.Printf("Warning: Failed to load preset connections: %v", err)
		} else {
			log.Printf("Loaded preset connections from: %s", config.Connections)
		}
	}

	// 创建Gin引擎
	if config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	// 设置日志中间件
	if config.Debug {
		engine.Use(gin.Logger())
	}
	engine.Use(gin.Recovery())

	// 如果启用认证，注册认证相关路由
	if config.EnableAuth {
		// 设置路由前缀（供中间件使用）
		SetRoutePrefix(config.RoutePrefix)
		// 注册认证中间件
		engine.Use(AuthMiddleware())

		// 登录相关路由（不需要认证）
		engine.GET(joinPath(config.RoutePrefix, "/login"), LoginPage)
		engine.POST(joinPath(config.RoutePrefix, "/api/auth/login"), LoginAPI)

		// 认证后的路由
		engine.GET(joinPath(config.RoutePrefix, "/api/auth/current"), GetCurrentUserAPI)
		engine.POST(joinPath(config.RoutePrefix, "/api/auth/logout"), LogoutAPI)
		engine.POST(joinPath(config.RoutePrefix, "/api/auth/password"), UpdatePasswordAPI)

		// 用户管理路由（需要管理员权限）
		adminGroup := engine.Group(joinPath(config.RoutePrefix, "/api/users"))
		adminGroup.Use(AdminMiddleware())
		{
			adminGroup.GET("", GetUsersAPI)
			adminGroup.POST("", CreateUserAPI)
			adminGroup.PUT("/:id", UpdateUserAPI)
			adminGroup.DELETE("/:id", DeleteUserAPI)
		}
	}

	// 注册核心服务器的路由
	ginRouter := handlers.NewGinRouter(engine)

	// 如果有路由前缀，使用PrefixRouter包装
	var router handlers.Router = ginRouter
	if config.RoutePrefix != "" {
		router = handlers.NewPrefixRouter(ginRouter, config.RoutePrefix)
	}

	// 支持数据库
	server.AddDatabase(func() database.Database {
		return database.NewBaseMysqlBasedDB("oceanbase")
	})
	server.AddDatabase(func() database.Database {
		return database.NewClickHouse()
	})
	server.AddDatabase(func() database.Database {
		return database.NewSQLite3()
	})
	server.AddDatabase(func() database.Database {
		return database.NewPostgreSQL()
	})
	server.AddDatabase(func() database.Database {
		return database.NewOracle()
	})
	server.AddDatabase(func() database.Database {
		return database.NewSQLServer()
	})
	server.AddDatabase(func() database.Database {
		return database.NewMongoDB()
	})

	server.RegisterRoutes(router)

	// 构建完整的URL
	url := fmt.Sprintf("http://localhost%s", config.Port)
	if config.RoutePrefix != "" {
		url = fmt.Sprintf("%s%s", url, config.RoutePrefix)
	}

	log.Printf("Server starting on %s", url)

	// 显示构建信息（如果可用）
	if Version != "" {
		log.Printf("Version: %s", Version)
	}
	if BuildTime != "" {
		log.Printf("Build Time: %s", BuildTime)
	}
	if Commit != "" {
		log.Printf("Commit: %s", Commit)
	}
	if GoVersion != "" {
		log.Printf("Go Version: %s", GoVersion)
	}

	if config.EnableAuth {
		log.Printf("Default admin account: admin / admin123")
		log.Printf("Database file: %s", config.DBPath)
	}
	if config.RoutePrefix != "" {
		log.Printf("Route prefix: %s", config.RoutePrefix)
	}

	// 启动服务器（异步）
	go func() {
		if err := engine.Run(config.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)

	// 如果启用自动打开浏览器
	if config.AutoOpen {
		openBrowser(url)
	}

	// 保持程序运行
	select {}
}

// parseFlags 解析命令行参数
func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.Port, "port", ":8080", "Server port (e.g., :8080 or 8080)")
	flag.BoolVar(&config.Debug, "debug", false, "Enable debug mode (prints debug logs)")
	flag.StringVar(&config.LogFile, "log", "", "Log file path (empty means no file logging)")
	flag.BoolVar(&config.EnableAuth, "auth", false, "Enable authentication and user management")
	flag.StringVar(&config.RoutePrefix, "prefix", "", "Route prefix (e.g., /v1, /api)")
	flag.BoolVar(&config.AutoOpen, "open", false, "Automatically open browser after startup")
	flag.StringVar(&config.DBPath, "db", "client.db", "Database file path (only used when auth is enabled)")
	flag.StringVar(&config.Connections, "connections", "", "Path to YAML file containing preset connections")

	flag.Parse()

	// 确保端口格式正确
	if config.Port[0] != ':' {
		config.Port = ":" + config.Port
	}

	return config
}

// initLogger 初始化日志
func initLogger(config *Config) {
	// 如果指定了日志文件，设置日志输出到文件
	if config.LogFile != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(config.LogFile)
		if logDir != "." && logDir != "" {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				log.Printf("Failed to create log directory: %v", err)
			}
		}

		// 打开日志文件（追加模式）
		logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("Failed to open log file: %v, using stdout", err)
			return
		}

		// 设置日志输出：同时输出到文件和控制台
		log.SetOutput(logFile)
	}
}

// joinPath 拼接路径（处理前缀）
func joinPath(prefix, path string) string {
	if prefix == "" {
		return path
	}
	if path == "/" {
		return prefix
	}
	return prefix + path
}

// openBrowser 自动打开浏览器
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		// 尝试使用 xdg-open
		cmd = exec.Command("xdg-open", url)
	default:
		log.Printf("Unsupported platform for auto-opening browser: %s", runtime.GOOS)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}

// loadPresetConnections 从 YAML 文件加载预设连接
func loadPresetConnections(server *handlers.Server, yamlPath string) error {
	// 读取 YAML 文件
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to read connections file: %w", err)
	}

	// 解析 YAML
	var config ConnectionsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 验证连接信息
	if len(config.Connections) == 0 {
		log.Printf("No connections found in YAML file")
		return nil
	}

	// 设置预设连接
	server.SetPresetConnections(config.Connections)
	log.Printf("Loaded %d preset connection(s)", len(config.Connections))

	return nil
}

// printVersion 打印版本信息
func printVersion() {
	fmt.Printf("Version: %s\n", AppVersion)
	if Version != "" {
		fmt.Printf("Build Version: %s\n", Version)
	}
	if BuildTime != "" {
		fmt.Printf("Build Time: %s\n", BuildTime)
	}
	if Commit != "" {
		fmt.Printf("Commit: %s\n", Commit)
	}
	if GoVersion != "" {
		fmt.Printf("Go Version: %s\n", GoVersion)
	}
}
