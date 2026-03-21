# Registry Shell Cutover Implementation Plan

**Goal:** Remove `core/registry` and relocate adapter/provider registration to the Wails shell composition root.

**Architecture:** Keep lookup state in `cmd/skillflow`, with agent gateway registration in `adapters.go` and backup provider registration in `providers.go`. Leave bounded context service constructors unchanged by continuing to inject resolvers from the shell.

**Tech Stack:** Go 1.23, Wails v2 backend, DDD modular monolith

---

### Task 1: Move shell wiring state out of `core/registry`

**Files:**
- Modify: `cmd/skillflow/adapters.go`
- Modify: `cmd/skillflow/providers.go`
- Delete: `core/registry/registry.go`

1. Add shell-local maps/helpers for registered agent gateways and cloud provider factories.
2. Switch registration and lookup call sites to those shell-local helpers.
3. Delete the old `core/registry` package file once no imports remain.

### Task 2: Update transport adapters and architecture docs

**Files:**
- Modify: `cmd/skillflow/app.go`
- Modify: `docs/architecture/migration.md`
- Modify: `docs/architecture/migration_zh.md`
- Create: `docs/plans/2026-03-21-registry-shell-cutover-design.md`

1. Remove remaining `core/registry` imports from shell transport code.
2. Update migration progress notes so `core/registry` is marked as removed.
3. Keep a short design and implementation record under `docs/plans/`.

### Task 3: Verify cutover

**Run:**
- `go test ./core/... -count=1`
- `go test ./cmd/skillflow -count=1`
- `rg -n 'core/registry' core cmd/skillflow -g'*.go'`

**Expected:**
- tests pass
- no production Go code imports `core/registry`
