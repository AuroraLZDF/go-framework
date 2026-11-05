# 脚手架开发文档

本文档面向维护 `scaffold` 的开发者，说明生成器结构、验证方式与版本管理约定。

## 命令

```bash
scaffold new <project-name>   # 生成新项目
scaffold version              # 打印构建版本信息（由 Makefile 注入）
```

## 代码结构

```
scaffold/
├── cmd/
│   └── main.go                     # CLI 入口：参数解析、new/version 命令
├── internal/
│   └── generator/
│       ├── generator.go            # 项目生成逻辑（字符串模板 + 文件写入）
│       └── generator_test.go       # 生成器单元测试
├── README.md                       # 使用说明
├── QUICKSTART.md                   # 5 分钟快速开始
└── DEVELOPMENT.md                  # 本文件
```

> 生成器**没有独立的模板文件**：所有生成内容都以内嵌字符串形式写在
> `generator.go` 中，`scaffold/templates/` 目录已移除。

## 生成器实现

`generator.NewProject(name)` 的主流程：

1. 校验项目名（非空、不含 `/\:*?"<>|` 等非法字符），目标目录已存在时报错；
2. 创建目录：`cmd/app`、`internal/{biz,controller,store,model,service}`、
   `configs`、`sql`、`assets`、`logs`；
3. 生成 `go.mod`：`go 1.26.0`，依赖 `gin v1.12.0`、`gorm v1.31.1` 与
   `github.com/AuroraLZDF/go-framework v0.2.0`；
4. 生成 `cmd/app/main.go`：`config.Load("")` → `cfg.Validate()` →
   `server.NewApplication(cfg)` → `RegisterRoutes(func(*gin.Engine, *gorm.DB))`，
   内置 `GET /health`、`GET /` 与 `/api/v1/users` 示例；
5. 生成 `config.yaml` 与内容相同的 `config.example.yaml`（可直接运行；
   JWT secret 使用 48 字节示例值，满足 ≥32 字节校验；MySQL/Redis 可留空跳过）；
6. 生成 `.gitignore` 与项目 `README.md`。

生成的 `go.mod` 使用框架 `v0.2.0`，因此发布前必须先推送同名 tag；本地开发时
`FRAMEWORK_LOCAL=1` 会向生成项目追加 `replace` 指向当前仓库。

## 修改生成内容

直接编辑 `scaffold/internal/generator/generator.go` 中对应的生成函数：

- `generateGoMod` — 依赖与 Go 版本；
- `generateMainFile` — 应用入口与示例路由；
- `generateConfigFile` — `config.yaml` / `config.example.yaml`；
- `generateGitignore`、`generateReadme` — 辅助文件。

修改后：

```bash
go test ./scaffold/...        # 生成器单测
make test-scaffold            # 临时目录端到端验证（生成→tidy/build/vet→启动→curl）
```

## 版本管理

版本号通过 Makefile 的 `-ldflags` 注入 `core/version` 的
`GitVersion/GitCommit/BuildDate/GitTreeState`。发布流程（仅在本地执行）：

```bash
make build
make test-scaffold
make release-tag   # 只打印 git tag 指引，不会创建 tag
git tag -a v0.2.0 -m 'release: v0.2.0'
git push origin main && git push origin v0.2.0
```

## 设计决策

- **字符串模板而非 text/template**：生成内容固定、字段少，直接在 Go 字符串中
  拼接更易阅读和审查；代价是较大的 `generator.go`。
- **一次生成完整可运行项目**：包含 `config.yaml`、示例路由与 README，
  用户 `go mod tidy` 后即可启动，无需复制配置文件。
- **不维护模板目录**：避免「模板存在但未被生成器使用」的不一致问题；
  需要模板引擎时可后续引入，届时同步更新本文件。

## 已知限制

- 仅支持 `new`（非交互式），暂无 `generate api/model/crud` 等子命令；
- 生成内容与框架 `v0.2.0` API 绑定，升级框架 API 时需同步修改生成器；
- 发布前必须推送 `v0.2.0` tag，否则新项目 `go mod tidy` 会失败。

## 相关文档

- [脚手架使用指南](README.md)
- [脚手架快速开始](QUICKSTART.md)
- [Framework 快速开始](../docs/quick-start.md)
- [CHANGELOG](../CHANGELOG.md)
