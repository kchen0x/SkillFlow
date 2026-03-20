# Promptcatalog DDD Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extract prompt library truth from `core/prompt` into `core/promptcatalog`, switch shell callers to the new app layer, and remove the old package without changing prompt behavior.

**Architecture:** Split the current prompt package into `domain`, `app`, and `infra/repository`. Keep prompt CRUD, category management, import/export, and validation in the bounded context while leaving Wails dialogs and prompt import session state in `cmd/skillflow`.

**Tech Stack:** Go 1.23, Wails v2 backend, testify unit tests, markdown docs

---

### Task 1: Create failing tests for the new `promptcatalog` package tree

**Files:**
- Create: `core/promptcatalog/domain/prompt_test.go`
- Create: `core/promptcatalog/app/service_test.go`
- Create: `core/promptcatalog/infra/repository/filesystem_storage_test.go`

**Step 1: Write the failing tests**

Port current `core/prompt` behavior checks to the new package tree so the new context expects:

- prompt/category normalization and validation
- prompt create/update/delete/category operations
- import/export JSON behavior
- legacy layout migration

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./core/promptcatalog/... -count=1
```

Expected: FAIL because the new packages do not exist yet.

**Step 3: Write the minimal implementation**

Create the new promptcatalog package tree and implement only enough code for those tests to pass.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./core/promptcatalog/... -count=1
```

Expected: PASS

### Task 2: Switch prompt shell entrypoints to `promptcatalog/app`

**Files:**
- Modify: `cmd/skillflow/app_prompt.go`
- Modify: `cmd/skillflow/app_prompt_session.go`
- Modify: `cmd/skillflow/app_prompt_session_test.go`

**Step 1: Write the failing tests**

Update prompt shell tests so they expect shell code to depend on `promptcatalog` types rather than `core/prompt`.

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./cmd/skillflow -run 'TestPrompt' -count=1
```

Expected: FAIL because shell code and prompt session structs still depend on `core/prompt`.

**Step 3: Write the minimal implementation**

Move shell prompt entrypoints onto a promptcatalog app service while keeping Wails file dialogs and prompt import session state in `cmd/skillflow`.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./cmd/skillflow -run 'TestPrompt' -count=1
```

Expected: PASS

### Task 3: Remove `core/prompt` and regenerate bindings

**Files:**
- Delete: `core/prompt/storage.go`
- Delete: `core/prompt/storage_test.go`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`
- Modify: `cmd/skillflow/frontend/wailsjs/go/models.ts`

**Step 1: Write the failing test/build check**

Verify the repo no longer compiles until all old imports and generated bindings are updated.

**Step 2: Run checks to verify they fail**

Run:

```bash
go test ./core/... ./cmd/skillflow -run '^$' -count=1
```

Expected: FAIL until remaining prompt imports and generated types are switched.

**Step 3: Write the minimal implementation**

Delete `core/prompt`, regenerate Wails bindings, and align remaining imports to `promptcatalog`.

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

Document that `promptcatalog` is now extracted alongside `skillcatalog`.

**Step 2: Run final verification**

Run:

```bash
rg -n '\"github.com/shinerio/skillflow/core/prompt\"|core/prompt\\b' core cmd/skillflow --glob '!core/prompt/**'
git status --short
```

Expected:

- no remaining active `core/prompt` code references
- worktree contains only intended promptcatalog migration changes
