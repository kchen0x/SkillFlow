# Performance Rearchitecture Design

**Date:** 2026-03-11
**Status:** Approved

## Goal

Reduce startup jank, route-transition lag, and card-animation stutter without changing existing product behavior or weakening cloud-sync correctness.

## Non-goals

- Do not change synced file formats or semantics in `config.json`, `meta/*.json`, or `star_repos.json`.
- Do not make local caches a source of truth.
- Do not require user interaction to reach the latest on-screen state after cloud restore, starred-repo refresh, or local mutations.

## Root Causes

### Backend

1. `ListSkills()` currently performs expensive derived-state work on the critical path.
   - `cmd/skillflow/app.go` loads all installed skills and immediately rebuilds tool push presence.
   - `cmd/skillflow/skill_state.go` rebuilds presence by scanning every configured tool `PushDir`.

2. Startup background work competes with first paint.
   - After `domReady`, the app schedules update checks, starred-repo refresh, app-update checks, and startup git pull close together.

3. Repeated page loads redo the same filesystem-heavy derivations.
   - Installed-skill summaries, pushed-tool mappings, and starred-repo derived views are recomputed instead of reused.

### Frontend

1. Page transitions animate entire routed trees.
   - `cmd/skillflow/frontend/src/App.tsx` wraps routes in `AnimatePresence` and route-level `motion.div`.

2. Card rendering does per-card measurement work.
   - `cmd/skillflow/frontend/src/components/SkillStatusStrip.tsx` creates a `ResizeObserver` per card and repeatedly measures child widths.

3. Pages rely on full reloads instead of targeted state updates.
   - `Dashboard`, `ToolSkills`, and `StarredRepos` each pull broad datasets and replace large slices of page state.

## Design Principles

1. Truth stays on disk in existing synced files.
2. Performance caches are local-only derived state under `cache/`.
3. Stale cached UI is acceptable briefly, but long-term stale UI is not.
4. Event-driven reconciliation is primary; low-frequency polling is only a safety net.
5. Any cache failure must degrade to recomputation, never to incorrect persisted data.

## Consistency Model

### Truth Layer

The authoritative state remains:

- `skills/`
- `meta/*.json`
- `config.json`
- `config_local.json`
- `star_repos.json`

Cloud sync, restore, starred-repo refresh, skill import/update/delete, and config writes continue to modify only these existing files.

### Derived Local Cache Layer

Add local-only derived cache files under `cache/`, for example:

- `cache/viewstate/installed_skills.json`
- `cache/viewstate/tool_presence.json`
- `cache/viewstate/starred_summary.json`

These files:

- contain only recomputable data
- are excluded from cloud backup and git sync
- never write back into synced files
- may differ across devices without affecting correctness

### Final Consistency Guarantee

The UI may render cached derived data first, then reconcile in the background, but it must converge to truth even if the user leaves the page untouched for a long time.

This guarantee is provided by:

1. targeted invalidation when known mutations complete
2. reconciliation after restore and starred-repo refresh events
3. low-frequency fingerprint checks while the app remains open
4. frontend subscriptions to backend "view state changed" events

## Backend Architecture

### New Package: `core/viewstate`

Create a reusable package responsible for:

- building derived view state from authoritative disk state
- persisting local-only cache snapshots
- computing dependency fingerprints
- deciding whether cached data is reusable or stale

Suggested files:

- `core/viewstate/model.go`
- `core/viewstate/cache.go`
- `core/viewstate/fingerprint.go`
- `core/viewstate/manager.go`

### Managed Derived Slices

The manager owns these derived slices:

1. Installed skills summary
   - `InstalledSkillEntry[]`
   - includes updatable state and pushed-tool list

2. Tool presence index
   - `logicalKey -> []toolName`
   - versioned by tool config and per-tool directory fingerprint

3. Starred repo summary
   - fast page snapshot for repo list and derived per-repo counts

### Fingerprints and Invalidations

Each derived slice stores:

- schema version
- source dependency fingerprint
- build timestamp

Dependency fingerprints include only inputs needed for that slice, such as:

- relevant config hash
- `meta/` directory mtime and file count
- `star_repos.json` mtime
- per-tool `PushDir` path and fingerprint
- app version or cache schema version

Any mismatch invalidates that slice and triggers background rebuild.

### Reconciliation Scheduler

Add an app-level reconciliation coordinator in `cmd/skillflow/` to:

- schedule background derived-state rebuilds
- coalesce repeated invalidation requests
- serialize overlapping rebuilds per slice
- publish frontend update events after successful rebuild

Priority order:

1. installed skills summary
2. tool presence index
3. starred summary
4. remote update checks

### Startup Flow

Replace "compute everything on first page load" with:

1. app startup loads config, storage, logger, and local cache manager
2. first page request reads cached installed-skill snapshot if valid
3. `domReady` schedules background reconciliation
4. route pages subscribe to updates instead of repeatedly forcing full recomputation

### Long-Lived Final Consistency

Add a low-frequency reconciler loop, roughly every 30-60 seconds, that checks lightweight fingerprints only.

If fingerprints changed, schedule precise slice rebuilds.

This loop exists to catch:

- cloud restore results
- external filesystem changes
- long-lived pages that receive no direct user interaction
- changes introduced by background git or cloud tasks

### Restore and Sync Hooks

After these flows complete, the app must explicitly invalidate and reconcile affected slices:

- startup git pull / restore
- manual cloud restore
- starred repo refresh
- skill import
- skill delete
- skill update
- category move / rename where installed list changes
- config changes affecting tools, push dirs, scan dirs, or scan depth

## Event Model

Add explicit backend-to-frontend events for derived state:

- `view.skills.changed`
- `view.toolPresence.changed`
- `view.starred.changed`
- `view.config.changed`

Payloads should include enough data for targeted refresh, but full-slice replacement is acceptable in phase one.

Important distinction:

- operational events report that work started or failed
- view events report that observable state has been rebuilt and is ready for UI consumption

## Frontend Architecture

### Snapshot + Subscription Model

Each major page switches from "load everything on mount, then reload broadly" to:

1. read current snapshot
2. subscribe to relevant `view.*.changed` events
3. replace only the affected page slice

Primary pages:

- `Dashboard`
- `ToolSkills`
- `StarredRepos`

### Remove Per-Card Measurement

`SkillStatusStrip` should stop using per-card `ResizeObserver` logic.

Replace with deterministic CSS rules:

- single-line truncation
- capped pushed-tool icon count
- overflow badge such as `+N`
- no per-card runtime measurement

### Adaptive Motion Degradation

Keep motion, but make it self-degrade:

- route transitions become lightweight fades only
- large lists disable staggered child animation
- card entry animation turns off above configured thresholds
- hover and drag affordances remain, but avoid layout-affecting transforms on dense lists

### Tooltip and Hover Requests

Metadata fetches triggered by hover should:

- debounce
- be cancellable by sequence guard
- cache successful responses in memory
- avoid updating UI with stale responses after pointer leaves

### Optional Second-Phase Virtualization

If post-refactor profiling still shows list rendering pressure on large datasets, add virtualization for dense card grids in phase two. This is a fallback, not a phase-one dependency.

## Cloud Sync Safety

The rearchitecture must not allow device-local caches to overwrite remote or synced truth.

Rules:

1. cache files must never be merged back into synced metadata
2. restore and sync flows must invalidate caches before publishing refreshed UI state
3. cache rebuilds always read from truth layer, never from prior cache as authority
4. cross-device differences in cache contents are expected and harmless

## Rollout Plan

### Phase 1

- add profiling logs / timings around current hotspots
- add `core/viewstate` skeleton and cache format
- keep existing UI behavior while enabling local snapshots

### Phase 2

- move installed-skill summary and tool presence off the `ListSkills()` critical path
- add reconcile scheduler and view update events
- hook restore / starred refresh invalidation paths

### Phase 3

- convert frontend pages to snapshot + subscription updates
- remove `SkillStatusStrip` measurement logic
- degrade route and card motion based on list size

### Phase 4

- add low-frequency fingerprint polling
- add regression tests and performance checks
- update architecture, feature, and config documentation

## Testing Strategy

### Backend

- cache hit returns existing snapshot
- fingerprint mismatch triggers rebuild
- restore invalidates and rebuilds installed and starred slices
- tool presence only rescans changed tools
- cache corruption falls back to rebuild

### Frontend

- first render shows cached snapshot without blank state flicker
- background reconcile updates visible data without user action
- long-lived pages receive backend-published updates
- motion degrades above thresholds
- hover metadata requests cannot race stale payloads into view

### Performance Verification

Track and compare at least:

- startup to first interactive paint
- `ListSkills()` latency before and after refactor
- dashboard first visible list time
- route transition time
- CPU and memory behavior for 100 / 300 / 1000 card scenarios

## Risks and Mitigations

### Risk: stale UI persists too long

Mitigation:

- event-driven invalidation
- periodic fingerprint checks
- explicit post-restore rebuild

### Risk: cache logic becomes a hidden second source of truth

Mitigation:

- cache stores derived data only
- cache never writes into synced files
- rebuild always starts from authoritative disk state

### Risk: refactor shifts work from startup to frequent background spikes

Mitigation:

- prioritize slices
- coalesce rebuild requests
- cap concurrent reconciliation work
- log timings per slice

## Expected Outcome

After rollout:

- app startup should reach usable UI substantially sooner
- route switches should stop animating large trees expensively
- dense card grids should scroll and animate smoothly
- backend CPU and I/O should drop because directory scans are reused and incremental
- cloud-synced changes should appear automatically without requiring user interaction
