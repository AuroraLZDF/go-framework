# 修复报告（v0.1.3 修复轮）

- 日期：2026-11-04
- 依据：上一轮复检报告 `PROJECT_RECHECK_2026-11-04.md` 的修复优先级（用户要求："按照建议修复顺序，修复以上问题"）。
- 范围：仅修改工作区文件（未 git add / commit / tag / push，避免重演上一轮误提交事件）。
- 结论：**五级优先级全部落地；构建/静态检查/测试/可执行回归/脚手架 e2e 全绿。唯一发布阻塞项是尚未提交并创建 `v0.1.3` tag/push**（见 §4 发布清单）。

---

## 1. 修复项与验证对照

### ① shutdown 日志顺序 + 测试（P1）
- `core/server/application.go`：`log.Info("Server exiting")` 移到 `app.Close()` 之前；`Close` 传播 `log.Close()` 错误；启动失败路径回滚 jwt Init 与日志引用计数；`gin.SetMode` 先于 `gin.New`；`signal.Notify` 先于 server goroutine；HTTP 超时/MaxHeaderBytes 从配置读取；TLS `MinVersion=TLS1.2`；移除 `gin.Logger()`，改为 zap 访问日志 `core/server/accesslog.go`（含 method/path/status/latency/request_id，5xx→Error）。
- `core/log/log.go` 重写：`Init(*Options) error` fail-loud（轮转目录/文件不可写返回错误）；`Reset()`；全局 `Close()` 引用计数（一个 Application.Close 不会关闭另一个 App 的日志文件）；全局 logger `atomic.Pointer` + nop 兜底（`C(nil)` 安全）。
- `core/log/rotate.go` 重写：`NewRotateWriter` 返回 error；打开成功才推进日期；write-after-close 返回 `ErrRotateWriterClosed`；Close 幂等。
- 测试：`core/log/log_test.go`（轮转失败、非法 format、引用计数、write-after-close）、`core/server/application_test.go`（优雅退出且日志含 "Server exiting"、日志目录不可写时启动报错）。
- 可执行证据：脚手架生成项目实测 SIGTERM 后日志依次为 "Shutting down server ..." → "Server exiting"，无 "file already closed"（见 §3）。

### ② JWT 弱密钥 + 文档密钥（P0）
- `core/auth/jwt/token.go`：`Init(*Options) error` 校验 nil/空/<32 字节（`MinSecretLength=32`），`atomic.Pointer` 并发安全，`Reset()`；Authorization 手写解析（大小写不敏感 Bearer、TrimSpace）；错误类型化 `*errno.Errno`（Missing/Expired/Invalid/Blacklisted）；`ExpiresIn` 改为剩余秒数；`Sign` 写 tokenId+jti+nbf/iat/exp；`Sign("")` 报错。
- `core/middleware/authn.go`：`errors.As` 透传，过期返回 401 "Token expired."，黑名单 401 "Token has been revoked."，缺头 401 "Token missing."，客户端可区分。
- `core/config/loader.go`：保留 `Validate` 强制 JWT secret ≥32 字节（`minJWTSecretLength=32`）；`core/server/application.go` 自动 `jwtpkg.Init(cfg.JWT)` 并传播 error。
- 文档：README/docs/examples 全部改 ≥32B secret（`change-this-to-a-random-secret-key-in-production` 等），删除/替换 `your-secret-key` 15B 示例。
- 可执行证据（/tmp 独立模块 + replace 当前仓库，实测输出）：
  - 未 Init 签发 → `jwt token is not initialized`
  - `Init(nil)` → `jwt options must not be nil`
  - 5 字节 secret → `jwt secret is too short: got 5 bytes, want at least 32`
  - 32B 后 Sign → `expires_in=7200`（Expire=2h，剩余秒数）
  - Authn 无 token → HTTP 401 `{"code":401,"message":"Token missing."}`；坏 token → 401；小写 `bearer` 合法 token → 200；过期 token → 401 `{"code":401,"message":"Token expired."}`。

### ③ v0.1.3 版本/脚手架收敛 + verify 去 replace（P0）
- `scaffold/internal/generator/generator.go`：项目名校验 `^[a-zA-Z][a-zA-Z0-9_-]*$`；step 化生成 + 失败 `os.RemoveAll` 回滚；生成 `go.mod` 要求 framework `v0.1.3`；生成 config 含新 server/redis 超时与池默认值、JWT ≥32B 注释；删除 3 个无人引用的 `templates/*.tpl`。
- `scaffold/cmd/main.go`：新增 `scaffold version`（打印 version.Get().Text()）；`new` 参数不足打印 usage。
- `core/version/version.go`：GitVersion `v0.1.3`，BuildDate 2026-11-04。
- `scripts/verify-scaffold.sh`：默认模式要求本地存在 `v0.1.3` tag 否则明确失败并给出打 tag 指引；`FRAMEWORK_LOCAL=1` 才注入本地 replace；保留 tidy/build/vet/启动/curl e2e。
- 可执行证据：`scaffold version` → gitVersion `v0.1.3`；`scaffold new demoapp` 生成 go.mod 含 `github.com/AuroraLZDF/go-framework v0.1.3`，追加 replace 后 `go mod tidy`/`go build` 成功；生成项目无 MySQL/Redis 可启动，`/healthz` 200、`/` 200、`/nope` 404 统一信封，SIGTERM 优雅退出。
- **注意**：本地 tag 只有 v0.1.0/v0.1.0-alpha/v0.1.1/v0.1.2，v0.1.3 未创建；未加 replace 的全新生成项目 `go mod tidy` 仍会失败，需提交后打 tag/push（§4）。

### ④ 日志生命周期 / jwt 原子 / Authn 错误细分（P1）
- 日志：见 ①（fail-loud + 引用计数 + write-after-close 明确错误）。
- jwt 全局配置：`atomic.Pointer[Options]`，多 Application 并发 Init 不再 data race；`Reset()` 测试友好。
- Authn 错误细分：见 ②；`core/errno/token.go` 新增 `ErrTokenBlacklisted`（HTTP 401）。
- `black.go`：jti-key 查询出错不再提前返回而跳过用户级吊销 epoch 检查；fail-open 查完再放行，fail-closed 返回错误；默认 fail-open 文档化，Redis 未配置时启动 WARNING。
- CORS：显式 `"*"` 正确匹配、无条件 `Vary: Origin`、Allow-Headers 补 Authorization/Content-Type/X-Request-ID/X-Requested-With/Accept/Origin 等、PATCH 恢复、credentials 仅白名单。
- validator：正则负缓存有界（256）；`WithFilters` 字段名仅简单标识符（禁 `.`，fail-closed）、LIKE 转义 `%/_`、未知 operator 报错。
- 数据层：`InitMySQL(nil)` 报错、重复 Init 返回 `ErrDBAlreadyInitialized`、close 先 Close 后置 nil、DSN 参数校验/转义；Redis nil 安全、池/超时接线、Ping 用 DialTimeout、`CloseRedis` 复位可重入；util 修复 StructToMap/snake_case/file/ip/time/tree；service 接口加 ctx。
- 测试补充：`core/database/db/db_test.go`、`core/util/util_edge_test.go`、`core/version/version_test.go`、`scaffold/internal/generator/generator_test.go`、`core/middleware/*_test.go`、`core/validator/*_test.go` 等。

### ⑤ 文档与工程化同步（P1）
- README：badge v0.1.3/Go 1.25+；全量真实 import 路径；`go get ...@v0.1.3`；JWT 段（token 包、Init 返回 error、`TokenId`、≥32B secret、显式 `middleware.Authn()`、Authorization 格式）；log 用 `OutputPaths`+`Rotation`；死链删除；重复章节清理。
- docs/quick-start、architecture、CONFIG_USAGE、CONFIG_QUICK_REFERENCE、CUSTOM_CONFIG×2、optimization-summary、DOCUMENTATION_GUIDE 全部同步（BlacklistPath→BlacklistPrefix+FailClosed、新增配置键、errno→HTTP 语义、显式 Authn、默认值由 `Validate`/`applyDefaults` 填充、池 idle≤open 等）。
- CHANGELOG：补 v0.1.2、v0.1.3（Added/Changed/Fixed/Security/Breaking）。
- 新增 `RELEASE_v0.1.3.md`；新增 `.github/workflows/ci.yml`（go 1.25.1：build/vet/gofmt/bash -n/test + `FRAMEWORK_LOCAL=1` 脚手架 e2e）。
- Makefile：install 尊重 GOBIN；test/test-scaffold/verify-scaffold/release-tag 分离；`.gitignore` 补 `.screenshots/`、`*.test`。
- 自检：`bash -n` 两个脚本通过；14 个改动 Markdown 相对链接 0 坏链；敏感旧模式 grep 无残留（除历史发布说明与 cfg 真实字段名 `TokenID`）。

---

## 2. 全量验证（Lead 亲自执行，修复后工作区）

| 检查 | 命令 | 结果 |
|---|---|---|
| 格式 | `gofmt -l .` | 无输出 |
| 构建 | `go build ./...` | 通过 |
| 静态检查 | `go vet ./...` | 通过 |
| 单元/集成测试 | `go test -count=1 ./...` | 12 个含测试包全 ok；`go test -v` 共 106 个 `--- PASS` 行；无 FAIL |
| 日志关键测试 | `go test -count=1 -v ./core/log/...` | 轮转失败/非法 format/引用计数/write-after-close 全 PASS |
| JWT/配置/服务器回归 | /tmp 独立模块 + replace 当前仓库 | 见 §1②，15 项断言全部 PASS（含过期区分、env 覆盖、坏日志目录启动报错、nil config 报错） |
| 脚手架 e2e | 生成→replace→tidy→build→启动→curl→SIGTERM | 全绿；访问日志走 zap；"Server exiting" 在 Close 前落盘；无 write-after-close 错误 |
| 示例 e2e | `go run .`（examples/simple-api） | `/` 200、`/api/v1/users` 200、`/api/v1/profile` 无 token 401 |
| 脚本/文档/CI | `bash -n`、链接检查、`grep` 敏感模式 | 通过 |

环境限制：沙箱 Go 工具链缺 `covdata` 且默认 GOCACHE 只读——`-cover`/`-race` 不可用（改用临时 GOCACHE 完成其余检查）；并发安全为代码审查（atomic.Pointer + 互斥引用计数）结论，未经 race detector 验证。

---

## 3. 关键可执行证据摘录

```
PASS sign-not-initialized       jwt token is not initialized
PASS init-nil                   jwt options must not be nil
PASS init-weak-secret           jwt secret is too short: got 5 bytes, want at least 32
PASS expires-in-remaining       7200
PASS authn-no-token-401         HTTP 401 {"code":401,"message":"Token missing."}
PASS authn-expired-401          HTTP 401 {"code":401,"message":"Token expired."}
PASS authn-valid-lowercase-bearer-200  HTTP 200
PASS env-override-app-version   9.9.9
PASS env-override-log-level     debug
PASS bad-log-dir-startup-error  failed to init log: ... mkdir /dev/null: not a directory
PASS nil-config-error           config is nil
```

```
scaffold version → gitVersion: v0.1.3
scaffold new demoapp → go.mod requires github.com/AuroraLZDF/go-framework v0.1.3
生成项目（replace 本地框架后）: tidy OK / build OK
GET /healthz → 200 {"code":200,...,"status":"ok"}
GET /        → 200
GET /nope    → 404 统一信封
SIGTERM → 日志尾部 "Shutting down server ..." → "Server exiting"，无 "file already closed"
```

---

## 4. 发布清单（必须执行，否则脚手架仍不可用）

工作区当前：`git diff` 54 个已跟踪文件 +2683/−1487，另有 24 个未跟踪文件（新增测试、`core/server/accesslog.go`、`docs/optimization-summary.md`、`RELEASE_v0.1.3.md`、`.github/workflows/ci.yml`、`examples/simple-api/config.example.yaml`、本报告与两份分析报告等）。**未做任何提交/tag/push。**

```bash
# 1) 提交（建议先检查两份分析报告是否要提交或移入 docs/internal/）
git add -A
git commit -m "feat: v0.1.3 — auth/log/server hardening, docs & scaffold sync"

# 2) 打 tag 并推送（v0.1.3 必须存在于远端，脚手架才能开箱 tidy）
git tag -a v0.1.3 -m "v0.1.3"
git push origin main --tags

# 3) 发布门槛（默认模式会真正检查 tag）
./scripts/verify-scaffold.sh        # 期望通过；本地开发可 FRAMEWORK_LOCAL=1 跳过 tag 检查
./scripts/pre-release-check.sh
```

## 5. 仍存在的已知限制（非本轮阻塞，建议后续处理）

1. **发布依赖**：v0.1.3 tag/commit 未创建；且仓库无远端推送权限验证（推送需用户执行）。
2. `/healthz` 仅 liveness，不检查 MySQL/Redis 依赖（文档已注明）；无 metrics/tracing/readiness。
3. gin 默认信任所有代理，`c.ClientIP()`/访问日志的 `client_ip` 可被 `X-Forwarded-For` 伪造；未提供 `trusted-proxies` 配置。
4. Redis 配置约定 `addr` 为 host、`port` 单独配置；若用户把 `addr` 写成 `host:port`，`net.JoinHostPort` 会重复拼接（docs/examples 均按 host+port 书写）。
5. `https-addr` 设置时 `http-addr` 被忽略（仅 TLS 监听，无 HTTP→HTTPS 重定向）。
6. `core/version/verflag` 仍无二进制消费（version 已由 `scaffold version` 使用）。
7. `docs/internal/` 下 11 个旧文件仍存在（已 gitignore，未清理）。
8. 环境无 `-race`/`-cover`，并发与覆盖率未用工具量化。
