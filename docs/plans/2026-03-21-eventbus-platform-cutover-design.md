# Eventbus Platform Cutover Design

## Goal

Move the reusable event hub from `core/notify` to `core/platform/eventbus` so the package location matches its actual role: a pure technical capability used by the shell.

## Why This Next

After the bounded-context extractions and the `core/registry` cutover, `core/notify` is another top-level cross-cutting package that no longer carries business ownership. The architecture docs already classify event bus behavior as a platform concern, so this is now a mechanical cutover with low semantic risk.

## Chosen Approach

Keep the implementation and payload model stable, and only move the package:

- create `core/platform/eventbus`
- move hub and event model there with the same API shape
- switch shell and tests to the new import path
- remove `core/notify`

This avoids premature redesign. Event publication still stays in the shell for now; only the technical container moves to the correct layer.

## Expected Outcome

After the cutover:

- `cmd/skillflow` depends on `core/platform/eventbus`
- `core/notify` is gone
- migration docs reflect that the platform event-bus move is complete
