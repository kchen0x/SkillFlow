# Prompt Import And Export Flow Design

**Date:** 2026-03-17

## Context

The prompt page currently exposes a single toolbar export action that immediately writes the entire prompt library to disk. Import also applies prompt updates by name without surfacing conflicts, which removes user control when an imported prompt would overwrite an existing local prompt.

The requested change keeps the toolbar as the only entry point while expanding it to support scoped export and explicit import conflict handling:

- export a selected subset of prompts
- export all prompts in the currently selected category
- keep imported prompt categories intact
- let the user choose skip or overwrite when imported prompts conflict with existing prompts
- allow one decision to apply to all remaining conflicts in the current import run

## Decision

1. Replace the current one-click export behavior with an inline export bar rendered below the prompt toolbar.
2. Keep the export entry unified under the existing toolbar `Export` button. Clicking it toggles the inline export bar instead of opening a dialog.
3. Keep export copy intentionally short:
   - `全部`
   - `<selected category name>` when a concrete category is selected in the left sidebar
   - `指定`
4. Scope `指定` to the current left-sidebar filter:
   - if the sidebar is on a concrete category, selection is limited to prompts in that category
   - if the sidebar is on all prompts, selection spans the full prompt list
5. Keep the export file format unchanged. Export still writes the existing JSON bundle format and each exported prompt keeps its own `category`, `imageURLs`, and `webLinks`.
6. Change import into a two-step flow:
   - prepare: parse the selected file, classify new prompts and conflicts, and return the conflict list without writing anything
   - complete: create new prompts plus only the conflicts the user chose to overwrite
7. Keep category semantics strict during import:
   - new prompts are created in the category declared by the import file
   - overwrite updates the existing prompt content, metadata, and category to the imported values
   - skip leaves the existing prompt and category untouched
8. Extend the existing conflict dialog with a checkbox labeled `对剩余 {count} 个冲突执行相同操作`. The checkbox applies only to the remaining conflicts in the current import session and is never persisted as a setting.

## Export Flow

1. User clicks the toolbar `Export` button.
2. The page expands an inline export bar below the toolbar controls.
3. The user chooses one of the available scope actions:
   - `全部`
   - `<selected category name>` when the left sidebar is focused on a concrete category
   - `指定`
4. For `全部` and category export, the frontend resolves the matching prompt names immediately and calls the backend export method.
5. For `指定`, the export bar switches into multi-select mode for the current scope and shows the prompt list plus selected count. Export runs only after the user confirms the selection.
6. The backend writes only the requested prompts but keeps the existing bundle schema unchanged.

## Import Flow

1. User clicks the toolbar `Import` button and selects a JSON file.
2. Backend parses the file and returns:
   - imported prompt candidates
   - prompts that can be created directly
   - conflicts where the imported prompt name already exists locally
3. If there are no conflicts, import completes immediately.
4. If conflicts exist, the frontend shows the existing conflict dialog one item at a time with:
   - `跳过`
   - `覆盖`
   - checkbox `对剩余 {count} 个冲突执行相同操作`
5. When the user finishes reviewing conflicts, the frontend sends one final overwrite set back to the backend to complete the import.
6. If the user cancels during conflict review, the temporary import session is discarded and no prompt files are modified.

## Data And API Shape

- Frontend export should pass explicit prompt names to the backend instead of implicit UI state such as "current category".
- Backend should add a scoped prompt export API that accepts zero or more prompt names:
  - empty list means export all prompts
  - non-empty list means export only the named prompts
- Backend should add import session APIs so conflict review can happen before any write:
  - prepare import
  - complete import with an explicit overwrite name list
- The on-disk prompt export/import schema does not change in this work.
- No persisted config changes are required.

## Error Handling

- Export refuses `指定` confirmation when no prompts are selected.
- Import preparation returns validation errors for malformed JSON or invalid prompt records before any prompt write occurs.
- Import completion ignores skipped conflicts and only mutates prompts in the overwrite set plus non-conflicting new prompts.
- Cancelling the export bar or conflict review leaves current prompt data unchanged.

## Scope

- `core/prompt` export/import orchestration
- `cmd/skillflow/app_prompt.go` Wails-facing prompt import/export methods
- prompt page toolbar and inline export bar behavior
- prompt import conflict dialog copy and state handling
- i18n strings and feature/readme documentation updates

## Out Of Scope

- changing the prompt JSON bundle schema
- adding a persistent import preference
- changing prompt editor layout or prompt card visuals outside the new export controls
