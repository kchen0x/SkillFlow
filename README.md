# SkillFlow

> 🌐 [中文](README_zh.md) | **English**

A cross-platform desktop app for managing LLM Skills (prompt libraries / slash commands) across multiple AI coding tools — with GitHub install, cloud backup, and cross-tool sync.

## Download & Install

Get the latest release from **[GitHub Releases →](https://github.com/shinerio/SkillFlow/releases/latest)**

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `SkillFlow-macos-apple-silicon` |
| macOS (Intel) | `SkillFlow-macos-intel` |
| Windows (x64) | `SkillFlow-windows-amd64.exe` |

## Features

| Feature | Description |
|---------|-------------|
| **Skill Library** | Central store with categories, real-time search, drag-and-drop organization, batch delete, and safe empty-category deletion |
| **GitHub Install** | Clone any repo, browse skill candidates, select and install with one click; auto-pulls on subsequent scans |
| **Cross-tool Sync** | Push or pull skills to/from Claude Code, OpenCode, Codex, Gemini CLI, OpenClaw, or any custom tool; conflict handling per skill |
| **Starred Repos** | Watch Git repos and browse/import their skills without adding them to your library first |
| **Cloud Backup** | Mirror your library to Aliyun OSS, Tencent COS, Huawei OBS, or any Git repo |
| **Update Checker** | Detects new commits for GitHub-sourced skills; one-click update |
| **App Auto-Update** | Modal dialog notifies when a new version is available; Windows supports one-click download and restart; macOS links to GitHub Releases; users can skip a version to suppress future startup prompts |
| **Background Tray** | Clicking the window close button keeps the app running in background; on macOS it hides the Dock icon and leaves a monochrome menu-bar status icon, on Windows it stays in the notification area |
| **Settings** | Per-tool enable/disable, push & scan paths, custom tools, proxy configuration, and local-only path settings kept out of sync; folder pickers reopen at the current location |
| **Light / Dark Theme** | One-click toggle between deep-space Dark theme (cyan accent) and sky-blue Light theme; persisted across restarts |

For a complete description of every button, dialog, and interaction, see **[docs/features.md](docs/features.md)**.

## Supported Tools

Built-in adapters for: **Claude Code** · **OpenCode** · **Codex** · **Gemini CLI** · **OpenClaw**

Custom tools can be added in Settings with any local directory path.

## Skill Format

A valid skill directory must contain a `skill.md` file at its root. Any directory satisfying this requirement can be imported locally or via GitHub.

```
my-skill/
  skill.md     ← required
  ...other files
```

## Cloud Backup

Configure in **Settings → Cloud Storage**.

- Sync-safe settings and metadata live under the app data directory and use relative paths for cross-platform restore.
- Local-only filesystem paths (such as `SkillsStorageDir` and tool scan/push directories) live in `config_local.json` and are excluded from backup/sync.
- App data directory:
  - macOS: `~/Library/Application Support/SkillFlow/`
  - Windows: `%USERPROFILE%\.skillflow\`

Supported providers and required fields:

| Provider | Required Fields |
|----------|----------------|
| Aliyun OSS | Access Key ID, Access Key Secret, Endpoint |
| Tencent COS | SecretId, SecretKey, Region |
| Huawei OBS | Access Key, Secret Key, Endpoint |
| Git Repo | Repo URL, Branch, Username, Access Token |

## Contributing & Building from Source

### Prerequisites

- macOS 11+ or Windows 10+
- Go 1.23+
- Node.js 18+
- Wails v2 CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Build Steps

```bash
git clone https://github.com/shinerio/SkillFlow
cd SkillFlow
make install-frontend   # install frontend dependencies
make dev                # hot-reload dev mode
make test               # run Go tests
make build              # production binary → build/bin/
```

Common `make` targets:

| Target | Description |
|--------|-------------|
| `make dev` | Hot-reload dev mode (Go + frontend) |
| `make build` | Build production binary |
| `make test` | Run all Go tests |
| `make tidy` | Sync Go module dependencies |
| `make generate` | Regenerate TypeScript bindings after App method changes |
| `make clean` | Remove build artifacts |

For internal architecture details, see **[docs/architecture.md](docs/architecture.md)**.
