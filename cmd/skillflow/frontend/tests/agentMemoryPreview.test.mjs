import test from 'node:test'
import assert from 'node:assert/strict'
import { buildAgentMemoryEntries, filterAgentMemoryEntries } from '../.tmp-tests/src/lib/agentMemoryPreview.js'

test('buildAgentMemoryEntries keeps main memory first and sorts rules', () => {
  const entries = buildAgentMemoryEntries({
    memoryPath: '/tmp/AGENTS.md',
    mainExists: true,
    mainContent: 'Use Go and write tests.',
    rulesDir: '/tmp/rules',
    rulesDirExists: true,
    rules: [
      { name: 'z-last.md', path: '/tmp/rules/z-last.md', content: 'last', managed: false },
      { name: 'sf-beta.md', path: '/tmp/rules/sf-beta.md', content: 'beta', managed: true },
      { name: 'sf-alpha.md', path: '/tmp/rules/sf-alpha.md', content: 'alpha', managed: true },
    ],
  })

  assert.deepEqual(
    entries.map(entry => entry.title),
    ['Main Memory', 'sf-alpha.md', 'sf-beta.md', 'z-last.md'],
  )
  assert.equal(entries[0].kind, 'main')
  assert.equal(entries[1].managed, true)
  assert.equal(entries[3].managed, false)
})

test('filterAgentMemoryEntries matches title and content', () => {
  const entries = buildAgentMemoryEntries({
    memoryPath: '/tmp/AGENTS.md',
    mainExists: true,
    mainContent: 'Focus on Go code reviews.',
    rulesDir: '/tmp/rules',
    rulesDirExists: true,
    rules: [
      { name: 'sf-testing.md', path: '/tmp/rules/sf-testing.md', content: 'Always add regression tests.', managed: true },
      { name: 'custom-notes.md', path: '/tmp/rules/custom-notes.md', content: 'Product context.', managed: false },
    ],
  })

  assert.deepEqual(
    filterAgentMemoryEntries(entries, 'regression').map(entry => entry.title),
    ['sf-testing.md'],
  )

  assert.deepEqual(
    filterAgentMemoryEntries(entries, 'go').map(entry => entry.title),
    ['Main Memory'],
  )
})
