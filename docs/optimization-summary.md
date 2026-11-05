# 框架优化总结（v0.1.3 开发中）

本文件记录截至 **2026-11-04** 的这一轮安全性、稳定性与可测试性优化。它是对当前
工作区状态的说明，而不是发布公告；发布流程见 [RELEASE_v0.1.3.md](../RELEASE_v0.1.3.md)。

## 概览

本轮改动覆盖 `core/`、`scaffold/`、`docs/`、`scripts/`、`examples/` 等目录，
当前工作区相对 `HEAD` 约有 53 个已跟踪文件变更（+2681/−1486），并新增了 17 个测试文件。
核心目标：

1. 消除 JWT/配置/数据库等基础设施中的不安全默认值；
2. 让错误响应具备正确的 HTTP 语义，并避免泄露内部错误细节；
3. 补齐优雅关闭、超时、日志轮转与请求上下文等稳定性细节；
4. 为关键路径补充可执行的单元测试，并提供脚手架端到端验证脚本。

## 安全性

- **JWT 算法与密钥**：只接受 HS256（`WithValidMethods`）且要求 `exp`；
  `token.Init()` 校验 Secret ≥ 32 字节，使用 `atomic.Pointer` 保存配置，
  未初始化时拒绝签发/验签。`server.NewApplication` 自动完成接线并传播错误。
- **Token 标识**：签发时自动写入 `jti`；`ParseRequest` 要求 jti 存在；
  黑名单支持 jti 级与用户级全端撤销。
- **黑名单失败策略**：新增 `jwt.blacklist-fail-closed`（默认 false，fail-open），
  Redis 不可用时可选择拒绝请求；未配置 Redis 时启动打印 WARNING。
- **CORS**：白名单模式回显 Origin、允许凭证，补充 PATCH 与 `X-Requested-With`；
  非白名单不返回 CORS 头。
- **SQL 注入防护**：`validator.WithFilters` 对字段名做白名单校验，
  非法字段转换为 `1 = 0`（fail-closed）。
- **随机数**：`GenerateRandomString` 改用 `crypto/rand` 并消除取模偏差。
- **错误响应**：`errno.Decode` 对未知错误返回统一 500 文案，不再回显 `err.Error()`；
  `response.Error` 记录 5xx 日志（含 path/request_id）但不泄露给客户端。
- **请求 ID**：只接受 `[A-Za-z0-9._-]{1,128}` 的客户端值，否则重新生成。

## 稳定性

- **优雅关闭**：`server.shutdown-timeout` 可配置（默认 10s）；关闭失败时回退
  `Close()`；`Application.Close()` 使用 `sync.Once` 且先关数据库再关日志。
- **HTTP 服务器**：ReadHeader/Read/Write/Idle 超时（10s/30s/60s/120s）、
  `max-header-bytes`（1MB）、TLS 最低 1.2；内置 `GET /healthz` 存活探针。
- **配置加载**：所有内置键通过 `BindEnv` 支持环境变量覆盖；允许无配置文件时用
  环境变量 + 默认值启动；默认值（超时、连接池、JWT 默认字段）集中到 `applyDefaults`。
- **可选依赖**：MySQL/Redis 未配置时跳过初始化，不影响启动。
- **连接管理**：MySQL DSN 增加 `timeout/readTimeout/writeTimeout` 且使用 `loc=UTC`；
  连接池参数有默认值与合法性钳制；Redis 增加连接池与超时配置，并提供 `CloseRedis()` 重置单例。
- **日志**：支持 `OutputPaths` 与 `Rotation`（daily/hourly、可选 stdout）；
  轮转初始化失败会使启动失败；`Close()` 带引用计数，避免共享 Writer 被提前关闭；
  `log.C(ctx)` 自动附加 request_id。

## 正确性

- `errno` 的 HTTP 状态码恢复语义：401（未认证/Token 错误）、404、429、500；
  业务码与 HTTP 码分离，`Decode` 支持 `errors.As`。
- `SetMessage` 返回副本，不再修改全局错误定义。
- 修复 `log.Warning` 误用、`ParamToInt64` 读取位置错误、`GenerateRandomString` 长度边界等问题。
- `validator` 包由 `tools` 更名，`RouteRegisterFunc` 明确为
  `func(*gin.Engine, *gorm.DB)`。
- 配置校验：`app.mode`/`log.format` 枚举、JWT secret ≥32 字节、
  MySQL 仅在配置 host 时校验、HTTPS 必须成对提供证书/私钥。

## 脚手架

- 生成项目 `go.mod` 使用 `go 1.26.0`，依赖 `github.com/AuroraLZDF/go-framework v0.1.3`。
- 生成 `config.yaml`（可直接运行）与 `config.example.yaml`，无需手动 `cp`；
  默认 JWT secret 为 48 字节示例值，满足校验要求。
- 生成路由注册使用 `*gorm.DB` 签名，并内置 `GET /health` 业务健康检查。
- 移除无效的 Viper 依赖与未使用的模板文件；新增 `scaffold version` 命令。
- `scripts/verify-scaffold.sh` 会生成项目、执行 `go mod tidy/build/vet`，
  启动应用并请求 `/health` 与 `/healthz`（`FRAMEWORK_LOCAL=1` 时使用本地 replace）。

## 测试

- 新增/更新测试覆盖：JWT（签名/解析/jti/黑名单/弱密钥）、配置（默认值/环境变量/
  无文件启动/校验）、errno、CORS、requestId、Authn、response、server 生命周期、
  MySQL DSN、DB 关闭、转换工具等。
- 新增 `middleware/authn_test.go`、`middleware/requestId_test.go`、
  `response/core_test.go`、`server/application_test.go` 等。
- 仍未覆盖：日志轮转/`log.Close` 的具体行为、`server.Run` 的完整启动-关闭流程、
  真实 Redis 黑名单、`validator.WithFilters` 的注入行为。后续可优先补这些用例。

## 已知限制

- 仍不提供配置热重载；配置变更需重启。
- 黑名单默认 fail-open，安全敏感场景应设置 `jwt.blacklist-fail-closed: true`。
- 脚手架依赖已发布的 `v0.1.3` tag；未发布前请使用 `FRAMEWORK_LOCAL=1` 本地验证。
- 破坏性变更（详见 [CHANGELOG.md](../CHANGELOG.md)）：JWT 现在要求 ≥32 字节密钥、
  存量 token 因 jti 校验需要重新签发；errno 的 HTTP 状态码不再恒为 200；
  日志 `Init`/`Close` 语义与轮转配置有调整。

## 验证方式

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
bash -n scripts/pre-release-check.sh scripts/verify-scaffold.sh
make test-scaffold   # FRAMEWORK_LOCAL=1，本地端到端验证脚手架
```

相关文档：[架构设计](architecture.md)、[快速开始](quick-start.md)、
[配置使用](CONFIG_USAGE.md)、[配置快速参考](CONFIG_QUICK_REFERENCE.md)。
