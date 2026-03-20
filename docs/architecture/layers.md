# Layers and Dependencies

## Layer Definitions

Under Wails, the transport adapter layer stays in `cmd/skillflow/`, while reusable backend code inside each bounded context is organized as:

```text
cmd/skillflow/
  transport adapters
  shell bootstrap
  tray/window/process integration

core/<context>/
  app/
    command/
    query/
    port/
      repository/
      gateway/
  domain/
  infra/
    repository/
    gateway/
    projection/
```

## Transport Adapters

The old `controller` concept maps most closely to transport adapters.

Current implementation location:

- `cmd/skillflow/` Wails `App` methods

Responsibilities:

- receive Wails calls
- validate request shape and simple invariants
- map request DTOs to application commands, orchestration calls, or read-model queries
- map results back to transport DTOs

Non-responsibilities:

- domain rule evaluation
- direct filesystem, Git, cloud, or SDK access
- hidden cross-context business coordination

Future CLI or API entrypoints should follow the same role at the module edge.

## `app`

The application layer owns use cases.

Responsibilities:

- coordinate aggregates and domain services
- define transaction boundaries
- invoke repositories and gateways through ports
- publish domain events after successful state changes
- expose commands and context-owned queries

Structure:

- `command/` for write use cases
- `query/` for context-owned read use cases
- `port/repository/` for persistence-facing interfaces
- `port/gateway/` for external-system interfaces

## `domain`

The domain layer owns business meaning.

Responsibilities:

- aggregate roots
- entities
- value objects
- domain services
- domain policies
- domain events

Rules:

- no dependency on `infra`, Wails, JSON, SDKs, or filesystem details
- no direct dependency on another context's aggregates

## `infra`

The infrastructure layer implements ports declared by the application layer.

Responsibilities:

- repository implementations
- gateway implementations
- projections or local cache implementations for the owning context
- filesystem, Git, cloud, or runtime adapters

Rules:

- `infra` may depend on `app` ports and `domain`
- one context must not depend on another context's `infra`

## Cross-Context Modules

## `orchestration/`

Use `orchestration/` for explicit write coordination that spans multiple contexts.

Examples:

- import skill from source and then auto-push
- update installed skill and refresh pushed copies
- restore backup and then rebuild derived state

Shell startup sequencing is not part of domain orchestration. It stays in `cmd/skillflow/bootstrap.go` or equivalent shell bootstrap code.

## `readmodel/`

Use `readmodel/` for composed read views spanning multiple contexts.

Examples:

- Dashboard
- Settings page
- My Agents aggregated view
- source candidate list enriched with installed state

Rules:

- `readmodel/` must depend only on explicit published query providers or published language DTOs from contexts
- `readmodel/` must not depend directly on another context's repositories or unpublished query internals
- `readmodel/` may cache projections, but it never owns business truth

## `platform/`

`platform/` contains pure technical capabilities:

- logging
- filesystem helpers
- Git client primitives
- HTTP client primitives
- event bus
- settings store
- upgrade / startup cutover helpers
- path normalization
- update-download primitives

`platform/` must stay business-agnostic.

## `shared/`

`shared/` contains the minimal shared kernel:

- logical keys
- common domain errors
- base event contracts

Do not use `shared/` as a dumping ground for context-local IDs or business behavior.

## Dependency Rules

Allowed direction:

```text
transport adapters -> app
transport adapters -> orchestration
transport adapters -> readmodel
orchestration -> app -> domain
readmodel -> published query providers / published language
infra -> app/port + domain
platform -> no business dependencies
shared -> no context-specific dependencies
```

Forbidden direction:

- `domain -> infra`
- `domain -> transport adapters`
- `context A -> context B infra`
- `transport adapters -> infra`
- `readmodel -> context repositories directly`
- `readmodel -> unpublished app internals`

## Repository vs Gateway

Classification rule:

- if it persists or retrieves this context's own truth, it is a `repository`
- if it talks to a system outside the context boundary, it is a `gateway`

Examples:

- installed skill metadata store: `repository`
- prompt library store: `repository`
- tracked source registry: `repository`
- GitHub Releases API client: `gateway`
- cloud object storage client: `gateway`
- agent workspace adapter: `gateway`
- Wails file dialog bridge: `gateway`

## Extension Rule

When adding a new capability:

1. decide the owning bounded context first
2. add or update an application command or query
3. keep domain rules in `domain`
4. introduce new ports in `app/port`
5. implement the ports in `infra`
6. expose the use case through a transport adapter, orchestration service, or read model

*Last updated: 2026-03-20*
