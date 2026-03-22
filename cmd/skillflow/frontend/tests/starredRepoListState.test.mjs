import test from 'node:test'
import assert from 'node:assert/strict'
import { getStarredRepoListState } from '../.tmp-tests/src/lib/starredRepoListState.js'

test('getStarredRepoListState keeps repo detail in loading state while fetch is in flight', () => {
  assert.equal(
    getStarredRepoListState({ isLoading: true, skillCount: 0 }),
    'loading',
  )
})

test('getStarredRepoListState returns empty only after loading completes with no skills', () => {
  assert.equal(
    getStarredRepoListState({ isLoading: false, skillCount: 0 }),
    'empty',
  )
})

test('getStarredRepoListState returns ready when the repo contains skills', () => {
  assert.equal(
    getStarredRepoListState({ isLoading: false, skillCount: 2 }),
    'ready',
  )
})
