# Skillsource DDD Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extract tracked repository and source-candidate semantics from `core/git` and `cmd/skillflow` into `core/skillsource`, move pure Git helpers into `core/platform/git`, switch shell callers, and remove `core/git` without changing user-visible behavior.

**Architecture:** Create `core/skillsource/domain`, `app`, `infra/repository`, and `infra/discovery`. Move repository tracking and candidate discovery into that context, move pure Git helpers into `core/platform/git`, and keep source-to-skillcatalog import plus UI-facing overlays in `cmd/skillflow` until `core/orchestration` and `core/readmodel` are introduced.

**Tech Stack:** Go 1.23, Wails v2 backend, testify unit tests, markdown docs

---

### Task 1: Create failing tests for the new `skillsource` and `platform/git` package trees

**Files:**
- Create: `core/skillsource/infra/repository/star_repo_storage_test.go`
- Create: `core/skillsource/infra/discovery/repo_scanner_test.go`
- Create: `core/skillsource/app/service_test.go`
- Create: `core/platform/git/client_test.go`

**Step 1: Write the failing tests**

Port and add behavior coverage for:

- tracked-repo storage and local-state split from `core/git/storage_test.go`
- source candidate scanning from `core/git/scanner_test.go`
- Git helper behavior from `core/git/client_test.go`
- track/list/untrack/refresh flows in a new `skillsource/app` service test

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./core/skillsource/... ./core/platform/git/... -count=1
```

Expected: FAIL because the new packages do not exist yet.

**Step 3: Write the minimal implementation**

Create the new `skillsource` and `platform/git` package trees and implement only enough code for those tests to pass.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./core/skillsource/... ./core/platform/git/... -count=1
```

Expected: PASS

### Task 2: Switch shell source-management entrypoints to `skillsource/app`

**Files:**
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/app_viewstate.go`
- Modify: `cmd/skillflow/skill_state.go`
- Modify: `cmd/skillflow/app_restore.go`
- Modify: relevant shell tests under `cmd/skillflow/*_test.go`
- Modify: `core/skillkey/skillkey.go`
- Modify: `core/update/checker.go`

**Step 1: Write the failing tests**

Update shell-facing tests so they expect starred-repo mutations and source-candidate reads to depend on `skillsource` and `platform/git` types rather than `core/git`.

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./cmd/skillflow -run 'Test(ListAllStarSkills|HandleRestoredCloudState|UpdateStarredRepo|UpdateAllStarredRepos)' -count=1
```

Expected: FAIL until the shell caller chain is moved onto `skillsource`.

**Step 3: Write the minimal implementation**

Move shell source-management entrypoints onto `skillsource/app` while keeping installed-state overlays, source import, backup scheduling, and restore orchestration in the shell.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./cmd/skillflow -count=1
```

Expected: PASS

### Task 3: Remove `core/git`, regenerate bindings, and clean imports

**Files:**
- Delete: `core/git/model.go`
- Delete: `core/git/storage.go`
- Delete: `core/git/scanner.go`
- Delete: `core/git/client.go`
- Delete: `core/git/*_test.go`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`
- Modify: `cmd/skillflow/frontend/wailsjs/go/models.ts`

**Step 1: Write the failing compile/build check**

Verify the repo no longer compiles until old imports and generated bindings are updated.

**Step 2: Run checks to verify they fail**

Run:

```bash
go test ./core/... ./cmd/skillflow -run '^$' -count=1
```

Expected: FAIL until every `core/git` caller is switched.

**Step 3: Write the minimal implementation**

Delete `core/git`, regenerate Wails bindings, and align remaining source-facing types to `skillsource` and `platform/git`.

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

Document that `skillsource` is now extracted, `core/git` has been removed, and pure Git helpers live under `core/platform/git`.

**Step 2: Run final verification**

Run:

```bash
rg -n '\"github.com/shinerio/skillflow/core/git\"|core/git\\b' core cmd/skillflow --glob '!docs/**'
git status --short
```

Expected:

- no remaining active `core/git` code references
- worktree contains only intended `skillsource` migration changes
