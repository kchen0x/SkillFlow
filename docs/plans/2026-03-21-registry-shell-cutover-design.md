# Registry Shell Cutover Design

## Goal

Remove `core/registry` and keep adapter/provider registration inside the Wails shell, matching the architecture rule that composition concerns belong in `cmd/skillflow`.

## Why This Next

All bounded contexts are now extracted. `core/registry` no longer owns business semantics; it only stores shell wiring state for agent gateways and backup providers. Keeping it in `core/` preserves an unnecessary cross-cutting package after the actual contexts have moved out.

## Chosen Approach

Move the lightweight registries into `cmd/skillflow`:

- agent gateway registration stays in [`cmd/skillflow/adapters.go`](/Users/shinerio/Workspace/code/SkillFlow/cmd/skillflow/adapters.go)
- cloud provider factory registration stays in [`cmd/skillflow/providers.go`](/Users/shinerio/Workspace/code/SkillFlow/cmd/skillflow/providers.go)
- transport adapters consume shell-local lookup helpers instead of importing `core/registry`

This keeps the change narrow:

- no bounded context API changes
- no user-facing behavior changes
- no new platform package introduced prematurely

## Boundaries After Cutover

- `core/agentintegration` still depends on an injected gateway resolver, but the resolver is now owned entirely by shell wiring.
- `core/backup/app` still depends on an injected cloud-provider resolver, but the resolver now comes from shell-owned provider factories.
- `cmd/skillflow` remains the composition root and is the only place that should know concrete adapter/provider registration.

## Verification

- `go test ./core/... -count=1`
- `go test ./cmd/skillflow -count=1`
- `rg -n 'core/registry' core cmd/skillflow -g'*.go'`

## Expected Outcome

After this cutover:

- `core/registry` is removed
- shell dependency wiring becomes more explicit
- migration docs accurately reflect that composition concerns are now back in `cmd/skillflow`
