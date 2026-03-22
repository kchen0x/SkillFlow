# My Memory Batch Push and Auto Sync Redesign

## Context

The initial **My Memory** implementation mixed three concerns inside the right-side drawer:

- authoring memory content
- configuring automatic push behavior
- manually pushing memory to agents

That coupling creates two UX problems:

- editing is interrupted by transport controls that do not belong to the editing task
- manual push and automatic push are configured in different places with overlapping concepts

There is also a functional bug: the preview tab does not render Markdown and instead shows plain text.

## Goals

- Keep the edit drawer focused on authoring only.
- Warn before closing the drawer when unsaved edits exist.
- Make the preview tab render Markdown.
- Move automatic sync configuration to the top of the **My Memory** page.
- Replace direct one-click push with an in-page batch-push selection flow.
- Keep main memory mandatory in batch push so the pushed main-memory file always maintains the module reference block for the selected modules.

## Non-Goals

- No background task system or push-job history.
- No per-module persistent push-target configuration.
- No support for pushing modules without also updating main memory.
- No attempt to preserve legacy `memory_local.json.modules` semantics in business code after cutover.

## UX Design

### Default Page State

The default **My Memory** page shows:

- the existing search box and memory cards
- a top **Auto Sync** panel with one row per enabled agent
- `New Module` and `Batch Push` actions on the right side of the toolbar

Each agent in the top panel exposes three persistent choices:

- `Off`
- `Auto Merge`
- `Auto Takeover`

These choices map to the existing per-agent local push config:

- `Off` => `autoPush=false`
- `Auto Merge` => `autoPush=true`, `mode=merge`
- `Auto Takeover` => `autoPush=true`, `mode=takeover`

### Batch Push Selection State

Clicking `Batch Push` does not open a modal. It switches the page into an inline selection state.

In selection state:

- every memory card shows a checkbox in the top-right corner
- the main-memory card is always selected and cannot be unchecked
- module-memory cards can be selected or deselected
- the top panel switches from **Auto Sync** to **Push Targets**
- users can multi-select target agents
- users choose one push mode for the entire push operation: `Merge` or `Takeover`
- the right-side toolbar actions become `Cancel` and `Start Push`

### Push Semantics

Batch push is an explicit snapshot operation for the selected agents:

- main memory is always included
- only the selected module memories are pushed
- non-selected managed module files are removed from the target agents
- the main-memory content is rebuilt so its managed module references match the selected module set

This keeps agent state internally consistent for agents that require explicit rule-file references in the main memory file.

### Edit Drawer

The edit drawer keeps only editing controls:

- `Edit`
- `Preview`
- `Save`
- `Delete Module` for module memories
- `Open in Editor`

Removed from the drawer:

- push targets
- push mode
- auto-push toggle
- push-now action

### Unsaved Changes Confirmation

When the user tries to close the drawer with unsaved changes, show a confirmation dialog with:

- `Discard`
- `Save and Close`
- `Keep Editing`

If there are no unsaved changes, closing remains immediate.

### Markdown Preview

The preview tab must render Markdown instead of showing raw text. The supported rendering surface should cover the markdown used in memory files:

- headings
- paragraphs
- unordered and ordered lists
- blockquotes
- fenced code blocks
- inline code
- links

Raw HTML is not a requirement.

## Data and Persistence Design

### Local Config

`memory/memory_local.json` keeps only:

- `pushConfigs`
- `pushState`

The legacy `modules` section is removed because module push targets are no longer persistent user configuration.

### Startup Cutover

Because the on-disk semantics change, the repository rules require an explicit startup cutover.

The startup upgrade should:

- load `memory/memory_local.json` if present
- remove the `modules` key
- rewrite the file in place only when a change is needed

After the cutover, business code reads only the new schema.

### Push State Semantics

`pushState.<agent>.lastPushedHash` remains the hash of the actual content most recently pushed to that agent.

This supports both cases:

- full automatic sync of all local memory => status becomes `synced`
- partial batch push of selected modules => status remains `pendingPush` because the stored pushed hash differs from the current full-library hash

## Backend Design

### Automatic Sync

Automatic sync should run after successful local mutations:

- save main memory
- create module memory
- save module memory
- delete module memory

For each enabled agent whose config is `autoPush=true`, SkillFlow performs a full push of main memory plus all current module memories.

Automatic sync is best-effort:

- the content mutation succeeds independently
- push failures are logged and reflected through status refresh

### Batch Push API

Add a new backend method for manual batch push instead of reusing the existing `PushAllMemory` signature.

Suggested transport shape:

```go
PushSelectedMemory(agentTypes []string, moduleNames []string, mode string) ([]*PushResultDTO, error)
```

Rules:

- `agentTypes` must be non-empty
- `moduleNames` contains only selected module memories
- main memory is always included implicitly
- `mode` applies uniformly to every target agent for that operation
- the temporary push mode does not overwrite persistent auto-sync config

### Push Service Extension

Add a selection-aware push path in `core/memorycatalog/app/push_service.go`.

The push service should support:

- full push for one agent using persistent config
- selection push for one agent using an override mode and a selected module set
- batch push to many agents using the same selected module set and mode

For selection push, the service should:

- load main memory
- resolve the selected modules by name
- remove every non-selected managed module file from the target agent
- write selected managed module files
- rebuild the rules index from the selected modules
- push main memory with the supplied override mode
- store the pushed hash for the selected snapshot

## Frontend Design

### Page State Model

`Memory.tsx` should distinguish:

- normal browsing state
- batch-push selection state

Selection state needs:

- selected module names
- selected target agents
- temporary batch-push mode

It should not reuse the persistent auto-sync config state.

### Derived Behavior

- entering batch-push mode preselects no modules and no agents
- the main-memory card is visually selected by default
- `Start Push` is enabled only when at least one module and one target agent are selected
- exiting batch-push mode clears temporary selections

### Preview Rendering

Use a small local Markdown renderer helper instead of adding a new dependency.

Reasoning:

- the repository already has unit-test infrastructure for pure frontend helpers
- avoiding a new package keeps the change self-contained and avoids dependency churn
- the required Markdown subset is small and predictable

## Testing Strategy

### Frontend

Add unit tests for pure helpers covering:

- auto-sync mode mapping
- batch-push state rules
- main-memory mandatory selection behavior
- Markdown rendering for headings, lists, blockquotes, fenced code blocks, inline code, and links

### Backend

Add tests covering:

- batch push writes only selected modules and updates the main-memory reference block
- automatic sync pushes all memories to enabled auto-sync agents after save
- startup upgrade removes legacy `modules` from `memory_local.json`

### Documentation

Update in the same change:

- `docs/features.md`
- `docs/features_zh.md`
- `docs/config.md`
- `docs/config_zh.md`

The feature docs must describe the new page flow, and the config docs must remove the old `modules` schema.
