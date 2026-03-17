import test from 'node:test'
import assert from 'node:assert/strict'
import {
  DEFAULT_PROXY_TEST_URL,
  normalizeProxyTestTargetURL,
  buildProxyConnectionPayload,
} from '../.tmp-tests/proxyConnection.js'

test('normalizeProxyTestTargetURL defaults github url', () => {
  assert.equal(normalizeProxyTestTargetURL(''), DEFAULT_PROXY_TEST_URL)
  assert.equal(normalizeProxyTestTargetURL('   '), DEFAULT_PROXY_TEST_URL)
})

test('normalizeProxyTestTargetURL trims explicit values', () => {
  assert.equal(
    normalizeProxyTestTargetURL('  https://example.com/docs  '),
    'https://example.com/docs',
  )
})

test('buildProxyConnectionPayload uses current proxy form state', () => {
  assert.deepEqual(
    buildProxyConnectionPayload(' https://example.com/health ', {
      mode: 'manual',
      url: ' http://127.0.0.1:7890 ',
    }),
    {
      targetURL: 'https://example.com/health',
      proxy: {
        mode: 'manual',
        url: 'http://127.0.0.1:7890',
      },
    },
  )
})

test('buildProxyConnectionPayload falls back to no proxy shape', () => {
  assert.deepEqual(
    buildProxyConnectionPayload('', null),
    {
      targetURL: DEFAULT_PROXY_TEST_URL,
      proxy: {
        mode: 'none',
        url: '',
      },
    },
  )
})
