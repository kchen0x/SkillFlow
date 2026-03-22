function escapeHtml(value: string): string {
  return value
    .split('&').join('&amp;')
    .split('<').join('&lt;')
    .split('>').join('&gt;')
    .split('"').join('&quot;')
    .split("'").join('&#39;')
}

function formatInlineMarkdown(text: string): string {
  let result = ''
  let index = 0

  while (index < text.length) {
    if (text[index] === '`') {
      const codeEnd = text.indexOf('`', index + 1)
      if (codeEnd > index + 1) {
        result += `<code>${escapeHtml(text.slice(index + 1, codeEnd))}</code>`
        index = codeEnd + 1
        continue
      }
    }

    if (text[index] === '[') {
      const textEnd = text.indexOf(']', index + 1)
      const urlStart = textEnd >= 0 ? text.indexOf('(', textEnd + 1) : -1
      const urlEnd = urlStart >= 0 ? text.indexOf(')', urlStart + 1) : -1
      if (textEnd > index + 1 && urlStart === textEnd + 1 && urlEnd > urlStart + 1) {
        const label = text.slice(index + 1, textEnd)
        const url = text.slice(urlStart + 1, urlEnd)
        result += `<a href="${escapeHtml(url)}" target="_blank" rel="noreferrer">${escapeHtml(label)}</a>`
        index = urlEnd + 1
        continue
      }
    }

    result += escapeHtml(text[index])
    index++
  }

  return result
}

function renderParagraph(lines: string[]): string {
  return `<p>${formatInlineMarkdown(lines.join(' '))}</p>`
}

function renderList(items: string[], ordered: boolean): string {
  const tag = ordered ? 'ol' : 'ul'
  return `<${tag}>${items.map(item => `<li>${formatInlineMarkdown(item)}</li>`).join('')}</${tag}>`
}

export function renderMemoryMarkdown(markdown: string): string {
  const lines = markdown.split('\r\n').join('\n').split('\n')
  const html: string[] = []
  let paragraphLines: string[] = []
  let listItems: string[] = []
  let listOrdered = false
  let quoteLines: string[] = []
  let codeLines: string[] = []
  let inCodeFence = false

  const flushParagraph = () => {
    if (paragraphLines.length === 0) return
    html.push(renderParagraph(paragraphLines))
    paragraphLines = []
  }

  const flushList = () => {
    if (listItems.length === 0) return
    html.push(renderList(listItems, listOrdered))
    listItems = []
  }

  const flushQuote = () => {
    if (quoteLines.length === 0) return
    html.push(`<blockquote>${renderParagraph(quoteLines)}</blockquote>`)
    quoteLines = []
  }

  const flushCode = () => {
    if (!inCodeFence) return
    html.push(`<pre><code>${escapeHtml(codeLines.join('\n'))}</code></pre>`)
    codeLines = []
    inCodeFence = false
  }

  for (const line of lines) {
    if (line.startsWith('```')) {
      flushParagraph()
      flushList()
      flushQuote()
      if (inCodeFence) {
        flushCode()
      } else {
        inCodeFence = true
        codeLines = []
      }
      continue
    }

    if (inCodeFence) {
      codeLines.push(line)
      continue
    }

    if (!line.trim()) {
      flushParagraph()
      flushList()
      flushQuote()
      continue
    }

    const headingMatch = line.match(/^(#{1,6})\s+(.*)$/)
    if (headingMatch) {
      flushParagraph()
      flushList()
      flushQuote()
      const level = headingMatch[1].length
      html.push(`<h${level}>${formatInlineMarkdown(headingMatch[2])}</h${level}>`)
      continue
    }

    const quoteMatch = line.match(/^>\s?(.*)$/)
    if (quoteMatch) {
      flushParagraph()
      flushList()
      quoteLines.push(quoteMatch[1])
      continue
    }

    const orderedMatch = line.match(/^\d+\.\s+(.*)$/)
    if (orderedMatch) {
      flushParagraph()
      flushQuote()
      if (listItems.length > 0 && !listOrdered) {
        flushList()
      }
      listOrdered = true
      listItems.push(orderedMatch[1])
      continue
    }

    const unorderedMatch = line.match(/^[-*+]\s+(.*)$/)
    if (unorderedMatch) {
      flushParagraph()
      flushQuote()
      if (listItems.length > 0 && listOrdered) {
        flushList()
      }
      listOrdered = false
      listItems.push(unorderedMatch[1])
      continue
    }

    flushList()
    flushQuote()
    paragraphLines.push(line.trim())
  }

  flushParagraph()
  flushList()
  flushQuote()
  if (inCodeFence) {
    flushCode()
  }

  return html.join('')
}
