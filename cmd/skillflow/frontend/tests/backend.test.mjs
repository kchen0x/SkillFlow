import test from 'node:test'
import assert from 'node:assert/strict'
import { createBackendClient, BackendClientError } from '../.tmp-tests/src/lib/backendClient.js'

test('backend client caches runtime config across calls', async () => {
  let configCalls = 0
  const requests = []
  const client = createBackendClient({
    getConfig: async () => {
      configCalls += 1
      return { baseUrl: 'http://127.0.0.1:9999', token: 'secret' }
    },
    fetchImpl: async (url, init) => {
      requests.push({ url, init })
      return new Response(JSON.stringify({ ok: true, result: ['done'] }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    },
  })

  await client.invoke('ListSkills')
  await client.invoke('ListCategories')

  assert.equal(configCalls, 1)
  assert.equal(requests.length, 2)
  assert.equal(requests[0].url, 'http://127.0.0.1:9999/invoke')
  assert.equal(requests[0].init.headers['X-SkillFlow-Token'], 'secret')
})

test('backend client sends method and params payload', async () => {
  let requestBody = null
  const client = createBackendClient({
    getConfig: async () => ({ baseUrl: 'http://127.0.0.1:9999', token: 'secret' }),
    fetchImpl: async (_url, init) => {
      requestBody = JSON.parse(init.body)
      return new Response(JSON.stringify({ ok: true, result: { success: true } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    },
  })

  const result = await client.invoke('MoveSkillCategory', 'skill-1', 'Tools')

  assert.deepEqual(requestBody, {
    method: 'MoveSkillCategory',
    params: ['skill-1', 'Tools'],
  })
  assert.deepEqual(result, { success: true })
})

test('backend client maps daemon errors to stable exceptions', async () => {
  const client = createBackendClient({
    getConfig: async () => ({ baseUrl: 'http://127.0.0.1:9999', token: 'secret' }),
    fetchImpl: async () => new Response(JSON.stringify({ ok: false, error: 'boom' }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }),
  })

  await assert.rejects(
    () => client.invoke('BackupNow'),
    (error) => {
      assert.equal(error instanceof BackendClientError, true)
      assert.equal(error.message, 'boom')
      assert.equal(error.status, 200)
      return true
    },
  )
})
