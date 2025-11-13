# 自定义JavaScript逻辑使用指南

本项目提供了灵活的扩展机制，允许外部项目自定义JavaScript逻辑，特别是可以在所有AJAX请求中添加自定义的header（比如从cookie中读取token）。

## 功能特性

1. **请求拦截器**：在发送请求前可以修改请求配置（URL、headers、body等）
2. **响应拦截器**：在收到响应后可以处理响应
3. **错误拦截器**：在请求出错时处理错误
4. **脚本注入**：通过Go代码注入自定义JavaScript脚本

## 使用方法

### 方法一：通过Go代码注入脚本（推荐）

这是最优雅的方式，适合在服务端配置：

```go
package main

import (
	"log"
	"github.com/chenhg5/simple-db-web/handlers"
)

func main() {
	server, err := handlers.NewServer()
	if err != nil {
		log.Fatalf("创建服务器失败: %v", err)
	}

	// 设置自定义脚本
	customScript := `
		// 从cookie中读取token并添加到所有请求的header
		function getCookie(name) {
			const value = '; ' + document.cookie;
			const parts = value.split('; ' + name + '=');
			if (parts.length === 2) return parts.pop().split(';').shift();
			return null;
		}

		if (window.SimpleDB && window.SimpleDB.config) {
			window.SimpleDB.config.requestInterceptor = function(url, options) {
				const token = getCookie('auth_token');
				if (token) {
					options.headers = options.headers || {};
					options.headers['Authorization'] = 'Bearer ' + token;
				}
				return { url, options };
			};
		}
	`

	server.SetCustomScript(customScript)

	// 注册路由并启动服务器
	router := handlers.NewGinRouter(nil)
	server.RegisterRoutes(router)
	router.Engine().Run(":8080")
}
```

### 方法二：直接在HTML中注入脚本

如果需要在运行时动态注入，可以在HTML模板中添加script标签（需要修改模板文件）：

```html
<script src="static/app.js"></script>
<script>
	// 自定义逻辑
	if (window.SimpleDB && window.SimpleDB.config) {
		window.SimpleDB.config.requestInterceptor = function(url, options) {
			const token = getCookie('auth_token');
			if (token) {
				options.headers = options.headers || {};
				options.headers['Authorization'] = 'Bearer ' + token;
			}
			return { url, options };
		};
	}
</script>
```

## API说明

### 全局配置对象

所有配置都通过 `window.SimpleDBConfig` 或 `window.SimpleDB.config` 访问：

```javascript
window.SimpleDBConfig = {
	requestInterceptor: null,   // 请求拦截器
	responseInterceptor: null,   // 响应拦截器
	errorInterceptor: null      // 错误拦截器
}
```

### 请求拦截器

在发送请求前调用，可以修改请求配置：

```javascript
window.SimpleDB.config.requestInterceptor = function(url, options) {
	// url: 请求URL
	// options: fetch选项（包含method, headers, body等）
	
	// 修改headers
	options.headers = options.headers || {};
	options.headers['Custom-Header'] = 'value';
	
	// 必须返回 { url, options }
	return { url, options };
};
```

### 响应拦截器

在收到响应后调用，可以处理响应：

```javascript
window.SimpleDB.config.responseInterceptor = function(response) {
	// response: fetch Response对象
	
	// 检查状态码
	if (response.status === 401) {
		// token过期，跳转登录
		window.location.href = '/login';
	}
	
	// 必须返回response对象
	return response;
};
```

### 错误拦截器

在请求出错时调用，可以处理错误：

```javascript
window.SimpleDB.config.errorInterceptor = function(error, url, options) {
	// error: 错误对象
	// url: 请求URL
	// options: 请求选项
	
	// 记录错误日志
	console.error('API请求失败:', url, error);
	
	// 必须返回error对象
	return error;
};
```

## 完整示例

### 示例1：从Cookie读取Token

```javascript
function getCookie(name) {
	const value = '; ' + document.cookie;
	const parts = value.split('; ' + name + '=');
	if (parts.length === 2) return parts.pop().split(';').shift();
	return null;
}

window.SimpleDB.config.requestInterceptor = function(url, options) {
	const token = getCookie('auth_token');
	if (token) {
		options.headers = options.headers || {};
		options.headers['Authorization'] = 'Bearer ' + token;
	}
	return { url, options };
};
```

### 示例2：从LocalStorage读取Token

```javascript
window.SimpleDB.config.requestInterceptor = function(url, options) {
	const token = localStorage.getItem('auth_token');
	if (token) {
		options.headers = options.headers || {};
		options.headers['Authorization'] = 'Bearer ' + token;
	}
	return { url, options };
};
```

### 示例3：添加多个自定义Header

```javascript
window.SimpleDB.config.requestInterceptor = function(url, options) {
	options.headers = options.headers || {};
	options.headers['X-Request-ID'] = generateUUID();
	options.headers['X-Client-Version'] = '1.0.0';
	options.headers['X-Platform'] = navigator.platform;
	return { url, options };
};
```

### 示例4：统一处理401错误

```javascript
window.SimpleDB.config.responseInterceptor = function(response) {
	if (response.status === 401) {
		// 清除token并跳转登录
		document.cookie = 'auth_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
		window.location.href = '/login';
	}
	return response;
};
```

### 示例5：请求日志记录

```javascript
window.SimpleDB.config.requestInterceptor = function(url, options) {
	console.log('发送请求:', url, options);
	return { url, options };
};

window.SimpleDB.config.responseInterceptor = function(response) {
	console.log('收到响应:', response.status, response.url);
	return response;
};

window.SimpleDB.config.errorInterceptor = function(error, url, options) {
	console.error('请求失败:', url, error);
	// 可以发送错误到日志服务
	// sendErrorToLogService(error, url, options);
	return error;
};
```

## 注意事项

1. **拦截器必须返回正确的值**：
   - `requestInterceptor` 必须返回 `{ url, options }`
   - `responseInterceptor` 必须返回 `response` 对象
   - `errorInterceptor` 必须返回 `error` 对象

2. **执行顺序**：
   - 请求拦截器在发送请求前执行
   - 响应拦截器在收到响应后执行（无论成功或失败）
   - 错误拦截器只在请求出错时执行

3. **错误处理**：
   - 如果拦截器抛出异常，会被捕获并记录警告，不会影响正常请求流程

4. **线程安全**：
   - 在Go端使用 `SetCustomScript` 是线程安全的
   - 在JavaScript端配置拦截器是异步的，建议在页面加载完成后配置

## 更多示例

查看 `examples/custom_js_example.go` 获取完整的使用示例。

