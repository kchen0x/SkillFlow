# SkillFlow Configuration File Reference

> 🌐 [中文版](config_zh.md) | **English**

This document explains the on-disk format of SkillFlow's persisted configuration and metadata files.

Examples below use placeholders such as `<AppDataDir>` and `<SyncRoot>`:

- `<AppDataDir>`: the app data directory returned by `config.AppDataDir()`
- `<SyncRoot>`: the root that contains `skills/` and `meta/`; by default it is the same as `<AppDataDir>`, but it can move when `skillsStorageDir` is relocated outside the default app data directory

The actual starred-repository file name is `star_repos.json` (plural), not `star_repo.json`.

Even when `<SyncRoot>` moves, `config.json`, `config_local.json`, and `star_repos.json` remain under `<AppDataDir>`. Only `skills/` and `meta/` move together.

## Quick Summary

| File | Purpose | Synced |
|------|---------|--------|
| `config.json` | Shared, sync-safe settings | Yes |
| `config_local.json` | Machine-specific paths, secrets, and local runtime state | No |
| `star_repos.json` | Starred repository cache metadata | Yes |
| `star_repos_local.json` | Local-only starred-repo runtime sync state overlay | No |
| `meta/<skill-id>.json` | One sidecar metadata file per installed skill | Yes |
| `meta_local/<skill-id>.local.json` | Local-only per-skill volatile metadata overlay | No |
| `cache/viewstate/*.json` | Local derived UI/cache snapshots | No |

## `cache/viewstate/*.json`

Path: `<AppDataDir>/cache/viewstate/*.json`

These files store local-only derived state used to speed up page entry and reduce repeated directory scans. Typical payloads include installed-skill card snapshots and tool-presence indexes.

Rules:

- They are optimization artifacts, not source-of-truth records.
- They must be rebuilt from `skills/`, `meta/`, tool directories, and other existing truth-layer files.
- They must not be uploaded by cloud backup or written back into synced metadata files.
- Cross-device cache differences are expected and harmless.

## `config.json`

Path: `<AppDataDir>/config.json`

`config.json` stores settings that are safe to move across devices. It must not contain machine-specific absolute paths or sensitive credentials.

### Example

```json
{
  "defaultCategory": "Default",
  "logLevel": "info",
  "repoScanMaxDepth": 5,
  "skillStatusVisibility": {
    "mySkills": ["updatable", "pushedTools"],
    "myTools": ["imported", "updatable", "pushedTools"],
    "pushToTool": ["pushedTools"],
    "pullFromTool": ["imported"],
    "starredRepos": ["imported", "pushedTools"],
    "githubInstall": ["imported", "updatable", "pushedTools"]
  },
  "tools": [
    { "name": "claude-code", "enabled": true },
    { "name": "codex", "enabled": true },
    { "name": "gemini-cli", "enabled": false }
  ],
  "cloud": {
    "provider": "git",
    "enabled": true,
    "syncIntervalMinutes": 30
  },
  "cloudProfiles": {
    "git": {
      "bucketName": "",
      "remotePath": "team-a/backup/skillflow/",
      "credentials": {
        "repo_url": "https://github.com/example/skillflow-backup.git",
        "branch": "main",
        "username": "alice"
      }
    },
    "aws": {
      "bucketName": "skillflow-backup",
      "remotePath": "nightly/skillflow/",
      "credentials": {
        "region": "us-east-1"
      }
    }
  },
  "skippedUpdateVersion": "v1.2.3"
}
```

### Keys

| Key | Type | Meaning |
|-----|------|---------|
| `defaultCategory` | string | Default category used when importing or creating skills. |
| `logLevel` | string | Backend log level. Valid values are `debug`, `info`, and `error`. Invalid values are normalized to `error`. |
| `repoScanMaxDepth` | number | Maximum recursive depth used when scanning tool directories and repos. Values are normalized to the `1-20` range, with `5` as the default. |
| `skillStatusVisibility` | object | Per-page visibility policy for skill status badges. |
| `skillStatusVisibility.mySkills` | string[] | Status badges shown on the "My Skills" page. Allowed values there are `updatable` and `pushedTools`. |
| `skillStatusVisibility.myTools` | string[] | Status badges shown on the "My Tools" page. Allowed values there are `imported`, `updatable`, and `pushedTools`. |
| `skillStatusVisibility.pushToTool` | string[] | Status badges shown on the "Push to Tool" page. Allowed value there is `pushedTools`. |
| `skillStatusVisibility.pullFromTool` | string[] | Status badges shown on the "Pull from Tool" page. Allowed value there is `imported`. |
| `skillStatusVisibility.starredRepos` | string[] | Status badges shown on the starred-repos page. Allowed values there are `imported` and `pushedTools`. |
| `skillStatusVisibility.githubInstall` | string[] | Status badges shown in the GitHub install flow. Allowed values there are `imported`, `updatable`, and `pushedTools`. |
| `tools` | object[] | Built-in tool enable/disable state only. Path-related tool settings are stored in `config_local.json`. |
| `tools[].name` | string | Built-in tool name such as `claude-code`, `codex`, `gemini-cli`, `opencode`, or `openclaw`. |
| `tools[].enabled` | boolean | Whether this built-in tool is enabled in the UI and scanning/push flows. |
| `cloud` | object | Active cloud-backup selection and scheduling state. |
| `cloud.provider` | string | Active provider name, such as `git`, `aws`, `aliyun`, `azure`, `google`, `huawei`, or `tencent`. |
| `cloud.enabled` | boolean | Whether cloud backup is enabled. |
| `cloud.syncIntervalMinutes` | number | Automatic backup interval in minutes. `0` means "only back up on mutations". |
| `cloudProfiles` | object | Provider-specific sync-safe settings keyed by provider name. |
| `cloudProfiles.<provider>.bucketName` | string | Bucket/container name for object-storage providers. Usually empty for `git`. |
| `cloudProfiles.<provider>.remotePath` | string | Remote prefix under the provider root. It is normalized to always end in `skillflow/`. |
| `cloudProfiles.<provider>.credentials` | object | Provider settings that are safe to sync, such as endpoints or repo URLs. Secrets are split into `config_local.json`. |
| `skippedUpdateVersion` | string | App version tag the user chose to skip in the startup update prompt. |

### Sync-safe cloud credential keys

Only the following credential keys are persisted in `config.json`:

| Provider | Keys stored in `config.json` | Meaning |
|----------|------------------------------|---------|
| `aliyun` | `endpoint` | OSS endpoint host. |
| `aws` | `region` | AWS S3 region. |
| `azure` | `account_name`, `service_url` | Azure Storage account identity and service endpoint. |
| `git` | `repo_url`, `branch`, `username` | Remote Git repo address, branch, and optional HTTPS username. |
| `google` | none | Google credentials are always local-only. |
| `huawei` | `endpoint` | OBS endpoint host. |
| `tencent` | `endpoint` | COS endpoint host. |

## `config_local.json`

Path: `<AppDataDir>/config_local.json`

`config_local.json` stores machine-specific paths, local runtime state, and secrets. It is intentionally excluded from cloud backup and Git sync.

### Example

```json
{
  "skillsStorageDir": "/Users/demo/Library/Application Support/SkillFlow/skills",
  "autoPushTools": ["codex", "gemini-cli"],
  "launchAtLogin": true,
  "tools": [
    {
      "name": "claude-code",
      "scanDirs": [
        "/Users/demo/.claude/skills",
        "/Users/demo/.claude/plugins/marketplaces"
      ],
      "pushDir": "/Users/demo/.claude/skills",
      "custom": false,
      "enabled": true
    },
    {
      "name": "my-custom-tool",
      "scanDirs": ["/Users/demo/work/my-tool/skills"],
      "pushDir": "/Users/demo/work/my-tool/skills",
      "custom": true,
      "enabled": true
    }
  ],
  "cloudCredentialsByProvider": {
    "git": {
      "token": "ghp_xxx"
    },
    "aws": {
      "access_key_id": "AKIA...",
      "secret_access_key": "secret"
    },
    "google": {
      "service_account_json": "{\"type\":\"service_account\",\"project_id\":\"demo\"}"
    }
  },
  "proxy": {
    "mode": "manual",
    "url": "http://127.0.0.1:7890"
  },
  "window": {
    "width": 1440,
    "height": 920
  }
}
```

### Keys

| Key | Type | Meaning |
|-----|------|---------|
| `skillsStorageDir` | string | Absolute local path of the installed `skills/` directory. |
| `autoPushTools` | string[] | Tool names that should receive auto-push after import/update flows. Values are trimmed and deduplicated. |
| `launchAtLogin` | boolean | Whether SkillFlow should register itself as a login/startup item on the current machine. |
| `tools` | object[] | Tool path configuration. This includes built-in tools and all custom tools. |
| `tools[].name` | string | Tool identifier. |
| `tools[].scanDirs` | string[] | Local directories scanned for external skills from this tool. |
| `tools[].pushDir` | string | Local target directory used when pushing skills to this tool. |
| `tools[].custom` | boolean | `true` for user-created custom tools, `false` for built-in tools. |
| `tools[].enabled` | boolean | Stored for every tool, but only meaningful for custom tools in `config_local.json`; built-in enable state comes from `config.json`. |
| `cloudCredentialsByProvider` | object | Sensitive provider credentials keyed by provider name. |
| `cloudCredentialsByProvider.<provider>` | object | Local-only credential map for one provider. |
| `proxy` | object | Local proxy settings used for outbound HTTP requests. |
| `proxy.mode` | string | Proxy mode: `none`, `system`, or `manual`. |
| `proxy.url` | string | Manual proxy URL. Used only when `mode` is `manual`. |
| `window` | object | Last persisted window size. This key is omitted until a valid size has been saved. |
| `window.width` | number | Window width in pixels. |
| `window.height` | number | Window height in pixels. |

### Local-only cloud credential keys

These keys are intentionally kept out of `config.json`:

| Provider | Keys stored in `config_local.json` | Meaning |
|----------|------------------------------------|---------|
| `aliyun` | `access_key_id`, `access_key_secret` | OSS access key pair. |
| `aws` | `access_key_id`, `secret_access_key` | AWS access key pair. |
| `azure` | `account_key` | Azure Storage account secret. |
| `git` | `token` | HTTPS access token. |
| `google` | `service_account_json` | Inline service-account JSON or a local key-file path. |
| `huawei` | `access_key_id`, `secret_access_key` | OBS access key pair. |
| `tencent` | `secret_id`, `secret_key` | COS credential pair. |

## `star_repos.json`

Path: `<AppDataDir>/star_repos.json`

`star_repos.json` stores the local cache state of starred repositories.

### Example

```json
[
  {
    "url": "https://github.com/example/awesome-skills.git",
    "name": "example/awesome-skills",
    "source": "github.com/example/awesome-skills",
    "localDir": "cache/github.com/example/awesome-skills"
  }
]
```

### Keys

| Key | Type | Meaning |
|-----|------|---------|
| `url` | string | Original Git clone URL entered or seeded for the repository. |
| `name` | string | Human-friendly repo name, usually `<owner>/<repo>` or `<group>/<subgroup>/<repo>`. |
| `source` | string | Canonical repo source key used for matching across modules, usually `<host>/<repo-path>`. |
| `localDir` | string | Local clone/cache directory. When it lives inside `<AppDataDir>`, it is stored as a forward-slash relative path such as `cache/github.com/example/awesome-skills`. |

## `star_repos_local.json`

Path: `<AppDataDir>/star_repos_local.json`

This local-only overlay stores per-repo volatile sync state that should not be synced across devices.

### Example

```json
{
  "repos": {
    "github.com/example/awesome-skills": {
      "lastSync": "2026-03-11T08:15:00Z",
      "syncError": "authentication failed"
    }
  }
}
```

### Keys

| Key | Type | Meaning |
|-----|------|---------|
| `repos` | object | Map keyed by repo source key (or URL fallback) to local sync state entries. |
| `repos.<key>.lastSync` | string | Last successful sync timestamp on the current device (RFC3339). |
| `repos.<key>.syncError` | string | Last sync error message on the current device. Omitted when empty. |

## `meta/<skill-id>.json`

Path: `<SyncRoot>/meta/<skill-id>.json`

Each installed skill gets one JSON sidecar file named after `Skill.ID` rather than the skill name.

### Example

```json
{
  "ID": "0f4b6f23-4f1e-4c56-a1fa-7fa0f7ce1234",
  "Name": "code-review",
  "Path": "skills/Engineering/code-review",
  "Category": "Engineering",
  "Source": "github",
  "SourceURL": "https://github.com/example/skill-collection.git",
  "SourceSubPath": "engineering/code-review",
  "SourceSHA": "8f3d4c2",
  "LatestSHA": "31ad9be",
  "InstalledAt": "2026-03-10T09:30:00Z",
  "UpdatedAt": "2026-03-11T07:45:00Z"
}
```

### Keys

| Key | Type | Meaning |
|-----|------|---------|
| `ID` | string | Stable instance UUID for the installed skill. This value also becomes the metadata file name. |
| `Name` | string | Skill directory name. |
| `Path` | string | Local skill directory path. When the directory is under the synchronized root, it is stored as a forward-slash relative path such as `skills/Engineering/code-review`. |
| `Category` | string | SkillFlow category folder that currently contains the skill. |
| `Source` | string | Install source type. Current values are `github` and `manual`. |
| `SourceURL` | string | Original source repository or source location URL for git-backed installs. Usually empty for manual imports. |
| `SourceSubPath` | string | Relative path inside the source repo when the skill was imported from a subdirectory. |
| `SourceSHA` | string | Commit SHA recorded when the installed skill was last imported or updated. |
| `LatestSHA` | string | Latest remote SHA most recently discovered by the update checker. |
| `InstalledAt` | string | Timestamp when the skill was first imported into SkillFlow. |
| `UpdatedAt` | string | Timestamp when the skill metadata was last changed, such as category moves or updates. |

### Important note

`meta/<skill-id>.json` stores installation state, not the YAML frontmatter parsed from `SKILL.md`. Frontmatter fields such as `name`, `description`, and `allowed-tools` stay in the skill content itself.

## `meta_local/<skill-id>.local.json`

Path: `<SyncRoot>/meta_local/<skill-id>.local.json`

This file stores local-only, high-churn per-skill fields that should not be synced across devices.

### Example

```json
{
  "lastCheckedAt": "2026-03-11T08:00:00Z"
}
```

### Keys

| Key | Type | Meaning |
|-----|------|---------|
| `lastCheckedAt` | string | Timestamp of the most recent update-check attempt on the current device. |
