# 贡献指南

感谢你愿意为 go-framework 做贡献！无论是报告问题、改进文档、补充测试还是提交代码，都非常欢迎。

本文档说明参与贡献的流程、规范与要求。提交贡献即表示你同意遵守本项目的
[行为准则](CODE_OF_CONDUCT.md)，并同意你的贡献以 [MIT 许可证](LICENSE) 授权。

## 目录

- [报告问题](#报告问题)
- [开发环境](#开发环境)
- [本地开发命令](#本地开发命令)
- [代码结构](#代码结构)
- [提交规范](#提交规范)
- [分支与 Pull Request](#分支与-pull-request)
- [API 稳定性要求](#api-稳定性要求)
- [测试要求](#测试要求)
- [文档要求](#文档要求)
- [发布流程（维护者）](#发布流程维护者)

## 报告问题

- 提交 issue 前请先搜索是否已有相同问题。
- 请使用 `.github/ISSUE_TEMPLATE/` 中的模板，并尽量提供**最小复现**、版本/commit、
  Go 版本、操作系统、相关配置（请脱敏，删除 secret/password）。
- **安全漏洞请勿公开报告**，请按照 [SECURITY.md](SECURITY.md) 中的私密渠道提交。
- 使用类问题（如何配置、如何接入 MySQL 等）请先阅读
  [docs/quick-start.md](docs/quick-start.md) 与 [docs/CONFIG_USAGE.md](docs/CONFIG_USAGE.md)。

## 开发环境

- Go **1.26+**（`go.mod` 要求 `go 1.26.0`；CI 使用当前 1.26.x 最新补丁版本）
- Git
- `make`（可选，用于快捷命令）
- MySQL 8 / Redis 7 仅在需要时配置；框架在未配置 `database.host` / `redis.addr`
  时会跳过后端初始化，因此本地单元测试不强制依赖它们。

```bash
git clone https://github.com/AuroraLZDF/go-framework.git
cd go-framework
go build ./...
```

## 本地开发命令

提交前请在本地全部执行通过：

```bash
gofmt -l .                 # 必须无输出；需要格式化时执行 gofmt -w .
go build ./...
go vet ./...
go test -count=1 ./...
```

其他常用命令：

```bash
make lint                          # golangci-lint（需本地安装，CI 使用 v2.6.2）
make sec                           # govulncheck 依赖漏洞扫描（需本地安装 v1.1.4）
make test-race                     # go test -race
make cover                         # 覆盖率报告（coverage.out 与总计）
make check                         # lint + test + test-race
make test                          # go test ./...
FRAMEWORK_LOCAL=1 make test-scaffold   # 用本地仓库生成脚手架项目并构建/启动验证
FRAMEWORK_LOCAL=1 ./scripts/verify-scaffold.sh  # 完整脚手架 E2E（推荐）
make verify-scaffold               # 需已发布对应 tag，从模块代理拉取
```

CI 使用 Go 1.26.x，包含以下门禁：build / vet / gofmt / shell 语法检查 / 单元测试、
golangci-lint（配置见 `.golangci.yml`）、govulncheck 依赖漏洞扫描 + `go mod verify`、
`go test -race`、覆盖率检查（当前阈值 60%，P0-6 之后提升到 75%）以及脚手架 E2E。
如果本地 `GOCACHE` 位于只读目录（例如受沙箱限制），可临时指定可写缓存：

```bash
export GOCACHE=$(mktemp -d)
```

## 代码结构

```text
core/           框架核心库（对外 API 所在）
  auth/         jwt / password
  config/       配置加载与校验
  database/db/  MySQL / Redis 单例
  errno/        统一错误码
  log/          zap 日志、轮转与访问日志
  middleware/   认证、CORS、RequestID、请求体限制等中间件
  response/     统一响应封装
  server/       Application、路由注册、HTTP/HTTPS 监听与优雅退出
  util/         工具函数（cache/time/tree/conversion/file/ip 等）
  validator/    校验器与 GORM 查询构造
  version/      版本信息与 verflag
scaffold/       项目脚手架 CLI 与生成器
examples/       可运行的示例
docs/           使用与设计文档
scripts/        发布检查与脚手架验证脚本
```

## 提交规范

采用 [Conventional Commits](https://www.conventionalcommits.org/)：

```text
<type>(<scope>): <简短描述>

<可选正文：动机、方案、breaking change 说明>

<可选 footer：Refs #123 / BREAKING CHANGE: ...>
```

常用 type：

| type | 用途 |
|---|---|
| `feat` | 新功能/新配置项/新导出 API |
| `fix` | 缺陷修复 |
| `docs` | 仅文档 |
| `refactor` | 不改变行为的重构 |
| `test` | 测试相关 |
| `perf` | 性能优化 |
| `ci` | CI/构建脚本 |
| `chore` | 依赖、版权头等杂项 |

要求：

- 一个提交只做一件逻辑上的事；不要在同一提交中混入格式化与功能改动。
- 提交前必须通过 `gofmt`、`go build`、`go vet`、`go test`。
- 不要提交密钥、个人配置、生成物（`bin/`、`logs/`、`config.yaml` 等已在 `.gitignore` 中忽略）。
- 破坏性变更必须在正文中以 `BREAKING CHANGE:` 明确说明。

## 分支与 Pull Request

1. Fork 本仓库并从 `main` 创建分支，命名建议 `feat/xxx`、`fix/xxx`、`docs/xxx`。
2. 保持 PR 聚焦且尽量小；一个 PR 对应一个主题。
3. 按 [PR 模板](.github/pull_request_template.md) 填写摘要、变更类型与检查清单。
4. 确保 CI 全部通过；公共 API 变更需由 CODEOWNERS 审核。
5. 根据评审意见修改后 rebase 到最新 `main`，保持提交历史清晰。
6. 合并方式由维护者决定（默认 squash merge）。

## API 稳定性要求

- 提交前请阅读 [docs/api-stability.md](docs/api-stability.md)：其中列出了已冻结的
  公共 API、稳定性分级（Stable/Experimental/Internal）、弃用政策与 semver 规则。
- 新增导出符号属于公共 API 扩张：需要补充测试与文档，并在 `docs/api-stability.md` 中登记。
- 任何破坏性变更只允许在 major 版本中进行，并必须同步更新
  [docs/migration-0.x-to-1.0.md](docs/migration-0.x-to-1.0.md) 与 CHANGELOG。
- `internal`/`examples`/`docs` 中的实现不是公共 API，可自由重构。

## 测试要求

- 行为变更、新增配置项、新增导出 API 必须附带测试；优先使用表驱动测试。
- 单元测试不得依赖网络、真实 MySQL/Redis 或固定端口；需要外部依赖时使用 `testing.Short()`
  跳过或在测试内 mock/替换。
- 涉及 HTTP 行为（中间件、状态码、统一响应）时请直接断言状态码与响应体。
- 修改 `scaffold/` 时需要运行 `FRAMEWORK_LOCAL=1 make test-scaffold` 或
  `FRAMEWORK_LOCAL=1 ./scripts/verify-scaffold.sh`。

## 文档要求

- 用户可见的行为/配置变更需要同步更新 README、`docs/` 下相关文档与示例配置。
- 在 CHANGELOG 的 `## [Unreleased]` 下按 Added/Changed/Fixed/Security/Deprecated
  分类添加条目；合并发布时由维护者移动到具体版本。
- 文档中的相对链接必须有效；代码示例必须可编译（至少通过 `go vet` 级别的检查）。

## 发布流程（维护者）

1. 确认 `main` 工作区干净、CI 全绿。
2. 更新 `core/version/version.go` 的 `GitVersion`/`BuildDate`、CHANGELOG、`RELEASE_vX.Y.Z.md`，
   以及 README badge、scaffold 依赖版本、脚本中的版本引用。
3. 执行 `bash scripts/pre-release-check.sh`（要求 0 FAIL）与
   `FRAMEWORK_LOCAL=1 ./scripts/verify-scaffold.sh`。
4. 提交后创建 annotated tag 并推送：

   ```bash
   git tag -a vX.Y.Z -m "go-framework vX.Y.Z"
   git push origin main
   git push origin refs/tags/vX.Y.Z
   ```

5. 在 GitHub Releases 发布 `RELEASE_vX.Y.Z.md` 内容。
6. tag 推送后确认 CI 通过，并用默认模式运行 `make verify-scaffold` 验证已发布模块可用。

---

再次感谢你的贡献！如有疑问，欢迎在 issue 中讨论。
