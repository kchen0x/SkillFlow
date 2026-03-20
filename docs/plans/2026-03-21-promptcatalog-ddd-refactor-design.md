# Promptcatalog DDD Refactor Design

**Date:** 2026-03-21

## Context

`skillcatalog` has already been extracted as the first end-to-end bounded context. The next migration target from `docs/architecture/migration.md` is `promptcatalog`.

The current prompt backend still centers on [`core/prompt`](/Users/shinerio/Workspace/code/SkillFlow/core/prompt), where one package currently owns:

- prompt domain model and validation
- category rules
- import/export flow objects
- prompt storage and legacy-layout migration
- web-link parsing and URL normalization

`cmd/skillflow/app_prompt.go` calls that package directly, so prompt business truth is still hosted in a horizontal package and shell methods are not yet reduced to thin transport adapters.

## Decision

This migration extracts the prompt library into:

- `core/promptcatalog/domain`
- `core/promptcatalog/app`
- `core/promptcatalog/infra/repository`

This is another **single-context direct migration**:

- the old `core/prompt` package will be removed after callers switch
- prompt shell session state in `cmd/skillflow/app_prompt_session.go` remains in the shell
- user-facing prompt behavior stays unchanged

## Scope

### In scope

- move prompt truth and validation rules into `promptcatalog/domain`
- move prompt CRUD, category management, import/export flow logic into `promptcatalog/app`
- move filesystem-backed prompt persistence and legacy-layout migration into `promptcatalog/infra/repository`
- update `cmd/skillflow/app_prompt.go` to become a shell transport adapter over `promptcatalog/app`
- update prompt import session structs in shell code to depend on `promptcatalog/app` types instead of `core/prompt`
- migrate tests and remove `core/prompt`

### Out of scope

- prompt UI redesign or behavior changes
- readmodel extraction for prompt pages
- moving prompt import session state out of `cmd/skillflow`
- settings namespace split
- any `agentintegration`, `skillsource`, or `backup` refactor

## Target Structure

### `core/promptcatalog/domain`

Owns prompt business meaning and rules:

- `Prompt`
- `PromptLink`
- prompt/category name normalization
- prompt image and web-link validation
- path-segment validation
- prompt URL normalization

It must stay free of Wails and repository-specific details.

### `core/promptcatalog/app`

Owns use cases and flow objects:

- list/get prompts
- list/create/rename/delete categories
- create/update/delete prompts
- move prompt category
- preview prompt import conflicts
- apply prompt import
- export prompt bundle
- parse markdown web links for shell-facing API calls

This layer also owns import/export DTO-like flow types such as:

- `ImportPrompt`
- `ImportPreview`
- export bundle representations

These are application-flow objects, not aggregate roots.

### `core/promptcatalog/infra/repository`

Owns the on-disk model and storage migration:

- `prompts/<category>/<name>/`
- `system.md`
- `prompt.json`
- legacy prompt-layout migration
- prompt file read/write

This keeps current storage behavior stable while moving repository code out of the old horizontal package.

## File Mapping

- `core/prompt/storage.go` domain parts -> `core/promptcatalog/domain/*`
- `core/prompt/storage.go` app/use-case parts -> `core/promptcatalog/app/*`
- `core/prompt/storage.go` filesystem and migration parts -> `core/promptcatalog/infra/repository/*`
- `core/prompt/storage_test.go` -> split across new `domain`, `app`, and `infra/repository` tests

## Shell Boundary

[`cmd/skillflow/app_prompt.go`](/Users/shinerio/Workspace/code/SkillFlow/cmd/skillflow/app_prompt.go) should keep only shell responsibilities:

- Wails file dialogs
- session id orchestration
- log messages
- backup scheduling
- DTO return values

It should no longer host prompt-library truth directly.

[`cmd/skillflow/app_prompt_session.go`](/Users/shinerio/Workspace/code/SkillFlow/cmd/skillflow/app_prompt_session.go) remains in `cmd/skillflow` because it stores UI-scoped import-session state, not prompt-domain truth.

## Behavior Invariants

This refactor must preserve:

- prompt storage under `prompts/<category>/<name>/`
- `system.md` and `prompt.json` file names
- `Default` category semantics
- name-based prompt import conflict detection
- current import/export JSON shape
- legacy prompt-layout migration behavior
- prompt web-link markdown parsing
- image URL limits and URL validation behavior

## Testing Strategy

Use TDD in the same pattern as `skillcatalog`:

1. add failing tests under `core/promptcatalog/...`
2. move or recreate old behavior coverage there
3. implement the new package tree
4. switch shell callers and prompt session types
5. remove `core/prompt`

Verification for this round:

- `make generate`
- `go test ./core/... ./cmd/skillflow -count=1`
- `cd cmd/skillflow/frontend && npm run build`

## Risks

### Risk: import/export behavior drifts while refactoring

Mitigation:

- keep import/export JSON coverage explicit in new app tests
- preserve current overwrite-on-name semantics

### Risk: shell prompt flows keep business logic

Mitigation:

- keep file dialogs and import session state in shell
- move prompt CRUD/import/export rules behind `promptcatalog/app`

### Risk: repository migration logic gets lost

Mitigation:

- port legacy-layout migration tests before deleting `core/prompt`
- keep storage migration in `infra/repository`, not in shell code
