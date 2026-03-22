# My Agents Skills Memory Toggle Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a top `Skills | Memory` segmented control to **My Agents** so the page shows only one content surface at a time, with search and result counts scoped to the active panel.

**Architecture:** Keep the existing backend calls unchanged. Extract a small frontend helper that computes panel-scoped result counts and filtered content, then update `ToolSkills` to render a toolbar toggle and conditional sections for the active panel.

**Tech Stack:** React, TypeScript, Node test runner, Markdown docs

---

### Task 1: Add failing frontend tests for panel-scoped derived state

**Files:**
- Create: `cmd/skillflow/frontend/src/lib/toolSkillsPanels.ts`
- Create: `cmd/skillflow/frontend/tests/toolSkillsPanels.test.mjs`
- Modify: `cmd/skillflow/frontend/package.json`

**Step 1: Write the failing test**

Add tests that prove:
- the default panel is `skills`
- `skills` result counts include pushed and scan-only skills but exclude memory entries
- `memory` result counts include only memory entries
- panel-scoped search only filters the active panel content

Suggested test names:
- `getDefaultToolSkillsPanel returns skills`
- `getVisibleResultCount uses only skill entries for skills panel`
- `getVisibleResultCount uses only memory entries for memory panel`
- `filterToolSkillsPanelContent scopes search to the active panel`

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `FAIL` because the helper does not exist and the unit-test script does not compile it yet.

**Step 3: Write minimal implementation**

Implement:
- panel type and default-panel helper
- pure filtering and count helpers for `skills` and `memory`
- unit-test script compilation entry for the new helper

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `ok`

### Task 2: Add the segmented control and split ToolSkills rendering by panel

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`

**Step 1: Write the failing test**

Use the helper tests from Task 1 as the behavior lock for:
- default `Skills` panel
- panel-specific result counts
- panel-specific search behavior

**Step 2: Run test to verify it fails against current page behavior**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `FAIL` until `ToolSkills` uses the new panel-scoped helpers and the segmented control state.

**Step 3: Write minimal implementation**

Implement:
- `activePanel` state in `ToolSkills`
- top segmented control near the selected agent title
- conditional toolbar actions so batch delete appears only for `skills`
- conditional page sections so `Memory` and `Skills` no longer render together
- search result label from panel-scoped counts
- new translations for panel labels and any panel-specific empty-state copy needed

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `ok`

### Task 3: Sync My Agents feature docs

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`

**Step 1: Write the failing check**

Create a manual checklist that confirms:
- **My Agents** documents the `Skills | Memory` segmented control
- the `Skills` panel documents skill sections and batch delete scope
- the `Memory` panel documents memory preview scope

**Step 2: Run the check**

Run:

```bash
rg -n "Skills \\| Memory|My Agents|我的智能体|Memory panel|记忆面板" docs/features.md docs/features_zh.md
```

Expected: missing or partial matches before the doc update.

**Step 3: Write minimal documentation updates**

Update:
- **My Agents** English and Chinese sections
- feature-doc last-updated dates

**Step 4: Run the check again**

Run:

```bash
rg -n "Skills \\| Memory|My Agents|我的智能体|Memory panel|记忆面板" docs/features.md docs/features_zh.md
```

Expected: matches present in both docs.

### Task 4: Final verification

**Files:**
- No code changes expected

**Step 1: Run frontend unit tests**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `ok`

**Step 2: Run frontend production build**

Run:

```bash
cd cmd/skillflow/frontend && npm run build
```

Expected: `ok`
