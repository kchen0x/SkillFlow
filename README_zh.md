[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/shinerio/SkillFlow)

# SkillFlow

> 🌐 **中文** | [English](README.md)

理念：学于一瞬，用于万境。

SkillFlow 是一款跨平台桌面应用，用来在不同智能体环境之间管理可复用的 Skill 和提示词。它提供本地 Skill 库、多智能体同步、仓库来源跟踪、更新检测，以及面向对象存储或 Git 的同步安全备份。

![SkillFlow](docs/skillflow.gif)

## SkillFlow 能做什么

- **📦 资产内化**：解决“收藏即遗忘”的困境。将散落于博客、视频中的碎片化 Skill 转化为可管理、可迭代的个人资产，明确技能效用，建立淘汰与更新机制。
- **🎛️ 透明掌控**：拒绝技能管理的“黑盒”状态。对海量技能按场景分类归档，确保开发者对智能体能力边界拥有完全知情权与控制权。
- **🔄 多端同步**：打破模型孤岛。针对 Claude Code、Codex、Gemini 等多智能体在特性、额度及使用场景上的差异，实现技能配置在不同智能体间的无缝同步与复用。
- **🖥️ 环境一致性**：构建跨设备云同步体系。屏蔽 macOS / Windows 多设备异构环境差异，确保在任何办公场景和设备上都能获得高度一致的开发环境与技能配置。
- **⚡ 自动迭代**：摒弃繁琐的人工维护。建立技能自动更新机制，解决来源分散导致的版本滞后问题，确保技能库时刻处于最新状态。
- **📝 提示词工程**：实现 Prompt 的版本化管理。杜绝重复编写造成的实际效果参差不齐，推动提示词的持续优化与标准化迭代。

## 下载安装

从 **[GitHub Releases](https://github.com/shinerio/SkillFlow/releases/latest)** 获取最新版本。

| 平台 | 文件 |
|------|------|
| macOS（Apple Silicon） | `SkillFlow-macos-apple-silicon.dmg` |
| macOS（Intel） | `SkillFlow-macos-intel.dmg` |
| Windows（x64） | `SkillFlow-windows.exe` |

## 功能概览

| 功能 | 说明 |
|------|------|
| **Skill 库** | 本地 Skill 库，支持分类、搜索、排序、拖拽整理、批量删除、更新检测和自动推送目标。 |
| **我的提示词** | 同步提示词卡片，支持分类管理、范围导出、冲突可选导入、一键复制、图片预览和网页链接。 |
| **跨智能体同步** | 在内置智能体和自定义智能体之间推送或拉取 Skill。 |
| **仓库收藏** | 跟踪 Git 仓库、浏览发现的 Skill，并在不整仓安装的前提下按需导入或推送。 |
| **云端备份** | 将可同步的 Skill、提示词和元数据备份到对象存储服务商或 Git。 |
| **桌面体验** | 支持中英文界面、多主题、托盘/菜单栏重开、开机自启、代理测试和应用内更新检查。 |

详细参考：

- UI / UX 行为： [docs/features_zh.md](docs/features_zh.md)
- 落盘文件与配置结构： [docs/config_zh.md](docs/config_zh.md)

## 支持的智能体

内置适配器：

- **Claude Code**
- **OpenCode**
- **Codex**
- **Gemini CLI**
- **OpenClaw**

你也可以在设置页通过配置本地扫描目录和推送目录来添加 **自定义智能体**。

## Skill 格式

有效的 Skill 目录必须在根目录包含 `skill.md` 文件；文件名匹配大小写不敏感。

```text
my-skill/
  skill.md
  ...其他文件
```

参与贡献和本地构建说明请查阅 **[contributing_zh.md](contributing_zh.md)**。
