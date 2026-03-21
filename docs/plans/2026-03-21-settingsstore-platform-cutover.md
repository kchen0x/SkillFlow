# Settingsstore Platform Cutover Plan

1. Add characterization tests for a new `core/platform/settingsstore` package covering path resolution and JSON round-trips.
2. Implement the generic settings store in `core/platform/settingsstore`.
3. Refactor `core/config.Service` to use the platform store for path resolution and shared/local JSON persistence.
4. Update architecture docs to reflect that the platform settings-store primitive now exists even though `AppConfig` remains transitional.
5. Run `gofmt`, targeted package tests, then `go test ./core/... -count=1` and `go test ./cmd/skillflow -count=1`.
