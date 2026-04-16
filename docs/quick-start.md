# 快速开始指南

本指南将帮助你在 5 分钟内使用 Framework 创建一个简单的 Web 应用。

## 前置要求

- Go 1.23+
- MySQL 5.7+（可选）
- Redis 6.0+（可选）

## Step 1: 创建项目

```bash
# 创建新项目目录
mkdir my-app && cd my-app

# 初始化 Go Module
go mod init my-app

# 添加 Framework 依赖
go get go-framework@latest
```

## Step 2: 创建配置文件

创建 `config.yaml`:

```yaml
app:
  name: my-app
  version: 1.0.0
  mode: debug

server:
  http-addr: :8080

database:
  mysql:
    host: 127.0.0.1:3306
    username: root
    password: root
    database: myapp_db
    max-idle-connections: 100
    max-open-connections: 20
    max-connection-life-time: 30s
    log-level: 4

jwt:
  secret: your-secret-key-change-in-production
  expire: 24
  token-id: X-User-ID

log:
  level: info
  format: json
  dir: logs
```

## Step 3: 创建主程序

创建 `main.go`:

```go
package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"go-framework/core/config"
	"go-framework/core/response"
	"go-framework/core/server"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		panic(err)
	}

	// 2. 创建应用
	app, err := server.NewApplication(cfg)
	if err != nil {
		panic(err)
	}

	// 3. 注册路由
	app.RegisterRoutes(func(engine *gin.Engine, db interface{}) {
		// 健康检查
		engine.GET("/health", func(c *gin.Context) {
			response.Success(c, gin.H{"status": "ok"}, "healthy")
		})

		// Hello World
		engine.GET("/hello", func(c *gin.Context) {
			name := c.DefaultQuery("name", "World")
			response.Success(c, gin.H{
				"message": "Hello, " + name + "!",
			}, "ok")
		})

		// API v1 路由组
		v1 := engine.Group("/api/v1")
		{
			v1.GET("/users", func(c *gin.Context) {
				response.Success(c, []gin.H{
					{"id": 1, "name": "Alice", "email": "alice@example.com"},
					{"id": 2, "name": "Bob", "email": "bob@example.com"},
				}, "ok")
			})
		}
	})

	// 4. 启动应用
	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
```

## Step 4: 运行应用

```bash
# 编译
go build -o my-app

# 运行
./my-app
```

你应该看到类似输出：

```
2024-01-01T00:00:00.000Z INFO Application initialized successfully
2024-01-01T00:00:00.000Z INFO Database connected successfully
2024-01-01T00:00:00.000Z INFO Start to listening the incoming requests on http address :8080
```

## Step 5: 测试 API

```bash
# 健康检查
curl http://localhost:8080/health

# Hello World
curl http://localhost:8080/hello?name=Framework

# 获取用户列表
curl http://localhost:8080/api/v1/users
```

## 🎉 完成！

你已经成功创建了一个基于 Framework 的 Web 应用。

## 下一步

- 学习如何使用 [JWT 认证](../examples/auth-example)
- 了解 [数据库操作](../examples/database-example)
- 查看 [完整示例项目](../examples/simple-api)
- 阅读 [架构设计文档](architecture.md)

## 常见问题

### Q: 如何不使用数据库？

A: 如果不需要数据库，可以在配置中省略 `database` 部分，或者在代码中不调用数据库相关功能。

### Q: 如何更改端口？

A: 修改配置文件中的 `server.http-addr` 字段，例如改为 `:3000`。

### Q: 如何启用 HTTPS？

A: 在配置文件中添加 TLS 证书路径：

```yaml
server:
  https-addr: :8443
  tls-cert: ./cert/server.crt
  tls-key: ./cert/server.key
```

### Q: 日志文件在哪里？

A: 默认在 `logs/` 目录下，按日期分割，例如 `logs/2024-01-01.log`。

## 需要帮助？

- 查看 [完整文档](../README.md)
- 提交 [Issue](https://github.com/yourorg/framework/issues)
- 加入社区讨论
