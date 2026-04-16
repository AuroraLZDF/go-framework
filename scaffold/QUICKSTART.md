# 脚手架快速开始

## 5分钟创建你的第一个 Framework 项目

### 步骤 1: 构建脚手架工具

```bash
cd framework
make build
```

### 步骤 2: 创建新项目

```bash
./bin/scaffold new my-awesome-api
```

### 步骤 3: 进入项目目录

```bash
cd my-awesome-api
```

### 步骤 4: 安装依赖

```bash
go mod tidy
```

### 步骤 5: 配置项目

```bash
# 复制配置示例
cp config.example.yaml config.yaml

# 编辑配置文件（可选）
vim config.yaml
```

### 步骤 6: 运行应用

```bash
go run cmd/app/main.go
```

### 步骤 7: 测试 API

打开另一个终端窗口：

```bash
# 健康检查
curl http://localhost:8080/health

# 欢迎页面
curl http://localhost:8080/

# 用户列表
curl http://localhost:8080/api/v1/users
```

你应该看到类似这样的响应：

```json
{
  "code": 0,
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
- 阅读 [Framework 文档](../docs/quick-start.md) 学习如何使用框架
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

---

**Happy Coding! 🚀**
