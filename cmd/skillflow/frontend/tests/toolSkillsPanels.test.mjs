import test from 'node:test'
import assert from 'node:assert/strict'
import {
  getDefaultToolSkillsPanel,
  getVisibleToolSkillsResultCount,
  filterToolSkillsPanelContent,
} from '../.tmp-tests/src/lib/toolSkillsPanels.js'

test('getDefaultToolSkillsPanel returns skills', () => {
  assert.equal(getDefaultToolSkillsPanel(), 'skills')
})

test('getVisibleToolSkillsResultCount uses only skill entries for skills panel', () => {
  assert.equal(
    getVisibleToolSkillsResultCount({
      activePanel: 'skills',
      filteredPushSkills: [{ name: 'alpha' }, { name: 'beta' }],
      filteredScanOnlySkills: [{ name: 'gamma' }],
      filteredMemoryEntries: [{ title: 'Main Memory' }, { title: 'sf-rules.md' }],
    }),
    3,
  )
})

test('getVisibleToolSkillsResultCount uses only memory entries for memory panel', () => {
  assert.equal(
    getVisibleToolSkillsResultCount({
      activePanel: 'memory',
      filteredPushSkills: [{ name: 'alpha' }, { name: 'beta' }],
      filteredScanOnlySkills: [{ name: 'gamma' }],
      filteredMemoryEntries: [{ title: 'Main Memory' }, { title: 'sf-rules.md' }],
    }),
    2,
  )
})

test('filterToolSkillsPanelContent scopes search to the active panel', () => {
  const skillsPanel = filterToolSkillsPanelContent({
    activePanel: 'skills',
    search: 'memory',
    sortOrder: 'asc',
    pushSkills: [
      { name: 'Go Review' },
      { name: 'Skill Memory Sync' },
    ],
    scanOnlySkills: [
      { name: 'Agent Memory Audit' },
      { name: 'Prompt Notes' },
    ],
    memoryEntries: [
      { title: 'Main Memory', path: '/tmp/AGENTS.md', content: 'Go conventions' },
      { title: 'sf-testing.md', path: '/tmp/rules/sf-testing.md', content: 'Always test' },
    ],
  })

  assert.deepEqual(
    skillsPanel.filteredPushSkills.map(item => item.name),
    ['Skill Memory Sync'],
  )
  assert.deepEqual(
    skillsPanel.filteredScanOnlySkills.map(item => item.name),
    ['Agent Memory Audit'],
  )
  assert.deepEqual(skillsPanel.filteredMemoryEntries, [])

  const memoryPanel = filterToolSkillsPanelContent({
    activePanel: 'memory',
    search: 'go',
    sortOrder: 'asc',
    pushSkills: [
      { name: 'Go Review' },
      { name: 'Skill Memory Sync' },
    ],
    scanOnlySkills: [
      { name: 'Agent Memory Audit' },
      { name: 'Prompt Notes' },
    ],
    memoryEntries: [
      { title: 'Main Memory', path: '/tmp/AGENTS.md', content: 'Go conventions' },
      { title: 'sf-testing.md', path: '/tmp/rules/sf-testing.md', content: 'Always test' },
    ],
  })

  assert.deepEqual(memoryPanel.filteredPushSkills, [])
  assert.deepEqual(memoryPanel.filteredScanOnlySkills, [])
  assert.deepEqual(
    memoryPanel.filteredMemoryEntries.map(item => item.title),
    ['Main Memory'],
  )
})
