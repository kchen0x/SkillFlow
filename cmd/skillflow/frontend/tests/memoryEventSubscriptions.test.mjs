import test from 'node:test'
import assert from 'node:assert/strict'
import { createMemoryEventSubscriptions } from '../.tmp-tests/memoryEventSubscriptions.js'

test('memory event subscriptions reload on memory content changes', () => {
  const load = () => {}
  const subscriptions = createMemoryEventSubscriptions(load)

  assert.deepEqual(subscriptions.map(([eventName]) => eventName), [
    'memory:content:changed',
  ])
  assert.ok(subscriptions.every(([, handler]) => handler === load))
})
