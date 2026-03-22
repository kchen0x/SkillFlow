import test from 'node:test'
import assert from 'node:assert/strict'
import { buildSettingsPathRows } from '../.tmp-tests/src/lib/settingsPaths.js'

test('settingsPaths exposes repo cache field and app data action', () => {
  const rows = buildSettingsPathRows(
    { repoCacheDir: '/tmp/repo-cache', skillsStorageDir: '/tmp/legacy-skills' },
    '/tmp/app-data',
  )

  assert.deepEqual(
    rows.map(row => row.key),
    ['appDataDir', 'repoCacheDir'],
  )
  assert.equal(rows[0].action, 'open')
  assert.equal(rows[0].value, '/tmp/app-data')
  assert.equal(rows[1].action, 'pick')
  assert.equal(rows[1].value, '/tmp/repo-cache')
})

test('settingsPaths omits legacy skills directory row', () => {
  const rows = buildSettingsPathRows(
    { repoCacheDir: '/custom/cache/repos', skillsStorageDir: '/tmp/legacy-skills' },
    '/tmp/app-data',
  )

  assert.equal(rows.some(row => row.key === 'appDataDir'), true)
  assert.equal(rows.some(row => row.key === 'repoCacheDir'), true)
  assert.equal(rows.some(row => row.key === 'skillsStorageDir'), false)
})
