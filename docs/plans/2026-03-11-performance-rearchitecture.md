# Performance Rearchitecture Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rebuild the startup, derived-state, and list-rendering pipeline so SkillFlow becomes responsive earlier, uses less CPU and memory, and still converges automatically to cloud-synced truth without user interaction.

**Architecture:** Introduce a local-only derived view-state layer under `core/viewstate`, move expensive tool-presence and starred-state derivation behind cacheable snapshots plus a reconciliation scheduler, and update the frontend to consume snapshots through targeted subscriptions rather than broad full-page reloads. Pair that with per-card render-cost removal and adaptive animation degradation so the UI remains smooth at higher item counts.

**Tech Stack:** Go 1.23, Wails desktop backend, React 18 + TypeScript, Framer Motion, filesystem-backed local storage, Go tests, frontend unit tests

---

### Task 1: Add backend timing coverage for current hotspots

**Files:**
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/skill_state.go`

**Step 1: Write the failing test**

- Add a focused backend test in `cmd/skillflow/app_skill_update_test.go` or a new `cmd/skillflow/app_viewstate_test.go` that expects timing/log helper paths to be callable without changing result data.
- The assertion should verify the returned installed-skill entries are unchanged while timing hooks are active.

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/skillflow -run 'TestListSkills'`

Expected: FAIL because the timing helper or test seam does not exist yet.

**Step 3: Write minimal implementation**

- Add lightweight timing around:
  - `ListSkills()`
  - `buildToolPresenceIndex()`
  - `ListAllStarSkills()`
- Keep logs stable and searchable.

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/skillflow -run 'TestListSkills'`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "chore: add performance timing hooks"
```

### Task 2: Create failing tests for local-only derived cache semantics

**Files:**
- Create: `core/viewstate/model.go`
- Create: `core/viewstate/cache.go`
- Create: `core/viewstate/fingerprint.go`
- Create: `core/viewstate/manager_test.go`

**Step 1: Write the failing test**

- Add tests that cover:
  - reading a valid cached installed-skills snapshot
  - invalidating cache on schema-version mismatch
  - invalidating cache on fingerprint mismatch
  - corrupt cache file falling back to rebuild
- Use temp directories only.

**Step 2: Run test to verify it fails**

Run: `go test ./core/viewstate/...`

Expected: FAIL because the package and manager do not exist yet.

**Step 3: Write minimal implementation**

- Define cache record models with:
  - schema version
  - fingerprint
  - built-at timestamp
  - payload
- Implement atomic read/write helpers and validation.

**Step 4: Run test to verify it passes**

Run: `go test ./core/viewstate/...`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: add local derived viewstate cache"
```

### Task 3: Create failing tests for incremental tool-presence indexing

**Files:**
- Create: `core/viewstate/tool_presence_test.go`
- Modify: `cmd/skillflow/skill_state.go`

**Step 1: Write the failing test**

- Add tests that:
  - build tool presence for two tools
  - mutate only one tool push directory
  - expect only that tool to be rescanned on rebuild
  - preserve identical pushed-tool results versus current behavior

**Step 2: Run test to verify it fails**

Run: `go test ./core/viewstate/... -run 'TestToolPresence'`

Expected: FAIL because incremental presence logic does not exist yet.

**Step 3: Write minimal implementation**

- Add per-tool directory fingerprinting and an incremental presence cache.
- Keep output keyed by logical skill identity, not name or absolute path.

**Step 4: Run test to verify it passes**

Run: `go test ./core/viewstate/... -run 'TestToolPresence'`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: cache incremental tool presence state"
```

### Task 4: Add failing app-level tests for snapshot-first `ListSkills()`

**Files:**
- Create: `cmd/skillflow/app_viewstate_test.go`
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/skill_state.go`

**Step 1: Write the failing test**

- Add tests that verify:
  - `ListSkills()` returns the cached snapshot when valid
  - stale snapshot schedules reconcile and then converges to rebuilt truth
  - cached and rebuilt results match existing logical semantics

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/skillflow -run 'TestListSkillsUsesViewState'`

Expected: FAIL because the app does not yet use the viewstate manager.

**Step 3: Write minimal implementation**

- Add viewstate manager wiring to `App`.
- Replace direct `buildToolPresenceIndex()` work on the `ListSkills()` critical path with snapshot-first reads.
- Keep fallback full rebuild when no usable cache exists.

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/skillflow -run 'TestListSkillsUsesViewState'`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: serve installed skills from derived snapshots"
```

### Task 5: Add failing tests for reconcile scheduler and no-interaction eventual consistency

**Files:**
- Modify: `cmd/skillflow/app.go`
- Create: `cmd/skillflow/app_reconcile_test.go`
- Modify: `cmd/skillflow/events.go`

**Step 1: Write the failing test**

- Add tests that:
  - simulate restore or starred-repo refresh changing disk truth
  - keep the page passive
  - expect backend reconciliation to publish `view.skills.changed` and/or `view.starred.changed`
  - verify the rebuilt snapshot reflects new truth without any explicit frontend reload call

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/skillflow -run 'TestReconcileScheduler'`

Expected: FAIL because the scheduler and view events do not exist yet.

**Step 3: Write minimal implementation**

- Add a reconcile scheduler with:
  - coalesced invalidation requests
  - per-slice rebuild serialization
  - explicit rebuild triggers after restore, skill mutation, starred refresh, and config changes
- Publish `view.skills.changed`, `view.toolPresence.changed`, `view.starred.changed`, and `view.config.changed` when slices are ready.

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/skillflow -run 'TestReconcileScheduler'`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: add reconcile scheduler for view state"
```

### Task 6: Add failing tests for low-frequency fingerprint polling

**Files:**
- Modify: `core/viewstate/manager_test.go`
- Modify: `cmd/skillflow/app.go`

**Step 1: Write the failing test**

- Add tests that emulate long-lived app runtime with changed fingerprints and expect the polling loop to schedule rebuilds.
- Ensure unchanged fingerprints do not trigger redundant work.

**Step 2: Run test to verify it fails**

Run: `go test ./core/viewstate/... ./cmd/skillflow -run 'Test(ViewStatePolling|ReconcilePolling)'`

Expected: FAIL because the polling loop does not exist yet.

**Step 3: Write minimal implementation**

- Add a low-frequency fingerprint check loop.
- Limit it to lightweight dependency checks; do not rescan heavy directories unless fingerprints changed.

**Step 4: Run test to verify it passes**

Run: `go test ./core/viewstate/... ./cmd/skillflow -run 'Test(ViewStatePolling|ReconcilePolling)'`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: add eventual-consistency polling for view state"
```

### Task 7: Add failing frontend tests for snapshot subscription flows

**Files:**
- Create: `cmd/skillflow/frontend/src/lib/viewEvents.ts`
- Create: `cmd/skillflow/frontend/tests/viewEvents.test.mjs`
- Modify: `cmd/skillflow/frontend/src/pages/Dashboard.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/StarredRepos.tsx`

**Step 1: Write the failing test**

- Add unit coverage for a small view-event helper that:
  - subscribes to backend `view.*.changed` events
  - deduplicates redundant reload requests
  - ignores stale async completions

**Step 2: Run test to verify it fails**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: FAIL because the helper and page integration do not exist yet.

**Step 3: Write minimal implementation**

- Add a shared view-event helper.
- Convert the three pages to read current snapshots first, then subscribe to targeted updates instead of broad reload loops.
- Preserve existing visible behavior and empty/loading states.

**Step 4: Run test to verify it passes**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: subscribe pages to derived view updates"
```

### Task 8: Add failing frontend tests for card render-cost removal

**Files:**
- Modify: `cmd/skillflow/frontend/src/components/SkillStatusStrip.tsx`
- Create: `cmd/skillflow/frontend/tests/skillStatusStrip.test.mjs`

**Step 1: Write the failing test**

- Add tests that verify:
  - pushed-tool overflow collapses to `+N`
  - badges still render in single-line mode
  - the component no longer depends on runtime measurement to choose layout

**Step 2: Run test to verify it fails**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: FAIL because the component still uses runtime measurement behavior.

**Step 3: Write minimal implementation**

- Remove per-card `ResizeObserver` and width measurement.
- Replace with deterministic CSS truncation and capped icon rendering.

**Step 4: Run test to verify it passes**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "perf: remove per-card status measurement"
```

### Task 9: Add failing frontend tests for adaptive motion degradation

**Files:**
- Modify: `cmd/skillflow/frontend/src/App.tsx`
- Modify: `cmd/skillflow/frontend/src/lib/motionVariants.ts`
- Modify: `cmd/skillflow/frontend/src/pages/Dashboard.tsx`
- Create: `cmd/skillflow/frontend/tests/motionVariants.test.mjs`

**Step 1: Write the failing test**

- Add tests that verify:
  - large lists disable staggered card motion
  - route transitions degrade to lightweight fades
  - small lists preserve existing motion affordances

**Step 2: Run test to verify it fails**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: FAIL because the adaptive thresholds do not exist yet.

**Step 3: Write minimal implementation**

- Add threshold-driven motion helpers.
- Remove heavy route movement transitions from list pages.
- Keep visual feedback for small datasets.

**Step 4: Run test to verify it passes**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "perf: adapt motion to list density"
```

### Task 10: Add failing tests for hover metadata cancellation

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/Dashboard.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- Create: `cmd/skillflow/frontend/tests/hoverMeta.test.mjs`

**Step 1: Write the failing test**

- Add tests that verify:
  - only the latest hover request may update state
  - leaving a card before metadata returns prevents stale tooltip updates
  - repeated hover over the same skill can reuse memoized metadata

**Step 2: Run test to verify it fails**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: FAIL because hover requests are not yet cancellable or memoized.

**Step 3: Write minimal implementation**

- Add request sequence guards and small in-memory metadata caches on both pages.
- Preserve current tooltip behavior and delay.

**Step 4: Run test to verify it passes**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "perf: stabilize hover metadata loading"
```

### Task 11: Update architecture, config, and feature docs

**Files:**
- Modify: `docs/architecture.md`
- Modify: `docs/architecture_zh.md`
- Modify: `docs/config.md`
- Modify: `docs/config_zh.md`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `README.md`
- Modify: `README_zh.md`

**Step 1: Document the behavior**

- Document the new `core/viewstate` package and reconciliation flow.
- Document local-only cache files under `cache/` and explicitly note that they are derived, rebuildable, and excluded from sync.
- Document that UI state auto-refreshes after background sync and restore without requiring page interaction.
- Update feature-doc last-updated footers.

**Step 2: Run verification**

Run: `go test ./core/... ./cmd/skillflow`

Expected: PASS

**Step 3: Commit**

```bash
git add -A
git commit -m "docs: document performance rearchitecture"
```

### Task 12: Run full verification and compare performance baseline

**Files:**
- Reuse: `cmd/skillflow/app.go`
- Reuse: `cmd/skillflow/frontend/src/App.tsx`
- Reuse: `core/viewstate/*.go`

**Step 1: Run backend verification**

Run: `go test ./core/... ./cmd/skillflow`

Expected: PASS

**Step 2: Run frontend verification**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: PASS

**Step 3: Run production frontend build**

Run: `cd cmd/skillflow/frontend && npm run build`

Expected: PASS

**Step 4: Capture post-change timings**

- Record fresh timings for:
  - startup to first interactive paint
  - `ListSkills()` duration
  - dashboard first visible list time
  - route transition time on dense lists

**Step 5: Commit**

```bash
git add -A
git commit -m "chore: verify performance rearchitecture"
```
