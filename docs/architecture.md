# SkillFlow Architecture

> 🌐 [中文](architecture_zh.md) | **English**

This is the entry page for SkillFlow's architecture documentation. The old single-file architecture reference has been replaced by a DDD-oriented document set under [`docs/architecture/`](./architecture/README.md).

SkillFlow's target backend architecture is a DDD-oriented modular monolith:

- `cmd/skillflow/` stays as the Wails desktop shell, process host, and composition root
- backend business code moves toward bounded contexts under `core/`
- each bounded context is organized as `entrypoint`, `app`, `domain`, and `infra`
- cross-context writes go through `workflow/`
- cross-context read views go through `readmodel/`

Use this document set as the source of truth for future backend refactoring.

## Reading Guide

- [Architecture Set Overview](./architecture/README.md)
- [Overview](./architecture/overview.md)
- [Layers and Dependencies](./architecture/layers.md)
- [Bounded Contexts and Domain Model](./architecture/contexts.md)
- [Application Use Cases](./architecture/use-cases.md)
- [Runtime, Repository Layout, and Storage](./architecture/runtime-and-storage.md)
- [Migration Blueprint](./architecture/migration.md)

## Invariants

- The repository root must contain no Go source files.
- `cmd/skillflow/*.go` must stay flat because Wails bindings require a single `package main` directory.
- SkillFlow remains a Wails desktop app with direct bindings, not a REST service.
- `Skill` and `Prompt` are parallel core business concepts.
- `Settings` is a UI composition surface, not a bounded context.

*Last updated: 2026-03-20*
