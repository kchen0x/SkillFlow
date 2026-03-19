import test from 'node:test'
import assert from 'node:assert/strict'
import * as useThemeModule from '../.tmp-tests/theme/useTheme.js'

function createStorage(entries = {}) {
  return {
    getItem(key) {
      return Object.prototype.hasOwnProperty.call(entries, key) ? entries[key] : null
    },
    setItem(key, value) {
      entries[key] = String(value)
    },
    removeItem(key) {
      delete entries[key]
    },
  }
}

test('resolveInitialThemeFromStorage defaults to young when no saved theme exists', () => {
  assert.equal(typeof useThemeModule.resolveInitialThemeFromStorage, 'function')
  assert.equal(useThemeModule.resolveInitialThemeFromStorage(createStorage()), 'young')
})

test('resolveInitialThemeFromStorage prefers the current storage key', () => {
  assert.equal(
    useThemeModule.resolveInitialThemeFromStorage(createStorage({
      'skillflow-theme-v2': 'dark',
      'skillflow-theme': 'light',
    })),
    'dark',
  )
})

test('resolveInitialThemeFromStorage migrates legacy light to young', () => {
  assert.equal(
    useThemeModule.resolveInitialThemeFromStorage(createStorage({
      'skillflow-theme': 'light',
    })),
    'young',
  )
})
