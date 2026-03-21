# Settingsstore Platform Cutover Design

## Goal

Extract the config file read/write mechanics from `core/config` into `core/platform/settingsstore` so the repository gains the documented platform storage primitive before the larger context-owned settings split.

## Why This Next

`core/config` is the last active top-level legacy package. Fully replacing `AppConfig` with context-owned namespaces is a larger refactor, but its file storage concerns are already business-agnostic and can move first without changing frontend or shell contracts.

## Chosen Approach

Keep `core/config` as the transitional compatibility facade and move only the storage primitive:

- add `core/platform/settingsstore`
- move data-dir, path resolution, JSON read, and JSON write mechanics into the new package
- refactor `core/config.Service` to delegate file IO to the platform store
- keep `AppConfig`, split/merge logic, and upgrade semantics in `core/config`
- update architecture docs to mark the settings-store primitive as introduced, while noting that namespace ownership is still pending

## Expected Outcome

After the cutover:

- config persistence mechanics live in `core/platform/settingsstore`
- `core/config` shrinks toward a transitional compatibility layer
- the later namespace split can build on an existing platform settings primitive instead of starting from file-path code
