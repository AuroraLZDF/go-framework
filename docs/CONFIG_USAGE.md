# Framework 配置使用指南

## 📋 快速开始

### 方法一：使用默认配置路径（推荐）

```go
package main

import (
    "log"
    
    "github.com/AuroraLZDF/go-framework/core/config"
)

func main() {
    // 不传参数，自动查找 config.yaml
    cfg, err := config.Load("")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 验证配置
    if err := cfg.Validate(); err != nil {
        log.Fatalf("Invalid config: %v", err)
    }
    
    // 使用配置
    log.Printf("App name: %s", cfg.App.Name)
    log.Printf("Server addr: %s", cfg.Server.HTTPAddr)
}
```

**搜索路径优先级：**
1. `./config.yaml` - 当前目录
2. `$HOME/.framework/config.yaml` - 用户主目录

---

### 方法二：指定配置文件路径

```go
package main

import (
    "log"
    
    "github.com/AuroraLZDF/go-framework/core/config"
)

func main() {
    // 指定配置文件路径
    cfg, err := config.Load("./config/config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 或者使用绝对路径
    // cfg, err := config.Load("/etc/myapp/config.yaml")
    
    log.Printf("Config loaded successfully")
}
```

---

## 📂 配置文件结构

### 完整的 config.yaml 示例

```yaml
# 应用配置
app:
  name: my-application
  version: 1.0.0
  mode: debug  # debug, release, test

# 服务器配置
server:
  http-addr: :8080
  https-addr: :8443
  tls-cert: ./cert/server.crt
  tls-key: ./cert/server.key

# 数据库配置
database:
  mysql:
    host: 127.0.0.1:3306
    username: root
    password: your_password
    database: myapp_db
    max-idle-connections: 100
    max-open-connections: 20
    max-connection-life-time: 30s
    log-level: 4  # 1: silent, 2: error, 3: warn, 4: info

# Redis 配置
redis:
  addr: 127.0.0.1
  port: 6379
  password: ""
  db: 0

# JWT 配置
jwt:
  secret: change-this-to-a-random-secret-key
  expire: 24  # hours
  token-id: X-User-ID
  blacklist-path: tmp/blacklist.log

# 日志配置
log:
  disable-caller: false
  disable-stacktrace: false
  level: info  # debug, info, warn, error
  format: json  # console, json
  dir: logs
```

---

## 🔧 访问配置项

### 访问应用配置

```go
// 应用名称
appName := cfg.App.Name

// 应用版本
version := cfg.App.Version

// 运行模式
mode := cfg.App.Mode  // debug, release, test
```

### 访问服务器配置

```go
// HTTP 地址
httpAddr := cfg.Server.HTTPAddr  // ":8080"

// HTTPS 地址
httpsAddr := cfg.Server.HTTPSAddr  // ":8443"

// TLS 证书
tlsCert := cfg.Server.TLSCert
tlsKey := cfg.Server.TLSKey
```

### 访问数据库配置

```go
// MySQL 配置
mysqlHost := cfg.Database.MySQL.Host
mysqlUser := cfg.Database.MySQL.Username
mysqlPass := cfg.Database.MySQL.Password
mysqlDB := cfg.Database.MySQL.Database
maxIdle := cfg.Database.MySQL.MaxIdleConnections
maxOpen := cfg.Database.MySQL.MaxOpenConnections
```

### 访问 Redis 配置

```go
redisAddr := cfg.Redis.Addr
redisPort := cfg.Redis.Port
redisPass := cfg.Redis.Password
redisDB := cfg.Redis.DB
```

### 访问 JWT 配置

```go
jwtSecret := cfg.JWT.Secret
jwtExpire := cfg.JWT.Expire  // 小时
tokenId := cfg.JWT.TokenID
blacklistPath := cfg.JWT.BlacklistPath
```

### 访问日志配置

```go
logLevel := cfg.Log.Level  // debug, info, warn, error
logFormat := cfg.Log.Format  // console, json
logDir := cfg.Log.Dir
disableCaller := cfg.Log.DisableCaller
disableStacktrace := cfg.Log.DisableStacktrace
```

---

## 💡 高级用法

### 1. 使用环境变量覆盖配置

Framework 支持通过环境变量覆盖配置文件中的值：

```bash
# 设置环境变量（前缀为 FRAMEWORK）
export FRAMEWORK_APP_NAME=my-app
export FRAMEWORK_SERVER_HTTP_ADDR=:3000
export FRAMEWORK_DATABASE_MYSQL_HOST=192.168.1.100:3306
export FRAMEWORK_JWT_SECRET=my-secret-key

# 运行应用
./my-app
```

**环境变量命名规则：**
- 使用 `_` 代替 `.` 和 `-`
- 全部大写
- 添加 `FRAMEWORK_` 前缀

例如：
- `app.name` → `FRAMEWORK_APP_NAME`
- `server.http-addr` → `FRAMEWORK_SERVER_HTTP_ADDR`
- `database.mysql.host` → `FRAMEWORK_DATABASE_MYSQL_HOST`

---

### 2. 多环境配置

创建不同环境的配置文件：

```
config/
├── config.dev.yaml      # 开发环境
├── config.test.yaml     # 测试环境
└── config.prod.yaml     # 生产环境
```

在代码中根据环境变量加载：

```go
package main

import (
    "log"
    "os"
    
    "github.com/AuroraLZDF/go-framework/core/config"
)

func main() {
    // 获取环境
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "dev"
    }
    
    // 加载对应环境的配置
    configPath := "./config/config." + env + ".yaml"
    cfg, err := config.Load(configPath)
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    log.Printf("Running in %s mode", env)
}
```

运行不同环境：

```bash
# 开发环境
APP_ENV=dev ./my-app

# 测试环境
APP_ENV=test ./my-app

# 生产环境
APP_ENV=prod ./my-app
```

---

### 3. 配置验证

在加载配置后进行验证：

```go
cfg, err := config.Load("./config/config.yaml")
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// 验证配置
if err := cfg.Validate(); err != nil {
    log.Fatalf("Invalid configuration: %v", err)
}

log.Println("Configuration is valid")
```

**内置验证规则：**
- ✅ MySQL host 不能为空
- ✅ MySQL username 不能为空
- ✅ MySQL database 不能为空
- ✅ JWT secret 不能为空
- ✅ Server HTTPAddr 如果为空，默认为 `:8080`

---

### 4. 自定义配置验证

你可以添加自己的验证逻辑：

```go
func validateCustomConfig(cfg *config.Config) error {
    // 验证端口范围
    if cfg.Server.HTTPAddr != "" {
        // 添加你的验证逻辑
    }
    
    // 验证数据库连接池大小
    if cfg.Database.MySQL.MaxOpenConnections <= 0 {
        return fmt.Errorf("max open connections must be positive")
    }
    
    // 验证 JWT 过期时间
    if cfg.JWT.Expire <= 0 || cfg.JWT.Expire > 168 {
        return fmt.Errorf("JWT expire must be between 1 and 168 hours")
    }
    
    return nil
}

// 使用
if err := validateCustomConfig(cfg); err != nil {
    log.Fatalf("Custom validation failed: %v", err)
}
```

---

## 🚀 完整示例

### main.go

```go
package main

import (
    "context"
    "log"
    
    "github.com/gin-gonic/gin"
    
    "github.com/AuroraLZDF/go-framework/core/config"
    "github.com/AuroraLZDF/go-framework/core/response"
    "github.com/AuroraLZDF/go-framework/core/server"
)

func main() {
    // 1. 加载配置
    cfg, err := config.Load("./config/config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 2. 验证配置
    if err := cfg.Validate(); err != nil {
        log.Fatalf("Invalid config: %v", err)
    }
    
    // 3. 打印配置信息（调试用）
    log.Printf("App: %s v%s", cfg.App.Name, cfg.App.Version)
    log.Printf("Mode: %s", cfg.App.Mode)
    log.Printf("Server: %s", cfg.Server.HTTPAddr)
    log.Printf("Database: %s@%s/%s", 
        cfg.Database.MySQL.Username,
        cfg.Database.MySQL.Host,
        cfg.Database.MySQL.Database)
    
    // 4. 创建应用
    app, err := server.NewApplication(cfg)
    if err != nil {
        log.Fatalf("Failed to create application: %v", err)
    }
    
    // 5. 注册路由
    app.RegisterRoutes(func(engine *gin.Engine, db interface{}) {
        // 健康检查
        engine.GET("/health", func(c *gin.Context) {
            response.Success(c, gin.H{
                "status": "healthy",
                "app": cfg.App.Name,
                "version": cfg.App.Version,
            }, "ok")
        })
        
        // 显示配置信息（仅用于演示）
        engine.GET("/config", func(c *gin.Context) {
            response.Success(c, gin.H{
                "app_name": cfg.App.Name,
                "mode": cfg.App.Mode,
                "server_addr": cfg.Server.HTTPAddr,
            }, "ok")
        })
    })
    
    // 6. 启动应用
    log.Printf("Starting %s on %s...", cfg.App.Name, cfg.Server.HTTPAddr)
    if err := app.Run(context.Background()); err != nil {
        log.Fatalf("Application failed: %v", err)
    }
}
```

---

## ❓ 常见问题

### Q1: 如何在不重启应用的情况下重新加载配置？

A: 目前 framework 不支持热重载。你需要重启应用来加载新配置。

### Q2: 配置文件必须是 YAML 格式吗？

A: 是的，目前只支持 YAML 格式。

### Q3: 如何加密配置文件中的敏感信息（如密码）？

A: 建议使用环境变量：

```yaml
database:
  mysql:
    password: ${DB_PASSWORD}  # 从环境变量读取
```

然后设置环境变量：
```bash
export FRAMEWORK_DATABASE_MYSQL_PASSWORD=your_password
```

### Q4: 配置文件的搜索顺序是什么？

A: 
1. 如果传入了路径，使用该路径
2. 否则按以下顺序搜索：
   - `./config.yaml`
   - `$HOME/.framework/config.yaml`

### Q5: 如何查看当前加载的配置？

A: 可以打印配置对象：

```go
import "encoding/json"

// 转换为 JSON 查看
jsonData, _ := json.MarshalIndent(cfg, "", "  ")
log.Printf("Current config:\n%s", string(jsonData))
```

---

## 📚 相关文档

- [配置结构定义](../core/config/options.go)
- [配置加载器实现](../core/config/loader.go)
- [配置示例](../examples/config.example.yaml)

---

**Happy Coding!** 🎉
