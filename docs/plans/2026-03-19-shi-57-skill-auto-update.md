# Skill Auto Update Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a My Skills auto-update toggle that automatically updates installed git-backed skills after starred-repo cache refresh, auto-pushes to selected auto-sync agents, and keeps non-auto-sync agent copies marked as updatable when they fall behind My Skills.

**Architecture:** Persist a new local-only `autoUpdateSkills` flag in config, then reuse the existing `UpdateSkill()` path from a new repo-refresh-driven auto-update coordinator. Refine agent list aggregation so `updatable` covers both upstream cache updates and agent copies that are stale relative to the installed library copy. Emit a shared post-update event so Dashboard and My Agents reload after background updates.

**Tech Stack:** Go 1.23, Wails v2 backend bindings, React + TypeScript frontend, testify Go tests, frontend `.mjs` tests, markdown docs

---

### Task 1: Add local config plumbing for `autoUpdateSkills`

**Files:**
- Modify: `core/config/model.go`
- Modify: `core/config/service.go`
- Test: `core/config/service_test.go`
- Test: `core/config/agent_terms_test.go`
- Modify: `docs/config.md`
- Modify: `docs/config_zh.md`

**Step 1: Write the failing tests**

Add tests that save and load `AppConfig.AutoUpdateSkills`, and assert:

- the value round-trips through `Service.Save()` / `Service.Load()`
- `config.json` does not contain `"autoUpdateSkills"`
- `config_local.json` does contain `"autoUpdateSkills"`

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./core/config/... -run 'Test(AutoUpdateSkills|AutoPushAgentsStoredOnlyInLocalConfig|SaveAndLoadConfig)' -count=1
```

Expected: FAIL because `AppConfig` and local/shared config split do not know about `autoUpdateSkills`.

**Step 3: Write the minimal implementation**

Update config structs and split/merge logic so `autoUpdateSkills` is stored only in local config and defaults to `false`.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./core/config/... -count=1
```

Expected: PASS

**Step 5: Commit**

```bash
git add core/config/model.go core/config/service.go core/config/service_test.go core/config/agent_terms_test.go docs/config.md docs/config_zh.md
git commit -m "feat: add local skill auto update config"
```

### Task 2: Add backend auto-update orchestration after repo refresh

**Files:**
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/app_startup.go`
- Modify: `core/notify/model.go`
- Test: `cmd/skillflow/app_skill_update_test.go`
- Test: `cmd/skillflow/app_startup_test.go`

**Step 1: Write the failing tests**

Add tests that cover:

- `UpdateStarredRepo()` auto-updates matching installed skills when `AutoUpdateSkills=true`
- `UpdateAllStarredRepos()` auto-updates only repos that refreshed successfully
- no automatic update occurs when `AutoUpdateSkills=false`
- startup background task order refreshes starred repos before `CheckUpdates()`

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./cmd/skillflow -run 'Test(UpdateStarredRepo|UpdateAllStarredRepos|StartupBackgroundTaskPlan)' -count=1
```

Expected: FAIL because repo refresh does not trigger installed auto update and startup order still checks updates before starred refresh.

**Step 3: Write the minimal implementation**

Implement:

- local-config gate for auto update
- repo-url-targeted installed-skill auto-update helper
- startup task reordering
- `skills.updated` event emission from successful `UpdateSkill()`

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./cmd/skillflow -run 'Test(UpdateStarredRepo|UpdateAllStarredRepos|StartupBackgroundTaskPlan|UpdateSkill)' -count=1
```

Expected: PASS

**Step 5: Commit**

```bash
git add cmd/skillflow/app.go cmd/skillflow/app_startup.go cmd/skillflow/app_skill_update_test.go cmd/skillflow/app_startup_test.go core/notify/model.go
git commit -m "feat: auto update installed skills after repo refresh"
```

### Task 3: Keep non-auto-sync agent copies marked as updatable

**Files:**
- Modify: `core/skill/index.go`
- Modify: `cmd/skillflow/skill_state.go`
- Test: `core/skill/index_test.go`
- Test: `cmd/skillflow/skill_state_test.go`

**Step 1: Write the failing tests**

Add tests that prove:

- an agent copy correlated to an installed skill is `updatable=true` when the installed skill content changed and the agent copy did not
- an agent copy already matching the installed content is not marked stale
- upstream-only installed update state still works as before

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./core/skill/... ./cmd/skillflow -run 'Test(Resolve|Agent).*Updatable' -count=1
```

Expected: FAIL because agent `updatable` is currently inherited only from installed upstream state.

**Step 3: Write the minimal implementation**

Extend installed/agent correlation so agent aggregation can compare agent content keys against the correlated installed content key and set `updatable=true` when the agent copy is stale.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./core/skill/... ./cmd/skillflow -run 'Test(Resolve|Agent).*Updatable' -count=1
```

Expected: PASS

**Step 5: Commit**

```bash
git add core/skill/index.go core/skill/index_test.go cmd/skillflow/skill_state.go cmd/skillflow/skill_state_test.go
git commit -m "feat: mark stale agent skill copies as updatable"
```

### Task 4: Add Dashboard auto-update toggle and frontend refresh hooks

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/Dashboard.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`
- Test: `cmd/skillflow/frontend/tests/appActivity.test.mjs`
- Test: `cmd/skillflow/frontend/tests/skillStatusStrip.test.mjs`
- Test: `cmd/skillflow/frontend/tests/listLoadState.test.mjs`

**Step 1: Write the failing tests**

Add frontend tests that verify:

- Dashboard renders and toggles the new auto-update control from config state
- failed save rolls UI state back
- `skills.updated` causes Dashboard and ToolSkills reload handlers to run

**Step 2: Run tests to verify they fail**

Run:

```bash
cd cmd/skillflow/frontend && npm test -- appActivity.test.mjs
```

Expected: FAIL because the new control and event handling do not exist yet.

**Step 3: Write the minimal implementation**

Add the toggle UI, copy, optimistic save/rollback behavior, and event subscriptions for `skills.updated`.

**Step 4: Run tests to verify they pass**

Run:

```bash
cd cmd/skillflow/frontend && npm test -- appActivity.test.mjs
```

Expected: PASS

**Step 5: Commit**

```bash
git add cmd/skillflow/frontend/src/pages/Dashboard.tsx cmd/skillflow/frontend/src/pages/ToolSkills.tsx cmd/skillflow/frontend/src/i18n/en.ts cmd/skillflow/frontend/src/i18n/zh.ts cmd/skillflow/frontend/tests/appActivity.test.mjs
git commit -m "feat: add dashboard skill auto update toggle"
```

### Task 5: Update feature and architecture docs, then run full verification

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `README.md`
- Modify: `README_zh.md`
- Modify: `docs/architecture/README.md`
- Modify: `docs/architecture/README_zh.md`
- Modify: `docs/config.md`
- Modify: `docs/config_zh.md`

**Step 1: Write the doc updates**

Document:

- the new My Skills auto-update toggle
- trigger timing after startup/manual starred repo refresh
- auto-sync vs non-auto-sync agent behavior
- local-only config persistence
- expanded meaning of agent `updatable`

**Step 2: Run verification**

Run:

```bash
go test ./core/... ./cmd/skillflow -count=1
cd cmd/skillflow/frontend && npm test
```

Expected: PASS

**Step 3: Run a focused sanity check**

Run:

```bash
git status --short
```

Expected: only intended implementation and docs changes remain.

**Step 4: Commit**

```bash
git add docs/features.md docs/features_zh.md README.md README_zh.md docs/architecture/README.md docs/architecture/README_zh.md docs/config.md docs/config_zh.md
git commit -m "docs: document skill auto update flow"
```
