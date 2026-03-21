# SkillFlow 架构文档集

这个目录保存 SkillFlow 基于 DDD 的架构参考文档。

SkillFlow 的目标后端架构是一个基于 DDD 的模块化单体：

- `cmd/skillflow/` 继续作为 Wails 桌面壳层、transport adapter、进程宿主和组合根
- 后端业务代码逐步迁移到 `core/` 下的各个 bounded context
- 每个上下文内部统一组织为 `app`、`domain`、`infra`
- 跨上下文写协调通过 `orchestration/`
- 跨上下文读视图通过 `readmodel/`

此前的 `docs/architecture.md` 将当前实现细节、仓库规则、运行时说明和领域概念混在同一个文件里。现在改为一组拆分文档，分别描述目标架构、运行时约束和迁移方案。

## 文档列表

- [总体架构](./overview_zh.md)
  - 高层架构风格、目标目录结构和真相归属规则
- [分层与依赖规则](./layers_zh.md)
  - transport adapter、`app`、`domain`、`infra`、`orchestration`、`readmodel`、`platform`、`shared` 的职责定义
- [限界上下文与领域模型](./contexts_zh.md)
  - bounded context 地图、聚合根、值对象、published language 和跨上下文身份规则
- [应用层用例](./use-cases_zh.md)
  - 各上下文的 command/query 归属、共享编排以及 read model 组合规则
- [运行时、仓库布局与存储](./runtime-and-storage_zh.md)
  - Wails 壳层约束、helper/UI 双进程、目标存储布局，以及 repository 与 gateway 的划分规则
- [迁移蓝图](./migration_zh.md)
  - 从当前代码结构迁移到目标 DDD 结构的映射关系和推荐顺序

## 不变约束

- 仓库根目录不能放 Go 源文件。
- `cmd/skillflow/*.go` 必须保持扁平，因为 Wails 绑定要求单一 `package main` 目录。
- SkillFlow 仍然是 Wails 桌面应用，不引入 REST 服务层。
- 由于 Wails 绑定限制，当前 transport entrypoint 保留在 `cmd/skillflow/`。
- `Skill` 和 `Prompt` 是并列的核心业务概念。
- `Settings` 是 UI 组合视图，不是独立 bounded context。

## 范围

这组文档只覆盖后端架构。用户可见行为仍然以 [`docs/features_zh.md`](../features_zh.md) 为准，落盘配置格式仍然以 [`docs/config_zh.md`](../config_zh.md) 为准。

## 当前状态

当前代码尚未完全迁移到这套架构。除非文中明确说明，这组文档描述的是后续重构要收敛到的目标结构。

截至 2026-03-21，已经有三个端到端 bounded context 落地到 `core/` 下：

- `core/skillcatalog/app`
- `core/skillcatalog/domain`
- `core/skillcatalog/infra`
- `core/promptcatalog/app`
- `core/promptcatalog/domain`
- `core/promptcatalog/infra`
- `core/agentintegration/app`
- `core/agentintegration/domain`
- `core/agentintegration/infra`

旧的 `core/skill`、`core/prompt` 和 `core/sync` 包已经移除。其余领域与横切模块仍需继续按同样方式迁移。

*最后更新：2026-03-21*
