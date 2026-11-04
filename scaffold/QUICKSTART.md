# 脚手架快速开始

## 5 分钟创建你的第一个 Framework 项目

### 步骤 1: 构建脚手架工具

```bash
cd framework
make build
```

> 也可以直接 `go run ./scaffold/cmd/main.go new my-app`，或 `make install` 后全局使用。

### 步骤 2: 创建新项目

```bash
./bin/scaffold new my-awesome-api
```

### 步骤 3: 进入项目目录并安装依赖

```bash
cd my-awesome-api
go mod tidy
```

> 生成的 `go.mod` 依赖 `github.com/AuroraLZDF/go-framework v0.1.3`。
> 如果该 tag 尚未发布，会提示 `unknown revision v0.1.3`；请先发布版本，
> 或在框架仓库内使用 `FRAMEWORK_LOCAL=1 make test-scaffold` 做本地验证。

### 步骤 4: 按需修改配置

脚手架已生成可直接运行的 `config.yaml`（`config.example.yaml` 保留为模板）：

```bash
vim config.yaml
```

- JWT secret 至少 32 字节（默认示例值已满足）
- 不需要 MySQL/Redis 时，将 `database.mysql.host`、`redis.addr` 留空即可

### 步骤 5: 运行应用

```bash
go run cmd/app/main.go
```

### 步骤 6: 测试 API

打开另一个终端窗口：

```bash
# 存活探针（框架内置）
curl http://localhost:8080/healthz

# 业务健康检查（生成的路由）
curl http://localhost:8080/health

# 欢迎页面
curl http://localhost:8080/

# 用户列表
curl http://localhost:8080/api/v1/users
```

`/health` 返回类似：

```json
{
  "code": 200,
  "message": "ok",
  "content": {
    "status": "healthy"
  }
}
```

## 🎉 恭喜！

你已经成功创建并运行了你的第一个 Framework 项目！

## 下一步

- 查看 [scaffold/README.md](README.md) 了解更多高级功能
- 阅读 [Framework 快速开始](../docs/quick-start.md) 学习如何使用框架
- 了解如何挂载 `middleware.Authn()` 保护需要登录的路由
- 开始编写你的业务逻辑代码

## 常见问题

### Q: 如何更改端口？

编辑 `config.yaml` 文件中的 `server.http-addr` 字段：

```yaml
server:
  http-addr: :3000  # 改为 3000 端口
```

### Q: 如何添加新的路由？

编辑 `cmd/app/main.go` 文件，在 `RegisterRoutes` 函数中添加：

```go
engine.GET("/my-route", func(c *gin.Context) {
    response.Success(c, gin.H{"message": "Hello"}, "ok")
})
```

### Q: 如何使用数据库？

1. 在 `config.yaml` 中配置 MySQL 连接
2. 在 `internal/model/` 中定义数据模型
3. 在 `internal/store/` 中实现数据访问层
4. 在 `internal/biz/` 中实现业务逻辑
5. 在 `internal/controller/` 中创建 HTTP 处理器

### Q: 脚手架版本信息？

```bash
./bin/scaffold version
```

---

**Happy Coding! 🚀**
