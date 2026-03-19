# Default Young Theme Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make first-open theme default to `young` while preserving existing stored-theme behavior and legacy migration.

**Architecture:** Keep theme selection logic in `useTheme.ts`, but expose the initial-theme resolution as a small pure helper so it can be tested without rendering React. Update the frontend unit test script to compile that module and add a regression test for the no-storage fallback.

**Tech Stack:** React, TypeScript, Node test runner, project docs

---

### Task 1: Lock the expected default behavior with a test

**Files:**
- Modify: `cmd/skillflow/frontend/package.json`
- Create: `cmd/skillflow/frontend/tests/useTheme.test.mjs`
- Test: `cmd/skillflow/frontend/tests/useTheme.test.mjs`

- [ ] **Step 1: Write the failing test**

Add a Node test that imports the compiled `useTheme` module and asserts the storage resolver returns `young` when no keys are present.

- [ ] **Step 2: Run test to verify it fails**

Run: `npm run test:unit`
Expected: the new `useTheme` test fails because the helper does not exist yet or the fallback is still `dark`.

- [ ] **Step 3: Write minimal implementation**

Expose a small helper from `useTheme.ts` and change the empty-storage fallback to `young`.

- [ ] **Step 4: Run test to verify it passes**

Run: `npm run test:unit`
Expected: all frontend unit tests pass.

### Task 2: Sync user-facing docs

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `README.md`
- Modify: `README_zh.md`

- [ ] **Step 1: Update feature docs**

Replace descriptions that say `Dark` is the default with `Young`, and update the last-updated date in both feature docs.

- [ ] **Step 2: Update README summaries**

Adjust the high-level desktop experience row to mention `Young` as the default theme.

- [ ] **Step 3: Re-run focused verification**

Run: `npm run test:unit`
Expected: PASS after documentation-only edits do not affect code behavior.
