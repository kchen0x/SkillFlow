# Agent Skill Sync Consolidation Design

## Context

The current app still exposes three separate user flows for one conceptual area:

- managing installed skills in **My Skills**
- pushing installed skills into agents from **Push to Agents**
- pulling scanned agent skills into the library from **Pull from Agents**

That split creates unnecessary navigation overhead and duplicates interaction patterns that already exist elsewhere in the app:

- **My Skills** already owns category filtering, search, sorting, and skill-card selection behavior.
- **My Agents** already owns agent switching, push-path skill browsing, scan-path skill browsing, and memory preview.
- **My Memory** already established the preferred inline batch-action pattern instead of sending users to a dedicated page.

There is also a consistency issue in Chinese mode: several primary labels still mix Chinese with English terms such as `skills` or `Skills | Memory`.

## Goals

- Reduce the main sidebar to four primary work surfaces:
  - **我的技能**
  - **我的提示词**
  - **我的记忆**
  - **我的智能体**
- Keep **仓库收藏 / 云备份 / 设置 / 意见反馈** available in the lower navigation area.
- Merge manual skill push into **My Skills** as an inline batch-push flow.
- Merge manual skill pull into **My Agents** as an inline batch-pull flow.
- Reuse existing backend transport methods wherever possible.
- Remove Chinese and English mixing in the changed primary UI surfaces.
- Preserve old `/sync/push` and `/sync/pull` routes through redirects instead of hard deletion.

## Non-Goals

- No new backend domain model for push or pull jobs.
- No new persisted config for manual push or manual pull selections.
- No redesign of Starred Repos, Backup, or Settings beyond keeping their navigation entry points.
- No change to the existing automatic push settings semantics in **My Skills**.
- No change to the existing agent memory preview behavior in **My Agents**.

## Navigation Design

### Primary Sidebar

The top section of the left sidebar will expose only:

- `/` => **我的技能**
- `/prompts` => **我的提示词**
- `/memory` => **我的记忆**
- `/tools` => **我的智能体**

### Secondary Sidebar Area

The bottom section remains:

- `/starred` => **仓库收藏**
- `/backup` => **云备份**
- `/settings` => **设置**
- feedback button => **意见反馈**

This keeps the primary section aligned with the user's day-to-day working objects while preserving access to utility areas.

### Legacy Routes

To avoid breaking existing links or assumptions in docs and internal navigation:

- `/sync/push` redirects to `/`
- `/sync/pull` redirects to `/tools`

The dedicated pages can remain in the codebase short-term if that lowers migration risk, but they should no longer be reachable from the sidebar.

## UX Design

### My Skills: Inline Manual Push

The existing **Auto Push Targets** strip remains unchanged in purpose. It continues to represent persistent local-only settings for automatic propagation.

Add a new toolbar action: **手动推送**.

When users click it, **My Skills** enters an inline batch-push mode similar to **My Memory**:

- skill cards become selectable
- the current category filter, search term, and sort order define the candidate set
- a target-agent strip appears in-page
- toolbar actions switch to:
  - `取消`
  - `开始推送`

#### Batch-Push Rules

- Users may select zero or more enabled agents before pushing.
- Users may select zero or more visible skill cards.
- `开始推送` is enabled only when at least one agent and one skill are selected.
- `全选 / 取消全选` affects only the currently visible filtered skills.
- Exiting batch-push mode clears all temporary selections.
- Manual push does not change the persistent automatic push target selection.

#### Push Execution

Push execution continues to use the current backend flow:

- `CheckMissingAgentPushDirs`
- `PushToAgents`
- `PushToAgentsForce` through the existing conflict dialog

This keeps all directory creation and overwrite semantics unchanged while moving the entry point into the right page.

### My Agents: Inline Manual Pull

The page keeps its existing top-level shape:

- left column = enabled-agent list
- main area = selected agent surface
- segmented control = **技能 / 记忆**

The segmented labels should be localized in Chinese mode and no longer display `Skills | Memory`.

Add a new action inside the **技能** panel: **手动拉取**.

When clicked, **My Agents** enters an inline pull mode for the currently selected agent:

- the page scans that agent's configured `ScanDirs`
- the content area switches from normal browse mode to scan-result selection mode
- target category is chosen directly in the top toolbar instead of a separate left import sidebar
- toolbar actions switch to:
  - `取消`
  - `开始拉取`

#### Pull-Mode Rules

- Pull mode always operates on the currently selected agent only.
- Entering pull mode triggers a fresh scan.
- Search and sorting apply to scanned candidates only.
- `全选 / 取消全选` affects only the currently visible scanned cards.
- `选择未导入` remains available and uses the scanned `imported` flag.
- `开始拉取` is enabled only when at least one scanned skill is selected.
- Exiting pull mode clears temporary selection, scan errors, and completion state.

#### Pull Execution

Pull execution continues to use the current backend flow:

- `ScanAgentSkills`
- `PullFromAgent`
- `PullFromAgentForce` through the existing conflict dialog

The target category is supplied inline from the selected category control.

### Normal My Agents Browsing

Outside pull mode, the existing behaviors remain available:

- view skills already present in the agent push path
- view scan-only skills detected from agent scan directories
- batch-delete pushed copies from the agent push path
- preview memory and rules in the **记忆** panel

The merge should add pull capability without removing current inspection and cleanup capabilities.

## Data and State Design

### My Skills Temporary State

Add page-local temporary state for manual push:

- `manualPushMode: boolean`
- `selectedPushAgents: string[]`
- `selectedSkillIDs: string[]`
- `pushing: boolean`
- `pushDone: boolean`
- `missingDirs`
- `pushConflicts`

This state is ephemeral and must not be persisted.

### My Agents Temporary State

Add page-local temporary state for manual pull:

- `pullMode: boolean`
- `pullTargetCategory: string`
- `scannedCandidates: AgentSkillCandidate[]`
- `selectedCandidatePaths: string[]`
- `scanning: boolean`
- `pulling: boolean`
- `pullDone: boolean`
- `scanError`
- `pullConflicts`

This state is also ephemeral and isolated from the normal browsing state.

### Shared Helper Logic

Move toggle-heavy selection rules into small testable frontend helpers instead of leaving everything inline in page components.

Expected helper coverage:

- dashboard manual-push mode transitions and validation
- tool-skills manual-pull mode transitions and validation
- route normalization or nav metadata if needed

## Backend Design

No new backend transport API is required for the requested UX.

Existing methods already cover the merged flows:

- `GetEnabledAgents`
- `ListSkills`
- `ListCategories`
- `CheckMissingAgentPushDirs`
- `PushToAgents`
- `PushToAgentsForce`
- `ListAgentSkills`
- `ScanAgentSkills`
- `PullFromAgent`
- `PullFromAgentForce`

Because the backend contract stays stable, this change is primarily a frontend composition refactor plus documentation update.

## Testing Design

### Frontend

Add or update unit tests for pure helper behavior:

- primary dashboard toolbar action order includes the new manual push entry
- dashboard manual push mode enters, exits, and validates correctly
- My Agents pull mode helper resets and validates correctly
- tool-skills panel helpers still scope search and visible counts correctly when the segmented control remains active

Then verify with:

- `npm run test:unit`
- `npm run build`

### Backend

Because no backend behavior changes are planned, full backend regression coverage should come from:

- `go test ./core/... ./cmd/skillflow`

This protects against accidental binding or DTO breakage while frontend wiring changes around Wails APIs.

## Documentation Impact

This is a user-facing flow change and must update, in the same change set:

- `docs/features.md`
- `docs/features_zh.md`

Expected documentation changes:

- sidebar route table
- removal of dedicated primary navigation for **Push to Agents** and **Pull from Agents**
- new inline **手动推送** flow under **My Skills**
- new inline **手动拉取** flow under **My Agents**
- updated Chinese labels to remove mixed-language wording on affected primary controls

`README.md` and `README_zh.md` should be updated only if the high-level product description explicitly mentions separate push/pull pages. Otherwise they can remain unchanged.
