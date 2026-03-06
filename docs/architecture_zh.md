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
- **基于 UUID 的 Skill 标识** — Skill 以 UUID 标识，元数据存储在 JSON 附属文件中
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

Skill 以 UUID 标识。`meta/` 目录始终为 `filepath.Join(filepath.Dir(root), "meta")`。

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
    Path          string     // skill 目录的绝对路径
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
    LocalDir  string    // 磁盘上的缓存目录
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
    Imported bool     // 是否已在"我的 Skills"中
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
- `StarStorage` — `[]StarredRepo` 的 JSON 持久化，存储于 `~/.skillflow/star_repos.json`

---

## 前端结构

```
frontend/src/
  App.tsx              ← BrowserRouter + 侧边栏布局 + 路由定义
  pages/               ← 每个路由对应一个文件
    Dashboard.tsx      ← 我的 Skills 列表（分类、搜索、拖拽）
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
