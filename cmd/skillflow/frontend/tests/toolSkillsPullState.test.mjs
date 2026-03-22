import test from 'node:test'
import assert from 'node:assert/strict'
import {
  createToolSkillsPullState,
  toggleToolSkillsPullPath,
  toggleToolSkillsPullAllVisible,
  toggleToolSkillsPullVisibleNotImported,
  syncToolSkillsPullVisibleSelection,
  getToolSkillsVisibleNotImportedPaths,
  isToolSkillsPullReady,
} from '../.tmp-tests/toolSkillsPullState.js'

test('createToolSkillsPullState starts with default category and empty selection', () => {
  assert.deepEqual(createToolSkillsPullState('Default'), {
    targetCategory: 'Default',
    selectedPaths: [],
  })
})

test('toggleToolSkillsPullPath adds and removes scanned skill paths', () => {
  const added = toggleToolSkillsPullPath([], '/tmp/alpha')
  assert.deepEqual(added, ['/tmp/alpha'])

  const removed = toggleToolSkillsPullPath(added, '/tmp/alpha')
  assert.deepEqual(removed, [])
})

test('toggleToolSkillsPullAllVisible selects and deselects visible scanned paths', () => {
  const selected = toggleToolSkillsPullAllVisible(['/tmp/base'], ['/tmp/a', '/tmp/b'])
  assert.deepEqual(selected, ['/tmp/base', '/tmp/a', '/tmp/b'])

  const deselected = toggleToolSkillsPullAllVisible(selected, ['/tmp/a', '/tmp/b'])
  assert.deepEqual(deselected, ['/tmp/base'])
})

test('getToolSkillsVisibleNotImportedPaths returns only not imported visible entries', () => {
  assert.deepEqual(
    getToolSkillsVisibleNotImportedPaths([
      { path: '/tmp/a', imported: false },
      { path: '/tmp/b', imported: true },
      { path: '/tmp/c', imported: false },
    ]),
    ['/tmp/a', '/tmp/c'],
  )
})

test('toggleToolSkillsPullVisibleNotImported toggles the visible not imported paths', () => {
  const selected = toggleToolSkillsPullVisibleNotImported(['/tmp/base'], ['/tmp/a', '/tmp/c'])
  assert.deepEqual(selected, ['/tmp/base', '/tmp/a', '/tmp/c'])

  const deselected = toggleToolSkillsPullVisibleNotImported(selected, ['/tmp/a', '/tmp/c'])
  assert.deepEqual(deselected, ['/tmp/base'])
})

test('syncToolSkillsPullVisibleSelection keeps only visible paths', () => {
  assert.deepEqual(
    syncToolSkillsPullVisibleSelection(['/tmp/a', '/tmp/b'], ['/tmp/b', '/tmp/c']),
    ['/tmp/b'],
  )
})

test('isToolSkillsPullReady requires at least one selected path', () => {
  assert.equal(isToolSkillsPullReady({ selectedPaths: [] }), false)
  assert.equal(isToolSkillsPullReady({ selectedPaths: ['/tmp/a'] }), true)
})
