//go:build ignore
// +build ignore

package main

import (
	"log"

	"github.com/chenhg5/simple-db-web/handlers"
)

// 这是一个展示如何自定义JavaScript逻辑的示例
// 运行前需要安装 Gin: go get -u github.com/gin-gonic/gin
// 运行方式: go run examples/custom_js_example.go
func main() {
	// 创建服务器实例
	server, err := handlers.NewServer()
	if err != nil {
		log.Fatalf("创建服务器失败: %v", err)
	}

	// 设置自定义JavaScript脚本
	// 这个脚本会在页面加载后执行，可以配置请求拦截器等
	customScript := `
		// 从cookie中读取token并添加到所有请求的header
		function getCookie(name) {
			const value = '; ' + document.cookie;
			const parts = value.split('; ' + name + '=');
			if (parts.length === 2) return parts.pop().split(';').shift();
			return null;
		}

		// 配置请求拦截器
		if (window.SimpleDB && window.SimpleDB.config) {
			window.SimpleDB.config.requestInterceptor = function(url, options) {
				// 从cookie中读取token
				const token = getCookie('auth_token');
				if (token) {
					options.headers = options.headers || {};
					options.headers['Authorization'] = 'Bearer ' + token;
				}
				
				// 也可以添加其他自定义header
				options.headers['X-Custom-Header'] = 'custom-value';
				
				return { url, options };
			};

			// 可选：配置响应拦截器
			window.SimpleDB.config.responseInterceptor = function(response) {
				// 可以在这里处理响应，比如检查token是否过期
				if (response.status === 401) {
					// token过期，跳转到登录页
					window.location.href = '/login';
				}
				return response;
			};

			// 可选：配置错误拦截器
			window.SimpleDB.config.errorInterceptor = function(error, url, options) {
				// 可以在这里处理错误，比如记录日志
				console.error('API请求失败:', url, error);
				return error;
			};
		}
	`

	server.SetCustomScript(customScript)

	// 创建 Gin 适配器
	ginRouter := handlers.NewGinRouter(nil)

	// 注册路由
	server.RegisterRoutes(ginRouter)

	// 启动服务器
	addr := ":8080"
	log.Printf("服务器启动在 %s", addr)
	log.Println("访问 http://localhost:8080 查看效果")
	log.Println("自定义脚本已注入，所有API请求会自动添加Authorization header（如果cookie中有auth_token）")
	if err := ginRouter.Engine().Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

