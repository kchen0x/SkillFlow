# 迁移蓝图

## 目标

在不中断产品迭代的前提下，把后端从当前按技术能力切分的结构迁移到目标的 DDD 模块化单体。

迁移应采用渐进式方式。目标架构可以严格，但实施过程不应采用高风险的一次性重写。

## 当前进度检查点

截至 2026-03-21：

- `skillcatalog` 已迁出到 `core/skillcatalog/app`、`domain`、`infra`
- `promptcatalog` 已迁出到 `core/promptcatalog/app`、`domain`、`infra`
- `agentintegration` 已迁出到 `core/agentintegration/app`、`domain`、`infra`
- `skillsource` 已迁出到 `core/skillsource/app`、`domain`、`infra`
- 纯 Git 基础能力已经迁到 `core/platform/git`
- `cmd/skillflow` 和相关后端包已经不再 import `core/skill`
- `cmd/skillflow` 已经不再 import `core/prompt`
- `cmd/skillflow` 和相关后端包已经不再 import `core/sync`
- `cmd/skillflow` 和相关后端包已经不再 import `core/git`
- 旧的 `core/skill`、`core/prompt`、`core/sync` 和 `core/git` 包已经移除

这意味着支撑域抽取已经开始，后续迁移可以把 `skillcatalog`、`promptcatalog`、`agentintegration` 和 `skillsource` 作为参考模式。

## 当前到目标的映射

| 当前包 / 区域 | 目标归属 |
|--------------|---------|
| `core/skill` | `core/skillcatalog/domain` + `core/skillcatalog/infra/repository` |
| `core/prompt` | `core/promptcatalog/domain` + `core/promptcatalog/infra/repository` |
| `core/sync` | `core/agentintegration/app/port/gateway` + `core/agentintegration/infra/gateway` |
| `core/git` 中与来源跟踪相关的部分 | `core/skillsource/domain` 中的 `StarRepo` 与 `SkillSource` 模型 + `core/skillsource/infra` |
| `core/git` 中的 Git 基础能力 | `core/platform/git` |
| `core/backup` 中的 provider 与 snapshot 逻辑 | `core/backup/domain` + `core/backup/infra` |
| `core/config` | `core/platform/settingsstore` + 各上下文拥有的配置 namespace |
| `core/applog` | `core/platform/logging` |
| `core/notify` | `core/platform/eventbus` |
| `core/pathutil` | `core/platform/pathutil` |
| `core/skillkey` | `core/shared/logicalkey` |
| `core/upgrade` | `core/platform/upgrade` |
| `core/viewstate` | `core/readmodel` 或上下文本地 `infra/projection` |
| `core/registry` | `cmd/skillflow` 中的壳层组合逻辑 |
| `core/update` 中的应用更新部分 | `core/platform/update` + `cmd/skillflow` 中的壳层适配 |
| `cmd/skillflow/app*.go` 中的业务方法 | `cmd/skillflow` 中的 Wails transport adapter，转调各上下文的 `app`、`orchestration` 与 `readmodel` |
| `cmd/skillflow/process_*.go`、`tray_*.go`、`window_*.go` | 继续保留在 `cmd/skillflow` |

## 迁移阶段

### 阶段 0：冻结架构方向

- 以这组文档作为目标架构
- 停止继续把新的可复用业务逻辑直接加到 `cmd/skillflow/app*.go`
- 停止新增模糊上下文边界的横向公共工具包

### 阶段 1：搭建新骨架

- 新建 `platform/`、`shared/`、`orchestration/`、`readmodel/`
- 为每个 bounded context 建立 `app`、`domain`、`infra`
- 初期如有必要允许薄适配，但必须明确它只是迁移脚手架

#### 退出条件

- 至少有一个上下文已经完整迁入 `app/domain/infra`
- 对应旧包被标记为不再承接新功能
- 该区域的新业务逻辑不再落入旧包或 `cmd/skillflow/app*.go`

### 阶段 2：先抽核心域

推荐顺序：

1. `skillcatalog`
2. `promptcatalog`
3. `agentintegration`

原因：

- 这些是主要核心域
- 大多数 UI 可见业务行为都依赖它们
- 支撑域要消费的真相大多由它们提供

### 阶段 3：再抽支撑域

推荐顺序：

1. `skillsource`
2. `backup`

当前状态：

- `skillsource` 已完成抽取
- `backup` 是下一个仍待迁出的 bounded context

启动时序、tray、窗口状态、开机自启、应用更新等壳层与 platform 关注点，应保留在 `cmd/skillflow` 与 `platform/`，而不是强行塞进 bounded context。

### 阶段 4：替换横切结构

- 把旧的 `AppConfig` 改造成按上下文 namespace 归属、由共享 settings store 承载
- 把事件发布迁移到应用服务和 orchestration 服务
- 把现有 view-state 缓存迁移到 `readmodel/` 或上下文本地 projection

### 阶段 5：收缩 `cmd/skillflow`

到这个阶段，`cmd/skillflow` 应主要只剩：

- Wails transport adapter
- 进程启动
- tray/window 集成
- 壳层协调
- 依赖装配

可复用的业务方法不应再留在这里。

## 推荐优先抽取的模块

### 1. `skillcatalog`

优先迁出：

- 已安装技能聚合模型
- 分类操作
- 已安装技能列表
- 删除、移动、更新等用例

必要时可以保留临时适配层，但新的 Wails 调用路径应先变成 transport adapter，再下沉到 `skillcatalog/app`。

### 2. `promptcatalog`

第二步迁出：

- Prompt 聚合模型
- 分类模型
- 导入导出用例
- Prompt 相关应用服务

### 3. `agentintegration`

第三步迁出：

- Agent 配置模型
- 推送/拉取规划器
- 冲突判定
- 扫描与存在状态语义

## 测试策略

- 单元测试应随代码一起迁移
- 在抽取模块前，为行为定义不清晰的区域先补 characterization tests
- 每个迁移阶段都应保持 `go test ./core/...` 通过
- 对于改动到的壳层代码，执行有针对性的 `go test ./cmd/skillflow/...`
- 引入 orchestration 后，为跨上下文流程补充聚焦的集成测试

## 迁移过程中要避免的反模式

- 把旧的 `App` 巨型对象原样复制到新的上下文目录
- 在 `core/` 下再建一个全局 `service` 包
- 引入公共大 `repository` 包
- 让一个上下文直接 import 另一个上下文的 `infra`
- 把业务真相继续留在 UI DTO 组装逻辑里
- 让临时薄封装演化成永久架构

## 验收标准

当下面这些条件成立时，说明迁移方向是对的：

- 业务用例可以按 bounded context 清晰定位
- Wails `App` 方法已经退化成薄 transport adapter，而不是业务逻辑宿主
- 配置的归属以 context namespace 明确可见
- 跨上下文写流程通过 `orchestration/` 或显式壳层协调处理
- 跨上下文读视图通过 `readmodel/`
- 旧包会被持续废弃，而不是继续悄悄接收新代码

*最后更新：2026-03-21*
