[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/shinerio/SkillFlow)

# SkillFlow

> 🌐 **中文** | [English](README.md)

理念：学于一瞬，用于万境。

SkillFlow 是一款跨平台的智能体SKILL管理神器。它打破了不同智能体环境的藩篱，让提示词与技能随处复用。集 GitHub 快速集成、多端同步、仓库发现、实时更新与云端归档于一体，在本地优先的逻辑下，为你重塑高效、纯净的开发体验。

![skilflow](docs/skillflow.gif)

## SkillFlow 解决什么问题

- **📦 资产内化**：解决“收藏即遗忘”的困境。将散落于博客、视频中的碎片化 Skill 转化为可管理、可迭代的个人资产，明确技能效用，建立淘汰与更新机制。
- **🎛️ 透明掌控**：拒绝技能管理的“黑盒”状态。对海量技能按场景分类归档，确保开发者对智能体能力边界拥有完全知情权与控制权。
- **🔄 多端同步**：打破模型孤岛。针对 Claude Code、Codex、Gemini 等多智能体在特性、额度及使用场景上的差异，实现技能配置在不同智能体间的无缝同步与复用。
- **🖥️ 环境一致性**：构建跨设备云同步体系。屏蔽 macOS / Windows 多设备异构环境差异，确保在任何办公场景和设备上都能获得高度一致的开发环境与技能配置。
- **⚡ 自动迭代**：摒弃繁琐的人工维护。建立技能自动更新机制，解决来源分散导致的版本滞后问题，确保技能库时刻处于最新状态。
- **📝 提示词工程**：实现 Prompt 的版本化管理。杜绝重复编写造成的实际效果参差不齐，推动提示词的持续优化与标准化迭代。
- **🛠️ 实战驱动**：拒绝低效的重复练习。以解决真实生产力痛点为导向，在掌握 AI 技术的同时打造真正赋能开发的实用工具，实现“学以致用”的闭环。

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
| **我的提示词** | 将可复用提示词保存为同步提示词卡片，支持范围导出、冲突可选导入、一键复制、关联图片预览和 markdown 风格网页链接。 |
| **GitHub 安装** | 克隆任意仓库，递归发现嵌套 Skill，并选择性安装到本地库。 |
| **跨智能体同步** | 在 Claude Code、OpenCode、Codex、Gemini CLI、OpenClaw 和自定义智能体之间推送或拉取 Skills。 |
| **仓库收藏** | 关注 Git 仓库后可直接浏览其中的 Skills，并按需导入或直接推送到智能体。 |
| **云端备份** | 将 Skills、提示词和可同步元数据备份到对象存储或 Git，同时把敏感信息与高频本地运行元数据保留在当前设备。 |
| **更新检测** | 检查 GitHub 来源 Skill 是否有新提交，并在应用内更新已安装副本以及已经推送到智能体目录的对应副本；同时会覆盖同步当前勾选“自动推送目标智能体”的副本，保持选中智能体中的 Skill 为最新版本。 |
| **桌面体验** | 支持中英文界面、多主题且默认使用 Young、helper 承载的托盘/菜单栏重开、开机自启、按智能体配置，以及窗口隐藏且长时间不活跃时的后台页面内存收缩。 |

完整的按钮、对话框和交互说明，以及云端备份的范围与服务商细节，请查阅 **[docs/features_zh.md](docs/features_zh.md)**。

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

参与贡献和本地构建说明请查阅 **[docs/contributing_zh.md](contributing_zh.md)**。
