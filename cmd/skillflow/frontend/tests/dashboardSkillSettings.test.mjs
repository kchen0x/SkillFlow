import test from 'node:test'
import assert from 'node:assert/strict'
import {
  readDashboardSkillSettings,
  toggleDashboardAutoPushAgent,
  createDashboardSkillEventSubscriptions,
  createToolSkillsEventSubscriptions,
  listDashboardToolbarActionKeys,
} from '../.tmp-tests/dashboardSkillSettings.js'

test('readDashboardSkillSettings defaults auto update to false', () => {
  assert.deepEqual(readDashboardSkillSettings(undefined), {
    autoPushAgents: [],
    autoUpdateSkills: false,
  })
})

test('readDashboardSkillSettings reads auto update and auto push agent config', () => {
  assert.deepEqual(readDashboardSkillSettings({
    autoPushAgents: ['codex'],
    autoUpdateSkills: true,
  }), {
    autoPushAgents: ['codex'],
    autoUpdateSkills: true,
  })
})

test('toggleDashboardAutoPushAgent adds and removes a target agent', () => {
  const added = toggleDashboardAutoPushAgent(['codex'], 'claude-code')
  assert.deepEqual(added, ['codex', 'claude-code'])

  const removed = toggleDashboardAutoPushAgent(added, 'codex')
  assert.deepEqual(removed, ['claude-code'])
})

test('dashboard skill event subscriptions reload on update.available and skills.updated', () => {
  const load = () => {}
  const subscriptions = createDashboardSkillEventSubscriptions(load)

  assert.deepEqual(subscriptions.map(([eventName]) => eventName), [
    'update.available',
    'skills.updated',
  ])
  assert.ok(subscriptions.every(([, handler]) => handler === load))
})

test('tool skills event subscriptions reload on skills.updated', () => {
  const load = () => {}
  const subscriptions = createToolSkillsEventSubscriptions(load)

  assert.deepEqual(subscriptions.map(([eventName]) => eventName), [
    'skills.updated',
  ])
  assert.ok(subscriptions.every(([, handler]) => handler === load))
})

test('dashboard toolbar actions replace remote install with auto update', () => {
  assert.deepEqual(listDashboardToolbarActionKeys(), [
    'update',
    'batchDelete',
    'import',
    'autoUpdate',
  ])
})
