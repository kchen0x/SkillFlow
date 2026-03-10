# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Directory Organization Rule — MANDATORY

The root directory must contain **no Go source files**. All code lives in clearly scoped subdirectories:

```
/ (project root — no .go files here)
  go.mod, go.sum         ← module definition (must stay at root)
  Makefile               ← build orchestration
  README.md, README_zh.md
  LICENSE, .gitignore, .github/
  docs/                  ← all documentation
  core/                  ← reusable internal packages (no package main)
    skill/               ← Skill model, Storage, Validator
    config/              ← AppConfig model, Service
    notify/              ← event Hub
    install/             ← Installer interface + implementations
    sync/                ← ToolAdapter interface + FilesystemAdapter
    backup/              ← CloudProvider interface + implementations
    update/              ← update Checker
    registry/            ← global adapter/provider maps
    git/                 ← git operations, starred repo storage
  cmd/
    skillflow/           ← package main (Wails desktop app)
      main.go            ← entry point + //go:embed all:frontend/dist
      app.go, app_update.go, app_log.go
      adapters.go, providers.go, events.go, version.go
      tray_darwin.go, tray_windows.go, tray_stub.go
      single_instance_other.go, single_instance_windows.go
      wails.json         ← Wails project config (must be co-located with frontend/)
      build/             ← Wails build assets + binary output
        darwin/          ← macOS resources (iconfile.icns, Info.plist)
        windows/         ← Windows resources (icon.ico, installer/)
        appicon.png
        bin/             ← compiled output (git-ignored)
      frontend/          ← React/TypeScript app
        src/             ← source code
        dist/            ← built output (git-ignored, embedded by Go)
        package.json, vite.config.ts, tsconfig.json
```

**Rules:**
- Never add `.go` files to the project root. New application code goes in `cmd/skillflow/`; new reusable packages go in `core/<name>/`.
- `wails.json` must be co-located with `frontend/` inside `cmd/skillflow/`. All `wails dev/build/generate` commands must be run from `cmd/skillflow/` (the Makefile handles this).
- The `//go:embed all:frontend/dist` directive in `main.go` works because both are in `cmd/skillflow/`.
- `go test ./core/...` is run from the module root (where `go.mod` is).
- Import paths use the full module path: `github.com/shinerio/skillflow/core/...` (no change from before).
- **`cmd/skillflow/*.go` files must remain flat (no subdirectories).** Go requires all files in a package to be in the same directory; since Wails binds to `package main`, splitting into subdirectories is not possible. Use file-name prefixes as the organization convention:
  - `app.go`, `app_log.go`, `app_update.go` — App struct and method groups
  - `events.go` — event type definitions and emitters
  - `adapters.go`, `providers.go` — registration of `core/` implementations
  - `tray_darwin.go`, `tray_windows.go`, `tray_stub.go` — platform-specific system tray
  - `single_instance_other.go`, `single_instance_windows.go` — platform-specific single-instance lock
  - `version.go` — build-time version constant; `main.go` — entry point
- When a concern grows large enough to warrant its own package, extract it to `core/<name>/` (reusable, no Wails dependency) rather than creating a subdirectory inside `cmd/skillflow/`.

## Documentation Organization Rule — MANDATORY

**Root directory must contain only `README.md` and `README_zh.md` as documentation files.**

All other documentation lives under `docs/`:

| File | Purpose |
|------|---------|
| `docs/features.md` | Complete UI/UX feature reference in English |
| `docs/features_zh.md` | Complete UI/UX feature reference in Chinese |
| `docs/architecture.md` | Internal architecture, packages, data models, extension guides (English) |
| `docs/architecture_zh.md` | Same in Chinese |
| `docs/config.md` | Persisted config and metadata file reference (English) |
| `docs/config_zh.md` | Same in Chinese |
| `docs/plans/` | Design and implementation plans |
| `docs/skill_directory.md` | Skill directory format specification |

**Rules:**
- `README.md` / `README_zh.md` — user-facing only: features overview, download/install links, skill format, cloud backup config, contributing/build instructions. **No internal code snippets, no package tables, no architecture diagrams.**
- Never add new standalone `.md` files to the root directory. If you need new documentation, put it under `docs/`.

## Documentation Sync Rule — MANDATORY

**Any time a meaningful user-facing feature is added, changed, or removed, you MUST update the following files in the same commit:**

| File | What to update |
|------|---------------|
| `docs/features.md` | Add / edit / remove the corresponding section(s) in English. Update the "Last updated" date at the bottom. |
| `docs/features_zh.md` | Same changes in Chinese. Update the "最后更新" date at the bottom. |
| `README.md` | Update the Features table row(s) if the high-level description changes. |
| `README_zh.md` | Same in Chinese. |

**Rules:**
- A "feature change" means a clear user-facing capability or usage-flow change: for example a new UI control that enables a new action, a behavior change, removal of a control, a new backend method callable from the frontend, or a new event type.
- Pure presentation-only adjustments are **not** feature changes for this rule. Do **not** update `docs/features.md`, `docs/features_zh.md`, `README.md`, or `README_zh.md` for small copy tweaks, label renames, icon swaps, visual reordering, spacing/layout polish, or other changes that do not obviously change how the feature is used.
- Do **not** leave the docs stale. Never commit a feature change without the corresponding doc update in the same commit.
- `docs/features.md` / `docs/features_zh.md` are the source of truth for UX details, but only for meaningful UX/feature behavior. Do not churn them for low-signal presentation-only edits unless the user explicitly asks.
- `README.md` / `README_zh.md` must stay high-level and only describe coarse-grained product capabilities. Avoid adding small UI wording/layout details that increase reading burden.
- If a frontend/backend feature change also changes architecture, module boundaries, cross-module contracts, core data flow, persistence model, or extension points, update `docs/architecture.md` in the same commit. When the Chinese architecture doc is being maintained in parallel, update `docs/architecture_zh.md` as well.
- If a change needs detailed module-level design or implementation documentation, create or update files under a coarse-grained module folder in `docs/` such as `docs/<module>/...` instead of adding more root-level markdown files. Keep module categories broad and stable; do not create many small folders for individual pages, components, or one-off tweaks.

## Configuration Documentation Sync Rule — MANDATORY

**Any time a repo-tracked persisted config or metadata file is added, removed, renamed, or its on-disk schema/semantics change, you MUST update the following files in the same commit:**

| File | What to update |
|------|---------------|
| `docs/config.md` | Update the English examples, key tables, file-name notes, path/sync-scope notes, and any added/removed/renamed fields. |
| `docs/config_zh.md` | Same changes in Chinese. |

**Rules:**
- This rule applies to persisted config and metadata files such as `config.json`, `config_local.json`, `star_repos.json`, `meta/*.json`, and future repo-tracked files that document on-disk configuration or sync metadata.
- Update the docs when a field is added, removed, renamed, split, merged, moved between files, changes meaning, changes persistence scope, or when the canonical file name/path changes.
- Keep examples aligned with the actual persisted format on disk, not just the in-memory runtime model.
- Pure formatting-only changes that do not change persisted keys, file names, storage location, or semantics do not require doc churn.
- Do not merge config/schema changes while `docs/config.md` and `docs/config_zh.md` are stale.

## Path Persistence Rule — MANDATORY

Any repo-tracked file that can be backed up or synced across devices must avoid machine-specific absolute paths.

- Synced files such as `config.json`, `meta/*.json`, `star_repos.json`, and future backup/sync data files must store local filesystem paths as **forward-slash relative paths** whenever the target is inside the synchronized root.
- The synchronized root is normally `config.AppDataDir()`. When `SkillsStorageDir` is moved outside that directory, treat the shared parent of `skills/` and `meta/` as the synchronized root for persisted skill metadata.
- Any path that points **outside** the synchronized root is platform-specific and must live only in `config_local.json`.
- `config_local.json` is local-only and must remain excluded from cloud backup and git sync.
- Runtime APIs may expand persisted relative paths back to absolute paths before returning them to frontend/backend callers, but the on-disk synced representation must stay relative.

## Logging Rule — MANDATORY

All backend code changes must follow consistent logging standards for troubleshooting.

### Log level policy

- `error`:
  - Required for any failed operation, exception, unexpected branch, external dependency failure.
- `info`:
  - Required for important flow milestones (`start` / `completed`) of key operations.
- `debug`:
  - For detailed diagnostics and branch-level context, must be suppressible by configured log level.

### Key operations that MUST log

The following operations must have reasonable logs (at minimum `info` on start/success, `error` on failure):

- Git operations:
  - clone, fetch, pull, push, conflict detection/resolution, reset/force update.
- API operations:
  - external API calls (GitHub / cloud providers / remote services), especially failures.
- Sync operations:
  - scan, import, update, push, pull, backup, restore.
- Resource mutations (state-changing operations):
  - create / delete / rename / move / overwrite.
- Config mutations:
  - settings save, log-level changes, provider/tool config updates.

### Message quality requirements

- Log message should include:
  - operation name
  - target/resource identifier (skill id/name, repo url/name, tool/provider, path, etc.)
  - result status (`started` / `completed` / `failed`)
  - failure reason for `error` logs
- Keep wording stable and searchable across the same operation.
- Avoid noisy/duplicated logs and avoid logging every trivial getter.

### Security requirements

- Never log secrets or sensitive data:
  - access token, password, secret key, credential raw content, authorization header, cookie.
- If needed for diagnosis, log only masked or non-sensitive metadata.

### Rotation and file-size rule

- Log file strategy must remain bounded:
  - keep only 2 files (`skillflow.log`, `skillflow.log.1`)
  - max 1MB per file
  - rotate and overwrite oldest when size limit is reached

## Python Tooling Rule — MANDATORY

Any Python-related work in this repository must use `uv` for interpreter management, dependency management, and script execution, while preserving functional correctness.

**Rules:**
- Never invoke system `python`, `python3`, `pip`, or `pip3` directly. If the interpreter itself is needed, run it via `uv run python ...`.
- Never install Python packages into the system environment.
- Prefer `uv run` for repo scripts, inline scripts, and module execution.
- When converting an existing Python command to `uv`, preserve the original entrypoint, arguments, working directory, environment variables, stdin/stdout behavior, and required Python version.
- Prefer `uv run python path/to/script.py` for script files, `uv run python - <<'PY'` for inline snippets, and `uv run -m <module>` for module-style commands.
- Use `uv add`, `uv remove`, and `uv sync` to manage project dependencies.
- Use `uv run --with <package>` or `uvx <tool>` for temporary or one-off tools instead of touching the system environment.
- If a documented or scripted Python workflow currently uses direct Python or pip invocation, convert it to the equivalent `uv` workflow before running it.
- Do not trade correctness for tooling purity: choose the `uv` invocation that faithfully reproduces the intended behavior.

## Commands

### Make targets (recommended)

```bash
make dev              # Run in dev mode (hot-reload for Go + frontend)
make build            # Build production binary
make test             # Run all Go tests
make tidy             # Sync Go module dependencies
make generate         # Regenerate TypeScript bindings after App method changes
make install-frontend # Install frontend npm dependencies
make clean            # Remove build artifacts
make help             # List all targets
```

### Development (manual)

```bash
# Run the app in dev mode (hot-reload for both Go and frontend)
cd cmd/skillflow && ~/go/bin/wails dev

# Build production binary
cd cmd/skillflow && ~/go/bin/wails build

# Regenerate TypeScript bindings after changing App struct methods
cd cmd/skillflow && ~/go/bin/wails generate module
```

### Go (backend)

```bash
# Run all tests (from project root)
go test ./core/...

# Run tests for a single package
go test ./core/skill/...
go test ./core/update/...
go test ./core/git/...

# Run a single test function
go test ./core/skill/... -run TestSkillHasUpdate

# Sync dependencies after modifying go.mod
go mod tidy
```

### Frontend

```bash
cd cmd/skillflow/frontend
npm install        # install dependencies
npm run dev        # Vite dev server (used by wails dev)
npm run build      # production build (output: cmd/skillflow/frontend/dist/)
```

## Architecture

SkillFlow is a Wails v2 desktop app (Go 1.23). The Go backend exposes methods directly to the React frontend via Wails bindings. There is **no REST API**.

For comprehensive architecture docs, data models, and extension guides, see **[docs/architecture.md](docs/architecture.md)**.

## Cross-Module Skill Identity Rule — MANDATORY

Any change touching skill identity, install/import/push/pull state, starred repo correlation, tool scan correlation, or skill update badges **must** follow the **"Unified Skill Identity & State Model"** section in `docs/architecture.md` and `docs/architecture_zh.md`.

- Distinguish **instance identity** (`Skill.ID`) from **logical identity** (stable cross-module key).
- Do **not** use `Name` or absolute `Path` as the primary cross-module identity.
- `imported` is the external-source wording for `installed`.
- `pushed` means the logical skill exists in a tool's configured `PushDir`.
- `seenInToolScan` means the logical skill was detected in a tool's configured `ScanDirs`; it does **not** imply SkillFlow previously pushed it.
- Git-backed update detection must be keyed by normalized repo source + subpath, and compare installed `SourceSHA` against the latest remote SHA for that same logical source.

### Key Design Decisions

- **`core/sync` import alias** — always import as `toolsync "github.com/shinerio/skillflow/core/sync"` (conflicts with stdlib `sync`)
- **`package main` files in `cmd/skillflow/`** — `app.go`, `adapters.go`, `providers.go`, `events.go` are all `package main` alongside `main.go` in `cmd/skillflow/` because Wails requires the app struct in the same package as `main`
- **Wails bindings are auto-generated** — after adding/removing exported methods on `App`, run `make generate` to update `cmd/skillflow/frontend/wailsjs/go/main/App.{js,d.ts}`; also manually add entries to `App.js` and `App.d.ts` if Wails CLI is unavailable
- **Installed skill instances are UUID-based, but cross-module identity must use a stable logical key** — see `docs/architecture.md#unified-skill-identity--state-model`
- **GitHub as source of truth** — update checker polls GitHub Commits API to compare SHA values
- **`SkippedUpdateVersion` in AppConfig** — persists which app version the user chose to skip on startup; `checkAppUpdateOnStartup` respects this; `CheckAppUpdateAndNotify` (manual check) always notifies regardless

### Adding a New App Method (Frontend-callable)

1. Add exported method to `App` struct in `cmd/skillflow/app.go` (or a new `package main` file in `cmd/skillflow/`)
2. Run `make generate` (or `cd cmd/skillflow && wails generate module`) to update `cmd/skillflow/frontend/wailsjs/go/main/App.{js,d.ts}`
3. Import and call from frontend: `import { MyNewMethod } from '../../wailsjs/go/main/App'`

### Adding a New Cloud Provider

1. Create `core/backup/<name>.go` implementing `backup.CloudProvider`
2. Register in `cmd/skillflow/providers.go`: `registry.RegisterCloudProvider(NewXxxProvider())`
3. Settings page auto-renders credential fields from `RequiredCredentials()`

### Adding a New Tool Adapter

Standard flat-directory tools: add to `registerAdapters()` in `cmd/skillflow/adapters.go`.
Custom behavior: implement `toolsync.ToolAdapter` and register via `registry.RegisterAdapter()`.
