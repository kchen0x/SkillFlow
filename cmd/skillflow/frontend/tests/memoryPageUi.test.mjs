import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('../src/pages/Memory.tsx', import.meta.url), 'utf8')

test('memory page uses shared status entries for both main and module cards', () => {
  assert.match(source, /buildMemoryPushStatusEntries/)
  assert.match(source, /const memoryStatusEntries = buildMemoryPushStatusEntries\(availableAgents, pushStatuses\)/)
  assert.match(source, /memoryStatusEntries\.map\(entry => \{/)
})

test('memory page drawer is fixed to the right edge with widened metrics', () => {
  assert.match(source, /const drawerMetrics = getMemoryDrawerMetrics\(window\.innerWidth\)/)
  assert.match(source, /position: 'fixed'/)
  assert.match(source, /right: 0/)
  assert.match(source, /width: drawerMetrics\.width/)
})
