# Skillcatalog DDD Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extract installed skill truth from `core/skill` into the first DDD bounded context `core/skillcatalog`, switch backend callers to the new package tree, and remove the old package without changing user-facing behavior.

**Architecture:** Create `skillcatalog/domain`, `skillcatalog/app`, and `skillcatalog/infra/repository`, then migrate the installed-skill model, filesystem persistence, and installed-index query logic into those layers. Update `cmd/skillflow` to use the application layer for core installed-skill CRUD/query entrypoints while allowing not-yet-migrated backend packages to depend on `skillcatalog/domain` types.

**Tech Stack:** Go 1.23, Wails v2 backend, testify unit tests, markdown docs

---

### Task 1: Create failing tests for the new `skillcatalog` package tree

**Files:**
- Create: `core/skillcatalog/domain/installed_skill_test.go`
- Create: `core/skillcatalog/domain/validator_test.go`
- Create: `core/skillcatalog/app/query/installed_index_test.go`
- Create: `core/skillcatalog/infra/repository/filesystem_storage_test.go`

**Step 1: Write the failing tests**

Port the current `core/skill` behavior checks so the new package tree expects:

- `InstalledSkill` source helpers and update detection
- case-insensitive `skill.md` validation
- installed-index logical-key/content-key/name fallback behavior
- filesystem storage import/delete/category/meta persistence behavior

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./core/skillcatalog/... -count=1
```

Expected: FAIL because the new packages do not exist yet.

**Step 3: Write the minimal implementation**

Create the new package directories and move the minimal production code needed for those tests to pass.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./core/skillcatalog/... -count=1
```

Expected: PASS

### Task 2: Introduce `skillcatalog/app` services and migrate shell-owned skill CRUD/query calls

**Files:**
- Create: `core/skillcatalog/app/service.go`
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/app_viewstate.go`
- Modify: `cmd/skillflow/skill_state.go`
- Modify: `cmd/skillflow/app_restore.go`
- Test: `cmd/skillflow/app_viewstate_test.go`
- Test: `cmd/skillflow/app_restore_test.go`
- Test: `cmd/skillflow/app_skill_update_test.go`

**Step 1: Write the failing tests**

Add or update tests so shell code expects:

- app startup still wires installed-skill storage correctly
- installed skill listing still works through viewstate-backed paths
- restore compensation still resolves installed skills after the new extraction
- skill update flows still mutate installed metadata

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./cmd/skillflow -run 'Test(App|Storage|Restore|SkillUpdate|Viewstate)' -count=1
```

Expected: FAIL because `cmd/skillflow` still imports `core/skill` and has no `skillcatalog/app` boundary.

**Step 3: Write the minimal implementation**

Introduce a pragmatic app service that exposes the existing installed-skill CRUD/query capabilities and switch the shell code to it where the shell owns transport concerns.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./cmd/skillflow -run 'Test(App|Storage|Restore|SkillUpdate|Viewstate)' -count=1
```

Expected: PASS

### Task 3: Switch remaining backend callers and remove `core/skill`

**Files:**
- Modify: `core/sync/adapter.go`
- Modify: `core/sync/filesystem_adapter.go`
- Modify: `core/update/checker.go`
- Delete: `core/skill/model.go`
- Delete: `core/skill/validator.go`
- Delete: `core/skill/meta.go`
- Delete: `core/skill/index.go`
- Delete: `core/skill/storage.go`
- Delete: `core/skill/model_test.go`
- Delete: `core/skill/validator_test.go`
- Delete: `core/skill/index_test.go`
- Delete: `core/skill/storage_test.go`

**Step 1: Write the failing tests**

Ensure the migrated tests and caller packages cover:

- agent adapter push/pull using the new installed-skill type
- update checker using the new installed-skill type
- no remaining `core/skill` imports

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./core/... ./cmd/skillflow -count=1
```

Expected: FAIL until all imports and old package references are removed.

**Step 3: Write the minimal implementation**

Switch remaining imports, delete `core/skill`, and keep package responsibilities under `skillcatalog`.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./core/... ./cmd/skillflow -count=1
```

Expected: PASS

### Task 4: Final verification and migration hygiene

**Files:**
- Modify as needed: `docs/plans/2026-03-21-skillcatalog-ddd-refactor-design.md`
- Modify as needed: `docs/plans/2026-03-21-skillcatalog-ddd-refactor.md`

**Step 1: Run repository checks**

Run:

```bash
rg -n 'core/skill\\b' core cmd/skillflow
go test ./core/... ./cmd/skillflow -count=1
git status --short
```

Expected:

- no remaining `core/skill` imports
- tests pass
- worktree contains only intended migration changes

