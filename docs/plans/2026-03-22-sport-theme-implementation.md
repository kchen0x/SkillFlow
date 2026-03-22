# Add Sport Theme Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a fourth `sport` appearance theme that uses a green athletic palette, appears in Settings and the quick theme cycle, and keeps the existing frontend accessibility baseline.

**Architecture:** Reuse the current frontend theme pipeline without introducing a new abstraction. Extend the `Theme` enum and cycle order, add a new CSS token block, wire a fourth Settings preview card and i18n copy, then expand the existing theme contrast test and update feature docs.

**Tech Stack:** React, TypeScript, Tailwind CSS, plain CSS variables, Node test runner, Markdown docs

---

### Task 1: Add the failing theme coverage for the new sport preset

**Files:**
- Modify: `cmd/skillflow/frontend/tests/themeContrast.test.mjs`

**Step 1: Write the failing test**

Update the existing theme iteration so it expects `sport` to exist and satisfy the same contrast assertions as the current themes.

Change:

```js
for (const themeName of ['dark', 'young', 'light']) {
```

to:

```js
for (const themeName of ['dark', 'young', 'light', 'sport']) {
```

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit -- themeContrast.test.mjs
```

Expected: `FAIL` because `style.css` does not yet contain a `sport` theme block.

**Step 3: Write minimal implementation**

Do not touch implementation yet beyond the failing expectation in this task.

**Step 4: Run test to verify it fails for the correct reason**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit -- themeContrast.test.mjs
```

Expected: failure message indicating the missing `sport` theme block or missing contrast tokens.

**Step 5: Commit**

Run:

```bash
git add cmd/skillflow/frontend/tests/themeContrast.test.mjs
git commit -m "test: require sport theme contrast coverage"
```

### Task 2: Implement the sport theme across the frontend theme system

**Files:**
- Modify: `cmd/skillflow/frontend/src/hooks/useTheme.ts`
- Modify: `cmd/skillflow/frontend/src/style.css`
- Modify: `cmd/skillflow/frontend/src/pages/Settings.tsx`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`

**Step 1: Write the failing test**

Use the failing contrast test from Task 1 as the current red state. If needed, add only the smallest source-level expectation necessary so the following are all required:

- `sport` is a valid theme value
- `getNextTheme()` reaches `sport`
- the Settings page contains a fourth theme option

Prefer updating an existing source-level test if one already covers theme metadata. If none exists, rely on the unit build/typecheck and the contrast test rather than adding redundant tests.

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit -- themeContrast.test.mjs
```

Expected: `FAIL`

**Step 3: Write minimal implementation**

Implement:

- add `sport` to `THEMES`, `Theme`, and `THEME_LABELS`
- keep the default theme unchanged
- extend `style.css` with a full `[data-theme="sport"]` token block
- add a fourth theme preview card in `Settings.tsx`
- update theme descriptions and theme-cycle hint text in `en.ts` and `zh.ts`

Use a palette direction based on:

- mint-tinted shell background
- pale green surfaces
- deep field-green primary accent
- teal-green secondary accent
- readable dark text and controlled glow

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit -- themeContrast.test.mjs
```

Expected: `PASS`

**Step 5: Run build verification**

Run:

```bash
cd cmd/skillflow/frontend && npm run build
```

Expected: successful build with no TypeScript errors from the new `Theme` value.

**Step 6: Commit**

Run:

```bash
git add cmd/skillflow/frontend/src/hooks/useTheme.ts cmd/skillflow/frontend/src/style.css cmd/skillflow/frontend/src/pages/Settings.tsx cmd/skillflow/frontend/src/i18n/en.ts cmd/skillflow/frontend/src/i18n/zh.ts
git commit -m "feat: add sport appearance theme"
```

### Task 3: Update the user-facing feature docs

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`

**Step 1: Write the failing doc check**

Run:

```bash
rg -n "Dark → Young → Light|Three visual presets|三张预览卡片|Dark → Young → Light" docs/features.md docs/features_zh.md
```

Expected: matches show stale three-theme copy that must be updated.

**Step 2: Write minimal documentation updates**

Update the docs so they reflect:

- four themes instead of three
- `sport` as a green athletic preset
- the updated quick-cycle order
- the new last-updated date if applicable

**Step 3: Verify the docs**

Run:

```bash
rg -n "Dark → Young → Light → Sport|Four visual presets|四张预览卡片|Sport" docs/features.md docs/features_zh.md
```

Expected: updated wording is present.

**Step 4: Commit**

Run:

```bash
git add docs/features.md docs/features_zh.md
git commit -m "docs: add sport theme to feature docs"
```

### Task 4: Run final verification before handoff

**Files:**
- No additional file changes expected

**Step 1: Run frontend unit tests**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `PASS`

**Step 2: Run frontend production build**

Run:

```bash
cd cmd/skillflow/frontend && npm run build
```

Expected: `PASS`

**Step 3: Review final diff**

Run:

```bash
git status --short
git diff -- cmd/skillflow/frontend/src/hooks/useTheme.ts cmd/skillflow/frontend/src/style.css cmd/skillflow/frontend/src/pages/Settings.tsx cmd/skillflow/frontend/src/i18n/en.ts cmd/skillflow/frontend/src/i18n/zh.ts cmd/skillflow/frontend/tests/themeContrast.test.mjs docs/features.md docs/features_zh.md docs/plans/2026-03-22-sport-theme-design.md docs/plans/2026-03-22-sport-theme-implementation.md
```

Expected: only the planned sport-theme files are changed.

**Step 4: Commit**

Run:

```bash
git add docs/plans/2026-03-22-sport-theme-design.md docs/plans/2026-03-22-sport-theme-implementation.md
git commit -m "docs: add sport theme design and plan"
```
