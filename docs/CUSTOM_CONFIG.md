# 自定义配置读取指南

## 🎯 问题场景

当你在 `config.yaml` 中添加了 framework 未预定义的配置选项时，如何读取这些配置？

例如，你的配置文件：

```yaml
app:
  name: my-app
  version: 1.0.0

# 自定义配置 - framework 未定义
custom:
  api-key: abc123
  timeout: 30
  features:
    enable-cache: true
    max-retries: 3
    
third-party:
  wechat:
    app-id: wx123456
    app-secret: secret789
  alipay:
    app-id: alipay123
```

---

## ✅ 解决方案

### 方案一：使用 Viper 直接读取（最简单）⭐

Framework 内部使用 Viper 加载配置，你可以直接使用 Viper 读取任意配置项。

#### 步骤 1：导入 Viper

```go
import "github.com/spf13/viper"
```

#### 步骤 2：读取配置

```go
package main

import (
    "log"
    
    "github.com/AuroraLZDF/go-framework/core/config"
    "github.com/spf13/viper"
)

func main() {
    // 加载 framework 配置
    cfg, err := config.Load("./config/config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 使用 viper 读取自定义配置
    apiKey := viper.GetString("custom.api-key")
    timeout := viper.GetInt("custom.timeout")
    enableCache := viper.GetBool("custom.features.enable-cache")
    maxRetries := viper.GetInt("custom.features.max-retries")
    
    // 读取嵌套配置
    wechatAppId := viper.GetString("third-party.wechat.app-id")
    wechatSecret := viper.GetString("third-party.wechat.app-secret")
    
    log.Printf("API Key: %s", apiKey)
    log.Printf("Timeout: %d", timeout)
    log.Printf("WeChat AppID: %s", wechatAppId)
}
```

#### Viper 常用方法

```go
// 字符串
value := viper.GetString("key")

// 整数
value := viper.GetInt("key")

// 浮点数
value := viper.GetFloat64("key")

// 布尔值
value := viper.GetBool("key")

// 字符串切片
value := viper.GetStringSlice("key")

// 字符串映射
value := viper.GetStringMapString("key")

// 任意类型
value := viper.Get("key")

// 检查键是否存在
exists := viper.IsSet("key")
```

---

### 方案二：创建自定义配置结构体（推荐用于复杂配置）⭐⭐⭐

如果你的自定义配置比较复杂，建议创建专门的结构体。

#### 步骤 1：定义配置结构

```go
package config

// CustomConfig 自定义配置
type CustomConfig struct {
    APIKey   string          `mapstructure:"api-key"`
    Timeout  int             `mapstructure:"timeout"`
    Features FeatureConfig   `mapstructure:"features"`
}

// FeatureConfig 功能配置
type FeatureConfig struct {
    EnableCache bool `mapstructure:"enable-cache"`
    MaxRetries  int  `mapstructure:"max-retries"`
}

// ThirdPartyConfig 第三方服务配置
type ThirdPartyConfig struct {
    WeChat WeChatConfig `mapstructure:"wechat"`
    Alipay AlipayConfig `mapstructure:"alipay"`
}

// WeChatConfig 微信配置
type WeChatConfig struct {
    AppID     string `mapstructure:"app-id"`
    AppSecret string `mapstructure:"app-secret"`
}

// AlipayConfig 支付宝配置
type AlipayConfig struct {
    AppID string `mapstructure:"app-id"`
}
```

#### 步骤 2：扩展 Framework Config

在你的项目中创建扩展配置：

```go
package config

import (
    "github.com/AuroraLZDF/go-framework/core/config"
    "github.com/spf13/viper"
)

// ExtendedConfig 扩展配置
type ExtendedConfig struct {
    *config.Config              // 嵌入 framework 的 Config
    
    Custom     CustomConfig     `mapstructure:"custom"`
    ThirdParty ThirdPartyConfig `mapstructure:"third-party"`
}

// LoadExtended 加载扩展配置
func LoadExtended(configPath string) (*ExtendedConfig, error) {
    // 先加载 framework 配置
    baseCfg, err := config.Load(configPath)
    if err != nil {
        return nil, err
    }
    
    // 再加载自定义配置
    var extended ExtendedConfig
    extended.Config = baseCfg
    
    if err := viper.Unmarshal(&extended); err != nil {
        return nil, err
    }
    
    return &extended, nil
}
```

#### 步骤 3：使用扩展配置

```go
package main

import (
    "log"
    
    "your-project/config"  // 你的自定义配置包
)

func main() {
    // 加载扩展配置
    cfg, err := config.LoadExtended("./config/config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 访问 framework 配置
    log.Printf("App: %s", cfg.App.Name)
    log.Printf("Server: %s", cfg.Server.HTTPAddr)
    
    // 访问自定义配置
    log.Printf("API Key: %s", cfg.Custom.APIKey)
    log.Printf("Timeout: %d", cfg.Custom.Timeout)
    log.Printf("WeChat AppID: %s", cfg.ThirdParty.WeChat.AppID)
}
```

---

### 方案三：使用 Map 存储动态配置（最灵活）

如果配置项经常变化，可以使用 map 来存储。

```go
package main

import (
    "log"
    
    "github.com/AuroraLZDF/go-framework/core/config"
    "github.com/spf13/viper"
)

func main() {
    cfg, err := config.Load("./config/config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 获取整个 custom 节点作为 map
    customConfig := viper.GetStringMap("custom")
    log.Printf("Custom config: %+v", customConfig)
    
    // 获取 features 节点
    features := viper.GetStringMap("custom.features")
    log.Printf("Features: %+v", features)
    
    // 获取所有第三方配置
    thirdParty := viper.GetStringMap("third-party")
    log.Printf("Third party: %+v", thirdParty)
}
```

---

## 💡 实际示例

### 示例 1：读取业务配置

**config.yaml:**
```yaml
business:
  order:
    max-amount: 10000
    timeout-minutes: 30
    auto-cancel: true
    
  user:
    max-login-attempts: 5
    lock-duration-minutes: 15
```

**读取代码:**
```go
import "github.com/spf13/viper"

// 方式 1：直接读取
maxAmount := viper.GetInt("business.order.max-amount")
timeout := viper.GetInt("business.order.timeout-minutes")
autoCancel := viper.GetBool("business.order.auto-cancel")

// 方式 2：定义结构体
type OrderConfig struct {
    MaxAmount      int  `mapstructure:"max-amount"`
    TimeoutMinutes int  `mapstructure:"timeout-minutes"`
    AutoCancel     bool `mapstructure:"auto-cancel"`
}

var orderCfg OrderConfig
viper.UnmarshalKey("business.order", &orderCfg)
```

---

### 示例 2：读取第三方 API 配置

**config.yaml:**
```yaml
services:
  sms:
    provider: aliyun
    access-key: LTAI5t...
    access-secret: xxx
    sign-name: 我的应用
    
  email:
    smtp-host: smtp.qq.com
    smtp-port: 465
    username: noreply@example.com
    password: xxx
```

**读取代码:**
```go
import "github.com/spf13/viper"

// SMS 配置
smsProvider := viper.GetString("services.sms.provider")
accessKey := viper.GetString("services.sms.access-key")
signName := viper.GetString("services.sms.sign-name")

// Email 配置
smtpHost := viper.GetString("services.email.smtp-host")
smtpPort := viper.GetInt("services.email.smtp-port")
```

---

### 示例 3：读取功能开关配置

**config.yaml:**
```yaml
features:
  enable-registration: true
  enable-social-login: false
  maintenance-mode: false
  beta-features:
    - new-dashboard
    - ai-assistant
```

**读取代码:**
```go
import "github.com/spf13/viper"

// 布尔开关
enableReg := viper.GetBool("features.enable-registration")
maintenance := viper.GetBool("features.maintenance-mode")

// 列表
betaFeatures := viper.GetStringSlice("features.beta-features")
// betaFeatures = ["new-dashboard", "ai-assistant"]
```

---

## 🔧 高级技巧

### 1. 设置默认值

```go
import "github.com/spf13/viper"

// 在加载配置前设置默认值
viper.SetDefault("custom.timeout", 30)
viper.SetDefault("custom.features.enable-cache", true)
viper.SetDefault("custom.features.max-retries", 3)

// 然后加载配置
cfg, _ := config.Load("./config/config.yaml")

// 如果配置文件中没有这些值，将使用默认值
timeout := viper.GetInt("custom.timeout")  // 30
```

### 2. 监听配置变化（开发环境）

```go
import (
    "github.com/spf13/viper"
    "log"
)

// 监听配置文件变化
viper.WatchConfig()
viper.OnConfigChange(func(e fsnotify.Event) {
    log.Printf("Config file changed: %s", e.Name)
    
    // 重新读取配置
    newTimeout := viper.GetInt("custom.timeout")
    log.Printf("New timeout value: %d", newTimeout)
})
```

**注意**: Framework 当前不支持热重载，这个功能需要你自行实现。

### 3. 验证自定义配置

```go
func validateCustomConfig() error {
    apiKey := viper.GetString("custom.api-key")
    if apiKey == "" {
        return fmt.Errorf("custom.api-key is required")
    }
    
    timeout := viper.GetInt("custom.timeout")
    if timeout <= 0 || timeout > 300 {
        return fmt.Errorf("custom.timeout must be between 1 and 300")
    }
    
    return nil
}

// 使用
if err := validateCustomConfig(); err != nil {
    log.Fatalf("Invalid custom config: %v", err)
}
```

---

## 📊 方案对比

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| **Viper 直接读取** | 简单快速，无需修改代码 | 类型不安全，容易拼写错误 | 少量简单配置 |
| **自定义结构体** | 类型安全，IDE 支持，易维护 | 需要编写额外代码 | 复杂配置，团队项目 |
| **Map 存储** | 最灵活，支持动态配置 | 无类型检查，访问不便 | 配置项频繁变化 |

---

## ✅ 最佳实践

### 1. 组织配置文件结构

```yaml
# 按功能模块分组
app:
  # ...framework 配置

# 业务配置
business:
  order:
    # ...
  user:
    # ...

# 第三方服务
services:
  sms:
    # ...
  email:
    # ...

# 功能开关
features:
  # ...
```

### 2. 使用常量避免魔法字符串

```go
package config

// Viper 配置键常量
const (
    CustomAPIKey       = "custom.api-key"
    CustomTimeout      = "custom.timeout"
    WeChatAppID        = "third-party.wechat.app-id"
    WeChatAppSecret    = "third-party.wechat.app-secret"
)

// 使用
apiKey := viper.GetString(config.CustomAPIKey)
```

### 3. 提供配置示例

创建 `config.example.yaml` 包含所有可用的配置项和注释：

```yaml
# 自定义配置示例
custom:
  # API 密钥
  api-key: your-api-key-here
  
  # 超时时间（秒）
  timeout: 30
  
  # 功能开关
  features:
    enable-cache: true
    max-retries: 3

# 第三方服务配置
third-party:
  wechat:
    app-id: your-wechat-appid
    app-secret: your-wechat-secret
```

---

## ❓ 常见问题

### Q1: 应该选择哪种方案？

**A:** 
- 少量简单配置 → 方案一（Viper 直接读取）
- 复杂配置或团队项目 → 方案二（自定义结构体）
- 配置项频繁变化 → 方案三（Map）

### Q2: 可以将自定义配置添加到 framework 吗？

**A:** 可以！如果你的配置是通用的，可以提交 PR 到 framework 仓库。

### Q3: 环境变量能覆盖自定义配置吗？

**A:** 可以！使用相同的命名规则：

```bash
export FRAMEWORK_CUSTOM_API_KEY=my-key
export FRAMEWORK_THIRD_PARTY_WECHAT_APP_ID=wx123
```

### Q4: 如何调试配置是否正确加载？

**A:** 

```go
import (
    "encoding/json"
    "github.com/spf13/viper"
)

// 打印所有配置
allSettings := viper.AllSettings()
jsonData, _ := json.MarshalIndent(allSettings, "", "  ")
log.Printf("All config:\n%s", string(jsonData))
```

---

## 🚀 快速开始

对于大多数情况，推荐使用**方案一**：

```go
package main

import (
    "log"
    
    "github.com/AuroraLZDF/go-framework/core/config"
    "github.com/spf13/viper"
)

func main() {
    // 1. 加载 framework 配置
    cfg, err := config.Load("./config/config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 读取自定义配置
    apiKey := viper.GetString("custom.api-key")
    timeout := viper.GetInt("custom.timeout")
    
    // 3. 使用配置
    log.Printf("Using API key: %s, timeout: %d", apiKey, timeout)
}
```

就是这么简单！🎉

---

**更多详情**: 
- [Viper 官方文档](https://github.com/spf13/viper)
- [Framework 配置使用](CONFIG_USAGE.md)
