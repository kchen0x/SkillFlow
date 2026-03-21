# Overview

## Architectural Style

SkillFlow's backend architecture is a DDD-oriented modular monolith hosted inside a Wails desktop application.

The design goals are:

1. Make business truth ownership explicit.
2. Separate domain logic from shell and OS integration concerns.
3. Replace technology-oriented package boundaries with bounded contexts.
4. Keep Wails-specific transport and shell code in `cmd/skillflow/` and reusable business code in `core/`.
5. Keep cross-context writes in `core/orchestration/` and cross-context reads in `core/readmodel/`.

## High-Level Shape

```text
cmd/skillflow/          Wails shell, transport adapters, OS integration, bootstrap
core/
  platform/             pure technical capabilities
  shared/               minimal shared kernel
  orchestration/        cross-context write coordination
  readmodel/            cross-context read composition
  skillcatalog/         core domain
  promptcatalog/        core domain
  agentintegration/     core domain
  skillsource/          supporting domain
  backup/               supporting domain
```

## Source of Truth Rules

- `skillcatalog` owns the truth for installed skills.
- `promptcatalog` owns the truth for prompts.
- `agentintegration` owns the truth for agent profiles, push/pull semantics, and agent-side presence rules.
- `skillsource` owns tracked repositories, logical skill sources, and source-side discovery state.
- `backup` owns backup and restore planning, not the business truth of skills or prompts.
- shell concerns such as tray, window state, single-instance behavior, launch-at-login, and app update belong to `cmd/skillflow/` plus `platform/`, not to a bounded context.
- `Settings` is not a bounded context. It is a UI composition surface over multiple contexts.

In the current product shape, `skillsource` contains two different domain concepts:

- `StarRepo`: the user-tracked GitHub repository
- `SkillSource`: the logical source of one skill, identified by `repo + subpath`

One `StarRepo` may expose many `SkillSource` records.

## Core Architectural Principles

### Truth ownership over page ownership

Pages such as Dashboard, Settings, My Agents, and Starred Repos are not domain boundaries. Backend ownership follows business truth, not frontend navigation.

### Transport adapters stay at the module edge

Because Wails requires bound methods to live in `cmd/skillflow/package main`, transport adapters stay in `cmd/skillflow/`. They translate Wails requests into application use cases and read models.

If CLI or API entrypoints are added later, they should follow the same transport-adapter role at the module edge rather than introducing Wails-specific code into `core/`.

### Vertical contexts over horizontal mega-layers

The preferred organization is by bounded context first, then by layer inside each context. This prevents the codebase from collapsing into giant shared `service`, `repository`, or `domain` packages.

### Orchestration and read model split

- Use `orchestration/` when a write operation spans multiple contexts.
- Use `readmodel/` when a query view combines data from multiple contexts.
- Do not hide cross-context coordination inside one context's domain model.

### Shared kernel must stay small

Only stable concepts shared by multiple contexts belong in `shared/`, such as logical keys, common domain errors, and base event contracts. Context-local instance IDs should stay in their owning bounded contexts unless proven otherwise.

## Core Business Concepts

`Skill` and `Prompt` are parallel first-class business concepts. They should not be collapsed into a generic `Asset` domain model at the domain layer.

If a unified content view is needed, build it in `readmodel/` or application queries rather than by forcing a common domain parent type.

*Last updated: 2026-03-21*
