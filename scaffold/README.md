# 脚手架工具使用指南

Framework 提供了一个强大的脚手架工具，帮助你快速创建基于 Framework 的新项目。

## 📦 安装

### 方法一：从源码构建

```bash
cd framework
make build
```

构建完成后，可执行文件位于 `bin/scaffold`。

### 方法二：全局安装

```bash
cd framework
make install
```

这会将 `scaffold` 安装到 `GOBIN`（未设置时为 `$GOPATH/bin`），确保该目录在你的 PATH 中。

### 方法三：直接运行

```bash
go run ./scaffold/cmd/main.go new my-app
```

## 🚀 快速开始

### 创建新项目

```bash
scaffold new my-app
```

这将创建一个名为 `my-app` 的新项目，包含完整的目录结构和配置文件。

### 项目结构

```
my-app/
├── cmd/app/              # 应用入口
│   └── main.go          # 主程序文件
├── internal/            # 内部包
│   ├── biz/            # 业务逻辑层
│   ├── controller/     # HTTP 控制器
│   ├── store/          # 数据访问层
│   ├── model/          # 数据模型
│   └── service/        # 外部服务
├── configs/             # 配置文件目录
├── sql/                 # 数据库脚本
├── assets/              # 静态资源
├── logs/                # 日志目录
├── config.yaml          # 可直接运行的配置（已生成）
├── config.example.yaml  # 配置模板
├── go.mod              # Go Module 文件（依赖 framework v0.2.0）
└── README.md           # 项目说明
```

## 📝 使用步骤

### 1. 创建项目

```bash
scaffold new my-api-service
```

### 2. 进入项目目录

```bash
cd my-api-service
```

### 3. 安装依赖

```bash
go mod tidy
```

### 4. 配置项目

生成项目时已创建可直接运行的 `config.yaml`，按需修改（例如 JWT secret 至少 32 字节）：

```bash
vim config.yaml
```

配置默认包含 `max-body-bytes: 4194304`（4MB，超限 413）与 `trusted-proxies: []`
（不信任任何代理；部署在反向代理后按需填入 IP/CIDR）。

### 5. 运行项目

```bash
go run cmd/app/main.go
```

### 6. 测试 API

```bash
# 存活探针（框架内置，仅进程存活）
curl http://localhost:8080/healthz

# 就绪探针（框架内置，检查已配置的 MySQL/Redis，未配置标记 skipped）
curl http://localhost:8080/readyz

# 业务健康检查（生成的路由）
curl http://localhost:8080/health

# 欢迎页面
curl http://localhost:8080/

# 用户列表
curl http://localhost:8080/api/v1/users
```

## 🔧 自定义生成内容

脚手架没有独立的模板文件，所有内容由 `scaffold/internal/generator/generator.go`
中的字符串模板生成。修改该文件中的生成函数即可定制项目结构、示例代码与配置；
修改后运行 `make test-scaffold` 做端到端验证。

## 💡 最佳实践

### 1. 分层架构

遵循 Controller → Biz → Store 的分层架构：

```go
// Controller 层 - 处理 HTTP 请求
type UserController struct {
    userBiz biz.UserBiz
}

// Biz 层 - 业务逻辑
type UserBiz interface {
    GetUserByID(ctx context.Context, id int64) (interface{}, error)
}

// Store 层 - 数据访问
type UserStore interface {
    FindByID(ctx context.Context, id int64) (*model.User, error)
}
```

### 2. 依赖注入

使用构造函数注入依赖：

```go
func NewUserController(userBiz biz.UserBiz) *UserController {
    return &UserController{userBiz: userBiz}
}
```

### 3. 错误处理

使用框架的统一响应：

```go
import "github.com/AuroraLZDF/go-framework/core/response"

func (ctrl *UserController) GetUser(c *gin.Context) {
    user, err := ctrl.userBiz.GetUserByID(c.Request.Context(), id)
    if err != nil {
        response.Error(c, err)
        return
    }
    response.Success(c, user, "ok")
}
```

## 🛠️ Makefile 命令

项目根目录提供了 Makefile，简化常用操作：

```bash
# 构建脚手架
make build

# 安装脚手架（GOBIN / GOPATH/bin）
make install

# 清理构建文件
make clean

# 运行单元测试
make test

# 本地模式端到端验证脚手架（FRAMEWORK_LOCAL=1）
make test-scaffold

# 发布模式验证（要求 v0.2.0 tag 存在）
make verify-scaffold

# 查看帮助
make help
```

## ❓ 常见问题

### Q: 如何修改生成的项目名称？

A: 在 `scaffold new` 命令中指定你想要的项目名称即可：

```bash
scaffold new my-awesome-project
```

### Q: 可以自定义生成的代码吗？

A: 可以！修改 `scaffold/internal/generator/generator.go` 中的模板内容，或者添加新的生成逻辑。

### Q: 脚手架如何输出版本？

A: `scaffold version` 或 `scaffold --version` 都会打印 GitVersion/BuildDate
等构建信息；Makefile 通过 `-ldflags` 注入这些值。

### Q: 脚手架支持哪些 Go 版本？

A: 生成的 `go.mod` 使用 Go 1.26.0，并依赖 `github.com/AuroraLZDF/go-framework v0.2.0`。
版本 v0.2.0 需要先打 tag 并推送，否则 `go mod tidy` 会提示 unknown revision。

## 📚 相关文档

- [Framework 快速开始](../docs/quick-start.md)
- [Framework 架构设计](../docs/architecture.md)
- [Framework README](../README.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进脚手架工具！

---

**Happy Coding! 🎉**
