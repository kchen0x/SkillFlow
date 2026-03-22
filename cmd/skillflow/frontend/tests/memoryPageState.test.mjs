import test from 'node:test'
import assert from 'node:assert/strict'
import {
  createMemoryBatchPushState,
  getMemoryAutoSyncMode,
  getMemoryPushConfigForAutoSyncMode,
  isMemoryBatchPushReady,
  isMemoryBatchSelected,
  toggleMemoryBatchAgent,
  toggleMemoryBatchModule,
} from '../.tmp-tests/memoryPageState.js'

test('getMemoryAutoSyncMode maps stored push config to off merge and takeover', () => {
  assert.equal(getMemoryAutoSyncMode(undefined), 'off')
  assert.equal(getMemoryAutoSyncMode({ mode: 'merge', autoPush: false }), 'off')
  assert.equal(getMemoryAutoSyncMode({ mode: 'merge', autoPush: true }), 'merge')
  assert.equal(getMemoryAutoSyncMode({ mode: 'takeover', autoPush: true }), 'takeover')
})

test('getMemoryPushConfigForAutoSyncMode maps ui mode back to stored config', () => {
  assert.deepEqual(getMemoryPushConfigForAutoSyncMode('off'), {
    mode: 'merge',
    autoPush: false,
  })
  assert.deepEqual(getMemoryPushConfigForAutoSyncMode('merge'), {
    mode: 'merge',
    autoPush: true,
  })
  assert.deepEqual(getMemoryPushConfigForAutoSyncMode('takeover'), {
    mode: 'takeover',
    autoPush: true,
  })
})

test('createMemoryBatchPushState starts with empty temporary selections', () => {
  assert.deepEqual(createMemoryBatchPushState(), {
    mode: 'merge',
    selectedAgents: [],
    selectedModules: [],
  })
})

test('main memory stays selected while module and agent selections toggle independently', () => {
  const withModule = toggleMemoryBatchModule([], 'testing-rules')
  assert.deepEqual(withModule, ['testing-rules'])
  assert.equal(isMemoryBatchSelected('main', '', withModule), true)
  assert.equal(isMemoryBatchSelected('module', 'testing-rules', withModule), true)
  assert.equal(isMemoryBatchSelected('module', 'style', withModule), false)

  const withoutModule = toggleMemoryBatchModule(withModule, 'testing-rules')
  assert.deepEqual(withoutModule, [])

  const withAgent = toggleMemoryBatchAgent([], 'codex')
  assert.deepEqual(withAgent, ['codex'])
  assert.deepEqual(toggleMemoryBatchAgent(withAgent, 'codex'), [])
})

test('isMemoryBatchPushReady requires at least one module and one agent', () => {
  assert.equal(isMemoryBatchPushReady(createMemoryBatchPushState()), false)
  assert.equal(isMemoryBatchPushReady({
    mode: 'merge',
    selectedAgents: ['codex'],
    selectedModules: [],
  }), false)
  assert.equal(isMemoryBatchPushReady({
    mode: 'merge',
    selectedAgents: [],
    selectedModules: ['testing-rules'],
  }), false)
  assert.equal(isMemoryBatchPushReady({
    mode: 'merge',
    selectedAgents: ['codex'],
    selectedModules: ['testing-rules'],
  }), true)
})
