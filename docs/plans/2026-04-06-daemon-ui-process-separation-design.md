# Daemon 与 UI 进程分离设计

## 背景

当前仓库里的进程模型在 macOS 和 Windows 上并不一致：

- macOS 目前实际是由完整的 Wails UI 进程自己托管菜单栏，关闭窗口后隐藏窗口但不退出整个进程。
- Windows 目前实际是轻量 helper 常驻，关闭窗口后 UI 进程退出，但 helper 并不承载完整后端。

这个现状带来两个直接问题：

1. 平台行为不一致，后续后台任务模型无法统一。
2. “关闭窗口后仅保留后端、释放前端内存”这个目标目前无法成立：
   - macOS 关闭窗口后仍保留 WebView 和前端运行时。
   - Windows 关闭窗口后只剩 helper，不是完整后端。

本次设计的目标是把两端统一为真正的“常驻后端 + 按需 UI”架构。

## 目标

- 在 macOS 和 Windows 上统一为双角色架构：
  - `daemon` 常驻进程
  - `ui` 临时进程
- 窗口显示期间同时存在 `daemon` 和 `ui` 两个进程。
- 窗口关闭到托管区或菜单栏时，`ui` 进程退出，只保留 `daemon`。
- `daemon` 承载完整后端能力：
  - 配置加载
  - 仓储访问
  - 扫描与同步
  - 定时任务
  - 托盘/菜单栏
  - 单实例控制
  - UI 拉起与回收
- 允许 UI 冷启动，不要求恢复关闭前的页面状态。
- 窗口关闭后仅剩 `daemon` 时，目标常驻内存占用低于 50MB。

## 非目标

- 不追求在本次改造中保留关闭前的页面树、滚动位置或临时表单输入。
- 不引入远程网络 API，本次通信仅限本机 IPC。
- 不新增第二个发布二进制；仍保持单一可执行文件，通过不同进程角色运行。
- 不在业务层保留旧 helper 语义；旧 helper 逻辑要演进为真正的 `daemon` 角色。

## 方案选择

本次采用“单一可执行文件，多角色双进程”方案：

- 入口程序根据命令行角色标记启动为 `daemon` 或 `ui`。
- 普通用户启动应用时，先确保 `daemon` 存在，再由 `daemon` 拉起 `ui`。
- `daemon` 统一托管菜单栏/托盘和后台任务。
- `ui` 仅负责窗口和用户交互，不再作为完整后端宿主。

不采用“双二进制”方案，原因是当前仓库已经有角色判断、托盘、控制通道等基础设施，继续沿单一可执行文件演进可以降低打包、签名、升级和安装复杂度。

## 目标进程模型

### 角色定义

#### `daemon`

常驻进程，负责：

- 初始化完整后端运行时
- 托盘/菜单栏生命周期
- 后台定时任务
- 单实例控制
- UI 进程拉起、关闭和存活探测
- 对 UI 暴露本地 IPC 服务

#### `ui`

临时进程，负责：

- 启动 Wails 窗口
- 承载 React 前端页面
- 调用 shell 级能力
- 通过本地 IPC 调用 `daemon`

### 生命周期

1. 用户首次启动应用：
   - 若 `daemon` 不存在，则先启动 `daemon`
   - `daemon` 再拉起 `ui`
2. 窗口显示期间：
   - `daemon` 与 `ui` 同时存在
3. 用户关闭窗口：
   - `ui` 主动退出
   - `daemon` 保留并继续常驻
4. 用户点击托盘/菜单栏 `Show SkillFlow`：
   - 若没有活动 `ui`，`daemon` 拉起新 `ui`
   - 若已有活动 `ui`，则聚焦已有窗口
5. 用户点击 `Quit SkillFlow`：
   - `daemon` 先关闭 `ui`
   - 再停止自身

### 平台行为

- Windows：关闭窗口后只保留托管区图标与 `daemon`
- macOS：关闭窗口后只保留菜单栏图标与 `daemon`

两端的真实进程语义保持一致，只是入口位置不同。

## 模块边界设计

### `cmd/skillflow/`

保留：

- 进程入口
- 角色判定
- Wails UI 壳
- 托盘/菜单栏宿主接线
- 极薄的本地 IPC client 接线

继续保持平铺的 `package main` 文件组织，不在 `cmd/skillflow/` 下创建子目录。

### `core/platform/daemon/`

新增 daemon 运行时编排层，负责：

- 后端容器初始化
- 定时任务调度
- daemon 控制服务
- UI 进程管理
- 平台无关的生命周期编排

### `core/platform/ipc/`

新增 IPC 协议与通道层，负责：

- endpoint 状态文件
- token 生成与校验
- request / response 协议
- server / client 封装
- 错误映射与超时

### 业务模块

`core/orchestration/`、`core/readmodel/` 和各 bounded context 保持业务职责不变，但不得再依赖 Wails 生命周期。真正业务逻辑必须能由 `daemon` 直接初始化和调用。

## 通信设计

本次改造采用两层本地 IPC：

### 1. 控制通道

用于轻量生命周期控制：

- `show-ui`
- `close-ui`
- `quit-ui`
- `quit-daemon`
- `ping`

主要服务于：

- 二次启动唤醒
- 托盘/菜单栏操作
- 进程间显示/退出协调

### 2. 服务通道

用于业务调用。前端页面原先直接通过 `wailsjs/go/main/App` 调用的大部分业务方法，要改为通过本地服务通道访问 `daemon`。

约束：

- 仅监听 `127.0.0.1`
- 启动时写入运行时 endpoint 文件
- 使用随机 token 鉴权
- 不允许 UI 直接访问业务仓储或后台调度器

## Shell 能力与业务能力拆分

本次改造后，Wails 导出方法只保留 shell 级能力，不再承载完整业务后端。

### 继续保留在 UI 进程内的能力

- 原生对话框
- 打开系统路径
- 剪贴板
- 窗口控制
- 仅与当前窗口实例绑定的事件订阅

### 迁移到 daemon 的能力

- 配置读写
- 技能、提示词、记忆、智能体相关读写
- 仓库扫描、更新检查、同步、推送、拉取
- 备份与恢复主流程
- 后台定时任务

这能保证 UI 退出后，真正需要常驻的能力仍然可用。

## 前端改造策略

前端不能继续在页面里直接散落调用 `wailsjs/go/main/App` 的业务方法。为降低迁移风险，新增一个统一 gateway：

- `cmd/skillflow/frontend/src/lib/backend.ts`

页面层统一改为调用 gateway，gateway 在本次架构下负责：

- 调用 daemon 服务通道
- 统一 token / base URL / 错误处理
- 为页面提供稳定 API

Wails 运行时能力和 shell 专用方法仍可保留在前端对 `wailsjs` 的少量调用中，但业务调用必须收敛到 gateway。

## 关闭与重开行为

### 关闭窗口

用户点关闭按钮后：

1. `ui` 发起关闭请求
2. `daemon` 记录 UI 已关闭并保留常驻运行
3. `ui` 主动退出进程

这意味着关闭窗口后：

- WebView 被释放
- React 页面树被释放
- JS runtime 被释放

### 重新打开窗口

用户点击 `Show SkillFlow` 后：

1. `daemon` 检查当前 `ui` 是否仍存活
2. 若不存在，则拉起新的 `ui`
3. 新 `ui` 冷启动后重新拉取必要数据

由于产品要求允许页面状态丢失，本设计不引入额外页面状态恢复机制。

## 内存预算策略

目标口径是“窗口关闭后仅剩 `daemon` 时的常驻占用”，不包含被重新拉起的 UI 进程。

控制原则：

1. `daemon` 不持有 WebView、前端资源或页面缓存。
2. 大对象结果按需加载，不在常驻阶段长期保留。
3. 后台任务默认懒触发，不因为 daemon 启动就预热所有列表数据。
4. 仅保留必要配置、索引、调度状态和轻量缓存。
5. 后续需要增加一套可重复执行的手工验证流程，分别在 macOS 和 Windows 上测量关闭窗口后的 `daemon` RSS。

## 测试与验收

### 自动化验证重点

- 角色判定与入口路由
- daemon 单实例与控制通道
- UI 关闭后可被重新拉起
- IPC token 与 endpoint 状态文件
- 前端 gateway 的请求与错误映射

### 手工验收标准

1. 启动应用后，系统中存在 `daemon + ui` 两个进程
2. 关闭窗口后，只剩 `daemon`
3. 点击托盘/菜单栏 `Show SkillFlow` 后，出现新的 `ui`
4. `daemon` 在无 `ui` 时仍能继续后台定时任务
5. `Quit SkillFlow` 能同时终止 `daemon` 和活动 `ui`
6. 关闭窗口后 `daemon` 的空闲 RSS 小于 50MB

## 文档影响面

本次改造不仅是功能变化，也是架构变化，后续实现必须同步更新：

- `docs/features.md`
- `docs/features_zh.md`
- `docs/config.md`
- `docs/config_zh.md`
- `docs/architecture/overview.md`
- `docs/architecture/overview_zh.md`
- `docs/architecture/runtime-and-storage.md`
- `docs/architecture/runtime-and-storage_zh.md`
- 必要时更新架构索引文档

## 风险

### 风险 1：业务逻辑仍残留在 Wails `App`

如果仅把进程拆开，但业务初始化仍残留在 `App.startup()`，则 UI 退出后后台能力会丢失，等于没有完成改造。

### 风险 2：前端仍直接依赖 `wailsjs` 业务方法

如果页面层继续散落调用 Wails 业务方法，后续通信切换和测试会非常脆弱，也无法真正让 UI 成为可回收的薄壳。

### 风险 3：daemon 过度缓存导致无法压到 50MB

如果把现有 UI 里的大对象状态原样搬进 daemon，而不做缓存分级和懒加载，最终很可能无法满足内存目标。

### 风险 4：平台行为再次分叉

如果 macOS 与 Windows 在托盘/菜单栏实现上继续保留不同的生命周期语义，后续定时任务和问题排查会重新复杂化。

## 结论

本次应将 SkillFlow 统一改造成“`daemon` 常驻 + `ui` 临时进程”的单二进制双角色架构：

- 窗口激活时存在两个进程
- 窗口关闭时只保留 `daemon`
- 完整后端只驻留在 `daemon`
- `ui` 成为可随时退出和重建的前端壳

这是满足“后台定时任务常驻”与“关闭窗口释放前端内存”这两个要求的最小正确架构。
