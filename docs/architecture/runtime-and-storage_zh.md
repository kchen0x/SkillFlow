# 运行时、仓库布局与存储

## Wails 壳层约束

SkillFlow 仍然是一个 Wails 桌面应用。

当前和未来架构都必须保留这些约束：

- 仓库根目录不能包含 Go 源文件
- `cmd/skillflow/*.go` 必须保持扁平，因为 Wails 绑定要求单一 `package main` 目录
- Wails 绑定的 transport adapter 必须保留在 `cmd/skillflow/`
- 桌面壳层继续通过 Wails 直接绑定后端能力给前端，而不是引入 REST API

## 重构后 `cmd/skillflow/` 的职责

`cmd/skillflow/` 应收缩为面向壳层的组合层。

它应保留：

- Wails 启动和绑定注册
- 面向 Wails 的 transport adapter
- helper/UI 进程启动
- tray 与 window 集成
- 单实例协调
- 壳层启动时序
- 保存 Settings 时对多个上下文的 fan-out 协调
- 各上下文、orchestration 服务和 read model 的依赖装配

它不应继续承载可复用的业务用例逻辑，例如技能导入规则、Prompt CRUD 规则、来源同步规则或 Agent 推拉语义。

## Helper / UI 双进程

当前 helper/UI 双进程模型在 DDD 重构后仍然成立：

- `helper` 进程负责托盘、本地控制端点和 UI 拉起/聚焦
- `ui` 进程负责承载 Wails、React 以及调用后端应用服务的 transport adapter
- 关闭主窗口会退出 UI 进程，但不会终止 helper 壳层

DDD 重构改变的是后端逻辑的归属，不改变这套桌面壳层拓扑。

## 目标仓库结构

```text
/
  go.mod
  Makefile
  docs/
    architecture.md
    architecture_zh.md
    architecture/
  core/
    platform/
    shared/
    orchestration/
    readmodel/
    skillcatalog/
    promptcatalog/
    agentintegration/
    skillsource/
    backup/
  cmd/
    skillflow/
      main.go
      bootstrap.go
      app.go
      app_*.go
      process_*.go
      tray_*.go
      window_*.go
      frontend/
```

## 存储方向

逻辑归属应按 bounded context 拆分，但物理存储应尽量保持简单可运维。

推荐的目标结构：

```text
<AppDataDir>/
  config.json          # shared，按上下文 namespace 分段
  config_local.json    # local-only，按上下文与壳层关注点 namespace 分段
  star_repos.json      # skillsource 的仓库跟踪状态
  skills/
    library/
    meta/
    meta_local/
  prompts/
    library/
  cache/
    sources/
    readmodel/
  runtime/
  logs/
```

`config.json` 和 `config_local.json` 内部按上下文 namespace 划分，每个上下文通过 platform settings store 读取和写入自己的视图。物理文件不需要与 bounded context 做 1:1 映射。

## 配置归属

旧的 `AppConfig` 应视为迁移期兼容结构，而不是长期领域模型。

推荐的逻辑归属：

- `skillcatalog`
  - 技能库路径
  - 默认技能分类
- `promptcatalog`
  - Prompt 库路径
  - Prompt 导入导出默认策略
- `agentintegration`
  - Agent 配置
  - 自动推送策略
- `skillsource`
  - 来源凭据元数据
  - 来源刷新默认策略
- `backup`
  - 当前备份配置
  - provider 与同步间隔
- 壳层 / platform
  - 开机自启
  - 窗口状态
  - 跳过更新版本
  - 代理与日志级别偏好

当前迁移说明：

- app data 路径归属现在位于 `core/platform/appdata`
- 壳层的 proxy、window、log level、skipped update 偏好现在位于 `core/platform/shellsettings` 与 `core/platform/settingsstore`
- Skill 状态可见性偏好现在位于 `core/readmodel/preferences`
- import、push/pull、update、restore compensation 等跨上下文写流程现在位于 `core/orchestration`
- installed skills、starred skills、agent presence 的组合现在位于 `core/readmodel/skills` 与 `core/readmodel/viewstate`
- `core/config` 现在主要作为面向前端的兼容 DTO 与 split/merge 持久化门面，底层委托给这些上下文与 platform 拥有的设置组件

## Repository 与 Gateway 示例

Repository：

- 已安装技能元数据存储
- Prompt 库存储
- Agent 配置存储
- 来源跟踪存储
- 按 namespace 提供视图的 settings store

Gateway：

- Agent 工作区适配器
- 用于同步外部来源的 Git 客户端封装
- 云备份 provider 适配器
- GitHub Releases API 客户端
- Wails runtime 适配器，例如文件对话框或系统打开路径能力

## 事件与派生状态

转发事件给前端仍然属于壳层集成职责，但事件发布点应逐步下沉到应用服务和 orchestration 服务。

安装技能卡片、Agent 聚合存在状态等派生快照，应放在 `readmodel/` 或上下文本地的 `infra/projection/`，不应塞进领域模型。

## 日志与路径可移植性

以下规则仍然有效：

- 日志仍然限制为 `skillflow.log` 和 `skillflow.log.1`
- 位于同步根目录下的路径，应以正斜杠相对路径持久化
- 机器相关路径只保留在本地配置 namespace 中
- 敏感信息不能写入日志

*最后更新：2026-03-21*
