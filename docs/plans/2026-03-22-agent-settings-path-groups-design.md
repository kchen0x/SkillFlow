# Agent Settings Path Groups and Modal Creation Design

**Date:** 2026-03-22

## Context

The current **Settings → Agents** page keeps each agent's path fields in one flat card:

- `pushDir`
- `memoryPath`
- `rulesDir`
- `scanDirs`

That shape works functionally, but it mixes two different concerns:

- skill distribution paths
- memory synchronization paths

The current custom-agent creation flow also has two problems:

1. it is an inline form instead of a focused dialog
2. it only collects `name + pushDir` and immediately calls `AddCustomAgent(name, pushDir)`

Immediate persistence is the wrong write boundary for this page because Settings already uses a draft-first model: users edit multiple fields and then click **Save Settings** once. Calling `AddCustomAgent` or `RemoveCustomAgent` in the middle of that flow can overwrite or discard other unsaved settings edits.

The user requested:

- split each agent into separate **Skill paths** and **Memory paths** modules
- change custom-agent creation to a modal dialog
- keep the ability to configure multiple `scanDirs`
- keep the existing `memoryPath + rulesDir` pair, only grouped under a dedicated memory module

## Decision

1. Keep the persisted agent model unchanged:
   - `name`
   - `enabled`
   - `custom`
   - `pushDir`
   - `scanDirs`
   - `memoryPath`
   - `rulesDir`
2. Refactor the Agents tab UI so every agent card contains two explicit modules:
   - **Skill Paths**
   - **Memory Paths**
3. Put these fields into **Skill Paths**:
   - `pushDir`
   - `scanDirs`
4. Put these fields into **Memory Paths**:
   - `memoryPath`
   - `rulesDir`
5. Replace the inline custom-agent form with a centered modal dialog opened from an **Add Custom Agent** button.
6. The modal collects:
   - agent name
   - skill path (`pushDir`)
   - memory file (`memoryPath`)
   - rules directory (`rulesDir`)
7. Saving the modal appends a new custom agent to the current in-memory Settings draft only. It does not persist immediately.
8. When a custom agent is created from the modal, initialize:
   - `scanDirs = [pushDir]`
   - `enabled = true`
   - `custom = true`
9. Keep the existing multi-`scanDir` editing UI inside the **Skill Paths** module so users can add more scan paths after the custom agent appears in the list.
10. Custom-agent deletion should also mutate only the current draft. Final persistence remains the page-level **Save Settings** action.

## Why This Shape

### Separate modules are the simplest correct model

The underlying data already distinguishes skill paths from memory paths. The current problem is presentation, not schema. Splitting the UI into two modules exposes the model the app already has, without adding migration risk or backend churn.

### Modal creation should follow the page write boundary

The page already batches edits into `cfg` and persists them through `SaveConfig(cfg)`. Custom-agent creation should obey the same boundary. A modal that mutates only the current Settings draft avoids accidental loss of unrelated unsaved edits and keeps user expectations consistent.

### Multi-scan support stays where it matters

The modal should stay narrow and focused. The user only asked for name plus memory and skill paths during creation, not a full advanced scan-path editor. The correct compromise is:

- create the agent with one initial scan path derived from `pushDir`
- preserve full multiple-`scanDir` editing on the card immediately after creation

This keeps the creation flow short without removing the advanced capability.

## UX Flow

### Existing agent cards

For each agent:

- top row keeps icon, name, and enable toggle
- first module is **Skill Paths**
  - push directory input + picker
  - scan directory list with picker and delete per row
  - add-scan-directory row
- second module is **Memory Paths**
  - memory file input + picker
  - rules directory input + picker
- custom agents keep their delete action

### Add custom agent

1. User clicks **Add Custom Agent**
2. Modal opens
3. User fills:
   - name
   - skill path
   - memory file
   - rules directory
4. User clicks **Save**
5. Modal validates and, if valid:
   - appends a new custom agent into `cfg.agents`
   - seeds `scanDirs` from `pushDir`
   - closes the modal
   - clears the modal draft
6. The new card appears immediately in the current Settings view
7. User may add extra scan paths before clicking the page-level **Save Settings**

## Validation and Error Handling

- Modal save is blocked when any required field is blank.
- Modal save is blocked when the agent name duplicates an existing agent name after trimming.
- Path fields use the existing folder-picker behavior where applicable.
- No new backend validation or config migration is required.

Notes:

- `pushDir` is a directory path.
- `rulesDir` is a directory path.
- `memoryPath` remains a file path but can continue using the existing folder-picker-based path assist pattern if the implementation keeps that behavior unchanged.

## Scope

- `cmd/skillflow/frontend/src/pages/Settings.tsx`
- `cmd/skillflow/frontend/src/components/ui/AnimatedDialog.tsx` usage only
- `cmd/skillflow/frontend/src/i18n/en.ts`
- `cmd/skillflow/frontend/src/i18n/zh.ts`
- frontend unit/source tests under `cmd/skillflow/frontend/tests/`
- `docs/features.md`
- `docs/features_zh.md`

## Out Of Scope

- changing agent config schema
- changing `SaveConfig` backend behavior
- removing `AddCustomAgent` / `RemoveCustomAgent` backend APIs
- adding multi-`scanDir` editing directly inside the creation modal
- changing README wording
