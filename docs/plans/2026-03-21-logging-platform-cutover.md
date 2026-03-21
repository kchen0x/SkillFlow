# Logging Platform Cutover Implementation Plan

**Goal:** Move the reusable file logger from `core/applog` to `core/platform/logging`.

**Architecture:** Keep the logger behavior and API unchanged, but relocate it to `platform/` because it is a business-agnostic technical capability. Update shell imports and tests, then remove the old package.

**Tech Stack:** Go 1.23, Wails v2 backend

---

### Task 1: Move logging package

**Files:**
- Create: `core/platform/logging/logger.go`
- Create: `core/platform/logging/logger_test.go`
- Delete: `core/applog/logger.go`
- Delete: `core/applog/logger_test.go`

1. Copy the logger implementation and tests into `core/platform/logging`.
2. Keep exported API and rotation behavior unchanged.
3. Remove the old `core/applog` files.

### Task 2: Switch shell imports

**Files:**
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/app_log.go`
- Modify: `cmd/skillflow/process_helper.go`

1. Change shell code to import `core/platform/logging`.
2. Keep logger construction and level handling unchanged.

### Task 3: Update architecture progress docs and verify

**Files:**
- Modify: `docs/architecture/README.md`
- Modify: `docs/architecture/README_zh.md`
- Modify: `docs/architecture/migration.md`
- Modify: `docs/architecture/migration_zh.md`

**Run:**
- `go test ./core/... -count=1`
- `go test ./cmd/skillflow -count=1`
- `rg -n 'github.com/shinerio/skillflow/core/applog' core cmd/skillflow -g'*.go'`

**Expected:**
- tests pass
- no production Go code imports `core/applog`
