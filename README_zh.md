# SkillFlow

> 🌐 **中文** | [English](README.md)

一款跨平台桌面应用，用于统一管理多个 AI 编程工具中的 LLM Skills（提示词库 / 斜杠命令），支持 GitHub 安装、云端备份和跨工具同步。

## 下载安装

从 **[GitHub Releases →](https://github.com/shinerio/SkillFlow/releases/latest)** 下载最新版本

| 平台 | 文件 |
|------|------|
| macOS（Apple Silicon） | `SkillFlow-macos-apple-silicon` |
| macOS（Intel） | `SkillFlow-macos-intel` |
| Windows（x64） | `SkillFlow-windows-amd64.exe` |

## 功能概览

| 功能 | 说明 |
|------|------|
| **Skill 库** | 集中存储所有 Skills，支持分类管理、实时搜索、拖拽整理和批量删除 |
| **GitHub 安装** | 克隆任意仓库，浏览 Skill 候选项，一键选择安装；后续扫描自动拉取更新 |
| **跨工具同步** | 推送或拉取 Skills 到/从 Claude Code、OpenCode、Codex、Gemini CLI、OpenClaw 及自定义工具；逐条冲突处理 |
| **仓库收藏** | 关注 Git 仓库，无需导入即可浏览和使用其中的 Skills |
| **云端备份** | 将 Skill 库镜像至阿里云 OSS、腾讯云 COS、华为云 OBS 或任意 Git 仓库 |
| **更新检测** | 自动检测 GitHub 来源 Skills 的新提交；一键更新 |
| **应用自动更新** | 弹出模态对话框提示新版本；Windows 支持一键下载并重启；macOS 链接至 GitHub Releases 页面；用户可跳过当前版本以抑制后续启动弹窗 |
| **托盘驻留** | 点击窗口关闭按钮仅隐藏窗口；macOS 驻留顶部状态栏，Windows 驻留系统托管区 |
| **设置** | 按工具独立配置启用状态、推送/扫描路径、自定义工具、云服务凭据、代理设置 |

每个按钮、对话框和交互的完整说明，请查阅 **[docs/features_zh.md](docs/features_zh.md)**。

## 支持的工具

内置适配器：**Claude Code** · **OpenCode** · **Codex** · **Gemini CLI** · **OpenClaw**

可在设置中添加自定义工具，指定任意本地目录路径即可。

## Skill 格式

有效的 Skill 目录须在根目录下包含 `skill.md` 文件，满足此要求的目录均可通过本地导入或 GitHub 安装。

```
my-skill/
  skill.md     ← 必须存在
  ...其他文件
```

## 云端备份配置

在**设置 → 云存储**中配置，凭据保存在本地配置文件中：

- macOS：`~/Library/Application Support/SkillFlow/config.json`
- Windows：`%APPDATA%\SkillFlow\config.json`

各云服务商所需配置字段：

| 云服务商 | 必填字段 |
|----------|---------|
| 阿里云 OSS | Access Key ID、Access Key Secret、Endpoint |
| 腾讯云 COS | SecretId、SecretKey、Region |
| 华为云 OBS | Access Key、Secret Key、Endpoint |
| Git 仓库 | 仓库地址、分支、用户名、访问令牌 |

## 参与贡献 & 自行构建

### 环境要求

- macOS 11+ 或 Windows 10+
- Go 1.23+
- Node.js 18+
- Wails v2 CLI：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### 构建步骤

```bash
git clone https://github.com/shinerio/SkillFlow
cd SkillFlow
make install-frontend   # 安装前端依赖
make dev                # 热重载开发模式
make test               # 运行 Go 测试
make build              # 构建生产版本 → build/bin/
```

常用 `make` 目标：

| 目标 | 说明 |
|------|------|
| `make dev` | 热重载开发模式（Go + 前端） |
| `make build` | 构建生产版本二进制 |
| `make test` | 运行所有 Go 测试 |
| `make tidy` | 同步 Go 模块依赖 |
| `make generate` | App 方法变更后重新生成 TypeScript 绑定 |
| `make clean` | 删除构建产物 |

内部架构详情请查阅 **[docs/architecture_zh.md](docs/architecture_zh.md)**。
