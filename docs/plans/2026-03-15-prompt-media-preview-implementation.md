# Prompt Media Preview Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add enlarged in-app image preview inside the prompt editor, move media input rows to the bottom of their panels, and simplify prompt cards by removing media clutter.

**Architecture:** Prompt editor thumbnails will set local preview state and render a secondary overlay dialog above the editor. Image and link panels will keep rendered media at the top, while new items are added through a shared attachment area at the end of the editor body. Prompt cards will stop rendering images and links entirely, keeping list density focused on text scanning.

**Tech Stack:** React, TypeScript, Wails frontend, Node test runner

---

### Task 1: Lock preview and image-append helpers with tests

**Files:**
- Modify: `cmd/skillflow/frontend/tests/promptRichContent.test.mjs`
- Modify: `cmd/skillflow/frontend/src/lib/promptRichContent.ts`

1. Add failing tests for exported helpers that normalize a single previewable HTTP image URL and append one image URL into the editor state.
2. Run the prompt rich-content tests and confirm the new test fails first.
3. Implement the helper with minimal code.
4. Re-run the same test file until it passes.

### Task 2: Add editor-only image preview overlay and bottom input rows

**Files:**
- Modify: `cmd/skillflow/frontend/src/components/PromptEditorDialog.tsx`

1. Add local state for the selected preview image URL.
2. Turn image thumbnails into buttons that open a larger in-app preview and add delete actions directly on the thumbnail corners.
3. Replace the image multi-input editor with one input row and right-aligned add button in the shared attachment area, matching the web-link interaction pattern.
4. Move the web-link input row to that same shared attachment area while keeping the existing display area in place.
5. Render a secondary overlay dialog with the enlarged image and close controls.

### Task 3: Simplify prompt cards and docs

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/Prompts.tsx`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`

1. Remove image thumbnails and web-link chips from prompt cards.
2. Add any lightweight preview affordance copy needed by the editor.
3. Update prompt feature docs in English and Chinese.

### Task 4: Verify

**Files:**
- None

1. Run `cd cmd/skillflow/frontend && npm run test:unit`.
2. Run `cd cmd/skillflow/frontend && npm run build`.
