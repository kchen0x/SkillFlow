# SkillFlow — Architecture & Developer Reference

> 🌐 [中文版](architecture_zh.md) | **English**

This document covers the internal architecture, package design, data models, and extension guides for contributors.

---

## Overview

SkillFlow is a **Wails v2** desktop app (Go 1.23, Wails v2.11.0). The Go backend exposes methods directly to the React frontend via Wails method bindings. There is **no REST API** — frontend calls Go methods as async functions.

**Tech stack:**
- Backend: Go 1.23, Wails v2
- Frontend: React 18, TypeScript, React Router v7, Tailwind CSS, Lucide React, Radix UI
- Build: Wails CLI, Vite

---

## Key Design Decisions

- **`core/sync` package name conflicts with Go stdlib `sync`** — always import it with alias: `toolsync "github.com/shinerio/skillflow/core/sync"`
- **Wails bindings are auto-generated** — after adding/removing exported methods on `App`, run `wails generate module` to update `frontend/wailsjs/go/main/App.{js,d.ts}`
- **`package main` files at root** — `app.go`, `adapters.go`, `providers.go`, `events.go` are all `package main` alongside `main.go` because Wails requires the app struct in the same package as `main`
- **No REST API** — direct Wails method bindings; faster and simpler
- **Installed skill instances are UUID-based, but cross-module identity must use a stable logical key** — see [Unified Skill Identity & State Model](#unified-skill-identity--state-model)
- **Filesystem adapters** — all built-in tools share the same `FilesystemAdapter` pattern
- **GitHub as source of truth** — update checker polls GitHub API, not local timestamps

---

## Data Storage Layout

```
~/.skillflow/
  skills/              ← SkillsStorageDir (configured)
    <category>/
      <skill-name>/    ← copied skill directory
        skill.md       ← main file with YAML frontmatter
        ...other files
  meta/                ← JSON sidecars (sibling of skills/)
    <uuid>.json        ← one per skill, contains Skill struct
  config.json          ← AppConfig (tools, cloud, proxy)
  star_repos.json      ← StarredRepo[] array
  cache/               ← temporary cloned repos for starred repos
    <cached-repo-dirs>/
```

Installed skill instances are identified by UUID. Cross-module correlation must follow the logical-key rules in [Unified Skill Identity & State Model](#unified-skill-identity--state-model). The `meta/` directory is always `filepath.Join(filepath.Dir(root), "meta")`.

---

## Backend Package Responsibilities

| Package | Responsibility |
|---------|---------------|
| `core/skill` | `Skill` model, `Storage` (CRUD + categories), `Validator` (skill.md check) |
| `core/config` | `AppConfig` model, `Service` (load/save JSON), `DefaultToolsDir()` per tool |
| `core/notify` | `Hub` (buffered channel pub/sub), `EventType` constants |
| `core/install` | `Installer` interface, `GitHubInstaller` (scan/download/SHA), `LocalInstaller` |
| `core/sync` | `ToolAdapter` interface, `FilesystemAdapter` (shared by all built-in tools) |
| `core/backup` | `CloudProvider` interface, Aliyun/Tencent/Huawei implementations |
| `core/update` | `Checker` (GitHub Commits API SHA comparison) |
| `core/registry` | Global maps for Installer/ToolAdapter/CloudProvider — registered at startup |
| `core/git` | Git clone/update, repo scanning for skills, starred repo storage |

---

## Key Data Models

### Skill (`core/skill/model.go`)

```go
type Skill struct {
    ID            string     // UUID
    Name          string     // skill name (dir name)
    Path          string     // absolute runtime path; persisted in meta/*.json as a relative path within the synced root
    Category      string     // user-defined category
    Source        SourceType // "github" | "manual"
    SourceURL     string     // GitHub repo URL for GitHub sources
    SourceSubPath string     // relative path within repo (e.g. "skills/my-skill")
    SourceSHA     string     // installed commit SHA (from GitHub)
    LatestSHA     string     // detected newer SHA (for update checking)
    InstalledAt   time.Time
    UpdatedAt     time.Time
    LastCheckedAt time.Time
}

const (
    SourceGitHub SourceType = "github"
    SourceManual SourceType = "manual"
)
```

## Unified Skill Identity & State Model

This section is normative for all future work that touches skill cards, import/install/push/pull flows, starred repos, tool scans, or update badges.

### Identity layers

SkillFlow must distinguish between two different identities:

- **Instance identity** — `Skill.ID` identifies one installed copy in **My Skills**. This is the correct key for CRUD operations on installed items such as delete, move category, rename category membership, and manual update.
- **Logical identity** — a stable cross-module identity that answers: “is this the same skill shown on Dashboard, Starred Repos, Tool Skills, Sync Pull, and Sync Push?”

`Name` and absolute `Path` are display or location metadata only. They must not be treated as the primary cross-module identity.

### Logical key rules

- **Git-backed skills** must use a logical key derived from normalized repository source plus repository subpath:
  - format: `git:<repo-source>#<subpath>`
  - `repo-source` is the canonical host/path form such as `github.com/owner/repo`
  - `<subpath>` is the forward-slash relative path inside the repo such as `skills/my-skill`
- **Non-git skills** should use a stable content-derived key, e.g. `content:<hash>`, so the same skill can still be recognized across tool scans and local imports.
- **Temporary fallback heuristics** are allowed only when no stable logical key can be derived yet, and those heuristics must be treated as weak matches.

### Module mapping

| Module / page | Primary entity | Identity that drives behavior |
|---------------|----------------|-------------------------------|
| Dashboard / My Skills | installed `Skill` | `Skill.ID` for instance actions; logical key for cross-module correlation |
| Sync Push | installed `Skill` | `Skill.ID` for selection; logical key for pushed-state resolution |
| GitHub scan/install | remote candidate | logical key derived from repo source + subpath |
| Starred Repos | `StarSkill` | logical key derived from repo source + subpath |
| Tool Skills | tool-local candidate / aggregate | logical key for dedupe and status; path only for open/delete within that tool |
| Sync Pull | tool-local candidate | logical key for import/conflict detection |

### Unified status semantics

- **installed** — at least one installed instance exists in My Skills for the logical key.
- **imported** — external-source wording for `installed`; when viewing GitHub, Starred Repos, or Tool scans, “imported” means “already installed into My Skills”.
- **pushed** — the logical skill is present in a tool’s configured `PushDir`. This means SkillFlow can regard it as already pushed to that tool.
- **seenInToolScan** — the logical skill is detected in one of a tool’s configured `ScanDirs`. This means the skill is visible in the tool ecosystem, but not necessarily managed by SkillFlow or present in `PushDir`.
- **updatable** — at least one installed git-backed instance has a newer remote commit than its installed `SourceSHA`.

### Status rules

- `pushed` is narrower than “exists somewhere in the tool”; it refers specifically to the configured push target.
- `seenInToolScan` is observational. It helps distinguish “the tool already has this skill” from “SkillFlow already pushed this skill”.
- A skill may have both `pushed=true` and `seenInToolScan=true` when the push directory is also scanned, or when the same logical skill exists in both places.
- A skill with `seenInToolScan=true` and `pushed=false` should generally be shown as “already detected in tool”, not “already pushed”.

### Conflict and dedupe rules

- Cross-module dedupe must prefer logical key equality.
- Same-name items from different repos must be treated as different skills when their logical keys differ.
- Same-path items are not automatically the same skill unless they resolve to the same logical key.
- Name-only matching is acceptable only as a last-resort compatibility fallback and must never overwrite a stronger logical-key match.

### Update detection rules

- Only git-backed installed skills with a stable repo source and subpath participate in remote update checks.
- Update eligibility must be keyed by the same logical source used for install/import correlation.
- Remote update detection compares the installed `SourceSHA` with the newest remote commit SHA for the same repo subpath.
- `LastCheckedAt` should be updated on every completed check attempt, not only when an update is found.
- `LatestSHA` should be cleared when a fresh check confirms that the installed `SourceSHA` is already current.
- Update badges and “Update” actions must be derived from the unified status model rather than page-local heuristics.

### Implementation guidance

- The backend should own cross-module skill correlation and expose normalized statuses to the frontend.
- Frontend pages should not independently decide “same skill”, “already imported”, or “already pushed” from `Name` or `Path` alone.
- Any future catalog / aggregate layer should group all module-specific representations under one logical skill record and keep installed instances as child references.

### AppConfig (`core/config/model.go`)

```go
type ToolConfig struct {
    Name     string   // e.g. "claude-code", "opencode", "codex", "gemini-cli", "openclaw"
    ScanDirs []string // directories to scan for existing skills
    PushDir  string   // default directory to push skills to
    Enabled  bool
    Custom   bool     // true if user-added via Settings
}

type CloudConfig struct {
    Provider    string            // "aliyun", "tencent", "huawei", "git"
    Enabled     bool
    BucketName  string
    RemotePath  string            // e.g. "skillflow/"
    Credentials map[string]string // provider-specific credentials
}

type ProxyConfig struct {
    Mode   ProxyMode // "none" | "system" | "manual"
    URL    string    // used when Mode == "manual"
}

type AppConfig struct {
    SkillsStorageDir     string        // default: ~/.skillflow/skills
    DefaultCategory      string        // default: "Default"
    LogLevel             string        // "debug" | "info" | "error"
    Tools                []ToolConfig
    Cloud                CloudConfig
    Proxy                ProxyConfig
    SkippedUpdateVersion string        // version tag to suppress startup update prompt
}
```

### StarredRepo (`core/git/model.go`)

```go
type StarredRepo struct {
    URL       string    // user-provided git repo URL
    Name      string    // parsed "owner/repo"
    Source    string    // canonical key "<host>/<path>"
    LocalDir  string    // absolute runtime cache dir; persisted in star_repos.json as a relative path under AppDataDir()
    LastSync  time.Time
    SyncError string
}

type StarSkill struct {
    Name     string
    Path     string   // absolute local path to skill dir
    SubPath  string   // relative path in repo
    RepoURL  string
    RepoName string
    Source   string
    Imported bool     // already in My Skills?
}
```

---

## Startup Flow

`main.go` → `app.startup()`:
1. Load app data directory
2. Initialize `config.Service`, load config
3. Create `skill.Storage` with configured `SkillsStorageDir`
4. Call `registerAdapters()` (5 built-in tools → `FilesystemAdapter`)
5. Call `registerProviders()` (Aliyun, Tencent, Huawei)
6. Start `forwardEvents(ctx, hub)` goroutine — subscribes to Hub, emits each event via `runtime.EventsEmit`
7. Start `checkUpdatesOnStartup()` goroutine — scan skills for GitHub updates
8. Start `updateStarredReposOnStartup()` goroutine — sync starred repos

---

## Main App Struct

`app.go` (`package main`) contains the `App` struct and all exported methods:

```go
type App struct {
    ctx         context.Context
    hub         *notify.Hub           // event pub/sub
    storage     *skill.Storage        // skill CRUD
    config      *config.Service       // config persistence
    starStorage *coregit.StarStorage  // starred repos JSON persistence
    cacheDir    string                // ~/.skillflow/cache/
}
```

**Key exported methods (50+) — all callable from frontend:**

| Category | Methods |
|----------|---------|
| Skills | `ListSkills()`, `ListCategories()`, `DeleteSkill()`, `MoveSkillCategory()` |
| Import | `ScanGitHub()`, `InstallFromGitHub()`, `ImportLocal()` |
| Sync | `GetEnabledTools()`, `ScanToolSkills()`, `PushToTools()`, `PullFromTool()` |
| Config | `GetConfig()`, `SaveConfig()`, `AddCustomTool()`, `RemoveCustomTool()` |
| Backup | `BackupNow()`, `ListCloudFiles()`, `RestoreFromCloud()`, `ListCloudProviders()` |
| Updates | `CheckUpdates()`, `UpdateSkill()`, `CheckAppUpdate()`, `CheckAppUpdateAndNotify()` |
| Starred repos | `AddStarredRepo()`, `ListAllStarSkills()`, `ImportStarSkills()`, `UpdateAllStarredRepos()` |
| UI helpers | `OpenFolderDialog()`, `OpenPath()`, `OpenURL()` |

Auto-backup (`autoBackup()`) is triggered after mutations (delete, import, push, pull) when cloud backup is enabled.

---

## Event System

Backend → Frontend events flow through `core/notify.Hub`:
- Backend publishes via `hub.Publish(notify.Event{Type: ..., Payload: ...})`
- `forwardEvents()` goroutine subscribes to Hub, marshals `Payload` to JSON, and calls `runtime.EventsEmit(ctx, eventType, jsonData)`
- Frontend subscribes via `EventsOn('backup.progress', handler)` from `wailsjs/runtime/runtime`

Event types are defined in `core/notify/model.go`:

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

The Hub uses a buffered channel (size 32) with drop-oldest behavior for slow subscribers.

---

## Tool Adapters

All 5 built-in tools use `FilesystemAdapter` from `core/sync`. Default push directories per tool:

| Tool | Default Push Directory |
|------|----------------------|
| `claude-code` | `~/.claude/skills` |
| `opencode` | `~/.config/opencode/skills` |
| `codex` | `~/.agents/skills` |
| `gemini-cli` | `~/.gemini/skills` |
| `openclaw` | `~/.openclaw/skills` |

**Adapter behavior:**
- `Pull()` — recursively scan directory tree for `skill.md` files, import each as a skill
- `Push()` — copy skill directories flat (no category subdir) into the target directory

Custom tools added via Settings also use `FilesystemAdapter` with user-provided directory.

---

## Installer Interface (`core/install`)

```go
type Installer interface {
    Type() string
    Scan(ctx context.Context, source InstallSource) ([]SkillCandidate, error)
    Install(ctx context.Context, source InstallSource, selected []SkillCandidate, category string) error
}
```

- `GitHubInstaller` — scans GitHub repos via Contents API, downloads skill directories, records commit SHA
- `LocalInstaller` — imports from local filesystem path

---

## Cloud Provider Interface (`core/backup`)

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

The Settings page automatically renders credential input fields from `RequiredCredentials()`.

---

## Git Package (`core/git`)

Handles starred repo workflows:
- `CloneOrUpdate(ctx, repoURL, localDir, proxyURL)` — git clone or fetch+pull
- `ScanSkills(localDir, repoURL, repoName, source)` — find skill dirs in cloned repo
- `GetSubPathSHA(ctx, repoDir, subPath)` — get latest commit SHA for a path
- `ParseRepoRef()`, `ParseRepoName()`, `RepoSource()` — URL parsing utilities
- `StarStorage` — JSON persistence for `[]StarredRepo` at `<AppDataDir>/star_repos.json`

---

## Frontend Structure

```
frontend/src/
  App.tsx              ← BrowserRouter + sidebar layout + route definitions
  pages/               ← one file per route
    Dashboard.tsx      ← My Skills listing (categories, search, drag-drop)
    SyncPush.tsx       ← Push skills to external tools
    SyncPull.tsx       ← Pull skills from external tools
    StarredRepos.tsx   ← Browse and import from starred/watched repos
    Backup.tsx         ← Cloud backup management
    Settings.tsx       ← Tool config, cloud provider, proxy settings
  components/          ← shared UI components
    SkillCard.tsx      ← Individual skill display card
    SkillTooltip.tsx   ← Hover tooltips showing skill metadata
    CategoryPanel.tsx  ← Category sidebar/filter
    GitHubInstallDialog.tsx  ← GitHub repo scanner UI
    ConflictDialog.tsx ← Handle skill name conflicts on sync
    SyncSkillCard.tsx  ← Skill card for sync pages
    ContextMenu.tsx    ← Right-click context menus
  config/
    toolIcons.tsx      ← Tool name → icon mapping
  wailsjs/             ← auto-generated (do not edit manually)
    go/main/App.js     ← Go method bindings
    go/main/App.d.ts   ← TypeScript type declarations
    runtime/runtime.js ← Wails runtime (EventsOn, EventsEmit, etc.)
```

Frontend calls Go methods directly: `import { ListSkills } from '../../wailsjs/go/main/App'`. Go struct field names are PascalCase in JSON (e.g. `cfg.Tools`, `t.SkillsDir`, `cfg.Cloud.Enabled`).

---

## Testing Approach

Tests use `httptest.NewServer` to mock GitHub API calls. Pass the mock server URL to `NewChecker(srv.URL)` or `NewGitHubInstaller(srv.URL)`. Filesystem tests use `t.TempDir()`.

**Test coverage by package:**

| Package | Test files | Notes |
|---------|-----------|-------|
| `core/skill` | `model_test.go`, `storage_test.go`, `validator_test.go` | Full coverage |
| `core/config` | `service_test.go` | Full coverage |
| `core/notify` | `hub_test.go` | Full coverage |
| `core/install` | `github_test.go`, `local_test.go` | Mocked GitHub API |
| `core/update` | `checker_test.go` | Mocked GitHub API |
| `core/sync` | `filesystem_adapter_test.go` | TempDir filesystem tests |
| `core/git` | `client_test.go`, `scanner_test.go`, `storage_test.go` | TempDir + mock |
| `core/backup` | none | Requires real cloud credentials |
| `core/registry` | none | Thin wrapper, tested via integration |

---

## Extension Guides

### Adding a New Cloud Provider

1. Create `core/backup/<name>.go` implementing `backup.CloudProvider`
2. Register in `providers.go`: `registry.RegisterCloudProvider(NewXxxProvider())`
3. The Settings page automatically renders credential fields from `RequiredCredentials()`

### Adding a New Tool Adapter

If the tool uses a flat directory of skills (standard), just add it to `registerAdapters()` in `adapters.go`. For custom behavior, implement `toolsync.ToolAdapter` and register via `registry.RegisterAdapter()`.

### Adding a New App Method (Frontend-callable)

1. Add exported method to `App` struct in `app.go` (or a new `package main` file at root)
2. Run `make generate` (or `wails generate module`) to update `frontend/wailsjs/go/main/App.{js,d.ts}`
3. Import and call from frontend: `import { MyNewMethod } from '../../wailsjs/go/main/App'`

---

*Last updated: 2026-03-06*
