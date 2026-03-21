# Logicalkey Shared Cutover Implementation Plan

**Goal:** Move the reusable skill logical-key helpers from `core/skillkey` to `core/shared/logicalkey`.

**Architecture:** Keep the helper API and behavior unchanged, but relocate it to `shared/` because these identifiers are part of the stable shared kernel used across contexts. Update imports and tests, then remove the old package.

**Tech Stack:** Go 1.23 backend

---

### Task 1: Move logical-key package

**Files:**
- Create: `core/shared/logicalkey/logicalkey.go`
- Create: `core/shared/logicalkey/logicalkey_test.go`
- Delete: `core/skillkey/skillkey.go`
- Delete: `core/skillkey/skillkey_test.go`

1. Copy the logical-key implementation and tests into `core/shared/logicalkey`.
2. Keep exported API and matching semantics unchanged.
3. Remove the old `core/skillkey` files.

### Task 2: Switch imports across contexts and shell

**Files:**
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/skill_state.go`
- Modify: `core/agentintegration/app/service.go`
- Modify: `core/platform/update/checker.go`
- Modify: `core/skillcatalog/domain/installed_skill.go`
- Modify: `core/skillcatalog/app/query/installed_index.go`
- Modify: related tests

1. Change all imports to `core/shared/logicalkey`.
2. Leave correlation and matching behavior unchanged.

### Task 3: Update architecture progress docs and verify

**Files:**
- Modify: `docs/architecture/README.md`
- Modify: `docs/architecture/README_zh.md`
- Modify: `docs/architecture/migration.md`
- Modify: `docs/architecture/migration_zh.md`

**Run:**
- `go test ./core/... -count=1`
- `go test ./cmd/skillflow -count=1`
- `rg -n 'github.com/shinerio/skillflow/core/skillkey' core cmd/skillflow -g'*.go'`

**Expected:**
- tests pass
- no production Go code imports `core/skillkey`
