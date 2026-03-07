# SkillFlow — 架构与开发者参考

> 🌐 **中文** | [English](architecture.md)

本文档面向贡献者，涵盖内部架构、包设计、数据模型和扩展指南。

---

## 概述

SkillFlow 是一款基于 **Wails v2** 的桌面应用（Go 1.23，Wails v2.11.0）。Go 后端通过 Wails 方法绑定直接将方法暴露给 React 前端。**没有 REST API** — 前端以异步函数形式直接调用 Go 方法。

**技术栈：**
- 后端：Go 1.23、Wails v2
- 前端：React 18、TypeScript、React Router v7、Tailwind CSS、Lucide React、Radix UI
- 构建：Wails CLI、Vite

---

## 关键设计决策

- **`core/sync` 包名与 Go 标准库 `sync` 冲突** — 必须使用别名导入：`toolsync "github.com/shinerio/skillflow/core/sync"`
- **Wails 绑定为自动生成** — 在 `App` 上增删导出方法后，需运行 `wails generate module` 更新 `frontend/wailsjs/go/main/App.{js,d.ts}`
- **根目录下的 `package main` 文件** — `app.go`、`adapters.go`、`providers.go`、`events.go` 均与 `main.go` 同属 `package main`，因为 Wails 要求 App 结构体与 `main` 位于同一包
- **无 REST API** — 直接使用 Wails 方法绑定，更快更简洁
- **已安装 Skill 实例使用 UUID，但跨模块关联必须使用稳定的逻辑主键** — 见下文的[统一的 Skill 身份与状态模型](#统一的-skill-身份与状态模型)
- **文件系统适配器** — 所有内置工具共享同一 `FilesystemAdapter` 模式
- **以 GitHub 为数据来源** — 更新检测器轮询 GitHub API，而非本地时间戳

---

## 数据存储目录结构

```
~/.skillflow/
  skills/              ← SkillsStorageDir（可配置）
    <category>/
      <skill-name>/    ← 复制的 Skill 目录
        skill.md       ← 带 YAML 前置元数据的主文件
        ...其他文件
  meta/                ← JSON 附属文件（与 skills/ 同级）
    <uuid>.json        ← 每个 Skill 一个，包含 Skill 结构体
  config.json          ← AppConfig（工具、云端、代理）
  star_repos.json      ← StarredRepo[] 数组
  cache/               ← 收藏仓库的临时克隆目录
    <cached-repo-dirs>/
```

已安装的 Skill 实例以 UUID 标识。跨模块关联必须遵循下文[统一的 Skill 身份与状态模型](#统一的-skill-身份与状态模型)中的逻辑主键规则。`meta/` 目录始终为 `filepath.Join(filepath.Dir(root), "meta")`。

---

## 后端包职责

| 包 | 职责 |
|----|------|
| `core/skill` | `Skill` 模型、`Storage`（增删改查 + 分类）、`Validator`（skill.md 校验） |
| `core/config` | `AppConfig` 模型、`Service`（JSON 加载/保存）、各工具的 `DefaultToolsDir()` |
| `core/notify` | `Hub`（缓冲通道发布/订阅）、`EventType` 常量 |
| `core/install` | `Installer` 接口、`GitHubInstaller`（扫描/下载/SHA）、`LocalInstaller` |
| `core/sync` | `ToolAdapter` 接口、`FilesystemAdapter`（所有内置工具共用） |
| `core/backup` | `CloudProvider` 接口、阿里云/腾讯云/华为云实现 |
| `core/update` | `Checker`（GitHub Commits API SHA 对比） |
| `core/registry` | Installer/ToolAdapter/CloudProvider 的全局映射 — 启动时注册 |
| `core/git` | Git 克隆/更新、仓库 Skill 扫描、收藏仓库存储 |

---

## 关键数据模型

### Skill（`core/skill/model.go`）

```go
type Skill struct {
    ID            string     // UUID
    Name          string     // skill 名称（目录名）
    Path          string     // 运行时绝对路径；在 meta/*.json 中以同步根目录下的相对路径持久化
    Category      string     // 用户定义的分类
    Source        SourceType // "github" | "manual"
    SourceURL     string     // GitHub 源的仓库 URL
    SourceSubPath string     // 仓库内的相对路径（如 "skills/my-skill"）
    SourceSHA     string     // 已安装的 commit SHA（来自 GitHub）
    LatestSHA     string     // 检测到的更新 SHA（用于更新检测）
    InstalledAt   time.Time
    UpdatedAt     time.Time
    LastCheckedAt time.Time
}

const (
    SourceGitHub SourceType = "github"
    SourceManual SourceType = "manual"
)
```

## 统一的 Skill 身份与状态模型

本节对所有后续涉及 skill 卡片、导入/安装/推送/拉取流程、收藏仓库、工具扫描或更新徽标的改动都具有约束力。

### 身份分层

SkillFlow 必须区分两类身份：

- **实例身份** — `Skill.ID` 标识“我的 skills”中的某一份已安装副本。删除、移动分类、实例级更新等安装实例操作应使用它。
- **逻辑身份** — 稳定的跨模块身份，用来回答“Dashboard、Starred Repos、Tool Skills、Sync Pull、Sync Push 中显示的是不是同一个 skill”。

`Name` 和绝对 `Path` 只是展示信息或位置信息，不能作为跨模块主身份键。

### 逻辑主键规则

- **Git 来源 skill** 必须使用“规范化仓库来源 + 仓库内子路径”生成逻辑主键：
  - 格式：`git:<repo-source>#<subpath>`
  - `repo-source` 为规范化后的 host/path，例如 `github.com/owner/repo`
  - `<subpath>` 为仓库内使用正斜杠的相对路径，例如 `skills/my-skill`
- **非 Git skill** 应使用稳定的内容型主键，例如 `content:<hash>`，这样在工具扫描和本地导入时也能识别为同一个 skill。
- **临时兜底启发式** 仅可在尚无法生成稳定逻辑主键时使用，并且必须视为弱匹配。

### 模块映射

| 模块 / 页面 | 主要实体 | 驱动行为的身份 |
|-------------|----------|----------------|
| Dashboard / 我的 Skills | 已安装 `Skill` | 实例操作使用 `Skill.ID`；跨模块关联使用逻辑主键 |
| Sync Push | 已安装 `Skill` | 选择使用 `Skill.ID`；推送状态解析使用逻辑主键 |
| GitHub 扫描/安装 | 远端候选项 | 使用 repo source + subpath 派生的逻辑主键 |
| Starred Repos | `StarSkill` | 使用 repo source + subpath 派生的逻辑主键 |
| Tool Skills | 工具侧候选项 / 聚合项 | 去重与状态使用逻辑主键；仅在该工具内的打开/删除使用 path |
| Sync Pull | 工具侧候选项 | 导入与冲突检测使用逻辑主键 |

### 统一状态语义

- **installed** — 该逻辑主键在“我的 skills”中至少存在一个已安装实例。
- **imported** — 外部来源页面对 `installed` 的文案别名；在 GitHub、Starred Repos、Tool 扫描视图中，“已导入”表示“已经安装进我的 skills”。
- **pushed** — 该逻辑 skill 已存在于某工具配置的 `PushDir` 中，可视为已经推送到该工具。
- **seenInToolScan** — 该逻辑 skill 已在某工具配置的 `ScanDirs` 中被扫描到，表示该 skill 已出现在工具生态中，但不一定由 SkillFlow 管理，也不一定在 `PushDir` 中。
- **updatable** — 至少有一个已安装的 Git skill 实例，其远端最新提交与本地 `SourceSHA` 不一致。

### 状态规则

- `pushed` 比“工具里某处存在”更窄，它特指配置的推送目标中存在。
- `seenInToolScan` 是观察型状态，用来区分“工具里已经有这个 skill”和“SkillFlow 已经把它推过去了”。
- 当推送目录也被扫描，或同一个逻辑 skill 同时存在于 push/scan 两处时，`pushed=true` 与 `seenInToolScan=true` 可以同时成立。
- 当 `seenInToolScan=true` 且 `pushed=false` 时，界面通常应表达为“工具中已检测到”，而不是“已推送”。

### 冲突与去重规则

- 跨模块去重必须优先使用逻辑主键相等。
- 不同仓库中的同名项，只要逻辑主键不同，就必须视为不同 skill。
- 相同路径不自动代表同一个 skill，除非它们能解析为同一个逻辑主键。
- 基于名称的匹配只能作为最后的兼容性兜底，且不能覆盖更强的逻辑主键匹配。

### 更新检测规则

- 只有具备稳定 repo source 和 subpath 的已安装 Git skill 才参与远端更新检测。
- 更新资格必须使用与安装/导入关联相同的逻辑来源键。
- 远端更新检测通过比较本地 `SourceSHA` 与同一 repo subpath 的最新远端 commit SHA。
- `LastCheckedAt` 应在每次完成检查后更新，而不是仅在发现更新时更新。
- 当最新检查确认本地 `SourceSHA` 已是最新时，应清空 `LatestSHA`。
- 更新徽标与“更新”动作必须来源于统一状态模型，而不是页面私有的启发式判断。

### 实现指导

- 跨模块的 skill 关联应由后端统一负责，并向前端暴露规范化后的状态。
- 前端页面不应再基于 `Name` 或 `Path` 独自判断“是不是同一个 skill”“是否已导入”或“是否已推送”。
- 后续若引入 catalog / aggregate 层，应将各模块中的不同表示统一归并到同一个逻辑 skill 记录下，并把已安装实例作为其子引用。

### AppConfig（`core/config/model.go`）

```go
type ToolConfig struct {
    Name     string   // 如 "claude-code"、"opencode"、"codex"、"gemini-cli"、"openclaw"
    ScanDirs []string // 扫描现有 Skill 的目录列表
    PushDir  string   // 默认推送目录
    Enabled  bool
    Custom   bool     // true 表示用户通过设置添加的自定义工具
}

type CloudConfig struct {
    Provider    string            // "aliyun"、"tencent"、"huawei"、"git"
    Enabled     bool
    BucketName  string
    RemotePath  string            // 如 "skillflow/"
    Credentials map[string]string // 各服务商特定的凭据
}

type ProxyConfig struct {
    Mode   ProxyMode // "none" | "system" | "manual"
    URL    string    // Mode 为 "manual" 时使用
}

type AppConfig struct {
    SkillsStorageDir     string        // 默认：~/.skillflow/skills
    DefaultCategory      string        // 默认："Default"
    LogLevel             string        // "debug" | "info" | "error"
    Tools                []ToolConfig
    Cloud                CloudConfig
    Proxy                ProxyConfig
    SkippedUpdateVersion string        // 用于抑制启动时更新提示的版本 tag
}
```

### StarredRepo（`core/git/model.go`）

```go
type StarredRepo struct {
    URL       string    // 用户提供的 git 仓库 URL
    Name      string    // 解析后的 "owner/repo"
    Source    string    // 规范化键值 "<host>/<path>"
    LocalDir  string    // 运行时绝对缓存目录；在 star_repos.json 中以 AppDataDir() 下的相对路径持久化
    LastSync  time.Time
    SyncError string
}

type StarSkill struct {
    Name     string
    Path     string   // 本地 Skill 目录的绝对路径
    SubPath  string   // 仓库内的相对路径
    RepoURL  string
    RepoName string
    Source   string
    Imported bool     // 是否已在"我的skills"中
}
```

---

## 启动流程

`main.go` → `app.startup()`：
1. 加载应用数据目录
2. 初始化 `config.Service`，加载配置
3. 使用已配置的 `SkillsStorageDir` 创建 `skill.Storage`
4. 调用 `registerAdapters()`（5 个内置工具 → `FilesystemAdapter`）
5. 调用 `registerProviders()`（阿里云、腾讯云、华为云）
6. 启动 `forwardEvents(ctx, hub)` goroutine — 订阅 Hub，通过 `runtime.EventsEmit` 发送事件
7. 启动 `checkUpdatesOnStartup()` goroutine — 扫描 GitHub 来源 Skill 的更新
8. 启动 `updateStarredReposOnStartup()` goroutine — 同步收藏仓库

---

## 主 App 结构体

`app.go`（`package main`）包含 `App` 结构体及所有导出方法：

```go
type App struct {
    ctx         context.Context
    hub         *notify.Hub           // 事件发布/订阅
    storage     *skill.Storage        // Skill 增删改查
    config      *config.Service       // 配置持久化
    starStorage *coregit.StarStorage  // 收藏仓库 JSON 持久化
    cacheDir    string                // ~/.skillflow/cache/
}
```

**主要导出方法（50+），均可从前端调用：**

| 类别 | 方法 |
|------|------|
| Skills | `ListSkills()`、`ListCategories()`、`DeleteSkill()`、`MoveSkillCategory()` |
| 导入 | `ScanGitHub()`、`InstallFromGitHub()`、`ImportLocal()` |
| 同步 | `GetEnabledTools()`、`ScanToolSkills()`、`PushToTools()`、`PullFromTool()` |
| 配置 | `GetConfig()`、`SaveConfig()`、`AddCustomTool()`、`RemoveCustomTool()` |
| 备份 | `BackupNow()`、`ListCloudFiles()`、`RestoreFromCloud()`、`ListCloudProviders()` |
| 更新 | `CheckUpdates()`、`UpdateSkill()`、`CheckAppUpdate()`、`CheckAppUpdateAndNotify()` |
| 收藏仓库 | `AddStarredRepo()`、`ListAllStarSkills()`、`ImportStarSkills()`、`UpdateAllStarredRepos()` |
| UI 辅助 | `OpenFolderDialog()`、`OpenPath()`、`OpenURL()` |

变更操作（删除、导入、推送、拉取）完成后，在云端备份已启用时会自动触发 `autoBackup()`。

---

## 事件系统

后端 → 前端的事件通过 `core/notify.Hub` 流转：
- 后端通过 `hub.Publish(notify.Event{Type: ..., Payload: ...})` 发布事件
- `forwardEvents()` goroutine 订阅 Hub，将 `Payload` 序列化为 JSON 后调用 `runtime.EventsEmit(ctx, eventType, jsonData)`
- 前端通过 `wailsjs/runtime/runtime` 中的 `EventsOn('backup.progress', handler)` 订阅

事件类型定义在 `core/notify/model.go`：

```go
const (
    EventBackupStarted         EventType = "backup.started"
    EventBackupProgress        EventType = "backup.progress"
    EventBackupCompleted       EventType = "backup.completed"
    EventBackupFailed          EventType = "backup.failed"
    EventSyncCompleted         EventType = "sync.completed"
    EventUpdateAvailable       EventType = "update.available"
    EventSkillConflict         EventType = "skill.conflict"
    EventStarSyncProgress      EventType = "star.sync.progress"
    EventStarSyncDone          EventType = "star.sync.done"
    EventAppUpdateAvailable    EventType = "app.update.available"
    EventAppUpdateDownloadDone EventType = "app.update.download.done"
    EventAppUpdateDownloadFail EventType = "app.update.download.fail"
)
```

Hub 使用大小为 32 的缓冲通道，对慢速订阅者采用丢弃最旧事件的策略。

---

## 工具适配器

5 个内置工具均使用 `core/sync` 中的 `FilesystemAdapter`。各工具默认推送目录：

| 工具 | 默认推送目录 |
|------|------------|
| `claude-code` | `~/.claude/skills` |
| `opencode` | `~/.config/opencode/skills` |
| `codex` | `~/.agents/skills` |
| `gemini-cli` | `~/.gemini/skills` |
| `openclaw` | `~/.openclaw/skills` |

**适配器行为：**
- `Pull()` — 递归扫描目录树中的 `skill.md` 文件，将每个文件作为 Skill 导入
- `Push()` — 将 Skill 目录平铺（无分类子目录）复制到目标目录

通过设置添加的自定义工具也使用 `FilesystemAdapter`，目录由用户指定。

---

## Installer 接口（`core/install`）

```go
type Installer interface {
    Type() string
    Scan(ctx context.Context, source InstallSource) ([]SkillCandidate, error)
    Install(ctx context.Context, source InstallSource, selected []SkillCandidate, category string) error
}
```

- `GitHubInstaller` — 通过 Contents API 扫描 GitHub 仓库，下载 Skill 目录，记录 commit SHA
- `LocalInstaller` — 从本地文件系统路径导入

---

## Cloud Provider 接口（`core/backup`）

```go
type CloudProvider interface {
    Name() string
    Init(credentials map[string]string) error
    Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(file string)) error
    Restore(ctx context.Context, bucket, remotePath, localDir string) error
    List(ctx context.Context, bucket, remotePath string) ([]RemoteFile, error)
    RequiredCredentials() []CredentialField
}
```

设置页会根据 `RequiredCredentials()` 自动渲染凭据输入字段。

---

## Git 包（`core/git`）

处理收藏仓库工作流：
- `CloneOrUpdate(ctx, repoURL, localDir, proxyURL)` — git clone 或 fetch+pull
- `ScanSkills(localDir, repoURL, repoName, source)` — 在克隆的仓库中查找 Skill 目录
- `GetSubPathSHA(ctx, repoDir, subPath)` — 获取某路径的最新 commit SHA
- `ParseRepoRef()`、`ParseRepoName()`、`RepoSource()` — URL 解析工具函数
- `StarStorage` — `[]StarredRepo` 的 JSON 持久化，存储于 `<AppDataDir>/star_repos.json`

---

## 前端结构

```
frontend/src/
  App.tsx              ← BrowserRouter + 侧边栏布局 + 路由定义
  pages/               ← 每个路由对应一个文件
    Dashboard.tsx      ← 我的skills 列表（分类、搜索、拖拽）
    SyncPush.tsx       ← 推送 Skills 到外部工具
    SyncPull.tsx       ← 从外部工具拉取 Skills
    StarredRepos.tsx   ← 浏览和导入收藏仓库的 Skills
    Backup.tsx         ← 云端备份管理
    Settings.tsx       ← 工具配置、云服务商、代理设置
  components/          ← 共享 UI 组件
    SkillCard.tsx      ← 单个 Skill 展示卡片
    SkillTooltip.tsx   ← 鼠标悬停时显示 Skill 元数据
    CategoryPanel.tsx  ← 分类侧边栏/筛选
    GitHubInstallDialog.tsx  ← GitHub 仓库扫描 UI
    ConflictDialog.tsx ← 同步时处理 Skill 名称冲突
    SyncSkillCard.tsx  ← 同步页的 Skill 卡片
    ContextMenu.tsx    ← 右键上下文菜单
  config/
    toolIcons.tsx      ← 工具名称 → 图标映射
  wailsjs/             ← 自动生成（请勿手动编辑）
    go/main/App.js     ← Go 方法绑定
    go/main/App.d.ts   ← TypeScript 类型声明
    runtime/runtime.js ← Wails runtime（EventsOn、EventsEmit 等）
```

前端直接调用 Go 方法：`import { ListSkills } from '../../wailsjs/go/main/App'`。Go 结构体字段名在 JSON 中为 PascalCase（如 `cfg.Tools`、`t.SkillsDir`、`cfg.Cloud.Enabled`）。

---

## 测试方法

测试使用 `httptest.NewServer` 模拟 GitHub API 调用。将 mock 服务器 URL 传给 `NewChecker(srv.URL)` 或 `NewGitHubInstaller(srv.URL)`。文件系统测试使用 `t.TempDir()`。

**各包测试覆盖情况：**

| 包 | 测试文件 | 说明 |
|----|---------|------|
| `core/skill` | `model_test.go`、`storage_test.go`、`validator_test.go` | 完整覆盖 |
| `core/config` | `service_test.go` | 完整覆盖 |
| `core/notify` | `hub_test.go` | 完整覆盖 |
| `core/install` | `github_test.go`、`local_test.go` | 模拟 GitHub API |
| `core/update` | `checker_test.go` | 模拟 GitHub API |
| `core/sync` | `filesystem_adapter_test.go` | TempDir 文件系统测试 |
| `core/git` | `client_test.go`、`scanner_test.go`、`storage_test.go` | TempDir + mock |
| `core/backup` | 无 | 需要真实云凭据 |
| `core/registry` | 无 | 薄封装，通过集成测试覆盖 |

---

## 扩展指南

### 新增云服务商

1. 创建 `core/backup/<name>.go`，实现 `backup.CloudProvider`
2. 在 `providers.go` 中注册：`registry.RegisterCloudProvider(NewXxxProvider())`
3. 设置页会根据 `RequiredCredentials()` 自动渲染凭据字段

### 新增工具适配器

如果工具使用标准的平铺 Skill 目录，只需在 `adapters.go` 的 `registerAdapters()` 中添加即可。如需自定义行为，实现 `toolsync.ToolAdapter` 并通过 `registry.RegisterAdapter()` 注册。

### 新增 App 方法（前端可调用）

1. 在 `app.go`（或根目录下的新 `package main` 文件）中为 `App` 结构体添加导出方法
2. 运行 `make generate`（或 `wails generate module`）更新 `frontend/wailsjs/go/main/App.{js,d.ts}`
3. 在前端导入并调用：`import { MyNewMethod } from '../../wailsjs/go/main/App'`

---

*最后更新：2026-03-06*
