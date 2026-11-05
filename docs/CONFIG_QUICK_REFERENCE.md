# Framework 配置快速参考

## 🚀 三种加载方式

### 方式 1：默认路径（最简单）

```go
cfg, err := config.Load("")
// 自动查找：./config.yaml 或 $HOME/.framework/config.yaml
```

### 方式 2：指定相对路径

```go
cfg, err := config.Load("./config/config.yaml")
cfg, err := config.Load("config.yaml")
```

### 方式 3：指定绝对路径

```go
cfg, err := config.Load("/etc/myapp/config.yaml")
```

---

## 📋 常用配置项

```go
// 应用信息
cfg.App.Name        // 应用名称
cfg.App.Version     // 版本号
cfg.App.Mode        // 运行模式 (debug/release/test)

// 服务器
cfg.Server.HTTPAddr          // HTTP 地址，如 ":8080"
cfg.Server.HTTPSAddr         // HTTPS 地址
cfg.Server.TLSCert           // TLS 证书路径
cfg.Server.TLSKey            // TLS 私钥路径
cfg.Server.ShutdownTimeout   // 优雅关闭超时（默认 10s）
cfg.Server.ReadHeaderTimeout // 读请求头超时（默认 10s）
cfg.Server.ReadTimeout       // 读超时（默认 30s）
cfg.Server.WriteTimeout      // 写超时（默认 60s）
cfg.Server.IdleTimeout       // 空闲超时（默认 120s）
cfg.Server.MaxHeaderBytes    // 请求头大小上限（默认 1MB）
cfg.Server.MaxBodyBytes      // 请求体上限（默认 4MB，超限 413）
cfg.Server.TrustedProxies    // 可信代理 []string（默认空 = 不信任任何代理）

// MySQL
cfg.Database.MySQL.Host               // 主机地址
cfg.Database.MySQL.Username           // 用户名
cfg.Database.MySQL.Password           // 密码
cfg.Database.MySQL.Database           // 数据库名
cfg.Database.MySQL.MaxIdleConnections // 最大空闲连接
cfg.Database.MySQL.MaxOpenConnections // 最大打开连接

// Redis
cfg.Redis.Addr        // 地址（留空则跳过 Redis 初始化）
cfg.Redis.Port        // 端口
cfg.Redis.Password    // 密码
cfg.Redis.DB          // 数据库编号
cfg.Redis.PoolSize    // 连接池大小（默认 10）
cfg.Redis.DialTimeout // 连接超时（默认 5s）
cfg.Redis.ReadTimeout // 读超时（默认 3s）
cfg.Redis.WriteTimeout// 写超时（默认 3s）

// JWT
cfg.JWT.Secret            // 密钥，至少 32 字节（Validate 强制校验）
cfg.JWT.Expire            // 过期时间（小时）
cfg.JWT.TokenID           // Token 中存储用户 ID 的 claim 名
cfg.JWT.BlacklistPrefix   // 黑名单 Redis key 前缀（默认 jwt:blacklist:）
cfg.JWT.BlacklistFailClosed // Redis 不可用时是否拒绝请求（默认 false）

// 日志
cfg.Log.Level             // 日志级别 (debug/info/warn/error)
cfg.Log.Format            // 格式 (console/json)
cfg.Log.Dir               // 日志目录
cfg.Log.DisableCaller     // 是否禁用调用者信息
cfg.Log.DisableStacktrace // 是否禁用堆栈跟踪
cfg.Log.OutputPaths       // 额外输出目标（如 "stdout"）
cfg.Log.Rotation          // 轮转配置（*log.RotateConfig）
```

> 校验规则：`app.mode` ∈ debug/release/test；`log.format` ∈ console/json；
> JWT secret 至少 32 字节；MySQL 仅在配置了 `host` 时校验用户名/库名；
> HTTPS 只配置地址时要求同时提供证书与私钥；`max-body-bytes` 必须为正数；
> `trusted-proxies` 每项必须是合法 IP 或 CIDR。

---

## ✅ 最佳实践

### 1. 总是验证配置

```go
cfg, err := config.Load("./config/config.yaml")
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

if err := cfg.Validate(); err != nil {
    log.Fatalf("Invalid config: %v", err)
}
```

### 2. 使用环境变量覆盖敏感信息

```bash
# 设置环境变量
export FRAMEWORK_DATABASE_MYSQL_PASSWORD=secret
export FRAMEWORK_JWT_SECRET=change-me-please-use-a-random-32byte-secret
export FRAMEWORK_SERVER_TRUSTED_PROXIES=10.0.0.0/8,192.168.0.0/16

# 运行时自动覆盖配置文件中的值
```

### 3. 多环境配置

```go
env := os.Getenv("APP_ENV")
if env == "" {
    env = "dev"
}

configPath := fmt.Sprintf("./config/config.%s.yaml", env)
cfg, err := config.Load(configPath)
```

---

## 🔧 完整示例

```go
package main

import (
    "context"
    "log"
    
    "github.com/AuroraLZDF/go-framework/core/config"
    "github.com/AuroraLZDF/go-framework/core/server"
)

func main() {
    // 加载配置
    cfg, err := config.Load("./config/config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 验证配置
    if err := cfg.Validate(); err != nil {
        log.Fatalf("Invalid config: %v", err)
    }
    
    // 打印配置信息
    log.Printf("App: %s v%s (%s)", cfg.App.Name, cfg.App.Version, cfg.App.Mode)
    log.Printf("Server: %s", cfg.Server.HTTPAddr)
    log.Printf("Database: %s@%s/%s", 
        cfg.Database.MySQL.Username,
        cfg.Database.MySQL.Host,
        cfg.Database.MySQL.Database)
    
    // 创建并启动应用
    app, err := server.NewApplication(cfg)
    if err != nil {
        log.Fatalf("Failed to create app: %v", err)
    }
    
    app.Run(context.Background())
}
```

---

## 💡 提示

- ✅ 配置文件必须是 YAML 格式
- ✅ 支持环境变量自动覆盖（默认监听全部内置配置键）
- ✅ 无配置文件时可用环境变量 + 默认值启动（`Load("")`）
- ✅ 内置基本验证（含 JWT secret ≥32 字节）
- ✅ 可以自定义验证逻辑
- ✅ 仅配置 `https-addr` 时只监听 HTTPS，不再隐式监听 `:8080`；同时配置 `http-addr`
  与 `https-addr` 时，HTTP 监听器返回 308 重定向到 HTTPS
- ✅ 健康检查：`/healthz`（liveness）与 `/readyz`（检查 MySQL/Redis，超时 2s，失败 503）
- ✅ 请求体默认上限 4MB（`server.max-body-bytes`），声明 Content-Length 超限返回 413；
  未知长度的 chunked 请求体由 `http.MaxBytesReader` 截断（可用 `middleware.IsRequestTooLarge` 判定）
- ❌ 不支持热重载（需要重启应用）

---

**更多详情**: [CONFIG_USAGE.md](CONFIG_USAGE.md)
