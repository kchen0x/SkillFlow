# Backup DDD Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extract backup execution semantics, snapshot comparison, and provider adapters from the flat `core/backup` package and `cmd/skillflow` into `core/backup/app`, `domain`, and `infra`, while keeping post-restore cross-context compensation in the shell.

**Architecture:** Create `core/backup/domain`, `app`, `infra/provider`, and `infra/snapshot`. Move provider metadata and result types into `domain`, move providers and snapshot helpers into `infra`, move backup and restore use cases into `app`, and keep Wails events plus post-restore cross-context rebuild in `cmd/skillflow`.

**Tech Stack:** Go 1.23, Wails v2 backend, testify unit tests, markdown docs

---

### Task 1: Create failing tests for the new `backup` package tree

**Files:**
- Create: `core/backup/domain/snapshot_test.go`
- Create: `core/backup/infra/snapshot/path_filter_test.go`
- Create: `core/backup/infra/snapshot/git_migration_test.go`
- Create: `core/backup/infra/provider/git_provider_test.go`
- Create: `core/backup/app/service_test.go`

**Step 1: Write the failing tests**

Port and add behavior coverage for:

- snapshot build/load/save/diff behavior
- backup path exclusion rules
- nested `.git` migration behavior
- git provider sync/restore/conflict behavior
- app-level run/list/restore/resolve-conflict flows

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./core/backup/... -count=1
```

Expected: FAIL because the new package tree and moved symbols do not exist yet.

**Step 3: Write the minimal implementation**

Create the new `backup` package tree and implement only enough code for those tests to pass.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./core/backup/... -count=1
```

Expected: PASS

### Task 2: Switch shell backup entrypoints to `backup/app`

**Files:**
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/app_backup.go`
- Modify: `cmd/skillflow/providers.go`
- Modify: `core/registry/registry.go`
- Modify: relevant shell tests under `cmd/skillflow/*_test.go`

**Step 1: Write the failing tests**

Update shell-facing tests so they expect backup run/list/restore/conflict logic to depend on `backup/app` and `backup/domain` types rather than the old flat `core/backup` package.

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./cmd/skillflow -run 'Test(HandleRestoredCloudState|TestProxyConnection|Backup|GitConflict)' -count=1
```

Expected: FAIL until shell backup callers are moved to `backup/app`.

**Step 3: Write the minimal implementation**

Move shell backup entrypoints onto `backup/app` while keeping Wails event publication, last-result caching, and post-restore compensation in the shell.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./cmd/skillflow -count=1
```

Expected: PASS

### Task 3: Remove the old flat `core/backup` package layout and regenerate bindings

**Files:**
- Delete: old flat files in `core/backup/*.go` after their replacements exist
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`
- Modify: `cmd/skillflow/frontend/wailsjs/go/models.ts`

**Step 1: Write the failing compile/build check**

Verify the repo no longer compiles until all imports and generated bindings are switched.

**Step 2: Run checks to verify they fail**

Run:

```bash
go test ./core/... ./cmd/skillflow -run '^$' -count=1
```

Expected: FAIL until all `core/backup` callers are aligned to the new package tree.

**Step 3: Write the minimal implementation**

Delete the old flat files, regenerate Wails bindings, and align remaining types to the new `backup` package tree.

**Step 4: Run tests/build to verify they pass**

Run:

```bash
make generate
go test ./core/... -count=1
go test ./cmd/skillflow -count=1
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

Document that `backup` is now extracted and that the old flat `core/backup` layout has been replaced by `app/domain/infra`.

**Step 2: Run final verification**

Run:

```bash
rg -n 'package backup$' core/backup -g '*.go'
git status --short
```

Expected:

- `core/backup` contains only the new `app/domain/infra` layout
- worktree contains only intended backup migration changes
