# Window State Settingsstore Cutover Plan

1. Add failing tests for `core/platform/settingsstore` window-state normalization and local-file persistence.
2. Implement `WindowState`, normalization, and load/save helpers in `core/platform/settingsstore`.
3. Refactor `core/config` to delegate window-state methods and type ownership to the platform store.
4. Update architecture docs to reflect that window-state persistence is now a platform concern implemented through `settingsstore`.
5. Run `gofmt`, targeted tests, then `go test ./core/... -count=1` and `go test ./cmd/skillflow -count=1`.
