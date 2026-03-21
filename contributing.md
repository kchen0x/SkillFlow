# Contributing

## Prerequisites

- macOS 11+ or Windows 10+
- Go 1.23+
- Node.js 18+
- Wails v2 CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Build Steps

```bash
git clone https://github.com/shinerio/SkillFlow
cd SkillFlow
make install-frontend
make dev
make test
make build
make build-cloud PROVIDERS="aws,google"
```

Notes:

- `make dev`, `make build`, and `make generate` run Wails from `cmd/skillflow/`.
- `make test` runs `go test ./core/...` plus frontend unit tests from `cmd/skillflow/frontend`.
- When you touch shell/runtime code under `cmd/skillflow/`, also run `go test ./cmd/skillflow`.
- Production binaries are written under `cmd/skillflow/build/bin/`.

## Common Make Targets

| Target | Description |
|--------|-------------|
| `make dev` | Run Wails dev mode with frontend hot reload |
| `make build` | Build the production app with all cloud providers |
| `make build-cloud PROVIDERS="aws,google"` | Build with selected cloud providers only |
| `make test` | Run core Go tests and frontend unit tests |
| `make test-cloud PROVIDERS="aws,google"` | Run Go tests with selected cloud providers only |
| `make tidy` | Sync Go module dependencies |
| `make generate` | Regenerate Wails TypeScript bindings |
| `make install-frontend` | Install frontend npm dependencies |
| `make clean` | Remove build artifacts |

For contributor-facing internals, see **[docs/architecture/README.md](docs/architecture/README.md)**.
