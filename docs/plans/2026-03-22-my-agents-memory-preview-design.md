# My Agents Memory Preview Design

**Date:** 2026-03-22

## Context

SkillFlow already has a dedicated **My Memory** page for editing source memory content, but the **My Agents** page only shows pushed and scanned skills. Users cannot inspect what memory content is currently configured for a specific agent from that agent-centric view.

There is also a configuration gap: agent profiles already define `memoryPath` and `rulesDir` in the runtime model, and Settings already exposes those fields, but the local config split/merge path does not persist them today.

## Decision

1. Add an in-page **Memory** section to **My Agents** for the currently selected agent.
2. Read the preview from the agent's actual configured filesystem targets:
   - the agent main memory file at `memoryPath`
   - the agent rules directory at `rulesDir`
3. Keep the preview read-only in **My Agents**. Editing remains on **My Memory**.
4. Provide direct actions from **My Agents** to:
   - open the agent main memory file
   - open the agent rules directory
   - refresh the preview
5. Persist `memoryPath` and `rulesDir` in local agent config so Settings changes survive reload.

## Why This Approach

- It matches the user mental model: from an agent page, you can inspect the current memory that agent will read.
- It avoids duplicating memory editing flows inside **My Agents**.
- It shows actual on-disk agent content rather than only SkillFlow source memory, which is more useful for troubleshooting merge-mode output and per-agent rule files.

## Read Model

Add a small read-model package under `core/readmodel/agentmemory/` that loads:

- agent name
- configured `memoryPath`
- configured `rulesDir`
- whether the main memory file exists
- the current main memory file content
- whether the rules directory exists
- flat `.md` files inside the rules directory, sorted for preview

Rule-file preview is flat and non-recursive. That keeps the first version predictable and aligned with SkillFlow's current push behavior.

`sf-*.md` files are marked as SkillFlow-managed in the read model so the UI can distinguish managed module memories from other rule files.

## Frontend Behavior

Within **My Agents**, add a third section above skill areas:

- **Memory File**: path, open action, preview content, empty-state text when missing
- **Rules Directory**: path, open action, preview cards for `.md` files, managed badge for `sf-*`

The existing search box should also filter memory rule previews by file name or content so the page keeps one shared search affordance.

## Data and Persistence Changes

Persist these local-only agent fields in `config_local.json`:

- `agents[].memoryPath`
- `agents[].rulesDir`

This is machine-specific path data, so it belongs in local config and must remain excluded from backup and sync.

## Error Handling

- If an agent has no configured `memoryPath`, show the configured-missing empty state.
- If the memory file path is configured but the file does not exist, show the path plus a missing-file state.
- If `rulesDir` is configured but missing, show a missing-directory state.
- If preview loading fails, surface an inline error in **My Agents** without breaking the rest of the page.

## Scope

- `core/agentintegration/app`
- `core/config`
- `core/readmodel/agentmemory`
- `cmd/skillflow/`
- `cmd/skillflow/frontend/`
- `docs/features.md`
- `docs/features_zh.md`
- `docs/config.md`
- `docs/config_zh.md`

## Out Of Scope

- Editing agent memory directly inside **My Agents**
- Recursive rules-directory browsing
- Per-agent diff views between SkillFlow source memory and agent memory
