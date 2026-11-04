# Release v0.1.3

**发布日期**: 2026-11-04

v0.1.3 是一次安全性、稳定性与可测试性升级，包含若干 **破坏性变更**。升级前请阅读
「Breaking Changes」与「升级步骤」。

## ✨ 亮点

- **安全加固**：JWT 强制 HS256 + 32 字节密钥 + jti/黑名单校验；错误响应不再泄露
  内部错误；请求 ID 校验；SQL 过滤字段白名单；`crypto/rand` 随机数。
- **稳定性**：可配置优雅关闭与 HTTP 超时；TLS 最低 1.2；所有内置配置键支持
  `FRAMEWORK_*` 环境变量（无配置文件也能启动）；MySQL/Redis 可选；
  日志轮转失败即启动失败、`Close()` 引用计数。
- **正确性**：HTTP 状态码语义（401/404/429/500）；请求 ID 写入 `r.Context()`；
  CORS 补充 PATCH；MySQL 连接池与 DSN 时区/超时修正。
- **可测试性**：新增 17 个测试文件，覆盖 JWT、配置、errno、中间件、response、
  server 生命周期、数据库、工具函数；新增脚手架端到端验证脚本与 GitHub Actions CI。
- **脚手架**：生成 Go 1.25.1 + framework `v0.1.3` 的可运行项目（含 `config.yaml`
  与 `/health`），新增 `scaffold version`，移除无用的 viper 依赖与模板文件。

## 💥 Breaking Changes

1. **JWT 密钥至少 32 字节**：`config.Validate()` 与 `token.Init()` 会拒绝短密钥；
   请使用 `openssl rand -base64 32` 生成（44 字符）。
2. **存量 Token 失效**：Token 现在必须携带 `jti`，升级后客户端需重新登录。
3. **JWT 包名为 `token`**（导入路径不变：`core/auth/jwt`）；
   `token.Init(opts *Options) error` 返回错误，`ExpiresIn` 为剩余秒数。
4. **错误响应 HTTP 状态码变化**：未认证/Token 错误为 401，404/429/500 恢复语义，
   不再统一返回 200；依赖旧行为的客户端需调整。
5. **日志配置语义变化**：`log.Init()` 返回 error，`Options.Dir` 已移除；
   使用 `OutputPaths` + `Rotation *RotateConfig`；`log.Close()` 带引用计数。
6. **脚手架依赖 `v0.1.3` tag**：未发布该 tag 时，生成项目 `go mod tidy` 会失败；
   本地开发可 `FRAMEWORK_LOCAL=1 make test-scaffold`。
7. **`validator` 包更名**：原 `tools` 包更名为 `validator`，旧导入路径需更新。
8. **`RouteRegisterFunc` 签名**：`func(*gin.Engine, *gorm.DB)`。

## 📦 升级步骤

1. 更新依赖：

   ```bash
   go get github.com/AuroraLZDF/go-framework@v0.1.3
   go mod tidy
   ```

2. 修改 `config.yaml`：

   ```yaml
   server:
     shutdown-timeout: 10s
     read-header-timeout: 10s
     read-timeout: 30s
     write-timeout: 60s
     idle-timeout: 120s
     max-header-bytes: 1048576

   jwt:
     secret: "至少 32 字节的随机密钥"   # 必填，否则 Validate 失败
     blacklist-prefix: "jwt:blacklist:"
     blacklist-fail-closed: false
   ```

3. 调整代码：
   - `token.Init(...)` / `jwt.Init(...)` 现在返回 error（`NewApplication` 会自动调用）；
   - JWT 字段 `TokenId`（原 `TokenID`）；
   - 受保护路由显式挂载 `middleware.Authn()`；
   - 日志初始化改用 `OutputPaths` / `Rotation`，并检查 `log.Init` 返回值；
   - `RegisterRoutes` 回调签名为 `func(*gin.Engine, *gorm.DB)`。

4. 如使用脚手架，重新生成或对照新版 `generator.go` 更新模板。

## 🔧 配置新增

- `server.shutdown-timeout` / `read-header-timeout` / `read-timeout` /
  `write-timeout` / `idle-timeout` / `max-header-bytes`
- `redis.pool-size` / `dial-timeout` / `read-timeout` / `write-timeout`
- `jwt.blacklist-fail-closed`

## 🧪 验证

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
bash -n scripts/pre-release-check.sh scripts/verify-scaffold.sh
FRAMEWORK_LOCAL=1 make test-scaffold
```

CI 配置见 `.github/workflows/ci.yml`。

## ⚠️ 已知问题

- 尚无配置热重载；
- 黑名单默认 fail-open，安全敏感场景请开启 `jwt.blacklist-fail-closed`；
- 日志轮转与 `server.Run` 完整流程的单元测试仍待补充。

## 🔗 相关链接

- [CHANGELOG](CHANGELOG.md)
- [快速开始](docs/quick-start.md)
- [架构设计](docs/architecture.md)
- [配置使用](docs/CONFIG_USAGE.md)
- [GitHub Releases](https://github.com/AuroraLZDF/go-framework/releases)
