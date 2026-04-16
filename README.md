# Go Framework

一个基于 Gin 的高性能、可扩展的 Go Web 开发框架，提供完整的基础设施组件和最佳实践。

[![Release](https://img.shields.io/badge/release-v0.1.0--alpha-blue.svg)](https://github.com/yourorg/framework/releases)
[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)]()
[![Go Report Card](https://goreportcard.com/badge/github.com/yourorg/framework)](https://goreportcard.com/report/github.com/yourorg/framework)

## 🌟 特性

- ✅ **分层架构** - Controller → Biz → Store 清晰的分层设计
- ✅ **开箱即用** - 日志、数据库、认证、中间件等核心组件
- ✅ **灵活配置** - 基于 Viper 的配置管理，支持多环境和环境变量
- ✅ **JWT 认证** - 完整的 Token 签发、验证和黑名单机制
- ✅ **优雅关闭** - 支持信号处理和超时控制的服务关闭
- ✅ **统一响应** - 标准化的 JSON 响应格式和错误处理
- ✅ **易于扩展** - 接口抽象设计，轻松替换实现

## 📦 快速开始

### 方法一：使用脚手架工具（推荐）

```bash
# 克隆框架仓库
git clone https://github.com/yourorg/framework.git
cd framework

# 构建脚手架工具
make build

# 创建新项目
./bin/scaffold new my-app

# 进入项目目录
cd my-app

# 安装依赖
go mod tidy

# 配置项目
cp config.example.yaml config.yaml
# 编辑 config.yaml

# 运行应用
go run cmd/app/main.go
```

### 方法二：手动创建

#### 安装

```bash
go get go-framework@latest
```

### 最小示例

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
    // 加载配置
    cfg, _ := config.Load("config.yaml")
    
    // 创建应用
    app, _ := server.NewApplication(cfg)
    
    // 注册路由
    app.RegisterRoutes(func(engine *gin.Engine, db interface{}) {
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

database:
  mysql:
    host: 127.0.0.1:3306
    username: root
    password: root
    database: mydb

jwt:
  secret: your-secret-key
  expire: 24

log:
  level: info
  format: json
```

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
│   ├── internal/         # 内部实现
│   │   └── generator/    # 代码生成器
│   └── templates/        # 项目模板
├── examples/              # 示例项目
└── docs/                  # 文档
```

## 🚀 核心模块

### 1. 配置管理 (config)

```go
import "go-framework/core/config"

cfg, err := config.Load("config.yaml")
if err != nil {
    panic(err)
}

// 验证配置
if err := cfg.Validate(); err != nil {
    panic(err)
}
```

### 2. 应用服务器 (server)

```go
import "go-framework/core/server"

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

### 3. JWT 认证

```go
import "go-framework/core/auth/jwt"

// 初始化 JWT
jwt.Init(&jwt.Options{
    Secret:  "your-secret",
    Expire:  24,
    TokenID: "X-User-ID",
})

// 签发 Token
token, err := jwt.Sign("user-123")

// 验证 Token（在中间件中自动完成）
```

### 4. 数据库连接

```go
import "go-framework/core/database/db"

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
import "go-framework/core/log"

// 初始化日志
log.Init(&log.Options{
    Level:  "info",
    Format: "json",
    Dir:    "logs",
})

// 使用日志
log.Info("Server started", "port", 8080)
log.Error("Something failed", "err", err)
```

### 6. 统一响应

```go
import "go-framework/core/response"

// 成功响应
response.Success(c, data, "ok")

// 错误响应
response.Error(c, errno.ErrTokenInvalid)
```

### 7. 中间件

```go
import "go-framework/core/middleware"

// 认证中间件
v1.Use(middleware.Authn())

// 请求 ID 中间件（自动注册）
// CORS 中间件（自动注册）
```

## 📖 文档

- [快速开始指南](docs/quick-start.md)
- [架构设计说明](docs/architecture.md)
- [API 参考文档](docs/api-reference.md)
- [配置说明](docs/configuration.md)

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
- [GORM](https://github.com/go-gorm/gorm) - ORM 库
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Zap](https://github.com/uber-go/zap) - 日志库

---

**Star ⭐ 这个项目如果它对你有帮助！**
