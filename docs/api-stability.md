# API 稳定性与版本策略

> 状态：**v1.0.0 冻结草案**（自 `v1.0.0-rc.1` 起生效）。
> 本文档是 1.x 系列的兼容性承诺：除本文明确列出的实验性/内部包外，1.x 的
> minor 版本不会破坏源码兼容，破坏性变更只在 major 版本（v2.0.0+）发生。
>
> 相关文档：[配置使用指南](CONFIG_USAGE.md) · [架构设计](architecture.md) ·
> [0.x → 1.0 迁移指南](migration-0.x-to-1.0.md) · [CHANGELOG](../CHANGELOG.md)

---

## 1. 版本与兼容承诺

- 本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)：
  - **PATCH**（1.0.x）：只包含向后兼容的缺陷与安全修复；
  - **MINOR**（1.x.0）：新增向后兼容的 API、配置项、错误码与行为改进；
  - **MAJOR**（2.0.0+）：允许删除/重命名导出标识符、修改签名或改变行为契约。
- **源码兼容**指：升级到 1.x 的新版本后，使用稳定 API 的代码无需修改即可
  编译通过，且不会因升级而出现新的运行时错误。
- 模块路径固定为 `github.com/AuroraLZDF/go-framework`，1.x 内不会修改。
- 支持的 Go 版本：`go.mod` 声明的版本及其后两个 minor（当前为 Go 1.25+）。
- 依赖大版本出现在公开签名中的（见 §5），1.x 内不会升级其 major 版本。

### 1.1 什么算破坏性变更

| 变更 | 是否破坏兼容 |
|------|--------------|
| 删除或重命名导出标识符（函数/类型/变量/常量/字段/方法） | ✅ 破坏 |
| 修改导出函数签名、返回值或接口方法集合 | ✅ 破坏 |
| 修改结构体导出字段的类型、标签或删除字段 | ✅ 破坏 |
| 修改导出常量的值、错误变量的语义 | ✅ 破坏 |
| 删除/重命名内置配置键或环境变量 | ✅ 破坏 |
| 修改统一响应信封字段、HTTP 状态码语义、内置路由语义 | ✅ 破坏 |
| 新增导出标识符、新增配置键（带向后兼容默认值） | ❌ 不破坏 |
| 增加错误分支、修正文档、内部实现重构 | ❌ 不破坏 |
| 修复安全漏洞导致的行为变化 | ⚠️ 例外，会在 CHANGELOG 的 Security 段说明 |

---

## 2. 稳定性分级

| 级别 | 范围 | 承诺 |
|------|------|------|
| **Stable（稳定）** | `core/` 下除 `core/service` 外的所有包；`scaffold` 命令行 | 1.x 内源码兼容；按 §4 弃用流程演进 |
| **Experimental（实验性）** | `core/service` | 不承诺兼容，可能在 minor 中调整或删除 |
| **Internal / 非 API** | `scaffold/internal/**`、`examples/**`、`scripts/**`、`.github/**`、`Makefile`、`docs/**` | 不承诺任何兼容性；可随时修改 |

### 2.1 Stable 包清单

| 包（导入路径末尾） | 包标识符 | 说明 |
|--------------------|----------|------|
| `core/auth/jwt` | `token` | JWT 签发、解析、吊销（注意包名与目录名不同） |
| `core/auth/password` | `password` | bcrypt 密码哈希 |
| `core/config` | `config` | 配置加载、默认值、校验 |
| `core/context` | `known` | 请求上下文键与请求头常量 |
| `core/database/db` | `db` | MySQL / Redis 单例初始化与访问 |
| `core/errno` | `errno` | 统一错误定义与解码 |
| `core/log` | `log` | zap 日志、轮转、访问日志 |
| `core/middleware` | `middleware` | Authn / CORS / RequestID / BodyLimit 等 |
| `core/response` | `response` | 统一响应封装 |
| `core/server` | `server` | Application 生命周期与路由注册 |
| `core/util` | `util` | 缓存、分页、文件、时间、树等工具 |
| `core/validator` | `validator` | 校验器、查询过滤条件 |
| `core/version` | `version` | 版本信息 |
| `core/version/verflag` | `verflag` | CLI `--version` 辅助（供命令行工具使用） |

`scaffold` 命令的**契约**：`scaffold new <name|path>`、`scaffold version`、
`scaffold --version` 三个入口的行为与退出码；生成的项目文件集合可能在
minor 中调整（新增/精简文件不算破坏），但生成的项目必须能在对应框架版本下
`go build` 并启动。

---

## 3. 冻结的公共 API（v1.0）

以下签名自 v1.0.0 起冻结。完整列表可用 `go doc -all <包>` 查看。

### 3.1 core/auth/jwt（package token）

```go
const MinSecretLength = 32

type Options struct {
    TokenId             string // 存放用户 ID 的 claim 名，默认 "uid"
    Secret              string // 至少 MinSecretLength 字节
    Expire              int    // 有效期（小时），默认 24
    BlacklistPrefix     string // 黑名单 key 前缀，默认 "jwt:blacklist:"
    BlacklistFailClosed bool   // Redis 不可用时是否拒绝请求，默认 false
}

func Init(opts *Options) error   // 校验 Secret；可由 server.NewApplication 自动调用
func Reset()                     // 清空配置，仅供测试/进程重配
func Sign(tokenId string) (*response.TokenResponse, error)
func ParseToken(r *http.Request) (*jwt.Token, error)   // 依赖 golang-jwt/v5 类型，高级 API
func ParseRequest(r *http.Request) (string, error)     // 返回数字用户 ID；含 jti/黑名单校验
func IsBlacklisted(token *jwt.Token) (bool, error)     // 高级 API
func AddToBlacklist(r *http.Request) error
func FailureAllClient(tokenId string) error            // 吊销某用户的全部 token
```

已弃用但保留至 v2 的错误变量：`ErrMissingHeader`、`ErrBlacklistToken`、
`ErrParseToken`（错误分类改由 `*errno.Errno` 表达）。

### 3.2 其余 Stable 包

```go
// core/auth/password
func Encrypt(source string) (string, error)
func Compare(hashedPassword, password string) error

// core/config
func Load(configPath string) (*Config, error) // configPath 为空时搜索默认位置，可仅靠环境变量
func (c *Config) Validate() error             // 填充默认值并校验；NewApplication 会调用
type Config / AppConfig / ServerConfig / DatabaseConfig / MySQLOptions / RedisConfig / JWTConfig / LogConfig

// core/context (package known)
const XRequestIDKey = "X-Request-ID"
const XUsernameKey = "X-Username"
const XAuthorization = "Authorization"
const XTokenType = "Bearer"

// core/database/db
var ErrDBAlreadyInitialized error
type MySQLOptions struct{ Host, Username, Password, Database string; MaxIdleConnections, MaxOpenConnections int; MaxConnectionLifeTime time.Duration; LogLevel int }
func (o *MySQLOptions) DSN() (string, error)
type RedisOptions struct{ Addr, Password string; DB, PoolSize, MinIdleConns int; DialTimeout, ReadTimeout, WriteTimeout time.Duration }
func InitMySQL(opts *MySQLOptions) (*gorm.DB, error)
func InitRedis(opts *RedisOptions) (*redis.Client, error)
func GetDB() *gorm.DB
func GetRedis() *redis.Client
func Close() error
func CloseRedis() error

// core/errno
type Errno struct{ HTTP, Code int; Message string }
func (e *Errno) Error() string
func (e *Errno) SetMessage(format string, args ...interface{}) *Errno
func Decode(err error) (httpCode, code int, message string)
var OK, InternalServerError, ErrPageNotFound, ErrRequestEntityTooLarge, ErrBind,
    ErrInvalidParameter, ErrSignToken, ErrTokenInvalid, ErrUnauthorized,
    ErrServiceUnavailable, ErrDBQuery, ErrPhoneCodeLimit, ErrPhoneCodeTimeLimit,
    ErrTokenMissing, ErrTokenExpired, ErrTokenBlacklisted,
    ErrTokenRefreshExpired, ErrTokenRefreshFailed, ErrTokenRefreshInvalid *Errno

// core/log
type Logger interface{ Debug/Info/Warning/Error/Panic/Fatal(msg string, kv ...interface{}); C(context.Context) Logger; Sync(); Close() }
type Options struct{ Level, Format string; DisableCaller, DisableStacktrace bool; OutputPaths []string; Rotation *RotateConfig }
type RotateConfig struct{ Interval, BasePath string; Stdout bool }
func Init(opts *Options) error
func Reset()
func Close() error
func Sync()
func Debug/Info/Warning/Error/Panic/Fatal(msg string, kv ...interface{})
func C(ctx context.Context) Logger
func NewLogger(opts *Options) Logger   // 不注册为全局 logger；失败回退 stdout
func NewOptions() *Options
type RotateWriter struct{ /* ... */ }
func NewRotateWriter(basePath, interval string) (*RotateWriter, error)
var ErrRotateWriterClosed error

// core/middleware
func Authn() gin.HandlerFunc
func CORS(allowedOrigins ...string) gin.HandlerFunc
func BodyLimit(maxBytes int64) gin.HandlerFunc
func IsRequestTooLarge(err error) bool
func RequestID() gin.HandlerFunc
func NoCache(c *gin.Context)
func Secure(c *gin.Context)
func Cors(c *gin.Context) // Deprecated: 使用 CORS()

// core/response
type Response struct{ Code int `json:"code"`; Message string `json:"message"`; Content interface{} `json:"content,omitempty"` }
type TokenResponse struct{ AccessToken string `json:"access_token"`; TokenType string `json:"token_type"`; ExpiresIn int64 `json:"expires_in"` }
func Success(c *gin.Context, data any, message string)
func Error(c *gin.Context, err error)

// core/server
type RouteRegisterFunc func(*gin.Engine, *gorm.DB)
func NewApplication(cfg *config.Config) (*Application, error)
func (app *Application) Run(ctx context.Context) error
func (app *Application) Close() error
func (app *Application) GetEngine() *gin.Engine
func (app *Application) GetDB() *gorm.DB
func (app *Application) RegisterRoutes(registerFunc RouteRegisterFunc)

// core/util
var ErrCacheUnavailable error
type Cache interface{ Get/Set/Delete/Exists/Close }
func NewCache() Cache
func AddDir(path string) error
func AddFile(path string) error
func AddFileIfNotExists(path string) error
func AssignIfNotNil[T any](dst *T, src *T)
func GenerateRandomString(length int) string
func GetRealClientIP(r *http.Request) string
func IsExpired(lastTime time.Time, duration time.Duration) bool
func IsWithinDuration(lastTime time.Time, duration time.Duration) bool
func JoinInt64s(nums []int64) string
func ParamToInt64(c *gin.Context, key string) (int64, error)
func StructToMap(obj interface{}) map[string]string
type Time struct{ /* ... */ } // FromTime/Now/MarshalJSON/UnmarshalJSON/Scan/Value/String/ToTime
type TreeService struct{ /* ... */ } // NewTreeService/MakeTree/MakeTreeForHTML

// core/validator
func DefaultLimit(limit int) int
func RegisterCustomValidations(v *validator.Validate) error
func ValidateRegex(fl validator.FieldLevel) bool
func JSONToMarkdown(data interface{}) (string, error)
type Filter / FilterOperator // 运算符 =, LIKE, >, <, >=, <=, IN, NOT IN
type QueryOption func(*gorm.DB) *gorm.DB // WithArchived/WithCode/WithFilters/WithId/WithMcId/WithName/WithPhone/WithStatus/WithTitle/WithType/WithUserId

// core/version
var GitVersion, GitCommit, GitTreeState, BuildDate string
type Info struct{ GitVersion, GitCommit, GitTreeState, BuildDate, GoVersion, Compiler, Platform string }
func Get() Info
func (i Info) String() string
func (i Info) Text() ([]byte, error)
func (i Info) ToJSON() string

// core/version/verflag（CLI 辅助）
type VersionValue int
const VersionFalse, VersionTrue, VersionRaw VersionValue
func VersionVar(p *VersionValue, name string, value VersionValue, usage string)
func Version(name string, value VersionValue, usage string) *VersionValue
func PrintAndExitIfRequested()
```

---

## 4. 弃用（Deprecation）政策

1. 弃用必须使用 Go 官方约定：在文档注释中以 `Deprecated: ` 开头说明替代方案。
2. 被弃用的导出标识符**至少保留一个 minor 版本**且保持功能可用，只允许在
   **下一个 major**（v2.0.0）删除。
3. 每次弃用都要在同版本 CHANGELOG 的 `Deprecated` 段记录（若无该段则新增），
   并在迁移指南中给出替代写法。
4. 弃用不打印运行期告警，避免污染日志；以编译期文档提示为准。
5. 紧急安全修复可以立即改变被弃用标识符的行为，但需在 CHANGELOG 的
   `Security` 段明确说明。

当前弃用项：

| 标识符 | 自 | 替代 | 计划移除 |
|--------|----|------|----------|
| `middleware.Cors` | v1.0.0-rc.1 | `middleware.CORS()` | v2.0.0 |
| `token.ErrMissingHeader` | v0.1.3 | `errno.ErrTokenMissing` | v2.0.0 |
| `token.ErrBlacklistToken` | v0.1.3 | `errno.ErrTokenBlacklisted` | v2.0.0 |
| `token.ErrParseToken` | v0.1.3 | `errno.ErrTokenInvalid` 等 | v2.0.0 |

---

## 5. 依赖面与公开签名

以下第三方类型出现在 Stable API 的签名中，1.x 内保持其 major 版本不变：

| 依赖 | 出现在 | 承诺 |
|------|--------|------|
| `github.com/gin-gonic/gin` (v1) | `*gin.Context`、`*gin.Engine`、`gin.HandlerFunc` | v1.x 不变 |
| `gorm.io/gorm` (v1) | `*gorm.DB`、`QueryOption` | v1.x 不变 |
| `github.com/redis/go-redis/v9` | `*redis.Client` | v9.x 不变 |
| `github.com/golang-jwt/jwt/v5` | `*jwt.Token`（`ParseToken`/`IsBlacklisted`） | v5.x 不变 |
| `github.com/go-playground/validator/v10` | `validator.FieldLevel`、`*validator.Validate` | v10.x 不变 |

`ParseToken`、`IsBlacklisted` 是**高级 API**：它们把 `*jwt.Token` 暴露给调用方，
适合需要自行读取 claims 的场景。普通业务应优先使用 `ParseRequest` 和
`middleware.Authn`。若未来需要更换 JWT 实现，只能在 major 版本进行。

---

## 6. 进程模型与全局状态（冻结决策）

框架的 **受支持进程模型是：一个进程一个 `server.Application`**。
以下状态是进程级全局的，多个 Application 并存不受支持：

| 全局状态 | 行为 | 线程安全 |
|----------|------|----------|
| JWT 配置 | `token.Init` 原子替换配置，重复调用 **last-wins**；`server.NewApplication` 自动接线 | `atomic.Pointer`，读写无数据竞争 |
| 日志 logger | `log.Init` 替换全局 logger 并增加引用计数；`log.Close` 递减，归零时才关闭轮转文件 | 互斥锁 + 原子指针 |
| MySQL/Redis 单例 | `db.Init*` 复用已初始化实例；MySQL 重复初始化返回 `ErrDBAlreadyInitialized`；`Close/CloseRedis` 后可以重新初始化 | 互斥锁 |

约定：

- 多个 Application 共用同一份日志与 JWT 配置，后创建者会覆盖先创建者的配置，
  因此**不要**在一个进程中启动多个 Application。需要多实例时请使用多进程
  （容器/进程管理器）。
- 测试中需要隔离时使用 `token.Reset()`、`log.Reset()`、`db.Close()`/`db.CloseRedis()`，
  并在 `t.Cleanup` 中恢复。这些 Reset 函数仅用于测试，不属于生产用法。
- 业务代码可以安全地并发调用 `token.Sign/ParseRequest`、日志函数与 `db.GetDB()`
  的只读访问；不要并发调用 `Init/Close/Reset` 与其他操作。

---

## 7. 配置与运行时契约

### 7.1 配置键

内置配置键、环境变量映射与默认值见 [CONFIG_USAGE.md](CONFIG_USAGE.md)。
冻结规则：

- 键名（YAML 路径）与环境变量 `FRAMEWORK_<PATH>` 属于 Stable API，删除或改名是破坏性变更；
- 新增配置键必须提供向后兼容的默认值（0 值/未配置时行为不变）；
- 除安全修复外，**默认值变更**视为行为变更，需在 CHANGELOG 中说明并在迁移指南中提示。

当前默认值（`Config.Validate` 自动填充）：

| 键 | 默认值 |
|----|--------|
| `app.mode` | `release` |
| `server.http-addr` | `:8080`（仅当 http-addr 与 https-addr 都为空） |
| `server.shutdown-timeout` | `10s` |
| `server.read-header-timeout` / `read-timeout` / `write-timeout` / `idle-timeout` | `10s` / `30s` / `60s` / `120s` |
| `server.max-header-bytes` | `1048576`（1MB） |
| `server.max-body-bytes` | `4194304`（4MB） |
| `server.trusted-proxies` | `[]`（不信任任何代理） |
| `database.mysql` 连接池 | `max-open 100` / `max-idle 10` / `lifetime 1h` |
| `redis.pool-size` / `dial-timeout` / `read-timeout` / `write-timeout` | `10` / `5s` / `3s` / `3s` |
| `jwt.expire` | `24`（小时） |
| `jwt.token-id` | `uid` |
| `jwt.blacklist-prefix` | `jwt:blacklist:` |
| `jwt.blacklist-fail-closed` | `false` |
| `log.level` / `log.format` | `info` / `console` |

### 7.2 HTTP 契约

- 统一响应信封：`{"code": int, "message": string, "content": any?}`；
  **错误响应省略 `content`**，成功响应总是 HTTP 200。
- 错误响应的 HTTP 状态码由 `errno.Errno.HTTP` 决定（401/404/413/429/5xx 等），
  `code` 为业务码；5xx 不向客户端泄漏原始错误。
- 内置端点：`GET /healthz`（liveness，恒 200）、`GET /readyz`
  （MySQL/Redis 就绪检查，200/503，响应含 `checks`）；未知路由返回统一 404 信封。
- 请求头：认证使用 `Authorization: Bearer <token>`；`X-Request-ID` 会被校验
  （`[A-Za-z0-9._-]{1,128}`，非法值由服务端生成）并写入响应头与请求上下文。
- 访问日志字段（method/path/status/latency/client_ip/request_id）属于可观测性输出，
  **不是**机器契约，字段可能在 minor 中调整。

### 7.3 错误语义

- 业务代码通过返回 `*errno.Errno`（或包装它的 error）控制 HTTP 状态码、
  业务码与用户可读消息；`errno.Decode` 负责解包（支持 `errors.As`）。
- `Errno.HTTP` 属于 Stable API；`Errno.Message` 是面向用户的可读文案，可能被改进。
- **业务码 `Errno.Code` 的取值将在 `v1.0.0-rc.1` 之前统一**（当前多个错误共用
  `Code: 1`，属于已知缺陷，见 V1 就绪评估 P1 项）；统一完成后 `Code` 冻结，
  客户端应基于 `Code` 而非 `Message` 做分支判断。

---

## 8. 不属于 API 的部分

以下内容不构成兼容承诺，可在任意版本变更：

- 日志的输出格式、字段名、轮转文件命名与保留策略；
- 错误 `Message` 的具体文案（结构稳定，文案可能优化）；
- 生成的脚手架项目模板的逐行内容（但生成项目必须可构建、可启动）；
- 性能/内存特性、并发度、内部实现细节、未导出标识符；
- `docs/internal/**` 内部文档、`examples/**`、`scripts/**`、CI 配置、Makefile targets；
- `scaffold/internal/**`（Go `internal` 规则下外部不可导入）。

---

## 9. API 决策记录（ADR）

| # | 决策 | 理由 |
|---|------|------|
| D1 | `core/auth/jwt` 的包名保留 `token`，不改名为 `jwt` | 包名表达职责（签发/校验 token），目录表达机制；`*jwt.Token` 已出现在高级 API 签名中，重名会强迫调用方别名；改名对 0.x 用户是纯破坏性收益低。文档推荐按导入路径导入后使用 `token.` 前缀 |
| D2 | 全局状态保留（jwt/log/db 单例），不改为显式注入 | 0.x 用户与脚手架依赖该模型；改为依赖注入会全面破坏 API，且 1.0 前收益不足。通过文档明确“一个进程一个 Application”的受支持模型，并提供测试用 Reset |
| D3 | `ParseToken`/`IsBlacklisted` 继续暴露 `*jwt.Token` | 高级用法需要读取 claims；隐藏需新增包装类型。冻结为高级 API，并承诺 1.x 内保持 golang-jwt/v5 依赖 |
| D4 | 导出 `verflag.VersionValue` 类型 | 原类型未导出，导致 `VersionVar`/`Version` 对外不可调用（无效 API）；1.0 前修正使该包可用 |
| D5 | `log.NewLogger` 返回 `Logger` 接口 | 原返回未导出的 `*zapLogger`，调用方无法命名类型、无法用于接口签名；返回接口与 `Init` 语义一致 |
| D6 | `core/service` 标记为 Experimental | 该包是无引用、无实现的预留扩展点（历史模板遗留）；删除会破坏潜在用户，冻结为稳定 API 又会锁死设计，故明确不承诺兼容 |
| D7 | `errno.Code` 在 rc.1 前统一后冻结 | 当前 6 个错误共用 `Code: 1`，客户端无法区分；统一需在 API 冻结生效前完成 |
| D8 | 受支持模型为“一进程一个 Application” | jwt/log/db 均为进程级状态；支持多实例需要大规模重构，且与 1.0 目标无关 |

---

## 10. 冻结检查与变更流程

发布 `v1.0.0-rc.1` 前必须完成：

- [ ] `go doc -all ./...` 逐包审阅，确认所有导出标识符有意为之且文档齐全；
- [ ] `go vet` 与 `go test ./...` 全绿；`scripts/pre-release-check.sh` 0 FAIL；
- [ ] 本文档与 `docs/migration-0.x-to-1.0.md` 的签名/配置表与代码一致；
- [ ] CI 中启用 `go test -race` 与 `golint/vet` 门禁（见 P0-5）；
- [ ] 业务码 `Code` 统一完成（D7）。

1.x 期间修改 Stable API 的流程：

1. 提交 PR，并在描述中标注 `API CHANGE: additive` 或 `API CHANGE: breaking`；
2. breaking 变更在 1.x 中一律拒绝，除非属于安全修复并走例外评审；
3. additive 变更需同步更新本文档 §3 与 CHANGELOG 的 `Added` 段；
4. 弃用按 §4 执行，且在迁移指南补充替代写法。

> 对本文档的修改本身也遵循 minor/major 规则：新增澄清、补充示例属于 minor；
> 扩大或收缩承诺范围（如把 Stable 降级为 Experimental）属于 major。
