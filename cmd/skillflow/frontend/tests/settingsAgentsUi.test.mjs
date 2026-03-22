import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('../src/pages/Settings.tsx', import.meta.url), 'utf8')

test('settings agents tab uses AnimatedDialog for custom agent creation', () => {
  assert.match(source, /import AnimatedDialog from '\.\.\/components\/ui\/AnimatedDialog'/)
  assert.match(source, /<AnimatedDialog open=\{[^}]+\}/)
})

test('settings agents tab exposes separate skill and memory path sections', () => {
  assert.match(source, /t\('settings\.skillPathsSection'\)/)
  assert.match(source, /t\('settings\.memoryPathsSection'\)/)
})

test('settings custom agent dialog includes required fields', () => {
  assert.match(source, /t\('settings\.addCustomToolDialogTitle'\)/)
  assert.match(source, /t\('settings\.addCustomToolSave'\)/)
  assert.match(source, /createEmptyCustomAgentDraft/)
})

test('settings agents tab no longer directly persists custom agent edits', () => {
  assert.doesNotMatch(source, /AddCustomAgent\(/)
  assert.doesNotMatch(source, /RemoveCustomAgent\(/)
})
