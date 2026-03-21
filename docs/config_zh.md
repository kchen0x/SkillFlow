# SkillFlow 配置文件参考

> 🌐 **中文** | [English](config.md)

本文档说明 SkillFlow 持久化到磁盘的配置文件与元数据文件格式。

下文使用了 `<AppDataDir>` 与 `<SyncRoot>` 这类占位符：

- `<AppDataDir>`：`config.AppDataDir()` 返回的应用数据目录
- `<SyncRoot>`：包含 `skills/` 与 `meta/` 的同步根目录；默认等于 `<AppDataDir>`，但当 `skillsStorageDir` 被移到外部目录时会随之变化

即使 `<SyncRoot>` 发生变化，`config.json`、`config_local.json`、`star_repos.json` 仍然保留在 `<AppDataDir>` 下；只有 `skills/` 与 `meta/` 会一起移动。

## 快速概览

| 文件 | 作用 | 是否参与同步 |
|------|------|--------------|
| `config.json` | 共享的、可同步的安全配置 | 是 |
| `config_local.json` | 机器相关路径、敏感信息和本地运行状态 | 否 |
| `star_repos.json` | 收藏仓库的本地缓存元数据 | 是 |
| `star_repos_local.json` | 收藏仓库本地同步状态覆盖文件 | 否 |
| `prompts/<category>/<name>/prompt.json` | 提示词卡片元数据，如描述、关联图片和网页链接 | 是 |
| `meta/<skill-id>.json` | 每个已安装 Skill 的 sidecar 元数据 | 是 |
| `meta_local/<skill-id>.local.json` | 每个 Skill 的本地易变元数据覆盖文件 | 否 |
| `cache/viewstate/*.json` | 本地派生 UI / 缓存快照 | 否 |
| `runtime/*.json`、`runtime/helper.lock` | 本地 helper/UI 进程协调状态 | 否 |

## `cache/viewstate/*.json`

路径：`<AppDataDir>/cache/viewstate/*.json`

这些文件保存只属于当前设备的派生状态，用于加快页面进入速度并减少重复目录扫描。典型内容包括已安装 Skill 卡片快照和智能体 presence 索引。

规则：

- 它们只是性能优化产物，不是权威真值。
- 必须从 `skills/`、`meta/`、智能体目录等现有真值层文件重建。
- 不能被云备份上传，也不能反向写回任何可同步元数据文件。
- 不同设备上的缓存内容不同是正常且无害的。

## `runtime/*.json`、`runtime/helper.lock`

路径：`<AppDataDir>/runtime/`

这些文件保存 helper/UI 双进程在当前设备上的本地协调状态。典型文件包括 `helper-control.json`、`ui-control.json` 和 `helper.lock`。

规则：

- 文件里保存的是回环地址、随机 token 和 PID，只对当前设备、当前这次运行有效。
- 它们必须被排除在云备份、Git 同步和跨设备恢复之外。
- 如果旧版本曾把 `runtime/` 跟踪进 Git 备份，下一次推送时会自动清理掉。
- 在 SkillFlow 完全退出时可以删除；应用下次启动会自动重建。

## `config.json`

路径：`<AppDataDir>/config.json`

`config.json` 只保存可跨设备移动的安全配置，不应包含机器相关绝对路径或敏感凭据。

在真正加载配置之前，SkillFlow 会先执行 `core/platform/upgrade` 里的单次术语割接。旧版基于 `tools` 的键会被原地改写为新版基于 `agents` 的 schema，运行时代码只读取新 schema。

### 示例

```json
{
  "defaultCategory": "Default",
  "logLevel": "info",
  "repoScanMaxDepth": 5,
  "skillStatusVisibility": {
    "mySkills": ["updatable", "pushedAgents"],
    "myAgents": ["imported", "updatable", "pushedAgents"],
    "pushToAgent": ["pushedAgents"],
    "pullFromAgent": ["imported"],
    "starredRepos": ["imported", "pushedAgents"]
  },
  "agents": [
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

### 键说明

| 键 | 类型 | 作用 |
|----|------|------|
| `defaultCategory` | string | 导入或创建 Skill 时默认使用的分类。 |
| `logLevel` | string | 后端日志级别。可选值为 `debug`、`info`、`error`；非法值会被归一化为 `error`。 |
| `repoScanMaxDepth` | number | 扫描智能体目录与仓库时允许的最大递归深度。会被归一化到 `1-20` 范围内，默认值是 `5`。 |
| `skillStatusVisibility` | object | 不同页面上 Skill 状态徽标的显示策略。 |
| `skillStatusVisibility.mySkills` | string[] | “我的 Skills” 页面显示哪些状态徽标。该页允许的值是 `updatable`、`pushedAgents`。 |
| `skillStatusVisibility.myAgents` | string[] | “我的智能体” 页面显示哪些状态徽标。该页允许的值是 `imported`、`updatable`、`pushedAgents`。 |
| `skillStatusVisibility.pushToAgent` | string[] | “推送到智能体” 页面显示哪些状态徽标。该页允许的值是 `pushedAgents`。 |
| `skillStatusVisibility.pullFromAgent` | string[] | “从智能体拉取” 页面显示哪些状态徽标。该页允许的值是 `imported`。 |
| `skillStatusVisibility.starredRepos` | string[] | 收藏仓库页面显示哪些状态徽标。该页允许的值是 `imported`、`pushedAgents`。 |
| `agents` | object[] | 只保存内置智能体的启用/停用状态。路径相关设置保存在 `config_local.json`。 |
| `agents[].name` | string | 内置智能体名，例如 `claude-code`、`codex`、`gemini-cli`、`opencode`、`openclaw`。 |
| `agents[].enabled` | boolean | 该内置智能体是否在界面与扫描/推送流程中启用。 |
| `cloud` | object | 当前选中的云备份提供方及调度状态。 |
| `cloud.provider` | string | 当前激活的 provider 名称，例如 `git`、`aws`、`aliyun`、`azure`、`google`、`huawei`、`tencent`。 |
| `cloud.enabled` | boolean | 是否启用云备份。 |
| `cloud.syncIntervalMinutes` | number | 自动备份间隔（分钟）。`0` 表示“仅在状态变更后备份”。 |
| `cloudProfiles` | object | 以 provider 名称为键的 provider 级同步安全配置。 |
| `cloudProfiles.<provider>.bucketName` | string | 对象存储 provider 使用的 bucket/container 名称；`git` 一般为空。 |
| `cloudProfiles.<provider>.remotePath` | string | provider 根目录下的远端前缀。保存时会被规范化为以 `skillflow/` 结尾。 |
| `cloudProfiles.<provider>.credentials` | object | 允许同步的 provider 配置项，例如 endpoint 或 repo URL。敏感项会拆分到 `config_local.json`。 |
| `skippedUpdateVersion` | string | 用户在启动更新提示里选择“跳过此版本”后记录的应用版本号。 |

### 会写入 `config.json` 的云配置键

只有下面这些凭据键会出现在 `config.json` 中：

| Provider | 保存在 `config.json` 的键 | 作用 |
|----------|---------------------------|------|
| `aliyun` | `endpoint` | OSS 的 endpoint 主机名。 |
| `aws` | `region` | AWS S3 区域。 |
| `azure` | `account_name`, `service_url` | Azure Storage 账号标识与服务地址。 |
| `git` | `repo_url`, `branch`, `username` | 远端 Git 仓库地址、分支、可选的 HTTPS 用户名。 |
| `google` | 无 | Google 凭据始终只保存在本地。 |
| `huawei` | `endpoint` | OBS 的 endpoint 主机名。 |
| `tencent` | `endpoint` | COS 的 endpoint 主机名。 |

## `config_local.json`

路径：`<AppDataDir>/config_local.json`

`config_local.json` 保存机器相关路径、本地运行状态与敏感凭据。它会被明确排除在云备份和 Git 同步之外。

同一套启动割接逻辑也会在读取 `config_local.json` 之前，把旧版 `autoPushTools` / `tools` 改写成新版 `autoPushAgents` / `agents`。

### 示例

```json
{
  "skillsStorageDir": "/Users/demo/Library/Application Support/SkillFlow/skills",
  "autoUpdateSkills": true,
  "autoPushAgents": ["codex", "gemini-cli"],
  "launchAtLogin": true,
  "agents": [
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
      "name": "my-custom-agent",
      "scanDirs": ["/Users/demo/work/my-agent/skills"],
      "pushDir": "/Users/demo/work/my-agent/skills",
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

### 键说明

| 键 | 类型 | 作用 |
|----|------|------|
| `skillsStorageDir` | string | 已安装 `skills/` 目录的本机绝对路径。 |
| `autoUpdateSkills` | boolean | 当前设备在刷新收藏仓库后，是否应自动把匹配的已安装 Git Skill 更新到 **我的skills**。 |
| `autoPushAgents` | string[] | 在导入/更新后自动推送到哪些智能体。保存前会去空格并去重。 |
| `launchAtLogin` | boolean | 当前设备上是否把 SkillFlow 注册为开机/登录自启动项。 |
| `agents` | object[] | 智能体路径配置，既包含内置智能体，也包含自定义智能体。 |
| `agents[].name` | string | 智能体标识名。 |
| `agents[].scanDirs` | string[] | 扫描该智能体外部 Skill 的本地目录列表。 |
| `agents[].pushDir` | string | 将 Skill 推送到该智能体时使用的目标目录。 |
| `agents[].custom` | boolean | `true` 表示用户创建的自定义智能体，`false` 表示内置智能体。 |
| `agents[].enabled` | boolean | 每个智能体都会带上这个字段，但在 `config_local.json` 中它只对自定义智能体有实际意义；内置智能体的启用状态以 `config.json` 为准。 |
| `cloudCredentialsByProvider` | object | 以 provider 名称为键的敏感凭据集合。 |
| `cloudCredentialsByProvider.<provider>` | object | 某个 provider 的本地专属凭据键值表。 |
| `proxy` | object | 出站 HTTP 请求使用的本地代理设置。 |
| `proxy.mode` | string | 代理模式：`none`、`system`、`manual`。 |
| `proxy.url` | string | 手动代理 URL。仅在 `mode` 为 `manual` 时使用。 |
| `window` | object | 最近一次保存的窗口尺寸。只有保存过有效尺寸后才会出现。 |
| `window.width` | number | 窗口宽度（像素）。 |
| `window.height` | number | 窗口高度（像素）。 |

### 只保存在 `config_local.json` 的云凭据键

下面这些键不会写入 `config.json`：

| Provider | 保存在 `config_local.json` 的键 | 作用 |
|----------|---------------------------------|------|
| `aliyun` | `access_key_id`, `access_key_secret` | OSS 访问密钥对。 |
| `aws` | `access_key_id`, `secret_access_key` | AWS 访问密钥对。 |
| `azure` | `account_key` | Azure Storage 账号密钥。 |
| `git` | `token` | HTTPS 访问令牌。 |
| `google` | `service_account_json` | 内联 service-account JSON，或本地密钥文件路径。 |
| `huawei` | `access_key_id`, `secret_access_key` | OBS 访问密钥对。 |
| `tencent` | `secret_id`, `secret_key` | COS 凭据对。 |

## `star_repos.json`

路径：`<AppDataDir>/star_repos.json`

`star_repos.json` 保存收藏仓库在本地的缓存状态。

### 示例

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

### 键说明

| 键 | 类型 | 作用 |
|----|------|------|
| `url` | string | 用户输入或系统内置的原始 Git 克隆地址。 |
| `name` | string | 便于展示的仓库名，通常是 `<owner>/<repo>` 或 `<group>/<subgroup>/<repo>`。 |
| `source` | string | 跨模块关联使用的规范化仓库源键，通常形如 `<host>/<repo-path>`。 |
| `localDir` | string | 本地克隆/缓存目录。如果目录位于 `<AppDataDir>` 内，会以正斜杠相对路径保存，例如 `cache/github.com/example/awesome-skills`。 |

## `star_repos_local.json`

路径：`<AppDataDir>/star_repos_local.json`

这个仅本地文件用于保存每个收藏仓库变化频繁、且不应跨设备同步的同步状态。

### 示例

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

### 字段说明

| 键 | 类型 | 说明 |
|----|------|------|
| `repos` | object | 以仓库 source 键（或 URL 兜底）为 key 的本地同步状态映射。 |
| `repos.<key>.lastSync` | string | 当前设备最近一次成功同步时间（RFC3339）。 |
| `repos.<key>.syncError` | string | 当前设备最近一次同步失败错误信息；为空时省略。 |

## `prompts/<category>/<name>/prompt.json`

路径：`<SyncRoot>/prompts/<category>/<name>/prompt.json`

每个提示词的正文保存在同级的 `system.md` 中，提示词卡片元数据则保存在 `prompt.json`。

### 示例

```json
{
  "name": "Review API",
  "description": "Review backend changes",
  "imageURLs": [
    "https://cdn.example.com/review-1.png",
    "https://cdn.example.com/review-2.png"
  ],
  "webLinks": [
    {
      "label": "PRD",
      "url": "https://docs.example.com/prd"
    },
    {
      "label": "Preview",
      "url": "https://preview.example.com/review"
    }
  ],
  "createdAt": "2026-03-15T13:00:00Z",
  "updatedAt": "2026-03-15T13:05:00Z"
}
```

### 字段说明

| 键 | 类型 | 说明 |
|----|------|------|
| `name` | string | 提示词名称。通常与提示词目录名一致，并在整个提示词库内保持全局唯一。 |
| `description` | string | 可选的简短描述，会显示在提示词卡片上。 |
| `imageURLs` | string[] | 可选的关联图片 URL。当前最多支持 3 条，且必须是 `http` 或 `https` URL。 |
| `webLinks` | object[] | 可选的结构化网页链接，会渲染为提示词卡片上的可点击胶囊。 |
| `webLinks[].label` | string | 链接展示文本。 |
| `webLinks[].url` | string | 点击卡片胶囊后打开的外部 URL。持久化时只允许 `http` 和 `https`。 |
| `createdAt` | string | 提示词创建时间（RFC3339）。 |
| `updatedAt` | string | 最近一次元数据更新时间（RFC3339）。 |

## `meta/<skill-id>.json`

路径：`<SyncRoot>/meta/<skill-id>.json`

每个已安装 Skill 都会有一个 sidecar JSON，文件名使用 `Skill.ID`，而不是 Skill 名称。

### 示例

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

### 键说明

| 键 | 类型 | 作用 |
|----|------|------|
| `ID` | string | 已安装 Skill 的稳定实例 UUID，同时也是元数据文件名。 |
| `Name` | string | Skill 目录名。 |
| `Path` | string | Skill 的本地目录路径。如果位于同步根目录内，会以正斜杠相对路径保存，例如 `skills/Engineering/code-review`。 |
| `Category` | string | 当前 Skill 所在的 SkillFlow 分类目录。 |
| `Source` | string | 安装来源类型。当前取值有 `github` 和 `manual`。 |
| `SourceURL` | string | Git 安装来源的原始仓库或来源地址；手动导入通常为空。 |
| `SourceSubPath` | string | 如果 Skill 来自仓库子目录，这里保存其仓库内相对路径。 |
| `SourceSHA` | string | 该 Skill 最近一次导入或更新时记录的提交 SHA。 |
| `LatestSHA` | string | 更新检测最近一次发现的远端最新 SHA。 |
| `InstalledAt` | string | Skill 首次导入到 SkillFlow 的时间。 |
| `UpdatedAt` | string | 最近一次修改元数据的时间，例如移动分类或完成更新。 |

### 重要说明

`meta/<skill-id>.json` 保存的是安装状态，不是 `SKILL.md` 里的 YAML frontmatter。像 `name`、`description`、`allowed-agents` 这类 frontmatter 字段仍然保存在 Skill 内容本身。

## `meta_local/<skill-id>.local.json`

路径：`<SyncRoot>/meta_local/<skill-id>.local.json`

这个文件用于保存当前设备本地、变化频繁且不应跨设备同步的 Skill 字段。

### 示例

```json
{
  "lastCheckedAt": "2026-03-11T08:00:00Z"
}
```

### 字段说明

| 键 | 类型 | 说明 |
|----|------|------|
| `lastCheckedAt` | string | 当前设备最近一次执行更新检查的时间。 |
