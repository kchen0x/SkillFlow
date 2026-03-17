[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/shinerio/SkillFlow)

# SkillFlow

> 🌐 [中文](README_zh.md) | **English**

Slogan: Master once. Apply everywhere.

SkillFlow is a cross-platform desktop app for managing skills and reusable prompts across diverse agentic AI environments. It combines GitHub install, cross-environment sync, starred repo browsing, update checking, and cloud backup in one local-first workflow.

![skilflow](docs/skillflow.gif)

## What Problems Does SkillFlow Solve?

- **📦 Turn discoveries into assets**: Solve the "saved means forgotten" problem by turning scattered skills from blogs and videos into personal assets you can manage, iterate on, retire, and update deliberately.
- **🎛️ Stay in control**: Avoid black-box skill management by organizing large skill collections by scenario, so developers always understand and control an agent's capability boundaries.
- **🔄 Sync across agents**: Break model silos by reusing one skill setup across Claude Code, Codex, Gemini, and other agents with different strengths, limits, and use cases.
- **🖥️ Keep environments consistent**: Build a cloud-backed workflow across macOS and Windows devices so your development environment and skill configuration stay aligned wherever you work.
- **⚡ Iterate automatically**: Replace tedious manual maintenance with automatic skill updates, so distributed sources do not turn into stale versions.
- **📝 Version prompts intentionally**: Manage prompts as versioned assets instead of rewriting them ad hoc, improving consistency, optimization, and standardization over time.
- **🛠️ Focus on real work**: Move beyond repetitive toy exercises and use AI practice to solve real productivity problems while building tools that matter in day-to-day development.

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
| **My Prompts** | Store reusable prompts as synced prompt cards with scoped export, conflict-aware import, one-click copy, related image previews, and markdown-style web links. |
| **GitHub Install** | Clone any repo, recursively discover nested skill candidates, and install selected ones into your library. |
| **Cross-agent Sync** | Push or pull skills across Claude Code, OpenCode, Codex, Gemini CLI, OpenClaw, and custom agents. |
| **Starred Repos** | Watch Git repos, browse their skills, and import or push them without installing everything into My Skills first. |
| **Cloud Backup** | Back up skills, prompts, and sync-safe metadata to object storage providers or Git, while keeping secrets and high-churn local runtime metadata on-device only. |
| **Update Detection** | Check GitHub-sourced skills for newer commits and update installed copies from the app, with per-skill spinner feedback and a top-of-page status banner for success or failure, including push-dir refresh and auto-push target overwrite to keep selected agents current. |
| **Desktop Experience** | Bilingual UI, multiple themes, helper-backed tray/menu bar reopen after window close, launch-at-login, per-agent settings, and background route-memory trimming when a hidden window stays inactive. |

For the complete UI/UX reference, including cloud backup behavior and provider details, see **[docs/features.md](docs/features.md)**.

## Supported Agents

Built-in adapters:

- **Claude Code**
- **OpenCode**
- **Codex**
- **Gemini CLI**
- **OpenClaw**

You can also add **custom agents** in Settings by pointing SkillFlow at local scan and push directories.

## Skill Format

A valid skill directory must contain a `skill.md` file at its root.

```text
my-skill/
  skill.md
  ...other files
```

For contributing guidelines and building from source, see **[docs/contributing.md](contributing.md)**.
