# Go Framework

一个基于 Gin 的高性能、可扩展的 Go Web 开发框架，提供完整的基础设施组件和最佳实践。

[![Release](https://img.shields.io/badge/release-v0.2.0-blue.svg)](https://github.com/AuroraLZDF/go-framework/releases/tag/v0.2.0)
[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/AuroraLZDF/go-framework)](https://goreportcard.com/report/github.com/AuroraLZDF/go-framework)

## 🌟 特性

- ✅ **分层架构** - Controller → Biz → Store 清晰的分层设计
- ✅ **开箱即用** - 日志、数据库、认证、中间件等核心组件
- ✅ **灵活配置** - 基于 Viper 的配置管理，支持环境变量和自定义配置
- ✅ **JWT 认证** - 完整的 Token 签发、验证和黑名单机制
- ✅ **优雅关闭** - 支持信号处理和超时控制的服务关闭
- ✅ **健康检查** - 内置 `/healthz` 存活探针与 `/readyz` 就绪探针（MySQL/Redis）
- ✅ **安全默认** - 默认不信任任何代理、请求体上限 4MB、可选 HTTP→HTTPS 强制重定向
- ✅ **统一响应** - 标准化的 JSON 响应格式和错误处理
- ✅ **易于扩展** - 接口抽象设计，轻松替换实现

## 📦 快速开始

### 方法一：使用脚手架工具（推荐）

```bash
# 克隆框架仓库
git clone https://github.com/AuroraLZDF/go-framework.git
cd framework

# 构建脚手架工具
make build

# 创建新项目
./bin/scaffold new my-app

# 进入项目目录
cd my-app

# 安装依赖
go mod tidy

# 编辑生成的 config.yaml（已包含可直接启动的默认配置）
# 注意：JWT secret 至少 32 字节，config.Validate 会校验

# 运行应用
go run cmd/app/main.go
```

### 方法二：手动创建

#### 安装

```bash
go get github.com/AuroraLZDF/go-framework@v0.2.0
```

### 最小示例

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
    // 加载配置
    cfg, _ := config.Load("config.yaml")
    
    // 创建应用
    app, _ := server.NewApplication(cfg)
    
    // 注册路由
    app.RegisterRoutes(func(engine *gin.Engine, db *gorm.DB) {
        engine.GET("/hello", func(c *gin.Context) {
            response.Success(c, gin.H{"message": "Hello World"}, "ok")
        })
    })
    
    // 启动服务
    app.Run(context.Background())
}
```

### 配置文件示例

创建 `config.yaml`:

```yaml
app:
  name: my-app
  mode: debug

server:
  http-addr: :8080
  read-header-timeout: 10s
  read-timeout: 30s
  write-timeout: 60s
  idle-timeout: 120s
  max-header-bytes: 1048576
  max-body-bytes: 4194304        # 请求体上限 4MB，声明长度超限返回 413
  trusted-proxies: []            # 默认不信任任何代理；按需填 IP/CIDR

database:
  mysql:
    host: 127.0.0.1:3306
    username: root
    password: root
    database: mydb

jwt:
  secret: change-me-please-use-a-random-32byte-secret  # 至少 32 字节
  expire: 24
  token-id: X-User-ID
  blacklist-prefix: "jwt:blacklist:"
  blacklist-fail-closed: false

log:
  level: info
  format: json
  dir: logs
```

> HTTPS：仅配置 `https-addr` 时只监听 HTTPS，不再隐式监听 `:8080`；同时显式配置
> `http-addr` 与 `https-addr` 时，HTTP 监听器返回 `308` 永久重定向到 HTTPS（保留
> path/query）。`trusted-proxies` 为空时不会采信 `X-Forwarded-For`，
> `c.ClientIP()` 与访问日志直接使用 RemoteAddr。

## 📁 项目结构

```
framework/
├── core/                    # 框架核心
│   ├── auth/               # 认证模块
│   │   ├── jwt/           # JWT Token 管理
│   │   └── password/      # 密码加密
│   ├── config/            # 配置管理
│   ├── context/           # 上下文常量
│   ├── database/          # 数据库连接
│   ├── errno/             # 错误码定义
│   ├── log/               # 日志系统
│   ├── middleware/        # Gin 中间件
│   ├── response/          # HTTP 响应封装
│   ├── server/            # 应用服务器
│   ├── service/           # 服务接口定义
│   ├── util/              # 工具函数
│   ├── validator/         # 数据验证
│   └── version/           # 版本管理
├── scaffold/              # 脚手架工具
│   ├── cmd/              # 命令行入口
│   └── internal/         # 内部实现
│       └── generator/    # 代码生成器
├── examples/              # 示例项目
└── docs/                  # 文档
```

## 🚀 核心模块

### 1. 配置管理 (config)

Framework 提供强大的配置管理能力，支持 YAML 格式、环境变量覆盖和多环境配置。

#### 基础用法

```go
import "github.com/AuroraLZDF/go-framework/core/config"

cfg, err := config.Load("config.yaml")
if err != nil {
    panic(err)
}

// 验证配置
if err := cfg.Validate(); err != nil {
    panic(err)
}
```

#### 读取自定义配置

`config.Load` 使用独立的 Viper 实例，不会写入全局 Viper。推荐扩展 `config.Config`
结构（见 [自定义配置读取](docs/CUSTOM_CONFIG.md)）；也可以自行创建 Viper 实例读取同一文件：

```go
import "github.com/spf13/viper"

v := viper.New()
v.SetConfigFile("config.yaml")
if err := v.ReadInConfig(); err != nil {
    panic(err)
}

// config.yaml:
// custom:
//   api-key: abc123
apiKey := v.GetString("custom.api-key")
```

**📚 更多详情：**
- [配置使用指南](docs/CONFIG_USAGE.md) - 完整的配置加载和使用说明
- [自定义配置读取](docs/CUSTOM_CONFIG.md) - 如何读取自定义配置项
- [配置快速参考](docs/CONFIG_QUICK_REFERENCE.md) - 常用配置项速查

### 2. 应用服务器 (server)

```go
import "github.com/AuroraLZDF/go-framework/core/server"

app, err := server.NewApplication(cfg)
if err != nil {
    panic(err)
}

// 注册路由
app.RegisterRoutes(func(engine *gin.Engine, db *gorm.DB) {
    // 注册你的路由
})

// 启动服务
app.Run(context.Background())
```

框架自动注册两个探针：`GET /healthz`（liveness，仅进程存活）与 `GET /readyz`
（检查已配置的 MySQL/Redis，单次检查超时 2s；正常 200，失败 503）。

### 3. JWT 认证

```go
import (
    token "github.com/AuroraLZDF/go-framework/core/auth/jwt"
    "github.com/AuroraLZDF/go-framework/core/middleware"
)

// 初始化 JWT：Secret 至少 32 字节；server.NewApplication(cfg) 会自动完成此接线
if err := token.Init(&token.Options{
    Secret:  "change-me-please-use-a-random-32byte-secret", // ≥32 字节
    Expire:  24, // 单位：小时
    TokenId: "X-User-ID",
}); err != nil {
    panic(err)
}

// 签发 Token：参数必须是数字型用户 ID（Authn 会将其解析为 int64）；
// 每个 Token 均带 jti，响应中的 ExpiresIn 为剩余秒数
tokenResp, err := token.Sign("12345")

// 受保护路由需显式挂载认证中间件
// 请求头格式：Authorization: Bearer <token>
v1.Use(middleware.Authn())
```

### 4. 数据库连接

```go
import "github.com/AuroraLZDF/go-framework/core/database/db"

// MySQL
mysqlOpts := &db.MySQLOptions{
    Host:     "127.0.0.1:3306",
    Username: "root",
    Password: "root",
    Database: "mydb",
}
db, err := db.InitMySQL(mysqlOpts)

// Redis
redisOpts := &db.RedisOptions{
    Addr:     "127.0.0.1:6379",
    Password: "",
    DB:       0,
}
redis, err := db.InitRedis(redisOpts)
```

### 5. 日志系统

```go
import "github.com/AuroraLZDF/go-framework/core/log"

// 初始化日志：默认输出到 stdout；配置 Rotation 可写文件并启用轮转
if err := log.Init(&log.Options{
    Level:       "info",
    Format:      "json",
    OutputPaths: []string{"stdout"},
    Rotation: &log.RotateConfig{
        BasePath: "logs/app",
        Interval: "daily",
        Stdout:   true,
    },
}); err != nil {
    panic(err)
}

// 使用日志
log.Info("Server started", "port", 8080)
log.Error("Something failed", "err", err)
```

### 6. 统一响应

```go
import (
    "github.com/AuroraLZDF/go-framework/core/errno"
    "github.com/AuroraLZDF/go-framework/core/response"
)

// 成功响应
response.Success(c, data, "ok")

// 错误响应：HTTP 状态码与业务码由 errno 决定（如 401 / 404 / 500）
response.Error(c, errno.ErrTokenInvalid)
```

### 7. 中间件

```go
import "github.com/AuroraLZDF/go-framework/core/middleware"

// 认证中间件需显式挂载，请求头格式：Authorization: Bearer <token>
v1.Use(middleware.Authn())

// Request ID / CORS 中间件由 server.NewApplication 自动注册
```

## 📖 文档

- [快速开始指南](docs/quick-start.md)
- [架构设计说明](docs/architecture.md)
- [API 稳定性与版本策略](docs/api-stability.md)
- [从 0.x 迁移到 1.0](docs/migration-0.x-to-1.0.md)
- [配置使用指南](docs/CONFIG_USAGE.md)
- [配置快速参考](docs/CONFIG_QUICK_REFERENCE.md)
- [自定义配置](docs/CUSTOM_CONFIG.md)
- [文档索引](docs/DOCUMENTATION_GUIDE.md)
- [脚手架使用指南](scaffold/README.md)
- [脚手架快速开始](scaffold/QUICKSTART.md)

## 💡 最佳实践

### 1. 项目结构建议

```
your-project/
├── cmd/app/              # 应用入口
├── internal/
│   ├── biz/             # 业务逻辑层
│   ├── controller/      # 控制器层
│   ├── store/           # 数据访问层
│   ├── model/           # 数据模型
│   └── service/         # 外部服务实现
├── configs/             # 配置文件
├── sql/                 # 数据库脚本
└── assets/              # 静态资源
```

### 2. 分层架构

```
Controller (HTTP 层)
    ↓ 调用
Biz (业务逻辑层)
    ↓ 调用
Store (数据访问层)
    ↓ 使用
Framework Core (基础设施)
```

### 3. 依赖注入

```go
// 在 Controller 中注入 Biz
type UserController struct {
    userBiz biz.UserBiz
}

func NewUserController(biz biz.UserBiz) *UserController {
    return &UserController{userBiz: biz}
}
```

## 🔧 开发工具

### 代码格式化

```bash
gofmt -s -w .
```

### 依赖管理

```bash
go mod tidy
```

### 运行测试

```bash
go test ./...
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 👥 作者

- auroralzdf - 初始工作

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [GORM](https://gorm.io/gorm) - ORM 库
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Zap](https://github.com/uber-go/zap) - 日志库

---

**Star ⭐ 这个项目如果它对你有帮助！**
