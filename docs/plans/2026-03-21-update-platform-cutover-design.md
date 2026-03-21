# Update Platform Cutover Design

## Goal

Move the GitHub skill-update checker from `core/update` to `core/platform/update` so the package location matches its role as an external-service technical adapter.

## Why This Next

`core/update` is now a narrow technical package with no remaining production call sites outside generic update-check logic. It fits the architecture rule that app-release and external update primitives belong under `platform/`.

## Chosen Approach

Keep the checker API unchanged and only relocate the package:

- create `core/platform/update`
- move the checker and tests there
- remove `core/update`
- update migration docs

This keeps the cutover mechanical and low risk.

## Expected Outcome

After the cutover:

- update-check primitives live under `core/platform/update`
- `core/update` is removed
- architecture docs reflect the completed platform move
