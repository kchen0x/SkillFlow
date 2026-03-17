import test from 'node:test'
import assert from 'node:assert/strict'
import { applyPromptImportDecision } from '../.tmp-tests/promptImportConflicts.js'

const conflicts = [
  { name: 'Prompt A' },
  { name: 'Prompt B' },
  { name: 'Prompt C' },
]

test('applyPromptImportDecision marks one conflict by default', () => {
  assert.deepEqual(
    applyPromptImportDecision(conflicts, [], 'overwrite', false),
    {
      overwriteNames: ['Prompt A'],
      remainingConflicts: [
        { name: 'Prompt B' },
        { name: 'Prompt C' },
      ],
    },
  )
})

test('applyPromptImportDecision skips one conflict by default', () => {
  assert.deepEqual(
    applyPromptImportDecision(conflicts, [], 'skip', false),
    {
      overwriteNames: [],
      remainingConflicts: [
        { name: 'Prompt B' },
        { name: 'Prompt C' },
      ],
    },
  )
})

test('applyPromptImportDecision applies same action to remaining conflicts when requested', () => {
  assert.deepEqual(
    applyPromptImportDecision(conflicts, [], 'overwrite', true),
    {
      overwriteNames: ['Prompt A', 'Prompt B', 'Prompt C'],
      remainingConflicts: [],
    },
  )

  assert.deepEqual(
    applyPromptImportDecision(conflicts, ['Prompt A'], 'skip', true),
    {
      overwriteNames: ['Prompt A'],
      remainingConflicts: [],
    },
  )
})
