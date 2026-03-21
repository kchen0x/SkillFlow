# Agent Profile Type Cutover Design

## Goal

Replace the transitional `config.AgentConfig` struct with the context-owned `agentintegration/domain.AgentProfile` type inside `AppConfig`, while preserving the persisted JSON shape and Wails-facing behavior.

## Why This Next

`core/config` still duplicates agent profile shape even after the default-catalog move. Using the context-owned type directly is a low-risk way to keep shrinking the compatibility layer without forcing the full settings-namespace refactor yet.

## Chosen Approach

Keep the field names and JSON schema unchanged and switch only the Go ownership:

- alias `config.AgentConfig` to `agentintegration/domain.AgentProfile`
- update `AppConfig.Agents` to use the alias-backed context type
- remove shell-side conversion helpers that only existed to hop between identical structs
- keep Wails method signatures stable by continuing to return `config.AgentConfig`, which now aliases the context type
- update architecture docs to note that the transitional settings layer now reuses the context-owned agent profile type

## Expected Outcome

After the cutover:

- `AppConfig.Agents` is directly consumable by `agentintegration`
- `core/config` no longer owns a duplicate agent profile struct
- shell code loses a layer of adapter churn around agent settings
