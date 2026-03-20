# Application Use Cases

## Purpose

This document defines which commands and queries belong to each bounded context, which flows should be handled through shared orchestration, and which UI-oriented views should be served through read models instead of domain services.

The main rule is:

- if an operation changes business truth owned by one context, it belongs to that context's application layer
- if an operation changes truth across multiple contexts, it should use explicit orchestration
- if an operation only assembles data for UI consumption, it belongs to `readmodel/`

## Use-Case Ownership by Context

## `skillcatalog`

### Commands

- `ImportLocalSkill`
- `CreateInstalledSkillFromSource`
- `DeleteInstalledSkill`
- `MoveInstalledSkillToCategory`
- `CreateSkillCategory`
- `RenameSkillCategory`
- `DeleteSkillCategory`
- `UpdateInstalledSkillFromSource`
- `ReconcileInstalledSkillVersionState`

### Queries

- `GetInstalledSkill`
- `ListInstalledSkills`
- `ListSkillCategories`
- `GetInstalledSkillVersionState`
- `FindInstalledSkillByLogicalKey`

### Notes

- `CreateInstalledSkillFromSource` should enforce the `repo + subpath -> one installed skill` constraint for repository-backed skills.
- `ImportLocalSkill` should generate `LogicalSkillKey` from a canonicalized content snapshot for manually imported skills and persist that key because there is no `SkillSourceRef` to derive from later.
- update checks that only refresh version state can stay in this context even if source hints come from `skillsource`.

## `promptcatalog`

### Commands

- `CreatePrompt`
- `UpdatePrompt`
- `DeletePrompt`
- `CreatePromptCategory`
- `RenamePromptCategory`
- `DeletePromptCategory`
- `MovePromptToCategory`
- `PreparePromptImport`
- `ApplyPromptImport`
- `ExportPromptBundle`

### Queries

- `GetPrompt`
- `ListPrompts`
- `ListPromptCategories`
- `PreviewPromptImportConflicts`

### Notes

- import sessions are application-flow objects, not aggregate roots
- current prompt import conflict behavior is name-based overwrite detection
- semantic prompt diffing is a future enhancement, not current domain behavior
- export formatting belongs to the application layer or infrastructure, not the prompt domain model

## `agentintegration`

### Commands

- `RegisterAgentProfile`
- `UpdateAgentProfile`
- `RemoveAgentProfile`
- `EnableAgentProfile`
- `DisableAgentProfile`
- `PushInstalledSkills`
- `PushInstalledSkillsForce`
- `PullSkillsFromAgent`
- `PullSkillsFromAgentForce`
- `DeleteAgentSkill`
- `ReconcileAutoPushPolicy`

### Queries

- `GetAgentProfile`
- `ListAgentProfiles`
- `ScanAgentSkills`
- `ListAgentSkills`
- `CheckMissingAgentPushDirs`
- `ResolveAgentSkillPresence`

### Notes

- push and pull conflict detection belongs here, not in shell adapters or source-management code
- this context should consume installed-skill summaries rather than holding `skillcatalog` aggregates directly

## `skillsource`

### Commands

- `TrackStarRepo`
- `TrackStarRepoWithCredentials`
- `UntrackStarRepo`
- `RefreshStarRepo`
- `RefreshAllStarRepos`
- `MarkStarRepoSyncFailure`
- `ClearStarRepoSyncFailure`

### Queries

- `GetStarRepo`
- `ListStarRepos`
- `ListSkillSourcesByRepo`
- `ListAllSkillSources`
- `ListSourceSkillCandidates`
- `GetSourceVersionHint`

### Notes

- `StarRepo` is the repository-level model
- `SkillSource` is the logical per-skill source model identified by `repo + subpath`
- installed/imported/updatable flags for source candidates should be derived by combining `skillsource` data with `skillcatalog` and `agentintegration` published language, usually in read models

## `backup`

### Commands

- `SaveBackupProfile`
- `RunBackup`
- `RunAutoBackup`
- `RestoreBackup`
- `ResolveGitBackupConflict`
- `RecordBackupSnapshot`

### Queries

- `GetBackupProfile`
- `ListRemoteBackupFiles`
- `GetLastBackupResult`
- `GetLastBackupCompletedAt`
- `PreviewBackupChanges`

### Notes

- backup owns backup execution semantics but not the post-restore business rebuild of every context
- restore compensation that spans contexts should use orchestration

## Shell and Platform Operations

The following operations are shell or platform concerns, not bounded-context use cases:

- `SetLaunchAtLogin`
- `PersistWindowState`
- `ShowMainWindow`
- `HideMainWindow`
- `CheckAppUpdate`
- `DownloadAppUpdate`
- `ApplyAppUpdate`
- `SetSkippedUpdateVersion`

They should remain in `cmd/skillflow/` and `platform/` rather than being modeled as a separate bounded context.

## Shared Orchestration

The following flows should not be owned by a single bounded context.

### `ImportSkillFromSourceOrchestrator`

Typical sequence:

1. read candidate source data from `skillsource`
2. create or reconcile installed skill in `skillcatalog`
3. optionally push to configured agents through `agentintegration`
4. optionally trigger backup through `backup`

### `ImportLocalSkillOrchestrator`

Typical sequence:

1. validate local source directory
2. create installed skill in `skillcatalog`
3. optionally push to configured agents
4. optionally trigger backup

### `UpdateInstalledSkillOrchestrator`

Typical sequence:

1. resolve source hint from `skillsource`
2. update installed skill content and version state in `skillcatalog`
3. refresh pushed agent copies through `agentintegration`
4. optionally trigger backup

### `RestoreSystemOrchestrator`

Typical sequence:

1. restore backup payload through `backup`
2. rebuild context-local settings, caches, and projections
3. refresh derived read models
4. emit post-restore events

## Shell Coordination

These flows are coordination concerns, but they are better owned by the shell composition layer than by `core/orchestration/`.

### `SettingsSaveCoordinator`

`Settings` is a composition surface. Saving it should dispatch to multiple contexts:

- skill-library settings -> `skillcatalog`
- prompt-library settings -> `promptcatalog`
- agent profiles and auto-push -> `agentintegration`
- source credentials and tracking settings -> `skillsource`
- backup provider settings -> `backup`
- shell preferences -> `cmd/skillflow` and `platform/`

### `StartupBootstrapSequence`

Startup sequencing should live in `cmd/skillflow/bootstrap.go` or equivalent shell bootstrap code:

- load settings
- initialize shell adapters
- refresh or repair runtime state
- schedule background source refresh, update checks, and backup tasks

## Read Models

The following views should live in `readmodel/` rather than inside one bounded context.

### `DashboardReadModel`

Combines:

- installed skills from `skillcatalog`
- version hints from `skillsource`
- pushed state from `agentintegration`

### `MyAgentsReadModel`

Combines:

- agent profiles from `agentintegration`
- agent-side scan results from `agentintegration`
- installed-state overlays from `skillcatalog`

### `SettingsReadModel`

Combines context-owned settings from:

- `skillcatalog`
- `promptcatalog`
- `agentintegration`
- `skillsource`
- `backup`
- shell/platform settings namespaces

### `StarRepoReadModel`

Combines:

- tracked repositories and skill sources from `skillsource`
- installed-state overlays from `skillcatalog`
- pushed-state overlays from `agentintegration`

## Transport Adapter Mapping

The current Wails-facing `App` methods in `cmd/skillflow/` should gradually become thin transport adapters that delegate to context application services, orchestration services, or read models.

Examples:

- `ListSkills` -> `readmodel/dashboard` or `skillcatalog/query`
- `ImportStarSkills` -> `orchestration/ImportSkillFromSourceOrchestrator`
- `PushToAgents` -> `agentintegration/app`
- `CreatePrompt` -> `promptcatalog/app`
- `CheckAppUpdate` -> shell/platform update service

## Design Constraints

- no use case should directly depend on another context's infrastructure implementation
- command handlers should return domain-oriented results, then let transport adapters map them to transport DTOs
- cross-context writes must remain explicit
- UI labels such as `imported` may differ from internal semantics such as `installed`, but the mapping belongs outside the domain layer

*Last updated: 2026-03-20*
