# 运行时、仓库布局与存储

## Wails 壳层约束

SkillFlow 仍然是一个 Wails 桌面应用。

当前后端架构保留这些约束：

- 仓库根目录不能包含 Go 源文件
- `cmd/skillflow/*.go` 必须保持扁平，因为 Wails 绑定要求单一 `package main` 目录
- Wails 绑定的 transport adapter 必须保留在 `cmd/skillflow/`
- 桌面壳层继续通过 Wails 直接绑定后端能力给前端，而不是引入 REST API

## `cmd/skillflow/` 的职责

`cmd/skillflow/` 是面向壳层的组合层。

它负责：

- Wails 启动和绑定注册
- 面向 Wails 的 transport adapter
- helper/UI 进程启动
- tray 与 window 集成
- 单实例协调
- 壳层启动时序
- 保存 Settings 时对多个上下文的 fan-out 协调
- 各上下文、orchestration 服务和 read model 的依赖装配

它不承载可复用的业务用例逻辑，例如技能导入规则、Prompt CRUD 规则、来源同步规则或 Agent 推拉语义。

## Helper / UI 双进程

SkillFlow 采用 helper/UI 双进程模型：

- `helper` 进程负责托盘、本地控制端点和 UI 拉起/聚焦
- `ui` 进程负责承载 Wails、React 以及调用后端应用服务的 transport adapter
- 关闭主窗口会退出 UI 进程，但不会终止 helper 壳层

## 仓库结构

```text
/
  go.mod
  go.sum
  Makefile
  README.md
  README_zh.md
  contributing.md
  contributing_zh.md
  docs/
    agents/
    architecture/
    config.md
    config_zh.md
    features.md
    features_zh.md
    plans/
    superpowers/
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
    config/
  cmd/
    skillflow/
      main.go
      app.go
      app_*.go
      app_startup.go
      adapters.go
      providers.go
      events.go
      process_*.go
      tray_*.go
      window_*.go
      single_instance_*.go
      frontend/
```

## 存储布局

逻辑归属按 bounded context 拆分，物理存储保持简单可运维。

当前持久化布局分为固定的应用数据根目录，以及可选的本地仓库缓存根目录两部分：

```text
<AppDataDir>/
  config.json          # shared 的可同步设置负载
  config_local.json    # local-only 设置、路径、密钥与运行时状态
  star_repos.json      # skillsource 的仓库跟踪状态
  star_repos_local.json
  skills/
    <category>/<skill>/
  meta/
  meta_local/
  prompts/
    <category>/<name>/
  cache/
    viewstate/
  runtime/
  logs/

<RepoCacheDir>/        # 本地专属的仓库 clone cache 根目录；默认值为 <AppDataDir>/cache/repos
  <git-cache-hosts...>
```

`config.json` 和 `config_local.json` 是通过 `core/config` 管理的扁平 shared/local 负载。字段归属是逻辑上的，而不是字面上的顶层 namespace，因此物理文件不需要与 bounded context 做 1:1 映射。

## 配置归属

`config.json` 和 `config_local.json` 是共享存储文件，但其中字段的逻辑归属仍然按上下文和 platform 关注点划分。

逻辑归属：

- `skillcatalog`
  - 默认技能分类
- `skillsource`
  - 收藏仓库 clone cache 的本地根目录
- `agentintegration`
  - Agent 配置
  - 自动推送策略
  - 仓库/智能体递归扫描深度
- `backup`
  - 当前备份配置
  - provider 与同步间隔
  - shared/local 配置中的云配置与凭据拆分
- 壳层 / platform
  - 开机自启
  - 窗口状态
  - 跳过更新版本
  - 代理与日志级别偏好

位于 `config*.json` 之外的额外持久化归属：

- `promptcatalog`
  - `prompts/` 下的 Prompt 内容与元数据
- `skillsource`
  - `star_repos.json` / `star_repos_local.json` 中的收藏仓库状态
  - 当前 `repoCacheDir` 下的 repo cache（默认 `<AppDataDir>/cache/repos`）

实现归属：

- app data 路径归属位于 `core/platform/appdata`
- 壳层的 proxy、window、log level、skipped update 偏好位于 `core/platform/shellsettings` 与 `core/platform/settingsstore`
- import、push/pull、update、restore compensation 等跨上下文写流程位于 `core/orchestration`
- installed skills、starred skills、agent presence 的组合位于 `core/readmodel/skills` 与 `core/readmodel/viewstate`
- `core/config` 是面向前端的设置门面与 split/merge 持久化适配器，底层委托给这些上下文与 platform 拥有的设置组件

## Repository 与 Gateway 示例

Repository：

- 已安装技能元数据存储
- Prompt 库存储
- Agent 配置存储
- 来源跟踪存储
- 设置门面的持久化视图

Gateway：

- Agent 工作区适配器
- 用于同步外部来源的 Git 客户端封装
- 云备份 provider 适配器
- GitHub Releases API 客户端
- Wails runtime 适配器，例如文件对话框或系统打开路径能力

## 事件与派生状态

转发事件给前端仍然属于壳层集成职责。事件发布点应位于应用服务和 orchestration 服务附近。

安装技能卡片、Agent 聚合存在状态等派生快照，应放在 `readmodel/` 或上下文本地的 `infra/projection/`，不应塞进领域模型。

## 日志与路径可移植性

以下规则仍然有效：

- 日志仍然限制为 `skillflow.log` 和 `skillflow.log.1`
- 位于同步根目录下的路径，应以正斜杠相对路径持久化
- 机器相关路径只保留在本地配置 namespace 中
- 敏感信息不能写入日志

*最后更新：2026-03-21*
