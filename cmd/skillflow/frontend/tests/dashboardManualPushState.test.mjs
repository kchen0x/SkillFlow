import test from 'node:test'
import assert from 'node:assert/strict'
import {
  createDashboardManualPushState,
  toggleDashboardManualPushAgent,
  toggleDashboardManualPushSkill,
  toggleDashboardManualPushAllVisible,
  syncDashboardManualPushVisibleSelection,
  isDashboardManualPushReady,
} from '../.tmp-tests/src/lib/dashboardManualPushState.js'

test('createDashboardManualPushState starts empty', () => {
  assert.deepEqual(createDashboardManualPushState(), {
    selectedAgents: [],
    selectedSkillIDs: [],
  })
})

test('toggleDashboardManualPushAgent adds and removes agents', () => {
  const added = toggleDashboardManualPushAgent([], 'codex')
  assert.deepEqual(added, ['codex'])

  const removed = toggleDashboardManualPushAgent(added, 'codex')
  assert.deepEqual(removed, [])
})

test('toggleDashboardManualPushSkill adds and removes skills', () => {
  const added = toggleDashboardManualPushSkill([], 'skill-1')
  assert.deepEqual(added, ['skill-1'])

  const removed = toggleDashboardManualPushSkill(added, 'skill-1')
  assert.deepEqual(removed, [])
})

test('toggleDashboardManualPushAllVisible selects and deselects visible skills only', () => {
  const selected = toggleDashboardManualPushAllVisible(['skill-0'], ['skill-1', 'skill-2'])
  assert.deepEqual(selected, ['skill-0', 'skill-1', 'skill-2'])

  const deselected = toggleDashboardManualPushAllVisible(selected, ['skill-1', 'skill-2'])
  assert.deepEqual(deselected, ['skill-0'])
})

test('syncDashboardManualPushVisibleSelection keeps only visible skills', () => {
  assert.deepEqual(
    syncDashboardManualPushVisibleSelection(['skill-1', 'skill-2'], ['skill-2', 'skill-3']),
    ['skill-2'],
  )
})

test('isDashboardManualPushReady requires at least one agent and one skill', () => {
  assert.equal(isDashboardManualPushReady({
    selectedAgents: [],
    selectedSkillIDs: [],
  }), false)

  assert.equal(isDashboardManualPushReady({
    selectedAgents: ['codex'],
    selectedSkillIDs: [],
  }), false)

  assert.equal(isDashboardManualPushReady({
    selectedAgents: [],
    selectedSkillIDs: ['skill-1'],
  }), false)

  assert.equal(isDashboardManualPushReady({
    selectedAgents: ['codex'],
    selectedSkillIDs: ['skill-1'],
  }), true)
})
