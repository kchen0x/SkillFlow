# GitHub 收藏仓库功能设计

Date: 2026-03-05

## 背景

现有"从 GitHub 安装"功能基于 GitHub Contents API 逐文件下载，存在以下问题：
- 每次扫描都需要网络请求，速度慢
- 无法管理多个常用 GitHub 仓库
- 没有集中的收藏/订阅入口

## 目标

1. 所有 GitHub 操作统一改为 git clone/pull，本地缓存复用
2. 新增"GitHub 收藏"侧边栏入口，管理收藏的 GitHub 仓库
3. 收藏仓库启动时自动后台更新，支持手动刷新
4. 支持从收藏仓库导入 Skills（单个 + 批量）

## 存储布局

```
~/.skillflow/                         (AppDataDir)
  cache/                              ← 新增，git clone 缓存
    <owner>/
      <repo>/                         ← 每个仓库一个目录
  star_repos.json                     ← 新增，收藏列表
  skills/                             ← 不变
  meta/                               ← 不变
  config.json                         ← 不变
```

`cache/` 目录不参与云同步。

## Git 操作策略

**检查 git：** 调用 git 前先 `git --version`，失败则返回提示性错误："请先安装 git 后再使用此功能"。

**Clone/Update（force-push 安全）：**
- 目录不存在：`git clone <url> <dir>`
- 目录已存在：`git fetch origin && git reset --hard origin/HEAD`

本地扫描 Skills：遍历目录下含 `SKILLS.md` 的子目录。

## 数据模型

```go
// core/git/model.go
type StarredRepo struct {
    URL       string    `json:"url"`
    Name      string    `json:"name"`     // "owner/repo"
    LocalDir  string    `json:"localDir"` // 绝对路径
    LastSync  time.Time `json:"lastSync"`
    SyncError string    `json:"syncError,omitempty"`
}

type StarSkill struct {
    Name     string `json:"name"`
    Path     string `json:"path"`      // 本地绝对路径
    RepoURL  string `json:"repoUrl"`
    RepoName string `json:"repoName"`  // "owner/repo"
    Imported bool   `json:"imported"`  // 是否已在我的 Skills 中
}
```

## Backend

### 新增 `core/git` 包

```
core/git/
  client.go      ← CheckGitInstalled(), CloneOrUpdate(url, dir)
  scanner.go     ← ScanSkills(dir) []StarSkill
  storage.go     ← StarStorage: 读写 star_repos.json
  model.go       ← StarredRepo, StarSkill
```

### App 方法变更

**现有方法改造：**

| 方法 | 原实现 | 新实现 |
|---|---|---|
| `ScanGitHub(url)` | GitHub API 逐文件请求 | `CloneOrUpdate` → 本地扫描 |
| `InstallFromGitHub(url, candidates, cat)` | 再次 API 下载文件 | 直接从 cache 目录 `os.CopyFS` 复制 |

**新增方法：**

```go
AddStarredRepo(url string) error
RemoveStarredRepo(url string) error
ListStarredRepos() ([]StarredRepo, error)
ListAllStarSkills() ([]StarSkill, error)           // 平铺视图
ListRepoStarSkills(url string) ([]StarSkill, error) // 文件夹视图（进入某 repo）
UpdateAllStarredRepos() error                       // 手动刷新全部
UpdateStarredRepo(url string) error                 // 单个仓库刷新
ImportStarSkills(skillPaths []string, repoURL, category string) error // 批量导入
```

### Startup 自动更新

`app.startup()` 新增：
```go
go a.updateStarredReposOnStartup()
```

后台异步拉取所有收藏仓库，发布事件通知前端进度。

### 新增事件

```go
EventStarSyncProgress EventType = "star.sync.progress"  // 单个仓库同步完成
EventStarSyncDone     EventType = "star.sync.done"       // 全部同步完成
```

## Frontend

### 侧边栏

```
我的 Skills
── 同步管理
   推送到工具
   从工具拉取
GitHub 收藏          ← 新增（Star 图标）
云备份
设置
```

### 新页面 `StarredRepos.tsx`

**工具栏：**
- 添加仓库按钮（输入 GitHub URL，触发 clone）
- 全部更新按钮（触发所有仓库 git pull）
- 视图切换：文件夹 ∙ 平铺
- 批量导入按钮（平铺/文件夹内 skill 列表中出现）

**文件夹视图（默认）：**
- 每个 repo 一张卡片：名称、最后同步时间、同步状态（成功/失败/同步中）
- 卡片操作：单独更新、删除收藏、进入查看 Skills
- 进入 repo 后显示该 repo 所有 Skills，支持批量选择 + 批量导入

**平铺视图：**
- 所有收藏仓库的 Skills 合并显示为网格
- 每个 skill 卡片显示来源 repo 徽标（防同名混淆）
- 已导入的 skill 显示"已导入"标签
- 支持批量选择 + 批量导入（弹出分类选择对话框）

**`GitHubInstallDialog` 改造：**
- 扫描阶段显示"正在 clone/更新仓库..."提示
- 安装直接从 cache 复制，无需重新下载

## 错误处理

- git 未安装：提示"请先安装 git（https://git-scm.com）"
- clone/pull 失败：在 repo 卡片上显示错误信息，不影响其他仓库
- 导入冲突：复用现有 `ConflictDialog` 逻辑

## 不在范围内

- 私有仓库（需要 git 凭证管理，后续考虑）
- cache/ 目录大小管理（后续考虑）
