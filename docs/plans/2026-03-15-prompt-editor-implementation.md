# Prompt Editor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Keep the prompt editor usable inside the desktop window while changing web-link editing to a single-line markdown input backed by structured link state.

**Architecture:** The frontend draft model will hold `PromptWebLink[]` and only serialize back to markdown when saving through the existing Wails API. Helper functions in `promptRichContent.ts` will own single-link parsing and markdown rebuilding so the dialog stays UI-focused.

**Tech Stack:** React, TypeScript, Wails frontend, Node test runner

---

### Task 1: Lock helper behavior with tests

**Files:**
- Modify: `cmd/skillflow/frontend/tests/promptRichContent.test.mjs`
- Modify: `cmd/skillflow/frontend/src/lib/promptRichContent.ts`

1. Add failing tests for parsing one markdown link line and ignoring invalid single-line input.
2. Run the prompt rich-content unit tests and confirm the new assertions fail first.
3. Implement the minimal helper functions needed by the dialog.
4. Re-run the same tests until they pass.

### Task 2: Refactor prompt draft model and dialog UI

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/Prompts.tsx`
- Modify: `cmd/skillflow/frontend/src/components/PromptEditorDialog.tsx`

1. Change `PromptDraft` to store `webLinks: PromptWebLink[]` instead of raw markdown text.
2. Hydrate the editor draft from saved prompt links and serialize links back to markdown only in `handleSave`.
3. Replace the link textarea with a single-line markdown input plus add/remove interactions.
4. Constrain the dialog body and content textarea with internal scrolling so the footer remains reachable.

### Task 3: Update copy and docs

**Files:**
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`

1. Update prompt-editor helper text to describe the single-link markdown input and add flow.
2. Update the English and Chinese feature docs to describe scrollable prompt editing and link-chip rendering.
3. Refresh the "Last updated" dates in the feature docs.

### Task 4: Verify

**Files:**
- None

1. Run `cd cmd/skillflow/frontend && npm run test:unit`.
2. Confirm the prompt rich-content tests and existing unit suite both pass.
