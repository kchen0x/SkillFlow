import test from 'node:test'
import assert from 'node:assert/strict'
import { renderMemoryMarkdown } from '../.tmp-tests/src/lib/memoryMarkdown.js'

test('renderMemoryMarkdown renders common markdown blocks and inline syntax', () => {
  const html = renderMemoryMarkdown(`# Title

Paragraph with [link](https://example.com) and \`inline\`.

> Quote line

- First
- Second

\`\`\`
const answer = 42
\`\`\`
`)

  assert.match(html, /<h1>Title<\/h1>/)
  assert.match(html, /<p>Paragraph with <a href="https:\/\/example.com"/)
  assert.match(html, /<code>inline<\/code>/)
  assert.match(html, /<blockquote><p>Quote line<\/p><\/blockquote>/)
  assert.match(html, /<ul><li>First<\/li><li>Second<\/li><\/ul>/)
  assert.match(html, /<pre><code>const answer = 42<\/code><\/pre>/)
})

test('renderMemoryMarkdown escapes html instead of trusting raw markup', () => {
  const html = renderMemoryMarkdown('<script>alert(1)</script>')

  assert.doesNotMatch(html, /<script>/)
  assert.match(html, /&lt;script&gt;alert\(1\)&lt;\/script&gt;/)
})
