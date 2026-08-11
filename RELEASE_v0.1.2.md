# Release v0.1.2

**发布日期**: 2026-08-11  
**类型**: Minor Release (小版本更新)

## 🎉 概述

v0.1.2 版本主要新增了日志轮转功能，支持按天或按小时自动切分日志文件，便于日志管理和归档。

---

## ✨ 新增功能

### 1. 日志轮转支持 📝

新增了完整的日志轮转功能，支持按时间自动切分日志文件：

- **按天轮转** (`daily`) - 每天生成一个日志文件，格式：`app.20260811.log`
- **按小时轮转** (`hourly`) - 每小时生成一个日志文件，格式：`app.2026081114.log`
- **自动目录创建** - 自动创建日志目录
- **控制台同时输出** - 可选同时输出到文件和控制台

### 2. 使用方式

```go
import "github.com/AuroraLZDF/go-framework/core/log"

// 初始化带日志轮转的 logger
log.Init(&log.Options{
    Level:  "info",
    Format: "json", // 或 "console"
    Rotation: &log.RotateConfig{
        BasePath: "./logs/app",  // 日志文件基础路径
        Interval: "daily",       // "daily" 或 "hourly"
        Stdout:   true,         // 同时输出到控制台
    },
})
```

### 3. 新增 API

- `RotateConfig` - 日志轮转配置结构体
  - `Interval`: 轮转间隔 ("daily" | "hourly")
  - `BasePath`: 日志文件基础路径
  - `Stdout`: 是否同时输出到控制台

---

## 🔄 变更内容

### 核心更改

1. **[core/log/options.go](file:///home/code/go/src/framework/core/log/options.go)**
   - 新增 `RotateConfig` 结构体
   - `Options` 结构体新增 `Rotation` 字段

2. **[core/log/rotate.go](file:///home/code/go/src/framework/core/log/rotate.go)** (新增)
   - 实现 `RotateWriter` 日志轮转写入器
   - 支持线程安全的日志写入和轮转

3. **[core/log/log.go](file:///home/code/go/src/framework/core/log/log.go)**
   - 新增 `createRotateCore` 函数
   - `NewLogger` 支持日志轮转配置

---

## 📦 如何升级

### 方法一：使用新版本（推荐）

在你的项目 `go.mod` 中：

```go
require github.com/AuroraLZDF/go-framework v0.1.2
```

然后运行：

```bash
go mod tidy
```

### 方法二：使用 replace 指令（开发时）

如果你在本地开发：

```go
require github.com/AuroraLZDF/go-framework v0.1.2

replace github.com/AuroraLZDF/go-framework => /path/to/local/go-framework
```

---

## 💡 迁移指南

### 从 v0.1.1 升级到 v0.1.2

#### 1. 更新依赖

```bash
go get github.com/AuroraLZDF/go-framework@v0.1.2
go mod tidy
```

#### 2. 启用日志轮转（可选）

如果需要日志轮转功能，更新日志初始化代码：

```go
// 旧方式（无轮转）
log.Init(&log.Options{
    Level:  "info",
    Format: "console",
})

// 新方式（带轮转）
log.Init(&log.Options{
    Level:  "info",
    Format: "console",
    Rotation: &log.RotateConfig{
        BasePath: "./logs/app",
        Interval: "daily",
        Stdout:   true,
    },
})
```

---

## 🐛 Bug 修复

无

---

## 📝 已知问题

无

---

## 🔗 相关链接

- [GitHub Release](https://github.com/AuroraLZDF/go-framework/releases/tag/v0.1.2)
- [完整文档](https://github.com/AuroraLZDF/go-framework/tree/main/docs)

---

## 👥 贡献者

- auroralzdf

---

## 🙏 致谢

感谢所有使用 Framework 的开发者！

---

**Happy Coding!** 🎉
