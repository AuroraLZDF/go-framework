# 架构设计说明

## 概述

Framework 采用经典的三层架构设计，结合依赖注入和接口抽象，实现了高内聚、低耦合的代码组织方式。

## 架构图

```
┌─────────────────────────────────────────┐
│          HTTP Request / Response         │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│         Controller Layer (控制器层)      │
│  - 接收 HTTP 请求                        │
│  - 参数验证                              │
│  - 调用 Biz 层                           │
│  - 返回统一响应                          │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│           Biz Layer (业务逻辑层)         │
│  - 业务流程编排                          │
│  - 业务规则验证                          │
│  - 事务管理                              │
│  - 调用 Store 层                         │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│          Store Layer (数据访问层)        │
│  - 数据库 CRUD 操作                      │
│  - 查询构建                              │
│  - 数据转换                              │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│       Framework Core (框架核心)          │
│  - Database (MySQL/Redis)               │
│  - Log (Zap)                            │
│  - Auth (JWT)                           │
│  - Middleware                           │
│  - Config                               │
└─────────────────────────────────────────┘
```

## 核心模块

### 1. Server (应用服务器)

**职责**: 
- 应用生命周期管理
- HTTP/HTTPS 服务器启动
- 中间件注册
- 优雅关闭

**关键代码**:
```go
type Application struct {
    config *config.Config
    engine *gin.Engine
    db     *gorm.DB
}

func NewApplication(cfg *config.Config) (*Application, error)
func (app *Application) RegisterRoutes(registerFunc RouteRegisterFunc)
func (app *Application) Run(ctx context.Context) error
```

### 2. Config (配置管理)

**职责**:
- 配置文件加载（YAML）
- 环境变量支持
- 配置验证

**配置结构**:
```
Config
├── App (应用配置)
├── Server (服务器配置)
├── Database (数据库配置)
│   └── MySQL
├── Redis (Redis 配置)
├── JWT (认证配置)
└── Log (日志配置)
```

### 3. Database (数据库)

**职责**:
- MySQL 连接池管理
- Redis 连接管理
- 单例模式确保唯一实例

**特性**:
- 连接池优化
- 自动重连
- SQL 日志记录

### 4. Auth (认证)

#### JWT Token 管理
- Token 签发（Sign）
- Token 验证（ParseRequest）
- Token 黑名单

#### 密码加密
- bcrypt 加密算法
- 安全的密码存储

### 5. Middleware (中间件)

**内置中间件**:
- `Authn()` - JWT 认证
- `Cors()` - CORS 跨域
- `RequestID()` - 请求 ID 追踪
- `Logger()` - 请求日志
- `Recovery()` - Panic 恢复

### 6. Response (响应封装)

**统一响应格式**:
```json
{
  "code": 0,
  "message": "ok",
  "content": {...}
}
```

**错误处理**:
```go
response.Success(c, data, "ok")
response.Error(c, errno.ErrTokenInvalid)
```

### 7. Log (日志系统)

**特性**:
- 结构化日志（Zap）
- 多输出（文件 + 控制台）
- 日志轮转（按日期）
- 请求 ID 追踪

**使用示例**:
```go
log.Info("User logged in", "user_id", 123)
log.Error("Database error", "err", err)
```

## 设计原则

### 1. 分层架构

每一层只依赖下一层，避免循环依赖：

```
Controller → Biz → Store → Framework Core
```

### 2. 依赖注入

通过构造函数注入依赖，便于测试：

```go
type UserController struct {
    userBiz biz.UserBiz
}

func NewUserController(userBiz biz.UserBiz) *UserController {
    return &UserController{userBiz: userBiz}
}
```

### 3. 接口抽象

定义接口而非具体实现，提高可扩展性：

```go
type SMSProvider interface {
    SendCode(phone, code string) error
}

// 业务代码依赖接口，不依赖具体实现
type UserBiz struct {
    smsProvider service.SMSProvider
}
```

### 4. 单一职责

每个模块只负责一个功能：
- `config` - 配置管理
- `database` - 数据库连接
- `auth` - 认证授权
- `log` - 日志记录

### 5. 开闭原则

对扩展开放，对修改关闭：
- 通过接口扩展新功能
- 不修改现有代码

## 数据流示例

### 用户登录流程

```
1. HTTP POST /auth/login
         ↓
2. Controller: 接收请求，验证参数
         ↓
3. Biz: 验证手机号和验证码
         ↓
4. Store: 查询用户信息
         ↓
5. Biz: 生成 JWT Token
         ↓
6. Controller: 返回 Token
         ↓
7. HTTP Response: {access_token: "..."}
```

### 代码示例

```go
// Controller
func (ctrl *UserController) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, errno.ErrValidationFailed)
        return
    }
    
    token, err := ctrl.userBiz.Login(c, req.Phone, req.Code)
    if err != nil {
        response.Error(c, err)
        return
    }
    
    response.Success(c, token, "ok")
}

// Biz
func (b *userBiz) Login(ctx context.Context, phone, code string) (*response.TokenResponse, error) {
    // 1. 验证验证码
    if !b.phoneCodeBiz.VerifyCode(phone, code) {
        return nil, errno.ErrInvalidCode
    }
    
    // 2. 查询用户
    user, err := b.ds.Users().GetByPhone(ctx, phone)
    if err != nil {
        return nil, err
    }
    
    // 3. 生成 Token
    token, err := jwt.Sign(strconv.FormatInt(user.ID, 10))
    if err != nil {
        return nil, err
    }
    
    return token, nil
}

// Store
func (s *userStore) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
    var user model.User
    if err := s.db.Where("phone = ?", phone).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}
```

## 扩展点

### 1. 自定义中间件

```go
func MyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 前置处理
        c.Next()
        // 后置处理
    }
}

// 注册中间件
app.GetEngine().Use(MyMiddleware())
```

### 2. 自定义服务实现

```go
// 实现短信服务接口
type AliyunSMS struct {
    client *aliyun.Client
}

func (s *AliyunSMS) SendCode(phone, code string) error {
    // 调用阿里云 API
    return s.client.SendVerifyCode(phone, code)
}

// 注入到 Biz 层
userBiz := biz.NewUserBiz(store, &AliyunSMS{})
```

### 3. 自定义配置项

```yaml
# config.yaml
extra:
  my-custom-config: value
```

```go
// 读取自定义配置
customConfig := cfg.Extra["my-custom-config"]
```

## 性能优化

### 1. 数据库连接池

```yaml
database:
  mysql:
    max-idle-connections: 100
    max-open-connections: 20
    max-connection-life-time: 30s
```

### 2. Redis 连接池

```go
redisOpts := &db.RedisOptions{
    PoolSize:     50,
    MinIdleConns: 10,
}
```

### 3. 日志异步写入

Zap 默认使用异步写入，不会阻塞主流程。

## 安全考虑

### 1. JWT Secret

⚠️ **重要**: 生产环境必须使用强随机密钥！

```bash
# 生成随机密钥
openssl rand -base64 32
```

### 2. 密码加密

使用 bcrypt 算法，自动加盐：

```go
import "go-framework/core/auth/password"

hashedPassword, _ := password.Encrypt(rawPassword)
password.Match(hashedPassword, rawPassword)
```

### 3. SQL 注入防护

使用 GORM 的参数化查询：

```go
// ✅ 安全
db.Where("name = ?", userInput).Find(&users)

// ❌ 危险
db.Where(fmt.Sprintf("name = '%s'", userInput)).Find(&users)
```

## 测试策略

### 1. 单元测试

测试单个函数或方法：

```go
func TestUserBiz_Login(t *testing.T) {
    // 准备测试数据
    // 调用被测函数
    // 验证结果
}
```

### 2. 集成测试

测试多个组件协作：

```go
func TestLoginFlow(t *testing.T) {
    // 启动测试服务器
    // 发送 HTTP 请求
    // 验证响应
}
```

### 3. Mock 外部依赖

```go
// 创建 Mock SMS 服务
mockSMS := &MockSMSProvider{}
mockSMS.On("SendCode", "13800138000", "1234").Return(nil)

// 注入 Mock
userBiz := biz.NewUserBiz(store, mockSMS)
```

## 总结

Framework 的架构设计遵循以下原则：

1. **清晰的分层** - 职责明确，易于维护
2. **依赖注入** - 松耦合，易测试
3. **接口抽象** - 可扩展，易替换
4. **约定优于配置** - 减少样板代码
5. **开箱即用** - 提供常用组件

这种设计使得 Framework 既适合小型项目快速开发，也能支撑大型项目的复杂需求。
