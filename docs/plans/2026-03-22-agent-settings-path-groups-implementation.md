# Agent Settings Path Groups and Modal Creation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Split each Settings agent card into dedicated Skill and Memory path modules, replace inline custom-agent creation with a modal dialog, and preserve multi-scan-directory editing.

**Architecture:** Keep the existing persisted `AgentProfile` schema and page-level `SaveConfig(cfg)` flow. Implement the feature entirely in the frontend by introducing a small draft helper for custom-agent creation, then refactoring the Agents tab UI so all edits continue to accumulate in the in-memory Settings draft until the user clicks **Save Settings**.

**Tech Stack:** React, TypeScript, Wails, Node test runner, Markdown docs

---

### Task 1: Add testable helper logic for custom-agent draft creation

**Files:**
- Create: `cmd/skillflow/frontend/src/lib/agentSettings.ts`
- Create: `cmd/skillflow/frontend/tests/agentSettings.test.mjs`
- Modify: `cmd/skillflow/frontend/package.json`

**Step 1: Write the failing test**

Add unit tests that prove:

- an empty custom-agent draft starts with blank `name`, `pushDir`, `memoryPath`, and `rulesDir`
- building a custom agent profile trims values and seeds `scanDirs` with exactly one entry from `pushDir`
- duplicate agent names are rejected after trimming
- modal-save validation fails when any required field is blank

Suggested test names:

- `createEmptyCustomAgentDraft returns blank required fields`
- `buildCustomAgentProfile seeds scanDirs from pushDir`
- `validateCustomAgentDraft rejects duplicate names`
- `validateCustomAgentDraft requires all fields`

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `FAIL` because `agentSettings.ts` is not compiled or does not exist yet.

**Step 3: Write minimal implementation**

Implement a narrow helper module that exports:

- an empty custom-agent draft factory
- draft validation
- custom-agent profile builder

Keep the helper free of UI concerns. It should only prepare and validate data for `Settings.tsx`.

Suggested shape:

```ts
export function createEmptyCustomAgentDraft() {
  return { name: '', pushDir: '', memoryPath: '', rulesDir: '' }
}

export function buildCustomAgentProfile(draft: CustomAgentDraft) {
  const pushDir = draft.pushDir.trim()
  return {
    name: draft.name.trim(),
    pushDir,
    scanDirs: [pushDir],
    memoryPath: draft.memoryPath.trim(),
    rulesDir: draft.rulesDir.trim(),
    enabled: true,
    custom: true,
  }
}
```

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: the new helper tests pass.

**Step 5: Commit**

Run:

```bash
git add cmd/skillflow/frontend/src/lib/agentSettings.ts cmd/skillflow/frontend/tests/agentSettings.test.mjs cmd/skillflow/frontend/package.json
git commit -m "test: add custom agent settings draft helper"
```

### Task 2: Refactor the Agents tab into Skill and Memory modules and add the modal

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/Settings.tsx`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`
- Create: `cmd/skillflow/frontend/tests/settingsAgentsUi.test.mjs`

**Step 1: Write the failing test**

Add source-level tests that prove:

- `Settings.tsx` uses `AnimatedDialog` for custom-agent creation
- the Agents tab renders separate labeled sections for skill paths and memory paths
- the modal draft includes `name`, `pushDir`, `memoryPath`, and `rulesDir`
- the page no longer calls `AddCustomAgent(...)` or `RemoveCustomAgent(...)` directly from the Agents tab source

Suggested test names:

- `settings agents tab uses AnimatedDialog for custom agent creation`
- `settings agents tab exposes separate skill and memory path sections`
- `settings custom agent dialog includes required fields`
- `settings agents tab no longer directly persists custom agent edits`

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `FAIL` because the page still uses the inline add form and direct Wails calls.

**Step 3: Write minimal implementation**

In `Settings.tsx`:

- remove `AddCustomAgent` and `RemoveCustomAgent` from imports
- add local state for:
  - dialog open/close
  - custom-agent draft
  - validation error message if needed
- replace the dashed inline add form with:
  - an **Add Custom Agent** trigger button
  - an `AnimatedDialog` modal
- on modal save:
  - validate the draft
  - append the new custom agent to `cfg.agents`
  - close the modal
  - reset the draft
- on custom-agent delete:
  - remove that agent from `cfg.agents` locally only

Refactor each agent card into two visual modules:

- **Skill Paths**
  - `pushDir`
  - `scanDirs`
- **Memory Paths**
  - `memoryPath`
  - `rulesDir`

Keep the existing picker helpers and existing multi-scan-directory editing behavior.

Update translations with new labels such as:

- `settings.skillPathsSection`
- `settings.memoryPathsSection`
- `settings.addCustomToolDialogTitle`
- `settings.addCustomToolDialogHint`
- `settings.addCustomToolSkillPath`
- `settings.addCustomToolSave`
- `settings.agentNameDuplicate`
- `settings.agentFieldsRequired`

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `PASS`

**Step 5: Verify the page still uses the existing save boundary**

Run:

```bash
rg -n "AddCustomAgent|RemoveCustomAgent" cmd/skillflow/frontend/src/pages/Settings.tsx
rg -n "SaveConfig\\(" cmd/skillflow/frontend/src/pages/Settings.tsx
```

Expected:

- no `AddCustomAgent` or `RemoveCustomAgent` usage in `Settings.tsx`
- existing page-level `SaveConfig` flow remains intact

**Step 6: Commit**

Run:

```bash
git add cmd/skillflow/frontend/src/pages/Settings.tsx cmd/skillflow/frontend/src/i18n/en.ts cmd/skillflow/frontend/src/i18n/zh.ts cmd/skillflow/frontend/tests/settingsAgentsUi.test.mjs
git commit -m "feat: split agent settings paths and add custom agent dialog"
```

### Task 3: Update user-facing feature docs for the new Settings flow

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`

**Step 1: Write the failing test**

Add or update documentation expectations manually by identifying the stale sections before editing:

- **Settings → Agents** still describes a flat agent card and inline custom-agent section
- the memory settings subsection does not explain the new grouped module layout

Use these checks to make the required stale text explicit:

```bash
rg -n "Add Custom Agent|inline|Push directory|Scan directories|Memory File|Rules Directory" docs/features.md
rg -n "添加自定义智能体|推送路径|扫描路径|记忆文件|规则目录" docs/features_zh.md
```

Expected: current wording reflects the old layout.

**Step 2: Write minimal documentation updates**

Update `docs/features.md` to describe:

- each agent card now has **Skill Paths** and **Memory Paths** sections
- custom-agent creation uses a modal dialog
- the modal collects name, skill path, memory file, and rules directory
- extra scan paths are still edited after creation inside the Skill Paths section

Update `docs/features_zh.md` with the same behavior in Chinese.

Also update the bottom `Last updated` / `最后更新` date if needed.

Do not change `README.md` or `README_zh.md`.

**Step 3: Verify the docs**

Run:

```bash
rg -n "Skill Paths|Memory Paths|modal dialog|scan paths are still edited after creation" docs/features.md
rg -n "Skill 路径|记忆路径|弹窗|创建后仍可继续编辑扫描路径" docs/features_zh.md
```

Expected: the new grouped layout and modal flow are documented in both languages.

**Step 4: Commit**

Run:

```bash
git add docs/features.md docs/features_zh.md
git commit -m "docs: update agent settings flow"
```

### Task 4: Run final verification before handoff

**Files:**
- No code changes expected

**Step 1: Run the full verification commands**

Run:

```bash
go test ./core/...
cd cmd/skillflow/frontend && npm run test:unit
```

Expected:

- `go test ./core/...` returns `ok`
- frontend unit tests return `PASS`

**Step 2: Check the final diff**

Run:

```bash
git status --short
git diff -- docs/plans/2026-03-22-agent-settings-path-groups-design.md docs/plans/2026-03-22-agent-settings-path-groups-implementation.md docs/features.md docs/features_zh.md cmd/skillflow/frontend/src/pages/Settings.tsx cmd/skillflow/frontend/src/lib/agentSettings.ts cmd/skillflow/frontend/tests/agentSettings.test.mjs cmd/skillflow/frontend/tests/settingsAgentsUi.test.mjs cmd/skillflow/frontend/src/i18n/en.ts cmd/skillflow/frontend/src/i18n/zh.ts cmd/skillflow/frontend/package.json
```

Expected: only the planned frontend and docs files are changed.

**Step 3: Commit**

Run:

```bash
git add docs/plans/2026-03-22-agent-settings-path-groups-design.md docs/plans/2026-03-22-agent-settings-path-groups-implementation.md
git commit -m "docs: add agent settings implementation plan"
```
