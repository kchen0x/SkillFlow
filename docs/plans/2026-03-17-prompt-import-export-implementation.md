# Prompt Import And Export Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add inline scoped prompt export plus import conflict review without changing the on-disk prompt JSON schema.

**Architecture:** Keep export selection logic in the frontend and pass explicit prompt names into a scoped backend export API. Move import to a prepare/complete session flow so the backend can parse once, report conflicts before any writes, and then apply only the overwrite choices confirmed by the user.

**Tech Stack:** Go, Wails, React, TypeScript, Node test runner, Markdown docs

---

### Task 1: Lock scoped export behavior in `core/prompt`

**Files:**
- Modify: `core/prompt/storage.go`
- Modify: `core/prompt/storage_test.go`

**Step 1: Write the failing test**

Add Go tests that prove:
- exporting with no explicit names still includes all prompts
- exporting with a name subset only includes those prompts
- subset export keeps each prompt's `category`, `imageURLs`, and `webLinks`

Suggested test names:
- `TestStorageExportJSONByNamesReturnsAllWhenEmpty`
- `TestStorageExportJSONByNamesFiltersPromptSubset`

**Step 2: Run test to verify it fails**

Run: `go test ./core/prompt -run 'TestStorageExportJSONByNamesReturnsAllWhenEmpty|TestStorageExportJSONByNamesFiltersPromptSubset'`

Expected: `FAIL` because the scoped export API does not exist yet.

**Step 3: Write minimal implementation**

Implement a scoped export path in `storage.go`, for example:
- keep `ExportJSON()` as the public all-prompts wrapper
- add `ExportJSONByNames(names []string)` for filtered export
- reuse the same bundle-building logic for both paths

**Step 4: Run test to verify it passes**

Run: `go test ./core/prompt -run 'TestStorageExportJSONByNamesReturnsAllWhenEmpty|TestStorageExportJSONByNamesFiltersPromptSubset'`

Expected: `ok`

**Step 5: Commit**

Run:

```bash
git add core/prompt/storage.go core/prompt/storage_test.go
git commit -m "feat: add scoped prompt export"
```

### Task 2: Lock import preview and apply semantics in `core/prompt`

**Files:**
- Modify: `core/prompt/storage.go`
- Modify: `core/prompt/storage_test.go`

**Step 1: Write the failing tests**

Add Go tests that prove:
- import preview classifies new prompts and conflicts without writing files
- skip leaves existing prompt content and category untouched
- overwrite updates existing prompt content, metadata, and category to the imported values

Suggested test names:
- `TestStoragePreviewImportJSONSeparatesCreatesAndConflicts`
- `TestStorageApplyImportSkipsConflicts`
- `TestStorageApplyImportOverwritesConflictAndCategory`

**Step 2: Run test to verify it fails**

Run: `go test ./core/prompt -run 'TestStoragePreviewImportJSONSeparatesCreatesAndConflicts|TestStorageApplyImportSkipsConflicts|TestStorageApplyImportOverwritesConflictAndCategory'`

Expected: `FAIL` because preview/apply helpers do not exist yet.

**Step 3: Write minimal implementation**

Implement preview/apply helpers in `storage.go`, for example:
- parse once into an import bundle
- return explicit create and conflict slices
- apply creates immediately
- apply only the conflict names included in an explicit overwrite set

Keep the existing file format and category data untouched.

**Step 4: Run test to verify it passes**

Run: `go test ./core/prompt -run 'TestStoragePreviewImportJSONSeparatesCreatesAndConflicts|TestStorageApplyImportSkipsConflicts|TestStorageApplyImportOverwritesConflictAndCategory'`

Expected: `ok`

**Step 5: Commit**

Run:

```bash
git add core/prompt/storage.go core/prompt/storage_test.go
git commit -m "feat: add prompt import preview flow"
```

### Task 3: Add app-level import session and scoped export bindings

**Files:**
- Create: `cmd/skillflow/app_prompt_session.go`
- Create: `cmd/skillflow/app_prompt_session_test.go`
- Modify: `cmd/skillflow/app_prompt.go`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.js`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`

**Step 1: Write the failing test**

Add Go tests for the app-level session helper that prove:
- a prepared import session can be stored and retrieved by ID
- a consumed or cancelled session is removed
- completing one session does not affect another session

Suggested test names:
- `TestPromptImportSessionStoreRoundTrip`
- `TestPromptImportSessionStoreDeleteRemovesSession`

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/skillflow -run 'TestPromptImportSessionStoreRoundTrip|TestPromptImportSessionStoreDeleteRemovesSession'`

Expected: `FAIL` because the session store does not exist yet.

**Step 3: Write minimal implementation**

Implement:
- an in-memory import session store in `app_prompt_session.go`
- new app methods in `app_prompt.go` for:
  - preparing prompt import
  - completing prompt import with overwrite names
  - exporting prompt subsets by explicit names
- dialog handling and stable logging for the new flows

If Wails generation is available, regenerate bindings after the exported app methods change. If not, update `App.js` and `App.d.ts` manually to match the new method names and signatures.

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/skillflow -run 'TestPromptImportSessionStoreRoundTrip|TestPromptImportSessionStoreDeleteRemovesSession'`

Expected: `ok`

**Step 5: Commit**

Run:

```bash
git add cmd/skillflow/app_prompt.go cmd/skillflow/app_prompt_session.go cmd/skillflow/app_prompt_session_test.go cmd/skillflow/frontend/wailsjs/go/main/App.js cmd/skillflow/frontend/wailsjs/go/main/App.d.ts
git commit -m "feat: add prompt import sessions"
```

### Task 4: Add frontend export-range helpers and inline export bar

**Files:**
- Create: `cmd/skillflow/frontend/src/lib/promptExport.ts`
- Create: `cmd/skillflow/frontend/tests/promptExport.test.mjs`
- Modify: `cmd/skillflow/frontend/package.json`
- Modify: `cmd/skillflow/frontend/src/pages/Prompts.tsx`

**Step 1: Write the failing test**

Add frontend unit tests that prove:
- available export actions are `全部`, selected category name when relevant, and `指定`
- the category action disappears when the left sidebar already points to all prompts
- `指定` resolves only the prompts inside the current sidebar scope
- export confirmation blocks empty manual selections

Suggested test names:
- `buildPromptExportActions hides duplicate all action`
- `resolveScopedPromptSelection limits specified export to current filter`

**Step 2: Run test to verify it fails**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: `FAIL` because the new helper file is not compiled yet and the new tests are not implemented.

**Step 3: Write minimal implementation**

Implement:
- a small helper module in `promptExport.ts` for action visibility and selection resolution
- `package.json` test compilation updates so the new helper is included in `.tmp-tests`
- inline export bar state in `Prompts.tsx`
- short export labels: `全部`, selected category name, `指定`
- multi-select list only when `指定` is active

**Step 4: Run test to verify it passes**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: `PASS`

**Step 5: Commit**

Run:

```bash
git add cmd/skillflow/frontend/package.json cmd/skillflow/frontend/src/lib/promptExport.ts cmd/skillflow/frontend/src/pages/Prompts.tsx cmd/skillflow/frontend/tests/promptExport.test.mjs
git commit -m "feat: add scoped prompt export bar"
```

### Task 5: Add frontend import conflict resolution flow

**Files:**
- Create: `cmd/skillflow/frontend/src/lib/promptImportConflicts.ts`
- Create: `cmd/skillflow/frontend/tests/promptImportConflicts.test.mjs`
- Modify: `cmd/skillflow/frontend/package.json`
- Modify: `cmd/skillflow/frontend/src/components/ConflictDialog.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/Prompts.tsx`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`

**Step 1: Write the failing test**

Add frontend unit tests that prove:
- one overwrite choice only marks the current conflict
- one skip choice only marks the current conflict
- checking `对剩余 {count} 个冲突执行相同操作` applies the same decision to every remaining conflict in the current session

Suggested test names:
- `applyPromptImportDecision marks one conflict by default`
- `applyPromptImportDecision applies same action to remaining conflicts when requested`

**Step 2: Run test to verify it fails**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: `FAIL`

**Step 3: Write minimal implementation**

Implement:
- a helper module for conflict-decision accumulation
- checkbox support in `ConflictDialog.tsx`
- prompt-page import state that calls prepare first and complete only after conflict decisions are finalized
- localized copy for the checkbox and any revised import status messages

**Step 4: Run test to verify it passes**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: `PASS`

**Step 5: Commit**

Run:

```bash
git add cmd/skillflow/frontend/package.json cmd/skillflow/frontend/src/lib/promptImportConflicts.ts cmd/skillflow/frontend/src/components/ConflictDialog.tsx cmd/skillflow/frontend/src/pages/Prompts.tsx cmd/skillflow/frontend/src/i18n/en.ts cmd/skillflow/frontend/src/i18n/zh.ts cmd/skillflow/frontend/tests/promptImportConflicts.test.mjs
git commit -m "feat: add prompt import conflict choices"
```

### Task 6: Update docs, regenerate bindings, and verify end to end

**Files:**
- Modify: `README.md`
- Modify: `README_zh.md`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.js`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`

**Step 1: Update docs**

Document:
- inline export bar with `全部` / selected category / `指定`
- scoped multi-select export
- import conflict review with skip, overwrite, and same-action checkbox

Refresh the `Last updated` / `最后更新` dates in the feature docs.

**Step 2: Regenerate or reconcile Wails bindings**

Run: `make generate`

Expected: generated bindings reflect the new exported app methods. If generation is unavailable, manually reconcile `App.js` and `App.d.ts`.

**Step 3: Run verification**

Run:
- `go test ./core/...`
- `go test ./cmd/skillflow -run 'TestPromptImportSessionStoreRoundTrip|TestPromptImportSessionStoreDeleteRemovesSession'`
- `cd cmd/skillflow/frontend && npm run test:unit`

Expected:
- all targeted Go tests pass
- frontend unit suite passes

**Step 4: Commit**

Run:

```bash
git add README.md README_zh.md docs/features.md docs/features_zh.md cmd/skillflow/frontend/wailsjs/go/main/App.js cmd/skillflow/frontend/wailsjs/go/main/App.d.ts
git commit -m "feat: enhance prompt import and export flows"
```
