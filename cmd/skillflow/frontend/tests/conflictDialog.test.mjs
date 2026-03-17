import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('../src/components/ConflictDialog.tsx', import.meta.url), 'utf8')

test('conflict dialog only completes after conflicts transition from non-empty to empty', () => {
  assert.match(source, /useRef\(/)
  assert.match(source, /hadConflicts/)
  assert.match(source, /if\s*\(\s*hadConflicts\s*&&\s*!hasConflicts\s*\)\s*{\s*onDone\(\)/)
  assert.doesNotMatch(source, /if\s*\(\s*conflicts\.length\s*===\s*0\s*\)\s*{\s*onDone\(\)/)
})
