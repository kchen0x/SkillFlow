# Remove Card Status Visibility Customization Design

**Date:** 2026-03-22

## Context

SkillFlow currently lets users customize which status badges appear on skill cards per page through a **Card Status Visibility** section in **Settings → General**. That feature affects:

- frontend Settings UI
- frontend status-visibility context and helpers
- shared persisted config schema via `skillStatusVisibility`
- upgrade logic that still migrates legacy visibility keys
- config, feature, and architecture docs

The user explicitly wants this customization removed because it creates too much cognitive overhead. The desired outcome is not "hide the setting while old values still influence behavior"; it is to stop supporting user customization entirely.

The user also clarified that this removal does not need extra new test cases beyond what is already necessary to keep the existing product behavior safe.

## Decision

1. Remove the **Card Status Visibility** section from **Settings → General**.
2. Remove `skillStatusVisibility` from the persisted shared config schema.
3. Stop reading `skillStatusVisibility` from runtime config.
4. Keep status badge behavior fixed to the current default policy:
   - `mySkills`: `updatable`, `pushedAgents`
   - `myAgents`: `imported`, `updatable`, `pushedAgents`
   - `pushToAgent`: `pushedAgents`
   - `pullFromAgent`: `imported`
   - `starredRepos`: `imported`, `pushedAgents`
5. Add startup cutover logic that deletes any legacy `skillStatusVisibility` field from `config.json`.
6. Remove config and architecture documentation that presents status-visibility customization as a supported user-facing or persisted feature.

## Why This Approach

### Fixed defaults are simpler than hidden state

If we remove the Settings UI but keep the stored value alive, users who changed the setting previously would still see non-default behavior without any visible explanation. That is worse than the current design because the behavior becomes invisible and harder to reason about.

### No runtime compatibility branch

The repo rules already prefer explicit startup cutovers over long-lived business compatibility branches for repo-tracked on-disk schema changes. Removing `skillStatusVisibility` follows that same rule:

- rewrite persisted shared config to the latest schema
- keep runtime code on one current model only

### Preserve actual page behavior

This change removes customization, not statuses themselves. Pages should keep showing the same default statuses they already show today when no customization is applied. That preserves core UX while removing complexity.

## Runtime Behavior After Removal

### Frontend

- All card-status consumers use fixed default visibility.
- No page can toggle per-status visibility from Settings.
- The status-visibility context can remain as a thin fixed-default provider if that is the least invasive path for existing consumers.

### Config

- `config.AppConfig` no longer includes `skillStatusVisibility`.
- `config.json` no longer stores `skillStatusVisibility`.
- `config_local.json` remains unchanged by this feature.

### Upgrade

- On startup, if `config.json` contains `skillStatusVisibility`, the field is removed.
- Existing terminology migration for `tools -> agents` remains intact.
- Legacy visibility-key migration becomes unnecessary once the whole field is removed.

## UX Impact

Users will no longer see any setting for status-badge customization.

Status badges remain visible using the built-in defaults. There is no user action required after upgrade, and no new setting is introduced.

## Scope

- `cmd/skillflow/frontend/src/pages/Settings.tsx`
- `cmd/skillflow/frontend/src/contexts/SkillStatusVisibilityContext.tsx`
- `cmd/skillflow/frontend/src/lib/skillStatusVisibility.ts`
- `cmd/skillflow/frontend/src/i18n/en.ts`
- `cmd/skillflow/frontend/src/i18n/zh.ts`
- existing frontend tests that mention card-status visibility
- `core/config/model.go`
- `core/config/service.go`
- existing config tests in `core/config/service_test.go`
- `core/platform/upgrade/upgrade.go`
- existing upgrade tests in `core/platform/upgrade/config_terms_test.go`
- `docs/features.md`
- `docs/features_zh.md`
- `docs/config.md`
- `docs/config_zh.md`
- relevant architecture docs that still mention persisted card-status visibility

## Out Of Scope

- changing the default status policy itself
- removing status badges from skill cards
- redesigning any status-strip visuals
- changing unrelated settings or config fields
