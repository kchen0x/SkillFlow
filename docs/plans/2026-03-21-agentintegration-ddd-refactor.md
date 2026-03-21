# Agentintegration DDD Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extract agent-side scan/push/pull presence semantics from `core/sync` and `cmd/skillflow` into `core/agentintegration`, switch shell callers to the new app layer, and remove `core/sync` without changing user-visible behavior.

**Architecture:** Create `core/agentintegration/domain`, `app`, and `infra/gateway`. Move gateway contracts and filesystem transport into the context, move push/scan/presence semantics into `app` and `domain`, and keep cross-context pull/import orchestration in `cmd/skillflow` until `orchestration` exists.

**Tech Stack:** Go 1.23, Wails v2 backend, testify unit tests, markdown docs

---

### Task 1: Create failing tests for the new `agentintegration` package tree

**Files:**
- Create: `core/agentintegration/domain/agent_test.go`
- Create: `core/agentintegration/app/service_test.go`
- Create: `core/agentintegration/infra/gateway/filesystem_adapter_test.go`

**Step 1: Write the failing tests**

Port and add behavior coverage for:

- filesystem push/pull/max-depth scan behavior from `core/sync`
- enabled-agent filtering
- push conflict detection
- agent scan candidate aggregation
- pushed and seen presence behavior

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./core/agentintegration/... -count=1
```

Expected: FAIL because the new packages do not exist yet.

**Step 3: Write the minimal implementation**

Create the new agentintegration package tree and implement only enough code for those tests to pass.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./core/agentintegration/... -count=1
```

Expected: PASS

### Task 2: Switch shell agent entrypoints to `agentintegration/app`

**Files:**
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/skill_state.go`
- Modify: `cmd/skillflow/adapters.go`
- Modify: `core/registry/registry.go`
- Modify: relevant shell tests under `cmd/skillflow/*_test.go`

**Step 1: Write the failing tests**

Update shell-facing tests so they expect agent push/scan logic to delegate to `agentintegration` types and services rather than `core/sync` and shell-local aggregation helpers.

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./cmd/skillflow -run 'Test(GetEnabledAgents|Prompt|AutoPush|Restore|SkillUpdate)' -count=1
```

Expected: FAIL until the shell caller chain is moved onto `agentintegration`.

**Step 3: Write the minimal implementation**

Move shell agent entrypoints onto `agentintegration/app` while keeping config loading, skillcatalog import, and backup scheduling in the shell.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./cmd/skillflow -count=1
```

Expected: PASS

### Task 3: Remove `core/sync` and regenerate bindings

**Files:**
- Delete: `core/sync/adapter.go`
- Delete: `core/sync/filesystem_adapter.go`
- Delete: `core/sync/filesystem_adapter_test.go`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`
- Modify: `cmd/skillflow/frontend/wailsjs/go/models.ts`

**Step 1: Write the failing compile/build check**

Verify the repo no longer compiles until the old sync imports and generated bindings are updated.

**Step 2: Run checks to verify they fail**

Run:

```bash
go test ./core/... ./cmd/skillflow -run '^$' -count=1
```

Expected: FAIL until `core/sync` callers and generated types are fully switched.

**Step 3: Write the minimal implementation**

Delete `core/sync`, regenerate Wails bindings, and align the remaining agent-facing types to `agentintegration`.

**Step 4: Run tests/build to verify they pass**

Run:

```bash
make generate
go test ./core/... ./cmd/skillflow -count=1
cd cmd/skillflow/frontend && npm run build
```

Expected: PASS

### Task 4: Update architecture progress docs and verify migration hygiene

**Files:**
- Modify: `docs/architecture/README.md`
- Modify: `docs/architecture/README_zh.md`
- Modify: `docs/architecture/migration.md`
- Modify: `docs/architecture/migration_zh.md`

**Step 1: Write the doc updates**

Document that `agentintegration` is now extracted and `core/sync` is gone.

**Step 2: Run final verification**

Run:

```bash
rg -n '\"github.com/shinerio/skillflow/core/sync\"|core/sync\\b' core cmd/skillflow --glob '!core/sync/**'
git status --short
```

Expected:

- no remaining active `core/sync` code references
- worktree contains only intended agentintegration migration changes
