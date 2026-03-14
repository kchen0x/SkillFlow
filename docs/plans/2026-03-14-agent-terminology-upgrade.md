# Agent Terminology Upgrade Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rename SkillFlow's "tool" business concept to "agent" everywhere, and perform a startup-time config schema cutover so the application only reads the new `agent` schema after upgrade.

**Architecture:** Add a dedicated `core/upgrade` layer that rewrites persisted config files before normal startup, then rename backend config/types/APIs and frontend bindings/routes/state so the whole app consistently uses `agent` terminology. Keep `core/sync/` as the package boundary for synchronization, but rename business-facing types, DTOs, and user-visible strings to `agent`, and rebuild derived caches instead of migrating them field-by-field.

**Tech Stack:** Go 1.23, Wails desktop backend, React 18 + TypeScript, JSON config files, Go tests, frontend unit tests, Wails binding generation

---

### Task 1: Add failing tests for startup config cutover

**Files:**
- Create: `core/upgrade/config_terms_test.go`
- Create: `core/upgrade/upgrade.go`

**Step 1: Write the failing test**

- Add table-driven tests that create temp `config.json` and `config_local.json` files containing old keys:
  - `tools`
  - `autoPushTools`
  - `myTools`
  - `pushToTool`
  - `pullFromTool`
  - `pushedTools`
- Assert that running the upgrade rewrites them to:
  - `agents`
  - `autoPushAgents`
  - `myAgents`
  - `pushToAgent`
  - `pullFromAgent`
  - `pushedAgents`
- Add a second test that reruns the upgrade on already-upgraded files and expects no further change.

**Step 2: Run test to verify it fails**

Run: `go test ./core/upgrade/...`

Expected: FAIL because the package or upgrade entrypoint does not exist yet.

**Step 3: Write minimal implementation**

- Create `core/upgrade`.
- Implement a small startup upgrade runner that:
  - reads `config.json` and `config_local.json` if they exist
  - rewrites old JSON keys to the new schema
  - returns an error on malformed JSON or failed writes
- Keep the implementation file-focused and independent from higher-level app logic.

**Step 4: Run test to verify it passes**

Run: `go test ./core/upgrade/...`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: add startup config terminology upgrade"
```

### Task 2: Wire upgrade execution into startup before config load

**Files:**
- Modify: `cmd/skillflow/app.go`
- Test: `cmd/skillflow/app_startup_test.go`

**Step 1: Write the failing test**

- Add or extend an app startup test that uses a temp app-data directory with old-schema config files.
- Expect startup to run the upgrade before `config.NewService(...).Load()` consumes those files.
- Assert that the loaded config exposes only new `agent` fields after startup.

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/skillflow -run 'TestStartup.*Upgrade'`

Expected: FAIL because startup does not invoke the new upgrade layer yet.

**Step 3: Write minimal implementation**

- Call the `core/upgrade` entrypoint at the top of `App.startup()`, after resolving `dataDir` and before constructing the config service.
- Add required `info` / `error` logs with stable operation naming.
- On upgrade failure, abort startup instead of falling through into mixed-schema execution.

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/skillflow -run 'TestStartup.*Upgrade'`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: run config upgrade before startup config load"
```

### Task 3: Rename config models and status-visibility schema to `agent`

**Files:**
- Modify: `core/config/model.go`
- Modify: `core/config/defaults.go`
- Modify: `core/config/service.go`
- Modify: `core/config/skill_status_visibility.go`
- Modify: `core/config/service_test.go`

**Step 1: Write the failing test**

- Update or add config tests that save and load:
  - `AgentConfig`
  - `AutoPushAgents`
  - `Agents`
  - status visibility keys `myAgents`, `pushToAgent`, `pullFromAgent`
  - status value `pushedAgents`
- Verify `config.json` and `config_local.json` contain only new field names.

**Step 2: Run test to verify it fails**

Run: `go test ./core/config/...`

Expected: FAIL because the config package still uses `tool` field names and JSON tags.

**Step 3: Write minimal implementation**

- Rename:
  - `ToolConfig` -> `AgentConfig`
  - `AutoPushTools` -> `AutoPushAgents`
  - `Tools` -> `Agents`
  - `SkillStatusPushedTools` -> `SkillStatusPushedAgents`
  - `MyTools` -> `MyAgents`
  - `PushToTool` -> `PushToAgent`
  - `PullFromTool` -> `PullFromAgent`
- Update JSON tags, defaults, normalize helpers, and shared/local split logic.
- Remove old-schema compatibility reads from `core/config`; that responsibility now belongs to `core/upgrade`.

**Step 4: Run test to verify it passes**

Run: `go test ./core/config/...`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "refactor: rename config tool schema to agent"
```

### Task 4: Rename sync adapter types and backend DTO fields

**Files:**
- Modify: `core/sync/adapter.go`
- Modify: `core/sync/filesystem_adapter.go`
- Modify: `core/sync/filesystem_adapter_test.go`
- Modify: `core/registry/registry.go`
- Modify: `cmd/skillflow/adapters.go`
- Modify: `cmd/skillflow/skill_state.go`
- Modify: `cmd/skillflow/skill_state_test.go`
- Modify: `cmd/skillflow/app_viewstate.go`
- Modify: `core/viewstate/tool_presence.go`
- Modify: `core/viewstate/tool_presence_test.go`

**Step 1: Write the failing test**

- Update backend tests that cover pushed-state aggregation, scan correlation, and viewstate fingerprints so they expect:
  - `pushedAgents`
  - `seenInAgentScan`
  - renamed page/status keys
- Keep identity logic unchanged; only the business terminology should move.

**Step 2: Run test to verify it fails**

Run: `go test ./core/viewstate/... ./cmd/skillflow -run 'Test.*(Presence|State|ViewState)'`

Expected: FAIL because DTOs and visibility logic still expose `tool` names.

**Step 3: Write minimal implementation**

- Rename business-facing sync types:
  - `ToolAdapter` -> `AgentAdapter`
  - `toolsync` alias -> `agentsync`
- Rename DTO fields and internal maps from `tool`-centric names to `agent`-centric names where they represent the SkillFlow concept.
- Keep package path `core/sync` intact.

**Step 4: Run test to verify it passes**

Run: `go test ./core/viewstate/... ./cmd/skillflow -run 'Test.*(Presence|State|ViewState)'`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "refactor: rename sync state from tool to agent"
```

### Task 5: Rename backend app methods and conflict structures to `agent`

**Files:**
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/push_conflict.go`
- Modify: `cmd/skillflow/app_autopush_test.go`
- Modify: `cmd/skillflow/app_restore_test.go`
- Modify: `cmd/skillflow/app_viewstate_test.go`

**Step 1: Write the failing test**

- Update or add app-level tests for:
  - `GetEnabledAgents()`
  - `ListAgentSkills()`
  - `DeleteAgentSkill()`
  - `ScanAgentSkills()`
  - `PullFromAgent()`
  - `PushToAgents()`
  - `PushToAgentsForce()`
  - `CheckMissingAgentPushDirs()`
- Assert auto-push behavior uses `AutoPushAgents`.

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/skillflow -run 'Test.*(Agent|AutoPush|Pull|Push)'`

Expected: FAIL because exported methods and related code still use `tool` names.

**Step 3: Write minimal implementation**

- Rename the exported Wails-facing app methods and related helpers.
- Update log messages so operation names and target labels use `agent`.
- Keep operation semantics the same while replacing names.

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/skillflow -run 'Test.*(Agent|AutoPush|Pull|Push)'`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "refactor: rename backend app APIs from tool to agent"
```

### Task 6: Rename frontend status config, bindings usage, and page state

**Files:**
- Modify: `cmd/skillflow/frontend/src/lib/skillStatusVisibility.ts`
- Modify: `cmd/skillflow/frontend/src/contexts/SkillStatusVisibilityContext.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/Settings.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/Dashboard.tsx`
- Modify: `cmd/skillflow/frontend/src/components/SkillCard.tsx`
- Modify: `cmd/skillflow/frontend/src/components/SyncSkillCard.tsx`
- Modify: `cmd/skillflow/frontend/src/components/SkillStatusStrip.tsx`
- Modify: `cmd/skillflow/frontend/tests/skillStatusStrip.test.mjs`

**Step 1: Write the failing test**

- Update frontend unit tests to expect:
  - `pushedAgents`
  - `myAgents`
  - `pushToAgent`
  - `pullFromAgent`
- Add a focused test for the status-strip helper if needed so renamed props still truncate and count overflow correctly.

**Step 2: Run test to verify it fails**

Run: `cd cmd/skillflow/frontend && npm test -- skillStatusStrip.test.mjs`

Expected: FAIL because the frontend still uses old status keys and prop names.

**Step 3: Write minimal implementation**

- Rename frontend status config keys and prop names.
- Update page state and component props from `tool` terminology to `agent` terminology where they reflect SkillFlow business concepts.
- Keep existing behavior and rendering order unchanged.

**Step 4: Run test to verify it passes**

Run: `cd cmd/skillflow/frontend && npm test -- skillStatusStrip.test.mjs`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "refactor: rename frontend status and state to agent"
```

### Task 7: Rename frontend pages, routes, bindings, and generated models

**Files:**
- Modify: `cmd/skillflow/frontend/src/App.tsx`
- Move/Rename: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/SyncPush.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/SyncPull.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/StarredRepos.tsx`
- Modify: `cmd/skillflow/frontend/src/config/toolIcons.tsx`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.js`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`
- Modify: `cmd/skillflow/frontend/wailsjs/go/models.ts`

**Step 1: Write the failing test**

- Update TypeScript usages so compilation expects:
  - `GetEnabledAgents`
  - `ListAgentSkills`
  - `DeleteAgentSkill`
  - `ScanAgentSkills`
  - `PushToAgents`
  - `PushToAgentsForce`
  - `CheckMissingAgentPushDirs`
- Rename the "My Tools" page component and route to the new `agent` naming.

**Step 2: Run test to verify it fails**

Run: `cd cmd/skillflow/frontend && npm run build`

Expected: FAIL because bindings, imports, or route/page names still reference old `tool` identifiers.

**Step 3: Write minimal implementation**

- Rename frontend page/component files and route usage to `agent` terminology.
- Regenerate Wails bindings after the exported Go methods are renamed:

```bash
make generate
```

- If regeneration is unavailable, update the generated binding files manually to match the new exported method names and models.

**Step 4: Run test to verify it passes**

Run: `cd cmd/skillflow/frontend && npm run build`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "refactor: rename frontend pages and bindings to agent"
```

### Task 8: Update English and Chinese user-visible copy

**Files:**
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`

**Step 1: Write the failing test**

- Add a small coverage check if needed for renamed translation keys, or rely on the TypeScript build to fail on missing keys.
- Update call sites so any missing renamed translation key becomes a compile-time or build-time error.

**Step 2: Run test to verify it fails**

Run: `cd cmd/skillflow/frontend && npm run build`

Expected: FAIL because `tool` translation keys or labels are still referenced.

**Step 3: Write minimal implementation**

- Replace user-visible "tool" / "工具" strings with "agent" / "智能体" across navigation, settings, push/pull pages, status labels, dialogs, and messages.
- Keep non-business external terminology untouched where it intentionally refers to something outside SkillFlow's own concept model.

**Step 4: Run test to verify it passes**

Run: `cd cmd/skillflow/frontend && npm run build`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: update UI terminology to agent"
```

### Task 9: Update repository guidance and all affected documentation

**Files:**
- Modify: `AGENTS.md`
- Modify: `README.md`
- Modify: `README_zh.md`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `docs/architecture.md`
- Modify: `docs/architecture_zh.md`
- Modify: `docs/config.md`
- Modify: `docs/config_zh.md`

**Step 1: Write the failing check**

- Review each document and mark all user-facing or schema-facing `tool` terminology that now conflicts with the approved design.
- Specifically verify `docs/config*.md` examples still show old keys before editing.

**Step 2: Run check to verify it fails**

Run: `rg -n "\\btool(s)?\\b|Tool|工具|推送到工具|从工具拉取|我的工具|autoPushTools|pushedTools|myTools|pushToTool|pullFromTool" AGENTS.md README.md README_zh.md docs`

Expected: Matches found in business terminology and config schema descriptions that need to be updated.

**Step 3: Write minimal implementation**

- Update docs to describe the new `agent` terminology and schema.
- Add a persistent `AGENTS.md` rule that future breaking config/terminology upgrades must:
  - live under `core/upgrade`
  - run at startup before business init
  - migrate persisted files directly
  - avoid old-schema business compatibility code

**Step 4: Run check to verify it passes**

Run: `rg -n "autoPushTools|pushedTools|myTools|pushToTool|pullFromTool" AGENTS.md README.md README_zh.md docs`

Expected: No matches in maintained docs unless a retained historical/example reference is explicitly intentional.

**Step 5: Commit**

```bash
git add -A
git commit -m "docs: rename tool terminology to agent"
```

### Task 10: Run final backend, frontend, and binding verification

**Files:**
- Verify only; no planned file creation

**Step 1: Run focused backend tests**

Run: `go test ./core/upgrade/... ./core/config/... ./core/viewstate/... ./cmd/skillflow`

Expected: PASS

**Step 2: Run frontend test/build verification**

Run: `cd cmd/skillflow/frontend && npm test -- skillStatusStrip.test.mjs`

Expected: PASS

**Step 3: Run final frontend build**

Run: `cd cmd/skillflow/frontend && npm run build`

Expected: PASS

**Step 4: Run final terminology grep**

Run: `rg -n "autoPushTools|pushedTools|myTools|pushToTool|pullFromTool|GetEnabledTools|ListToolSkills|DeleteToolSkill|ScanToolSkills|PushToTools|PushToToolsForce|CheckMissingPushDirs" cmd/skillflow core docs AGENTS.md README.md README_zh.md`

Expected: No matches in active app code or maintained docs.

**Step 5: Commit**

```bash
git add -A
git commit -m "chore: verify agent terminology upgrade"
```
