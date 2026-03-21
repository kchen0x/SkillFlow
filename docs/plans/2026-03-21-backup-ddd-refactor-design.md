# Backup DDD Refactor Design

**Date:** 2026-03-21

## Context

`skillcatalog`, `promptcatalog`, `agentintegration`, and `skillsource` have already been extracted. The next recommended migration target from [`docs/architecture/migration.md`](/Users/shinerio/Workspace/code/SkillFlow/docs/architecture/migration.md) is `backup`.

The current backup backend is still split across two places:

- [`core/backup`](/Users/shinerio/Workspace/code/SkillFlow/core/backup), which currently mixes:
  - provider gateway contracts
  - remote file and credential field value objects
  - snapshot building and diffing
  - backup path filtering rules
  - git backup provider conflict semantics
  - provider factory registration
- [`cmd/skillflow/app.go`](/Users/shinerio/Workspace/code/SkillFlow/cmd/skillflow/app.go), which currently hosts:
  - backup execution entrypoints
  - provider initialization and dispatch
  - backup preview computation
  - restore execution
  - git conflict resolution
  - last-backup result caching
  - startup git pull flow

This means the `backup` business truth still lives half in shell code and half in a flat technical package.

## Decision

This migration extracts backup semantics into:

- `core/backup/domain`
- `core/backup/app`
- `core/backup/infra/provider`
- `core/backup/infra/snapshot`

This round is a **direct context extraction with shell restore-compensation retained**:

- the old flat `core/backup` package will be removed after callers switch
- `backup` will own backup execution semantics, snapshot comparison, provider metadata, restore execution, and git-conflict semantics
- shell code will still coordinate post-restore cross-context compensation such as reloading skill state, auto-pushing restored skills, and cloning restored source caches
- provider registration will stay shell-owned, but providers will be instantiated from `backup/infra/provider`

That boundary lets `backup` own its real use cases without prematurely pulling cross-context rebuild logic into it.

## Scope

### In scope

- move provider-facing contracts and backup result/conflict value objects into `backup/domain`
- move snapshot building, diffing, path filtering, and nested `.git` migration into `backup/infra/snapshot`
- move git provider and cloud providers into `backup/infra/provider`
- move run/list/restore/resolve-conflict/preview use cases into `backup/app`
- update shell code to delegate backup operations into `backup/app`
- migrate tests and remove the old flat `core/backup` package

### Out of scope

- moving cloud config persistence out of `core/config`
- introducing `core/orchestration/RestoreSystemOrchestrator`
- refactoring shell event hub ownership into `platform/eventbus`
- changing Backup page behavior or provider UX

## Target Structure

### `core/backup/domain`

Owns backup-side business language:

- `RemoteFile`
- `CredentialField`
- `Snapshot`
- `SnapshotEntry`
- `GitConflict`
- provider metadata and operation result types

This layer should define the meaning of:

- backup change actions
- restore conflict payloads
- backup preview results

### `core/backup/app`

Owns use cases:

- run backup
- list remote backup files
- restore backup
- resolve git conflict
- preview backup changes
- compute backup root and snapshot persistence

This layer should not own post-restore business rebuild for skills, prompts, or source caches.

### `core/backup/infra/provider`

Owns provider adapters:

- Git provider
- Aliyun / Tencent / Huawei / AWS / Azure / Google providers

### `core/backup/infra/snapshot`

Owns technical backup-storage helpers:

- snapshot build/load/save/diff
- path exclusion rules
- legacy nested `.git` migration

## File Mapping

- `core/backup/provider.go` -> `core/backup/domain/provider.go`
- `core/backup/snapshot.go` -> `core/backup/domain/snapshot.go` plus `core/backup/infra/snapshot/store.go`
- `core/backup/path_filter.go` -> `core/backup/infra/snapshot/path_filter.go`
- `core/backup/git_migration.go` -> `core/backup/infra/snapshot/git_migration.go`
- `core/backup/git_provider.go` -> `core/backup/infra/provider/git_provider.go`
- `core/backup/{aliyun,tencent,huawei,aws,azure,google}.go` -> `core/backup/infra/provider/*`
- `core/backup/provider_catalog.go` -> shell wiring or `backup/infra/provider/catalog.go`
- shell backup methods in [`cmd/skillflow/app.go`](/Users/shinerio/Workspace/code/SkillFlow/cmd/skillflow/app.go) -> `core/backup/app/service.go`

## Shell Boundary

[`cmd/skillflow/app.go`](/Users/shinerio/Workspace/code/SkillFlow/cmd/skillflow/app.go) should keep only shell responsibilities:

- loading config and building a `backup/app.Service`
- publishing Wails-facing events
- caching last backup result for frontend readback
- invoking post-restore compensation after `backup/app` reports a successful restore

This means:

- `BackupNow`, `ListCloudFiles`, `RestoreFromCloud`, `ResolveGitConflict`, `GetGitConflictPending`, backup preview helpers, and startup git pull should delegate into `backup/app`
- `handleRestoredCloudState()` stays outside `backup` this round, because it writes into `skillcatalog`, `agentintegration`, and `skillsource`
- `OpenGitBackupDir()` remains a shell helper, but it should call a `backup/app` helper for backup-root resolution

## Behavior Invariants

This refactor must preserve:

- current provider names and required credential fields
- current excluded-path rules for backup content
- current git backup behavior, including local commit-signing disable and conflict detection
- current startup git pull behavior
- current restore conflict events and pending-flag behavior
- current snapshot diff semantics for non-git providers
- current post-restore compensation behavior

## Testing Strategy

Use the same TDD flow as the previous context extractions:

1. create failing tests under `core/backup/...`
2. port snapshot, path-filter, git-migration, and git-provider coverage into the new package tree
3. add service tests for run/list/restore/resolve-conflict flows
4. switch shell entrypoints and shell tests
5. remove the old flat `core/backup`

Verification for this round:

- `make generate`
- `go test ./core/... -count=1`
- `go test ./cmd/skillflow -count=1`
- `cd cmd/skillflow/frontend && npm run build`

## Risks

### Risk: restore compensation gets tangled with backup execution

Mitigation:

- keep only backup execution semantics in `backup/app`
- leave post-restore cross-context rebuild in shell/orchestration boundary

### Risk: provider registration becomes ambiguous

Mitigation:

- keep registration shell-owned in this round
- make shell wire concrete providers from `backup/infra/provider`

### Risk: git backup behavior drifts during the split

Mitigation:

- port existing git-provider characterization tests first
- preserve the same command sequencing and conflict payload behavior
