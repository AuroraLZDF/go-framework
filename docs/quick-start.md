# 快速开始指南

本指南将帮助你在 5 分钟内使用 Framework 创建一个简单的 Web 应用。

## 前置要求

- Go 1.25+
- MySQL 5.7+（可选）
- Redis 6.0+（可选）

## Step 1: 创建项目

```bash
# 创建新项目目录
mkdir my-app && cd my-app

# 初始化 Go Module
go mod init my-app

# 添加 Framework 依赖
go get github.com/AuroraLZDF/go-framework@v0.2.0
```

## Step 2: 创建配置文件

创建 `config.yaml`（也可以直接复用 `examples/config.example.yaml`）：

```yaml
app:
  name: my-app
  version: 1.0.0
  mode: debug

server:
  http-addr: :8080
  read-header-timeout: 10s
  read-timeout: 30s
  write-timeout: 60s
  idle-timeout: 120s
  max-header-bytes: 1048576
  max-body-bytes: 4194304  # 请求体上限 4MB，超限返回 413
  trusted-proxies: []      # 默认不信任任何代理；按需填 IP/CIDR
  shutdown-timeout: 10s

database:
  mysql:
    host: 127.0.0.1:3306
    username: root
    password: root
    database: myapp_db
    max-idle-connections: 10
    max-open-connections: 100
    max-connection-life-time: 30s
    log-level: 4

redis:
  addr: 127.0.0.1
  port: "6379"
  password: ""
  db: 0
  pool-size: 10
  dial-timeout: 5s
  read-timeout: 3s
  write-timeout: 3s

jwt:
  secret: change-me-please-use-a-random-32byte-secret  # 至少 32 字节，Validate 会校验
  expire: 24  # 小时
  token-id: X-User-ID
  blacklist-prefix: "jwt:blacklist:"
  blacklist-fail-closed: false

log:
  level: info
  format: json
  dir: logs
```

> `database.mysql.host` 与 `redis.addr` 留空时，应用会跳过对应组件的初始化。

## Step 3: 创建主程序

创建 `main.go`:

```go
package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/AuroraLZDF/go-framework/core/config"
	"github.com/AuroraLZDF/go-framework/core/response"
	"github.com/AuroraLZDF/go-framework/core/server"
)

func main() {
	// 1. 加载配置（Load 使用独立 Viper 实例，不会写入全局 Viper）
	cfg, err := config.Load("config.yaml")
	if err != nil {
		panic(err)
	}

	// 2. 创建应用：NewApplication 会校验配置并自动初始化 JWT
	app, err := server.NewApplication(cfg)
	if err != nil {
		panic(err)
	}

	// 3. 注册路由
	app.RegisterRoutes(func(engine *gin.Engine, db *gorm.DB) {
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

	// 4. 启动应用（收到 SIGINT/SIGTERM 时优雅关闭）
	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
```

> 框架已内置 `GET /healthz`（存活探针）与 `GET /readyz`（就绪探针，检查已配置的
> MySQL/Redis，超时 2s）；请求体默认上限 4MB（`server.max-body-bytes`），声明
> Content-Length 超限时返回 413。需要鉴权的路由请显式挂载 `middleware.Authn()`，请求头格式为
> `Authorization: Bearer <token>`。

## Step 4: 运行应用

```bash
# 编译
go build -o my-app

# 运行
./my-app
```

你应该看到类似输出：

```
2026-11-04T12:20:35.000Z	INFO	Start to listening the incoming requests	{"addr": ":8080", "tls": false}
2026-11-04T12:20:35.000Z	INFO	Server exiting
```

## Step 5: 测试 API

```bash
# 存活探针（框架内置，仅进程存活）
curl http://localhost:8080/healthz

# 就绪探针（框架内置，检查 MySQL/Redis；未配置的依赖标记 skipped）
curl http://localhost:8080/readyz

# Hello World
curl http://localhost:8080/hello?name=Framework

# 获取用户列表
curl http://localhost:8080/api/v1/users
```

## 🎉 完成！

你已经成功创建了一个基于 Framework 的 Web 应用。

## 下一步

- 查看 [完整示例项目](../examples/simple-api)
- 阅读 [架构设计说明](architecture.md)
- 了解 [配置使用指南](CONFIG_USAGE.md) 与 [配置快速参考](CONFIG_QUICK_REFERENCE.md)

## 常见问题

### Q: 如何不使用数据库？

A: 省略 `database` 配置（或将 `database.mysql.host` 留空），框架会跳过 MySQL 初始化；Redis 同理。

### Q: 如何更改端口？

A: 修改配置文件中的 `server.http-addr` 字段，例如改为 `:3000`。

### Q: 如何启用 HTTPS？

A: 在配置文件中添加 TLS 证书路径（框架强制最低 TLS 1.2）：

```yaml
server:
  https-addr: :8443
  tls-cert: ./cert/server.crt
  tls-key: ./cert/server.key
```

- 仅配置 `https-addr` 时只监听 HTTPS，**不再隐式监听 `:8080`**；
- 如需保留 HTTP 入口，同时显式配置 `http-addr`：此时 HTTP 监听器返回
  `308` 永久重定向到 HTTPS（保留 path/query，非 443 端口会带在 Location 中）。

### Q: JWT 报错 "jwt secret must be at least 32 bytes"？

A: `config.Validate()` 要求 JWT secret 至少 32 字节；可用
`openssl rand -base64 32` 生成（结果为 44 个字符）。

### Q: 请求体过大返回什么？

A: 全局请求体上限由 `server.max-body-bytes` 控制（默认 4194304 = 4MB）。请求声明
`Content-Length` 且超限时返回 `HTTP 413` 与统一响应
`{"code":413,"message":"Request entity too large."}`；未声明长度的 chunked 请求体由
`http.MaxBytesReader` 截断，绑定错误可用 `middleware.IsRequestTooLarge(err)` 判断后
映射为 413。

### Q: 如何让 `c.ClientIP()` 识别反向代理传来的真实 IP？

A: 默认 `server.trusted-proxies: []` 不信任任何代理；请按实际代理网段配置
IP/CIDR（如 `["10.0.0.0/8"]`，或环境变量
`FRAMEWORK_SERVER_TRUSTED_PROXIES=10.0.0.0/8`），配置后才会解析
`X-Forwarded-For`。

### Q: 日志文件在哪里？

A: 默认输出到 stdout（`log.dir` 未配置时）。设置 `log.dir: logs` 后写入
`logs/<app-name>.YYYYMMDD.log`，并可按 `daily`/`hourly` 轮转。

## 需要帮助？

- 查看 [完整文档](../README.md)
- 提交 [Issue](https://github.com/AuroraLZDF/go-framework/issues)
- 加入社区讨论
