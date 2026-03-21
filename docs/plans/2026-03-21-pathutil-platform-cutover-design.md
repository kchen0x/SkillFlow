# Pathutil Platform Cutover Design

## Goal

Move portable path normalization helpers from `core/pathutil` to `core/platform/pathutil` so the package location matches their role as a business-agnostic technical capability.

## Why This Next

`pathutil` is small, isolated, and currently used only by repository implementations in `skillcatalog` and `skillsource`. That makes it a low-risk cutover that still reduces the number of top-level cross-cutting packages under `core/`.

## Chosen Approach

Keep the API unchanged and only relocate the package:

- create `core/platform/pathutil`
- move the helpers and tests there
- switch the repository imports
- remove `core/pathutil`

This keeps the migration mechanical and avoids incidental redesign.

## Expected Outcome

After the cutover:

- repositories depend on `core/platform/pathutil`
- `core/pathutil` is removed
- migration docs reflect that the path utility move is complete
