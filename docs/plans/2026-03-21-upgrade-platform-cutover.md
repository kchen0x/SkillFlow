# Upgrade Platform Cutover Implementation Plan

**Goal:** Move startup cutover code from `core/upgrade` to `core/platform/upgrade`.

**Architecture:** Keep the startup upgrade API and behavior unchanged, but relocate it to `platform/` because persisted-schema cutover is a business-agnostic technical capability. Update shell imports and tests, then remove the old package.

**Tech Stack:** Go 1.23 backend

---

### Task 1: Move upgrade package

**Files:**
- Create: `core/platform/upgrade/upgrade.go`
- Create: `core/platform/upgrade/config_terms_test.go`
- Delete: `core/upgrade/upgrade.go`
- Delete: `core/upgrade/config_terms_test.go`

1. Copy the cutover implementation and tests into `core/platform/upgrade`.
2. Keep exported API and file-rewrite behavior unchanged.
3. Remove the old `core/upgrade` files.

### Task 2: Switch shell imports

**Files:**
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/app_startup_test.go`

1. Change shell startup code and tests to import `core/platform/upgrade`.
2. Keep startup ordering and behavior unchanged.

### Task 3: Update architecture progress docs and verify

**Files:**
- Modify: `docs/architecture/README.md`
- Modify: `docs/architecture/README_zh.md`
- Modify: `docs/architecture/migration.md`
- Modify: `docs/architecture/migration_zh.md`

**Run:**
- `go test ./core/... -count=1`
- `go test ./cmd/skillflow -count=1`
- `rg -n 'github.com/shinerio/skillflow/core/upgrade' core cmd/skillflow -g'*.go'`

**Expected:**
- tests pass
- no production Go code imports `core/upgrade`
