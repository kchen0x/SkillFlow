# SkillFlow Architecture Set

This directory contains the DDD-oriented architecture reference for SkillFlow.

SkillFlow's target backend architecture is a DDD-oriented modular monolith:

- `cmd/skillflow/` stays as the Wails desktop shell, transport adapter layer, process host, and composition root
- backend business code moves toward bounded contexts under `core/`
- each bounded context is organized as `app`, `domain`, and `infra`
- cross-context write coordination goes through `orchestration/`
- cross-context read views go through `readmodel/`

The previous `docs/architecture.md` document mixed current implementation detail, repository rules, runtime notes, and domain concepts into one page. This folder replaces that monolithic format with smaller documents that separate target architecture, runtime constraints, and migration guidance.

## Documents

- [Overview](./overview.md)
  - high-level architectural style, target repository shape, and source-of-truth rules
- [Layers and Dependencies](./layers.md)
  - definitions for transport adapters, `app`, `domain`, `infra`, `orchestration`, `readmodel`, `platform`, and `shared`
- [Bounded Contexts and Domain Model](./contexts.md)
  - bounded context map, aggregate roots, value objects, published language, and cross-context identity rules
- [Application Use Cases](./use-cases.md)
  - command/query ownership by context, shared orchestration, and read-model composition rules
- [Runtime, Repository Layout, and Storage](./runtime-and-storage.md)
  - Wails shell constraints, helper/UI runtime split, target storage layout, and repository vs gateway rules
- [Migration Blueprint](./migration.md)
  - mapping from the current package layout to the target DDD layout and the recommended migration order

## Invariants

- The repository root must contain no Go source files.
- `cmd/skillflow/*.go` must stay flat because Wails bindings require a single `package main` directory.
- SkillFlow remains a Wails desktop app with direct bindings, not a REST service.
- Current transport entrypoints live in `cmd/skillflow/` because of Wails binding constraints.
- `Skill` and `Prompt` are parallel core business concepts.
- `Settings` is a UI composition surface, not a bounded context.

## Scope

These documents cover backend architecture only. User-facing behavior remains documented in [`docs/features.md`](../features.md), and persisted file schemas remain documented in [`docs/config.md`](../config.md).

## Status

The codebase is not yet fully aligned with this architecture. Unless otherwise noted, this document set describes the target structure that future backend refactoring should converge toward.

*Last updated: 2026-03-20*
