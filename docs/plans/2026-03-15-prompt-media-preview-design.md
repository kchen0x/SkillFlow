# Prompt Editor Image Preview Design

**Date:** 2026-03-15

## Context

Prompt cards currently expose too much secondary media. Users only need richer media interactions while editing a prompt, not while scanning the prompt list.

## Decision

1. Prompt cards will remove both image thumbnails and web-link chips.
2. Prompt cards will stay text-first: name, description, content preview, category, copy, and delete.
3. Prompt editor image thumbnails will open an in-app enlarged preview overlay instead of opening the external browser.
4. Image and web-link display blocks stay where they are, but their add-input rows move to a shared attachment area at the end of the editor body to preserve reading flow.
5. Saved images expose a delete affordance directly on the thumbnail.
6. The enlarged preview will exist only inside the editor flow and will not change prompt persistence or metadata format.

## Scope

- `PromptEditorDialog` preview interaction
- `Prompts` card rendering cleanup
- frontend docs for prompt cards and prompt editor behavior

## Out Of Scope

- gallery navigation between images
- fullscreen browser or OS-native preview
- changing stored image metadata
