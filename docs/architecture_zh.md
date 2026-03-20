# SkillFlow 架构文档

> **中文** | [English](architecture.md)

这是 SkillFlow 架构文档的入口页。原先的单文件架构说明已经被替换为位于 [`docs/architecture/`](./architecture/README_zh.md) 下的一组基于 DDD 的拆分文档。

SkillFlow 的目标后端架构是一个基于 DDD 的模块化单体：

- `cmd/skillflow/` 继续作为 Wails 桌面壳层、进程宿主和组合根
- 后端业务代码逐步迁移到 `core/` 下的各个 bounded context
- 每个上下文内部统一组织为 `entrypoint`、`app`、`domain`、`infra`
- 跨上下文写流程通过 `workflow/`
- 跨上下文读视图通过 `readmodel/`

后续后端重构应以这组文档为准。

## 阅读顺序

- [文档总览](./architecture/README_zh.md)
- [总体架构](./architecture/overview_zh.md)
- [分层与依赖规则](./architecture/layers_zh.md)
- [限界上下文与领域模型](./architecture/contexts_zh.md)
- [应用层用例](./architecture/use-cases_zh.md)
- [运行时、仓库布局与存储](./architecture/runtime-and-storage_zh.md)
- [迁移蓝图](./architecture/migration_zh.md)

## 不变约束

- 仓库根目录不能放 Go 源文件。
- `cmd/skillflow/*.go` 必须保持扁平，因为 Wails 绑定要求单一 `package main` 目录。
- SkillFlow 仍然是 Wails 桌面应用，不引入 REST 服务层。
- `Skill` 和 `Prompt` 是并列的核心业务概念。
- `Settings` 是 UI 组合视图，不是独立 bounded context。

*最后更新：2026-03-20*
