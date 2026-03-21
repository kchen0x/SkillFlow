# Background Memory Trim Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Reduce app-owned memory while SkillFlow is hidden or inactive by unmounting the routed page tree after a delay and reloading it on resume.

**Architecture:** Emit stable backend window-visibility events with change deduplication, then let the frontend combine those events with browser focus/visibility signals. After a background delay, App unmounts the routed page subtree so page-local arrays and prompt content are released. On resume, the subtree mounts again and reloads current route data.

**Tech Stack:** Go 1.23, Wails desktop runtime, React 18 + TypeScript, Node test runner, Go tests

---

### Task 1: Add failing frontend tests for background trim state logic

**Files:**
- Create: `cmd/skillflow/frontend/src/lib/appActivity.ts`
- Create: `cmd/skillflow/frontend/tests/appActivity.test.mjs`
- Modify: `cmd/skillflow/frontend/package.json`

**Step 1: Write the failing test**

- Cover:
  - foreground requires window visible + document visible + focused
  - background trim activates only after an explicit timeout event
  - returning to foreground clears trimmed state and bumps resume token

**Step 2: Run test to verify it fails**

Run: `npm run test:unit`

Expected: FAIL because `appActivity.ts` does not exist yet.

**Step 3: Write minimal implementation**

- Add a small reducer/state helper that drives trim scheduling and resume behavior.

**Step 4: Run test to verify it passes**

Run: `npm run test:unit`

Expected: PASS

### Task 2: Add failing backend tests for visibility-event dedupe

**Files:**
- Create: `cmd/skillflow/app_visibility_test.go`
- Modify: `cmd/skillflow/app.go`
- Modify: `core/notify/model.go`

**Step 1: Write the failing test**

- Cover:
  - visible -> hidden publishes one event
  - duplicate hidden does not republish
  - hidden -> visible publishes one event

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/skillflow -run 'TestWindowVisibility'`

Expected: FAIL because the helper/event does not exist yet.

**Step 3: Write minimal implementation**

- Add a dedicated app visibility event type.
- Add App helper to publish only when visibility actually changes.

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/skillflow -run 'TestWindowVisibility'`

Expected: PASS

### Task 3: Wire frontend/background trim behavior

**Files:**
- Modify: `cmd/skillflow/frontend/src/App.tsx`
- Modify: `cmd/skillflow/frontend/src/lib/wailsEvents.ts`
- Modify: `cmd/skillflow/tray_darwin_callbacks.go`
- Modify: `cmd/skillflow/window_darwin.go`
- Modify: `cmd/skillflow/window_other.go`

**Step 1: Implement minimal UI behavior**

- Subscribe to backend `app.window.visibility.changed`.
- Track browser `focus`, `blur`, and `visibilitychange`.
- After the configured background delay, unmount the routed page subtree and show a lightweight placeholder.
- On foreground, remount the subtree and reload the current route.

**Step 2: Run targeted tests**

Run: `npm run test:unit`
Run: `go test ./cmd/skillflow -run 'TestWindowVisibility'`

Expected: PASS

### Task 4: Update docs and verify the full path

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `docs/architecture/README.md`
- Modify: `docs/architecture/README_zh.md`
- Modify: `README.md`
- Modify: `README_zh.md`

**Step 1: Document behavior**

- Explain that hidden/inactive windows release routed page memory after a delay and reload current route data on resume.

**Step 2: Run full verification**

Run: `npm run test:unit`
Run: `npm run build`
Run: `go test ./cmd/skillflow`
Run: `go test ./core/viewstate/...`

Expected: PASS
