# Loglevel Logging Cutover Plan

1. Add failing tests for level-string normalization in `core/platform/logging`.
2. Implement level-string constants and normalization helpers in `core/platform/logging`.
3. Refactor `core/config` to alias or wrap the logging-owned semantics.
4. Update architecture docs to reflect that log-level semantics now belong to the logging platform module.
5. Run `gofmt`, targeted tests, then `go test ./core/... -count=1` and `go test ./cmd/skillflow -count=1`.
