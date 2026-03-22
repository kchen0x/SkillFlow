import test from 'node:test'
import assert from 'node:assert/strict'
import {
  appendPromptImageURL,
  buildPromptLinksMarkdown,
  normalizePromptImageURLs,
  normalizePromptPreviewImageURL,
  parsePromptWebLinkLine,
  parsePromptWebLinks,
} from '../.tmp-tests/src/lib/promptRichContent.js'

test('normalizePromptImageURLs trims blanks, preserves order, and caps at three items', () => {
  assert.deepEqual(
    normalizePromptImageURLs([
      '  https://cdn.example.com/1.png  ',
      '',
      'https://cdn.example.com/2.png',
      'https://cdn.example.com/3.png',
      'https://cdn.example.com/4.png',
    ]),
    [
      'https://cdn.example.com/1.png',
      'https://cdn.example.com/2.png',
      'https://cdn.example.com/3.png',
    ],
  )
})

test('parsePromptWebLinks returns markdown links with trimmed labels and urls', () => {
  assert.deepEqual(
    parsePromptWebLinks(`
      [ 产品文档 ]( https://docs.example.com/product )
      [预览环境](https://preview.example.com)
    `),
    [
      {
        label: '产品文档',
        url: 'https://docs.example.com/product',
      },
      {
        label: '预览环境',
        url: 'https://preview.example.com',
      },
    ],
  )
})

test('parsePromptWebLinkLine parses a single markdown link line', () => {
  assert.deepEqual(
    parsePromptWebLinkLine('  [ 产品文档 ]( https://docs.example.com/product )  '),
    {
      label: '产品文档',
      url: 'https://docs.example.com/product',
    },
  )
})

test('parsePromptWebLinkLine returns null for invalid single-line input', () => {
  assert.equal(parsePromptWebLinkLine('https://docs.example.com/product'), null)
})

test('normalizePromptPreviewImageURL trims and validates a previewable image URL', () => {
  assert.equal(
    normalizePromptPreviewImageURL('  https://cdn.example.com/prompt.png  '),
    'https://cdn.example.com/prompt.png',
  )
  assert.equal(normalizePromptPreviewImageURL('ftp://cdn.example.com/prompt.png'), null)
})

test('appendPromptImageURL appends one valid image URL and rejects overflow', () => {
  assert.deepEqual(
    appendPromptImageURL(['https://cdn.example.com/1.png'], '  https://cdn.example.com/2.png  '),
    ['https://cdn.example.com/1.png', 'https://cdn.example.com/2.png'],
  )
  assert.equal(
    appendPromptImageURL([
      'https://cdn.example.com/1.png',
      'https://cdn.example.com/2.png',
      'https://cdn.example.com/3.png',
    ], 'https://cdn.example.com/4.png'),
    null,
  )
})

test('buildPromptLinksMarkdown rebuilds markdown lines from parsed links', () => {
  assert.equal(
    buildPromptLinksMarkdown([
      {
        label: 'PRD',
        url: 'https://docs.example.com/prd',
      },
      {
        label: 'Preview',
        url: 'https://preview.example.com/review',
      },
    ]),
    [
      '[PRD](https://docs.example.com/prd)',
      '[Preview](https://preview.example.com/review)',
    ].join('\n'),
  )
})
