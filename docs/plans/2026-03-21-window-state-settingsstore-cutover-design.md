# Window State Settingsstore Cutover Design

## Goal

Move window-state type and persistence helpers out of `core/config` and into `core/platform/settingsstore`, because window size is a shell/platform concern rather than part of the transitional domain-facing config compatibility layer.

## Why This Next

`window state` is explicitly documented as shell/platform-owned. The current `config.Service` methods are thin local-file helpers that fit the new settings-store primitive better than the compatibility facade.

## Chosen Approach

Keep the persisted JSON shape unchanged and relocate only the platform concern:

- add `WindowState` and normalization helpers to `core/platform/settingsstore`
- add `Store.LoadWindowState` / `Store.SaveWindowState`
- refactor `core/config` to alias the type and delegate the persistence methods
- keep shell call sites stable by preserving `config.WindowState` and `config.Service` methods as compatibility wrappers
- update architecture docs to note that window-state persistence now uses the platform settings store directly

## Expected Outcome

After the cutover:

- `window` persistence mechanics live in `core/platform/settingsstore`
- `core/config` keeps compatibility wrappers but no longer owns that shell concern
- no on-disk schema or frontend contract changes are introduced
