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
cfg.Server.HTTPAddr      // HTTP 地址，如 ":8080"
cfg.Server.HTTPSAddr     // HTTPS 地址
cfg.Server.TLSCert       // TLS 证书路径
cfg.Server.TLSKey        // TLS 私钥路径

// MySQL
cfg.Database.MySQL.Host               // 主机地址
cfg.Database.MySQL.Username           // 用户名
cfg.Database.MySQL.Password           // 密码
cfg.Database.MySQL.Database           // 数据库名
cfg.Database.MySQL.MaxIdleConnections // 最大空闲连接
cfg.Database.MySQL.MaxOpenConnections // 最大打开连接

// Redis
cfg.Redis.Addr     // 地址
cfg.Redis.Port     // 端口
cfg.Redis.Password // 密码
cfg.Redis.DB       // 数据库编号

// JWT
cfg.JWT.Secret         // 密钥
cfg.JWT.Expire         // 过期时间（小时）
cfg.JWT.TokenID        // Token ID 字段名
cfg.JWT.BlacklistPath  // 黑名单文件路径

// 日志
cfg.Log.Level             // 日志级别 (debug/info/warn/error)
cfg.Log.Format            // 格式 (console/json)
cfg.Log.Dir               // 日志目录
cfg.Log.DisableCaller     // 是否禁用调用者信息
cfg.Log.DisableStacktrace // 是否禁用堆栈跟踪
```

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
export FRAMEWORK_JWT_SECRET=my-jwt-secret

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
- ✅ 支持环境变量自动覆盖
- ✅ 内置基本验证
- ✅ 可以自定义验证逻辑
- ❌ 不支持热重载（需要重启应用）

---

**更多详情**: [CONFIG_USAGE.md](CONFIG_USAGE.md)
