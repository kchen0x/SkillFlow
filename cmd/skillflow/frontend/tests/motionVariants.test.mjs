import test from 'node:test'
import assert from 'node:assert/strict'
import {
  pageVariants,
  shouldAnimateSkillCards,
  shouldAnimateSkillGridIntro,
} from '../.tmp-tests/src/lib/motionVariants.js'

test('pageVariants uses fade-only transitions for route changes', () => {
  assert.deepEqual(pageVariants.initial, { opacity: 0 })
  assert.deepEqual(pageVariants.animate, {
    opacity: 1,
    transition: { duration: 0.16, ease: 'easeOut' },
  })
  assert.deepEqual(pageVariants.exit, {
    opacity: 0,
    transition: { duration: 0.1 },
  })
})

test('shouldAnimateSkillCards disables per-card motion for dense lists', () => {
  assert.equal(shouldAnimateSkillCards(18), true)
  assert.equal(shouldAnimateSkillCards(19), false)
})

test('shouldAnimateSkillGridIntro only runs on the first populated render', () => {
  assert.equal(shouldAnimateSkillGridIntro(6, false), true)
  assert.equal(shouldAnimateSkillGridIntro(6, true), false)
  assert.equal(shouldAnimateSkillGridIntro(24, false), false)
})
