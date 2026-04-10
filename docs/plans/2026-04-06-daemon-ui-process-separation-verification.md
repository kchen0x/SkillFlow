# Daemon / UI 进程分离验证说明

## 验证目标

确认 macOS 与 Windows 都满足新的运行时模型：

1. 主窗口激活时，同时存在 `daemon` 与 `ui` 两个进程。
2. 主窗口隐藏后，macOS 只保留菜单栏 `daemon`，Windows 只保留托盘 `daemon`。
3. 隐藏后不再保留 `ui` 进程。
4. 重新显示窗口时，`ui` 会冷启动，页面状态丢失。
5. 后台定时任务与业务读写仍通过 `daemon` 正常工作。
6. 隐藏后仅统计常驻 `daemon` 的内存占用。

## 预期运行模型

- 单二进制，双角色。
- 默认启动角色为 `daemon`。
- `daemon` 负责后端运行时、托盘或菜单栏、本地控制端点、loopback 服务网关、后台定时任务，以及 `ui` 的拉起与退出。
- `ui` 负责 Wails 窗口、React 页面，以及少量本地壳层能力。
- 业务调用路径为：前端页面 -> `frontend/src/lib/backend.ts` -> `frontend/src/lib/backendClient.ts` -> 本地 `daemon` `/invoke` 服务。
- 事件路径为：`daemon` 事件总线 -> `/events` NDJSON 流 -> `ui` 壳层转发 -> 前端 Wails 事件订阅。

## runtime 目录检查

应用启动后，`<AppDataDir>/runtime/` 目录应出现这些本地状态文件：

- `helper-control.json`
- `ui-control.json`
- `daemon-service.json`
- `helper.lock`

说明：

- `helper-control.json` 与 `helper.lock` 沿用历史命名，但当前语义对应后台 `daemon` 宿主。
- `daemon-service.json` 记录 UI 访问 `daemon` loopback 网关所需的地址、token 和 PID。

## macOS 手工验证

### 1. 启动应用

```bash
open -a SkillFlow
```

### 2. 记录激活窗口时的进程

```bash
ps -axo pid,ppid,rss,comm | rg 'SkillFlow'
```

验证点：

- 应看到两个相关进程。
- 一个是后台 `daemon` 宿主。
- 一个是当前可见的 `ui` 进程。

### 3. 隐藏窗口到菜单栏

操作方式：

- 点击窗口关闭按钮。
- 或在菜单栏菜单中点击 `Hide SkillFlow`。

### 4. 验证隐藏后的常驻进程

```bash
ps -axo pid,ppid,rss,comm | rg 'SkillFlow'
```

验证点：

- 菜单栏图标仍在。
- 只剩一个后台 `daemon` 进程。
- 不应再看到活动 `ui` 进程。

### 5. 测量隐藏后的 daemon RSS

```bash
ps -axo pid,rss,comm | rg 'SkillFlow'
```

判定方式：

- 仅取隐藏后仍存活的那个 `daemon` 进程。
- `rss` 单位为 KB。
- 通过 `rss / 1024` 换算为 MB。
- 目标是小于 `50 MB`。

### 6. 验证冷启动重开

操作方式：

- 在菜单栏点击 `Show SkillFlow`。

验证点：

- 会重新出现新的 `ui` 进程。
- 页面回到冷启动后的默认状态。
- 不恢复上一次路由、滚动位置和临时选择。

## Windows 手工验证

### 1. 启动应用

直接启动 `SkillFlow-windows.exe`。

### 2. 记录激活窗口时的进程

PowerShell：

```powershell
Get-Process SkillFlow | Select-Object Id, ProcessName, WorkingSet64, StartTime
```

验证点：

- 主窗口可见时，应看到 `daemon` 与 `ui` 两个相关进程。

### 3. 关闭窗口到托盘

操作方式：

- 点击窗口关闭按钮。

### 4. 验证隐藏后的常驻进程

PowerShell：

```powershell
Get-Process SkillFlow | Select-Object Id, ProcessName, WorkingSet64, StartTime
```

验证点：

- 托盘图标仍在。
- 只保留后台 `daemon` 进程。
- `ui` 进程应已退出。

### 5. 测量隐藏后的 daemon 工作集

PowerShell：

```powershell
Get-Process SkillFlow | Select-Object Id, @{Name='WorkingSetMB';Expression={[math]::Round($_.WorkingSet64 / 1MB, 2)}}
```

判定方式：

- 只统计隐藏后仍然存活的那个 `daemon` 进程。
- 目标是 `WorkingSetMB < 50`。

### 6. 验证冷启动重开

操作方式：

- 在托盘菜单点击 `Show SkillFlow`。

验证点：

- 新的 `ui` 进程被重新拉起。
- 页面状态不会被恢复。

## 回归验证命令

代码级验证：

```bash
go test ./cmd/skillflow ./core/...
```

前端验证：

```bash
cd cmd/skillflow/frontend
npm run test:unit
npm run build
```

## 通过标准

- `go test ./cmd/skillflow ./core/...` 通过。
- `npm run test:unit` 通过。
- `npm run build` 通过。
- macOS 隐藏窗口后只剩 `daemon`。
- Windows 隐藏窗口后只剩 `daemon`。
- 两端隐藏后常驻 `daemon` 内存都低于 `50 MB`。
- 两端重新显示窗口时都会冷启动 `ui`，且页面状态丢失。
