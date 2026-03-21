# Logging Platform Cutover Design

## Goal

Move the reusable file logger from `core/applog` to `core/platform/logging` so the package location matches the architecture rule that logging is a platform capability.

## Why This Next

After the registry and eventbus cutovers, `core/applog` is the next small technical package that still sits at the top level of `core/`. It has no business ownership and is only used by the shell helper and UI runtime.

## Chosen Approach

Keep the logger implementation unchanged and only move the package:

- create `core/platform/logging`
- copy logger implementation and tests there
- update shell imports and types
- remove `core/applog`

This keeps the cutover mechanical and low risk. The logger API, rotation policy, and call sites all stay the same.

## Expected Outcome

After this cutover:

- shell logging depends on `core/platform/logging`
- `core/applog` is removed
- architecture progress docs reflect that the logging platform move is complete
