import test from 'node:test'
import assert from 'node:assert/strict'
import { subscribeToEvents } from '../.tmp-tests/src/lib/wailsEvents.js'

test('subscribeToEvents unsubscribes every registered Wails event on cleanup', () => {
  const calls = []
  const unsubscribeCalls = []

  const cleanup = subscribeToEvents((eventName, handler) => {
    calls.push({ eventName, handlerType: typeof handler })
    return () => unsubscribeCalls.push(eventName)
  }, [
    ['backup.progress', () => {}],
    ['backup.completed', () => {}],
    ['backup.failed', () => {}],
  ])

  assert.deepEqual(calls.map(call => call.eventName), [
    'backup.progress',
    'backup.completed',
    'backup.failed',
  ])

  cleanup()

  assert.deepEqual(unsubscribeCalls, [
    'backup.progress',
    'backup.completed',
    'backup.failed',
  ])
})

test('subscribeToEvents ignores missing unsubscribe callbacks', () => {
  const cleanup = subscribeToEvents((eventName) => {
    if (eventName === 'backup.progress') return () => {}
    return undefined
  }, [
    ['backup.progress', () => {}],
    ['backup.failed', () => {}],
  ])

  assert.doesNotThrow(() => cleanup())
})
