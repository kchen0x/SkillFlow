# SkillFlow

> 🌐 **中文** | [English](README.md)

SkillFlow 是一款跨平台桌面应用，用于统一管理多个 AI 编程工具中的 LLM Skills 与可复用提示词。它把 GitHub 安装、跨工具同步、仓库收藏浏览、更新检测和云端备份整合到同一套本地优先工作流里。

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
| **Skill 库** | 提供本地集中库，支持分类、搜索、排序、拖拽整理和批量删除。 |
| **我的提示词** | 将可复用提示词保存为同步的 `prompts/<category>/<name>/system.md` 卡片，支持导入导出和一键复制。 |
| **GitHub 安装** | 克隆任意仓库，递归发现嵌套 Skill，并选择性安装到本地库。 |
| **跨工具同步** | 在 Claude Code、OpenCode、Codex、Gemini CLI、OpenClaw 和自定义工具之间推送或拉取 Skills。 |
| **仓库收藏** | 关注 Git 仓库后可直接浏览其中的 Skills，并按需导入或直接推送到工具。 |
| **云端备份** | 将 Skills、提示词和元数据备份到对象存储或 Git，并把敏感信息保留在本地。 |
| **更新检测** | 检查 GitHub 来源 Skill 是否有新提交，并在应用内更新已安装副本。 |
| **桌面体验** | 支持中英文界面、多主题、托盘驻留、开机自启和按工具配置。 |

完整的按钮、对话框和交互说明请查阅 **[docs/features_zh.md](docs/features_zh.md)**。

## 支持的工具

内置适配器：

- **Claude Code**
- **OpenCode**
- **Codex**
- **Gemini CLI**
- **OpenClaw**

你也可以在设置页添加 **自定义工具**，指定本地扫描目录和推送目录即可接入。

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
- 机器相关路径、代理设置、自动推送目标、开机自启、窗口尺寸以及敏感云凭据都只保存在本地 `config_local.json`。
- Git 备份支持启动拉取、定时自动同步和显式冲突处理操作。

## 参与贡献与本地构建

### 环境要求

- macOS 11+ 或 Windows 10+
- Go 1.23+
- Node.js 18+
- Wails v2 CLI：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### 构建步骤

```bash
git clone https://github.com/shinerio/SkillFlow
cd SkillFlow
make install-frontend
make dev
make test
make build
make build-cloud PROVIDERS="aws,google"
```

说明：

- `make dev`、`make build` 和 `make generate` 都会在 `cmd/skillflow/` 下执行 Wails 命令。
- 生产构建输出位于 `cmd/skillflow/build/bin/`。

常用 `make` 目标：

| 目标 | 说明 |
|------|------|
| `make dev` | 启动 Wails 开发模式并联动前端热更新 |
| `make build` | 构建包含全部云服务商的生产版本 |
| `make build-cloud PROVIDERS="aws,google"` | 仅构建指定云服务商版本 |
| `make test` | 运行 `./core/...` 下的 Go 测试 |
| `make test-cloud PROVIDERS="aws,google"` | 使用指定云服务商标签运行 Go 测试 |
| `make tidy` | 同步 Go 模块依赖 |
| `make generate` | 重新生成 Wails TypeScript 绑定 |
| `make clean` | 清理构建产物 |

面向贡献者的内部说明请查阅 **[docs/architecture_zh.md](docs/architecture_zh.md)**。
