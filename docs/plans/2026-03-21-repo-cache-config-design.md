# Repo Cache Path Cutover Design

**Date:** 2026-03-21

## Context

SkillFlow currently allows `skillsStorageDir` to move the installed `skills/` tree outside `<AppDataDir>`. That setting also changes the effective backup root because backup logic switches from `<AppDataDir>` to the shared parent of `skills/`, `meta/`, `meta_local/`, and `prompts/`.

The requested behavior is different:

- remove the ability to configure an external skills directory
- keep the installed skill library under `<AppDataDir>` only
- keep backup and restore rooted at `<AppDataDir>`
- make the large repository clone cache configurable instead
- expose a one-click action in Settings to open `<AppDataDir>`
- keep `README.md` and `README_zh.md` unchanged

The user also clarified the reason for the change: installed skills are not expected to be large, but the full cloned repositories under cache can be large enough to require placement on another disk.

## Decision

1. Remove `skillsStorageDir` from runtime configuration and persisted config files.
2. Fix all skill-owned persisted content under `<AppDataDir>`:
   - `skills/`
   - `meta/`
   - `meta_local/`
   - `prompts/`
3. Fix Git/cloud backup root to `<AppDataDir>` for all cases. Backup root no longer depends on a configurable skill location.
4. Add a new local-only config field `repoCacheDir` in `config_local.json`.
5. Default `repoCacheDir` to `<AppDataDir>/cache/repos`.
6. Limit `repoCacheDir` to repository clone caches only. Do not move:
   - `cache/viewstate/`
   - backup snapshots
   - `runtime/`
   - logs
7. Remove `localDir` from synced `star_repos.json`. Repository local paths become runtime-derived from `repoCacheDir` plus normalized repo source.
8. Add startup cutover code under `core/platform/upgrade/` that:
   - removes persisted `skillsStorageDir`
   - backfills `repoCacheDir` when missing
   - strips legacy `localDir` entries from `star_repos.json`
9. Do not keep business-layer compatibility branches for the old schema after the cutover runs.
10. Update detailed config / feature / architecture docs, but do not update `README.md` or `README_zh.md`.

## Storage Model

After the cutover, persisted storage becomes:

```text
<AppDataDir>/
  config.json
  config_local.json
  star_repos.json
  star_repos_local.json
  skills/
  meta/
  meta_local/
  prompts/
  cache/
    viewstate/
    backup_snapshot.json
  runtime/
  logs/

<RepoCacheDir>/        # defaults to <AppDataDir>/cache/repos
  <host>/<owner>/<repo>/
```

Key consequences:

- installed skills and their synced metadata always live under `<AppDataDir>`
- cloud backup and Git backup always read `<AppDataDir>`
- large Git clone caches can move elsewhere without affecting synced data layout
- `star_repos.json` stops carrying machine-specific clone paths

## Runtime Changes

### Config

- `config.AppConfig` removes `SkillsStorageDir`
- `config.AppConfig` adds `RepoCacheDir`
- `config_local.json` stores `repoCacheDir`
- `config.DefaultConfig(dataDir)` uses:
  - skills root: `<AppDataDir>/skills`
  - repo cache root: `<AppDataDir>/cache/repos`

### Skill and prompt paths

- installed skill storage always resolves from `<AppDataDir>/skills`
- skill metadata always resolves from `<AppDataDir>/meta`
- local skill overlay metadata always resolves from `<AppDataDir>/meta_local`
- prompt storage always resolves from `<AppDataDir>/prompts`

### Backup

- `BackupProfile` no longer needs `SkillsStorageDir`
- backup root calculation becomes a stable `<AppDataDir>`
- Git backup compatibility logic that depended on external skills roots is removed or simplified

### Starred repositories

- clone paths are derived from `repoCacheDir`
- `star_repos.json` persists only sync-safe repo identity data
- `star_repos_local.json` continues to store local runtime sync state such as `lastSync` and `syncError`

## Settings UX

Settings changes:

- remove the editable “Local Skills Storage Directory” field
- add an editable “Repository Cache Directory” field with folder picker
- add an “Open App Data Directory” action in the General settings section
- optionally show the resolved app-data path next to that action, matching the existing log-directory pattern

No UI is added for moving `viewstate`, runtime files, or logs.

## Error Handling

- blank `repoCacheDir` normalizes to `<AppDataDir>/cache/repos`
- changing `repoCacheDir` only affects future clone/update/read paths for repository caches
- if a configured repo cache directory does not exist, clone/update operations create parent directories as needed
- startup cutover silently removes obsolete `skillsStorageDir` and `star_repos.json.localDir` fields

## Scope

- `core/config`
- `core/platform/appdata`
- `core/platform/upgrade`
- `core/platform/git`
- `core/skillsource`
- `core/backup`
- `cmd/skillflow/`
- `cmd/skillflow/frontend/`
- `docs/config*.md`
- `docs/features*.md`
- `docs/architecture/use-cases*.md`
- `docs/architecture/runtime-and-storage*.md`

## Out Of Scope

- keeping compatibility for external installed skill directories after the cutover
- moving runtime helper state or logs outside `<AppDataDir>`
- introducing submodule-based backup
- changing README wording
