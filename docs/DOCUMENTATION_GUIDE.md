# Framework 文档组织指南

本文档说明了 Framework 项目中各类文档的组织结构和提交策略。

## 📂 文档目录结构

```
framework/
├── README.md                      ✅ 提交 - 项目主文档
├── CHANGELOG.md                   ✅ 提交 - 版本变更记录
├── LICENSE                        ✅ 提交 - 开源许可证
│
├── docs/                          # 公开文档目录
│   ├── quick-start.md            ✅ 提交 - 快速开始指南
│   ├── architecture.md           ✅ 提交 - 架构设计文档
│   └── internal/                 ❌ 不提交 - 内部开发文档
│       ├── README.md             (说明文件，也被忽略)
│       ├── COMPILE_STATUS.md     (内部记录)
│       ├── PROJECT_SUMMARY.md    (内部总结)
│       ├── RELEASE_*.md          (发布相关内部文档)
│       ├── TODO.md               (待办事项)
│       └── GITIGNORE_GUIDE.md    (参考指南)
│
└── scaffold/                      # 脚手架工具
    ├── README.md                  ✅ 提交 - 使用指南
    ├── QUICKSTART.md              ✅ 提交 - 快速开始
    └── DEVELOPMENT.md             ✅ 提交 - 开发文档
```

## ✅ 应该提交到 GitHub 的文档

### 1. 项目根目录

| 文件 | 用途 | 重要性 |
|------|------|--------|
| README.md | 项目介绍和快速开始 | ⭐⭐⭐⭐⭐ 必须 |
| CHANGELOG.md | 版本变更历史 | ⭐⭐⭐⭐⭐ 必须 |
| LICENSE | 开源许可证 | ⭐⭐⭐⭐⭐ 必须 |

### 2. docs/ 目录（公开文档）

| 文件 | 用途 | 目标读者 |
|------|------|----------|
| quick-start.md | 5分钟上手教程 | 新用户 |
| architecture.md | 架构设计和最佳实践 | 开发者、架构师 |

### 3. scaffold/ 目录

| 文件 | 用途 | 目标读者 |
|------|------|----------|
| README.md | 脚手架完整使用指南 | 使用脚手架的开发者 |
| QUICKSTART.md | 脚手架快速开始 | 新用户 |
| DEVELOPMENT.md | 脚手架开发文档 | 贡献者 |

## ❌ 不应该提交的文档

### docs/internal/ 目录

这些是**内部开发文档**，包含：

- 开发过程中的临时记录
- 内部决策和评估报告
- 待办事项清单
- 参考指南

**原因：**
1. 面向内部团队，不是最终用户
2. 内容可能过时或不准确
3. 避免仓库臃肿
4. 保护内部决策过程

**替代方案：**
- TODO → GitHub Issues
- 项目总结 → GitHub Wiki
- 发布检查 → GitHub Projects
- 技术讨论 → GitHub Discussions

## 🎯 文档分类原则

### 公开文档的特征

✅ **应该公开如果：**
- 面向最终用户或外部开发者
- 内容是稳定的、经过验证的
- 对项目使用有直接帮助
- 符合开源项目的标准文档

❌ **不应该公开如果：**
- 内部开发过程的记录
- 临时的、可能过时的信息
- 包含敏感的开发细节
- 只是个人笔记或草稿

## 📋 文档维护建议

### 对于公开文档

1. **保持更新** - 代码变化时同步更新文档
2. **质量保证** - 拼写检查、格式统一
3. **用户导向** - 从用户角度编写
4. **示例完整** - 提供可运行的示例代码

### 对于内部文档

1. **定期清理** - 删除过时的文档
2. **迁移重要内容** - 将有价值的内容迁移到 Issues/Wiki
3. **保持简洁** - 只保留必要的参考信息
4. **明确标记** - 在文件名或内容中标明"内部"

## 🔄 文档生命周期

```
创建 → 评审 → 公开/归档 → 维护 → 废弃
  ↓
内部文档 (docs/internal/)
  ↓ 如果有价值
公开文档 (docs/, README.md)
  ↓ 如果过时
归档或删除
```

## 💡 最佳实践

### 1. README.md 应该包含

- 项目简介
- 主要特性
- 快速开始
- 文档链接
- 贡献指南
- 许可证信息

### 2. CHANGELOG.md 格式

遵循 [Keep a Changelog](https://keepachangelog.com/) 规范：

```markdown
## [版本号] - 日期

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
```

### 3. 技术文档结构

- 概述
- 架构设计
- 核心模块
- API 参考
- 最佳实践
- 常见问题

## 🚀 发布前的文档检查清单

发布新版本前，确保：

- [ ] README.md 已更新（版本号、特性）
- [ ] CHANGELOG.md 已添加新版本记录
- [ ] 所有公开文档与代码同步
- [ ] 示例代码可以运行
- [ ] 没有内部文档被意外提交
- [ ] 文档链接都有效

## 📊 当前项目文档状态

### ✅ 已完成

- [x] README.md - 完整的项目介绍
- [x] CHANGELOG.md - v0.1.0-alpha 变更记录
- [x] LICENSE - MIT 许可证
- [x] docs/quick-start.md - 快速开始
- [x] docs/architecture.md - 架构设计
- [x] scaffold/README.md - 脚手架文档
- [x] .gitignore - 正确配置

### 📝 建议改进

- [ ] 添加 CONTRIBUTING.md - 贡献指南
- [ ] 添加 CODE_OF_CONDUCT.md - 行为准则
- [ ] 添加 SECURITY.md - 安全政策
- [ ] 完善 API 参考文档
- [ ] 添加更多示例

---

**最后更新**: 2026-04-16  
**版本**: v0.1.0-alpha
