# Skillcatalog DDD Refactor Design

**Date:** 2026-03-21

## Context

`docs/architecture/README.md`, `contexts.md`, `layers.md`, and `migration.md` define a target backend structure where installed skill truth belongs to the `skillcatalog` bounded context.

The current codebase is still centered on the horizontal package [`core/skill`](/Users/shinerio/Workspace/code/SkillFlow/core/skill), which mixes:

- installed skill domain model
- skill validation and frontmatter parsing
- installed-skill correlation/index logic
- filesystem-backed persistence and metadata rewriting

`cmd/skillflow` and several other backend packages still depend on that package directly. This keeps business truth discoverability low and prevents `cmd/skillflow` from shrinking toward a transport-adapter layer.

## Decision

This refactor extracts the installed skill library into the first end-to-end DDD context:

- `core/skillcatalog/domain`
- `core/skillcatalog/app`
- `core/skillcatalog/infra/repository`

This migration is **single-context direct migration**, not whole-repo big-bang migration:

- `core/skill` will be removed after callers are switched
- this round does not refactor `promptcatalog`, `agentintegration`, `skillsource`, or `backup`
- existing user-facing behavior remains unchanged unless required by the new application boundary

## Scope

### In scope

- move installed-skill model and rules into `skillcatalog/domain`
- move filesystem-backed installed-skill persistence into `skillcatalog/infra/repository`
- introduce `skillcatalog/app` services for installed-skill queries and core mutations
- update `cmd/skillflow` skill-library entrypoints to call the new application layer for skill-library truth
- update old callers in `core/sync`, `core/update`, and shell code to depend on `skillcatalog` instead of `core/skill`
- migrate and keep existing unit coverage green

### Out of scope

- cross-context orchestration extraction
- settings namespace split out of `core/config`
- readmodel extraction from `cmd/skillflow/app_viewstate.go`
- prompt, agent, source, or backup context migration
- user-facing feature changes

## Target Structure

### `core/skillcatalog/domain`

Owns business meaning for installed skills:

- `InstalledSkill`
- `SourceType`
- `SkillMeta`
- validation rules
- logical-key derivation for an installed skill

The domain package remains free of filesystem persistence wiring, Wails, and shell concerns.

### `core/skillcatalog/app`

Owns use-case entrypoints for installed skill truth:

- list installed skills
- get installed skill by id
- list categories
- create / rename / delete category
- import local skill
- delete installed skill
- move installed skill to category
- save/update installed metadata
- overwrite installed content during update flows

The first iteration keeps these use cases pragmatic and small. The goal is not maximal abstraction; the goal is to stop routing skill-library mutations straight through a repository-like type from shell code.

### `core/skillcatalog/infra/repository`

Owns the filesystem implementation:

- skill directory copy/delete/move
- `meta/*.json` and `meta_local/*.local.json`
- portable path persistence and in-place metadata normalization

This layer keeps the current on-disk behavior unchanged.

## File Mapping

- `core/skill/model.go` -> `core/skillcatalog/domain/installed_skill.go`
- `core/skill/validator.go` -> `core/skillcatalog/domain/validator.go`
- `core/skill/meta.go` -> `core/skillcatalog/domain/meta.go`
- `core/skill/index.go` -> `core/skillcatalog/app/query/installed_index.go`
- `core/skill/storage.go` -> `core/skillcatalog/infra/repository/filesystem_storage.go`

Tests move with those responsibilities into the new package tree.

## Caller Migration

The following existing code must stop importing `core/skill` in this round:

- `cmd/skillflow/app.go`
- `cmd/skillflow/app_viewstate.go`
- `cmd/skillflow/skill_state.go`
- `cmd/skillflow/app_restore.go`
- `core/sync/adapter.go`
- `core/sync/filesystem_adapter.go`
- `core/update/checker.go`

Migration rule:

- `cmd/skillflow` uses `skillcatalog/app` for installed-skill CRUD/query entrypoints where possible
- other not-yet-migrated backend packages may depend on stable `skillcatalog/domain` types

## Behavior Invariants

This refactor must preserve these behaviors:

- installed metadata still lives in `meta/*.json`
- local-only metadata still lives in `meta_local/*.local.json`
- synced path persistence still uses forward-slash relative paths where applicable
- import/delete/move/rename behavior remains unchanged
- update detection semantics via `SourceSHA`, `LatestSHA`, and `LastCheckedAt` remain unchanged
- frontmatter parsing and case-insensitive `skill.md` validation remain unchanged

## Testing Strategy

Use TDD for the new context surface:

1. add failing tests under the new `skillcatalog` package tree
2. move or recreate old behavior coverage there
3. implement the new package structure
4. switch callers
5. remove `core/skill`

Verification for this round:

- `go test ./core/...`
- `go test ./cmd/skillflow`

## Risks

### Risk: caller churn breaks shell flows

Mitigation:

- keep the first `skillcatalog/app` API intentionally close to the existing storage capabilities
- migrate tests together with the code
- run targeted `cmd/skillflow` tests in addition to core tests

### Risk: "DDD" becomes a directory rename only

Mitigation:

- introduce an application service boundary in this round
- switch shell CRUD/query entrypoints to that boundary instead of only renaming imports

### Risk: over-design slows down extraction

Mitigation:

- do not force orchestration/readmodel/settings migration into this step
- preserve current on-disk model and runtime behavior
