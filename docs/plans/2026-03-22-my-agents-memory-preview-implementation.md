# My Agents Memory Preview Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a read-only memory preview section to **My Agents** so users can inspect each selected agent's current main memory file and rules directory content from the agent panel.

**Architecture:** Persist `memoryPath` and `rulesDir` in local agent config, add a read-only `core/readmodel/agentmemory` loader for the agent filesystem view, expose it through a thin Wails App method, then render the preview section in the existing `ToolSkills` page with shared search filtering and open-path actions.

**Tech Stack:** Go, Wails, React, TypeScript, Markdown docs

---

### Task 1: Persist agent memory paths in local config

**Files:**
- Modify: `core/agentintegration/app/settings.go`
- Modify: `core/agentintegration/app/settings_test.go`
- Modify: `core/config/service.go`
- Modify: `core/config/service_test.go`

**Step 1: Write the failing test**

Add tests that prove:
- default local agent settings include `memoryPath` and `rulesDir`
- saving and loading config preserves `agents[].memoryPath` and `agents[].rulesDir`

Suggested test names:
- `TestDefaultSettingsIncludeBuiltinAgentMemoryPaths`
- `TestSaveAndLoadPreservesAgentMemoryPaths`

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./core/agentintegration/app ./core/config -run 'TestDefaultSettingsIncludeBuiltinAgentMemoryPaths|TestSaveAndLoadPreservesAgentMemoryPaths'
```

Expected: `FAIL` because the local settings model and config split/merge path do not persist those fields yet.

**Step 3: Write minimal implementation**

Implement:
- `MemoryPath` and `RulesDir` on local agent settings
- default local settings population from built-in agent defaults
- config split / merge / default-local persistence for those fields

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./core/agentintegration/app ./core/config -run 'TestDefaultSettingsIncludeBuiltinAgentMemoryPaths|TestSaveAndLoadPreservesAgentMemoryPaths'
```

Expected: `ok`

### Task 2: Add the agent-memory read model and App transport

**Files:**
- Create: `core/readmodel/agentmemory/model.go`
- Create: `core/readmodel/agentmemory/service.go`
- Create: `core/readmodel/agentmemory/service_test.go`
- Create: `cmd/skillflow/app_agent_memory.go`
- Modify: `cmd/skillflow/app_agent_api_test.go`

**Step 1: Write the failing test**

Add tests that prove:
- the read model returns configured paths, main-file content, and flat `.md` rule previews
- `sf-*.md` rule files are marked as managed
- the App method returns a DTO for the selected agent

Suggested test names:
- `TestLoadPreviewReadsMainMemoryAndRules`
- `TestLoadPreviewMarksManagedRuleFiles`
- `TestGetAgentMemoryPreviewReturnsAgentPreview`

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./core/readmodel/agentmemory ./cmd/skillflow -run 'TestLoadPreviewReadsMainMemoryAndRules|TestLoadPreviewMarksManagedRuleFiles|TestGetAgentMemoryPreviewReturnsAgentPreview'
```

Expected: `FAIL` because the read model and App method do not exist yet.

**Step 3: Write minimal implementation**

Implement:
- read-model structs for main memory and rule-file previews
- flat filesystem loading for the configured `memoryPath` and `rulesDir`
- managed-file detection via `sf-` prefix
- thin App transport method with start/completed/failed logging

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./core/readmodel/agentmemory ./cmd/skillflow -run 'TestLoadPreviewReadsMainMemoryAndRules|TestLoadPreviewMarksManagedRuleFiles|TestGetAgentMemoryPreviewReturnsAgentPreview'
```

Expected: `ok`

### Task 3: Add frontend preview helpers and render the My Agents memory section

**Files:**
- Create: `cmd/skillflow/frontend/src/lib/agentMemoryPreview.ts`
- Create: `cmd/skillflow/frontend/tests/agentMemoryPreview.test.mjs`
- Modify: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.js`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`
- Modify: `cmd/skillflow/frontend/wailsjs/go/models.ts`

**Step 1: Write the failing test**

Add frontend helper tests that prove:
- memory preview entries are ordered with main memory first and managed rule files grouped predictably
- search matches rule file names and content case-insensitively

Suggested test names:
- `buildAgentMemoryEntries keeps main memory first and sorts rules`
- `filterAgentMemoryEntries matches title and content`

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `FAIL` because the helper and its test target are not wired into the unit-test script yet.

**Step 3: Write minimal implementation**

Implement:
- pure helper functions for building and filtering preview entries
- `ToolSkills` loading state for agent memory preview alongside skills
- a new memory section with open-file / open-directory / refresh actions
- inline empty / missing states for file and rules directory
- new translations
- regenerated Wails bindings for the new App method

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `ok`

### Task 4: Sync docs for the new My Agents capability and config schema

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `docs/config.md`
- Modify: `docs/config_zh.md`

**Step 1: Write the failing check**

Create a manual checklist that confirms:
- **My Agents** documents the new memory preview section and actions
- local config documents `agents[].memoryPath` and `agents[].rulesDir`
- the feature docs keep their last-updated date in sync

**Step 2: Run the check**

Run:

```bash
rg -n "memoryPath|rulesDir|My Agents|我的智能体" docs/features.md docs/features_zh.md docs/config.md docs/config_zh.md
```

Expected: partial or missing matches before doc updates.

**Step 3: Write minimal documentation updates**

Update:
- **My Agents** feature documentation in English and Chinese
- `config_local.json` examples and key tables for local agent path fields
- feature-doc last-updated dates

**Step 4: Run the check again**

Run:

```bash
rg -n "memoryPath|rulesDir|My Agents|我的智能体" docs/features.md docs/features_zh.md docs/config.md docs/config_zh.md
```

Expected: matches present in all four docs.

### Task 5: Final verification

**Files:**
- No code changes expected

**Step 1: Run targeted Go verification**

Run:

```bash
go test ./core/agentintegration/app ./core/config ./core/readmodel/agentmemory ./cmd/skillflow
```

Expected: `ok`

**Step 2: Run frontend verification**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `ok`

**Step 3: Run full feature-oriented verification**

Run:

```bash
make test
```

Expected: `ok`, or record the exact failing slice if unrelated pre-existing failures remain.
