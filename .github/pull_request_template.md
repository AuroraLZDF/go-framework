## 摘要

<!-- 简述这个 PR 做了什么、为什么。 -->

## 变更类型

- [ ] 缺陷修复
- [ ] 新功能 / 新配置项 / 新导出 API
- [ ] 文档
- [ ] 重构 / 性能优化
- [ ] CI / 构建 / 发布脚本
- [ ] 破坏性变更（必须在描述中说明迁移方式，并更新 `docs/migration-0.x-to-1.0.md` 与 CHANGELOG）

## 关联 issue

<!-- 例如：Closes #123 -->

## 检查清单

- [ ] `gofmt -l .` 无输出
- [ ] `go build ./...` 通过
- [ ] `go vet ./...` 通过
- [ ] `go test -count=1 ./...` 通过
- [ ] 新增/变更行为有对应测试（表驱动优先）
- [ ] 用户可见变更已更新 README / `docs/`，并在 CHANGELOG `## [Unreleased]` 添加条目
- [ ] 涉及公共 API 时已更新 `docs/api-stability.md`（破坏性变更还需更新迁移指南）
- [ ] 未提交密钥、个人配置、生成物（bin/logs/config.yaml 等）
- [ ] 修改 `scaffold/` 时已运行 `FRAMEWORK_LOCAL=1 make test-scaffold` 或
      `FRAMEWORK_LOCAL=1 ./scripts/verify-scaffold.sh`

## 备注

<!-- 需要评审者特别关注的点、设计取舍、后续工作。 -->
