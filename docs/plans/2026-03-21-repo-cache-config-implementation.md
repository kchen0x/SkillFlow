# Repo Cache Path Cutover Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Remove configurable external skill storage, keep all synced app content under `<AppDataDir>`, and add a local-only configurable repository clone cache directory plus an AppDataDir open action in Settings.

**Architecture:** Collapse skill, meta, prompt, and backup roots back to `<AppDataDir>`, then isolate only the heavyweight cloned-repository cache behind a new local-only `repoCacheDir` setting. Run an explicit startup cutover to remove obsolete `skillsStorageDir` and synced `star_repos.json.localDir` fields so runtime code only reads the new schema.

**Tech Stack:** Go, Wails, React, TypeScript, Markdown docs

---

### Task 1: Replace the local skills-path setting with a local repo-cache setting

**Files:**
- Modify: `core/platform/appdata/appdata.go`
- Modify: `core/platform/appdata/appdata_test.go`
- Modify: `core/skillcatalog/app/settings.go`
- Modify: `core/skillcatalog/app/settings_test.go`
- Modify: `core/config/model.go`
- Modify: `core/config/defaults.go`
- Modify: `core/config/service.go`
- Modify: `core/config/service_test.go`

**Step 1: Write the failing tests**

Add or update Go tests that prove:
- default app config uses `<AppDataDir>/skills` for installed skills without any configurable override
- default app config uses `<AppDataDir>/cache/repos` for `repoCacheDir`
- saving config persists `repoCacheDir` only in `config_local.json`
- loading config normalizes an empty `repoCacheDir` to the default
- `config.json` no longer carries `skillsStorageDir`

Suggested test names:
- `TestDefaultConfigUsesDefaultRepoCacheDir`
- `TestSavePersistsRepoCacheDirOnlyInLocalConfig`
- `TestLoadNormalizesEmptyRepoCacheDir`
- `TestLegacySharedConfigDoesNotKeepSkillsStorageDir`

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./core/config ./core/platform/appdata ./core/skillcatalog/app -run 'TestDefaultConfigUsesDefaultRepoCacheDir|TestSavePersistsRepoCacheDirOnlyInLocalConfig|TestLoadNormalizesEmptyRepoCacheDir|TestLegacySharedConfigDoesNotKeepSkillsStorageDir'
```

Expected: `FAIL` because `repoCacheDir` does not exist yet and the code still persists `skillsStorageDir`.

**Step 3: Write minimal implementation**

Implement:
- a new app-data helper for the default repo cache root
- removal of local skill-path config from `skillcatalog` local settings
- `RepoCacheDir` on `config.AppConfig`
- local-config split/merge logic for `repoCacheDir`
- normalization that falls back to `<AppDataDir>/cache/repos`

Keep the installed skill root fixed at `<AppDataDir>/skills`.

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./core/config ./core/platform/appdata ./core/skillcatalog/app -run 'TestDefaultConfigUsesDefaultRepoCacheDir|TestSavePersistsRepoCacheDirOnlyInLocalConfig|TestLoadNormalizesEmptyRepoCacheDir|TestLegacySharedConfigDoesNotKeepSkillsStorageDir'
```

Expected: `ok`

**Step 5: Commit**

Run:

```bash
git add core/platform/appdata/appdata.go core/platform/appdata/appdata_test.go core/skillcatalog/app/settings.go core/skillcatalog/app/settings_test.go core/config/model.go core/config/defaults.go core/config/service.go core/config/service_test.go
git commit -m "refactor: replace skills path setting with repo cache path"
```

### Task 2: Cut over persisted storage and stop syncing starred-repo clone paths

**Files:**
- Modify: `core/platform/upgrade/upgrade.go`
- Create: `core/platform/upgrade/storage_path_cutover_test.go`
- Modify: `core/platform/git/client.go`
- Modify: `core/platform/git/client_test.go`
- Modify: `core/skillsource/infra/repository/star_repo_storage.go`
- Modify: `core/skillsource/infra/repository/star_repo_storage_test.go`
- Modify: `core/skillsource/app/service.go`
- Modify: `core/skillsource/app/service_test.go`

**Step 1: Write the failing tests**

Add or update tests that prove:
- startup upgrade removes `skillsStorageDir` from legacy config payloads
- startup upgrade removes `localDir` from legacy `star_repos.json` entries
- missing `repoCacheDir` is backfilled during cutover
- repo clone paths are derived from the configured cache root, not from `<AppDataDir>/cache`
- `star_repos.json` serialization no longer writes `localDir`
- loading starred repos still returns runtime `LocalDir` values derived from the current repo cache root

Suggested test names:
- `TestRunRemovesLegacySkillsStorageDirAndBackfillsRepoCacheDir`
- `TestRunRemovesLegacyStarRepoLocalDir`
- `TestCacheDirUsesConfiguredRepoCacheRoot`
- `TestStarRepoStorageDoesNotPersistLocalDir`
- `TestStarRepoStorageResolvesLocalDirFromRepoCacheRoot`

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./core/platform/upgrade ./core/platform/git ./core/skillsource/... -run 'TestRunRemovesLegacySkillsStorageDirAndBackfillsRepoCacheDir|TestRunRemovesLegacyStarRepoLocalDir|TestCacheDirUsesConfiguredRepoCacheRoot|TestStarRepoStorageDoesNotPersistLocalDir|TestStarRepoStorageResolvesLocalDirFromRepoCacheRoot'
```

Expected: `FAIL` because the upgrade path and star-repo persistence still use legacy fields.

**Step 3: Write minimal implementation**

Implement:
- upgrade-time migration for `config.json`, `config_local.json`, and `star_repos.json`
- repo cache path derivation rooted at `repoCacheDir`
- star-repo storage that persists only sync-safe repo identity fields
- runtime rehydration of `StarRepo.LocalDir` from the current cache root

Do not keep runtime compatibility branches for the old persisted shape after the cutover.

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./core/platform/upgrade ./core/platform/git ./core/skillsource/... -run 'TestRunRemovesLegacySkillsStorageDirAndBackfillsRepoCacheDir|TestRunRemovesLegacyStarRepoLocalDir|TestCacheDirUsesConfiguredRepoCacheRoot|TestStarRepoStorageDoesNotPersistLocalDir|TestStarRepoStorageResolvesLocalDirFromRepoCacheRoot'
```

Expected: `ok`

**Step 5: Commit**

Run:

```bash
git add core/platform/upgrade/upgrade.go core/platform/upgrade/storage_path_cutover_test.go core/platform/git/client.go core/platform/git/client_test.go core/skillsource/infra/repository/star_repo_storage.go core/skillsource/infra/repository/star_repo_storage_test.go core/skillsource/app/service.go core/skillsource/app/service_test.go
git commit -m "refactor: cut over repo cache and starred repo storage"
```

### Task 3: Rewire backup, prompt, viewstate, restore, and update flows to fixed AppDataDir roots

**Files:**
- Modify: `core/backup/domain/types.go`
- Modify: `core/backup/app/service.go`
- Modify: `core/backup/app/service_test.go`
- Modify: `cmd/skillflow/adapters.go`
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/app_backup.go`
- Modify: `cmd/skillflow/app_backup_test.go`
- Modify: `cmd/skillflow/app_restore_test.go`
- Modify: `cmd/skillflow/app_viewstate.go`
- Modify: `cmd/skillflow/app_viewstate_test.go`
- Modify: `cmd/skillflow/app_skill_update_test.go`

**Step 1: Write the failing tests**

Add or update tests that prove:
- Git/cloud backup root is always `<AppDataDir>`
- `BackupProfile` no longer depends on `SkillsStorageDir`
- `OpenGitBackupDir` opens `<AppDataDir>`
- installed-skill viewstate fingerprints read `meta/` and `meta_local/` from `<AppDataDir>`
- cached skill source lookup uses `repoCacheDir`
- restore compensation reclones starred repos into `repoCacheDir`

Suggested test names:
- `TestBackupProfileUsesConfigServiceDataDirOnly`
- `TestBackupRootDirAlwaysUsesAppDataDir`
- `TestInstalledSkillsFingerprintUsesAppDataMetaDirs`
- `TestCachedSkillSourceDirUsesRepoCacheDir`
- `TestRestoreReclonesStarredReposIntoRepoCacheDir`

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./core/backup ./cmd/skillflow -run 'TestBackupProfileUsesConfigServiceDataDirOnly|TestBackupRootDirAlwaysUsesAppDataDir|TestInstalledSkillsFingerprintUsesAppDataMetaDirs|TestCachedSkillSourceDirUsesRepoCacheDir|TestRestoreReclonesStarredReposIntoRepoCacheDir'
```

Expected: `FAIL` because backup and cache consumers still depend on legacy skill-path behavior.

**Step 3: Write minimal implementation**

Implement:
- a simplified backup profile and backup-root calculation
- fixed `<AppDataDir>` prompt / meta / backup path resolution
- repo cache root usage in cached-source and restore flows
- removal of legacy external-skills-root assumptions from shell wiring

Keep `cache/viewstate`, runtime state, and logs under `<AppDataDir>`.

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./core/backup ./cmd/skillflow -run 'TestBackupProfileUsesConfigServiceDataDirOnly|TestBackupRootDirAlwaysUsesAppDataDir|TestInstalledSkillsFingerprintUsesAppDataMetaDirs|TestCachedSkillSourceDirUsesRepoCacheDir|TestRestoreReclonesStarredReposIntoRepoCacheDir'
```

Expected: `ok`

**Step 5: Commit**

Run:

```bash
git add core/backup/domain/types.go core/backup/app/service.go core/backup/app/service_test.go cmd/skillflow/adapters.go cmd/skillflow/app.go cmd/skillflow/app_backup.go cmd/skillflow/app_backup_test.go cmd/skillflow/app_restore_test.go cmd/skillflow/app_viewstate.go cmd/skillflow/app_viewstate_test.go cmd/skillflow/app_skill_update_test.go
git commit -m "refactor: fix backup and runtime roots to app data"
```

### Task 4: Add Settings transport and UI for repo cache and AppDataDir actions

**Files:**
- Modify: `cmd/skillflow/app_settings.go`
- Create: `cmd/skillflow/app_settings_test.go`
- Modify: `cmd/skillflow/frontend/package.json`
- Create: `cmd/skillflow/frontend/src/lib/settingsPaths.ts`
- Create: `cmd/skillflow/frontend/tests/settingsPaths.test.mjs`
- Modify: `cmd/skillflow/frontend/src/pages/Settings.tsx`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.js`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`
- Modify: `cmd/skillflow/frontend/wailsjs/go/models.ts`

**Step 1: Write the failing tests**

Add tests that prove:
- saving config with a changed `repoCacheDir` rewires shell services that depend on the repo cache root
- a new frontend helper returns only `repoCacheDir` and `AppDataDir` actions, with no legacy skills-directory field
- the helper keeps the app-data open action available even when `repoCacheDir` is customized

Suggested test names:
- `TestSaveConfigRebuildsRepoCacheConsumers`
- `settingsPaths exposes repo cache field and app data action`
- `settingsPaths omits legacy skills directory row`

**Step 2: Run test to verify it fails**

Run backend test:

```bash
go test ./cmd/skillflow -run 'TestSaveConfigRebuildsRepoCacheConsumers'
```

Run frontend test:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `FAIL` because the new backend methods/helpers and UI wiring do not exist yet.

**Step 3: Write minimal implementation**

Implement:
- backend methods such as `GetAppDataDir()` and `OpenAppDataDir()`
- settings-save rewiring for repo-cache-dependent services
- Settings General UI that:
  - removes the editable skills directory field
  - adds the editable repo cache directory field with folder picker
  - adds the AppDataDir open action
- localized strings for the new repo cache and app-data controls
- regenerated Wails bindings and models reflecting `repoCacheDir`

**Step 4: Run test to verify it passes**

Run backend test:

```bash
go test ./cmd/skillflow -run 'TestSaveConfigRebuildsRepoCacheConsumers'
```

Run frontend test:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: backend `ok`, frontend `PASS`

**Step 5: Regenerate bindings and verify**

Run:

```bash
make generate
rg -n "OpenAppDataDir|GetAppDataDir|repoCacheDir" cmd/skillflow/frontend/wailsjs/go
```

Expected: generated bindings and models include the new methods/field, and no generated app config field named `skillsStorageDir` remains.

**Step 6: Commit**

Run:

```bash
git add cmd/skillflow/app_settings.go cmd/skillflow/app_settings_test.go cmd/skillflow/frontend/package.json cmd/skillflow/frontend/src/lib/settingsPaths.ts cmd/skillflow/frontend/tests/settingsPaths.test.mjs cmd/skillflow/frontend/src/pages/Settings.tsx cmd/skillflow/frontend/src/i18n/en.ts cmd/skillflow/frontend/src/i18n/zh.ts cmd/skillflow/frontend/wailsjs/go/main/App.js cmd/skillflow/frontend/wailsjs/go/main/App.d.ts cmd/skillflow/frontend/wailsjs/go/models.ts
git commit -m "feat: expose repo cache and app data settings"
```

### Task 5: Update config, feature, and architecture docs without touching README

**Files:**
- Modify: `docs/config.md`
- Modify: `docs/config_zh.md`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `docs/architecture/use-cases.md`
- Modify: `docs/architecture/use-cases_zh.md`
- Modify: `docs/architecture/runtime-and-storage.md`
- Modify: `docs/architecture/runtime-and-storage_zh.md`

**Step 1: Write the failing verification**

Verify current docs still mention the old skill-path behavior and synced starred-repo local clone paths.

Run:

```bash
rg -n "skillsStorageDir|Local Skills Storage Directory|localDir|shared parent of skills|外部目录|本地 Skills 存储目录" docs/config.md docs/config_zh.md docs/features.md docs/features_zh.md docs/architecture/use-cases.md docs/architecture/use-cases_zh.md docs/architecture/runtime-and-storage.md docs/architecture/runtime-and-storage_zh.md
```

Expected: matches show the legacy wording that must be removed.

**Step 2: Write minimal doc updates**

Update docs to state:
- `skills/`, `meta/`, `meta_local/`, and `prompts/` stay under `<AppDataDir>`
- `repoCacheDir` is the only configurable large-cache path
- `star_repos.json` no longer stores `localDir`
- Settings includes repo cache editing and an AppDataDir open action
- README files remain unchanged

Also update the “Last updated” dates where required.

**Step 3: Verify docs**

Run:

```bash
rg -n "repoCacheDir|Open App Data Directory|打开 AppDataDir|localDir|skillsStorageDir" docs/config.md docs/config_zh.md docs/features.md docs/features_zh.md docs/architecture/use-cases.md docs/architecture/use-cases_zh.md docs/architecture/runtime-and-storage.md docs/architecture/runtime-and-storage_zh.md
```

Expected:
- new `repoCacheDir` and AppDataDir wording appears
- no remaining intended user-facing guidance still describes configurable `skillsStorageDir`
- no doc change was made to `README.md` or `README_zh.md`

**Step 4: Commit**

Run:

```bash
git add docs/config.md docs/config_zh.md docs/features.md docs/features_zh.md docs/architecture/use-cases.md docs/architecture/use-cases_zh.md docs/architecture/runtime-and-storage.md docs/architecture/runtime-and-storage_zh.md
git commit -m "docs: document repo cache path cutover"
```

### Task 6: Final verification

**Files:**
- Verify only

**Step 1: Run core backend tests**

Run:

```bash
go test ./core/...
```

Expected: `ok`

**Step 2: Run shell tests**

Run:

```bash
go test ./cmd/skillflow
```

Expected: `ok`

**Step 3: Run frontend unit tests**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `PASS`

**Step 4: Run frontend build**

Run:

```bash
cd cmd/skillflow/frontend && npm run build
```

Expected: successful TypeScript + Vite production build

**Step 5: Review diff**

Run:

```bash
git diff --stat
```

Expected: changes are limited to the planned config, backup, skillsource, shell, frontend, and doc files. `README.md` and `README_zh.md` remain unchanged.
