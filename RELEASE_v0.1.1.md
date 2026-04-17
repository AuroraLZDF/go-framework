# Release v0.1.1

**发布日期**: 2026-04-16  
**类型**: Minor Release (小版本更新)

## 🎉 概述

v0.1.1 版本主要增强了配置管理能力，添加了完整的配置文档，并改进了脚手架工具以支持自定义配置读取。

---

## ✨ 新增功能

### 1. 配置文档体系 📚

新增了完整的配置管理文档：

- **CONFIG_USAGE.md** (469行) - 完整的配置加载和使用指南
  - 三种配置加载方式
  - 配置项访问方法
  - 环境变量覆盖
  - 多环境配置
  - 配置验证

- **CUSTOM_CONFIG.md** (581行) - 自定义配置读取完整指南
  - Viper 直接读取
  - 自定义结构体方案
  - Map 存储方案
  - 实际示例代码
  - 最佳实践

- **CONFIG_QUICK_REFERENCE.md** (166行) - 配置快速参考卡片
- **CUSTOM_CONFIG_QUICK_REF.md** (188行) - 自定义配置速查

### 2. 自定义配置支持 🔧

现在用户可以轻松读取 config.yaml 中的自定义配置项：

```go
import "github.com/spf13/viper"

// config.yaml:
// custom:
//   api-key: abc123
//   timeout: 30

apiKey := viper.GetString("custom.api-key")
timeout := viper.GetInt("custom.timeout")
```

### 3. 脚手架改进 🛠️

更新的脚手架生成的项目现在包含：

- ✅ 自动导入 `github.com/spf13/viper`
- ✅ go.mod 中包含 viper 依赖
- ✅ main.go 中添加自定义配置读取示例注释
- ✅ 更清晰的配置加载说明

### 4. README 改进 📖

- 添加配置管理详细说明
- 添加自定义配置读取示例
- 新增"📚 文档"章节，包含所有文档链接
- 更新快速开始部分的配置提示

---

## 🔄 变更内容

### 核心更改

1. **模块路径统一** - 所有导入路径更新为 `github.com/AuroraLZDF/go-framework`
2. **配置加载器** - 支持指定配置文件路径
3. **配置验证** - 增强的配置验证机制

### 文档更改

- 新增 4 个配置相关文档
- 更新 README.md
- 更新脚手架模板

### 依赖更改

- 脚手架生成的项目现在依赖 `github.com/spf13/viper v1.18.2`

---

## 📦 如何升级

### 方法一：使用新版本（推荐）

在你的项目 `go.mod` 中：

```go
require github.com/AuroraLZDF/go-framework v0.1.1
```

然后运行：

```bash
go mod tidy
```

### 方法二：使用最新代码

如果你想使用最新的开发代码：

```go
require github.com/AuroraLZDF/go-framework v0.0.0-20260416135318-c551e20d8f9a
```

获取最新 commit hash：

```bash
git ls-remote https://github.com/AuroraLZDF/go-framework.git HEAD
```

### 方法三：使用 replace 指令（开发时）

如果你在本地开发：

```go
require github.com/AuroraLZDF/go-framework v0.1.1

replace github.com/AuroraLZDF/go-framework => /path/to/local/go-framework
```

---

## 💡 迁移指南

### 从 v0.1.0 升级到 v0.1.1

#### 1. 更新依赖

```bash
go get github.com/AuroraLZDF/go-framework@v0.1.1
go mod tidy
```

#### 2. 更新导入路径（如果还没有）

确保所有导入使用新的模块路径：

```go
// 旧的路径（不再使用）
import "go-framework/core/config"

// 新的路径
import "github.com/AuroraLZDF/go-framework/core/config"
```

#### 3. 利用新功能

现在可以读取自定义配置了：

```go
import "github.com/spf13/viper"

// 读取你的自定义配置
myConfig := viper.GetString("my.custom.setting")
```

---

## 🐛 Bug 修复

- 修复了脚手架生成项目的模块路径问题
- 修复了配置加载示例代码

---

## 📝 已知问题

无

---

## 🔗 相关链接

- [GitHub Release](https://github.com/AuroraLZDF/go-framework/releases/tag/v0.1.1)
- [完整文档](https://github.com/AuroraLZDF/go-framework/tree/main/docs)
- [配置使用指南](https://github.com/AuroraLZDF/go-framework/blob/main/docs/CONFIG_USAGE.md)
- [自定义配置读取](https://github.com/AuroraLZDF/go-framework/blob/main/docs/CUSTOM_CONFIG.md)

---

## 👥 贡献者

- auroralzdf

---

## 🙏 致谢

感谢所有使用 Framework 的开发者！

---

**Happy Coding!** 🎉
