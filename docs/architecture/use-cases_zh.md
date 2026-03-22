# 应用层用例

## 目标

本文档定义各 bounded context 分别拥有哪些 command 和 query，哪些流程应通过共享 orchestration 处理，以及哪些面向 UI 的组合视图应通过 read model 提供，而不是塞进领域服务。

核心规则是：

- 如果某个操作只改变一个上下文拥有的业务真相，它属于该上下文的应用层
- 如果某个操作会同时改变多个上下文的真相，它应通过显式 orchestration 处理
- 如果某个操作只是为 UI 组装数据，它属于 `readmodel/`

## 各上下文的用例归属

## `skillcatalog`

### Commands

- `ImportLocalSkill`
- `CreateInstalledSkillFromSource`
- `DeleteInstalledSkill`
- `MoveInstalledSkillToCategory`
- `CreateSkillCategory`
- `RenameSkillCategory`
- `DeleteSkillCategory`
- `UpdateInstalledSkillFromSource`
- `ReconcileInstalledSkillVersionState`

### Queries

- `GetInstalledSkill`
- `ListInstalledSkills`
- `ListSkillCategories`
- `GetInstalledSkillVersionState`
- `FindInstalledSkillByLogicalKey`

### 说明

- 对于仓库来源的 Skill，`CreateInstalledSkillFromSource` 必须强制执行 `repo + subpath -> 一个已安装 Skill` 的约束
- `ImportLocalSkill` 需要在手动导入时基于规范化内容快照生成 `LogicalSkillKey`，并将其持久化保存，因为这类 skill 后续没有 `SkillSourceRef` 可供派生
- 只刷新版本状态的更新检查仍然留在该上下文，只是来源线索来自 `skillsource`

## `promptcatalog`

### Commands

- `CreatePrompt`
- `UpdatePrompt`
- `DeletePrompt`
- `CreatePromptCategory`
- `RenamePromptCategory`
- `DeletePromptCategory`
- `MovePromptToCategory`
- `PreparePromptImport`
- `ApplyPromptImport`
- `ExportPromptBundle`

### Queries

- `GetPrompt`
- `ListPrompts`
- `ListPromptCategories`
- `PreviewPromptImportConflicts`

### 说明

- 导入 session 是应用层流程对象，不是聚合根
- 当前 Prompt 导入冲突行为是基于名称已存在的覆盖检测
- 导出格式属于应用层或基础设施层，不属于 Prompt 领域模型

## `agentintegration`

### Commands

- `RegisterAgentProfile`
- `UpdateAgentProfile`
- `RemoveAgentProfile`
- `EnableAgentProfile`
- `DisableAgentProfile`
- `PushInstalledSkills`
- `PushInstalledSkillsForce`
- `PullSkillsFromAgent`
- `PullSkillsFromAgentForce`
- `DeleteAgentSkill`
- `ReconcileAutoPushPolicy`

### Queries

- `GetAgentProfile`
- `ListAgentProfiles`
- `ScanAgentSkills`
- `ListAgentSkills`
- `CheckMissingAgentPushDirs`
- `ResolveAgentSkillPresence`

### 说明

- push / pull 冲突判定属于这个上下文，不应散落在壳层适配器或来源管理代码里
- 该上下文应消费 `skillcatalog` 发布的 summary，而不是直接持有其聚合

## `skillsource`

### Commands

- `TrackStarRepo`
- `TrackStarRepoWithCredentials`
- `UntrackStarRepo`
- `RefreshStarRepo`
- `RefreshAllStarRepos`
- `MarkStarRepoSyncFailure`
- `ClearStarRepoSyncFailure`

### Queries

- `GetStarRepo`
- `ListStarRepos`
- `ListSkillSourcesByRepo`
- `ListAllSkillSources`
- `ListSourceSkillCandidates`
- `GetSourceVersionHint`

### 说明

- `StarRepo` 是仓库级模型
- `SkillSource` 是由 `repo + subpath` 标识的技能级来源模型
- 来源候选技能的已安装、已导入、可更新状态，应通过 `skillsource` 与 `skillcatalog`、`agentintegration` 的 published language 组合得出，通常由 read model 负责

## `backup`

### Commands

- `SaveBackupProfile`
- `RunBackup`
- `RunAutoBackup`
- `RestoreBackup`
- `ResolveGitBackupConflict`
- `RecordBackupSnapshot`

### Queries

- `GetBackupProfile`
- `ListRemoteBackupFiles`
- `GetLastBackupResult`
- `GetLastBackupCompletedAt`
- `PreviewBackupChanges`

### 说明

- `backup` 拥有备份执行语义，但不拥有恢复后各上下文的业务重建逻辑
- 跨上下文的 restore compensation 应通过 orchestration 处理

## 壳层与 Platform 操作

下面这些操作是壳层或 platform 关注点，不属于 bounded context 用例：

- `SetLaunchAtLogin`
- `PersistWindowState`
- `ShowMainWindow`
- `HideMainWindow`
- `CheckAppUpdate`
- `DownloadAppUpdate`
- `ApplyAppUpdate`
- `SetSkippedUpdateVersion`

它们应保留在 `cmd/skillflow/` 与 `platform/`，而不是被建模成单独的 bounded context。

## 共享 Orchestration

下面这些流程不应归属于单一 bounded context。

### `ImportSkillFromSourceOrchestrator`

典型顺序：

1. 从 `skillsource` 读取候选来源数据
2. 在 `skillcatalog` 中创建或对齐已安装 Skill
3. 如有配置，通过 `agentintegration` 自动推送到 Agent
4. 如有配置，通过 `backup` 触发自动备份

### `ImportLocalSkillOrchestrator`

典型顺序：

1. 校验本地来源目录
2. 在 `skillcatalog` 中创建已安装 Skill
3. 如有配置，自动推送到 Agent
4. 如有配置，触发自动备份

### `UpdateInstalledSkillOrchestrator`

典型顺序：

1. 从 `skillsource` 解析来源线索
2. 在 `skillcatalog` 中更新已安装 Skill 内容和版本状态
3. 通过 `agentintegration` 刷新已推送副本
4. 如有配置，触发自动备份

### `RestoreSystemOrchestrator`

典型顺序：

1. 通过 `backup` 恢复备份内容
2. 重建各上下文的本地设置、缓存和 projection
3. 刷新派生 read model
4. 发布恢复后的事件

这些 orchestrator 属于 `core/orchestration`，对应的 Wails transport method 必须走这条路径。

## 壳层协调

下面这些流程属于协调逻辑，但更适合由壳层组合层负责，而不是落进 `core/orchestration/`。

### `SettingsSaveCoordinator`

`Settings` 是组合界面。当前实现里，它会先经过壳层协调与 `core/config`，再拆发到各自的归属组件：

- 技能库设置，例如 `defaultCategory` -> `skillcatalog`
- 源仓库缓存设置，例如 `repoCacheDir` -> `skillsource`，同时由壳层重建依赖 repo cache 的运行时适配器
- Agent 配置、自动推送和递归扫描深度 -> `agentintegration`
- 备份 provider 选择、同步间隔与云配置拆分 -> `backup`
- 开机自启、代理、日志级别、跳过更新版本、窗口状态等壳层偏好 -> `cmd/skillflow` 与 `platform/`

Prompt 持久化目前直接位于 `prompts/` 目录树，且不通过 Settings 页面编辑。收藏仓库身份跟踪仍位于 `star_repos*.json`，而本地 repo cache 根目录则通过 Settings 编辑，保存后需要同步重建依赖 cache 的壳层服务。

### `StartupBootstrapSequence`

启动时序应保留在 `cmd/skillflow/app_startup.go`、`cmd/skillflow/main.go`、`cmd/skillflow/process_bootstrap_mode*.go` 这类壳层启动文件中：

- 加载设置
- 初始化壳层适配器
- 刷新或修复运行时状态
- 调度后台来源刷新、更新检查和备份任务

## Read Model

下面这些视图应落在 `readmodel/`，而不是塞进某个 bounded context。

### `DashboardReadModel`

组合：

- `skillcatalog` 的已安装技能
- `skillsource` 的版本线索
- `agentintegration` 的已推送状态

### `MyAgentsReadModel`

组合：

- `agentintegration` 的 Agent 配置
- `agentintegration` 的 Agent 扫描结果
- `skillcatalog` 的已安装状态覆盖信息

### `StarRepoReadModel`

组合：

- `skillsource` 的 StarRepo 与 SkillSource
- `skillcatalog` 的已安装状态覆盖
- `agentintegration` 的已推送状态覆盖

## Settings Facade

`Settings` 是“UI 组合一般走 `readmodel/`”这条规则的一个例外。

当前实现使用 `core/config` 作为 `GetConfig` / `SaveConfig` 的设置门面，统一合并：

- `skillcatalog` 的技能库设置
- `agentintegration` 的 Agent 与自动推送设置
- `backup` 的备份设置
- 壳层 / platform 设置

除非未来新增专门的设置字段，否则 Prompt 存储与收藏仓库跟踪不会进入这个门面。

## Transport Adapter 映射

`cmd/skillflow/` 中对 Wails 暴露的 `App` 方法保持为薄 transport adapter，转调上下文应用服务、orchestration 服务或 read model。

例如：

- `ListSkills` -> `readmodel/skills`
- `ListAllStarSkills` / `ListRepoStarSkills` -> `readmodel/skills`
- `ImportLocal` / `PushToAgents` / `PullFromAgent` / `UpdateSkill` -> `core/orchestration`
- `ImportStarSkills` -> `orchestration/ImportSkillFromSourceOrchestrator`
- `GetConfig` / `SaveConfig` -> `core/config` 门面 + 壳层 `SettingsSaveCoordinator`
- `CreatePrompt` -> `promptcatalog/app`
- `CheckAppUpdate` -> 壳层 / platform update service

## 设计约束

- 任何用例都不能直接依赖另一个上下文的基础设施实现
- command handler 应返回领域语义结果，再由 transport adapter 映射成传输 DTO
- 跨上下文写流程必须显式存在
- UI 文案如 `imported` 可以与内部语义 `installed` 不同，但这种映射不属于领域层

*最后更新：2026-03-21*
