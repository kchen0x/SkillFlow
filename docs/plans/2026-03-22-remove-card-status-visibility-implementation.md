# Remove Card Status Visibility Customization Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Remove user-configurable card-status visibility from Settings and persisted config, while keeping all pages on the existing built-in default status policy.

**Architecture:** Delete the Settings UI and shared-config field, then collapse runtime behavior onto a single fixed-default visibility model. Use startup cutover to remove legacy `skillStatusVisibility` from `config.json` so the app no longer carries invisible state that users cannot control.

**Tech Stack:** Go, React, TypeScript, Wails, Markdown docs

---

### Task 1: Remove the Settings UI and fixed-default the frontend visibility provider

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/Settings.tsx`
- Modify: `cmd/skillflow/frontend/src/contexts/SkillStatusVisibilityContext.tsx`
- Modify: `cmd/skillflow/frontend/src/lib/skillStatusVisibility.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`
- Modify: existing frontend tests that mention status-visibility controls or config syncing

**Step 1: Write the failing test**

Reuse or adjust existing tests instead of introducing extra standalone feature tests where possible.

Make the current frontend verification fail by updating the relevant expectations so they reflect the removed feature. At minimum, remove or invert any source-level expectations that depend on:

- `Card Status Visibility` strings in `Settings.tsx`
- `toggleSkillStatusForPage`
- `normalizeSkillStatusVisibility(cfg?.skillStatusVisibility)` in the Settings page

If no existing test covers this directly, add only the smallest source-level assertion needed.

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `FAIL` because the Settings page still renders the removed section or still depends on config-driven visibility state.

**Step 3: Write minimal implementation**

Implement:

- remove the Settings General section for card-status visibility
- remove related i18n entries
- remove page-level config syncing for `skillStatusVisibility`
- simplify the visibility context so it serves fixed defaults only
- keep `useSkillStatusVisibility(page)` working for consumers without changing page behavior

Prefer the least invasive path: retain the default visibility constants if existing consumers still use them, but stop exposing any user customization path.

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `PASS`

**Step 5: Run build verification**

Run:

```bash
cd cmd/skillflow/frontend && npm run build
```

Expected: successful TypeScript compile and production build.

**Step 6: Commit**

Run:

```bash
git add cmd/skillflow/frontend/src/pages/Settings.tsx cmd/skillflow/frontend/src/contexts/SkillStatusVisibilityContext.tsx cmd/skillflow/frontend/src/lib/skillStatusVisibility.ts cmd/skillflow/frontend/src/i18n/en.ts cmd/skillflow/frontend/src/i18n/zh.ts
git commit -m "refactor: remove status visibility settings ui"
```

### Task 2: Remove persisted config support and cut over legacy shared config

**Files:**
- Modify: `core/config/model.go`
- Modify: `core/config/service.go`
- Modify: `core/config/service_test.go`
- Modify: `core/platform/upgrade/upgrade.go`
- Modify: `core/platform/upgrade/config_terms_test.go`

**Step 1: Write the failing test**

Reuse existing config and upgrade tests instead of adding a new test suite unless necessary.

Update current expectations so they fail if:

- `config.AppConfig` still persists `skillStatusVisibility`
- `config.json` still contains `skillStatusVisibility`
- startup upgrade still preserves or migrates that field instead of deleting it

The existing tests most likely to adjust are:

- `TestLoadDefaultConfig`
- `TestSkillStatusVisibilityPersistsInSharedConfig`
- `TestSkillStatusVisibilityDropsStatusesOutsidePageDefaultPolicy`
- upgrade terminology tests in `core/platform/upgrade/config_terms_test.go`

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./core/config ./core/platform/upgrade
```

Expected: `FAIL` because config and upgrade code still carry `skillStatusVisibility`.

**Step 3: Write minimal implementation**

Implement:

- remove `SkillStatusVisibility` from `config.AppConfig`
- remove shared-config load/save normalization for that field
- delete the field during startup upgrade when present
- remove now-unneeded visibility migration logic if it becomes dead code

Keep the rest of config and terminology migration behavior unchanged.

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./core/config ./core/platform/upgrade
```

Expected: `ok`

**Step 5: Commit**

Run:

```bash
git add core/config/model.go core/config/service.go core/config/service_test.go core/platform/upgrade/upgrade.go core/platform/upgrade/config_terms_test.go
git commit -m "refactor: remove status visibility config"
```

### Task 3: Update config, feature, and architecture docs

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `docs/config.md`
- Modify: `docs/config_zh.md`
- Modify: `docs/architecture/use-cases.md`
- Modify: `docs/architecture/use-cases_zh.md`
- Modify: `docs/architecture/runtime-and-storage.md`
- Modify: `docs/architecture/runtime-and-storage_zh.md`

**Step 1: Write the failing test**

Use direct document checks to identify stale text before editing:

```bash
rg -n "Card status visibility|skillStatusVisibility|card-status visibility|readmodel/preferences" docs
rg -n "卡片状态显示|skillStatusVisibility|状态显示策略|readmodel/preferences" docs
```

Expected: matches show stale docs that still describe the removed feature.

**Step 2: Write minimal documentation updates**

Update docs so they reflect:

- no Settings option for status-visibility customization
- no persisted `skillStatusVisibility` field in config docs
- no architecture claim that status-visibility preferences are stored in readmodel/preferences

Preserve documentation for the default status behavior where still relevant, but remove any customization wording.

**Step 3: Verify the docs**

Run:

```bash
rg -n "Card status visibility|skillStatusVisibility" docs
rg -n "卡片状态显示|skillStatusVisibility" docs
```

Expected: no stale customization references remain, except where removal/cutover is being described intentionally in plan docs.

**Step 4: Commit**

Run:

```bash
git add docs/features.md docs/features_zh.md docs/config.md docs/config_zh.md docs/architecture/use-cases.md docs/architecture/use-cases_zh.md docs/architecture/runtime-and-storage.md docs/architecture/runtime-and-storage_zh.md
git commit -m "docs: remove status visibility customization docs"
```

### Task 4: Run final verification before handoff

**Files:**
- No additional file changes expected

**Step 1: Run full backend verification**

Run:

```bash
go test ./core/...
```

Expected: `ok`

**Step 2: Run full frontend verification**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
cd cmd/skillflow/frontend && npm run build
```

Expected:

- unit tests pass
- production build succeeds

**Step 3: Review final diff**

Run:

```bash
git status --short
git diff -- core/config/model.go core/config/service.go core/config/service_test.go core/platform/upgrade/upgrade.go core/platform/upgrade/config_terms_test.go cmd/skillflow/frontend/src/pages/Settings.tsx cmd/skillflow/frontend/src/contexts/SkillStatusVisibilityContext.tsx cmd/skillflow/frontend/src/lib/skillStatusVisibility.ts cmd/skillflow/frontend/src/i18n/en.ts cmd/skillflow/frontend/src/i18n/zh.ts docs/features.md docs/features_zh.md docs/config.md docs/config_zh.md docs/architecture/use-cases.md docs/architecture/use-cases_zh.md docs/architecture/runtime-and-storage.md docs/architecture/runtime-and-storage_zh.md
```

Expected: only the planned removal files are changed.

**Step 4: Commit**

Run:

```bash
git add docs/plans/2026-03-22-remove-card-status-visibility-design.md docs/plans/2026-03-22-remove-card-status-visibility-implementation.md
git commit -m "docs: add status visibility removal plan"
```
