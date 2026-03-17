# SkillFlow — Architecture & Developer Reference

> 🌐 [中文版](architecture_zh.md) | **English**

This document covers SkillFlow's internal architecture, repository layout, data models, and extension points for contributors. User-facing behavior belongs in **[docs/features.md](features.md)**.

---

## Overview

SkillFlow is a **Wails v2** desktop app built with **Go 1.23** and a **React 18 + TypeScript** frontend.

- The Go backend lives in `cmd/skillflow/`.
- The React frontend lives in `cmd/skillflow/frontend/`.
- Wails bindings connect them directly.
- There is **no REST API**.

Core stack:

- Backend: Go, Wails, Git, provider-specific cloud SDKs
- Frontend: React, TypeScript, React Router, Tailwind CSS, Radix UI, Lucide
- Build: `make`, Wails CLI, Vite

---

## Repository Layout

```text
/                              module root, no Go source files
  go.mod
  Makefile
  README.md
  README_zh.md
  docs/
  core/                        reusable packages, no package main
  cmd/
    skillflow/                 Wails desktop app, package main
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

Key repository rules:

- The root directory contains **no `.go` files**.
- `cmd/skillflow/*.go` stays **flat** because Wails requires the bound `package main` files to remain in one directory.
- New reusable backend code goes in `core/<name>/`, not in a subdirectory under `cmd/skillflow/`.
- Run `wails dev`, `wails build`, and `wails generate module` from `cmd/skillflow/`. From the repo root, prefer `make dev`, `make build`, and `make generate`.

---

## Runtime Lifecycle

### Wails entrypoint

`cmd/skillflow/main.go`:

- determines an internal process role (`helper` or `ui`) before starting the shell
- keeps the helper as the single-instance owner for tray/menu-bar control and UI relaunch/focus
- starts the Wails UI child with `//go:embed all:frontend/dist` and `HideWindowOnClose: false`
- binds the `App` instance directly to the frontend only inside the UI process

### Helper/UI split

- The **helper** process is a lightweight shell owner. It keeps the tray or menu-bar item alive, exposes a local control endpoint under `<AppDataDir>/runtime/`, and relaunches or focuses the UI process on demand.
- The **UI** process hosts Wails, React, and the frontend-callable `App` bindings. Closing the window exits this process, which also releases the embedded WebKit/WebView runtime.
- A repeated app launch no longer tries to focus an existing Wails window directly. Instead it forwards `show-ui` to the helper and exits.
- At this stage, backend `App` methods still execute in the UI process; the helper currently manages shell lifecycle only.

### App startup flow

`App.startup()` performs the core backend initialization:

1. Resolve `config.AppDataDir()`
2. Run the one-time startup cutover in `core/upgrade`
3. Load config via `core/config.Service`
4. Initialize the rotating file logger
5. Create skill storage, starred-repo storage, and the local derived view-state cache manager
6. Register agent adapters and cloud providers
7. Start backend event forwarding
8. Start the auto-sync timer for cloud backup

`App.domReady()` handles shell/UI setup:

1. Restore or compute the initial window size
2. Start the UI-side local control server used by the helper (`show`, `hide`, `quit`)
3. Schedule background startup tasks after a short delay
4. Stagger those tasks so the first interactive second does not fan out all remote checks at once
5. Maintain a frontend activity state machine that can trim the routed page tree after prolonged background or inactive time

Deferred background tasks currently include:

- Git backup startup pull (earliest)
- skill update check
- starred repo refresh
- app release update check

`App.beforeClose()` persists the current window size. `App.shutdown()` stops the UI control server.

---

## App Data Layout

By default, SkillFlow stores app data under `config.AppDataDir()`:

- macOS: `~/Library/Application Support/SkillFlow/`
- Windows: `%USERPROFILE%\\.skillflow\\`

```text
<AppDataDir>/
  skills/                 installed library
  meta/                   one JSON sidecar per installed skill
  prompts/                prompt library
  cache/
    <repo-cache>/         cloned starred repositories
    viewstate/            local-only derived UI snapshots
  runtime/
    helper-control.json   helper loopback control endpoint
    ui-control.json       UI loopback control endpoint
    helper.lock           local single-instance lock / coordination state
  logs/
    skillflow.log
    skillflow.log.1
  config.json             sync-safe shared settings
  config_local.json       local-only settings
  star_repos.json         starred repo metadata
```

Important storage rules:

- `config.json` contains only settings that are safe to sync across devices.
- `config_local.json` stores machine-specific paths, auto-push targets, launch-at-login state, proxy settings, window state, custom-agent path config, and sensitive cloud credentials.
- Synced files such as `meta/*.json` and `star_repos.json` persist local paths as **forward-slash relative paths** whenever the target is inside the synchronized root.
- If `SkillsStorageDir` is moved outside the default app data directory, the synchronized root becomes the shared parent of `skills/` and `meta/`.
- Logs are bounded to **two files** (`skillflow.log`, `skillflow.log.1`) at **1MB each**.
- `cache/viewstate/*.json` stores only rebuildable derived state such as installed-skill snapshots and agent-presence indexes. These files are local-only and must never be treated as synced truth.
- `runtime/*` stores helper/UI loopback endpoints, tokens, PIDs, and single-instance coordination state. These files are local-only, must never be synced, and are recreated automatically when needed.

---

## Package Responsibilities

| Path | Responsibility |
|------|----------------|
| `cmd/skillflow` | Wails entrypoint, frontend-callable `App` methods, tray/window/single-instance integration, adapter/provider registration |
| `core/applog` | Rotating file logger and log-level handling |
| `core/backup` | Backup snapshot logic, provider interfaces, Git provider, object-storage providers |
| `core/config` | Split shared/local config persistence, defaults, status-visibility normalization |
| `core/git` | Git clone/pull/push helpers, starred repo scanning, starred repo storage |
| `core/install` | GitHub and local install flows |
| `core/notify` | Buffered backend event hub and typed payloads |
| `core/pathutil` | Cross-platform path normalization and relative-path persistence helpers |
| `core/prompt` | Prompt library storage, scoped export, and prepared import preview/apply |
| `core/registry` | Global agent-adapter and cloud-provider registries |
| `core/skill` | Skill model, storage, validation, installed-skill indexing |
| `core/skillkey` | Stable logical-key derivation for git and content-based skills |
| `core/sync` | `AgentAdapter` interface and filesystem-based adapter implementation |
| `core/upgrade` | Startup terminology/schema cutover for persisted config files |
| `core/update` | Direct GitHub commit-check helper kept for tests and fallback-style utilities; installed skill update state now comes from local repo-cache SHA comparison |
| `core/viewstate` | Local-only derived snapshot cache, fingerprinting, and incremental agent-presence helpers |

---

## `cmd/skillflow/` File Organization

The Wails app package must remain flat, so responsibilities are grouped by file prefix:

| File group | Purpose |
|-----------|---------|
| `main.go`, `version.go` | entrypoint and build-time version |
| `app.go` | main `App` struct and most frontend-callable methods |
| `app_viewstate.go`, `app_perf.go` | local snapshot caching, fingerprints, and lightweight performance timing helpers |
| `app_prompt.go` | prompt CRUD, scoped export, and prompt import session orchestration |
| `app_update.go` | app release update check, download, apply, skip-version behavior |
| `app_log.go` | logger initialization and runtime/file log bridging |
| `app_restore.go`, `app_backup.go` | restore compensation, Git-backup helpers |
| `app_autostart.go`, `window_size.go`, `app_path.go` | OS integration helpers |
| `adapters.go`, `providers.go` | register `core/sync` adapters and `core/backup` providers |
| `events.go`, `push_conflict.go`, `skill_state.go` | frontend DTOs and aggregated card state |
| `tray_*.go`, `single_instance_*.go`, `window_*.go` | platform-specific shell behavior |

---

## Derived View-State Cache

To keep navigation and large-list rendering responsive, the backend now maintains a local-only derived cache under `cache/viewstate/`.

- `ListSkills()` is snapshot-first: when the installed-skill fingerprint matches, it returns the cached `InstalledSkillEntry[]` directly instead of rebuilding push presence immediately.
- The installed-skill fingerprint is based on sync-relevant skill metadata plus agent push-directory summaries, so switching away from a page and coming back will reload current truth when those dependencies changed.
- Agent push presence is rebuilt incrementally per agent using per-agent fingerprints rather than rescanning every configured `PushDir` on every request.

These caches are optimization artifacts only:

- they are rebuilt from authoritative disk state
- they are safe to delete
- they are not synced across devices

---

## Key Data Models

### Skill (`core/skill/model.go`)

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

Notes:

- `ID` is the installed-instance UUID.
- `Path` is a runtime absolute path; when persisted inside synced metadata it should be stored as a portable relative path whenever possible.
- `SourceURL + SourceSubPath` identify the logical git source for GitHub-installed skills.

### Prompt (`core/prompt/storage.go`)

```go
type PromptLink struct {
    Label string
    URL   string
}

type Prompt struct {
    Name        string
    Description string
    Category    string
    Path        string
    FilePath    string
    Content     string
    ImageURLs   []string
    WebLinks    []PromptLink
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

Notes:

- Prompt body lives in `prompts/<category>/<name>/system.md`.
- Prompt-card metadata such as description, related images, and web links lives in the sidecar `prompts/<category>/<name>/prompt.json`.
- `ImageURLs` is capped at 3 items and currently persists only `http` / `https` URLs.
- The editor accepts markdown-style web-link input, but persistence normalizes it into structured `PromptLink{Label, URL}` records.

### AppConfig (`core/config/model.go`)

```go
type AppConfig struct {
    SkillsStorageDir      string
    AutoPushAgents        []string
    LaunchAtLogin         bool
    DefaultCategory       string
    LogLevel              string
    RepoScanMaxDepth      int
    SkillStatusVisibility SkillStatusVisibilityConfig
    Agents                []AgentConfig
    Cloud                 CloudConfig
    CloudProfiles         map[string]CloudProviderConfig
    Proxy                 ProxyConfig
    SkippedUpdateVersion  string
}
```

Config split:

- **Shared / synced**: `DefaultCategory`, `LogLevel`, `RepoScanMaxDepth`, status visibility, built-in agent enabled state, active cloud provider state, non-sensitive cloud profile fields, skipped app version.
- **Local-only**: `SkillsStorageDir`, `AutoPushAgents`, `LaunchAtLogin`, agent paths, custom-agent definitions, proxy settings, window size, sensitive cloud credentials.

Relevant nested models:

```go
type AgentConfig struct {
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

### Starred repo models (`core/git/model.go`)

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

## Unified Skill Identity & State Model

This section is normative for any work touching skill cards, import/install/push/pull flows, starred repos, agent scans, or update badges.

### Identity layers

SkillFlow distinguishes two identities:

- **Instance identity** — `Skill.ID` identifies one installed copy in **My Skills**. Use it for delete, move, and installed-instance update operations.
- **Logical identity** — a stable cross-module identity used to answer whether Dashboard, Starred Repos, My Agents, Pull, and Push are referring to the same skill.

`Name` and absolute `Path` are display/location metadata only. They are **not** the primary cross-module key.

### Logical key rules

- **Git-backed skills** use `git:<repo-source>#<subpath>`
  - `repo-source`: canonical host/path such as `github.com/owner/repo`
  - `subpath`: forward-slash repo-relative path such as `skills/my-skill`
- **Non-git skills** should use a stable content-derived key such as `content:<hash>`
- Weak fallbacks are allowed only when no stable key can be derived yet

### Module mapping

| Module / page | Primary entity | Identity that drives behavior |
|---------------|----------------|-------------------------------|
| Dashboard / My Skills | installed `Skill` | `Skill.ID` for instance actions; logical key for cross-module correlation |
| Sync Push | installed `Skill` | `Skill.ID` for selection; logical key for pushed-state resolution |
| GitHub scan/install | remote candidate | logical key from repo source + subpath |
| Starred Repos | `StarSkill` | logical key from repo source + subpath |
| Agent Skills | agent-local candidate / aggregate | logical key for dedupe and status; path only for agent-local open/delete |
| Sync Pull | agent-local candidate | logical key for import and conflict detection |

### Unified status semantics

- **installed** — at least one installed My Skills entry exists for the logical key
- **imported** — wording alias for `installed` on external-source pages
- **pushed** — the logical skill exists in an agent's configured `PushDir`
- **seenInAgentScan** — the logical skill exists in an agent's configured `ScanDirs`; this does **not** imply SkillFlow pushed it
- **updatable** — at least one installed git-backed instance has a newer cached-repo SHA than its installed `SourceSHA`

### Status and dedupe rules

- `pushed` is narrower than "exists somewhere in the agent"; it specifically means present in the configured push target.
- `seenInAgentScan` is observational state. It must not be mislabeled as "already pushed".
- Cross-module dedupe prefers logical-key equality.
- Same-name items from different repos remain distinct when their logical keys differ.
- Name-only matching is a last-resort compatibility fallback and must never override a stronger logical-key match.

### Update rules

- Installed skill update checks apply only to git-backed skills with a stable repo source and subpath whose corresponding repo clone already exists under the local `cache/` tree.
- Cache lookup and installed-instance correlation must use the same logical git key.
- `CheckUpdates()` compares installed `SourceSHA` against the latest commit SHA for that same `SourceSubPath` inside the local cached repo clone; it does not call the GitHub Commits API directly.
- `UpdateSkill()` copies files from that cached repo subdirectory into the installed library directory, then refreshes any existing pushed agent copies from the updated installed instance.
- `LatestSHA` is cleared when a fresh check confirms the installed copy is already current.
- `LastCheckedAt` is updated on every completed check attempt and persisted in local-only `meta_local/<skill-id>.local.json` (not synced).

### Implementation guidance

- Backend code owns cross-module correlation and should return normalized status data to the frontend.
- Frontend pages should not infer "same skill", "already imported", or "already pushed" from `Name` or `Path` alone.
- `core/skillkey` derives logical keys.
- `core/skill.BuildInstalledIndex` correlates installed state across GitHub scan results, starred repo entries, and agent-scan entries.

---

## Events and Bindings

### Backend event flow

SkillFlow uses `core/notify.Hub` as a buffered event bus:

1. Backend code publishes `notify.Event`
2. `forwardEvents()` subscribes and forwards them through Wails `runtime.EventsEmit`
3. The frontend subscribes through `cmd/skillflow/frontend/wailsjs/runtime`

The hub currently uses a buffer of **32** entries with drop-oldest behavior for slow consumers.

Important event groups:

- backup: `backup.started`, `backup.progress`, `backup.completed`, `backup.failed`
- sync/update: `sync.completed`, `update.available`, `skill.conflict`
- starred repos: `star.sync.progress`, `star.sync.done`
- Git backup: `git.sync.started`, `git.sync.completed`, `git.sync.failed`, `git.conflict`
- app update: `app.update.available`, `app.update.download.done`, `app.update.download.fail`
- window lifecycle: `app.window.visibility.changed`

### Wails bindings

Generated bindings live under:

- `cmd/skillflow/frontend/wailsjs/go/main/App.js`
- `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`

Regenerate them with `make generate` after changing exported `App` methods.

---

## Frontend Structure

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

Key frontend areas:

- `src/pages/` — route-level screens such as Dashboard, Sync Push/Pull, Starred Repos, Backup, Settings, My Agents, and My Prompts
- `src/components/` — shared cards, dialogs, category panels, and list controls
- `src/contexts/` — language, theme, and status-visibility state
- `src/i18n/` — translation dictionaries
- `src/lib/` — shared list/search/clipboard/state helpers
- `tests/` — frontend unit tests run outside the Wails build

`App.tsx` also owns the app-activity state machine:

- backend hide/show signals and browser focus/visibility are merged into one foreground/background model
- after roughly 30 seconds in the background, the routed page subtree is unmounted to release page-local arrays and prompt content
- when the window becomes active again, the current route mounts from scratch and reloads fresh data

Frontend code imports backend methods from the generated Wails module, for example:

```ts
import { ListSkills } from '../../wailsjs/go/main/App'
```

---

## Logging and Path Portability

### Logging

- `core/applog.Logger` writes structured text logs to the app data `logs/` directory.
- `cmd/skillflow/app_log.go` mirrors enabled logs to the Wails runtime log APIs.
- Backend changes should log start/completion/failure for important mutations, sync flows, backup flows, Git operations, and external API calls.
- Secrets must never be logged.

### Path handling

- `core/pathutil` normalizes stored paths into forward-slash relative form.
- Runtime APIs may expand those paths back to absolute paths before returning them to callers.
- Portable path storage is required for synced files and restore flows.

---

## Testing and Build Workflow

- Run backend tests from the repo root: `go test ./core/...`
- Run Wails dev/build/generate from `cmd/skillflow/`, or use `make dev`, `make build`, `make generate`
- Frontend dependencies live in `cmd/skillflow/frontend/package.json`
- Frontend unit tests live in `cmd/skillflow/frontend/tests/`
- Production app output is written to `cmd/skillflow/build/bin/`

---

## Extension Guides

### Add a new frontend-callable App method

1. Add an exported method on `App` in `cmd/skillflow/app.go` or another flat `cmd/skillflow/*.go` file
2. Run `make generate`
3. Import it in the frontend from `../../wailsjs/go/main/App`

### Add a new cloud provider

1. Create `core/backup/<name>.go` implementing `backup.CloudProvider`
2. Register it in `cmd/skillflow/providers.go`
3. Expose its credential fields through `RequiredCredentials()`

### Add a new agent adapter

1. For standard filesystem-based agents, register it in `cmd/skillflow/adapters.go`
2. For custom behavior, implement `agentsync.AgentAdapter`
3. Always import `core/sync` as `agentsync` to avoid the stdlib `sync` name conflict

### Add a new reusable backend module

- Prefer `core/<name>/` for reusable logic
- Do not create Go subdirectories under `cmd/skillflow/`
- Keep Wails-specific shell code in `cmd/skillflow/` and reusable domain logic in `core/`

---

*Last updated: 2026-03-14*
