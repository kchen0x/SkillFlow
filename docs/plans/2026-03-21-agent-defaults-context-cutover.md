# Agent Defaults Context Cutover Plan

1. Add failing tests for built-in agent names and default scan/push directory helpers under `core/agentintegration/domain`.
2. Implement the default agent catalog in `core/agentintegration/domain`.
3. Refactor `core/config` and shell adapter wiring to consume the new context-owned defaults.
4. Update architecture docs to reflect that built-in agent profile defaults are now owned by `agentintegration`.
5. Run `gofmt`, targeted tests, then `go test ./core/... -count=1` and `go test ./cmd/skillflow -count=1`.
