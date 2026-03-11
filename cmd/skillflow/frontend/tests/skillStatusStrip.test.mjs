import test from 'node:test'
import assert from 'node:assert/strict'
import { summarizePushedTools } from '../.tmp-tests/skillStatusStrip.js'

test('summarizePushedTools returns all tools when under the visible cap', () => {
  assert.deepEqual(
    summarizePushedTools(['codex', 'claude'], 3),
    {
      visibleTools: ['codex', 'claude'],
      overflowCount: 0,
    },
  )
})

test('summarizePushedTools truncates visible tools and reports overflow count', () => {
  assert.deepEqual(
    summarizePushedTools(['codex', 'claude', 'gemini', 'opencode'], 2),
    {
      visibleTools: ['codex', 'claude'],
      overflowCount: 2,
    },
  )
})

test('summarizePushedTools handles a zero visible cap', () => {
  assert.deepEqual(
    summarizePushedTools(['codex', 'claude'], 0),
    {
      visibleTools: [],
      overflowCount: 2,
    },
  )
})
