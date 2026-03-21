# Agent Defaults Context Cutover Design

## Goal

Move the built-in agent catalog and default scan/push directory rules out of `core/config` and into `core/agentintegration/domain` so those defaults live with the context that owns agent profiles.

## Why This Next

After introducing `core/platform/settingsstore`, `core/config` is still carrying non-storage behavior. The built-in agent list and default profile paths are agent-profile semantics, not generic config mechanics, and they are reused by shell wiring outside the config package.

## Chosen Approach

Keep the external config shape unchanged and relocate only the agent-default rules:

- add built-in agent default helpers under `core/agentintegration/domain`
- cover them with domain tests
- refactor `core/config` to build default agent config from the agentintegration defaults
- refactor shell adapter registration to consume the same agentintegration defaults directly
- update architecture docs to note that built-in agent profile defaults now belong to `agentintegration`

## Expected Outcome

After the cutover:

- built-in agent names and default directories live in `agentintegration`
- `core/config` loses another slice of domain-adjacent behavior
- shell wiring and config defaults reuse the same context-owned source of truth
