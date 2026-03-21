# Eventbus Platform Cutover Implementation Plan

**Goal:** Move the reusable event hub from `core/notify` to `core/platform/eventbus`.

**Architecture:** Keep the event hub API and event payload model unchanged, but relocate the package to `platform/` because it is a business-agnostic runtime capability. Update shell imports and tests, then remove the old package.

**Tech Stack:** Go 1.23, Wails v2 backend

---

### Task 1: Move event bus package

**Files:**
- Create: `core/platform/eventbus/hub.go`
- Create: `core/platform/eventbus/model.go`
- Create: `core/platform/eventbus/hub_test.go`
- Delete: `core/notify/hub.go`
- Delete: `core/notify/model.go`
- Delete: `core/notify/hub_test.go`

1. Copy the event hub and event model into `core/platform/eventbus`.
2. Keep exported types and constants behavior-compatible.
3. Remove the old `core/notify` files.

### Task 2: Switch shell imports

**Files:**
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/app_update.go`
- Modify: `cmd/skillflow/app_visibility.go`
- Modify: `cmd/skillflow/events.go`
- Modify: `cmd/skillflow/app_visibility_test.go`

1. Change shell code and tests to import `core/platform/eventbus`.
2. Keep event publishing behavior and payloads unchanged.

### Task 3: Update architecture progress docs and verify

**Files:**
- Modify: `docs/architecture/README.md`
- Modify: `docs/architecture/README_zh.md`
- Modify: `docs/architecture/migration.md`
- Modify: `docs/architecture/migration_zh.md`

**Run:**
- `go test ./core/... -count=1`
- `go test ./cmd/skillflow -count=1`
- `rg -n 'github.com/shinerio/skillflow/core/notify' core cmd/skillflow -g'*.go'`

**Expected:**
- tests pass
- no production Go code imports `core/notify`
