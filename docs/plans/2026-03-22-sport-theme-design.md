# Add Sport Theme Design

**Date:** 2026-03-22

## Context

SkillFlow currently exposes three frontend-only appearance themes:

- `dark`
- `young`
- `light`

Those themes are implemented through one shared path:

- `cmd/skillflow/frontend/src/hooks/useTheme.ts` defines the theme enum, labels, and cycle order
- `cmd/skillflow/frontend/src/style.css` defines the CSS variable token blocks
- `cmd/skillflow/frontend/src/pages/Settings.tsx` renders theme preview cards
- `cmd/skillflow/frontend/src/i18n/en.ts` and `cmd/skillflow/frontend/src/i18n/zh.ts` describe the themes in the UI
- `cmd/skillflow/frontend/tests/themeContrast.test.mjs` verifies baseline accessibility contrast

The user wants a new `sport` color style added by referencing the green, mint, and field-like palette from the provided dashboard example. The user explicitly clarified that `sport` must be added as a fourth theme, not replace any existing theme.

## Decision

1. Add `sport` as a fourth theme alongside `dark`, `young`, and `light`.
2. Keep the existing frontend-only theme architecture:
   - no backend config changes
   - no new theme abstraction layer
   - no page-specific overrides
3. Extend the existing theme cycle order from:
   - `dark -> young -> light`
   to:
   - `dark -> young -> light -> sport`
4. Build `sport` as a complete CSS token set that all current screens can consume automatically.
5. Update the Settings page, translation strings, accessibility test coverage, and user-facing feature docs in the same change.

## Why This Approach

### The request is additive, not architectural

The user asked for one more theme. The current system already supports themed tokens cleanly. Adding another token block is the shortest path that fully solves the problem.

### Reusing the existing theme pipeline minimizes regression risk

If `sport` is added through the same enum, CSS variables, Settings preview, and test pipeline as the other themes, every existing screen continues to work without per-page rewrites.

### The visual direction should feel athletic without becoming neon

The reference palette is energetic, but SkillFlow is still a productivity app used for longer sessions. The theme should therefore lean toward:

- pale mint and stadium-turf background layers
- deep athletic green for primary actions
- restrained teal-green secondary emphasis
- readable dark text and visible borders on light surfaces

That preserves the "sport" signal without creating an exhausting, high-saturation UI.

## Runtime Behavior

### Theme model

- `sport` becomes a valid `Theme` value in the frontend.
- existing local storage continues to work
- no migration is required for stored values because this is an additive enum expansion
- default theme remains unchanged

### Theme cycling

- sidebar/topbar shortcut theme cycling uses the updated four-theme order
- theme switch titles use existing label plumbing and therefore pick up the new label automatically once `THEME_LABELS` is updated

### Visual tokens

`[data-theme="sport"]` will define the same token surface already used by the app:

- background layers
- text colors
- accent colors
- button colors
- border colors
- semantic status colors
- shadows, glow, shell, and active state tokens

Because the app already consumes these variables, pages should switch into `sport` automatically.

## UX Impact

Users will see:

- a fourth preview card in **Settings -> Appearance Theme**
- a fourth stop in the quick theme toggle cycle
- sport-oriented green/mint styling applied consistently across the shell and content surfaces

Users will not need to restart the app. The selected theme will continue to persist in `localStorage`.

## Testing Strategy

Update `cmd/skillflow/frontend/tests/themeContrast.test.mjs` so `sport` is included in the existing contrast assertions for:

- muted text on base and elevated surfaces
- primary button label contrast
- border visibility
- active surface distinction
- active text readability

This preserves the current standard that new themes must be visually distinctive without sacrificing readability.

## Documentation Impact

This is a meaningful user-facing feature change. Update:

- `docs/features.md`
- `docs/features_zh.md`

to reflect:

- four themes instead of three
- the new `sport` visual preset
- the updated quick-cycle order

`README.md` and `README_zh.md` do not need changes because they already describe the product at the coarse-grained "multiple themes" level, which remains accurate.

## Scope

- `cmd/skillflow/frontend/src/hooks/useTheme.ts`
- `cmd/skillflow/frontend/src/style.css`
- `cmd/skillflow/frontend/src/pages/Settings.tsx`
- `cmd/skillflow/frontend/src/i18n/en.ts`
- `cmd/skillflow/frontend/src/i18n/zh.ts`
- `cmd/skillflow/frontend/tests/themeContrast.test.mjs`
- `docs/features.md`
- `docs/features_zh.md`

## Out Of Scope

- backend config persistence changes
- theme-specific component rewrites
- a generic theme registry refactor
- dark-mode behavior changes
- README marketing copy changes
