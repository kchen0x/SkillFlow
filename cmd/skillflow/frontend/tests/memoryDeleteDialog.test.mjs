import test from 'node:test'
import assert from 'node:assert/strict'
import { buildModuleDeletePreview } from '../.tmp-tests/src/lib/memoryDeleteDialog.js'

test('buildModuleDeletePreview keeps the first non-empty lines for delete confirmation', () => {
  assert.equal(
    buildModuleDeletePreview('\n\nfirst line\n\nsecond line\nthird line\nfourth line\n'),
    'first line\nsecond line\nthird line',
  )
})

test('buildModuleDeletePreview returns empty string for blank content', () => {
  assert.equal(buildModuleDeletePreview(' \n\t '), '')
})
