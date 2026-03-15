import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('../src/components/PromptEditorDialog.tsx', import.meta.url), 'utf8')

test('prompt editor image cards opt into group hover state for overlay actions', () => {
  assert.match(
    source,
    /className="[^"]*\bgroup\b[^"]*\brelative\b[^"]*\boverflow-hidden\b[^"]*hover:scale-\[1\.01\][^"]*"/,
  )
})

test('prompt editor delete button stays hidden until the image card is hovered or focused', () => {
  assert.match(source, /className="[^"]*\bopacity-0\b[^"]*\bpointer-events-none\b[^"]*"/)
  assert.match(source, /className="[^"]*\bgroup-hover:opacity-100\b[^"]*\bgroup-hover:pointer-events-auto\b[^"]*"/)
  assert.match(source, /className="[^"]*\bgroup-focus-within:opacity-100\b[^"]*\bgroup-focus-within:pointer-events-auto\b[^"]*"/)
})
