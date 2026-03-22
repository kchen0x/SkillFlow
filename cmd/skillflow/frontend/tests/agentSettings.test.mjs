import test from 'node:test'
import assert from 'node:assert/strict'
import {
  buildCustomAgentProfile,
  createEmptyCustomAgentDraft,
  validateCustomAgentDraft,
} from '../.tmp-tests/src/lib/agentSettings.js'

test('createEmptyCustomAgentDraft returns blank required fields', () => {
  assert.deepEqual(createEmptyCustomAgentDraft(), {
    name: '',
    pushDir: '',
    memoryPath: '',
    rulesDir: '',
  })
})

test('buildCustomAgentProfile seeds scanDirs from pushDir', () => {
  assert.deepEqual(
    buildCustomAgentProfile({
      name: '  custom-agent  ',
      pushDir: '  /tmp/skills  ',
      memoryPath: '  /tmp/agent/AGENTS.md  ',
      rulesDir: '  /tmp/agent/rules  ',
    }),
    {
      name: 'custom-agent',
      pushDir: '/tmp/skills',
      scanDirs: ['/tmp/skills'],
      memoryPath: '/tmp/agent/AGENTS.md',
      rulesDir: '/tmp/agent/rules',
      enabled: true,
      custom: true,
    },
  )
})

test('validateCustomAgentDraft rejects duplicate names', () => {
  assert.deepEqual(
    validateCustomAgentDraft(
      {
        name: '  Codex  ',
        pushDir: '/tmp/skills',
        memoryPath: '/tmp/AGENTS.md',
        rulesDir: '/tmp/rules',
      },
      [{ name: 'codex' }],
    ),
    { ok: false, reason: 'duplicate_name' },
  )
})

test('validateCustomAgentDraft requires all fields', () => {
  assert.deepEqual(
    validateCustomAgentDraft(
      {
        name: 'custom-agent',
        pushDir: '',
        memoryPath: '/tmp/AGENTS.md',
        rulesDir: '/tmp/rules',
      },
      [],
    ),
    { ok: false, reason: 'required_fields' },
  )
})
