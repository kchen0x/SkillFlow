# Viewstate Readmodel Cutover Plan

1. Move `core/viewstate` files and tests into `core/readmodel/viewstate` without changing exported behavior.
2. Replace imports in `cmd/skillflow` to use the new readmodel package path.
3. Update architecture docs to reflect that `core/viewstate` has been removed and read-state caching now lives under `core/readmodel/viewstate`.
4. Run `gofmt`, verify no references to `core/viewstate` remain, and run `go test ./core/... -count=1` plus `go test ./cmd/skillflow -count=1`.
