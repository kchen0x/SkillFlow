# SkillFlow — Complete Feature Reference

> 🌐 [中文版](features_zh.md) | **English**
>
> This document enumerates every feature, button, interaction, and UX detail in SkillFlow.
> **Keep this file in sync whenever features are added, changed, or removed.**

---

## Table of Contents

1. [Navigation & Shell](#1-navigation--shell)
2. [My Skills (Dashboard)](#2-my-skills-dashboard)
3. [Push to Agents](#3-push-to-agents)
4. [Pull from Agents](#4-pull-from-agents)
5. [Starred Repos](#5-starred-repos)
6. [Cloud Backup](#6-cloud-backup)
7. [Settings](#7-settings)
8. [Skill Card](#8-skill-card)
9. [Skill Tooltip](#9-skill-tooltip)
10. [Shared Dialogs](#10-shared-dialogs)
11. [Backend Events](#11-backend-events)
12. [App Update Dialog](#12-app-update-dialog)
13. [My Agents](#13-my-agents)
14. [My Prompts](#14-my-prompts)
15. [My Memory](#15-my-memory)

---

## 1. Navigation & Shell

A fixed left sidebar (w-56) provides navigation throughout the app.

| Route | Icon | Label |
|-------|------|-------|
| `/` | Package | My Skills |
| `/tools` | Wrench | My Agents |
| `/prompts` | FileText | My Prompts |
| `/sync/push` | ArrowUpFromLine | Push to Agents |
| `/sync/pull` | ArrowDownToLine | Pull from Agents |
| `/starred` | Star | Starred Repos |
| `/backup` | Cloud | Cloud Backup |
| `/settings` | Settings | Settings |

- Active route: highlighted with a subtle theme-tinted surface, soft border, and restrained elevation shadow.
- Inactive routes: gray text with hover highlight.
- Top-left of sidebar: the `SkillFlow` wordmark shows the app icon immediately to the left; the icon is slightly taller than the text for clarity and preserves its aspect ratio.
- Top-right of sidebar: **Languages** shortcut button; toggles immediately between **Chinese** and **English**, and persists the preference to `localStorage`.
- Next to it: **Palette** theme shortcut button; cycles immediately through **Dark → Young → Light**.
- Bottom-left **Feedback** button: opens the GitHub "new issue" page in the default browser.
- Window close button behavior: clicking the top-left close button closes the current UI process instead of only hiding it. The lightweight helper remains in the tray/menu bar, so `Show SkillFlow` or launching the app again opens a fresh window.
- Primary-page refresh behavior: route transitions remount the page subtree keyed by `location.pathname`, so re-entering pages such as **My Skills**, **My Agents**, **My Prompts**, and **Starred Repos** fetches fresh backend state again without requiring a manual in-page refresh.
- Startup smoothing behavior: after the shell is ready, background startup jobs are staggered instead of launching together, so the first interactive second is less likely to compete with skill update checks, starred refresh, app update checks, and git startup pull at the same time.
- Window reactivation behavior: when the window is shown again after staying hidden in the tray/menu bar or inactive for a long time, SkillFlow now keeps the current routed page tree mounted. The visible page content, scroll position, and in-progress interactions remain available immediately instead of being replaced by a temporary memory-trim placeholder or forcing a route remount.
- Initial window sizing: on launch, SkillFlow first restores the most recently saved window size from local-only `config_local.json`; if none is saved yet, it sizes itself against the current display with a larger desktop-friendly default, clamps to the available screen, and centers the window.
- macOS tray behavior: a lightweight helper keeps a monochrome status icon in the menu-bar status area. Closing the main window exits the UI process, but the menu-bar item remains available with `Show SkillFlow`, `Hide SkillFlow`, and `Quit SkillFlow`; `Show SkillFlow` relaunches or focuses the UI window.
- Windows tray behavior: a lightweight helper remains in the system notification area with the app icon. Closing the main window exits the UI process; the tray menu can still `Show SkillFlow` to relaunch/focus the UI, or `Exit` to terminate both helper and UI.

---

## 2. My Skills (Dashboard)

Central library for managing your skill collection.

### Toolbar

| Control | Action |
|---------|--------|
| **Search input** | Wide search field for real-time case-insensitive filter by skill name; the toolbar wraps on narrower window widths so controls stay visible |
| **Sort toggle** | Two-button toggle for alphabetical order by skill name: **A-Z** or **Z-A** |
| **Update** (RefreshCw) | Calls backend `CheckUpdates()`; groups installed Git-backed skills by normalized repo source + subpath, compares installed `SourceSHA` against the same subpath inside the local repo cache, refreshes `LatestSHA` (synced) and `LastCheckedAt` (local-only), marks updatable cards with a red dot and Update action, and clears stale markers when an installed instance is already current. This checks only — it does not overwrite local files by itself |
| **Batch Delete** (CheckSquare) | Toggles multi-select mode |
| **Import** (FolderOpen) | Opens native folder-picker → `ImportLocal(dir)` |
| **Auto Update** (ToggleLeft / ToggleRight) | Uses a clear on/off state button in the old remote-install slot: when auto update is off, the control shows a muted "Turn Auto Update On" action; when it is on, it switches to a highlighted "Turn Auto Update Off" action. Successful toggles show a confirmation notice, and enabled state keeps the same local-only automatic update behavior after startup or manual starred-repo refresh |

### Auto Push Targets

- A compact single-row strip under the toolbar shows the **Auto Push Targets** title and agent chips together, using the same icon-chip selection style as **Push to Agents**.
- The selection is persisted locally on the current device and reused for future imports into **My Skills**.
- The toolbar **Auto Update** toggle is also persisted locally on the current device, alongside the auto-push target selection.
- The button now exposes the next action instead of a static label, and also uses left/right toggle icons plus contrastful styling so the current state is visible before you click it.
- Turning an agent on here immediately backfills the current library to that agent, so existing My Skills entries are pushed right away instead of waiting for the next import.
- Any newly added skill in **My Skills** is automatically copied to the selected agents after the library import succeeds. This applies to local folder import, Pull from Agents, and Starred Repo import.
- If a cloud restore brings skills onto the current device, SkillFlow auto-pushes any newly restored or newly updated library skills to the selected agents on this device.
- Import auto-push remains non-destructive: if a selected agent already contains a same-name skill in its `PushDir`, SkillFlow skips that target instead of overwriting it.
- Turning an agent off in this strip does not delete anything that was already pushed earlier; removing agent copies still requires manual deletion from **My Agents** or the agent directory.

### Select Mode (activated by "Batch Delete")

| Control | Action |
|---------|--------|
| **Select All / Deselect All** | Toggles all currently filtered skills |
| **Delete (n)** (Trash2, red) | `DeleteSkills(ids)` — disabled when nothing selected |
| **Cancel** | Exits select mode and clears selections |

### Category Sidebar

- Lists all categories; clicking one filters the skill grid.
- **"All" button** — shows every skill regardless of category.
- **Drag-and-drop target** — dragging a skill card onto a category moves it there.
- **Drop highlight** — the target category is highlighted more prominently while dragging a skill card over it.
- **Right-click context menu** on each category:
  - **Rename** — shows inline text input; confirm with Enter, cancel with Escape; calls `RenameCategory()`. (Not available for `Default`.)
  - **Delete** — deletes the category immediately when it is empty; if it still contains skills, shows a blocking dialog telling the user to clear the category first. (`Default` remains undeletable.)
- **New Category** (Plus icon at bottom) — shows inline text input; confirm with Enter or blur, cancel with Escape; calls `CreateCategory()`.

### Skill Grid

- Grid layout: 3 columns, 4 on wide screens.
- List performance behavior: page entry now prefers a local derived snapshot for installed skills when dependencies still match, and route transitions use a lighter fade-only animation to reduce navigation jank.
- Dense-list animation fallback: when the visible Dashboard list exceeds 18 cards, staggered per-card entry motion is disabled automatically and cards render immediately instead of animating one-by-one.
- Card status-strip overflow: status badges and pushed-agent icons now use deterministic truncation with compact `+N` overflow instead of per-card runtime measurement.
- **Empty state** — "No Skills found" message with usage hint.
- **Right-click skill menu** — includes move-to-category actions for every category other than the current one, plus delete and update where applicable.
- **Drag-and-drop** — drag a skill card to a category in the sidebar to move it; when drag starts, a smaller floating card follows the cursor; once a sidebar category is targeted, the original card collapses into a thin line until drag ends. Dragging a folder from the OS file manager onto the window imports it directly.
- **Window-level drag overlay** — semi-transparent indigo overlay with "Release to import Skill" message activates when a file is dragged over the window.
- **Hover tooltip** — appears after 300 ms hovering over a card (see [Skill Tooltip](#9-skill-tooltip)).

### Skill Update Flow

- Toolbar **Update** compares installed Git-backed skills against their locally cached repo clones and only marks cards as updatable.
- Card-level **Update** is the action that copies the latest files from the local repo cache, overwrites the installed copy in **My Skills**, refreshes any same-skill copies that already exist in agent `PushDir`s, and force-updates all agents currently selected in **Auto Push Targets** (creating missing copies and overwriting existing copies there).
- When **Auto Update** is enabled on the Dashboard, the same installed-skill update flow runs automatically after startup starred-repo refresh and after manual single-repo or refresh-all actions in **Starred Repos**.
- Automatic updates still only create or overwrite agent copies for the agents selected in **Auto Push Targets**. Other agents keep their existing copies and surface **Updatable** when those copies fall behind the installed library version.
- While a card update is running, that card keeps the Update action visible, disables repeat clicks, and spins the Refresh icon until the request finishes.
- The Dashboard also shows a temporary top-of-page status banner for skill update progress, success, or failure so users can tell whether the click really triggered work.
- Cache-based checks are grouped by the same logical git key used elsewhere in the app: normalized repo source + repo subpath.
- When multiple installed instances point at the same logical git skill, one cached SHA lookup updates their check state together, but each installed instance still decides its own `updatable` state by comparing its own `SourceSHA`.
- If the corresponding local repo cache is missing or stale, SkillFlow does not fall back to a direct GitHub API lookup here; users need the normal repo-cache sync path to refresh that clone first.

```text
[Dashboard toolbar Update]
          |
          v
     CheckUpdates()
          |
          v
Compare installed SourceSHA with cached repo subpath SHA
          |
          +--> no newer SHA  -> keep card unchanged
          |
          +--> newer SHA     -> store LatestSHA -> show red dot / Update action

[Dashboard card Update]
          |
          v
   UpdateSkill(skillID)
          |
          v
Copy latest repo subpath from local cache clone
          |
          v
Overwrite installed folder in My Skills
          |
          v
Refresh existing agent PushDir copies for that skill
          |
          v
SourceSHA = LatestSHA -> clear update marker
```

---

## 3. Push to Agents

Copies skills from your library to external agent directories.

### Layout

- Uses a two-column layout similar to My Skills.
- Left sidebar shows category filters: **All** plus every existing category.
- Right side shows the agent selector, search + A-Z/Z-A sort controls, push mode controls, and a skill-card grid for the current category scope.
- Spacing is tuned so the adaptive startup window can usually show the header controls, the current skill grid, and the bottom push action together on common laptop/desktop displays before scrolling is needed.

### Agent Selection

- One toggle button per enabled agent (icon + name).
- Multiple agents can be selected simultaneously.
- Active category, agent, and scope buttons use a brighter theme-tinted background, a lighter border, and a clearer glow so selection remains obvious in dark mode without adding extra symbols.

### Sync Scope

Two push behaviors based on the current left-sidebar category filter:

| Mode | Behavior |
|------|----------|
| **Manual Select** | Shown to the left of **Push All / Push Current Category** and selected by default; uses the current sidebar filter as the candidate list, shows selection checkboxes on cards, and allows select-all for the visible list |
| **Push All / Push Current Category** | If the sidebar is on **All**, pushes the whole library; if a category is selected, pushes only that category |

### Missing Directory Check

Before pushing, the app calls `CheckMissingPushDirs()`. If any target agent directory does not exist yet, a confirmation dialog appears:

- Lists each missing agent name and its full directory path.
- **"Create & Push"** — creates the directory then proceeds.
- **"Cancel"** — aborts without creating anything.

### Conflict Handling

If a skill already exists in the target directory, a conflict dialog appears for each one (see [Conflict Dialog](#101-conflict-dialog)).

### Skill Grid

- Library cards surface only push-relevant state on this page: which agents already contain that logical skill in their `PushDir`.
- Cards also keep the installed-source badge from My Skills so users can still distinguish manual imports from git-backed installs while deciding what to push.
- The pushed-agent indicator uses compact agent icons with ellipsis overflow and hover-to-reveal full lists.

### Bottom Bar

- **"Start Push (n)"** button — disabled when no agents selected or skill count is zero; shows "Pushing…" while in progress.
- **"Push complete ✓"** — green success message after all pushes finish.

---

## 4. Pull from Agents

Imports skills from external agent directories into your library.

### Layout

- Uses the same two-column shell as Push to Agents.
- Left sidebar lists all categories and controls the import target category.
- Right side contains the source-agent selector, scan feedback, search + A-Z/Z-A sort controls, selectable skill grid, and bottom action bar.

### Agent Selection

- Same toggle buttons as Push; selecting a different agent resets the scanned list.
- The active import target category and selected source agent use the same brighter background, lighter border, and glow treatment as Push so the current choice stays visually distinct in dark mode.

### Scan

- **"Scan"** button — calls `ScanAgentSkills(agentName)`; recursively searches the agent's configured scan directories for `skill.md` files.
- Local agent scanning uses the same configurable depth limit from **Settings → General** (default `5`, saved range `1-20`).
- Shows animated "Scanning…" state while in progress.
- **Error alert** (red) if scan fails; **warning alert** (yellow) if no skills found.
- Agent-scan candidates are deduplicated and correlated by logical key first; same-name items are kept distinct when their content-derived keys differ.

### Skill Grid

- Appears after a successful scan.
- Search field filters the scanned skill list by name in real time.
- Two-button sort toggle switches between **A-Z** and **Z-A** ordering by skill name.
- When a scanned item correlates to an already installed My Skills entry, the card also shows that installed entry's source badge so users can tell manual imports from git-backed installs at a glance.
- Cards still show the pull-specific imported state; newly discovered scan items that do not correlate to an installed entry keep the source badge empty instead of guessing.
- After each scan, all skills start unchecked by default.
- Select individual skills, use "Select All / Deselect All" for the currently visible list, or use the matching square-style "Select Not Imported" toggle to bulk-select only visible skills that are not yet imported.
- Selection and pull conflicts are tracked by scanned path, so same-name skills from different agent folders remain independent.

### Bottom Bar

- **"Start Pull (n)"** button — calls `PullFromAgent()`.
- **"Pull complete ✓"** — green success message.
- Conflicts handled by the same [Conflict Dialog](#101-conflict-dialog).

---

## 5. Starred Repos

Browse and import skills directly from watched Git repositories without installing them into your library first.

- Repo scanning is recursive across the full clone, so nested skill folders such as `plugins/<plugin>/skills/<name>` are included; `skill.md` matching is case-insensitive.
- Repos whose root directory itself contains `SKILL.md` / `skill.md` are also treated as a single skill candidate.
- Recursive repo scanning is bounded by the configurable **Remote Repo Recursive Scan Depth** setting in **Settings → General** (default `5`, saved range `1-20`).

### View Modes

| Mode | Icon | Description |
|------|------|-------------|
| **Folder** | Folder | Grid of repo cards; click a card to drill into its skills |
| **Flat** | LayoutGrid | All skills from all repos shown in a single grid |

### Toolbar (Normal Mode)

| Button | Action |
|--------|--------|
| **Search input** | Visible in flat view and repo detail view; filters the current skill grid by name |
| **Sort toggle** | Visible in flat view and repo detail view; switches the current skill grid between **A-Z** and **Z-A** |
| **Batch Import** (CheckSquare) | Enters select mode when a skill grid is visible |
| **Update All** (RefreshCw) | `UpdateAllStarredRepos()` — clones/pulls all repos in parallel; icon spins while syncing |
| **Add Repo** (Plus, indigo) | Opens "Add Repo" dialog |

### Toolbar (Select Mode)

| Button | Action |
|--------|--------|
| **Select All / Deselect All** | Toggles all visible skills |
| **Push to Agents (n)** | Opens the Push to Agents dialog (see below) |
| **Import to My Skills (n)** | Opens the Import dialog |
| **Cancel** | Exits select mode |

### Repo Card (Folder View)

- Click to open the skill list for that repo.
- Card header now highlights the repo host (for example `github.com`) and shows a compact skill count so users can quickly compare repo scale before drill-down.
- **Open in Browser** (ExternalLink icon) — opens repo URL in default browser.
- **Update** (RefreshCw icon) — `UpdateStarredRepo(url)` — pulls latest commits.
- **Delete** (Trash2 icon, red on hover) — removes from starred list.
- Shows last sync time and any sync error below the repo name.

### Builtin Starter Repos Initialization

- Builtin starter repos are only seeded when `star_repos.json` does not exist yet (first initialization on a device/profile).
- Current builtin starter repos:
  - `https://github.com/anthropics/skills.git`
  - `https://github.com/ComposioHQ/awesome-claude-skills.git`
  - `https://github.com/affaan-m/everything-claude-code.git`
- Once the user deletes a builtin repo, the persisted list remains authoritative and later app launches will not auto-add that repo again.

### Repo Detail View (Drill-down)

- Breadcrumb back button (ChevronLeft) to return to the repo grid.
- Skills grid with same select/import behavior as flat view.
- While repo skills are still loading, both the toolbar count label and the content pane show **Loading...** instead of a transient `0 skills` / `No Skills found` empty state.
- Repo skill cards show only imported and pushed-agent state on this page.
- Imported badges are resolved from normalized repo source + subpath, so same-name skills from different repos are not conflated.
- Pushed state is rendered as agent-brand icons with hover-to-reveal full agent lists.

### Repo Sync vs Installed Skill Update

- Repo-card **Update** and toolbar **Update All** refresh the locally cached clone for the starred repo.
- This makes the latest repo contents visible in **Starred Repos** so the user can browse or import newer skill files.
- If cloud restore syncs a newly starred repo onto this device, SkillFlow immediately clones that repo locally so search, import, and direct push-to-agents are ready without waiting for a manual update.
- It does **not** overwrite the already installed copy in **My Skills**.
- If a skill has already been imported into the library, updating that installed copy still happens from **My Skills (Dashboard)** via the card-level **Update** action.

```text
                    Cross-page update flow

[Starred Repo card Update]                  [My Skills card Update]
            |                                          |
            v                                          v
   UpdateStarredRepo(url)                       UpdateSkill(skillID)
            |                                          |
            v                                          v
Refresh cached repo clone                      Overwrite installed library copy
            |                                          |
            v                                          v
Refresh Starred Repos list                     Clear update marker in My Skills
            |
            +--> enables newer import candidates
            x--> does NOT change installed My Skills files
```

### Add Repo Dialog

- URL input (HTTPS or SSH format); Enter key triggers add.
- **"Add"** button — `AddStarredRepo(url)`.
- If the repo requires HTTP authentication, an **HTTP Auth Dialog** appears automatically.
- If SSH auth fails, an **SSH Auth Error Dialog** explains required setup.
- Shows clone-in-progress state ("Cloning…").

### HTTP Auth Dialog

- Username + Password inputs (password is masked); Enter on password field confirms.
- **"Confirm"** — retries with `AddStarredRepoWithCredentials(url, user, pass)`.
- **"Cancel"** — aborts.
- Shows error if credentials are wrong.

### SSH Auth Error Dialog

- Explains SSH key setup checklist:
  - Key generated with `ssh-keygen`
  - Public key added to GitHub / GitLab
  - SSH agent running (`ssh-add`)
  - Suggestion to use HTTPS instead
- **"Close"** button.

### Import Dialog (to My Skills)

- Category selector (dropdown).
- **"Import n"** — `ImportStarSkills(paths, repoURL, category)`.
- **"Cancel"**.

### Push to Agents Dialog

- Description: "Copies skills directly to the agent directory; no need to import first."
- Lists all enabled agents as checkboxes with their push directory paths shown.
- **Empty state** message if no agents are configured.
- **"Push to n agents"** button.
- If conflicts exist, a follow-up dialog lists the exact `skill → agent` pairs that were skipped; **Overwrite All** only overwrites those listed pairs.
- **"Cancel"**.

### Missing Directory Confirmation

Same behavior as [Push to Agents page](#missing-directory-check): confirms before creating absent push directories.

### Push Conflict Dialog

When skills already exist in the target agent directory:

- Lists all conflicting skill names.
- **"Overwrite All"** (amber) — `PushStarSkillsToAgentsForce()`.
- **"Skip Conflicts"** — `PushStarSkillsToTools()` (already resolved; conflicts discarded).

---

## 6. Cloud Backup

Mirror your skill library to cloud storage. Two backend types are supported: **Object Storage** (Aliyun OSS, AWS S3, Azure Blob Storage, Google Cloud Storage, Tencent COS, Huawei OBS) and **Git Repository**.

### Status

- **Cloud disabled banner** (yellow) — shown when cloud backup is not configured; links to Settings.

### Actions

| Button | Object Storage label | Git label |
|--------|---------------------|-----------|
| **Backup Now** (Upload icon) | Backup Now | Backup Now |
| **Restore / Pull** (Download icon) | Restore from Cloud | Pull Remote |
| **Refresh** (RefreshCw) | Reloads the latest backup change result | Same |

- Backup Now and Restore are disabled when cloud is not configured.
- **"Backup complete / Git sync complete"** (green) / **"Backup/sync failed"** (red) status messages.

### Backup Change List

- After each successful backup or restore, the page shows only the files involved in that operation, not the full remote file set.
- Backup page also shows **Last sync completed at** based on the most recent successful backup/restore event in the current app session.
- Both object storage and Git render the same per-run list with action badges: `Added`, `Modified`, `Deleted`.
- Deleted entries show a deletion label instead of a file size.
- Refresh reloads the latest in-memory backup result for the current app session.
- Scrollable, max-height container.
- **Unified backup scope (all providers)** — backup always uses `AppDataDir` as the Git/cloud working root. Synced content lives there under `skills/`, `meta/`, and `prompts/`, while local-only `cache/`, `runtime/`, `logs/`, `meta_local/`, `.git/`, `config_local.json`, and `star_repos_local.json` are excluded. The starred-repo clone cache at `repoCacheDir` is local-only and never part of backup scope.
- **Custom object-storage prefix** — object storage providers let the user choose a parent `remotePath`; SkillFlow always writes under `<bucket>/<remotePath>/skillflow/` (or `<bucket>/skillflow/` when the parent path is empty).
- **Provider-specific cloud profiles** — each cloud provider keeps its own bucket/path/credential set; switching providers in Settings restores that provider's saved values instead of overwriting another provider's form state.
- **Portable synced paths** — local paths persisted inside synced metadata (such as `meta/*.json`) are stored as forward-slash relative paths under `AppDataDir`, so restores continue to work across macOS and Windows.
- **Local-only volatile skill metadata** — high-churn per-skill check timestamps (`LastCheckedAt`) are stored in local-only `meta_local/*.local.json` overlays, so they do not create cross-device git/cloud merge churn.
- **Local-only starred-repo sync runtime state** — per-repo `lastSync` and `syncError` are stored in local-only `star_repos_local.json`, so background sync attempts on one device do not churn synced repo metadata on other devices.
- **Local-only path config** — `config_local.json` stores machine-specific filesystem paths such as `repoCacheDir`, agent `ScanDirs` / `PushDir`, and proxy settings; it is excluded from backup and git sync.
- **Other local-only device state** — per-device choices and runtime shell state such as auto-push targets, launch-at-login registration, and the last saved window size remain in `config_local.json`, so restoring on one machine does not overwrite another machine's local behavior.
- **Local-only cloud secrets** — sensitive cloud credentials (for example access key IDs, secret keys, and access tokens) are stored only in per-provider entries inside `config_local.json`; synced `config.json` keeps only non-sensitive cloud settings such as provider, bucket name, remote path, endpoint, repo URL, or branch.
- **Git backup compatibility** — when Git backup uses a parent directory as the working tree, SkillFlow automatically moves any legacy nested `skills/.git` metadata aside so actual skill files remain trackable.
- **Post-restore device compensation** — after a successful cloud restore, newly restored library skills are auto-pushed to this device's selected auto-push agents, and newly restored starred repos are cloned locally right away.

### Provider Coverage

- Object storage now also supports AWS S3 (bucket + region), Azure Blob Storage (container + account name + optional service URL), and Google Cloud Storage (bucket + service-account JSON or local key file path), in addition to Aliyun OSS, Tencent COS, and Huawei OBS.
- Sync-safe connection fields now include `region`, `account_name`, and `service_url` alongside existing endpoint / repo URL / branch settings. Sensitive values such as account keys and service-account JSON remain local-only in `config_local.json`.

### Auto-Backup

Triggered automatically after any of these mutations (when cloud is enabled):

- Delete skill / bulk delete
- Create / update / delete prompt
- Manual import
- Install from GitHub
- Pull from agent
- Update skill
- Import from starred repo

Progress events surface in the UI via the Wails event system (`backup.started`, `backup.progress`, `backup.completed`, `backup.failed`).

### Git Sync (Git provider only)

When the **git** provider is selected:

- **Repository bootstrap** — if the Skills directory is not a git repo, SkillFlow auto-initializes it and configures `origin` from the configured repo URL.
- **Remote binding self-heal** — if `origin` is missing or changed, SkillFlow auto-adds/updates it before pull/push.
- **Startup pull** — on every app launch, SkillFlow runs `git pull` on the Git backup root directory to fetch the latest remote changes.
- **Staggered startup background work** — the startup pull is still automatic, but it no longer starts in the same burst as every other startup check; SkillFlow spreads those jobs out after the UI becomes interactive.
- **Missing branch tolerance** — if the configured remote branch does not exist yet (first-time setup), startup pull is skipped without failing the backup page.
- **Auto-push after mutations** — same post-mutation trigger as object storage; runs `git add -A && git commit && git push`.
- **Excluded-path cleanup** — on every Git backup push, local-only excluded directories such as `cache/` and `runtime/` are removed from the Git index if an older version ever tracked them.
- **Periodic auto-sync** — controlled by the "Auto-sync interval" setting (in minutes, 0 = disabled). A background timer fires `autoBackup()` on the configured interval.
- **Manual actions with conflict detection** — both **Backup Now** and **Restore / Pull** detect git conflicts/divergence and emit `git.conflict` when user action is required.
- **Conflict resolution dialog** — if `git pull` or `git push` detects a conflict or diverged history, a modal appears:
  - The dialog includes a conflict file list when available.
  - **"Keep Local"** — aborts the merge, force-pushes local state to remote. Calls `ResolveGitConflict(true)`.
  - **"Keep Remote"** — aborts the merge, resets local to `origin/<branch>`. Calls `ResolveGitConflict(false)`.
  - **"Resolve Manually"** — opens the Git backup root in the system file manager so the user can inspect and fix conflicted files directly.
  - The keep-local and keep-remote actions both reload app state from disk (skills/meta/config) and emit `git.sync.completed` on success.
- **State refresh after pull** — after successful startup pull or manual pull, app state is immediately reloaded from disk so changed `meta/` and config files take effect.
- If a conflict is detected during startup (before the UI loads), it is stored as a pending flag and surfaced when the Backup page mounts (`GetGitConflictPending()`).

---

## 7. Settings

Configuration panel with four tabs in this order: Agents, Cloud, Proxy, General.

The Settings page content expands with the window up to a wider desktop-friendly maximum width instead of staying pinned to a narrow fixed column.

### Agents Tab

For each built-in or custom agent:

| Control | Description |
|---------|-------------|
| **Enable toggle** | Enables or disables the agent across the app |
| **Push directory** | Single directory where skills are copied on push; supports both manual text entry and folder-picker button (FolderOpen icon), which opens at the current path or nearest existing parent |
| **Scan directories** | Multiple directories searched when pulling; each row has a folder-picker button and a delete button; new directories added with an input + folder-picker + "Add" button, with the picker reopening at the current path or nearest existing parent |
| **Delete agent** (custom agents only) | Removes the custom agent entry |

**Add Custom Agent** section (dashed border):

- Agent name input.
- Push directory input with folder-picker button that reopens at the current path or nearest existing parent.
- **"Add"** button — `AddCustomAgent(name, pushDir)`.

### Cloud Tab

| Control | Description |
|---------|-------------|
| **Provider buttons** | Responsive provider cards shown in a wrapping grid. **Git Repo** is always pinned first; the remaining providers keep the backend-returned order. Each provider restores its own saved bucket/path/credential draft when selected |
| **Bucket name** | Object storage bucket name (hidden when git provider is selected) |
| **Remote path** | Object storage parent path (optional). Users enter the parent folder only; SkillFlow always appends `/skillflow/` to build the final backup prefix |
| **Final backup path preview** | Real-time rendered object-storage destination shown as `<bucket>/<remotePath>/skillflow/` so users can verify the exact remote backup location before saving |
| **Credential fields** | Dynamically rendered from `RequiredCredentials()` — text or password inputs per provider. Sync-safe connection fields such as endpoint / repo URL / branch remain in `config.json`, while sensitive credentials such as access keys and tokens are stored only in `config_local.json`. Git fields: repo URL, branch, username, access token |
| **Input normalization** | Aliyun OSS and Huawei OBS bucket fields accept either a plain bucket name or a common full bucket host/URL and normalize it automatically. Tencent COS uses the same bucket + endpoint model as the other object-storage providers. The bucket value always comes from the dedicated bucket field, while the endpoint field may contain either a plain endpoint host or a full bucket host/URL and is preserved as entered for display |
| **Additional provider details** | AWS S3 trims the region field before saving. Azure Blob Storage uses the bucket field as the container name, stores account name and optional service URL as sync-safe fields, and defaults the service URL to `https://<account>.blob.core.windows.net/` when empty. Google Cloud Storage accepts either an inline service-account JSON string or a local key file path, and keeps that credential local-only in `config_local.json`. |
| **Auto-sync interval** | Number input (minutes); 0 = sync only after mutations; positive value starts a background periodic timer |
| **Enable auto backup toggle** | Turns on/off automatic post-mutation backups and the periodic timer |

### General Tab

| Control | Description |
|---------|-------------|
| **Language** | Two buttons, **中文** and **English**, switch the entire frontend language immediately; shares the same state as the sidebar **Languages** button and persists to `localStorage` |
| **Appearance theme** | Three visual presets shown as preview cards: **Young** (default, a softened paper-blue evolution of the previous sky-blue Light palette), **Dark** (refined graphite with muted mist-blue accents), and **Light** (new low-saturation gray-white palette inspired by Messor); persisted to `localStorage`; changes apply immediately without restart; legacy stored `Light` preference auto-migrates to `Young` |
| **Card status visibility** | A compact per-page row list that lets users hide or show only the statuses that page supports by default. Unsupported statuses are not offered for that page. Default policy: **My Skills** = updatable + pushed agents; **My Agents** = imported + updatable + pushed agents; **Push to Agent** = pushed agents; **Pull from Agent** = imported; **Starred Repos** = imported + pushed agents |
| **App data directory** | Fixed root for installed skills, prompts, metadata, and backup working state on the current device; shown read-only with a one-click open action |
| **Repository cache directory** | Local-only root path for cloned starred repositories; manual text entry + folder-picker button that opens at the current path or nearest existing parent; defaults to `<AppDataDir>/cache/repos` |
| **Skill recursive scan depth** | Maximum recursion depth used when scanning local agent directories and starred repos; default `5`; saved values are clamped to `1-20` to avoid pathological nested trees |
| **Default category** | Fixed system fallback category `Default` (read-only), used when pulling/importing without specifying a category |
| **Log level buttons** | Toggle runtime log level between `debug`, `info`, and `error` (default: `error`); takes effect after saving settings |
| **Launch at login toggle** | Enables/disables OS login-item registration so SkillFlow auto-starts after sign-in on the current device; stored only in local `config_local.json`. Reconcile runs on startup and Settings save: already-missing disabled entries are treated as a no-op, while the enabled path is refreshed to the current executable so moved or updated app installs keep working on macOS and Windows |
| **Open log directory** | One-click open the local log folder in system file manager; missing targets fall back to the nearest existing parent directory |

Log files are stored under the app log directory, with rolling limits:
- At most **2 files** are kept: `skillflow.log` and `skillflow.log.1`.
- Each file is capped at **1MB**.
- When `skillflow.log` reaches the limit, it rotates and overwrites the older backup file.

### Proxy Tab

Proxy settings for all remote operations (repo scan, repo-cache sync):

| Mode | Description |
|------|-------------|
| **No proxy** | Direct connection |
| **System proxy** | Reads `HTTP_PROXY` / `HTTPS_PROXY` environment variables |
| **Manual** | Custom proxy URL (http://, https://, socks5://) |

When Manual is selected, a URL input appears with format hint. Proxy settings are persisted in `config_local.json`, so manual proxy values survive restart and are not included in backup/git sync.

- A dedicated **Test Connection** block lets users verify proxy reachability without saving first.
- The target URL input defaults to `https://github.com`, but users can replace it with any `http` / `https` URL.
- The test uses the current in-memory proxy form state, so unsaved edits are honored immediately; if nothing changed, that state naturally matches the saved config.
- Each test enforces a **5-second timeout** and shows inline feedback with target URL, elapsed time, and either the received HTTP status or the transport error message.
- Receiving an HTTP response still counts as successful connectivity even when the status code is not `2xx` (for example `301` or `403`).

### Save Button

- **"Save Settings"** — `SaveConfig(cfg)`; disabled while saving.
- **Keyboard shortcut** — while the Settings page is open, `Ctrl+S` on Windows/Linux and `Cmd+S` on macOS trigger the same save action and suppress the browser/webview default save shortcut.

---

## 8. Skill Card

Reusable card component shown in the My Skills grid and Sync pages.

### Variants

**Dashboard card** (`SkillCard`):

| Element | Description |
|---------|-------------|
| **Status strip** | Source badge plus any coexisting state badges (for example Update available + pushed-agent icons) rendered in one compact header row on the card; the strip prefers a single line when space allows, then automatically wraps instead of clipping badges away when cards become too narrow |
| **Skill name** | Two-line clamp; padded to avoid overlap with action buttons |
| **Open folder button** (FolderOpenDot, top-right) | `OpenPath(skill.path)` — opens directory in OS file manager; visible on hover only |
| **Select checkbox** (top-left) | Visible in select mode only |
| **Hover actions** (bottom-right) | Update (if available) · Copy · Delete — hidden until hover, except the Update action stays visible while that card is actively updating |
| **Update action feedback** | Hover adds a lift/highlight animation; clicking switches the action into a disabled spinner state until the update finishes |
| **Copy button** | Reads `skill.md` content, copies to clipboard, shows "Copied ✓" for 2 s |
| **Drag handle** | Cards are draggable in normal mode; dragged `skillId` moves skill to drop target category |
| **Right-click context menu** | Update (if available) · Move to [Category] (one item per other category) · Delete (red) |

**Sync card** (`SyncSkillCard`):

| Element | Description |
|---------|-------------|
| **Status strip** | Source badge plus only the page-selected state badges (for example imported, update-available, or pushed-agent icons) rendered together in one compact header row; cards keep imported and pushed-agent indicators on the same line when they fit, and automatically wrap them to a second row when the card width is too tight |
| **Pushed-agent indicator** | Shows the exact agents whose `PushDir` already contains this logical skill via small agent-brand icons without an extra arrow prefix; overflows collapse into a compact count badge while hover still reveals the full list |
| **Skill name** | Two-line clamp |
| **Subtitle** | Category or repo name |
| **Copy button** (hover) | Same clipboard behavior |
| **Open folder button** (hover) | Same as dashboard card |
| **Selection checkbox** (bottom-right) | Shown when `showSelection = true` |

### Unified Status Semantics

| State | Meaning |
|------|---------|
| **installed** | The logical skill already exists in **My Skills** as at least one installed instance |
| **imported** | External-page wording for **installed**; on GitHub / Starred Repos / agent views it means “already in My Skills” |
| **pushed** | The logical skill already exists in an agent's configured **PushDir** |
| **pushedAgents** | The exact agent names whose configured **PushDir** currently contains that logical skill; used for icon rendering on cards |
| **seenInAgentScan** | The logical skill was detected in one of an agent's configured **ScanDirs**; this means the agent already has it somewhere, but not necessarily because SkillFlow pushed it. This is currently used for grouping and correlation, not shown as a card badge. |
| **updatable** | An installed Git-backed skill has a newer remote SHA than its installed `SourceSHA` |

These statuses are resolved by the backend from a unified logical-key model; frontend pages no longer infer them independently from `Name` or `Path`.

```text
                 Unified skill status picture

[GitHub candidate] [Starred skill] [Agent scan candidate]
         \              |               /
          \             |              /
           +---- same logical skill ----+
                        |
                        v
                 [My Skills instance]
                  installed = true
           external pages show imported = true
                        |
                        +---- copy to agent PushDir ---> pushed = true
                        |
                        +---- remote newer SHA -------> updatable = true

[Agent ScanDirs] --- detect same logical skill ----> seenInAgentScan = true
```

---

## 9. Skill Tooltip

A floating info panel that appears 300 ms after hovering over any skill card (dashboard only).

### Positioning

- Fixed position, 300 px wide, max 400 px tall.
- Prefers right side of card; falls back to left if near the right window edge.
- Shifts up if near the bottom of the window.

### Content

| Section | Fields shown |
|---------|-------------|
| **Header** | Icon (GitHub / folder) · skill name · source badge · category |
| **Description** | Parsed from `skill.md` YAML frontmatter; shows "No description" if absent; "Loading…" while fetching |
| **Frontmatter fields** | `argument_hint` (Tag icon) · `allowed_tools` (Wrench icon) · `context` (GitBranch icon) |
| **Metadata** | Repository URL (trimmed, opens on click) · installed SHA · available update SHA (amber) · installed date · updated date |

---

## 10. Shared Dialogs

### 10.1 Conflict Dialog

Shown one conflict at a time during push or pull when a skill already exists at the destination.

- Displays: "[Skill name] already exists. How to handle?"
- **"Skip"** — leaves existing file untouched, moves to next conflict.
- **"Overwrite"** — calls the `*Force` variant, replaces only the exact conflicting skill-target pair.
- Auto-closes when the conflict queue is empty.

### 10.2 Missing Directory Dialog

Appears before any push when target directories do not exist.

- Lists each affected agent name and full directory path.
- **"Create & Push"** — auto-creates directories then proceeds.
- **"Cancel"** — aborts.

---

## 11. Backend Events

Events emitted from the Go backend to the frontend via Wails runtime:

| Event | When fired | Payload |
|-------|-----------|---------|
| `backup.started` | Auto-backup begins | — |
| `backup.progress` | Each file uploaded | `{ currentFile: string }` |
| `backup.completed` | Backup finished | `{ files: Array<{ path, size, action }> }` |
| `backup.failed` | Backup error | — |
| `update.available` | New cached commit found for a skill | `{ skillID, skillName, currentSHA, latestSHA }` |
| `star.sync.progress` | One repo synced | `{ repoURL, repoName, syncError }` |
| `star.sync.done` | All repos synced | — |
| `git.sync.started` | Git pull/push begins | — |
| `git.sync.completed` | Git sync succeeded | `{ files: Array<{ path, size, action }> }` when triggered by backup/restore/conflict resolution |
| `git.sync.failed` | Git sync error | — |
| `git.conflict` | Git merge conflict detected | `{ message: string, files?: string[] }` |

The Dashboard listens to `update.available` and marks affected skill cards with a red update dot in real time.
The Backup page listens to all `git.*` events and surfaces the conflict resolution dialog on `git.conflict`.
`App.tsx` listens to all three `app.update.*` events, parses the emitted `AppUpdateInfo` payload, and drives the update dialog state machine.

---

## 12. App Update Dialog

A centered modal dialog that appears when a new app version is detected. Triggered by both the automatic startup check and the manual check in Settings. Driven by a four-state machine:

| State | Trigger | Dialog content |
|-------|---------|---------------|
| `available` | `app.update.available` event | Version labels + release notes + platform-specific action buttons |
| `downloading` | User clicks "Download & Auto-restart" (Windows only) | Spinner + progress message |
| `ready_to_restart` | `app.update.download.done` event | Completion message + "Restart Now" / "Later" buttons |
| `download_failed` | `app.update.download.fail` event | Error message + "Go to Downloads" button |

### Platform Behavior

Both startup and manual checks surface the same modal dialog.

- **Windows** — Three choices in the `available` state:
  1. **Download & Auto-restart** — downloads the new exe in the background, then prompts restart.
  2. **Open Release Page** — opens the GitHub Releases page in the system browser.
  3. **Skip this version (don't remind on next start)** — persists the skipped version; the startup check will not prompt for this version again. The manual check always shows the dialog regardless.
- **macOS** — Two choices in the `available` state (auto-download not supported):
  1. **Open Release Page** — opens the GitHub Releases page.
  2. **Skip this version (don't remind on next start)** — same skip behavior as Windows.
- The `available` dialog always renders the current version, latest version, and the Release-page action from the same `AppUpdateInfo` payload on both platforms.

### Skip Version Behavior

- The skipped version tag is stored in `AppConfig.SkippedUpdateVersion` and persisted in the shared `config.json` file, so it survives app restarts.
- On app startup, if `latestVersion == skippedUpdateVersion` the `app.update.available` event is **not** emitted and no dialog appears.
- When the user manually clicks "Check for Updates" in Settings, `CheckAppUpdateAndNotify` always emits the event, bypassing the skip — the dialog always appears for manual checks.
- Clicking "Skip this version" calls `SetSkippedUpdateVersion(latestVersion)`.

### Manual Check Button (Settings Page)

A **"Check for Updates"** button in the top-right corner of the Settings page header:

- Displays current app version (`vX.Y.Z`) next to the button.
- Click → calls `CheckAppUpdateAndNotify()`; button shows a spinner while checking.
- If a new version is found, the update dialog opens automatically via the `app.update.available` event.
- If already up-to-date, inline text shows "Already up to date (vX.Y.Z)".
- On error: "Check failed: …" shown inline.

### Controls

| Control | Action |
|---------|--------|
| **Download & Auto-restart** (Windows, `available`) | `DownloadAppUpdate(downloadUrl)` — starts async download |
| **Open Release Page** (`available`) | `OpenURL(releaseUrl)` — opens release page in browser; closes dialog |
| **Skip this version** (`available`) | `SetSkippedUpdateVersion(latestVersion)` — persists skip; closes dialog |
| **Restart Now** (`ready_to_restart`) | `ApplyAppUpdate()` — writes bat script and exits; bat replaces exe and relaunches |
| **Later** (`ready_to_restart`) | Closes dialog without restarting |
| **Go to Downloads** (`download_failed`) | `OpenURL(releaseUrl)` — opens release page in browser; closes dialog |
| **×** (all states except `downloading`) | Closes dialog for the current session |

### Backend Methods

| Method | Description |
|--------|-------------|
| `GetAppVersion()` | Returns current version string (injected by `-ldflags` at build time; `"dev"` in local dev) |
| `CheckAppUpdate()` | Queries GitHub Releases API; returns `AppUpdateInfo` with platform-matched download URL |
| `CheckAppUpdateAndNotify()` | Calls `CheckAppUpdate()` and emits `EventAppUpdateAvailable` if an update is found; always notifies regardless of skipped version |
| `GetSkippedUpdateVersion()` | Returns the version tag stored in config as the user-skipped version |
| `SetSkippedUpdateVersion(version)` | Persists the skipped version tag to config; pass `""` to clear |
| `DownloadAppUpdate(url)` | Downloads new exe to temp file asynchronously; emits `app.update.download.done` or `app.update.download.fail` |
| `ApplyAppUpdate()` | Windows only — writes bat script for post-exit exe replacement, then calls `os.Exit(0)` |

### Version Injection (CI)

GitHub Actions builds inject the tag name at compile time:
```
wails build -ldflags "-X main.Version=${{ github.ref_name }}"
```
The startup check is skipped when `Version == "dev"` (local development).

---

## 13. My Agents

Browse the skills currently present inside each enabled agent.

### Layout

- Left sidebar lists enabled agents.
- Main area shows one toolbar plus two skill-card sections: **Push Path** and **Scan Path**.

### Toolbar

| Control | Description |
|---------|-------------|
| **Search input** | Filters both Push Path and Scan Path skill cards by name in real time |
| **Sort toggle** | Switches both sections between **A-Z** and **Z-A** ordering by skill name |
| **Batch Delete** | Available when the visible Push Path list is non-empty; enters select mode |

### Push Path Section

- Shows deletable agent-local skills under the configured push directory.
- Push-path discovery uses the same configurable depth limit from **Settings → General** (default `5`, saved range `1-20`).
- In select mode, **Select All / Deselect All** applies to the currently visible filtered Push Path cards only.
- When a agent-local skill correlates to an installed My Skills entry, the card also shows that installed entry's source badge so users can tell manual imports from git-backed installs.
- Cards continue showing imported, update-available, and "pushed to other agents" states. The current agent is excluded from the pushed-agent icon list so the card only surfaces cross-agent distribution that adds information.

### Scan Path Section

- Shows read-only skills discovered only from scan directories.
- Scan-path discovery uses the same configurable depth limit from **Settings → General** (default `5`, saved range `1-20`).
- Shares the same search and sort controls as Push Path.
- Scan-path cards use the same compact strip for source / imported / update-available / pushed-to-other-agents states whenever the scanned item can be correlated to an installed My Skills entry; the fact that they were found via scan paths is conveyed by the section itself instead of repeating a detected badge on every card.

---

## 14. My Prompts

Store reusable system prompts inside the synced `prompts/` directory.

### Navigation & Storage

- Sidebar adds **My Prompts** directly below **My Agents**.
- Each prompt is stored as `prompts/<category>/<name>/system.md` plus a sidecar `prompts/<category>/<name>/prompt.json` under the backup root, so both object-storage providers and the Git provider sync the same prompt files automatically.
- Prompt names are required, globally unique in the library, and used as the folder key.

### Layout

- The page mirrors **My Skills** with a left category sidebar and a right content pane.
- Categories support **Default** as the built-in fallback group.
- Just like the other primary routes, re-entering **My Prompts** remounts the page and reloads current prompt data instead of keeping a long-lived stale page instance.
- Prompt cards are filtered by selected category, keyword search, and A-Z / Z-A sorting.
- Search supports logical `and` / `or` syntax, for example `review and golang` or `summary or changelog`.

### Category Management

- Categories can be created from the sidebar.
- Non-default categories support rename and delete from the context menu.
- Prompt cards can be dragged onto categories to move them, matching the category-drop behavior from **My Skills**.

### Prompt Cards

- Cards reuse the same visual language as Skill cards (`card-base`, hover glow, compact actions).
- Each card shows **name**, optional **description**, and the opening content excerpt by default.
- Cards no longer show related image thumbnails or saved web-link chips in the list, keeping the prompt grid focused on text scanning.
- When the content is longer than the preview window, the card displays **Click to view more**.
- Top-right action button copies the full prompt content to the desktop clipboard in one click, preserving multi-line content on Windows.
- Prompt copy now falls back across desktop runtime, browser clipboard, and document copy APIs so the action still succeeds when one clipboard path is unavailable.
- Bottom-right **Delete** action opens a confirmation dialog before removing both the card and the underlying prompt folder.

### Prompt Editor

- Clicking **Add Prompt** opens a built-in editor with fields for **name** (required), **description** (optional), **category**, and full `system.md` content.
- The editor supports up to **3 related image URLs**. Saved images stay visible in the image panel, while the image add-input row moves to the shared attachment area after the main prompt content for cleaner reading flow.
- Each saved image thumbnail exposes a small delete action in the top-right corner.
- Clicking an editor thumbnail opens an enlarged in-app preview overlay above the editor instead of sending the user to the external browser.
- The editor now keeps the footer reachable inside smaller desktop windows by constraining the body and prompt content area with internal scrolling instead of growing beyond the app window.
- The editor supports **web links** through a single-line markdown input placed in the shared attachment area after the main prompt content. Clicking **Add** appends the parsed link above as a clickable chip using the markdown label text from `[Label](https://example.com/doc)`, then clears the input field.
- Clicking an existing prompt card opens the same editor pre-filled for editing and rename operations.
- Saving writes the prompt body back to `prompts/<category>/<name>/system.md` and writes metadata such as description, image URLs, and web links to `prompts/<category>/<name>/prompt.json`.

### Import / Export

- Toolbar **Import** reads a JSON prompt library file, preserves the imported category for each prompt, and when a prompt name already exists locally it asks whether to **Skip** or **Overwrite** before writing anything.
- The import conflict dialog includes **Apply the same action to the remaining {count} conflicts**, so one decision can be reused for the rest of the current import run without turning it into a saved global preference.
- Toolbar **Export** now expands an inline export bar instead of exporting immediately.
- The export bar supports **All**, the currently selected left-sidebar category, and **Pick**.
- **Pick** supports multi-select export within the current left-sidebar scope. When the sidebar is on **All**, selection spans the whole prompt library; when the sidebar is on a concrete category, selection is limited to that category.
- Exported JSON still preserves each prompt's own category, image URLs, and web links.

---

## 15. My Memory

SkillFlow provides a unified memory management interface. Users can author and manage personal AI coding assistant memories in one place and push them to multiple AI tools.

### Memory Types

- **Main Memory**: A single `main.md` file containing global instructions shared across all configured agents.
- **Module Memories**: Individual topic-focused markdown files (e.g., `coding-style`, `testing-rules`) stored under `rules/`.

### Page Layout

The My Memory page shows:
- A prominent **Main Memory card** with per-agent push status chips.
- A **two-column module grid** with search, agent filter, and sort controls.
- A right-side **Edit Drawer** (~55 % width) with Edit / Preview tabs, push target selection, push mode, auto-push toggle, and a **Push Now** button.

### Push Modes (per agent)

| Mode | Behaviour |
|------|-----------|
| **Merge** | SkillFlow writes a managed marker block inside the agent's memory file. Content outside the block is preserved. |
| **Takeover** | SkillFlow owns the entire memory file and rules directory. |

### Push Status

Each agent shows one of three statuses:

| Status | Meaning |
|--------|---------|
| ✓ Synced | Last push matches current local memory. |
| ⚠ Pending | Local memory has changed since the last push. |
| Never Pushed | Push targets configured but never pushed. |

A red dot appears on the sidebar entry and on individual cards when any target agent has a pending push.

### Module File Naming

Module memories are written with an `sf-` prefix to the agent rules directory (e.g., `coding-style` → `sf-coding-style.md`).

### Open in External Editor

Each memory card provides an **Open in Editor** button that opens the file in the system default text editor. SkillFlow detects file changes and refreshes the content automatically.

### Agent Settings

In **Settings → Agents**, each agent now exposes:
- **Memory File** – path to the agent's main memory file (default shown; user-overridable).
- **Rules Directory** – path to the agent's rules directory.
- **Memory Push Mode** – Merge or Takeover.
- **Auto Push Memory** – push automatically whenever memory content changes.

### Cloud Backup

Memory content files (`memory/main.md` and `memory/rules/*.md`) are included in cloud backup. The local configuration file (`memory/memory_local.json`) is excluded from backup.

---

*Last updated: 2026-03-22*
