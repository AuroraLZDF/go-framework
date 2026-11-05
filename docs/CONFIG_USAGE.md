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
  shutdown-timeout: 10s          # 优雅关闭最长等待
  read-header-timeout: 10s
  read-timeout: 30s
  write-timeout: 60s
  idle-timeout: 120s
  max-header-bytes: 1048576      # 1MB
  max-body-bytes: 4194304        # 请求体上限 4MB，声明长度超限返回 HTTP 413
  trusted-proxies: []            # 可信代理 IP/CIDR；空 = 不信任任何代理

# 数据库配置
database:
  mysql:
    host: 127.0.0.1:3306
    username: root
    password: your_password
    database: myapp_db
    max-idle-connections: 10
    max-open-connections: 100
    max-connection-life-time: 30s
    log-level: 4  # 1: silent, 2: error, 3: warn, 4: info

# Redis 配置
redis:
  addr: 127.0.0.1
  port: 6379
  password: ""
  db: 0
  pool-size: 10
  dial-timeout: 5s
  read-timeout: 3s
  write-timeout: 3s

# JWT 配置
jwt:
  # 至少 32 字节，config.Validate() 会强制校验
  secret: change-me-please-use-a-random-32byte-secret
  expire: 24  # hours
  token-id: X-User-ID
  blacklist-prefix: "jwt:blacklist:"
  blacklist-fail-closed: false  # Redis 不可用时是否拒绝请求

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

// 超时与请求头限制
shutdownTimeout := cfg.Server.ShutdownTimeout
readHeaderTimeout := cfg.Server.ReadHeaderTimeout
readTimeout := cfg.Server.ReadTimeout
writeTimeout := cfg.Server.WriteTimeout
idleTimeout := cfg.Server.IdleTimeout
maxHeaderBytes := cfg.Server.MaxHeaderBytes

// 请求体上限（字节）与可信代理列表
maxBodyBytes := cfg.Server.MaxBodyBytes
trustedProxies := cfg.Server.TrustedProxies  // []string，默认空
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
redisPoolSize := cfg.Redis.PoolSize
redisDialTimeout := cfg.Redis.DialTimeout
redisReadTimeout := cfg.Redis.ReadTimeout
redisWriteTimeout := cfg.Redis.WriteTimeout
```

### 访问 JWT 配置

```go
jwtSecret := cfg.JWT.Secret
jwtExpire := cfg.JWT.Expire  // 小时
tokenId := cfg.JWT.TokenID
blacklistPrefix := cfg.JWT.BlacklistPrefix
blacklistFailClosed := cfg.JWT.BlacklistFailClosed
```

### 访问日志配置

```go
logLevel := cfg.Log.Level  // debug, info, warn, error
logFormat := cfg.Log.Format  // console, json
logDir := cfg.Log.Dir
disableCaller := cfg.Log.DisableCaller
disableStacktrace := cfg.Log.DisableStacktrace
outputPaths := cfg.Log.OutputPaths
rotation := cfg.Log.Rotation
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
export FRAMEWORK_JWT_SECRET=change-me-please-use-a-random-32byte-secret
export FRAMEWORK_SERVER_TRUSTED_PROXIES=10.0.0.0/8,192.168.0.0/16

# 运行应用
./my-app
```

> 所有内置配置键都支持环境变量覆盖；当 `config.Load("")` 找不到配置文件时，
> 应用会使用环境变量 + 默认值启动（JWT secret 仍必须 ≥32 字节）。

**环境变量命名规则：**
- 使用 `_` 代替 `.` 和 `-`
- 全部大写
- 添加 `FRAMEWORK_` 前缀

例如：
- `app.name` → `FRAMEWORK_APP_NAME`
- `server.http-addr` → `FRAMEWORK_SERVER_HTTP_ADDR`
- `server.trusted-proxies` → `FRAMEWORK_SERVER_TRUSTED_PROXIES`（逗号分隔）
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
- ✅ `app.mode` 必须是 debug/release/test 之一
- ✅ `log.format` 必须是 console/json 之一
- ✅ JWT secret 必填且至少 32 字节
- ✅ 配置了 `database.mysql.host` 时，username/database 不能为空
- ✅ 配置了 `server.https-addr` 时，必须同时提供 tls-cert 与 tls-key
- ✅ HTTP 超时与 max-header-bytes 必须为正数
- ✅ `server.max-body-bytes` 必须为正数（0 视为未配置，默认 4MB）
- ✅ `server.trusted-proxies` 每项必须是合法 IP 或 CIDR（自动去空白、丢弃空项）
- ✅ `redis.addr` 支持空、`host`、`host:port`、裸 IPv6、`[IPv6]`、`[IPv6]:port`；
  拒绝带 scheme（如 `redis://`）或多余冒号的写法
- ✅ Server HTTPAddr 默认 `:8080`（由 applyDefaults 填充；仅配置 https-addr 时不再补 HTTP 地址）

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

## 🔐 HTTP/HTTPS 监听与健康检查

### 监听语义

- **仅配置 `server.https-addr`**：只启动 HTTPS 监听器，不再隐式监听 `:8080`
  （相对 v0.1.3 之前的行为变更）。
- **同时显式配置 `server.http-addr` 与 `server.https-addr`**：HTTP 监听器对请求返回
  `308 Permanent Redirect` 到 HTTPS，保留 path/query；HTTPS 端口不是 443 时
  Location 中带该端口。优雅关闭同时覆盖两个监听器。
- **`server.trusted-proxies` 为空（默认）**：不信任任何代理，`c.ClientIP()` 与访问日志
  的 client_ip 直接取 RemoteAddr；配置 IP/CIDR 后才解析 `X-Forwarded-For`。
  每项会去除首尾空白；配置 `0.0.0.0/0`、`::/0` 等全地址范围时启动会打印 WARNING
  （等价于信任任意来源的 XFF，仅应在明确知道后果时使用）。
- **重定向的 Host 处理**：HTTP→HTTPS 重定向沿用请求的 Host 拼接 Location。
  部署在反向代理之后时，应由代理清洗 Host 头；框架为 308 响应设置
  `Cache-Control: no-store`，降低共享缓存被投毒的风险。
- **`server.max-body-bytes`**：全局请求体上限（默认 4194304 = 4MB）。请求声明
  `Content-Length` 且超限时返回 `HTTP 413` 与统一响应
  `{"code":413,"message":"Request entity too large."}`（错误响应省略 content）；
  未声明长度的 chunked 请求体由 `http.MaxBytesReader` 截断，超出部分读取报错，
  业务绑定通常返回 400，但内存占用同样被限制在上限内。

### 健康检查

| 端点 | 用途 | 行为 |
|------|------|------|
| `GET /healthz` | liveness | 仅确认进程存活，不检查依赖，返回 200 |
| `GET /readyz` | readiness | 检查已配置的 MySQL/Redis（各自 Ping，超时 2s）|

`/readyz` 正常示例：

```json
{
  "code": 200,
  "message": "ok",
  "content": {
    "status": "ok",
    "checks": {"mysql": "ok", "redis": "skipped"}
  }
}
```

未配置的依赖标记为 `skipped`；任一依赖失败返回 `HTTP 503` 与统一响应
`{"code":503,"message":"Service unavailable.","content":{...}}`。

---

## 🚀 完整示例

### main.go

```go
package main

import (
    "context"
    "log"
    
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    "github.com/AuroraLZDF/go-framework/core/config"
    "github.com/AuroraLZDF/go-framework/core/response"
    "github.com/AuroraLZDF/go-framework/core/server"
)

func main() {
    // 1. 加载配置
    cfg, err := config.Load("./config.yaml")
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
    app.RegisterRoutes(func(engine *gin.Engine, db *gorm.DB) {
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
