# Upgrade Platform Cutover Design

## Goal

Move startup cutover code from `core/upgrade` to `core/platform/upgrade` so the package location matches the architecture rule that persisted-schema upgrade code belongs to the platform layer.

## Why This Next

`upgrade` is a small, isolated technical package with a single shell entrypoint. It is a clean platform migration and removes another remaining top-level cross-cutting package before tackling the much larger `viewstate` and `config` work.

## Chosen Approach

Keep the API unchanged and only relocate the package:

- create `core/platform/upgrade`
- move the startup cutover implementation and tests there
- switch shell imports and startup tests
- remove `core/upgrade`

This preserves startup behavior while aligning the codebase to the documented architecture.

## Expected Outcome

After the cutover:

- startup uses `core/platform/upgrade`
- `core/upgrade` is removed
- architecture docs reflect the completed platform move
