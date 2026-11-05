# Release v0.2.0

**发布日期**: 2026-11-05

v0.2.0 聚焦**服务器加固与生产默认值**：新增请求体上限、就绪探针、可信代理配置、
HTTP→HTTPS 强制重定向与 `scaffold --version`，并修复若干与监听、Redis 地址、
IPv6 重定向相关的问题。本版本包含**默认行为变化**，升级前请阅读「行为变化」。

## ✨ 亮点

- **请求体上限**：`server.max-body-bytes`（默认 4MB），超限返回 `HTTP 413` 统一响应；
  chunked 请求由 `http.MaxBytesReader` 截断，业务可用 `middleware.IsRequestTooLarge(err)`
  映射为 413。
- **就绪探针**：`GET /readyz` 检查已配置的 MySQL/Redis（各自独立 2s Ping），
  失败返回 503；`GET /healthz` 明确为 liveness。
- **可信代理**：`server.trusted-proxies`（IP/CIDR，环境变量逗号分隔），
  默认空 = 不信任任何代理，避免伪造 `X-Forwarded-For` 影响客户端 IP/审计。
- **双监听与强制 HTTPS**：同时配置 `http-addr` 与 `https-addr` 时，HTTP 统一
  `308 Permanent Redirect` 到 HTTPS（保留 path/query、正确处理 IPv6 与端口、
  `Cache-Control: no-store`）。
- **配置校验**：`redis.addr` 支持 host / host:port / IPv6，并拒绝 `redis://`、
  多余冒号、非法端口；`0.0.0.0/0` 等全地址 CIDR 启动打印 WARNING。
- **脚手架**：`scaffold --version`；生成配置包含新键与双监听说明。

## 💥 行为变化 / 升级注意

1. **仅配置 `server.https-addr` 时不再隐式监听 `:8080`**。如需同时提供 HTTP，
   请显式配置 `server.http-addr`（将返回 308 重定向）。
2. **默认不信任任何代理**：`c.ClientIP()` 与访问日志 `client_ip` 默认取
   `RemoteAddr`。部署在 Nginx/SLB/Ingress 后请配置 `server.trusted-proxies`，
   否则看到的是代理 IP。
3. **默认请求体上限 4MB**：超过返回 413；需要更大上传时调整
   `server.max-body-bytes`（负值启动报错，`0` 使用默认值）。
4. **chunked 超限**：未声明 `Content-Length` 的请求由 `MaxBytesReader` 截断，
   框架不会自动改写业务绑定错误；建议业务使用
   `middleware.IsRequestTooLarge(err)` 返回 413。
5. 脚手架生成项目依赖 `github.com/AuroraLZDF/go-framework v0.2.0`，
   需先发布本 tag；本地开发可用 `FRAMEWORK_LOCAL=1 make test-scaffold`。

## 🔧 配置新增

```yaml
server:
  # 默认空 = 不信任任何代理；IP 或 CIDR 列表，也可用
  # FRAMEWORK_SERVER_TRUSTED_PROXIES="127.0.0.1, 10.0.0.0/8"
  trusted-proxies: []
  # 请求体上限，默认 4MB；负值启动报错
  max-body-bytes: 4194304
```

## 🧪 验证

```bash
gofmt -l .
go build ./...
go vet ./...
go test -count=1 ./...
bash -n scripts/pre-release-check.sh scripts/verify-scaffold.sh
FRAMEWORK_LOCAL=1 make test-scaffold
```

## ⚠️ 已知限制

- HTTP→HTTPS 重定向沿用请求 Host，没有独立的对外域名白名单配置；
  请由边界代理清洗 `Host`（响应已设置 `no-store`）。
- 黑名单默认 fail-open，安全敏感场景请开启 `jwt.blacklist-fail-closed`。
- 暂无 metrics/tracing 内置支持（规划中）。

## 🔗 相关链接

- [CHANGELOG](CHANGELOG.md)
- [快速开始](docs/quick-start.md)
- [架构设计](docs/architecture.md)
- [配置使用](docs/CONFIG_USAGE.md)
- [GitHub Releases](https://github.com/AuroraLZDF/go-framework/releases)
