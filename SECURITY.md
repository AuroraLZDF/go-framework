# 安全策略

我们非常重视 go-framework 及其使用者的安全。请按照本文档描述的方式**私密**报告安全问题。

## 支持的版本

| 版本 | 安全更新 |
|---|---|
| `v0.2.x` | ✅ 支持 |
| `v1.0.x` | ✅ 支持（发布后） |
| `< v0.2.0` | ❌ 不再支持，请先升级 |

处于预发布状态的版本（`-alpha`、`-rc`）不承诺安全更新。

## 报告漏洞

**请勿通过公开 issue、讨论区或 PR 报告安全漏洞。**

推荐渠道：

1. 通过 GitHub 私密漏洞报告（Private Vulnerability Reporting）：
   <https://github.com/AuroraLZDF/go-framework/security/advisories/new>
2. 或发送邮件至 **auroralzdf@gmail.com**，标题以 `[SECURITY]` 开头。

报告内容建议包含：

- 受影响的版本/tag 或 commit
- 漏洞类型与影响面（例如认证绕过、信息泄漏、拒绝服务）
- 最小复现步骤或 PoC 代码
- 相关配置（请脱敏，删除真实 secret/password/DSN）
- 你希望的披露时间（如无特殊要求，按下方协调披露政策处理）

## 响应时间线

| 阶段 | 目标时间 |
|---|---|
| 确认收到报告 | 72 小时内 |
| 初步评估与定级（CVSS/严重性） | 7 天内 |
| 修复并发布补丁版本 | 高危 30 天内，其余与报告者协商 |

如未在 72 小时内收到确认，请再次发送邮件。

## 披露政策

- 采用**协调披露（coordinated disclosure）**：在补丁版本发布前，不公开漏洞细节。
- 修复发布后，我们会发布 GitHub Security Advisory，并在征得报告者同意后致谢。
- 如修复周期较长，最迟在 90 天内与报告者协商披露方案。
- 报告者请给我们合理的修复时间，避免在修复前公开 PoC。

## 框架安全模型

以下安全能力由框架提供并在 v0.1.3 起逐步收紧，使用方仍需正确配置：

- **JWT**：仅接受 `HS256`（使用 `WithValidMethods` 固定算法），要求 `exp`，
  密钥长度强制 ≥32 字节；`jti` 必填；吊销通过 Redis 黑名单实现
  （`jwt.blacklist-fail-closed` 可配置为失败即拒绝）。
- **认证中间件**：`middleware.Authn()` 不会自动挂载，需要业务显式注册到需要保护的路由。
- **反向代理信任**：默认**不信任**任何代理，配置 `server.trusted-proxies` 后才解析
  `X-Forwarded-For`；对 `0.0.0.0/0` 等宽泛 CIDR 会在启动时告警。
- **请求体限制**：`server.max-body-bytes` 默认 4MB，超限返回 413；
  未声明长度的 chunked 请求由 `http.MaxBytesReader` 截断。
- **错误信息**：未知错误统一返回通用 500 文案，不向客户端泄漏内部错误、
  SQL/DSN 等细节；5xx 会记录原始错误、路径与 request id。
- **HTTP 安全头**：`middleware.Secure` 提供 `X-Frame-Options`、`X-Content-Type-Options`、
  HSTS（仅 TLS）与 CSP；`NoCache` 提供禁止缓存响应头。
- **TLS**：最低版本 TLS 1.2；同时配置 `http-addr`/`https-addr` 时明文端口返回 308 跳转。
- **日志**：不会主动记录 Authorization 头、token 或密码。

## 不在范围内

- 使用方应用自身的业务逻辑漏洞（框架不是 WAF）。
- 使用方错误配置：如把 JWT secret 提交到仓库、将 `trusted-proxies` 设为 `0.0.0.0/0`、
  关闭认证中间件等。
- 第三方依赖的上游漏洞：请同时反馈给对应上游项目；我们会跟踪并升级依赖。
- 需要超大流量/资源的 DoS 攻击、社会工程、物理访问等。

## 安全港（Safe Harbor）

对于遵守本政策、出于善意进行安全研究和测试的行为，我们不会采取法律行动，
并会尽力在合理范围内提供协助。请勿访问、修改或删除他人数据，测试范围仅限本仓库代码。

感谢你帮助让 go-framework 更安全。
