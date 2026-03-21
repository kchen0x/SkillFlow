import test from 'node:test'
import assert from 'node:assert/strict'
import { orderCloudProviders } from '../.tmp-tests/cloudProviderOrder.js'

test('orderCloudProviders keeps git pinned first and preserves backend order for others', () => {
  assert.deepEqual(
    orderCloudProviders([
      { name: 'aliyun' },
      { name: 'huawei' },
      { name: 'git' },
      { name: 'tencent' },
    ]),
    [
      { name: 'git' },
      { name: 'aliyun' },
      { name: 'huawei' },
      { name: 'tencent' },
    ],
  )
})

test('orderCloudProviders leaves backend order untouched when git is absent', () => {
  assert.deepEqual(
    orderCloudProviders([
      { name: 'aws' },
      { name: 'huawei' },
      { name: 'google' },
    ]),
    [
      { name: 'aws' },
      { name: 'huawei' },
      { name: 'google' },
    ],
  )
})
