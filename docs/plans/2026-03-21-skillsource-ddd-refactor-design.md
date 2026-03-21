# Skillsource DDD Refactor Design

**Date:** 2026-03-21

## Context

`skillcatalog`, `promptcatalog`, and `agentintegration` have already been extracted. The next recommended migration target from [`docs/architecture/migration.md`](/Users/shinerio/Workspace/code/SkillFlow/docs/architecture/migration.md) is `skillsource`.

The current source-management backend is still split across:

- [`core/git`](/Users/shinerio/Workspace/code/SkillFlow/core/git), which currently mixes:
  - starred repository business models
  - tracked-repo storage in `star_repos.json`
  - source-side skill discovery by scanning cached repositories
  - low-level Git and remote URL helpers
- [`cmd/skillflow/app.go`](/Users/shinerio/Workspace/code/SkillFlow/cmd/skillflow/app.go), which currently hosts:
  - add/remove/list tracked repository flows
  - refresh-one / refresh-all flows
  - source-candidate listing with installed/pushed overlays
  - source-to-skillcatalog import orchestration
- [`cmd/skillflow/app_restore.go`](/Users/shinerio/Workspace/code/SkillFlow/cmd/skillflow/app_restore.go), which restores cloned repo caches after cloud restore

This means the actual `skillsource` business truth still lives in shell code and in a technically named package that mixes domain and infrastructure concerns.

## Decision

This migration extracts source-management semantics into:

- `core/skillsource/domain`
- `core/skillsource/app`
- `core/skillsource/infra/repository`
- `core/skillsource/infra/discovery`

At the same time, the pure Git and repository URL helpers move from `core/git` into:

- `core/platform/git`

This round is a **direct context extraction with temporary shell orchestration retained**:

- the old `core/git` package will be removed after callers switch
- `skillsource` will own tracked repositories, repo refresh status, and source-side skill discovery
- shell code will still coordinate cross-context writes that import source candidates into `skillcatalog`
- source settings persistence will continue to flow through `star_repos.json` and `star_repos_local.json` for now

That boundary keeps `skillsource` meaningful without forcing `orchestration/` or a full settings split early.

## Scope

### In scope

- move `StarredRepo` and source candidate models into `skillsource/domain`
- move starred-repo JSON storage into `skillsource/infra/repository`
- move cached-repo skill discovery into `skillsource/infra/discovery`
- move track/untrack/list/refresh flows into `skillsource/app`
- move pure Git, cache-dir, URL normalization, and SHA helpers into `platform/git`
- update shell code to delegate source management into `skillsource/app`
- migrate tests and remove `core/git`

### Out of scope

- creating `core/orchestration/ImportSkillFromSourceOrchestrator`
- moving Git-backed installed-skill update checks out of `cmd/skillflow/app.go`
- introducing non-Git source types beyond tracked repositories
- changing UI behavior for Starred Repos, batch import, or update badges

## Target Structure

### `core/skillsource/domain`

Owns source-side business language:

- `StarRepo`
- `SkillSourceCandidate`
- source sync status semantics

This layer should define the meaning of:

- tracked repository identity
- source-side discovery candidate identity (`repo + subpath`)
- local sync failure state

### `core/skillsource/app`

Owns use cases and coordination over tracked repositories:

- track repository
- track repository with credentials
- untrack repository
- list tracked repositories
- refresh one repository
- refresh all repositories
- list source candidates for one repository
- list all source candidates

This layer should not own installed-skill truth or pushed-state truth. Those overlays remain derived by combining `skillsource` data with `skillcatalog` and `agentintegration`.

### `core/skillsource/infra/repository`

Owns persisted tracked-repo state:

- `star_repos.json`
- `star_repos_local.json`
- builtin repo seeding
- relative-path persistence for local cache paths

### `core/skillsource/infra/discovery`

Owns scanning cached repositories for `SKILL.md` / `skill.md` directories with the current recursive behavior and depth limit.

### `core/platform/git`

Owns pure technical helpers:

- git availability checks
- clone/update
- clone/update with credentials
- remote URL normalization
- canonical repo source derivation
- cache-dir derivation
- repo subpath SHA lookup

## File Mapping

- `core/git/model.go` -> `core/skillsource/domain/source.go`
- `core/git/storage.go` -> `core/skillsource/infra/repository/star_repo_storage.go`
- `core/git/scanner.go` -> `core/skillsource/infra/discovery/repo_scanner.go`
- `core/git/client.go` -> `core/platform/git/client.go`
- `core/git/*_test.go` -> corresponding tests under `core/skillsource/...` and `core/platform/git/...`
- `cmd/skillflow/app.go` starred-repo flows -> `core/skillsource/app/service.go`

## Shell Boundary

[`cmd/skillflow/app.go`](/Users/shinerio/Workspace/code/SkillFlow/cmd/skillflow/app.go) should keep only shell responsibilities:

- loading config
- choosing installed-skill overlays for source candidates
- importing selected source candidates into `skillcatalog`
- backup scheduling
- Wails DTO transport

This means:

- `AddStarredRepo`, `AddStarredRepoWithCredentials`, `RemoveStarredRepo`, `ListStarredRepos`, `UpdateStarredRepo`, and `UpdateAllStarredRepos` should delegate into `skillsource/app`
- `ListAllStarSkills` and `ListRepoStarSkills` should obtain raw candidates from `skillsource/app` and then layer installed/pushed state in shell read composition
- `ImportStarSkills` remains shell-level cross-context orchestration for now, but it should consume `skillsource` terminology instead of `core/git`

[`cmd/skillflow/app_restore.go`](/Users/shinerio/Workspace/code/SkillFlow/cmd/skillflow/app_restore.go) should keep restore compensation orchestration, but it should consume `skillsource` storage types and `platform/git` clone helpers.

## Behavior Invariants

This refactor must preserve:

- current `star_repos.json` and `star_repos_local.json` layout
- builtin starred repo seeding on first init only
- relative local cache path persistence
- current auth-error mapping for HTTP and SSH clone failures
- current refresh-one and refresh-all behavior
- current source candidate scan semantics, including nested plugin skill discovery and max-depth handling
- current source candidate overlays for installed, imported, updatable, pushed, and pushedAgents
- current source import behavior into `skillcatalog`

## Testing Strategy

Use the same TDD flow as the previous context extractions:

1. create failing tests under `core/skillsource/...` and `core/platform/git/...`
2. port current storage, scanner, and Git helper coverage from `core/git`
3. add explicit service tests for track/list/untrack/refresh flows
4. switch shell entrypoints and shell tests
5. remove `core/git`

Verification for this round:

- `make generate`
- `go test ./core/... ./cmd/skillflow -count=1`
- `cd cmd/skillflow/frontend && npm run build`

## Risks

### Risk: shell keeps too much source logic

Mitigation:

- move all tracked-repo mutation flows into `skillsource/app`
- leave only installed/pushed overlay composition and source-import orchestration in shell

### Risk: moving `core/git` breaks unrelated callers

Mitigation:

- split pure Git helpers into `platform/git`
- update `skillkey`, update checks, restore logic, and shell tests in the same round

### Risk: restore compensation drifts from tracked-repo semantics

Mitigation:

- keep restore orchestration in shell
- reuse `skillsource` storage types and `platform/git` clone helpers so the repo-tracking truth stays singular
