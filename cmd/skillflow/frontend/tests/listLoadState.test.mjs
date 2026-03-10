import test from 'node:test'
import assert from 'node:assert/strict'
import { getListLoadState } from '../.tmp-tests/listLoadState.js'

test('getListLoadState stays loading while the first fetch is still in flight', () => {
  assert.equal(
    getListLoadState({ isLoading: true, itemCount: 0 }),
    'loading',
  )
})

test('getListLoadState reports empty only after loading finishes with no items', () => {
  assert.equal(
    getListLoadState({ isLoading: false, itemCount: 0 }),
    'empty',
  )
})

test('getListLoadState reports ready when items are available', () => {
  assert.equal(
    getListLoadState({ isLoading: false, itemCount: 3 }),
    'ready',
  )
})
