# SkillFlow Architecture Set

This directory contains the current backend architecture reference for SkillFlow.

SkillFlow's backend is a DDD-oriented modular monolith:

- `cmd/skillflow/` is the Wails desktop shell, transport adapter layer, process host, and composition root
- backend business code lives under bounded contexts in `core/`
- each bounded context is organized as `app`, `domain`, and `infra`
- cross-context write coordination goes through `core/orchestration/`
- cross-context read composition goes through `core/readmodel/`
- `core/config/` is a frontend-facing settings facade over context- and platform-owned settings
- pure technical capabilities live in `core/platform/`
- only highly stable shared kernel concepts live in `core/shared/`

## Documents

- [Overview](./overview.md)
  - high-level architectural style, repository shape, and source-of-truth rules
- [Layers and Dependencies](./layers.md)
  - definitions for transport adapters, `app`, `domain`, `infra`, `orchestration`, `readmodel`, `platform`, and `shared`
- [Bounded Contexts and Domain Model](./contexts.md)
  - bounded context map, aggregate roots, value objects, published language, and cross-context identity rules
- [Application Use Cases](./use-cases.md)
  - command/query ownership by context, shared orchestration, and read-model composition rules
- [Runtime, Repository Layout, and Storage](./runtime-and-storage.md)
  - Wails shell constraints, helper/UI runtime split, storage layout, and repository vs gateway rules

## Invariants

- The repository root must contain no Go source files.
- `cmd/skillflow/*.go` must stay flat because Wails bindings require a single `package main` directory.
- SkillFlow remains a Wails desktop app with direct bindings, not a REST service.
- Current transport entrypoints live in `cmd/skillflow/` because of Wails binding constraints.
- `Skill` and `Prompt` are parallel core business concepts.
- `Settings` is a UI composition surface, not a bounded context.
- `core/config/` is a settings facade, not a source-of-truth bounded context.

## Scope

These documents cover backend architecture only. User-facing behavior remains documented in [`docs/features.md`](../features.md), and persisted file schemas remain documented in [`docs/config.md`](../config.md).

*Last updated: 2026-03-21*
