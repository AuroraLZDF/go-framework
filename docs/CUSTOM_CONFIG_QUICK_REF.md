# 自定义配置快速参考

## 🚀 三种读取方式

### 方式 1：Viper 直接读取（最简单）⭐

```go
import "github.com/spf13/viper"

// config.yaml:
// custom:
//   api-key: abc123
//   timeout: 30

apiKey := viper.GetString("custom.api-key")
timeout := viper.GetInt("custom.timeout")
enableCache := viper.GetBool("custom.features.enable-cache")
```

**常用方法：**
- `viper.GetString("key")` - 字符串
- `viper.GetInt("key")` - 整数
- `viper.GetFloat64("key")` - 浮点数
- `viper.GetBool("key")` - 布尔值
- `viper.GetStringSlice("key")` - 字符串数组
- `viper.GetStringMap("key")` - Map
- `viper.IsSet("key")` - 检查是否存在

---

### 方式 2：自定义结构体（类型安全）⭐⭐⭐

```go
// 1. 定义结构体
type CustomConfig struct {
    APIKey  string `mapstructure:"api-key"`
    Timeout int    `mapstructure:"timeout"`
}

// 2. 反序列化
var custom CustomConfig
viper.UnmarshalKey("custom", &custom)

// 3. 使用
log.Printf("API Key: %s", custom.APIKey)
```

---

### 方式 3：Map 存储（最灵活）

```go
// 获取整个节点
customMap := viper.GetStringMap("custom")
log.Printf("Custom: %+v", customMap)

// 访问具体字段
apiKey := customMap["api-key"].(string)
```

---

## 💡 完整示例

### config.yaml

```yaml
app:
  name: my-app
  version: 1.0.0

# 自定义配置
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
```

### main.go

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
    enableCache := viper.GetBool("custom.features.enable-cache")
    
    wechatAppId := viper.GetString("third-party.wechat.app-id")
    
    // 3. 使用配置
    log.Printf("App: %s", cfg.App.Name)
    log.Printf("API Key: %s", apiKey)
    log.Printf("WeChat AppID: %s", wechatAppId)
}
```

---

## 🔧 高级用法

### 设置默认值

```go
viper.SetDefault("custom.timeout", 30)
viper.SetDefault("custom.features.enable-cache", true)
```

### 环境变量覆盖

```bash
export FRAMEWORK_CUSTOM_API_KEY=my-key
export FRAMEWORK_THIRD_PARTY_WECHAT_APP_ID=wx123
```

### 验证配置

```go
func validateCustomConfig() error {
    apiKey := viper.GetString("custom.api-key")
    if apiKey == "" {
        return fmt.Errorf("api-key is required")
    }
    
    timeout := viper.GetInt("custom.timeout")
    if timeout <= 0 {
        return fmt.Errorf("timeout must be positive")
    }
    
    return nil
}
```

### 调试：打印所有配置

```go
import "encoding/json"

allSettings := viper.AllSettings()
jsonData, _ := json.MarshalIndent(allSettings, "", "  ")
log.Printf("All config:\n%s", string(jsonData))
```

---

## 📊 方案选择

| 场景 | 推荐方案 |
|------|----------|
| 少量简单配置 | Viper 直接读取 |
| 复杂配置/团队项目 | 自定义结构体 |
| 配置频繁变化 | Map 存储 |

---

## ✅ 最佳实践

1. **组织配置**：按功能模块分组
2. **使用常量**：避免魔法字符串
3. **提供示例**：创建 config.example.yaml
4. **验证配置**：启动时验证必要字段
5. **文档化**：在 YAML 中添加注释

---

**详细文档**: [CUSTOM_CONFIG.md](CUSTOM_CONFIG.md)
