import test from 'node:test'
import assert from 'node:assert/strict'
import { summarizePushedAgents } from '../.tmp-tests/skillStatusStrip.js'

test('summarizePushedAgents returns all agents when under the visible cap', () => {
  assert.deepEqual(
    summarizePushedAgents(['codex', 'claude'], 3),
    {
      visibleAgents: ['codex', 'claude'],
      overflowCount: 0,
    },
  )
})

test('summarizePushedAgents truncates visible agents and reports overflow count', () => {
  assert.deepEqual(
    summarizePushedAgents(['codex', 'claude', 'gemini', 'opencode'], 2),
    {
      visibleAgents: ['codex', 'claude'],
      overflowCount: 2,
    },
  )
})

test('summarizePushedAgents handles a zero visible cap', () => {
  assert.deepEqual(
    summarizePushedAgents(['codex', 'claude'], 0),
    {
      visibleAgents: [],
      overflowCount: 2,
    },
  )
})
