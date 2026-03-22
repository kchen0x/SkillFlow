# Agent Skill Sync Consolidation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Merge manual skill push into **My Skills**, merge manual skill pull into **My Agents**, simplify the primary sidebar to four core pages, and keep the legacy sync routes redirecting to the merged destinations.

**Architecture:** Keep the backend Wails contract stable and move the consolidation into frontend composition. Reuse existing push/pull APIs, add small page-state helpers for temporary manual push and pull flows, and update route wiring plus i18n so the new UX is explicit and testable.

**Tech Stack:** React, TypeScript, Wails, Node test runner, Go test, Markdown docs

---

### Task 1: Lock the new page-state rules with failing frontend tests

**Files:**
- Create: `cmd/skillflow/frontend/src/lib/dashboardManualPushState.ts`
- Create: `cmd/skillflow/frontend/src/lib/toolSkillsPullState.ts`
- Create: `cmd/skillflow/frontend/tests/dashboardManualPushState.test.mjs`
- Create: `cmd/skillflow/frontend/tests/toolSkillsPullState.test.mjs`
- Modify: `cmd/skillflow/frontend/package.json`
- Modify: `cmd/skillflow/frontend/tests/dashboardSkillSettings.test.mjs`

**Step 1: Write the failing tests**

Add tests that prove:

- dashboard toolbar actions now include `manualPush`
- manual push mode starts with no selected agents and no selected skills
- manual push can toggle agents and visible skill ids
- manual push is only ready when both an agent and a skill are selected
- pull mode starts with default category and empty selection
- pull mode can toggle scanned paths and compute `select not imported`
- pull mode is only ready when at least one scanned path is selected

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `FAIL` because the new helper files and action key do not exist yet.

**Step 3: Write minimal implementation**

Implement the smallest pure helpers needed to satisfy those tests and update the temporary test compile list.

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `ok`

### Task 2: Redirect the old sync routes and simplify navigation

**Files:**
- Modify: `cmd/skillflow/frontend/src/App.tsx`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`

**Step 1: Re-run the frontend tests**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: existing helper tests pass; no route behavior has changed yet.

**Step 2: Write minimal implementation**

Implement:

- primary sidebar order with only:
  - My Skills
  - My Prompts
  - My Memory
  - My Agents
- bottom utility section keeping Starred Repos, Cloud Backup, Settings, and Feedback
- `Navigate` redirects:
  - `/sync/push` => `/`
  - `/sync/pull` => `/tools`
- Chinese labels updated to remove mixed-language primary labels on affected controls

**Step 3: Run targeted verification**

Run:

```bash
cd cmd/skillflow/frontend && npm run build
```

Expected: `ok`

### Task 3: Merge manual push into My Skills

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/Dashboard.tsx`
- Modify: `cmd/skillflow/frontend/src/lib/dashboardSkillSettings.ts`
- Modify: `cmd/skillflow/frontend/src/components/SkillCard.tsx`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`

**Step 1: Write or extend the failing tests**

Extend helper tests to prove:

- the toolbar action order includes `manualPush`
- visible skill selection survives search filtering only for still-visible ids
- canceling manual push clears agent and skill selections

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `FAIL` until Dashboard uses the new helper flow and action key.

**Step 3: Write minimal implementation**

Implement in `Dashboard.tsx`:

- a `手动推送` toolbar action in normal mode
- inline manual-push mode with:
  - target-agent chips
  - select all / deselect all
  - cancel / start push actions
  - existing missing-dir flow
  - existing conflict dialog flow
- temporary completion feedback
- no change to the persistent auto-push target strip semantics

Keep card rendering compatible with existing multi-select styling.

**Step 4: Run targeted verification**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Run:

```bash
cd cmd/skillflow/frontend && npm run build
```

Expected: both `ok`

### Task 4: Merge manual pull into My Agents

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- Modify: `cmd/skillflow/frontend/src/lib/toolSkillsPanels.ts`
- Modify: `cmd/skillflow/frontend/src/components/SyncSkillCard.tsx`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`
- Modify: `cmd/skillflow/frontend/tests/toolSkillsPanels.test.mjs`

**Step 1: Write or extend the failing tests**

Add tests that prove:

- the segmented labels remain scoped by active panel while using localized panel ids
- visible result counts in skills mode still include pushed and scan-only skill entries
- pull-mode helper derives the correct set of visible not-imported paths

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `FAIL` until the new pull-mode helper and tool-skills integration exist.

**Step 3: Write minimal implementation**

Implement in `ToolSkills.tsx`:

- localized segmented control labels: `技能` / `记忆`
- `手动拉取` action within the skills panel
- inline pull mode for the selected agent:
  - fresh scan on entry
  - target-category selector in the top toolbar
  - select all / select not imported
  - cancel / start pull
  - existing conflict dialog flow
- keep normal browse mode for pushed and scan-only skills outside pull mode

**Step 4: Run targeted verification**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Run:

```bash
cd cmd/skillflow/frontend && npm run build
```

Expected: both `ok`

### Task 5: Sync feature documentation

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`

**Step 1: Write the documentation updates**

Update the docs to reflect:

- the new primary sidebar structure
- legacy sync routes redirecting to merged destinations
- inline manual push under **My Skills**
- inline manual pull under **My Agents**
- localized Chinese primary labels without mixed `skills`/`Skills`

**Step 2: Run checks**

Run:

```bash
rg -n "Push to Agents|Pull from Agents|手动推送|手动拉取|/sync/push|/sync/pull|My Agents|我的智能体" docs/features.md docs/features_zh.md
```

Expected: matches describe redirects and merged flows, with no stale statement that the sync pages remain primary sidebar destinations.

### Task 6: Final verification

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

**Step 3: Run backend regression tests**

Run:

```bash
go test ./core/... ./cmd/skillflow
```

Expected: `ok`
