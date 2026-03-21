# Logicalkey Shared Cutover Design

## Goal

Move the reusable skill logical-key helpers from `core/skillkey` to `core/shared/logicalkey` so the package location matches the architecture rule that these identifiers belong to the shared kernel.

## Why This Next

`skillkey` is small, stable, and referenced from multiple bounded contexts plus the shell. That makes it a good fit for the shared kernel and a low-risk migration before tackling larger remaining modules like `viewstate` and `config`.

## Chosen Approach

Keep the API unchanged and only relocate the package:

- create `core/shared/logicalkey`
- move the helpers and tests there
- switch all imports
- remove `core/skillkey`

This keeps the migration mechanical while aligning the package layout with the DDD architecture docs.

## Expected Outcome

After the cutover:

- contexts and shell depend on `core/shared/logicalkey`
- `core/skillkey` is removed
- architecture docs reflect the shared-kernel move
