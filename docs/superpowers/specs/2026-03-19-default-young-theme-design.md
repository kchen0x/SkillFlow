# Default Young Theme Design

**Date:** 2026-03-19
**Status:** Approved

## Goal

Change SkillFlow's first-open theme default from `dark` to `young` without changing any persisted user preference behavior.

## Scope

- New installs or browsers with no saved theme should start in `young`.
- Existing saved `skillflow-theme-v2` values must still win.
- Legacy `skillflow-theme` migration must remain:
  - `light` -> `young`
  - `dark` -> `dark`

## Non-goals

- Do not change the theme cycle order.
- Do not rename theme ids or labels.
- Do not change settings UI copy beyond documentation that describes the default.

## Design

The initial theme resolution already lives in `cmd/skillflow/frontend/src/hooks/useTheme.ts`. The smallest safe change is to make the "no stored value" fallback return `young` instead of `dark`.

To keep this behavior testable without React rendering, extract the storage resolution logic into a small exported helper that accepts a storage-like object. `getInitialTheme()` continues to call that helper with `localStorage`, while unit tests exercise the pure helper directly.

## Affected Files

- `cmd/skillflow/frontend/src/hooks/useTheme.ts`
- `cmd/skillflow/frontend/package.json`
- `cmd/skillflow/frontend/tests/useTheme.test.mjs`
- `docs/features.md`
- `docs/features_zh.md`
- `README.md`
- `README_zh.md`
