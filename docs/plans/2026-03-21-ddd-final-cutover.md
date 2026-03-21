# Final DDD Cutover Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Finish the backend DDD refactor so `cmd/skillflow` is reduced to shell transport/coordinator code, legacy `core/config` runtime ownership is removed, cross-context write flows live in `core/orchestration`, and cross-context read composition lives in `core/readmodel`, without changing persisted behavior or frontend-visible functionality.

**Architecture:** Keep the frontend-facing `GetConfig` / `SaveConfig` contract behavior stable while moving internal ownership to context namespaces and platform helpers. Extract cross-context write flows into explicit orchestration services and move installed/starred/presence read composition into readmodel services. Leave shell-only concerns in `cmd/skillflow`, but only as thin transport and shell coordination layers.

**Tech Stack:** Go 1.23, Wails v2, React frontend bindings, testify, local filesystem persistence, JSON config split across `config.json` and `config_local.json`.

---

### Task 1: Introduce Final Settings Ownership

**Files:**
- Create: `core/platform/appdata/appdata.go`
- Create: `core/platform/shellsettings/*.go`
- Create: `core/skillcatalog/app/settings.go`
- Create: `core/agentintegration/app/settings.go`
- Create: `core/backup/app/settings.go`
- Create: `core/readmodel/preferences/*.go`
- Modify: `core/platform/settingsstore/*.go`
- Modify: `core/config/*.go`

**Step 1: Write failing tests**

- Add tests that prove namespace-owned settings can load/save the same effective values currently exposed by `config.Service`.
- Add tests that app-data path and proxy normalization come from platform packages, not `core/config`.

**Step 2: Run tests to verify they fail**

Run: `go test ./core/... -run 'Settings|Proxy|AppData' -count=1`

**Step 3: Implement minimal code**

- Move app-data path to `core/platform/appdata`.
- Move proxy types/normalization to `core/platform/shellsettings`.
- Add namespace structs/services for skillcatalog, agentintegration, backup, shell/platform, and UI visibility preferences.
- Refactor `core/platform/settingsstore` to own shared/local document composition.
- Reduce `core/config` to transport DTO compatibility only, or remove it entirely if Wails binding changes can be completed safely in the same pass.

**Step 4: Run tests to verify they pass**

Run: `go test ./core/... -count=1`

### Task 2: Extract Cross-Context Write Orchestration

**Files:**
- Create: `core/orchestration/*.go`
- Create: `core/orchestration/*_test.go`
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/app_restore.go`
- Modify: `cmd/skillflow/app_backup.go`

**Step 1: Write failing tests**

- Add orchestration tests for:
  - local import + optional auto-push + auto-backup
  - push/pull from agent flows
  - installed-skill update + pushed-copy refresh + auto-backup
  - restore compensation for restored skills and starred repos

**Step 2: Run tests to verify they fail**

Run: `go test ./core/orchestration/... -count=1`

**Step 3: Implement minimal code**

- Create explicit orchestrators for import, update, push/pull, and restore compensation.
- Move business coordination out of `App` and into orchestration services.
- Keep event emission and shell-specific timers/notifications in `cmd/skillflow`.

**Step 4: Run tests to verify they pass**

Run: `go test ./core/... -count=1`

### Task 3: Extract Cross-Context Read Composition

**Files:**
- Create: `core/readmodel/skills/*.go`
- Create: `core/readmodel/skills/*_test.go`
- Modify: `cmd/skillflow/app_viewstate.go`
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/skill_state.go`

**Step 1: Write failing tests**

- Add readmodel tests for:
  - installed skill list composition
  - starred skill list composition
  - agent presence snapshot reuse/rebuild
  - repo star skill composition with installed/updatable/pushed state

**Step 2: Run tests to verify they fail**

Run: `go test ./core/readmodel/... -count=1`

**Step 3: Implement minimal code**

- Move installed/starred/presence composition into `core/readmodel`.
- Keep cache storage under `core/readmodel/viewstate`.
- Make `cmd/skillflow` call readmodel services instead of assembling read DTOs inline.

**Step 4: Run tests to verify they pass**

Run: `go test ./core/... -count=1`

### Task 4: Thin Shell Transport and Settings Coordination

**Files:**
- Create: `cmd/skillflow/app_settings.go` or coordinator files as needed
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/process_helper.go`
- Modify: `cmd/skillflow/process_control.go`
- Modify: `cmd/skillflow/window_size.go`
- Modify: `cmd/skillflow/app_proxy.go`
- Modify: `cmd/skillflow/adapters.go`

**Step 1: Write failing tests**

- Add or tighten shell tests around:
  - startup config loading
  - settings save rollback for launch-at-login
  - helper logger config load
  - window state and proxy behavior

**Step 2: Run tests to verify they fail**

Run: `go test ./cmd/skillflow -count=1`

**Step 3: Implement minimal code**

- Extract a shell-side settings save coordinator from `App`.
- Keep `App` methods as thin transport adapters.
- Remove helper methods in `App` that only existed to host business logic now owned elsewhere.

**Step 4: Run tests to verify they pass**

Run: `go test ./cmd/skillflow -count=1`

### Task 5: Finish Contract and Documentation

**Files:**
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`
- Modify: `cmd/skillflow/frontend/wailsjs/go/models.ts`
- Modify: `docs/architecture/*.md`
- Modify: `docs/config.md`
- Modify: `docs/config_zh.md`
- Modify: relevant frontend pages only if Wails DTO namespace changes

**Step 1: Regenerate bindings**

Run: `make generate`

**Step 2: Update docs**

- Sync architecture docs with the final ownership model.
- Update config docs if any persisted schema/semantics changed.

**Step 3: Run final verification**

Run:
- `go test ./core/... -count=1`
- `go test ./cmd/skillflow -count=1`

**Step 4: Commit**

```bash
git add -A
git commit -m "refactor: finish ddd backend cutover"
```
