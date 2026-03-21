# Loglevel Logging Cutover Design

## Goal

Move log-level string constants and normalization rules out of `core/config` and into `core/platform/logging`, because log-level semantics are part of the logging platform capability rather than the transitional config compatibility layer.

## Why This Next

After moving window state into `settingsstore`, `core/config` still owns platform-only logic for log-level normalization. The shell already depends on `core/platform/logging`, so consolidating log-level semantics there removes another platform concern from the config package without changing persisted JSON keys.

## Chosen Approach

Keep the persisted `logLevel` field unchanged and relocate only the semantics:

- add level-string constants and normalization helpers to `core/platform/logging`
- cover them with logging package tests
- refactor `core/config` to expose compatibility aliases/wrappers
- switch shell logging code to call the logging package directly where practical
- update architecture docs to note that log-level semantics now live in `core/platform/logging`

## Expected Outcome

After the cutover:

- `core/platform/logging` owns log-level strings and normalization
- `core/config` keeps only compatibility wrappers for the persisted config model
- no on-disk schema or user-visible behavior changes are introduced
