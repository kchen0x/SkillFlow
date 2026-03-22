# My Memory Batch Push Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Redesign **My Memory** so editing is isolated from push controls, Markdown preview renders correctly, automatic sync lives in the page header, and manual push becomes an inline batch-push flow.

**Architecture:** Keep `memorycatalog` as the source of truth for persisted memory content and per-agent push config. Add a selection-aware batch-push path in the backend, remove persistent module push targets through a startup upgrade, and move page interaction complexity into small frontend helpers that can be unit-tested independently.

**Tech Stack:** Go, React, TypeScript, Wails, Node test runner, Go test, Markdown docs

---

### Task 1: Lock the new frontend state rules with failing tests

**Files:**
- Create: `cmd/skillflow/frontend/src/lib/memoryPageState.ts`
- Create: `cmd/skillflow/frontend/src/lib/memoryMarkdown.ts`
- Create: `cmd/skillflow/frontend/tests/memoryPageState.test.mjs`
- Create: `cmd/skillflow/frontend/tests/memoryMarkdown.test.mjs`
- Modify: `cmd/skillflow/frontend/package.json`

**Step 1: Write the failing tests**

Add tests that prove:

- auto-sync config maps to `off`, `merge`, and `takeover`
- entering batch-push mode starts with no selected modules and no selected agents
- main memory is always selected for batch push
- batch push is not submittable without at least one selected module and one selected agent
- Markdown rendering emits headings, lists, blockquotes, fenced code blocks, inline code, and links

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `FAIL` because the new helper files and compile entries do not exist yet.

**Step 3: Write minimal implementation**

Implement pure helpers for:

- auto-sync mode mapping
- batch-push selection state and validation
- minimal Markdown token rendering

Update the unit-test script so these helpers compile into `.tmp-tests`.

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `ok`

### Task 2: Lock the backend batch-push and config-cutover behavior with failing tests

**Files:**
- Create: `core/platform/upgrade/memory_local_config_test.go`
- Create: `core/memorycatalog/app/push_service_test.go`
- Modify: `cmd/skillflow/app_autopush_test.go`

**Step 1: Write the failing tests**

Add tests that prove:

- startup upgrade removes legacy `memory_local.json.modules`
- selection push writes only selected module files and removes unselected ones
- selection push updates the pushed hash so partial push still reports `pendingPush`
- saving memory auto-syncs all memories to agents whose memory config is `autoPush=true`

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./core/platform/upgrade ./core/memorycatalog/app ./cmd/skillflow -run 'Test(Memory|Save.*AutoPush)'
```

Expected: `FAIL` because the upgrade path and selection-push behavior do not exist yet.

**Step 3: Write minimal implementation**

Implement:

- startup migration for `memory/memory_local.json`
- selection-aware push helpers in `core/memorycatalog/app/push_service.go`
- app-level auto-sync after create, save, and delete memory mutations

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./core/platform/upgrade ./core/memorycatalog/app ./cmd/skillflow -run 'Test(Memory|Save.*AutoPush)'
```

Expected: `ok`

### Task 3: Update the My Memory page to use the new helpers and flow

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/Memory.tsx`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.js`
- Modify: `cmd/skillflow/frontend/wailsjs/go/models.ts`
- Modify: `cmd/skillflow/app_memory.go`

**Step 1: Re-run the failing frontend tests**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: helper tests may pass, but page behavior is still missing.

**Step 2: Write minimal implementation**

Implement in `Memory.tsx`:

- top auto-sync panel in default state
- inline batch-push state instead of the old direct push action
- checkbox rendering on cards during batch-push mode
- header swap from auto-sync controls to target-agent and mode controls
- drawer cleanup so it contains only editing features
- unsaved-changes confirmation dialog
- Markdown preview using the new helper renderer

Implement in Go:

- new `PushSelectedMemory` transport method
- keep `PushAllMemory` available if needed for compatibility, but do not use it from the page

Regenerate Wails bindings after the exported app method change.

**Step 3: Run targeted verification**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Run:

```bash
go test ./cmd/skillflow ./core/memorycatalog/app ./core/platform/upgrade
```

Expected: `ok`

### Task 4: Sync feature and config documentation

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `docs/config.md`
- Modify: `docs/config_zh.md`

**Step 1: Write the documentation updates**

Update the docs so they describe:

- `Batch Push` instead of `Push All`
- top-level auto-sync controls
- drawer-only editing behavior
- unsaved-change confirmation
- Markdown preview rendering
- removal of persistent module push targets from `memory_local.json`

**Step 2: Run checks**

Run:

```bash
rg -n "Batch Push|批量推送|Auto Sync|自动同步|memory_local.json|pushTargets|未保存" docs/features.md docs/features_zh.md docs/config.md docs/config_zh.md
```

Expected: matches in all four files, with no stale description of persistent module push targets.

### Task 5: Final verification

**Files:**
- No code changes expected

**Step 1: Run frontend unit tests**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `ok`

**Step 2: Run frontend production build**

Run:

```bash
cd cmd/skillflow/frontend && npm run build
```

Expected: `ok`

**Step 3: Run backend tests**

Run:

```bash
go test ./core/... ./cmd/skillflow
```

Expected: `ok`
