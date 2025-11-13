# Custom JavaScript Logic Usage Guide

This project provides a flexible extension mechanism that allows external projects to customize JavaScript logic, especially to add custom headers (such as reading tokens from cookies) to all AJAX requests.

## Features

1. **Request Interceptor**: Modify request configuration (URL, headers, body, etc.) before sending requests
2. **Response Interceptor**: Process responses after receiving them
3. **Error Interceptor**: Handle errors when requests fail
4. **Script Injection**: Inject custom JavaScript scripts through Go code

## Usage

### Method 1: Inject Script via Go Code (Recommended)

This is the most elegant approach, suitable for server-side configuration:

```go
package main

import (
	"log"
	"github.com/chenhg5/simple-db-web/handlers"
)

func main() {
	server, err := handlers.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Set custom script
	customScript := `
		// Read token from cookie and add to all request headers
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

	// Register routes and start server
	router := handlers.NewGinRouter(nil)
	server.RegisterRoutes(router)
	router.Engine().Run(":8080")
}
```

### Method 2: Inject Script Directly in HTML

If you need to inject dynamically at runtime, you can add a script tag in the HTML template (requires modifying the template file):

```html
<script src="static/app.js"></script>
<script>
	// Custom logic
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

## API Reference

### Global Configuration Object

All configurations are accessed through `window.SimpleDBConfig` or `window.SimpleDB.config`:

```javascript
window.SimpleDBConfig = {
	requestInterceptor: null,   // Request interceptor
	responseInterceptor: null,   // Response interceptor
	errorInterceptor: null      // Error interceptor
}
```

### Request Interceptor

Called before sending requests, can modify request configuration:

```javascript
window.SimpleDB.config.requestInterceptor = function(url, options) {
	// url: Request URL
	// options: fetch options (contains method, headers, body, etc.)
	
	// Modify headers
	options.headers = options.headers || {};
	options.headers['Custom-Header'] = 'value';
	
	// Must return { url, options }
	return { url, options };
};
```

### Response Interceptor

Called after receiving responses, can process responses:

```javascript
window.SimpleDB.config.responseInterceptor = function(response) {
	// response: fetch Response object
	
	// Check status code
	if (response.status === 401) {
		// Token expired, redirect to login
		window.location.href = '/login';
	}
	
	// Must return response object
	return response;
};
```

### Error Interceptor

Called when requests fail, can handle errors:

```javascript
window.SimpleDB.config.errorInterceptor = function(error, url, options) {
	// error: Error object
	// url: Request URL
	// options: Request options
	
	// Log error
	console.error('API request failed:', url, error);
	
	// Must return error object
	return error;
};
```

## Complete Examples

### Example 1: Read Token from Cookie

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

### Example 2: Read Token from LocalStorage

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

### Example 3: Add Multiple Custom Headers

```javascript
window.SimpleDB.config.requestInterceptor = function(url, options) {
	options.headers = options.headers || {};
	options.headers['X-Request-ID'] = generateUUID();
	options.headers['X-Client-Version'] = '1.0.0';
	options.headers['X-Platform'] = navigator.platform;
	return { url, options };
};
```

### Example 4: Unified 401 Error Handling

```javascript
window.SimpleDB.config.responseInterceptor = function(response) {
	if (response.status === 401) {
		// Clear token and redirect to login
		document.cookie = 'auth_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
		window.location.href = '/login';
	}
	return response;
};
```

### Example 5: Request Logging

```javascript
window.SimpleDB.config.requestInterceptor = function(url, options) {
	console.log('Sending request:', url, options);
	return { url, options };
};

window.SimpleDB.config.responseInterceptor = function(response) {
	console.log('Received response:', response.status, response.url);
	return response;
};

window.SimpleDB.config.errorInterceptor = function(error, url, options) {
	console.error('Request failed:', url, error);
	// Can send error to log service
	// sendErrorToLogService(error, url, options);
	return error;
};
```

## Notes

1. **Interceptors must return correct values**:
   - `requestInterceptor` must return `{ url, options }`
   - `responseInterceptor` must return `response` object
   - `errorInterceptor` must return `error` object

2. **Execution order**:
   - Request interceptor executes before sending request
   - Response interceptor executes after receiving response (whether successful or failed)
   - Error interceptor only executes when request fails

3. **Error handling**:
   - If an interceptor throws an exception, it will be caught and a warning will be logged, but it won't affect the normal request flow

4. **Thread safety**:
   - Using `SetCustomScript` on the Go side is thread-safe
   - Configuring interceptors on the JavaScript side is asynchronous, it's recommended to configure after page load

## More Examples

See `examples/custom_js_example.go` for complete usage examples.

