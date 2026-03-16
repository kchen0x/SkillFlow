import test from 'node:test'
import assert from 'node:assert/strict'
import {
  buildPromptExportActions,
  resolvePromptExportNames,
  canExportPromptSelection,
} from '../.tmp-tests/promptExport.js'

const prompts = [
  { name: 'Prompt A', category: 'Default' },
  { name: 'Prompt B', category: 'Writing' },
  { name: 'Prompt C', category: 'Writing' },
]

test('buildPromptExportActions hides duplicate all action', () => {
  assert.deepEqual(
    buildPromptExportActions(null),
    [
      { key: 'all', label: '全部' },
      { key: 'selected', label: '指定' },
    ],
  )
})

test('buildPromptExportActions includes selected category when one is active', () => {
  assert.deepEqual(
    buildPromptExportActions('Writing'),
    [
      { key: 'all', label: '全部' },
      { key: 'current', label: 'Writing' },
      { key: 'selected', label: '指定' },
    ],
  )
})

test('resolvePromptExportNames limits specified export to current filter', () => {
  assert.deepEqual(
    resolvePromptExportNames('selected', prompts, 'Writing', new Set(['Prompt A', 'Prompt B', 'Prompt C'])),
    ['Prompt B', 'Prompt C'],
  )
})

test('resolvePromptExportNames returns all scoped prompts for current export', () => {
  assert.deepEqual(
    resolvePromptExportNames('current', prompts, 'Writing', new Set()),
    ['Prompt B', 'Prompt C'],
  )
  assert.deepEqual(
    resolvePromptExportNames('all', prompts, null, new Set()),
    ['Prompt A', 'Prompt B', 'Prompt C'],
  )
})

test('canExportPromptSelection blocks empty specified selections', () => {
  assert.equal(canExportPromptSelection('selected', new Set()), false)
  assert.equal(canExportPromptSelection('selected', new Set(['Prompt B'])), true)
  assert.equal(canExportPromptSelection('all', new Set()), true)
})
