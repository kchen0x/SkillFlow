# 参与贡献

## 环境要求

- macOS 11+ 或 Windows 10+
- Go 1.25+
- Node.js 18+
- Wails v2 CLI：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## 构建步骤

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
- `make test` 会运行 `go test ./core/...`，并执行 `cmd/skillflow/frontend` 下的前端单元测试。
- 如果改动了 `cmd/skillflow/` 下的壳层或运行时代码，还应额外执行 `go test ./cmd/skillflow`。
- 生产构建输出位于 `cmd/skillflow/build/bin/`。

## 常用 Make 目标

| 目标 | 说明 |
|------|------|
| `make dev` | 启动 Wails 开发模式并联动前端热更新 |
| `make build` | 构建包含全部云服务商的生产版本 |
| `make build-cloud PROVIDERS="aws,google"` | 仅构建指定云服务商版本 |
| `make test` | 运行核心 Go 测试和前端单元测试 |
| `make test-cloud PROVIDERS="aws,google"` | 使用指定云服务商标签运行 Go 测试 |
| `make tidy` | 同步 Go 模块依赖 |
| `make generate` | 重新生成 Wails TypeScript 绑定 |
| `make install-frontend` | 安装前端 npm 依赖 |
| `make clean` | 清理构建产物 |

面向贡献者的内部说明请查阅 **[docs/architecture/README_zh.md](docs/architecture/README_zh.md)**。
