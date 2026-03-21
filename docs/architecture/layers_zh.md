# 分层与依赖规则

## 分层定义

在 Wails 约束下，transport adapter 留在 `cmd/skillflow/`，而每个 bounded context 内部的可复用后端代码统一组织为：

```text
cmd/skillflow/
  transport adapter
  shell bootstrap
  tray/window/process 集成

core/<context>/
  app/
    command/
    query/
    port/
      repository/
      gateway/
  domain/
  infra/
    repository/
    gateway/
    projection/

core/config/
  面向前端的设置门面
  shared/local split-merge 持久化适配器
```

## Transport Adapter

当前实现位置：

- `cmd/skillflow/` 中的 Wails `App` 方法

职责：

- 接收 Wails 调用
- 做请求结构校验和简单前置约束校验
- 把请求 DTO 转成应用命令、编排调用或 read model 查询
- 把结果映射回 transport DTO

不负责：

- 领域规则判断
- 直接访问文件系统、Git、云服务或 SDK
- 隐式的跨上下文业务协调

如果增加 CLI 或 API 入口，也应在模块边界承担同样的角色。

## `app`

应用层拥有用例。

职责：

- 协调聚合和领域服务
- 定义事务边界
- 通过端口调用 repository 和 gateway
- 在状态变更成功后发布领域事件
- 暴露 command 和上下文拥有的 query

结构：

- `command/` 保存写用例
- `query/` 保存上下文拥有的读用例
- `port/repository/` 保存持久化接口
- `port/gateway/` 保存外部系统接口

## `domain`

领域层拥有业务语义。

职责：

- 聚合根
- 实体
- 值对象
- 领域服务
- 领域策略
- 领域事件

规则：

- 不依赖 `infra`、Wails、JSON、SDK 或文件系统细节
- 不直接依赖其他上下文的聚合

## `infra`

基础设施层负责实现应用层定义的端口。

职责：

- repository 实现
- gateway 实现
- 当前上下文拥有的 projection 或本地缓存实现
- 文件系统、Git、云服务和 runtime 适配器

规则：

- `infra` 可以依赖 `app` 中定义的端口和 `domain`
- 一个上下文不能依赖另一个上下文的 `infra`

## 跨上下文模块

## `orchestration/`

当写操作跨越多个上下文时，使用 `orchestration/` 做显式协调。

例如：

- 从来源导入技能后再自动推送
- 更新已安装技能后刷新已推送副本
- 恢复备份后重建派生状态

壳层启动时序不属于领域编排，它应保留在 `cmd/skillflow/app_startup.go`、`cmd/skillflow/main.go`、`cmd/skillflow/process_bootstrap_mode*.go` 这类壳层启动文件中。

## `readmodel/`

当读视图需要组合多个上下文的数据时，使用 `readmodel/`。

例如：

- Dashboard
- My Agents 聚合视图
- 带安装状态的来源候选技能列表

规则：

- `readmodel/` 只能依赖显式发布的 query provider 或 published language DTO
- `readmodel/` 不能直接依赖其他上下文的 repository 或未发布的 query 内部实现
- `readmodel/` 可以维护 projection 缓存，但不拥有业务真相

## `core/config`

`core/config` 是面向前端的设置门面，不是 bounded context，也不是 read model。

职责：

- 合并/拆分共享 `config.json` 与本地 `config_local.json`
- 把字段逻辑归属委托给上下文拥有的设置组件与 platform 设置组件
- 为 `GetConfig` / `SaveConfig` 暴露一个 transport 友好的统一设置 DTO

规则：

- 业务真相仍归属于对应上下文与 `platform/`
- 不要把无关的业务规则堆进这个门面包

## `platform/`

`platform/` 保存纯技术能力：

- 日志
- 文件系统辅助
- Git 客户端基础能力
- HTTP 客户端基础能力
- 事件总线
- settings store
- 启动期升级辅助
- 路径规范化
- 更新下载基础能力

`platform/` 必须保持无业务语义。

## `shared/`

`shared/` 保存最小共享内核：

- 逻辑键
- 通用领域错误
- 基础事件契约

不要把上下文内部 ID 或具体业务行为塞进 `shared/`。

## 依赖规则

允许的方向：

```text
transport adapters -> app
transport adapters -> orchestration
transport adapters -> readmodel
transport adapters -> core/config
orchestration -> app -> domain
readmodel -> published query providers / published language
infra -> app/port + domain
core/config -> 上下文设置组件 + platform 设置组件
platform -> 不依赖业务
shared -> 不依赖具体上下文
```

禁止的方向：

- `domain -> infra`
- `domain -> transport adapters`
- `上下文 A -> 上下文 B 的 infra`
- `transport adapters -> infra`
- `readmodel -> 直接依赖上下文 repository`
- `readmodel -> 依赖未发布的应用层内部查询实现`

## Repository 与 Gateway 的划分

判断规则：

- 如果它负责读写本上下文拥有的真相，就是 `repository`
- 如果它负责与上下文边界之外的系统通信，就是 `gateway`

例如：

- 已安装技能元数据存储：`repository`
- Prompt 库存储：`repository`
- 来源跟踪清单：`repository`
- GitHub Releases API 客户端：`gateway`
- 云对象存储客户端：`gateway`
- Agent 工作区适配器：`gateway`
- Wails 文件选择对话框桥接：`gateway`

## 扩展规则

新增一个能力时，按这个顺序处理：

1. 先判断它归属哪个 bounded context
2. 新增或修改对应的应用层 command/query
3. 把业务规则放进 `domain`
4. 在 `app/port` 定义新端口
5. 在 `infra` 实现这些端口
6. 通过 transport adapter、orchestration 服务或 read model 对外暴露

*最后更新：2026-03-21*
