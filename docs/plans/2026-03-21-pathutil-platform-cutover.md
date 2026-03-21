# Pathutil Platform Cutover Implementation Plan

**Goal:** Move portable path normalization helpers from `core/pathutil` to `core/platform/pathutil`.

**Architecture:** Keep the path helper API unchanged, but relocate it to `platform/` because it is a pure technical capability used by repository implementations. Update repository imports, then remove the old package.

**Tech Stack:** Go 1.23 backend

---

### Task 1: Move path utility package

**Files:**
- Create: `core/platform/pathutil/portable.go`
- Create: `core/platform/pathutil/portable_test.go`
- Delete: `core/pathutil/portable.go`
- Delete: `core/pathutil/portable_test.go`

1. Copy the helper implementation and tests into `core/platform/pathutil`.
2. Keep exported API and behavior unchanged.
3. Remove the old `core/pathutil` files.

### Task 2: Switch repository imports

**Files:**
- Modify: `core/skillcatalog/infra/repository/filesystem_storage.go`
- Modify: `core/skillsource/infra/repository/star_repo_storage.go`

1. Change repository imports to `core/platform/pathutil`.
2. Leave persistence semantics unchanged.

### Task 3: Update architecture progress docs and verify

**Files:**
- Modify: `docs/architecture/README.md`
- Modify: `docs/architecture/README_zh.md`
- Modify: `docs/architecture/migration.md`
- Modify: `docs/architecture/migration_zh.md`

**Run:**
- `go test ./core/... -count=1`
- `go test ./cmd/skillflow -count=1`
- `rg -n 'github.com/shinerio/skillflow/core/pathutil' core cmd/skillflow -g'*.go'`

**Expected:**
- tests pass
- no production Go code imports `core/pathutil`
