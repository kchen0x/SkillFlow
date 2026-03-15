# Prompt Editor Layout And Link Flow Design

**Date:** 2026-03-15

## Context

The prompt editor currently uses a large freeform textarea for web links and a tall content textarea that can push the action bar outside the visible SkillFlow window. This makes the save action inaccessible in smaller desktop windows.

## Decision

1. Keep persisted prompt metadata compatible with the existing backend contract:
   - prompt content remains `system.md`
   - web links remain serialized as markdown lines when saved
2. Refactor frontend draft state from raw `webLinksMarkdown` text into structured `PromptWebLink[]`.
3. Replace the web-link textarea with a single-line markdown input plus an add action:
   - users enter one markdown link such as `[Doc](https://example.com)`
   - clicking add parses and appends a structured link
   - the input clears after a successful add
   - saved links render below as hyperlink chips using the markdown label text
4. Constrain the prompt content editor to the dialog viewport and allow vertical scrolling instead of expanding beyond the desktop window.

## Scope

- `PromptEditorDialog` layout and local state
- prompt rich-content helpers used by the prompt page
- prompt page draft creation / hydration / save serialization
- i18n copy for the revised link workflow
- feature documentation for the prompt editor

## Out Of Scope

- backend prompt schema changes
- import/export format changes
- prompt card layout changes outside the editor dialog
