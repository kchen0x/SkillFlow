# Agent Profile Type Cutover Plan

1. Add a failing test proving `config.AppConfig.Agents` can be consumed directly by `agentintegration`.
2. Alias `config.AgentConfig` to `agentintegration/domain.AgentProfile`.
3. Remove redundant shell conversion helpers and switch call sites to use `cfg.Agents` directly.
4. Update architecture docs to reflect that transitional settings reuse the context-owned agent profile type.
5. Run `gofmt`, targeted tests, then `go test ./core/... -count=1` and `go test ./cmd/skillflow -count=1`.
