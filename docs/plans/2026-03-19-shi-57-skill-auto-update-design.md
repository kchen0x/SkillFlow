# Skill Auto Update Design

**Date:** 2026-03-19

## Context

SHI-57 requires installed git-backed skills in **My Skills** to support automatic update after the local starred-repo cache is refreshed.

Current behavior already supports:

- checking installed git-backed skills against the local cached repo clone
- manually updating a single installed skill from cache
- refreshing already-pushed agent copies during manual update
- force-pushing the updated skill to agents selected in **Auto Sync**

The missing behavior is:

- a user-controlled auto-update toggle on the **My Skills** page
- backend orchestration that runs the existing update flow automatically after repo-cache refresh
- agent-page status that still shows **Updatable** when an agent copy is now behind the installed library copy because that agent was not selected for auto sync
- frontend refresh after background auto updates

## Decision

1. Add a device-local boolean config field `autoUpdateSkills`.
2. Surface that toggle on the **My Skills** page beside the existing **Auto Sync** controls.
3. Trigger automatic installed-skill updates only after starred-repo cache refresh:
   - startup starred-repo refresh
   - manual single-repo refresh in **Starred Repos**
   - manual refresh-all in **Starred Repos**
4. Reuse the existing `UpdateSkill()` path instead of adding a second update implementation.
5. Keep auto-push semantics unchanged:
   - selected auto-sync agents receive the updated skill automatically
   - non-selected agents are not created or overwritten automatically
6. Extend agent-card `updatable` state so it also becomes true when the agent copy is behind the installed **My Skills** copy.
7. Emit a lightweight post-update event so **My Skills** and **My Agents** refresh after background auto updates.

## Config Model

- `autoUpdateSkills` is a local-only device behavior and belongs in `config_local.json`.
- It follows the same persistence split as:
  - `autoPushAgents`
  - `launchAtLogin`
  - local filesystem paths
- It must not be stored in `config.json` because synced devices may have different local repo caches, agent directories, and desired background mutation behavior.

## Trigger Flow

### Startup

1. App schedules startup background tasks.
2. Starred repos refresh from remote into local cache.
3. If `autoUpdateSkills` is enabled, SkillFlow finds installed git-backed skills sourced from the refreshed repos and auto-updates the ones whose cached SHA is newer.
4. After that, `CheckUpdates()` runs to refresh any remaining `LatestSHA` / `updatable` state.

This requires reordering startup background tasks so cache refresh happens before the installed-skill update check.

### Manual Starred Repo Refresh

1. User refreshes one repo or all repos in **Starred Repos**.
2. Repo cache update succeeds.
3. If `autoUpdateSkills` is enabled, SkillFlow auto-updates installed skills mapped to those repo URLs.
4. The backend emits a refresh event so the frontend reloads list state.

## Backend Update Orchestration

Add a helper dedicated to automatic installed-skill updates triggered by repo refresh, conceptually:

- load config and return immediately when `autoUpdateSkills` is disabled
- load installed skills
- normalize incoming repo URLs to the same canonical form used by installed metadata
- keep only installed git-backed skills whose `SourceURL` matches the refreshed repo set
- for each matching installed skill:
  - resolve latest cached SHA from the refreshed local repo clone
  - skip when the cache is unchanged for that skill
  - otherwise call `UpdateSkill(skillID)`

This keeps all update side effects centralized in one place:

- overwrite installed library files
- clear `LatestSHA`
- update `SourceSHA`
- refresh already-pushed agent copies
- force-update selected auto-sync agents
- schedule backup
- publish the new refresh event

## Agent Updatable State

Current agent-card `updatable` state is inherited from the correlated installed skill group and only means "the installed copy has a newer upstream cache revision available".

SHI-57 needs a second condition:

- the agent copy correlates to an installed library skill
- the installed library copy is already newer than the agent copy

That second condition must mark the agent entry as `updatable=true` even when the installed library skill itself is no longer updatable against upstream.

Implementation rule:

- keep installed-skill update detection based on git logical key + cached SHA
- additionally compare the agent candidate content key against the correlated installed skill content key during agent list aggregation
- if they differ, mark the agent entry as updatable

This preserves the requested behavior:

- auto-sync agents get updated immediately
- agents outside auto-sync continue to show **Updatable**
- users can still manually bring those agents up to date through the existing push flow

## Frontend Changes

### My Skills

- Add a dedicated `Auto Update` toggle in the existing auto-sync control band.
- Keep agent chips focused only on target selection for auto sync.
- Toggle save uses the same `GetConfig()` / `SaveConfig()` flow already used by the Dashboard.
- Failed saves roll back local optimistic state.
- Copy should explain:
  - startup refresh triggers auto update
  - manual starred-repo refresh triggers auto update
  - only selected auto-sync agents receive automatic push

### My Agents

- No new per-card update button is added in this work.
- The page keeps using the existing `Updatable` badge, but its backend meaning expands to include "agent copy is behind My Skills".

### Events

- Add a new backend event for installed-skill changes caused by updates, for example `skills.updated`.
- Dashboard and My Agents listen to it and reload silently.
- Manual card update and background auto update both benefit from the same UI refresh path.

## Error Handling

- Auto update is best-effort and non-transactional across repos or skills.
- A failed repo refresh prevents auto update for that repo only.
- A failed automatic installed-skill update:
  - logs a stable error message with source and skill identifiers
  - does not stop other eligible skills from updating
  - leaves UI state refreshable through the next check/load cycle
- Background auto update should not show the per-card success/error banner used by manual Dashboard updates.

## Logging

New backend logs should include:

- auto update trigger source
- repo URL or logical scope
- skill identifier and name
- started / completed / failed outcome
- failure reason

Secret values must never be logged.

## Scope

- `core/config` local config model and persistence
- `cmd/skillflow/app.go` startup ordering, starred-repo refresh hooks, auto-update orchestration
- `cmd/skillflow/skill_state.go` / `core/skill/index.go` agent updatable-state refinement
- `core/notify` event type for installed skill refresh
- Dashboard and My Agents frontend event reload + auto-update toggle UI
- docs and config reference updates required by repo policy

## Out Of Scope

- per-skill auto-update preferences
- per-agent auto-update preferences separate from auto-sync targets
- adding a new manual update button on **My Agents**
- changing the existing manual **Update** behavior on Dashboard cards
