# SkillFlow — 架构与开发者参考

> 🌐 **中文** | [English](architecture.md)

本文档面向贡献者，说明 SkillFlow 的内部架构、仓库布局、关键数据模型和扩展点。面向用户的交互与功能说明请查看 **[docs/features_zh.md](features_zh.md)**。

---

## 概述

SkillFlow 是一个基于 **Wails v2** 的桌面应用，后端使用 **Go 1.23**，前端使用 **React 18 + TypeScript**。

- Go 后端位于 `cmd/skillflow/`
- React 前端位于 `cmd/skillflow/frontend/`
- 二者通过 Wails 生成绑定直接通信
- **没有 REST API**

核心技术栈：

- 后端：Go、Wails、Git、各云服务商 SDK
- 前端：React、TypeScript、React Router、Tailwind CSS、Radix UI、Lucide
- 构建：`make`、Wails CLI、Vite

---

## 仓库布局

```text
/                              模块根目录，不放 Go 源文件
  go.mod
  Makefile
  README.md
  README_zh.md
  docs/
  core/                        可复用包，不使用 package main
  cmd/
    skillflow/                 Wails 桌面应用，package main
      main.go
      app.go
      app_*.go
      adapters.go
      providers.go
      events.go
      version.go
      tray_*.go
      single_instance_*.go
      window_*.go
      wails.json
      build/
      frontend/
```

关键规则：

- 根目录 **不允许** 出现 `.go` 文件。
- `cmd/skillflow/*.go` 必须保持 **扁平**，因为 Wails 绑定要求同一个 `package main` 目录。
- 新的可复用后端逻辑应放到 `core/<name>/`，而不是再给 `cmd/skillflow/` 建 Go 子目录。
- `wails dev`、`wails build`、`wails generate module` 都要在 `cmd/skillflow/` 下执行；从仓库根目录优先使用 `make dev`、`make build`、`make generate`。

---

## 运行时生命周期

### Wails 入口

`cmd/skillflow/main.go` 负责：

- 在真正启动外壳前先判断内部进程角色（`helper` 或 `ui`）
- 让 helper 持有单实例、托盘/菜单栏控制，以及 UI 拉起/聚焦能力
- 在 UI 子进程里用 `//go:embed all:frontend/dist` 启动 Wails，并把 `HideWindowOnClose` 设为 `false`
- 仅在 UI 进程中创建并绑定 `App` 给前端

### Helper / UI 双进程

- **helper** 进程是轻量壳层，负责托盘/菜单栏、单实例和本地控制端点（位于 `<AppDataDir>/runtime/` 下），并按需重新拉起或聚焦 UI 进程。
- **UI** 进程承载 Wails、React 以及前端可调用的 `App` 绑定。关闭窗口时会直接退出该进程，并随之释放嵌入的 WebKit/WebView 运行时。
- 再次启动应用时，不再直接尝试聚焦旧的 Wails 窗口，而是向 helper 转发 `show-ui` 指令后退出当前启动实例。
- 当前阶段里，后端 `App` 方法仍运行在 UI 进程中；helper 目前只接管壳层生命周期。

### 应用启动流程

`App.startup()` 完成核心后端初始化：

1. 解析 `config.AppDataDir()`
2. 通过 `core/config.Service` 加载配置
3. 初始化滚动日志
4. 创建 Skill 存储、收藏仓库存储，以及本地派生视图缓存管理器
5. 注册工具适配器与云服务商
6. 启动后端事件转发
7. 启动云备份自动同步计时器

`App.domReady()` 负责外壳/UI 相关初始化：

1. 恢复或计算窗口初始尺寸
2. 启动 UI 侧本地控制服务（供 helper 发送 `show`、`hide`、`quit`）
3. 延迟调度启动后的后台任务
4. 将这些任务错峰展开，避免首个可交互阶段同时发起所有远端检查
5. 在前端维护一个活动状态机，用于在长时间后台或不活跃后卸载路由页面树

当前延迟后台任务包括：

- Git 备份启动拉取（最早）
- Skill 更新检测
- 收藏仓库刷新
- 应用版本更新检测

`App.beforeClose()` 会持久化当前窗口尺寸，`App.shutdown()` 会停止 UI 控制服务。

---

## 应用数据布局

默认情况下，SkillFlow 将数据保存在 `config.AppDataDir()`：

- macOS：`~/Library/Application Support/SkillFlow/`
- Windows：`%USERPROFILE%\\.skillflow\\`

```text
<AppDataDir>/
  skills/                 已安装 Skill 库
  meta/                   每个已安装 Skill 的 JSON sidecar
  prompts/                提示词库
  cache/
    <repo-cache>/         收藏仓库的本地克隆缓存
    viewstate/            本地专属的派生 UI 快照
  runtime/
    helper-control.json   helper 回环控制端点
    ui-control.json       UI 回环控制端点
    helper.lock           本地单实例锁 / 进程协调状态
  logs/
    skillflow.log
    skillflow.log.1
  config.json             可同步的共享配置
  config_local.json       本地专属配置
  star_repos.json         收藏仓库元数据
```

重要规则：

- `config.json` 只保存跨设备可同步的安全配置。
- `config_local.json` 保存机器相关路径、自动推送目标、开机自启、代理、窗口状态、自定义工具路径配置，以及敏感云凭据。
- `meta/*.json`、`star_repos.json` 等可同步文件中的本地路径，在目标位于同步根目录内时必须以 **正斜杠相对路径** 形式持久化。
- 如果 `SkillsStorageDir` 被移出默认应用数据目录，同步根目录会变成 `skills/` 与 `meta/` 的共同父目录。
- 日志文件固定为 **两个**，每个 **1MB** 上限：`skillflow.log` 与 `skillflow.log.1`。
- `cache/viewstate/*.json` 只保存可重建的派生状态，例如已安装 Skill 快照和工具 presence 索引。这些文件只存在于本机，绝不能被视为可同步真值。
- `runtime/*` 保存 helper/UI 的回环端点、token、PID 以及单实例协调状态。这些文件只属于当前设备，绝不能参与同步，需要时会自动重建。

---

## 包职责

| 路径 | 职责 |
|------|------|
| `cmd/skillflow` | Wails 入口、前端可调用的 `App` 方法、托盘/窗口/单实例集成、适配器与服务商注册 |
| `core/applog` | 日志滚动与日志级别处理 |
| `core/backup` | 备份快照、服务商接口、Git provider、对象存储 provider |
| `core/config` | 共享/本地配置拆分持久化、默认值、状态显示策略归一化 |
| `core/git` | Git clone/pull/push 辅助、收藏仓库扫描、收藏仓库存储 |
| `core/install` | GitHub 安装和本地导入流程 |
| `core/notify` | 缓冲事件总线与事件载荷类型 |
| `core/pathutil` | 跨平台路径归一化与相对路径持久化辅助 |
| `core/prompt` | 提示词库存储与导入导出 |
| `core/registry` | 工具适配器和云服务商的全局注册表 |
| `core/skill` | Skill 模型、存储、校验、已安装索引 |
| `core/skillkey` | Git Skill 与内容型 Skill 的稳定逻辑主键生成 |
| `core/sync` | `ToolAdapter` 接口和基于文件系统的适配器实现 |
| `core/update` | 保留给测试和直连式辅助场景的 GitHub commit 检测工具；已安装 Skill 的更新状态现在来自本地仓库缓存 SHA 比较 |
| `core/viewstate` | 本地派生快照缓存、fingerprint 计算，以及增量工具 presence 辅助逻辑 |

---

## `cmd/skillflow/` 文件组织

Wails 应用包必须保持扁平，因此通过文件名前缀划分职责：

| 文件组 | 作用 |
|--------|------|
| `main.go`、`version.go` | 入口与构建时版本号 |
| `app.go` | 主 `App` 结构体及大部分前端可调用方法 |
| `app_viewstate.go`、`app_perf.go` | 本地快照缓存、fingerprint 计算和轻量性能计时辅助 |
| `app_prompt.go` | 提示词 CRUD 与提示词导入导出 |
| `app_update.go` | 应用版本检测、下载、应用更新、跳过版本逻辑 |
| `app_log.go` | 日志初始化与 Wails runtime 日志桥接 |
| `app_restore.go`、`app_backup.go` | 恢复补偿逻辑、Git 备份辅助 |
| `app_autostart.go`、`window_size.go`、`app_path.go` | 操作系统集成辅助 |
| `adapters.go`、`providers.go` | 注册 `core/sync` 适配器与 `core/backup` 服务商 |
| `events.go`、`push_conflict.go`、`skill_state.go` | 前端 DTO 与跨页面状态聚合 |
| `tray_*.go`、`single_instance_*.go`、`window_*.go` | 平台相关的壳层行为 |

---

## 派生视图缓存

为了让页面切换和大列表渲染更流畅，后端现在会在 `cache/viewstate/` 下维护本地专属的派生缓存。

- `ListSkills()` 采用 snapshot-first：当已安装 Skill 的 fingerprint 匹配时，会直接返回缓存的 `InstalledSkillEntry[]`，而不是立刻重新构建 push presence。
- 已安装 Skill fingerprint 由同步相关的 skill 元数据和工具 push 目录摘要共同决定，因此用户切走页面再回来时，如果这些依赖发生变化，就会重新读取当前真值。
- 工具 push presence 按工具粒度增量重建，使用每个工具自己的 fingerprint，而不是每次请求都把所有配置过的 `PushDir` 全量重扫一遍。

这些缓存只是性能优化产物：

- 始终从权威磁盘状态重建
- 删除后可自动恢复
- 不会跨设备同步

---

## 关键数据模型

### Skill（`core/skill/model.go`）

```go
type Skill struct {
    ID            string
    Name          string
    Path          string
    Category      string
    Source        SourceType
    SourceURL     string
    SourceSubPath string
    SourceSHA     string
    LatestSHA     string
    InstalledAt   time.Time
    UpdatedAt     time.Time
    LastCheckedAt time.Time
}
```

说明：

- `ID` 是已安装实例的 UUID。
- `Path` 运行时为绝对路径；若写入可同步元数据，应尽量保存为可移植的相对路径。
- `SourceURL + SourceSubPath` 共同标识 GitHub 安装 Skill 的逻辑来源。

### AppConfig（`core/config/model.go`）

```go
type AppConfig struct {
    SkillsStorageDir      string
    AutoPushTools         []string
    LaunchAtLogin         bool
    DefaultCategory       string
    LogLevel              string
    RepoScanMaxDepth      int
    SkillStatusVisibility SkillStatusVisibilityConfig
    Tools                 []ToolConfig
    Cloud                 CloudConfig
    CloudProfiles         map[string]CloudProviderConfig
    Proxy                 ProxyConfig
    SkippedUpdateVersion  string
}
```

配置拆分规则：

- **共享 / 可同步**：`DefaultCategory`、`LogLevel`、`RepoScanMaxDepth`、状态显示策略、内置工具启用状态、当前云服务商状态、云服务商非敏感配置、跳过的应用版本。
- **本地专属**：`SkillsStorageDir`、`AutoPushTools`、`LaunchAtLogin`、工具路径、自定义工具定义、代理、窗口尺寸、敏感云凭据。

相关嵌套模型：

```go
type ToolConfig struct {
    Name     string
    ScanDirs []string
    PushDir  string
    Enabled  bool
    Custom   bool
}

type CloudConfig struct {
    Provider            string
    Enabled             bool
    BucketName          string
    RemotePath          string
    Credentials         map[string]string
    SyncIntervalMinutes int
}
```

### 收藏仓库模型（`core/git/model.go`）

```go
type StarredRepo struct {
    URL       string
    Name      string
    Source    string
    LocalDir  string
    LastSync  time.Time
    SyncError string
}

type StarSkill struct {
    Name        string
    Path        string
    SubPath     string
    RepoURL     string
    RepoName    string
    Source      string
    LogicalKey  string
    Installed   bool
    Imported    bool
    Updatable   bool
    Pushed      bool
    PushedTools []string
}
```

---

## 统一的 Skill 身份与状态模型

本节对所有涉及 Skill 卡片、导入/安装/推送/拉取流程、收藏仓库、工具扫描或更新徽标的改动都具有约束力。

### 身份分层

SkillFlow 区分两类身份：

- **实例身份**：`Skill.ID`，用于“我的 Skills”中的单个已安装副本操作，例如删除、移动分类、实例更新。
- **逻辑身份**：跨模块稳定身份，用来判断 Dashboard、Starred Repos、Tool Skills、Pull、Push 中展示的是不是同一个 Skill。

`Name` 和绝对 `Path` 只是展示或定位信息，**不能** 作为跨模块主键。

### 逻辑主键规则

- **Git 来源 Skill** 使用 `git:<repo-source>#<subpath>`
  - `repo-source`：规范化 host/path，例如 `github.com/owner/repo`
  - `subpath`：仓库内正斜杠路径，例如 `skills/my-skill`
- **非 Git Skill** 应使用稳定内容主键，例如 `content:<hash>`
- 只有在暂时无法生成稳定主键时，才允许弱匹配兜底

### 模块映射

| 模块 / 页面 | 主要实体 | 驱动行为的身份 |
|-------------|----------|----------------|
| Dashboard / 我的 Skills | 已安装 `Skill` | 实例操作用 `Skill.ID`；跨模块关联用逻辑主键 |
| Sync Push | 已安装 `Skill` | 选择用 `Skill.ID`；推送状态解析用逻辑主键 |
| GitHub 扫描/安装 | 远端候选项 | 用 repo source + subpath 派生的逻辑主键 |
| Starred Repos | `StarSkill` | 用 repo source + subpath 派生的逻辑主键 |
| Tool Skills | 工具侧候选项 / 聚合项 | 去重与状态用逻辑主键；工具内打开/删除才用 path |
| Sync Pull | 工具侧候选项 | 导入与冲突检测用逻辑主键 |

### 统一状态语义

- **installed**：该逻辑主键在“我的 Skills”中至少存在一个已安装实例
- **imported**：外部来源页面对 `installed` 的文案别名
- **pushed**：该逻辑 Skill 已存在于某工具配置的 `PushDir`
- **seenInToolScan**：该逻辑 Skill 已出现在某工具配置的 `ScanDirs`；这 **不代表** SkillFlow 已经推送过它
- **updatable**：至少有一个已安装 Git Skill 实例在本地缓存仓库中的 SHA 比本地 `SourceSHA` 更新

### 状态与去重规则

- `pushed` 比“工具里某处存在”更窄，它特指配置的推送目标。
- `seenInToolScan` 是观察型状态，不能被误写成“已推送”。
- 跨模块去重必须优先使用逻辑主键。
- 不同仓库中的同名项，只要逻辑主键不同，就必须视为不同 Skill。
- 仅按名称匹配只能作为最后兼容兜底，且不能覆盖更强的逻辑主键匹配。

### 更新规则

- 只有具备稳定 repo source + subpath，且其对应仓库 clone 已经存在于本地 `cache/` 树中的已安装 Git Skill 才参与更新检测。
- 缓存查询与已安装实例关联必须使用同一逻辑 Git 主键。
- `CheckUpdates()` 会把已安装 Skill 的 `SourceSHA` 与本地缓存仓库中同一 `SourceSubPath` 的最新提交 SHA 做比较，不会直接调用 GitHub Commits API。
- `UpdateSkill()` 会把该缓存仓库子目录的文件复制到已安装库目录中，然后再基于更新后的已安装副本刷新已推送到工具目录的副本。
- 当最新检查确认本地已是最新时，应清空 `LatestSHA`。
- 每次完成检查后都应更新 `LastCheckedAt`，并将其写入仅本地持久化的 `meta_local/<skill-id>.local.json`（不参与同步）。

### 实现指导

- 跨模块关联由后端统一负责，并将归一化状态返回给前端。
- 前端页面不应仅基于 `Name` 或 `Path` 自行判断“是否同一个 Skill”“是否已导入”“是否已推送”。
- `core/skillkey` 负责生成逻辑主键。
- `core/skill.BuildInstalledIndex` 负责把 GitHub 扫描、收藏仓库、工具扫描结果关联回已安装状态。

---

## 事件与绑定

### 后端事件流

SkillFlow 使用 `core/notify.Hub` 作为缓冲事件总线：

1. 后端发布 `notify.Event`
2. `forwardEvents()` 订阅并通过 Wails `runtime.EventsEmit` 转发
3. 前端通过 `cmd/skillflow/frontend/wailsjs/runtime` 订阅

当前事件总线缓冲区大小为 **32**，对慢消费者采用丢弃最旧事件策略。

主要事件分组：

- 备份：`backup.started`、`backup.progress`、`backup.completed`、`backup.failed`
- 同步/更新：`sync.completed`、`update.available`、`skill.conflict`
- 收藏仓库：`star.sync.progress`、`star.sync.done`
- Git 备份：`git.sync.started`、`git.sync.completed`、`git.sync.failed`、`git.conflict`
- 应用更新：`app.update.available`、`app.update.download.done`、`app.update.download.fail`
- 窗口生命周期：`app.window.visibility.changed`

### Wails 生成绑定

生成后的绑定位于：

- `cmd/skillflow/frontend/wailsjs/go/main/App.js`
- `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`

当导出的 `App` 方法发生变化时，需要运行 `make generate`。

---

## 前端结构

```text
cmd/skillflow/frontend/
  src/
    App.tsx
    main.tsx
    pages/
    components/
    contexts/
    i18n/
    lib/
    config/
  tests/
  wailsjs/
```

主要前端区域：

- `src/pages/`：Dashboard、Sync Push/Pull、Starred Repos、Backup、Settings、My Tools、My Prompts 等路由页
- `src/components/`：卡片、对话框、分类栏、列表控件等共享组件
- `src/contexts/`：语言、主题、状态显示策略等上下文状态
- `src/i18n/`：中英文翻译词典
- `src/lib/`：列表、搜索、剪贴板、状态管理等共享辅助
- `tests/`：不依赖 Wails 打包流程的前端单测

`App.tsx` 还负责应用活动状态机：

- 把后端的隐藏/显示信号与浏览器侧的 focus / visibility 合并成统一的前后台模型
- 进入后台约 30 秒后，卸载路由页面树，释放页面局部数组和提示词正文等内存
- 窗口再次激活时，当前路由会从头挂载并重新加载最新数据

前端通过 Wails 生成模块直接导入后端方法，例如：

```ts
import { ListSkills } from '../../wailsjs/go/main/App'
```

---

## 日志与路径可移植性

### 日志

- `core/applog.Logger` 将结构化文本日志写入应用数据目录下的 `logs/`
- `cmd/skillflow/app_log.go` 会把启用级别的日志同步输出到 Wails runtime 日志
- 后端改动应为关键变更、同步、备份、Git 操作和外部 API 调用提供 started / completed / failed 级别日志
- 严禁记录密钥、令牌、密码等敏感信息

### 路径处理

- `core/pathutil` 负责把可同步路径归一化为正斜杠相对路径
- 运行时 API 可以在返回给调用方前把它们还原为绝对路径
- 备份、恢复和跨设备同步都依赖这套可移植路径规则

---

## 测试与构建工作流

- 在仓库根目录运行后端测试：`go test ./core/...`
- Wails 的开发、构建、绑定生成都在 `cmd/skillflow/` 下执行；也可直接使用 `make dev`、`make build`、`make generate`
- 前端依赖位于 `cmd/skillflow/frontend/package.json`
- 前端单测位于 `cmd/skillflow/frontend/tests/`
- 生产构建输出位于 `cmd/skillflow/build/bin/`

---

## 扩展指南

### 新增前端可调用的 App 方法

1. 在 `cmd/skillflow/app.go` 或其他扁平的 `cmd/skillflow/*.go` 文件中添加导出方法
2. 运行 `make generate`
3. 在前端通过 `../../wailsjs/go/main/App` 导入

### 新增云服务商

1. 创建 `core/backup/<name>.go` 并实现 `backup.CloudProvider`
2. 在 `cmd/skillflow/providers.go` 中注册
3. 通过 `RequiredCredentials()` 暴露其凭据字段

### 新增工具适配器

1. 标准文件系统型工具可直接在 `cmd/skillflow/adapters.go` 中注册
2. 如需自定义行为，实现 `toolsync.ToolAdapter`
3. 导入 `core/sync` 时始终使用 `toolsync` 别名，避免与标准库 `sync` 冲突

### 新增可复用后端模块

- 可复用逻辑优先放到 `core/<name>/`
- 不要在 `cmd/skillflow/` 下再创建 Go 子目录
- Wails 壳层代码留在 `cmd/skillflow/`，可复用领域逻辑沉淀到 `core/`

---

*最后更新：2026-03-11*
