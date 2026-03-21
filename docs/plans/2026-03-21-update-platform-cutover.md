# Update Platform Cutover Implementation Plan

**Goal:** Move the GitHub skill-update checker from `core/update` to `core/platform/update`.

**Architecture:** Keep the checker API and behavior unchanged, but relocate it to `platform/` because it is a business-agnostic external update primitive. Update tests and docs, then remove the old package.

**Tech Stack:** Go 1.23 backend

---

### Task 1: Move update checker package

**Files:**
- Create: `core/platform/update/checker.go`
- Create: `core/platform/update/checker_test.go`
- Delete: `core/update/checker.go`
- Delete: `core/update/checker_test.go`

1. Copy the checker implementation and tests into `core/platform/update`.
2. Keep exported API and HTTP behavior unchanged.
3. Remove the old `core/update` files.

### Task 2: Update architecture progress docs

**Files:**
- Modify: `docs/architecture/README.md`
- Modify: `docs/architecture/README_zh.md`
- Modify: `docs/architecture/migration.md`
- Modify: `docs/architecture/migration_zh.md`

1. Mark `core/update` as migrated to `core/platform/update`.
2. Keep the progress summary aligned with the actual codebase.

### Task 3: Verify cutover

**Run:**
- `go test ./core/... -count=1`
- `go test ./cmd/skillflow -count=1`
- `rg -n 'github.com/shinerio/skillflow/core/update' core cmd/skillflow -g'*.go'`

**Expected:**
- tests pass
- no production Go code imports `core/update`
