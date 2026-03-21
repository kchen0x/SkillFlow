# Viewstate Readmodel Cutover Design

## Goal

Move the derived view-state cache package from `core/viewstate` to `core/readmodel/viewstate` so the package location matches the architecture rule that cross-context read-side projections belong under `readmodel/`.

## Why This Next

`viewstate` is already a read-side cache used by shell query assembly. It has a narrow import surface and no business truth ownership, which makes it a low-risk migration that establishes the first concrete package under `core/readmodel` before tackling larger settings work.

## Chosen Approach

Keep the package API and on-disk snapshot format unchanged and only relocate the code:

- create `core/readmodel/viewstate`
- move the cache, fingerprint, and agent-presence projection helpers with their tests
- switch shell imports
- remove `core/viewstate`
- update architecture docs to mark the readmodel move as completed

## Expected Outcome

After the cutover:

- shell read adapters depend on `core/readmodel/viewstate`
- `core/viewstate` is removed
- the repository has its first concrete `readmodel` package aligned with the architecture docs
