[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/shinerio/SkillFlow)

# SkillFlow

> 🌐 **中文** | [English](README.md)

理念：学于一瞬，用于万境。

SkillFlow 是一款跨平台的AI技能管理神器。它打破了不同智能体环境的藩篱，让提示词与技能随处复用。集 GitHub 快速集成、多端同步、仓库发现、实时更新与云端归档于一体，在本地优先的逻辑下，为你重塑高效、纯净的开发体验。

![skilflow](docs/skillflow.gif)

## 下载安装

从 **[GitHub Releases →](https://github.com/shinerio/SkillFlow/releases/latest)** 下载最新版本。

| 平台 | 文件 |
|------|------|
| macOS（Apple Silicon） | `SkillFlow-macos-apple-silicon` |
| macOS（Intel） | `SkillFlow-macos-intel` |
| Windows（x64） | `SkillFlow-windows-amd64.exe` |

## 功能概览

| 功能 | 说明 |
|------|------|
| **Skill 库** | 提供本地集中库，支持分类、搜索、排序、拖拽整理和批量删除，并通过本地派生快照、启动任务错峰和大列表动画降级提升启动与页面重入的流畅度。 |
| **我的提示词** | 将可复用提示词保存为同步提示词卡片，支持导入导出、一键复制、关联图片预览和 markdown 风格网页链接。 |
| **GitHub 安装** | 克隆任意仓库，递归发现嵌套 Skill，并选择性安装到本地库。 |
| **跨智能体同步** | 在 Claude Code、OpenCode、Codex、Gemini CLI、OpenClaw 和自定义智能体之间推送或拉取 Skills。 |
| **仓库收藏** | 关注 Git 仓库后可直接浏览其中的 Skills，并按需导入或直接推送到智能体。 |
| **云端备份** | 将 Skills、提示词和可同步元数据备份到对象存储或 Git，同时把敏感信息与高频本地运行元数据保留在当前设备。 |
| **更新检测** | 检查 GitHub 来源 Skill 是否有新提交，并在应用内更新已安装副本以及已经推送到智能体目录的对应副本；同时会覆盖同步当前勾选“自动推送目标智能体”的副本，保持选中智能体中的 Skill 为最新版本。 |
| **桌面体验** | 支持中英文界面、多主题、helper 承载的托盘/菜单栏重开、开机自启、按智能体配置，以及窗口隐藏且长时间不活跃时的后台页面内存收缩。 |

完整的按钮、对话框和交互说明请查阅 **[docs/features_zh.md](docs/features_zh.md)**。

## 支持的智能体

内置适配器：

- **Claude Code**
- **OpenCode**
- **Codex**
- **Gemini CLI**
- **OpenClaw**

你也可以在设置页添加 **自定义智能体**，指定本地扫描目录和推送目录即可接入。

## Skill 格式

一个有效的 Skill 目录必须在根目录包含 `skill.md` 文件。

```text
my-skill/
  skill.md
  ...其他文件
```

## 云端备份

在 **设置 → 云存储** 中配置备份。

- 支持的服务商：**阿里云 OSS**、**AWS S3**、**Azure Blob Storage**、**Google Cloud Storage**、**腾讯云 COS**、**华为云 OBS** 以及 **Git**。
- Skills、提示词和可同步元数据会一起备份，方便新设备恢复出相同的资料库状态。
- 可同步元数据会尽量保存为可移植的相对路径，从而保证 macOS 与 Windows 之间恢复行为一致。
- 高频变化的 Skill 检查时间会写入仅本地的 `meta_local/*.local.json`，不参与云端/Git 同步，以减少多机并行时的合并冲突。
- 收藏仓库的高频同步状态（`lastSync`、`syncError`）会写入仅本地的 `star_repos_local.json`，避免多机同步时反复冲突。
- 机器相关路径、代理设置、自动推送目标、开机自启、窗口尺寸以及敏感云凭据都只保存在本地 `config_local.json`。
- Git 备份支持启动拉取、定时自动同步和显式冲突处理操作。

参与贡献和本地构建说明请查阅 **[docs/contributing_zh.md](contributing_zh.md)**。
