# Daemon 与 UI 进程分离实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 SkillFlow 改造成单二进制双角色架构，在 macOS 和 Windows 上统一实现“`daemon` 常驻、`ui` 按需拉起、关闭窗口后退出 UI 仅保留后端”。

**Architecture:** 将现有依附在 Wails `App.startup()` 生命周期内的完整后端运行时抽离到 `daemon`，由 `daemon` 统一托管托盘或菜单栏、后台任务、UI 拉起与本地 IPC 服务。`ui` 进程保留 Wails 窗口和少量 shell 能力，前端业务调用通过统一 gateway 访问 `daemon`，不再直接依赖 Wails 导出的业务方法。

**Tech Stack:** Go, Wails v2, React, TypeScript, loopback IPC, Node test runner, Go test, Markdown docs

---

### Task 1: 固化新的进程角色与入口判定

**Files:**
- Modify: `cmd/skillflow/process_role.go`
- Modify: `cmd/skillflow/main.go`
- Modify: `cmd/skillflow/process_control_test.go`
- Create: `cmd/skillflow/main_process_role_test.go`
- Modify: `cmd/skillflow/process_bootstrap_mode.go`
- Modify: `cmd/skillflow/process_bootstrap_mode_darwin.go`

**Step 1: 写失败测试**

新增或扩展测试，证明：

- 默认启动角色会先进入 `daemon`
- `--internal-daemon` 会显式进入 `daemon`
- `--internal-ui` 只启动 `ui`
- macOS 与 Windows 都不再走“UI 自己托管托盘生命周期”的旧分支
- 入口在已有 daemon 存在时不会再启动第二个 daemon

**Step 2: 运行测试确认失败**

Run:

```bash
go test ./cmd/skillflow -run 'TestDetermineProcessRole|TestRunEntry|TestHelperBootstrap' 
```

Expected: `FAIL`，因为 `daemon` 角色与新的入口路由尚未实现。

**Step 3: 写最小实现**

实现：

- 新的 `processRoleDaemon`
- 新的内部角色参数，例如 `--internal-daemon`
- `runEntry()` 新的角色分发逻辑
- 移除 macOS 当前“UI 直接常驻”的特殊分支，让两端统一经由 daemon

**Step 4: 运行测试确认通过**

Run:

```bash
go test ./cmd/skillflow -run 'TestDetermineProcessRole|TestRunEntry|TestHelperBootstrap'
```

Expected: `ok`

**Step 5: Commit**

```bash
git add cmd/skillflow/process_role.go cmd/skillflow/main.go cmd/skillflow/process_control_test.go cmd/skillflow/main_process_role_test.go cmd/skillflow/process_bootstrap_mode.go cmd/skillflow/process_bootstrap_mode_darwin.go
git commit -m "refactor: introduce daemon and ui process roles"
```

### Task 2: 提取不依赖 Wails 的后端运行时容器

**Files:**
- Create: `core/platform/daemon/runtime.go`
- Create: `core/platform/daemon/runtime_test.go`
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/app_startup_test.go`
- Modify: `cmd/skillflow/app_tray_lifecycle.go`

**Step 1: 写失败测试**

新增测试，证明：

- 完整后端运行时可以在没有 Wails `context.Context` 的情况下初始化
- 配置升级、配置加载、日志、存储、memory service、event hub、startup scheduler 能由 daemon runtime 持有
- `App.startup()` 不再承担完整后端容器初始化职责

**Step 2: 运行测试确认失败**

Run:

```bash
go test ./cmd/skillflow ./core/platform/daemon -run 'TestStartup|TestDaemonRuntime'
```

Expected: `FAIL`，因为 daemon runtime 还不存在，后端初始化仍耦合在 `App.startup()`。

**Step 3: 写最小实现**

提取一个独立的 backend runtime，承接当前 `App.startup()` 中与业务和后台任务相关的初始化：

- config / upgrade
- logger
- storage / repository
- memory services
- event hub
- startup task scheduler
- auto sync timer

同时把 UI 进程的 `App` 缩减为薄壳，只保留窗口和 shell 相关状态。

**Step 4: 运行测试确认通过**

Run:

```bash
go test ./cmd/skillflow ./core/platform/daemon -run 'TestStartup|TestDaemonRuntime'
```

Expected: `ok`

**Step 5: Commit**

```bash
git add core/platform/daemon/runtime.go core/platform/daemon/runtime_test.go cmd/skillflow/app.go cmd/skillflow/app_startup_test.go cmd/skillflow/app_tray_lifecycle.go
git commit -m "refactor: extract backend runtime from wails app"
```

### Task 3: 建立 daemon 与 UI 共用的 IPC 基础层

**Files:**
- Create: `core/platform/ipc/endpoint.go`
- Create: `core/platform/ipc/protocol.go`
- Create: `core/platform/ipc/server.go`
- Create: `core/platform/ipc/client.go`
- Create: `core/platform/ipc/ipc_test.go`
- Modify: `cmd/skillflow/process_control.go`

**Step 1: 写失败测试**

新增测试，证明：

- endpoint 状态文件会写入地址、token、PID
- client 能使用 token 成功往返调用
- 错误 token 会被拒绝
- stale endpoint 会被清理
- 新服务通道可以与现有控制通道共享统一抽象

**Step 2: 运行测试确认失败**

Run:

```bash
go test ./cmd/skillflow ./core/platform/ipc -run 'TestLoopback|TestIPC'
```

Expected: `FAIL`，因为 `core/platform/ipc` 还不存在。

**Step 3: 写最小实现**

把现有 loopback 控制 server 的通用能力下沉到 `core/platform/ipc/`：

- endpoint state
- token
- server accept loop
- client request / response
- 超时与通用错误

保留 `cmd/skillflow/process_control.go` 作为兼容入口，后续逐步改为调用新层。

**Step 4: 运行测试确认通过**

Run:

```bash
go test ./cmd/skillflow ./core/platform/ipc -run 'TestLoopback|TestIPC'
```

Expected: `ok`

**Step 5: Commit**

```bash
git add core/platform/ipc/endpoint.go core/platform/ipc/protocol.go core/platform/ipc/server.go core/platform/ipc/client.go core/platform/ipc/ipc_test.go cmd/skillflow/process_control.go
git commit -m "refactor: add local ipc foundation for daemon and ui"
```

### Task 4: 让 daemon 真正托管托盘、UI 生命周期与后台任务

**Files:**
- Create: `core/platform/daemon/service.go`
- Create: `core/platform/daemon/service_test.go`
- Modify: `cmd/skillflow/process_helper.go`
- Modify: `cmd/skillflow/process_helper_darwin.go`
- Modify: `cmd/skillflow/process_helper_other.go`
- Modify: `cmd/skillflow/tray_controller.go`
- Modify: `cmd/skillflow/tray_darwin_callbacks.go`
- Modify: `cmd/skillflow/tray_windows.go`

**Step 1: 写失败测试**

新增测试，证明：

- daemon 可以在启动时初始化托盘/菜单栏
- daemon 能拉起 `ui`
- daemon 在 `ui` 退出后保持存活
- `Show SkillFlow` 能在没有活动 UI 时重拉起 `ui`
- `Quit SkillFlow` 能先停 `ui` 再停 daemon

**Step 2: 运行测试确认失败**

Run:

```bash
go test ./cmd/skillflow ./core/platform/daemon -run 'TestHelper|TestDaemonService|TestTray'
```

Expected: `FAIL`，因为当前 helper 仍是轻量控制器，不承载完整 runtime。

**Step 3: 写最小实现**

把当前 `helperController` 演进成 `daemon` 宿主：

- 持有 daemon runtime
- 持有控制通道与服务通道
- 持有 UI 子进程引用
- 托盘/菜单栏回调全部收敛到 daemon 生命周期

必要时重命名 `helper` 相关类型与日志文案为 `daemon`，避免概念继续混乱。

**Step 4: 运行测试确认通过**

Run:

```bash
go test ./cmd/skillflow ./core/platform/daemon -run 'TestHelper|TestDaemonService|TestTray'
```

Expected: `ok`

**Step 5: Commit**

```bash
git add core/platform/daemon/service.go core/platform/daemon/service_test.go cmd/skillflow/process_helper.go cmd/skillflow/process_helper_darwin.go cmd/skillflow/process_helper_other.go cmd/skillflow/tray_controller.go cmd/skillflow/tray_darwin_callbacks.go cmd/skillflow/tray_windows.go
git commit -m "feat: run tray and background services from daemon"
```

### Task 5: 缩减 UI 进程为 Wails 薄壳并保留 shell 能力

**Files:**
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/app_settings.go`
- Modify: `cmd/skillflow/app_backup.go`
- Modify: `cmd/skillflow/app_memory.go`
- Modify: `cmd/skillflow/app_prompt.go`
- Modify: `cmd/skillflow/app_agent_memory.go`
- Modify: `cmd/skillflow/app_startup.go`
- Modify: `cmd/skillflow/app_visibility.go`
- Modify: `cmd/skillflow/app_visibility_test.go`

**Step 1: 写失败测试**

新增测试，证明：

- UI 进程关闭窗口时会真正退出，而不是隐藏保活
- UI 只保留窗口与 shell 相关方法
- 业务方法不再直接从 Wails `App` 访问完整 runtime

**Step 2: 运行测试确认失败**

Run:

```bash
go test ./cmd/skillflow -run 'TestBuildUIOptions|TestWindowVisibility|TestApp'
```

Expected: `FAIL`，因为 UI 仍然承担业务和后台生命周期。

**Step 3: 写最小实现**

实现：

- UI 关闭时始终退出进程
- `HideWindowOnClose` 和关闭处理逻辑改为匹配新模型
- Wails `App` 导出方法只保留 shell 级能力
- 原业务方法改为通过 daemon client 转发，或从 UI `App` 中删除并准备前端迁移

**Step 4: 运行测试确认通过**

Run:

```bash
go test ./cmd/skillflow -run 'TestBuildUIOptions|TestWindowVisibility|TestApp'
```

Expected: `ok`

**Step 5: Commit**

```bash
git add cmd/skillflow/app.go cmd/skillflow/app_settings.go cmd/skillflow/app_backup.go cmd/skillflow/app_memory.go cmd/skillflow/app_prompt.go cmd/skillflow/app_agent_memory.go cmd/skillflow/app_startup.go cmd/skillflow/app_visibility.go cmd/skillflow/app_visibility_test.go
git commit -m "refactor: slim ui process into shell-only wails app"
```

### Task 6: 新增前端 backend gateway 并迁移业务调用

**Files:**
- Create: `cmd/skillflow/frontend/src/lib/backend.ts`
- Create: `cmd/skillflow/frontend/src/lib/backendClient.ts`
- Create: `cmd/skillflow/frontend/tests/backend.test.mjs`
- Modify: `cmd/skillflow/frontend/src/App.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/Dashboard.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/Memory.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/Prompts.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/Settings.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/Backup.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/StarredRepos.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/SyncPull.tsx`
- Modify: `cmd/skillflow/frontend/src/pages/SyncPush.tsx`
- Modify: `cmd/skillflow/frontend/src/components/CategoryPanel.tsx`
- Modify: `cmd/skillflow/frontend/src/components/PromptCategoryPanel.tsx`
- Modify: `cmd/skillflow/frontend/src/components/SkillCard.tsx`
- Modify: `cmd/skillflow/frontend/src/components/SyncSkillCard.tsx`

**Step 1: 写失败测试**

新增测试，证明：

- gateway 能统一请求 daemon 服务通道
- base URL 与 token 从运行时配置读取
- 常见错误会映射成稳定前端异常
- 页面层不再直接依赖 Wails 业务方法

**Step 2: 运行测试确认失败**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `FAIL`，因为 gateway 文件不存在，页面仍直接 import `wailsjs/go/main/App` 的业务方法。

**Step 3: 写最小实现**

实现：

- 新建前端 backend gateway
- 将业务方法改为经由 gateway 调 daemon
- 保留 shell 专用调用，例如剪贴板、窗口运行时事件、少量原生对话框能力

避免在本任务中重写页面业务逻辑，只替换通信入口。

**Step 4: 运行测试确认通过**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Run:

```bash
cd cmd/skillflow/frontend && npm run build
```

Expected: 两条命令都返回 `ok`

**Step 5: Commit**

```bash
git add cmd/skillflow/frontend/src/lib/backend.ts cmd/skillflow/frontend/src/lib/backendClient.ts cmd/skillflow/frontend/tests/backend.test.mjs cmd/skillflow/frontend/src/App.tsx cmd/skillflow/frontend/src/pages/Dashboard.tsx cmd/skillflow/frontend/src/pages/Memory.tsx cmd/skillflow/frontend/src/pages/ToolSkills.tsx cmd/skillflow/frontend/src/pages/Prompts.tsx cmd/skillflow/frontend/src/pages/Settings.tsx cmd/skillflow/frontend/src/pages/Backup.tsx cmd/skillflow/frontend/src/pages/StarredRepos.tsx cmd/skillflow/frontend/src/pages/SyncPull.tsx cmd/skillflow/frontend/src/pages/SyncPush.tsx cmd/skillflow/frontend/src/components/CategoryPanel.tsx cmd/skillflow/frontend/src/components/PromptCategoryPanel.tsx cmd/skillflow/frontend/src/components/SkillCard.tsx cmd/skillflow/frontend/src/components/SyncSkillCard.tsx
git commit -m "refactor: route frontend business calls through daemon gateway"
```

### Task 7: 同步文档与运行时文件说明

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `docs/config.md`
- Modify: `docs/config_zh.md`
- Modify: `docs/architecture/overview.md`
- Modify: `docs/architecture/overview_zh.md`
- Modify: `docs/architecture/runtime-and-storage.md`
- Modify: `docs/architecture/runtime-and-storage_zh.md`
- Modify: `docs/architecture/README.md`
- Modify: `docs/architecture/README_zh.md`

**Step 1: 写文档更新**

更新内容至少包括：

- 新的 `daemon + ui` 进程模型
- macOS / Windows 统一的窗口关闭行为
- `runtime/*.json` 的新含义
- 后端从 Wails `App` 抽离到 daemon 的架构变化
- UI 冷启动而非页面状态恢复的说明

**Step 2: 运行检查**

Run:

```bash
rg -n "helper|daemon|菜单栏|托管区|HideWindowOnClose|runtime/" docs/features.md docs/features_zh.md docs/config.md docs/config_zh.md docs/architecture/overview.md docs/architecture/overview_zh.md docs/architecture/runtime-and-storage.md docs/architecture/runtime-and-storage_zh.md docs/architecture/README.md docs/architecture/README_zh.md
```

Expected: 匹配结果反映新的 daemon/UI 模型，不再保留与当前实现冲突的旧 helper 描述。

**Step 3: Commit**

```bash
git add docs/features.md docs/features_zh.md docs/config.md docs/config_zh.md docs/architecture/overview.md docs/architecture/overview_zh.md docs/architecture/runtime-and-storage.md docs/architecture/runtime-and-storage_zh.md docs/architecture/README.md docs/architecture/README_zh.md
git commit -m "docs: document daemon and ui process separation"
```

### Task 8: 端到端验证进程行为与内存预算

**Files:**
- Create: `docs/plans/2026-04-06-daemon-ui-process-separation-verification.md`
- Modify: `cmd/skillflow/main_darwin_test.go`
- Create: `cmd/skillflow/main_windows_process_test.go`

**Step 1: 写失败测试或验证占位**

新增测试或验证说明，覆盖：

- 启动后存在两个角色
- 关闭窗口后只剩 daemon
- `Show SkillFlow` 能重新拉起 UI
- `Quit SkillFlow` 能终止全部角色

同时在验证文档里写明 macOS / Windows 的手工内存测量命令和通过标准。

**Step 2: 运行测试确认失败**

Run:

```bash
go test ./cmd/skillflow -run 'TestRunEntry|TestBuildUIOptions|TestDaemon'
```

Expected: `FAIL`，直到新的进程行为覆盖完成。

**Step 3: 写最小实现与验证文档**

补齐缺失测试，并在验证文档中写清：

- macOS 使用 `ps` 或 Activity Monitor 检查 daemon RSS
- Windows 使用 `tasklist`、PowerShell 或资源监视器检查 daemon 工作集
- 验收口径只统计“关闭窗口后仍存活的 daemon”

**Step 4: 运行最终验证**

Run:

```bash
go test ./cmd/skillflow ./core/...
```

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Run:

```bash
cd cmd/skillflow/frontend && npm run build
```

Expected: 三条命令都返回 `ok`

**Step 5: 手工验收**

在 macOS 和 Windows 上分别验证：

1. 启动后存在 `daemon + ui`
2. 关闭窗口后只剩 `daemon`
3. 托盘或菜单栏可以重开 UI
4. 关闭窗口后的 daemon RSS `< 50MB`

**Step 6: Commit**

```bash
git add docs/plans/2026-04-06-daemon-ui-process-separation-verification.md cmd/skillflow/main_darwin_test.go cmd/skillflow/main_windows_process_test.go
git commit -m "test: verify daemon-ui split lifecycle"
```
