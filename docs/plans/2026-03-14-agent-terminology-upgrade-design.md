# Agent Terminology Upgrade Design

**Date:** 2026-03-14
**Status:** Approved

## Goal

Rename SkillFlow's business concept of "tool" to "agent" across backend code, frontend code, persisted config schema, routes, Wails bindings, and user-facing documentation.

## Decisions

### 1. Use one-time startup migration, not runtime compatibility

- Create `core/upgrade/` for schema-cutover code.
- Run the upgrade before `config.NewService(...).Load()` inside `App.startup()`.
- Upgrade old config files in place, then let the rest of the app read only the new schema.
- Do not keep business-layer compatibility branches for old `tool*` fields or old Wails/API names.

### 2. Rename the persisted config schema to `agent` terminology

These persisted keys move to the new schema:

- `config.json`
  - `tools` -> `agents`
- `config_local.json`
  - `tools` -> `agents`
  - `autoPushTools` -> `autoPushAgents`
- `skillStatusVisibility`
  - `myTools` -> `myAgents`
  - `pushToTool` -> `pushToAgent`
  - `pullFromTool` -> `pullFromAgent`
  - `pushedTools` -> `pushedAgents`

The migration only touches authoritative persisted files. Local derived cache files under `cache/viewstate/` are not migrated; they should be invalidated or rebuilt automatically.

### 3. Rename business types and APIs end-to-end

Representative rename targets:

- `ToolConfig` -> `AgentConfig`
- `ToolAdapter` -> `AgentAdapter`
- `ToolSkillEntry` / `ToolSkillCandidate` -> `AgentSkillEntry` / `AgentSkillCandidate`
- `GetEnabledTools()` -> `GetEnabledAgents()`
- `ListToolSkills()` -> `ListAgentSkills()`
- `DeleteToolSkill()` -> `DeleteAgentSkill()`
- `ScanToolSkills()` -> `ScanAgentSkills()`
- `PullFromTool()` -> `PullFromAgent()`
- `PushToTools()` / `PushToToolsForce()` -> `PushToAgents()` / `PushToAgentsForce()`
- `CheckMissingPushDirs()` -> `CheckMissingAgentPushDirs()`

Wails bindings, generated TypeScript models, page state, and DTO fields must be updated in the same change so the codebase does not carry old names forward.

### 4. Keep technical package boundaries where they still make sense

- `core/sync/` remains the package path because it describes synchronization capability, not the old business term.
- Inside that package and its consumers, business-facing type names should move from `tool` to `agent`.
- Import aliases such as `toolsync` should move to `agentsync` to match the new terminology.

### 5. Rename frontend navigation, page names, and route names

- "My Tools" becomes "My Agents".
- "Push to Tool" becomes "Push to Agent".
- "Pull from Tool" becomes "Pull from Agent".
- Route `/tools` becomes `/agents`.
- File and component names should also follow the new terminology where they represent business concepts, for example `ToolSkills.tsx` -> `AgentSkills.tsx`.

Action-only routes such as `/sync/push` and `/sync/pull` can stay unchanged because they are not themselves the old business term.

### 6. Update all user-facing documentation and repository guidance

Update the following in the same implementation:

- `README.md`
- `README_zh.md`
- `docs/features.md`
- `docs/features_zh.md`
- `docs/architecture.md`
- `docs/architecture_zh.md`
- `docs/config.md`
- `docs/config_zh.md`
- `AGENTS.md`

`AGENTS.md` must gain a durable rule for future breaking schema/terminology upgrades:

- place cutover code under `core/upgrade/`
- run it automatically at startup before business initialization
- migrate persisted files directly to the new schema
- do not retain old-schema compatibility in business code

## Upgrade Flow

1. `App.startup()` resolves the app data directory.
2. Startup calls `core/upgrade` before creating or loading the config service.
3. The upgrade layer reads `config.json` and `config_local.json` as raw JSON.
4. If old `tool*` keys are present, it rewrites them to the new `agent*` schema.
5. If upgrade succeeds, normal startup continues and the app reads only new fields.
6. If upgrade fails, startup stops and logs a stable, searchable failure message.

## Logging Requirements

The upgrade must follow the repository logging policy:

- `info`: upgrade started / completed
- `error`: upgrade failed, including the reason
- never log tokens, passwords, or raw credential content

## Testing Strategy

### `core/upgrade`

- old-schema `config.json` migrates to `agents`
- old-schema `config_local.json` migrates to `agents` and `autoPushAgents`
- old `skillStatusVisibility` keys migrate to the new names
- rerunning the migration on already-upgraded files is a no-op
- malformed JSON or write failure returns an error

### `core/config`

- config loading and saving use only the new `agent` schema
- defaults and normalization use `AgentConfig`, `AutoPushAgents`, and `pushedAgents`
- no old-schema compatibility logic remains in config loading

### `cmd/skillflow`

- startup upgrade happens before config loading
- push/pull/scan/delete/check-missing operations work through renamed agent APIs
- auto-push and derived state use renamed config and DTO fields

### Frontend

- TypeScript compiles with the renamed Wails bindings and models
- route and page renames are wired correctly
- status visibility keys use the renamed page/status values
- user-visible strings all use "agent" / "智能体"

## Non-goals

- Do not rename external standards or concepts that are not SkillFlow's business term, such as explicit references to an external "MCP tool".
- Do not migrate local derived cache payloads field-by-field.
- Do not preserve backwards-compatible runtime reads of old config fields after startup migration.
