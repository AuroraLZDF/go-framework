# 从 0.x 迁移到 1.0

> 适用对象：使用 **v0.1.x / v0.2.0** 的项目升级到 **v1.0.0**。
> 本文档逐条列出破坏性变更、最小修改方式与验证步骤；API 冻结范围见
> [API 稳定性与版本策略](api-stability.md)。

---

## 1. 升级路径

推荐直接从 **v0.2.0** 升级到 **v1.0.0**：

```bash
go get github.com/AuroraLZDF/go-framework@v1.0.0
go mod tidy
go build ./...
go test ./...
```

如果仍停留在 v0.1.0–v0.1.2，建议先升到 **v0.2.0** 跑通（该版本已包含 v0.1.3 的
全部破坏性变更与 v0.2.0 的请求/安全加固），再升 v1.0.0。v1.0.0 相对 v0.2.0
的破坏性变更很小（见 §3 最后一组），主要是 API 冻结与文档承诺。

**升级窗口提醒**：v0.1.3 起签发的 token 带 `jti`，且解析时强制要求 `jti`。
从 v0.1.2 及更早版本升级时，**所有已签发的 token 会在升级后立即失效**，
需要安排客户端重新登录（见 §5）。

---

## 2. 迁移清单（Checklist）

- [ ] 升级依赖：`go get github.com/AuroraLZDF/go-framework@v1.0.0`
- [ ] 把 `JWT secret` 换成至少 32 字节的随机值（`openssl rand -base64 32`）
- [ ] YAML 中 `jwt.blacklist-path` 改为 `jwt.blacklist-prefix`（若曾配置）
- [ ] 处理 `token.Init`、`log.Init`、`config.Load` 返回的 error（不再忽略）
- [ ] 确认 JWT 的用户 ID 是**数字**（`Authn` 会 `ParseInt` 成 `int64`）
- [ ] 受保护路由显式挂载 `middleware.Authn()`
- [ ] 更新路由注册函数为 `func(engine *gin.Engine, db *gorm.DB)`
- [ ] 检查业务代码是否依赖"错误响应也返回 HTTP 200"
- [ ] 检查是否依赖 `response.content` 在错误响应中存在（现在省略）
- [ ] 更新日志初始化：`log.Init` 返回 error；`log.NewLogger` 返回 `log.Logger`
- [ ] 配置补齐新键（超时、`max-body-bytes`、`trusted-proxies`、Redis 池）
- [ ] 重新生成/升级脚手架项目（`scaffold new` 或手动改 `go.mod` 依赖版本）
- [ ] 运行本文 §8 的验证清单

---

## 3. 破坏性变更总表

### 3.1 v0.1.3 引入（从更早版本升级时必看）

| # | 变更 | 影响 | 迁移动作 |
|---|------|------|----------|
| 1 | JWT Secret 必须 ≥32 字节 | 短密钥在 `config.Validate`/`token.Init` 阶段直接报错，服务无法启动 | 更换随机密钥 |
| 2 | 每个 token 带 `jti`，解析时**强制要求** | 升级前签发的 token 全部 401 | 安排重新登录（§5） |
| 3 | `token.Init(opts *Options)` → `Init(opts *Options) error` | 不处理返回值会编译错误/忽略错误 | `if err := token.Init(...); err != nil { ... }`；通常无需手动调用（`NewApplication` 自动接线） |
| 4 | 认证错误细分 | `ParseRequest` 返回 `*errno.Errno`（缺失/过期/非法/已吊销） | 需要分支判断时用 `errors.Is(err, errno.ErrTokenExpired)` 等 |
| 5 | `ExpiresIn` 由绝对 Unix 时间戳改为**剩余秒数** | 客户端若把 `expires_in` 当时间戳用会错误 | 按"相对秒数"处理，或改成 `now.Add(expires_in)` |
| 6 | 认证中间件必须**显式挂载** | 未挂载的受保护路由不再被保护（行为/安全变化） | `v1.Use(middleware.Authn())` |
| 7 | errno 的 HTTP 状态码语义化 | 错误响应不再是统一 HTTP 200（401/404/413/429/5xx） | 客户端按 HTTP 状态 + `code` 判断；网关/监控规则同步 |
| 8 | `errno.Decode` 对未知错误返回统一 500 文案 | 不再回显 `err.Error()` | 需要排查时看服务端日志（含 request_id） |
| 9 | `RouteRegisterFunc` → `func(*gin.Engine, *gorm.DB)` | 旧签名 `func(*gin.Engine, interface{})` 编译失败 | 回调第二个参数改为 `*gorm.DB` |
| 10 | `log.Init` 返回 error；轮转/目录不可写时启动失败 | 依赖"静默降级"的部署会启动失败 | 修正日志目录权限/配置，处理 error |
| 11 | `log.Close` 改为引用计数 | 多次 Init 后需同样次数 Close；单次 Close 不再关闭仍在使用的文件 | 通常由 `Application.Close` 管理，无需手动配对 |
| 12 | `Request ID` 校验并写入 `r.Context()` | 非法 `X-Request-ID` 会被替换为服务端 UUID | 如需透传自定义 ID，确保符合 `[A-Za-z0-9._-]{1,128}` |
| 13 | `validator` 包由 `tools` 改名 | 旧导入路径 `core/validator/tools` 不存在 | 改为 `github.com/AuroraLZDF/go-framework/core/validator` |
| 14 | 配置支持全量环境变量覆盖，配置文件可选 | `Load("")` + 环境变量即可启动 | 可选：移除必须存在的 `config.yaml` 依赖 |
| 15 | MySQL DSN：`loc=UTC` + 显式超时；连接池默认值与钳制 | 时间统一按 UTC；DSN 变化影响旧库时区假设 | 确认业务时间处理；`DSN()` 现在返回 `(string, error)` |
| 16 | Go 版本要求 1.26.0（v1.0.0 起） | 旧工具链无法构建 | 升级 Go |

### 3.2 v0.2.0 引入

| # | 变更 | 影响 | 迁移动作 |
|---|------|------|----------|
| 17 | `server.trusted-proxies` 默认 `[]` | `c.ClientIP()`/访问日志不再采信 `X-Forwarded-For` | 部署在反向代理后时显式配置代理 IP/CIDR |
| 18 | `server.max-body-bytes`（默认 4MB） | 超过上限的请求返回 413 | 大文件上传类接口调大该值，或改用流式接口 |
| 19 | 仅配 `https-addr` 时不再隐式监听 `:8080` | 依赖 `:8080` 的健康检查/客户端连不上 | 显式配置 `http-addr`（此时该端口只做 308 重定向） |
| 20 | 新增 `GET /readyz`；`/healthz` 仅为 liveness | 探针语义变化 | 依赖检查改用 `/readyz` |

### 3.3 v1.0.0 引入

| # | 变更 | 影响 | 迁移动作 |
|---|------|------|----------|
| 21 | `verflag.VersionValue` 导出（原类型不可命名） | 使用 `verflag.Version/VersionVar` 的 CLI 代码现在可用 | 无需改动；若曾因无法编译而自行实现，可迁回 |
| 22 | `log.NewLogger` 返回 `log.Logger` 接口（原为未导出 `*zapLogger`） | 声明为具体类型的变量会编译失败 | 用 `logger := log.NewLogger(opts)` 或显式 `var l log.Logger` |
| 23 | `middleware.Cors` 标记 Deprecated | 仍可用，v2 移除 | 改用 `middleware.CORS()` |
| 24 | `token.ErrMissingHeader`/`ErrBlacklistToken`/`ErrParseToken` 标记 Deprecated | 仍可用，v2 移除 | 改用 `errno.ErrTokenMissing/ErrTokenBlacklisted/ErrTokenInvalid` |

> 业务码 `errno.Code` 的取值会在 `v1.0.0-rc.1` 之前统一（当前多个错误共用 `1`）。
> 如果你的客户端按 `code` 做分支，请在 rc 发布说明中确认最终取值后再固化；
> HTTP 状态码与响应信封结构已稳定。

---

## 4. 逐项迁移示例

### 4.1 JWT

```go
// 旧（v0.1.2）：Init 无返回值，且只生效一次；Secret 无长度校验
token.Init(&token.Options{TokenId: "uid", Secret: "short"})

// 新（v1.0）：Init 返回 error，Secret ≥32 字节；NewApplication 会自动接线
if err := token.Init(&token.Options{
    Secret:  "请使用 openssl rand -base64 32 生成", // ≥32 字节
    Expire:  24,                                  // 小时
    TokenId: "uid",                               // 必须是数字用户 ID 的 claim
}); err != nil {
    log.Fatal("init jwt failed", "err", err)
}

// 签发：参数须为数字用户 ID（Authn 内部 ParseInt）
resp, err := token.Sign("12345")

// 吊销：登出/踢人
token.AddToBlacklist(r)        // 吊销当前会话（jti 级别）
token.FailureAllClient("12345") // 吊销该用户全部 token（epoch 级别）
```

错误分类：

```go
_, err := token.ParseRequest(r)
switch {
case errors.Is(err, errno.ErrTokenMissing):   // 401 缺 Authorization
case errors.Is(err, errno.ErrTokenExpired):   // 401 过期
case errors.Is(err, errno.ErrTokenBlacklisted): // 401 已吊销
case errors.Is(err, errno.ErrTokenInvalid):   // 401 非法
}
```

黑名单配置迁移：

```yaml
# 旧
jwt:
  blacklist-path: ./blacklist.json

# 新
jwt:
  blacklist-prefix: "jwt:blacklist:"   # Redis key 前缀
  blacklist-fail-closed: false         # Redis 不可用时：false=放行（默认），true=拒绝
```

> 黑名单已从 JSON 文件迁移到 Redis。未配置 Redis 时默认 **fail-open**（放行并打
> WARNING）；对安全要求高的场景请配置 `blacklist-fail-closed: true`。

### 4.2 errno 与响应

```go
// 自定义错误：保持 *errno.Errno 或包装它（errors.As 可解包）
return errno.InternalServerError.SetMessage("查询用户失败: %d", uid)

// 业务错误：直接 response.Error(c, err)
response.Error(c, errno.ErrUnauthorized)

// 成功响应：HTTP 200，信封 code=200
response.Success(c, data, "ok")
```

客户端变化：

- 失败响应的 HTTP 状态码不再固定 200：401/404/413/429/5xx；
- 错误响应省略 `content` 字段（`omitempty`），应为 `{"code":401,"message":"Token missing."}`；
- 成功响应 `code` 从旧版的 `0` 变为 `200`（`errno.OK.Code`）；
- 5xx 只返回通用文案，原始错误见服务端日志。

### 4.3 日志

```go
// 旧
log.Init(log.NewOptions()) // 无返回值

// 新
if err := log.Init(opts); err != nil { // 轮转目录/文件不可写会返回错误
    return fmt.Errorf("init log: %w", err)
}

// 派生带请求 ID 的 logger：NewApplication 已把 request_id 写入 r.Context()
logger := log.C(c.Request.Context())
logger.Info("handled request")

// NewLogger 返回接口（不再返回未导出的具体类型）
logger := log.NewLogger(log.NewOptions())
```

- 全局 logger 仍由 `Application` 初始化/关闭；应用内不要重复 `Init` 后再手动
  `Close`，除非你清楚引用计数语义。
- 访问日志已改由 zap 输出（不再使用 `gin.Logger()`），格式字段见 API 稳定性文档 §7.2。

### 4.4 配置

```yaml
# 旧（v0.1.2 典型配置）
server:
  http-addr: ":8080"
jwt:
  blacklist-path: "./blacklist.json"

# 新（v1.0 完整示例，未列出的键会使用默认值）
app:
  name: my-app
  mode: release
server:
  http-addr: ":8080"
  shutdown-timeout: 10s
  read-header-timeout: 10s
  read-timeout: 30s
  write-timeout: 60s
  idle-timeout: 120s
  max-header-bytes: 1048576
  max-body-bytes: 4194304
  trusted-proxies: []      # 反向代理后填 IP/CIDR
  # https-addr: ":8443"
  # tls-cert: ./cert/server.crt
  # tls-key: ./cert/server.key
jwt:
  secret: "<32 字节以上随机值>"
  expire: 24
  token-id: uid
  blacklist-prefix: "jwt:blacklist:"
  blacklist-fail-closed: false
redis:
  addr: 127.0.0.1:6379     # 也支持只写 host + port: 6379
  pool-size: 10
  dial-timeout: 5s
  read-timeout: 3s
  write-timeout: 3s
log:
  level: info
  format: json
  dir: logs
```

环境变量：所有内置键都可用 `FRAMEWORK_` 前缀覆盖（`.`/`-` 转 `_`），例如
`FRAMEWORK_SERVER_HTTP_ADDR`、`FRAMEWORK_JWT_SECRET`、
`FRAMEWORK_SERVER_TRUSTED_PROXIES="127.0.0.1,10.0.0.0/8"`（列表用逗号分隔）。
`config.Load("")` 在找不到配置文件时不再报错，可仅靠环境变量 + 默认值启动。

### 4.5 服务器与路由

```go
// 旧：Second 参数是 interface{}
app.RegisterRoutes(func(engine *gin.Engine, db interface{}) { ... })

// 新：*gorm.DB
app.RegisterRoutes(func(engine *gin.Engine, db *gorm.DB) {
    engine.GET("/hello", func(c *gin.Context) {
        response.Success(c, gin.H{"message": "hello"}, "ok")
    })

    v1 := engine.Group("/api/v1")
    v1.Use(middleware.Authn()) // 受保护路由必须显式挂载
    v1.GET("/profile", profileHandler)
})

// Run 返回 error（监听失败/关闭失败可感知）；信号处理已内置
if err := app.Run(context.Background()); err != nil {
    log.Error("server exited", "err", err)
}
```

### 4.6 数据库

```go
// DSN() 现在必须处理 error
dsn, err := opts.DSN()
if err != nil { return err }

// InitMySQL 的参数为 *db.MySQLOptions（注意与 config.MySQLOptions 区分）
gdb, err := db.InitMySQL(&db.MySQLOptions{
    Host: "127.0.0.1:3306", Username: "root", Password: "root", Database: "mydb",
    MaxOpenConnections: 100, MaxIdleConnections: 10,
    MaxConnectionLifeTime: time.Hour, LogLevel: 4,
})
if err != nil { return err }

// 重复初始化返回 db.ErrDBAlreadyInitialized；需要重来先 Close()
// 未初始化时 GetDB()/GetRedis() 返回 nil（不再 panic）
```

### 4.7 工具与校验

- `util.ParamToInt64(c, key)` 现在读取 **path 参数**（`c.Param`），不再读 query；
- `util.GenerateRandomString` 使用 `crypto/rand`，`length <= 0` 返回空串；
- `util.StructToMap` 返回 `map[string]string`，支持指针/嵌入/`json` tag，忽略未导出字段；
- `util.AddFile` 使用 `O_EXCL`（文件已存在时报错），需要"不存在才创建"用 `AddFileIfNotExists`；
- `validator.WithFilters`：字段名仅允许字母/数字/下划线（不再接受 `.`），
  LIKE 值会转义 `%`/`_`/`\`，未知运算符会返回错误（fail-closed，同时写入 `db.AddError`）。

### 4.8 脚手架项目

生成项目现在依赖框架 **v0.2.0**（v1.0.0 发布后为 v1.0.0）：

```bash
# 方式一：重新生成（推荐）
go install github.com/AuroraLZDF/go-framework/scaffold/cmd@v1.0.0
scaffold new my-app

# 方式二：手动升级已有项目
go get github.com/AuroraLZDF/go-framework@v1.0.0
```

生成的项目包含可直接启动的 `config.yaml`（JWT secret 已满足 ≥32 字节）与
`config.example.yaml`、`/health` 业务示例路由、`scaffold version` 子命令。

---

## 5. 旧 Token 与升级顺序

升级到 v0.1.3+ 后，所有旧 token 因缺少 `jti` 会被拒绝。可选策略：

1. **直接失效（最简单）**：发布说明告知用户重新登录；适合 token 有效期较短的场景。
2. **平滑过渡**：先在旧版本中把 token 有效期调短，再在低峰期升级；或维护一个
   短期的双解析兼容层（仅解析、不校验 jti），过渡期内逐步淘汰旧 token。
3. **灰度**：新版本上线后保留旧版本实例一小段时间（需自行实现按 token 路由）。

注意：即使旧 token 失效，**签名密钥不要更换为空/默认值**；如需强制全量重登，
可以在受控时间点轮换 secret。

---

## 6. 客户端契约变化（务必同步前端/调用方）

| 项目 | 旧行为 | 新行为 |
|------|--------|--------|
| 成功响应 `code` | `0` | `200` |
| 失败响应 HTTP 状态 | 几乎总是 `200` | `401/404/413/429/5xx` |
| 失败响应 body | 含 `content`（通常 `null`） | 省略 `content` 字段 |
| 认证失败文案 | 统一"Token was invalid." | 区分 missing/expired/revoked/invalid |
| `expires_in` | 绝对 Unix 时间戳 | 剩余秒数 |
| `X-Request-ID` | 未校验、未进日志上下文 | 校验格式，非法则重生成，并贯穿日志 |
| 413 | 不存在 | 请求体超过 `max-body-bytes` 时返回 |
| `/readyz` | 不存在 | 新增就绪探针（200/503） |

---

## 7. 常见报错对照

| 现象/报错 | 原因 | 修复 |
|-----------|------|------|
| `jwt secret is too short: got N bytes, want at least 32` | Secret <32 字节 | 换随机 32B+ 密钥 |
| `jwt token is not initialized` | 直接使用 token 包但未调用 `Init`（不经过 `NewApplication`） | 调用 `token.Init` 或改用 `NewApplication` |
| `Token missing.` / `Token expired.` | 缺少/过期 Authorization | 客户端补 token / 刷新 |
| 升级后所有请求 401 | 旧 token 无 `jti` | 重新登录 |
| `Authn` 通过但用户 ID 为 0/报 invalid | token 的用户 ID 非数字 | 改用数字用户 ID，或自定义认证中间件 |
| `log: failed to ...` 启动失败 | 日志目录不可写 | 修正目录权限或 `log.dir` |
| `invalid server.trusted-proxies entry "..."` | 配置了非法 IP/CIDR | 填 `127.0.0.1`、`10.0.0.0/8` 等 |
| `redis.addr` 启动报错 | 写了 `redis://` scheme 或 `host:6379:extra` | 写 `host` 或 `host:port` |
| 客户端拿到 413 | 请求体 > `max-body-bytes` | 调大配置或改流式上传 |
| 客户端报 `x509` / 连不上 8080 | 仅配 HTTPS 时不再监听明文 | 显式配置 `http-addr` 做 308 |
| 编译错误 `cannot use func(*gin.Engine, interface{})` | `RouteRegisterFunc` 双参变更 | 第二参改 `*gorm.DB` |
| 编译错误 `token.Init(...) used as value` | Init 现返回 error | 处理返回值 |
| 编译错误 `unknown field BlacklistPath` | 配置字段改名 | YAML 用 `blacklist-prefix` |

---

## 8. 升级验证

```bash
# 1. 构建与测试
go build ./... && go vet ./... && go test ./...

# 2. 启动并验证探针
curl -i http://127.0.0.1:8080/healthz   # 200
curl -i http://127.0.0.1:8080/readyz    # 200（未配置依赖时 checks 为 skipped）

# 3. 认证链路
curl -i http://127.0.0.1:8080/api/v1/profile            # 401 {"code":401,"message":"Token missing."}
TOKEN=...  # 登录接口返回
curl -i -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8080/api/v1/profile  # 200

# 4. 统一信封与 404
curl -i http://127.0.0.1:8080/not-exist   # 404 统一信封

# 5. 请求体上限（可选）
head -c 5000000 /dev/zero | curl -i -X POST --data-binary @- http://127.0.0.1:8080/upload  # 413
```

使用框架自带的发布检查（推荐在 CI 中执行）：

```bash
FRAMEWORK_LOCAL=1 make test-scaffold   # 用本地代码跑脚手架 e2e
bash scripts/pre-release-check.sh      # 版本/构建/gofmt/文档结构检查
```

---

## 9. 回滚说明

- **数据层**：v1.0.0 未引入数据库 schema 变更，回滚到 v0.2.0 不需要迁移数据库。
- **Token**：v0.2.0 与 v1.0.0 签发的 token 格式一致（均含 `jti`），回滚不会使
  新 token 失效；密钥保持一致即可。
- **配置**：v1.0.0 新增的配置键在 v0.2.0 中会被忽略（不影响启动），但依赖
  `trusted-proxies`/`max-body-bytes` 的保护在回滚后失效，需评估风险。
- **客户端**：如果已按新契约（HTTP 状态码/`code`/`expires_in`）适配，回滚到
  0.1.x 前需确认兼容；建议回滚目标为 v0.2.0，而不是更早版本。

---

## 10. FAQ

**Q：我只用 `config.Load` + 自己的 gin 路由，不用 `server.NewApplication`，需要做什么？**
A：需要自己调用 `token.Init`（否则签发/解析报未初始化）、`log.Init`（否则使用默认
stdout logger）、`db.Init*`（如需数据库），并按需挂载 `middleware.Authn`、
`RequestID`、`CORS` 等中间件。

**Q：为什么业务码还不能直接用于客户端分支？**
A：当前多个错误共用 `Code: 1`（历史遗留）。`v1.0.0-rc.1` 前会统一为可区分的
业务码；在此之前建议用 HTTP 状态码 + `message` 判断，或从 rc 起以 `code` 为准。

**Q：能否继续用 `jwt.Init` 这个写法？**
A：包标识符是 `token`，不是 `jwt`。可以写
`import fwjwt "github.com/AuroraLZDF/go-framework/core/auth/jwt"` 后使用
`fwjwt.Init(...)`；推荐直接用包名 `token`，与 `golang-jwt/jwt/v5` 同时导入时更清晰。

**Q：v1.0.0 会删除 `core/service` 吗？**
A：不会。它被标记为 Experimental（实验性），不承诺兼容，但会保留；详见
[API 稳定性文档](api-stability.md#2-稳定性分级)。

---

## 相关链接

- [API 稳定性与版本策略](api-stability.md)
- [配置使用指南](CONFIG_USAGE.md) · [配置快速参考](CONFIG_QUICK_REFERENCE.md)
- [快速开始](quick-start.md) · [架构设计](architecture.md)
- [CHANGELOG](../CHANGELOG.md) · [Releases](https://github.com/AuroraLZDF/go-framework/releases)
