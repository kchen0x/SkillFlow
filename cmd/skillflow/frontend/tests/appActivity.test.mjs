import test from 'node:test'
import assert from 'node:assert/strict'
import {
  createAppActivityState,
  reduceAppActivityState,
  BACKGROUND_MEMORY_TRIM_DELAY_MS,
} from '../.tmp-tests/appActivity.js'

test('createAppActivityState starts in the foreground without trimming', () => {
  assert.deepEqual(createAppActivityState(), {
    windowVisible: true,
    documentVisible: true,
    focused: true,
    foreground: true,
    trimScheduled: false,
    trimmed: false,
    resumeToken: 0,
  })
})

test('backgrounding schedules trimming but does not trim until timeout elapses', () => {
  const hidden = reduceAppActivityState(createAppActivityState(), {
    type: 'window_visibility_changed',
    visible: false,
  })

  assert.equal(hidden.foreground, false)
  assert.equal(hidden.trimScheduled, true)
  assert.equal(hidden.trimmed, false)
  assert.equal(BACKGROUND_MEMORY_TRIM_DELAY_MS > 0, true)

  const trimmed = reduceAppActivityState(hidden, {
    type: 'trim_timeout_elapsed',
  })

  assert.equal(trimmed.trimmed, true)
  assert.equal(trimmed.trimScheduled, false)
  assert.equal(trimmed.foreground, false)
})

test('returning to the foreground clears trimmed state and bumps the resume token', () => {
  const trimmed = reduceAppActivityState(
    reduceAppActivityState(createAppActivityState(), {
      type: 'window_visibility_changed',
      visible: false,
    }),
    { type: 'trim_timeout_elapsed' },
  )

  const resumed = reduceAppActivityState(trimmed, {
    type: 'focus_changed',
    focused: true,
  })

  assert.equal(resumed.windowVisible, true)
  assert.equal(resumed.foreground, true)
  assert.equal(resumed.trimmed, false)
  assert.equal(resumed.trimScheduled, false)
  assert.equal(resumed.resumeToken, 1)
})

test('window visibility restore clears trimmed state even when focus events do not fire', () => {
  const trimmed = reduceAppActivityState(
    reduceAppActivityState(
      reduceAppActivityState(createAppActivityState(), {
        type: 'window_visibility_changed',
        visible: false,
      }),
      {
        type: 'focus_changed',
        focused: false,
      },
    ),
    { type: 'trim_timeout_elapsed' },
  )

  const resumed = reduceAppActivityState(trimmed, {
    type: 'window_visibility_changed',
    visible: true,
  })

  assert.equal(resumed.windowVisible, true)
  assert.equal(resumed.documentVisible, true)
  assert.equal(resumed.focused, true)
  assert.equal(resumed.foreground, true)
  assert.equal(resumed.trimmed, false)
  assert.equal(resumed.trimScheduled, false)
  assert.equal(resumed.resumeToken, 1)
})
