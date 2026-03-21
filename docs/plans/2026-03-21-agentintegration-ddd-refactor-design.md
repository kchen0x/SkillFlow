# Agentintegration DDD Refactor Design

**Date:** 2026-03-21

## Context

`skillcatalog` and `promptcatalog` have already been extracted. The next recommended migration target from [`docs/architecture/migration.md`](/Users/shinerio/Workspace/code/SkillFlow/docs/architecture/migration.md) is `agentintegration`.

The current agent backend is split across two places:

- [`core/sync`](/Users/shinerio/Workspace/code/SkillFlow/core/sync), which currently owns only the agent filesystem adapter contract and implementation
- [`cmd/skillflow/app.go`](/Users/shinerio/Workspace/code/SkillFlow/cmd/skillflow/app.go) plus [`cmd/skillflow/skill_state.go`](/Users/shinerio/Workspace/code/SkillFlow/cmd/skillflow/skill_state.go), which currently host:
  - enabled-agent filtering
  - scan and push directory checks
  - push conflict detection
  - push and force-push flows
  - scan/pull candidate resolution
  - pushed/seen presence aggregation
  - refresh of pushed copies after skill updates

This means the actual `agentintegration` business semantics still live in shell code, while `core/sync` is only a thin technical package.

## Decision

This migration extracts agent-side semantics into:

- `core/agentintegration/domain`
- `core/agentintegration/app`
- `core/agentintegration/infra/gateway`

This round is a **direct context extraction with temporary shell orchestration retained**:

- the old `core/sync` package will be removed after callers switch
- `agentintegration` will own scan/push/presence semantics and gateway selection
- shell code will still coordinate cross-context writes that import agent skills into `skillcatalog`
- agent config persistence will continue to flow through `core/config` for now

That boundary keeps the context meaningful without forcing the settings-namespace split early.

## Scope

### In scope

- move agent gateway contracts and the filesystem adapter into `agentintegration`
- move agent profile runtime types, push conflicts, scan candidates, and agent skill entry read types into `agentintegration/domain`
- move scan, list, push, force-push, missing-dir checks, enabled-agent filtering, and pushed-copy refresh planning into `agentintegration/app`
- move pushed/seen presence resolution and agent skill candidate aggregation out of `cmd/skillflow/skill_state.go`
- update shell code to delegate agent operations into `agentintegration/app`
- migrate tests and remove `core/sync`

### Out of scope

- moving persisted agent settings out of `core/config`
- refactoring `SaveConfig`, `AddCustomAgent`, and `RemoveCustomAgent` into a context-owned config namespace
- creating `core/orchestration` for agent-to-skillcatalog pull/import flows
- UI behavior changes for My Agents / Push / Pull pages

## Target Structure

### `core/agentintegration/domain`

Owns agent-side business language:

- `AgentProfile`
- `PushConflict`
- `AgentSkillCandidate`
- `AgentSkillEntry`
- agent-side presence and grouping rules

This layer should define the meaning of:

- `pushed`
- `seenInAgentScan`
- conflict records
- fallback identity/grouping for agent-side skill observations

### `core/agentintegration/app`

Owns use cases and coordination over agent gateways:

- list enabled agent profiles
- check missing push directories
- push installed skills
- push external/starred skill directories
- force push
- scan agent skills
- list agent skills with pushed and seen overlays
- refresh already-pushed copies after installed skill changes

This layer should consume installed-skill summaries from `skillcatalog`, not own installed skill truth itself.

### `core/agentintegration/infra/gateway`

Owns the file-based gateway implementation currently in `core/sync`:

- filesystem push
- filesystem pull / recursive scan
- max-depth pull support

This remains the place for agent-specific filesystem transport details.

## File Mapping

- `core/sync/adapter.go` -> `core/agentintegration/app/port/gateway/adapter.go`
- `core/sync/filesystem_adapter.go` -> `core/agentintegration/infra/gateway/filesystem_adapter.go`
- `core/sync/filesystem_adapter_test.go` -> `core/agentintegration/infra/gateway/filesystem_adapter_test.go`
- `cmd/skillflow/skill_state.go` agent candidate/presence logic -> `core/agentintegration/domain/*` plus `core/agentintegration/app/*`
- `cmd/skillflow/app.go` agent push/scan/list logic -> `core/agentintegration/app/*`

## Shell Boundary

[`cmd/skillflow/app.go`](/Users/shinerio/Workspace/code/SkillFlow/cmd/skillflow/app.go) should keep only shell responsibilities:

- loading and saving config
- choosing which installed skills from `skillcatalog` to pass into `agentintegration`
- importing selected agent-side skills back into `skillcatalog`
- backup scheduling
- Wails DTO transport

This means:

- `PushToAgents`, `PushToAgentsForce`, `PushStarSkillsToAgents`, `ScanAgentSkills`, `ListAgentSkills`, `CheckMissingAgentPushDirs`, and pushed-copy refresh should delegate into `agentintegration/app`
- `PullFromAgent` and `PullFromAgentForce` will remain shell-level cross-context orchestration for now, but they should delegate candidate discovery and selection to `agentintegration`

## Behavior Invariants

This refactor must preserve:

- current filesystem-based push and pull behavior
- flattened push semantics into agent push directories
- recursive scan semantics with the same max-depth handling
- existing push conflict detection based on target directory existence
- current `pushed` and `seenInAgentScan` meaning
- existing auto-push and refresh-on-update behavior
- current Wails-facing My Agents / Push / Pull page behavior

## Testing Strategy

Use the same TDD flow as the previous two context extractions:

1. create failing tests under `core/agentintegration/...`
2. port current filesystem adapter coverage from `core/sync`
3. add explicit service tests for push conflicts, scan aggregation, and presence behavior
4. switch shell entrypoints and shell tests
5. remove `core/sync`

Verification for this round:

- `make generate`
- `go test ./core/... ./cmd/skillflow -count=1`
- `cd cmd/skillflow/frontend && npm run build`

## Risks

### Risk: shell keeps most agent logic

Mitigation:

- move candidate resolution and presence aggregation out of `skill_state.go`
- keep shell limited to config loading, skill selection, and cross-context orchestration

### Risk: cross-context pull/import logic gets tangled during extraction

Mitigation:

- keep actual import into `skillcatalog` in shell for now
- let `agentintegration/app` own scan, grouping, and candidate-selection semantics only

### Risk: config ownership becomes muddled

Mitigation:

- keep `core/config` as temporary persistence for agent profiles this round
- introduce `domain.AgentProfile` as the runtime business type, while deferring full config namespace migration
