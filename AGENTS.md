# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Documentation Organization Rule — MANDATORY

**Root directory must contain only `README.md` and `README_zh.md` as documentation files.**

All other documentation lives under `docs/`:

| File | Purpose |
|------|---------|
| `docs/features.md` | Complete UI/UX feature reference in English |
| `docs/features_zh.md` | Complete UI/UX feature reference in Chinese |
| `docs/architecture.md` | Internal architecture, packages, data models, extension guides (English) |
| `docs/architecture_zh.md` | Same in Chinese |
| `docs/plans/` | Design and implementation plans |
| `docs/skill_directory.md` | Skill directory format specification |

**Rules:**
- `README.md` / `README_zh.md` — user-facing only: features overview, download/install links, skill format, cloud backup config, contributing/build instructions. **No internal code snippets, no package tables, no architecture diagrams.**
- Never add new standalone `.md` files to the root directory. If you need new documentation, put it under `docs/`.

## Documentation Sync Rule — MANDATORY

**Any time a feature is added, changed, or removed, you MUST update the following files in the same commit:**

| File | What to update |
|------|---------------|
| `docs/features.md` | Add / edit / remove the corresponding section(s) in English. Update the "Last updated" date at the bottom. |
| `docs/features_zh.md` | Same changes in Chinese. Update the "最后更新" date at the bottom. |
| `README.md` | Update the Features table row(s) if the high-level description changes. |
| `README_zh.md` | Same in Chinese. |

**Rules:**
- A "feature change" includes: any new UI element (button, dialog, toggle, input), behavior change, removal of a control, new backend method callable from the frontend, and new event type.
- Do **not** leave the docs stale. Never commit a feature change without the corresponding doc update in the same commit.
- `docs/features.md` / `docs/features_zh.md` are the source of truth for UX details. README files only carry high-level summaries with links to the feature files.

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
~/go/bin/wails dev

# Build production binary
~/go/bin/wails build

# Regenerate TypeScript bindings after changing App struct methods
~/go/bin/wails generate module
```

### Go (backend)

```bash
# Run all tests
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
cd frontend
npm install        # install dependencies
npm run dev        # Vite dev server (used by wails dev)
npm run build      # production build (output: frontend/dist/)
```

## Architecture

SkillFlow is a Wails v2 desktop app (Go 1.23). The Go backend exposes methods directly to the React frontend via Wails bindings. There is **no REST API**.

For comprehensive architecture docs, data models, and extension guides, see **[docs/architecture.md](docs/architecture.md)**.

### Key Design Decisions

- **`core/sync` import alias** — always import as `toolsync "github.com/shinerio/skillflow/core/sync"` (conflicts with stdlib `sync`)
- **`package main` files at root** — `app.go`, `adapters.go`, `providers.go`, `events.go` are all `package main` alongside `main.go` because Wails requires the app struct in the same package as `main`
- **Wails bindings are auto-generated** — after adding/removing exported methods on `App`, run `make generate` to update `frontend/wailsjs/go/main/App.{js,d.ts}`; also manually add entries to `App.js` and `App.d.ts` if Wails CLI is unavailable
- **UUID-based skills** — skills are identified by UUID, metadata stored in JSON sidecars under `meta/`
- **GitHub as source of truth** — update checker polls GitHub Commits API to compare SHA values
- **`SkippedUpdateVersion` in AppConfig** — persists which app version the user chose to skip on startup; `checkAppUpdateOnStartup` respects this; `CheckAppUpdateAndNotify` (manual check) always notifies regardless

### Adding a New App Method (Frontend-callable)

1. Add exported method to `App` struct in `app.go` (or a new `package main` file at root)
2. Run `make generate` (or `wails generate module`) to update `frontend/wailsjs/go/main/App.{js,d.ts}`
3. Import and call from frontend: `import { MyNewMethod } from '../../wailsjs/go/main/App'`

### Adding a New Cloud Provider

1. Create `core/backup/<name>.go` implementing `backup.CloudProvider`
2. Register in `providers.go`: `registry.RegisterCloudProvider(NewXxxProvider())`
3. Settings page auto-renders credential fields from `RequiredCredentials()`

### Adding a New Tool Adapter

Standard flat-directory tools: add to `registerAdapters()` in `adapters.go`.
Custom behavior: implement `toolsync.ToolAdapter` and register via `registry.RegisterAdapter()`.
