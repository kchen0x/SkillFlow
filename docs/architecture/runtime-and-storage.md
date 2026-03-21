# Runtime, Repository Layout, and Storage

## Wails Shell Constraints

SkillFlow remains a Wails desktop application.

Current and future architecture must preserve these constraints:

- repository root contains no Go source files
- `cmd/skillflow/*.go` stays flat because Wails bindings require a single `package main` directory
- Wails-bound transport adapters remain in `cmd/skillflow/`
- the desktop shell binds backend use cases directly to the frontend instead of exposing a REST API

## `cmd/skillflow/` Responsibilities After Refactor

`cmd/skillflow/` should become a shell-oriented composition layer.

It should contain:

- Wails startup and binding registration
- Wails-facing transport adapters
- helper/UI process bootstrapping
- tray and window integration
- single-instance coordination
- shell-level startup sequencing
- settings-save fan-out coordination where multiple contexts must be updated together
- dependency assembly for backend contexts, orchestration services, and read models

It should not remain the home for reusable business use cases such as skill import rules, prompt CRUD rules, source synchronization rules, or agent push/pull semantics.

## Helper and UI Runtime Split

The current helper/UI split remains valid after the DDD refactor:

- the `helper` process owns tray presence, local control endpoints, and UI relaunch/focus
- the `ui` process hosts Wails, React, and the transport adapters that call backend application services
- closing the main window exits the UI process without killing the helper shell

The DDD refactor changes where backend logic lives, but not this shell topology.

## Target Repository Shape

```text
/
  go.mod
  Makefile
  docs/
    architecture.md
    architecture_zh.md
    architecture/
  core/
    platform/
    shared/
    orchestration/
    readmodel/
    skillcatalog/
    promptcatalog/
    agentintegration/
    skillsource/
    backup/
  cmd/
    skillflow/
      main.go
      bootstrap.go
      app.go
      app_*.go
      process_*.go
      tray_*.go
      window_*.go
      frontend/
```

## Storage Direction

Logical ownership should be split by bounded context, but physical storage should remain operationally simple.

Recommended target layout:

```text
<AppDataDir>/
  config.json          # shared, namespaced by context
  config_local.json    # local-only, namespaced by context and shell concerns
  star_repos.json      # tracked repository state for skillsource
  skills/
    library/
    meta/
    meta_local/
  prompts/
    library/
  cache/
    sources/
    readmodel/
  runtime/
  logs/
```

Within `config.json` and `config_local.json`, each context should read and write its own namespace through a platform settings store. The physical files do not need to mirror bounded contexts 1:1.

## Configuration Ownership

The old `AppConfig` structure should be treated as a transitional compatibility structure, not the long-term domain model.

Recommended logical ownership:

- `skillcatalog`
  - skill library location
  - default skill category
- `promptcatalog`
  - prompt library location
  - prompt-specific import/export defaults
- `agentintegration`
  - agent profiles
  - auto-push policy
- `skillsource`
  - source credentials metadata
  - source refresh defaults
- `backup`
  - active backup profile
  - provider selection and interval
- shell/platform
  - launch-at-login
  - window state
  - skipped update version
  - proxy and log-level preferences

Current migration note:

- app-data path ownership now lives in `core/platform/appdata`
- shell proxy, window, log-level, and skipped-update preferences now live in `core/platform/shellsettings` plus `core/platform/settingsstore`
- skill-status visibility preferences now live in `core/readmodel/preferences`
- cross-context write flows for import, push/pull, update, and restore compensation now live in `core/orchestration`
- installed-skill, starred-skill, and agent-presence composition now lives in `core/readmodel/skills` plus `core/readmodel/viewstate`
- `core/config` remains as the frontend-facing compatibility DTO and split/merge persistence facade around these context- and platform-owned settings components

## Repository vs Gateway Examples

Repositories:

- installed skill metadata store
- prompt library store
- agent profile store
- source-tracking store
- namespaced settings store views

Gateways:

- agent workspace adapter
- Git client wrapper used to sync external sources
- cloud backup provider adapter
- GitHub release API client
- Wails runtime adapters such as file dialogs or shell open operations

## Events and Derived State

Event forwarding to the frontend remains a shell integration concern, but event publication should move closer to application services and orchestration services.

Derived snapshots such as installed-skill cards or aggregated agent presence should live under `readmodel/` or context-local `infra/projection/`, not inside domain models.

## Logging and Path Portability

Existing constraints remain in force:

- logs stay bounded to `skillflow.log` and `skillflow.log.1`
- synced paths should be stored as forward-slash relative paths when under the synchronized root
- local-only machine-specific paths stay in local settings namespaces
- secrets must never be written to logs

*Last updated: 2026-03-21*
