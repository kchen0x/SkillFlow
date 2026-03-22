# My Agents Skills Memory Toggle Design

**Date:** 2026-03-22

## Context

The current **My Agents** page shows the new memory preview and the existing skill sections in one continuous surface. That makes the page visually dense and mixes two different jobs:

- inspect pushed and scanned skills
- inspect the selected agent's memory file and rules directory

Users want a cleaner agent panel with an explicit way to switch between these views.

## Decision

1. Add a two-state segmented control at the top of **My Agents**:
   - `Skills`
   - `Memory`
2. Default the selected panel to `Skills`.
3. Keep the selected panel when the user switches to another agent.
4. Show only skill-related content in `Skills`:
   - Push Path section
   - Scan Path section
   - batch delete controls for pushed skills
5. Show only memory-related content in `Memory`:
   - Memory refresh action
   - main memory file preview
   - rules directory preview
6. Keep one shared search box, but scope the search results, result count, and empty-state copy to the active panel only.

## Why This Approach

- It resolves the clutter directly instead of adding more vertical grouping inside the same page.
- It preserves one agent-centric page instead of sending users to a separate route.
- It keeps the backend unchanged because the existing skill list and memory preview loaders are already sufficient.

## Interaction Rules

- Initial page load:
  - first enabled agent is selected
  - active panel is `Skills`
- Switching agents:
  - keeps the active panel unchanged
  - reloads skills and memory preview for the new agent
- Search behavior:
  - `Skills` searches pushed and scan-only skills only
  - `Memory` searches main memory and rule files only
- Result count:
  - `Skills` shows pushed plus scan-only visible items
  - `Memory` shows visible memory entries only
- Batch delete:
  - available only in `Skills`
  - hidden in `Memory`

## UI Placement

Place the segmented control in the page toolbar beside the selected agent identity so it is visible before the search and action row.

The visual treatment should match the rest of the app's existing highlighted toggle buttons instead of introducing a new design language.

## Scope

- `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- `cmd/skillflow/frontend/src/lib/`
- `cmd/skillflow/frontend/tests/`
- `cmd/skillflow/frontend/src/i18n/en.ts`
- `cmd/skillflow/frontend/src/i18n/zh.ts`
- `docs/features.md`
- `docs/features_zh.md`

## Out Of Scope

- backend API changes
- changes to **My Memory**
- a third `All` panel
- persisting the selected panel between app launches
