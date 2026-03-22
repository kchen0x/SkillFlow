import test from 'node:test'
import assert from 'node:assert/strict'
import { copyTextWithDocumentCommand, copyTextWithFallbacks } from '../.tmp-tests/src/lib/clipboardCore.js'

test('copyTextWithFallbacks prefers runtime clipboard when available', async () => {
  let runtimeText = null
  let browserCalled = false

  await copyTextWithFallbacks('prompt body', {
    runtimeWriteText: async (text) => {
      runtimeText = text
      return true
    },
    browserWriteText: async () => {
      browserCalled = true
    },
  })

  assert.equal(runtimeText, 'prompt body')
  assert.equal(browserCalled, false)
})

test('copyTextWithFallbacks prefers document execCommand before async fallbacks', async () => {
  let runtimeCalled = false
  let browserCalled = false

  await copyTextWithFallbacks('prompt body', {
    execCommandCopy: () => true,
    runtimeWriteText: async () => {
      runtimeCalled = true
      return true
    },
    browserWriteText: async () => {
      browserCalled = true
    },
  })

  assert.equal(runtimeCalled, false)
  assert.equal(browserCalled, false)
})

test('copyTextWithFallbacks falls back to browser clipboard when runtime returns false', async () => {
  let browserText = null

  await copyTextWithFallbacks('prompt body', {
    runtimeWriteText: async () => false,
    browserWriteText: async (text) => {
      browserText = text
    },
  })

  assert.equal(browserText, 'prompt body')
})

test('copyTextWithFallbacks accepts runtime clipboard with undefined return value', async () => {
  let runtimeCalls = 0

  await copyTextWithFallbacks('prompt body', {
    runtimeWriteText: async () => {
      runtimeCalls += 1
    },
    browserWriteText: async () => {
      throw new Error('browser fallback should not run')
    },
  })

  assert.equal(runtimeCalls, 1)
})

test('copyTextWithFallbacks falls back to document execCommand', async () => {
  let execCopied = null

  await copyTextWithFallbacks('prompt body', {
    runtimeWriteText: async () => false,
    execCommandCopy: (text) => {
      execCopied = text
      return true
    },
  })

  assert.equal(execCopied, 'prompt body')
})

test('copyTextWithDocumentCommand copies text and restores focus', () => {
  const events = []
  const previousFocus = {
    focus() {
      events.push('restore-focus')
    },
  }
  const textarea = {
    value: '',
    style: { position: '', left: '', top: '', opacity: '' },
    setAttribute(name, value) {
      events.push(`attr:${name}=${value}`)
    },
    focus() {
      events.push('focus-textarea')
    },
    select() {
      events.push('select')
    },
    setSelectionRange(start, end) {
      events.push(`range:${start}-${end}`)
    },
  }
  const doc = {
    activeElement: previousFocus,
    body: {
      appendChild(node) {
        events.push(node === textarea ? 'append' : 'append-other')
      },
      removeChild(node) {
        events.push(node === textarea ? 'remove' : 'remove-other')
      },
    },
    createElement(tagName) {
      assert.equal(tagName, 'textarea')
      return textarea
    },
    execCommand(command) {
      events.push(`exec:${command}`)
      return true
    },
  }

  const ok = copyTextWithDocumentCommand('prompt body', doc)

  assert.equal(ok, true)
  assert.deepEqual(events, [
    'attr:readonly=',
    'append',
    'focus-textarea',
    'select',
    'range:0-11',
    'exec:copy',
    'remove',
    'restore-focus',
  ])
  assert.equal(textarea.value, 'prompt body')
})

test('copyTextWithFallbacks throws when all clipboard strategies fail', async () => {
  await assert.rejects(
    () => copyTextWithFallbacks('prompt body', {}),
    /clipboard unavailable/,
  )
})
