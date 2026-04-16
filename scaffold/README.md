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

这会将 `scaffold` 安装到你的 `$GOPATH/bin` 目录，确保该目录在你的 PATH 中。

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
├── config.example.yaml  # 配置示例
├── go.mod              # Go Module 文件
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

```bash
cp config.example.yaml config.yaml
# 编辑 config.yaml，修改数据库、Redis 等配置
```

### 5. 运行项目

```bash
go run cmd/app/main.go
```

### 6. 测试 API

```bash
# 健康检查
curl http://localhost:8080/health

# 欢迎页面
curl http://localhost:8080/

# 用户列表
curl http://localhost:8080/api/v1/users
```

## 🔧 自定义模板

脚手架提供了以下模板文件（位于 `scaffold/templates/`）：

- `internal_biz_user.go.tpl` - 业务逻辑层示例
- `internal_controller_user.go.tpl` - 控制器示例
- `sql_users.sql.tpl` - 数据库表结构示例

你可以根据需要修改这些模板，或者添加自己的模板。

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
import "go-framework/core/response"

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

# 安装脚手架
make install

# 清理构建文件
make clean

# 测试脚手架
make test

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

### Q: 如何添加新的模板文件？

A: 
1. 在 `scaffold/templates/` 目录下创建模板文件
2. 在 `generator.go` 中添加相应的生成函数
3. 在 `NewProject` 函数中调用新的生成函数

### Q: 脚手架支持哪些 Go 版本？

A: 默认使用 Go 1.23+，你可以在生成的 `go.mod` 文件中修改版本要求。

## 📚 相关文档

- [Framework 快速开始](../docs/quick-start.md)
- [Framework 架构设计](../docs/architecture.md)
- [Framework README](../README.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进脚手架工具！

---

**Happy Coding! 🎉**
