# Migration Blueprint

## Goal

Move the backend from the current technology-oriented package layout to the target DDD-oriented modular monolith without stopping product work.

The migration should be incremental. The target architecture is strict, but the transition plan should avoid a high-risk big-bang rewrite.

## Progress Checkpoint

As of 2026-03-21:

- `skillcatalog` is extracted under `core/skillcatalog/app`, `domain`, and `infra`
- `promptcatalog` is extracted under `core/promptcatalog/app`, `domain`, and `infra`
- `cmd/skillflow` and supporting backend packages no longer import `core/skill`
- `cmd/skillflow` no longer imports `core/prompt`
- the old `core/skill` and `core/prompt` packages have been removed

This means Phase 1 exit criteria are now met with two extracted core contexts. `skillcatalog` and `promptcatalog` are now the reference extraction patterns for the next migrations.

## Current-to-Target Mapping

| Current package / area | Target destination |
|------------------------|-------------------|
| `core/skill` | `core/skillcatalog/domain` + `core/skillcatalog/infra/repository` |
| `core/prompt` | `core/promptcatalog/domain` + `core/promptcatalog/infra/repository` |
| `core/sync` | `core/agentintegration/app/port/gateway` + `core/agentintegration/infra/gateway` |
| `core/git` starred-source parts | `core/skillsource/domain` with `StarRepo` and `SkillSource` models + `core/skillsource/infra` |
| `core/git` Git primitives | `core/platform/git` |
| `core/backup` provider and snapshot logic | `core/backup/domain` + `core/backup/infra` |
| `core/config` | `core/platform/settingsstore` + context-owned config namespaces |
| `core/applog` | `core/platform/logging` |
| `core/notify` | `core/platform/eventbus` |
| `core/pathutil` | `core/platform/pathutil` |
| `core/skillkey` | `core/shared/logicalkey` |
| `core/upgrade` | `core/platform/upgrade` |
| `core/viewstate` | `core/readmodel` or context-local `infra/projection` |
| `core/registry` | shell composition concerns in `cmd/skillflow` |
| `core/update` app-release pieces | `core/platform/update` plus shell-facing adapters in `cmd/skillflow` |
| `cmd/skillflow/app*.go` business methods | Wails transport adapters in `cmd/skillflow` delegating to context `app`, `orchestration`, and `readmodel` |
| `cmd/skillflow/process_*.go`, `tray_*.go`, `window_*.go` | remain in `cmd/skillflow` |

## Migration Phases

### Phase 0: Freeze Architectural Direction

- adopt this document set as the architectural target
- stop introducing new reusable business logic directly into `cmd/skillflow/app*.go`
- stop creating new horizontal utility packages that blur context ownership

### Phase 1: Introduce New Skeleton

- create `platform/`, `shared/`, `orchestration/`, and `readmodel/`
- create bounded-context directories with `app`, `domain`, and `infra`
- keep initial adapters thin where necessary, but treat them as temporary migration scaffolding

#### Exit criteria

- at least one context is extracted end-to-end into `app/domain/infra`
- the corresponding old package is marked deprecated for new feature work
- new business logic for that area no longer lands in the old package or in `cmd/skillflow/app*.go`

### Phase 2: Extract Core Domains First

Recommended order:

1. `skillcatalog`
2. `promptcatalog`
3. `agentintegration`

Reason:

- these are the primary core domains
- most UI-facing business behavior depends on them
- they define the truth that supporting domains consume

### Phase 3: Extract Supporting Domains

Recommended order:

1. `skillsource`
2. `backup`

Shell and platform concerns such as startup, tray, window state, launch-at-login, and app update should remain in `cmd/skillflow` plus `platform/` rather than being forced into a bounded context.

### Phase 4: Replace Cross-Cutting Structures

- split the old `AppConfig` into context-owned namespaces backed by a shared settings store
- move event publication onto application services and orchestration services
- migrate current view-state caches into `readmodel/` or context-local projections

### Phase 5: Shrink `cmd/skillflow`

At this stage, `cmd/skillflow` should mostly contain:

- Wails transport adapters
- process bootstrap
- tray/window integration
- shell coordination
- dependency wiring

Reusable business methods should no longer live there.

## Recommended First Extractions

### 1. `skillcatalog`

Move first:

- installed skill aggregate model
- category operations
- installed skill listing
- delete, move, and update use cases

Keep temporary adapters if needed, but route new Wails calls through transport methods that delegate into `skillcatalog/app`.

### 2. `promptcatalog`

Move second:

- prompt aggregate model
- category model
- import/export use cases
- prompt-related application services

### 3. `agentintegration`

Move third:

- agent profile model
- push/pull planners
- conflict detection
- scan and presence semantics

## Testing Strategy

- migrate unit tests together with the code they cover
- before extracting a module, add characterization tests for behavior that is poorly specified
- each migration phase should keep `go test ./core/...` passing
- for touched shell code, run targeted `go test ./cmd/skillflow/...`
- when orchestration is introduced, add focused integration tests around cross-context flows

## Anti-Patterns to Avoid During Migration

- copying the old `App` god-object into a new context package
- creating one global `service` package under `core/`
- introducing a shared mega-repository package
- letting one context import another context's `infra`
- leaving business truth in UI-facing DTO builders
- letting temporary thin wrappers become the permanent architecture

## Verification Criteria

The migration is on track when these statements become true:

- business use cases are discoverable by bounded context
- Wails `App` methods are thin transport adapters rather than business-method hosts
- settings ownership is explicit per context namespace
- cross-context writes happen in `orchestration/` or explicit shell coordination
- cross-context views happen in `readmodel/`
- old packages are steadily deprecated instead of quietly accepting new code forever

*Last updated: 2026-03-21*
