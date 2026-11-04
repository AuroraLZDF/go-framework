# go-framework 修复后复检报告（2026-11-04）

- 复检对象：`/home/code/go/src/framework`（module `github.com/AuroraLZDF/go-framework`）
- 基线：`HEAD ad185b8` + 未提交工作区（35 个已修改文件，`+1022/−505`；13 个测试文件；`git tag` 最新 `v0.1.2`）
- 方式：Lead + 4 个只读 teammate（runtime / security / data+scaffold / docs+tests），逐条复核上一轮 75+ 项问题；全部结论按 **FIXED / PARTIAL / STILL PRESENT / NEW（修复引入）** 标注
- 可执行验证：`go build ./...`、`go vet ./...`、`gofmt -l .`、`go test -count=1 ./...` 全绿（13 个测试文件、9 个包有测试；58 顶层测试/65 RUN）；另在 `/tmp` 独立模块（`replace` 到本仓库）真实跑 JWT 认证链路、配置覆盖、脚手架生成项目 build/启动/健康检查。仓库代码零修改（仅新增本报告文件）。
- 环境限制：本机 Go 工具链被裁剪，`-race` / `-cover` 无法构建，竞态与覆盖率结论来自代码审查。

---

## 1. 总体结论

**上一轮的 4 个 P0（JWT 零初始化/空密钥伪造、manifest 环境变量失效、启动期 panic、脚手架生成物不可用）中，前三个已在关键路径上真实关闭，且有回归测试；脚手架"代码内容"已修好，但"用户流程"仍被一个未发布版本阻断。**

当前**不建议直接合并/发布**，还有 2 个 P0 级收尾项和 4 个 P1 级新问题：

| # | 级别 | 问题 | 证据 |
|---|---|---|---|
| P0-1 | 文档回归 | 新增"JWT secret ≥32B"强制校验，但 README/CONFIG 示例仍是 11–15 字节，照抄即启动失败 | `loader.go:19,169-174`；`README.md:112,218`；`CONFIG_QUICK_REFERENCE.md:91`；`CONFIG_USAGE.md:204` |
| P0-2 | 发布阻断 | 脚手架生成物固定依赖 `go-framework v0.1.3`，但该 tag 不存在，真实用户 `go mod tidy` 必失败；`verify-scaffold.sh` 用本地 `replace` 掩盖 | `generator.go:109`；`scripts/verify-scaffold.sh:70,79`；本机 tag 仅 ≤v0.1.2 |
| P1-1 | 新回归 | 关闭时 `log.Info("Server exiting")` 在 `app.Close()` 之后执行，轮转日志必现 `write error: ... file already closed`，且该行不落盘 | `application.go:169-176`；`log/rotate.go:65-72` |
| P1-2 | 新回归 | 日志是全局单例，一个 `Application.Close()` 会关闭另一个 App 的轮转文件 | `log/log.go:44-56`；`application.go:183-193` |
| P1-3 | 新风险 | `jwt.Init` 去掉 `sync.Once` 后全局 config 无锁，运行期 Init 并发 Sign/Parse 有 data race；同进程多 App 互相覆盖 JWT 配置 | `auth/jwt/token.go:37-47,61-63` |
| P1-4 | 仍存在 | 日志轮转失败仍静默：坏目录下 `NewApplication` 返回 nil error、文件日志全丢 | `application.go:207-233`；`rotate.go:34,88,96-99`（实测） |

此外：包级 `token.Init` 仍不校验密钥长度（5 字节可签发、Authn 放行 200，与 config 层双标准）；`Authn` 仍把所有错误扁平化为 401 "Token was invalid."；黑名单默认 fail-open；README/docs 大面积未同步（含不存在链接、错误包名/字段名、旧 badge）；无 CI；`docs/optimization-summary.md` 多处与现状矛盾。

---

## 2. 上一轮问题逐项状态

### 2.1 FIXED（已实证）

| 上一轮问题 | 修复点 | 验证 |
|---|---|---|
| JWT 零初始化仍可签发/验签、空 Secret 伪造 token（P0） | `token.go:50-56` checkInitialized + Sign/Parse 入口校验 | 实测未 Init：`Sign err: jwt token is not initialized`；伪造 token → HTTP 401、handler 未执行 |
| config.JWT 未接线、Authn 未挂载（P0） | `application.go:57-64` 从 cfg 初始化；README 双参 RegisterRoutes | `application_test.go` TestNewApplicationWiresJWT；实测 合法 token → 200 uid=999 |
| config 层无 secret 必填/强度校验（P0） | `loader.go:19` `minJWTSecretLength=32`，`:169-174` | 实测 `Validate(secret="short")` → `jwt secret must be at least 32 bytes, got 5` |
| Env 变量对 YAML 缺失键无效（P0） | `loader.go:23-57` envKeys 显式列表 + `:88-101` 逐个 BindEnv | 实测 `FRAMEWORK_JWT_TOKEN_ID→sub`、`FRAMEWORK_LOG_LEVEL→debug`、`FRAMEWORK_APP_VERSION→9.9.9` 全部生效 |
| 启动期 panic（app.mode/log.format 非法、nil cfg） | `loader.go:149-192` Validate + `application.go:48-55` | 实测非法 mode/format 返回 error 不再 panic；`-cover` 环境下亦未触发 |
| 错误全部 HTTP 200、泄漏内部错误（P1） | `errno/code.go:13-40` 语义化状态码；`errno.go:38-45` errors.As + 500 通用文案；`response/core.go:38-56` | 实测 401/404/500 正确，body 不泄漏 `err.Error()` |
| requestId 可伪造/不传播（P1） | `middleware/requestId.go:18-38` 正则白名单 + 写入 `r.Context()` + 响应头 | 实测 header==gin ctx==r.Context；129 字符非法值被替换 UUID |
| 无 jti、无法按 token 吊销（P1） | `token.go:168-174` 签发 jti；`black.go:26-108` jti 会话级 + 用户级 epoch 吊销 | 实测 无 jti token → 401；有 jti → 200 |
| `FailureAllClient` TTL 漂移（P1） | `black.go:141-148` 写永久吊销 epoch，按 iat 比较 | 代码路径确认 |
| 黑名单空 TokenId key 退化 | jti/user 双 key + `loader.go:140-142` 默认 uid | 代码确认 |
| CORS 丢 PATCH / X-Requested-With | `cors.go:37-38` 恢复 | 代码确认 |
| Close 顺序/幂等、Shutdown 无兜底 | `application.go:183-193` sync.Once + db→log；`:153-166` 超时后强制 Close | `application_test.go` CloseIdempotent；实测 ctx cancel 后 ~300ms 返回 |
| 脚手架生成内容不可用（P0） | `generator.go:102-131,153,203-246`：正确 module path、go 1.25.1、无 viper 死 import、双参签名、默认无 MySQL/Redis、HTTPS 注释、同时生成 config.yaml | 追加 replace 后：tidy/build/vet 通过；启动 `/health`、`/healthz`、`/`、404 均统一 envelope；SIGTERM 优雅退出 |
| DSN 无超时/loc 不一致 | `db.go:43-54` 加 `loc=UTC&timeout=5s&readTimeout=30s&writeTimeout=30s` | 代码 + db_test |
| cache.Close 关共享 Redis 无法重置 | `cache.go:79-83` 调 `db.CloseRedis()`；`redis.go:80-88` 复位单例 | 代码确认 |
| pre-release-check 计数器 `set -e` 早退 | `scripts/pre-release-check.sh:25,30,35` 改 `PASS=$((PASS+1))` | `bash -n` 通过 |
| Makefile test 实为脚手架测试 | 新增独立 `test`=go test；`test-scaffold`；ldflags 注入 version 包 | 代码确认 |

### 2.2 PARTIAL（改善但未闭环）

| 项 | 现状 | 证据 |
|---|---|---|
| 日志轮转 fail-loud | 坏目录仍 `NewApplication` 成功、仅 stderr 刷 `write error: invalid argument`；`initLog` 恒 nil、`log.Init` 无 error、rotate 忽略初始错误 | `application.go:207-233`；`rotate.go:34,88,96-99`（实测 `/dev/null/logs`） |
| 包级 `token.Init` secret 校验 | 只判空，不校验 ≥32；实测 5B secret → Sign 成功 → Authn 200，与 config 层双标准 | `token.go:41-56`（实测） |
| Authn 错误映射 | 过期/非法/黑名单全部映射 ErrTokenInvalid，客户端无法 refresh | `authn.go:26-27`；实测过期 token 401 "Token was invalid." |
| 黑名单 fail-open | 新增 `BlacklistFailClosed` 可配，但默认 false；且 Exists 出错提前返回会跳过用户级吊销检查 | `black.go:74-77,88-95`；`examples/config.example.yaml:42` |
| CORS | PATCH/X-Requested-With 已回；`X-Request-ID` 等自定义头仍不在 Allow-Headers；显式传 `"*"` 仍匹配不到 Origin、allowAll 无 Vary | `cors.go` |
| WithFilters 注入面 | 字段正则允许 `.`（可 `users.password`）、Field 仍 fmt.Sprintf 插值、LIKE 未转义 `%/_`、未知 operator 静默按 `=` | `validator/query.go:16,116-140` |
| 正则缓存 | 编译失败不缓存（文档声称已避免重复编译）；sync.Map 无上限 | `validator/helper.go:46-48,20` |
| 超时可配置 | ShutdownTimeout 可配；Read/Write/Idle/ReadHeader 仍硬编码 10/30/60/120s，无 ServerConfig 字段；WriteTimeout 60s 会砍 SSE | `application.go:311-319`；`config/options.go:38-43` |
| TLS | 只认 https-addr 且静默丢 http-addr；无 TLS MinVersion/CipherSuites；证书仅校验非空 | `application.go:117-120`；`loader.go:187-189` |
| 轮转选项 | 忽略 DisableStacktrace/OutputPaths；interval 硬编码 daily；无 retention/大小上限 | `log/log.go:92-102`；`application.go:225` |
| version 链路 | Makefile 注入 LDFLAGS，但无任何 binary 打印版本、无 version 命令，verflag 无人 import | `Makefile:10-13`；`scaffold/cmd/main.go` |
| 默认值 | applyDefaults 仍只在 Validate 内，直接 Load 的消费者拿零值 | `loader.go:121-146` |
| 文档同步 | README 仅 +3 行、quick-start 2 处、CONFIG_USAGE 4 处；JWT 示例仍不可编译/会启动失败 | 见 §4 |
| 生成器健壮性 | 命名仅校验非法字符、无失败回滚；模板 3 个 `.tpl` 为死文件 | `generator.go:15-74`；`scaffold/templates/*` |
| 数据层收尾 | cache.Get 无 ErrCacheMiss、NewCache 快照客户端 | `cache.go:36-38` |

### 2.3 STILL PRESENT（未修）

- **DB/Redis**：`InitMySQL` 二次调用忽略新 options（`db.go:64-67`）；`closeMySQL` 先置 nil 再 Close，失败丢池（`:127-129`）；`GetDB` 每次读加 Mutex（`:113-115`）；`InitMySQL(nil)` nil 解引用；Redis `Ping` 硬编码 5s、忽略 DialTimeout（`redis.go:56`）；Redis pool/timeout 配置未接线（`RedisConfig` 仅 Addr/Port/Password/DB）；DSN username/password 未转义（特殊字符可破坏 DSN）。
- **util**：`StructToMap` 未导出字段 panic（`utils.go:32`）；`toSnakeCase` UserID→user_i_d；`file.go` AddFile 丢弃 os.Create 句柄（fd 泄漏，`:42-47`）、AddFileIfNotExists TOCTOU（`:14`）；`ip.go` GetRealClientIP 可伪造且首个非法头不回退（`:15-33`）；`tree.go` 忽略配置、重复追加、改入参；`time.go` 仅两种格式、UnmarshalJSON recover 吞 panic。
- **其他**：password bcrypt cost 10、无策略/调用点（`password/auth.go`）；service 接口无 ctx、StreamCompletion 无 error（`service/interface.go:12,29,67`）；`Response.Content` 错误响应仍可能 `content:null` 之外的细节；`/healthz` 不检查依赖（`application.go:305-307`）；gin 默认信任所有代理；无请求体大小限制。
- **生命周期**：SIGTERM 注册晚于 server goroutine（`application.go:123-141`），窗口内信号直接杀进程；`NewApplication` 失败路径不释放已初始化的 log/DB（实测 fd +1）；`log.C(nil)` panic；`With()` 不在 Logger 接口；全局 logger 读取无锁（除写 std 外）。
- **工程化**：无 CI/.github、无 golangci 配置、无 CONTRIBUTING；CHANGELOG 只到 v0.1.1、缺 v0.1.2 条目且保留"No unit tests yet/HTTPS 未实现"等过期陈述；`core/version/version.go:18-21` 仍 v0.1.0-alpha/BuildDate 2026-04-16/占位 GitCommit；`docs/internal/` 11 个过期文件仍在（已被 .gitignore 忽略）；版权头 biddingAI/elgo 混杂；无 CI 下 README badge "build-passing" 名不副实；`PROJECT_ANALYSIS_2026-11-04.md`（307 行）未跟踪，勿误提交。

### 2.4 NEW（修复引入的新问题/回归）

1. **N1 关闭日志 write-after-close（P1）**：`application.go:169` 先 `app.Close()`（内部 log.Sync/log.Close），`:176` 再 `log.Info("Server exiting")`。实测本框架与生成项目均报 `write error: write logs/demo.2026-11-04.log: file already closed`，轮转文件缺该行。修：将最后日志移到 Close 前，或 Close 内先 flush；补 server.Run 回归测试。
2. **N2 全局 logger 被单 App 关闭（P1）**：`log/log.go` 全局 `std`；`a1.Close()` 关闭 appB 的轮转文件（实测 logB 出现 "before closing app1" 后无 "after" 并报 file already closed）。修：per-App logger 或引用计数 + 显式 Reset。
3. **N3 jwt 全局 config 竞态/多实例覆盖（P1）**：去掉 `sync.Once` 后 `config = *opts` 无同步；并发 NewApplication/Sign 存在 data race，后建 App 覆盖先建 App 的 JWT 配置。修：`atomic.Pointer[Options]` 或锁 + 显式 Reset；单测隔离。
4. **N4 破坏性变更未提示（P1）**：强制 jti（`token.go:128-131`）使升级前签发的所有 token 立即 401；JWT secret ≥32B 为新增无条件要求（不涉及 JWT 的应用也必须配）；errno 由全 200 改为语义状态码影响前端/网关。三条都需写进发版说明。
5. **N5 文档回归（P0-1）**：README/CONFIG_USAGE/CONFIG_QUICK_REFERENCE 的示例密钥 11–15B，照抄启动失败；所有文档均未提 32 字节要求。
6. **N6 脚手架发布阻断（P0-2）**：`generator.go:109` 固定 v0.1.3；真实 registry 无该版本（本地 tag 仅 ≤v0.1.2，且 v0.1.2 的 `RouteRegisterFunc` 还是 `interface{}` 旧签名），生成项目 `go mod tidy` 失败；`scripts/verify-scaffold.sh:70,79` 硬 grep v0.1.3 并注入本地 replace 掩盖。修：打并推送 v0.1.3（含全部工作区修复），或生成器改用已发布版本/伪版本；CI 改为"无 replace 必须通过"。
7. **N7 README JWT 片段仍不可编译**：包名实际 `token`（示例写 `jwt.Init`）、字段 `TokenId`（示例写 `TokenID`）、import 路径 `go-framework/...` 非真实 module；"验证 Token（中间件自动完成）"也不实（Authn 需业务显式 `Use`）。
8. **N8 文档/脚本假绿**：`pre-release-check.sh:9,134-137,211-213` 仍校验/指导 v0.1.0-alpha + `push origin master`（实际 main）；`docs/optimization-summary.md:4` "29 文件 +595/−349"、"61 行 Validate 不强制 Secret"、"77 建议补测试"均与现状矛盾；`docs/architecture.md:345-347` "Zap 异步写入"错误；`docs/quick-start.md:43-44` max-idle 100 > max-open 20。

---

## 3. 可执行验证记录（摘录）

JWT 链路（`/tmp` 独立模块 + replace）：

```
stage0 未 Init: Sign err: jwt token is not initialized
                forged token -> HTTP=401 body={"code":401,"message":"Token was invalid."} reached=false
stage1 Init(Secret:""): Sign err: jwt token is not initialized; Authn 401
stage2 Init(Secret:"short"): Sign err=<nil> -> Authn HTTP=200 uid=42   ← 包级弱密钥未拦截
stage3 Init(32B): valid -> 200; missing/expired/forged/no-jti/HS512/alg=none -> 全 401
stage4 无 Redis: IsBlacklisted(false, "redis client is not initialized");默认 fail-open 放行;
                BlacklistFailClosed=true -> "token has in blacklist and expire"
config.Validate(secret="short"): jwt secret must be at least 32 bytes, got 5
env 覆盖: FRAMEWORK_JWT_TOKEN_ID=sub / FRAMEWORK_LOG_LEVEL=debug / FRAMEWORK_APP_VERSION=9.9.9 均生效
```

脚手架 e2e（`/tmp`，单条命令内完成）：

```
go build -o scaffold ./scaffold/cmd && ./scaffold new demoapp
demoapp/go.mod: module demoapp; go 1.25.1; require github.com/AuroraLZDF/go-framework v0.1.3 ...
无 replace: GOPROXY=off go mod tidy -> unknown revision v0.1.3（或 read-only 锁）  ← 阻断
加 replace: tidy/build/vet 全 0；启动(默认配置, 无 MySQL/Redis):
  /health  -> 200 {"code":200,"message":"ok","content":{"status":"healthy"}}
  /healthz -> 200 {"code":200,"message":"ok","content":{"status":"ok"}}
  /nope    -> 404 {"code":404,"message":"Page not found."}
  SIGTERM  -> exit 0 (优雅), stderr: write error: write logs/demoapp.2026-11-04.log: file already closed
坏日志目录 /dev/null/logs: NewApplication err=<nil>，zap 刷 write error: invalid argument，文件日志全丢
```

仓库侧：`go build ./...`、`go vet ./...`、`gofmt -l .`、`go test -count=1 ./...` 全绿（9 个包有测试 / 13 个测试文件，58 顶层测试/65 RUN），无 FAIL。

---

## 4. 文档逐项状态（docs-tests-auditor）

- **FIXED**：`code 200` 描述（architecture:126、QUICKSTART:65）、RegisterRoutes 双参签名（README:81/203、quick-start:89、CONFIG_USAGE:384、examples）、CONFIG_USAGE `blacklist-prefix`（:109）、gorm import、README:158 全路径 import、examples 池配置/shutdown-timeout/fail-closed。
- **STILL PRESENT**：
  - README：`:56 go get go-framework@latest`；`:68-70,195,214,232,255,272,284` bare `go-framework/...`；`:220 TokenID`；`:261 Dir:"logs"`（无此字段）；`:297 docs/api-reference.md`、`:298 docs/configuration.md` 不存在；`:5 v0.1.0-alpha`、`:6 go-1.23+`、`:8 静态 build-passing`；文档章节重复（:293 与 :373）。
  - quick-start：`:7 Go 1.23+`；`:21 bare go get`；`:43-44 max-idle 100 > max-open 20`；`:70-72 bare import`；`:159-160` 指向不存在的 examples；`:187 "logs/ 日期分片"` FAQ 错（默认 stdout）；`:192 yourorg`。
  - architecture：`:232 errno.ErrValidationFailed`、`:249 errno.ErrInvalidCode` 不存在；`:345-347 Zap 异步写入` 错；`:365` bare import。
  - CONFIG_QUICK_REFERENCE `:59,:91` BlacklistPath/13B secret；CONFIG_USAGE `:178,:204` 同。
  - optimization-summary：`:4,:61,:63,:77` 过时/矛盾，全部相对链接坏。
  - CHANGELOG：`:8` 最新 v0.1.1（无 v0.1.2）；`:93 "No unit tests yet"`、`:94 "HTTPS 未实现"` 过期。
  - scaffold README/DEVELOPMENT：仍宣传模板系统 + "所有验证测试通过"，与 generator 不引用 templates 矛盾。
- **新增**：secret ≥32B 要求未进任何文档；脚手架 v0.1.3 与 tag 矛盾；shutdown 日志回归；pre-release 假绿；PROJECT_ANALYSIS 未跟踪；.gitignore 缺 .screenshots/。

**测试缺口**（未覆盖高风险）：`core/log`（轮转/Init/Sync/Close/C(ctx)）、`server.Run`（优雅关闭/TLS/信号）、`db.InitMySQL/InitRedis`、真实 Redis cache、`validator.WithFilters`、generator/scaffold、password、version/verflag、service、context；token_test 直接改包级 config，无 parallel 保护。

---

## 5. 修复优先级建议

**P0（发布/合并前必须）**
1. 修 shutdown 日志顺序：`application.go:176` 的 `log.Info("Server exiting")` 移到 `app.Close()` 之前（或 Close 内先 flush 再关 writer），并加 server.Run + 轮转 write-after-close 回归测试。
2. 统一密钥文档与校验：README/quick-start/CONFIG_USAGE/CONFIG_QUICK_REFERENCE 示例改为 ≥32 字节并显式说明；`token.Init` 返回 error 并复用同一最小长度校验（避免 jwt 包反向依赖 config，可共享常量）。
3. 打并推送 `v0.1.3` tag（含工作区全部修复 + 双参签名），或让生成器依赖已发布版本；`verify-scaffold.sh` 去掉本地 replace 兜底，AI/CI 改为"无 replace 必须通过"。

**P1（下个迭代）**
4. 日志组件：per-App logger 或引用计数，避免单 Close 关全局；轮转失败 fail-loud（initLog 返回 error、rotate 失败重试/回退 console、Set dateStr 后再 OpenFile）；`log.C(nil)` 容错、`With()` 入接口。
5. JWT：`atomic.Pointer[Options]`/锁 + Reset；Authn 用 `errors.As` 映射 Expired/Blacklist/Missing；`BlacklistFailClosed` 默认值与 Redis 未配置时启动告警对齐；`ParseToken` 全路径返回 `*errno.Errno`；`ErrMissingHeader` 带 HTTP 语义；考虑 scheme 大小写兼容。
6. 生命周期：signal.Notify 提前注册；`NewApplication` 失败路径释放 log/DB；HTTP 超时/TLS 配置化；`/healthz` 区分 liveness/readiness；gin 设置 trusted proxies 与 MaxHeaderBytes；gin.Logger 移除/改 zap（当前 access log 绕过轮转且格式不统一）。
7. 发版说明标注三项 breaking：jti 强制（存量 token 全 401）、JWT secret ≥32B、errno HTTP 状态码语义化。

**P2（收尾）**
8. 数据层：InitMySQL 二次调用语义/重建、closeMySQL 先 Close 后置 nil、GetDB atomic.Pointer、Redis pool/timeout 接线、Ping 用 DialTimeout、DSN 字段转义、InitMySQL(nil) 防护。
9. util/validator：StructToMap panic、file fd 泄漏、ip 伪造回退、tree/time 缺陷；WithFilters 改显式列白名单 + LIKE 转义 + 未知 op 报错；正则 negative cache + 上限。
10. 文档批量同步（README/quick-start/architecture/CONFIG_* /optimization-summary/CHANGELOG/scaffold 文档）；pre-release 由 git tag/git describe 推导版本、校验 tag 与 version.go 一致、push origin main；删除/归档 templates 死文件；examples 可运行；补 CI（build/vet/test/test-scaffold）、.golangci.yml、.screenshots/ 忽略。
11. 测试补齐：log/server.Run/db init/WithFilters/generator；token_test 用 Init 而不是改包级 config。

**P3（可选）**
gin.New 早于 SetMode；ExpiresIn 应为剩余秒数而非绝对时间戳；errno Code 仍多为 1 无法区分；requestId 信任策略；`X-Username` 键名与 int64 类型语义；CORS 自定义头；password cost 12 + 策略；service 接口 ctx。

---

## 6. 结论

修复质量总体不错：上一轮最危险的 JWT 绕过链路已被真实关闭，配置/错误码/请求 ID/关闭幂等/脚手架生成内容都有实质改善，且新增测试确实在测行为而非空断言。

但"是否还有其他问题"的答案是**有，且仍有阻断项**：两条 P0 收尾（文档密钥示例导致启动失败、v0.1.3 未发布导致脚手架开箱不可用）+ 四条 P1 新回归（关闭日志顺序、全局 logger 生命周期、jwt 全局态竞态、公开 API 破坏性变更未说明）+ 日志轮转静默失败仍存。建议按 §5 顺序处理后再评估发布；当前状态下仍不建议对外宣称"开箱即用/生产可用"。
