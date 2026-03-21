# Runtime, Repository Layout, and Storage

## Wails Shell Constraints

SkillFlow remains a Wails desktop application.

The backend architecture preserves these constraints:

- repository root contains no Go source files
- `cmd/skillflow/*.go` stays flat because Wails bindings require a single `package main` directory
- Wails-bound transport adapters remain in `cmd/skillflow/`
- the desktop shell binds backend use cases directly to the frontend instead of exposing a REST API

## `cmd/skillflow/` Responsibilities

`cmd/skillflow/` is the shell-oriented composition layer.

It contains:

- Wails startup and binding registration
- Wails-facing transport adapters
- helper/UI process bootstrapping
- tray and window integration
- single-instance coordination
- shell-level startup sequencing
- settings-save fan-out coordination where multiple contexts must be updated together
- dependency assembly for backend contexts, orchestration services, and read models

It is not the home for reusable business use cases such as skill import rules, prompt CRUD rules, source synchronization rules, or agent push/pull semantics.

## Helper and UI Runtime Split

SkillFlow uses a helper/UI split:

- the `helper` process owns tray presence, local control endpoints, and UI relaunch/focus
- the `ui` process hosts Wails, React, and the transport adapters that call backend application services
- closing the main window exits the UI process without killing the helper shell

## Repository Shape

```text
/
  go.mod
  go.sum
  Makefile
  README.md
  README_zh.md
  docs/
    architecture/
    config.md
    config_zh.md
    features.md
    features_zh.md
    plans/
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
    config/
  cmd/
    skillflow/
      main.go
      bootstrap.go
      app.go
      app_*.go
      adapters.go
      providers.go
      events.go
      process_*.go
      tray_*.go
      window_*.go
      single_instance_*.go
      frontend/
```

## Storage Layout

Logical ownership is split by bounded context, while physical storage remains operationally simple.

Current persisted layout:

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

Within `config.json` and `config_local.json`, each context reads and writes its own namespace through the platform settings store. The physical files do not need to mirror bounded contexts 1:1.

## Configuration Ownership

`config.json` and `config_local.json` are shared storage files. Ownership of the fields inside them is still split by context and platform concern.

Logical ownership:

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

Implementation ownership:

- app-data path ownership lives in `core/platform/appdata`
- shell proxy, window, log-level, and skipped-update preferences live in `core/platform/shellsettings` plus `core/platform/settingsstore`
- skill-status visibility preferences live in `core/readmodel/preferences`
- cross-context write flows for import, push/pull, update, and restore compensation live in `core/orchestration`
- installed-skill, starred-skill, and agent-presence composition lives in `core/readmodel/skills` plus `core/readmodel/viewstate`
- `core/config` is the frontend-facing settings facade and split/merge persistence adapter around these context- and platform-owned settings components

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

Event forwarding to the frontend remains a shell integration concern. Event publication belongs close to application services and orchestration services.

Derived snapshots such as installed-skill cards or aggregated agent presence should live under `readmodel/` or context-local `infra/projection/`, not inside domain models.

## Logging and Path Portability

Existing constraints remain in force:

- logs stay bounded to `skillflow.log` and `skillflow.log.1`
- synced paths should be stored as forward-slash relative paths when under the synchronized root
- local-only machine-specific paths stay in local settings namespaces
- secrets must never be written to logs

*Last updated: 2026-03-21*
