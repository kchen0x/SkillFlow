# SkillFlow

> 🌐 [中文](README_zh.md) | **English**

Slogan: Master once. Apply everywhere.

SkillFlow is a cross-platform desktop app for managing skills and reusable prompts across diverse agentic AI environments. It combines GitHub install, cross-environment sync, starred repo browsing, update checking, and cloud backup in one local-first workflow.

![skilflow](docs/skillflow.gif)

## Download & Install

Get the latest release from **[GitHub Releases →](https://github.com/shinerio/SkillFlow/releases/latest)**

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `SkillFlow-macos-apple-silicon` |
| macOS (Intel) | `SkillFlow-macos-intel` |
| Windows (x64) | `SkillFlow-windows-amd64.exe` |

## Highlights

| Feature | Description |
|---------|-------------|
| **Skill Library** | Manage a local library of skills with categories, search, sorting, drag-and-drop organization, batch delete, smoother startup/page re-entry through local derived snapshots, and adaptive large-list motion. |
| **My Prompts** | Store reusable prompts as synced `prompts/<category>/<name>/system.md` cards with import/export and one-click copy. |
| **GitHub Install** | Clone any repo, recursively discover nested skill candidates, and install selected ones into your library. |
| **Cross-tool Sync** | Push or pull skills across Claude Code, OpenCode, Codex, Gemini CLI, OpenClaw, and custom tools. |
| **Starred Repos** | Watch Git repos, browse their skills, and import or push them without installing everything into My Skills first. |
| **Cloud Backup** | Back up skills, prompts, and sync-safe metadata to object storage providers or Git, while keeping secrets and high-churn local runtime metadata on-device only. |
| **Update Detection** | Check GitHub-sourced skills for newer commits and update installed copies from the app, with per-skill spinner feedback and a top-of-page status banner for success or failure, including push-dir refresh and auto-push target overwrite to keep selected tools current. |
| **Desktop Experience** | Bilingual UI, multiple themes, helper-backed tray/menu bar reopen after window close, launch-at-login, per-tool settings, and background route-memory trimming when a hidden window stays inactive. |

For the complete UI/UX reference, see **[docs/features.md](docs/features.md)**.

## Supported Tools

Built-in adapters:

- **Claude Code**
- **OpenCode**
- **Codex**
- **Gemini CLI**
- **OpenClaw**

You can also add **custom tools** in Settings by pointing SkillFlow at local scan and push directories.

## Skill Format

A valid skill directory must contain a `skill.md` file at its root.

```text
my-skill/
  skill.md
  ...other files
```

## Cloud Backup

Configure backup in **Settings → Cloud Storage**.

- Supported providers: **Aliyun OSS**, **AWS S3**, **Azure Blob Storage**, **Google Cloud Storage**, **Tencent COS**, **Huawei OBS**, and **Git**.
- Skills, prompts, and synced metadata are backed up together so a new device can restore the same library state.
- Synced metadata stores portable relative paths where possible, which keeps cross-device restore working across macOS and Windows.
- High-churn per-skill check timestamps are kept in local-only `meta_local/*.local.json`, so multi-device backup syncs avoid unnecessary merge conflicts.
- High-churn starred-repo sync state (`lastSync`, `syncError`) is kept in local-only `star_repos_local.json` to reduce multi-device merge churn.
- Machine-specific paths, proxy settings, auto-push targets, launch-at-login state, window size, and sensitive cloud credentials stay local in `config_local.json`.
- Git backup supports startup pull, periodic auto-sync, and explicit conflict-resolution actions.

## Contributing & Building From Source

### Prerequisites

- macOS 11+ or Windows 10+
- Go 1.23+
- Node.js 18+
- Wails v2 CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Build Steps

```bash
git clone https://github.com/shinerio/SkillFlow
cd SkillFlow
make install-frontend
make dev
make test
make build
make build-cloud PROVIDERS="aws,google"
```

Notes:

- `make dev`, `make build`, and `make generate` run Wails from `cmd/skillflow/`.
- Production binaries are written under `cmd/skillflow/build/bin/`.

Common `make` targets:

| Target | Description |
|--------|-------------|
| `make dev` | Run Wails dev mode with frontend hot reload |
| `make build` | Build the production app with all cloud providers |
| `make build-cloud PROVIDERS="aws,google"` | Build with selected cloud providers only |
| `make test` | Run Go tests under `./core/...` |
| `make test-cloud PROVIDERS="aws,google"` | Run Go tests with selected cloud providers only |
| `make tidy` | Sync Go module dependencies |
| `make generate` | Regenerate Wails TypeScript bindings |
| `make clean` | Remove build artifacts |

For contributor-facing internals, see **[docs/architecture.md](docs/architecture.md)**.
