export type PromptWebLink = {
  label: string
  url: string
}

const promptMarkdownLinkPattern = /^\[(.+?)\]\((.+?)\)$/

export function normalizePromptImageURLs(imageURLs: string[]) {
  const normalized: string[] = []
  for (const rawURL of imageURLs) {
    const normalizedURL = normalizePromptPreviewImageURL(rawURL)
    if (!normalizedURL) {
      continue
    }
    normalized.push(normalizedURL)
    if (normalized.length >= 3) {
      break
    }
  }
  return normalized
}

export function appendPromptImageURL(imageURLs: string[], rawURL: string): string[] | null {
  const normalizedExisting = normalizePromptImageURLs(imageURLs)
  const normalizedURL = normalizePromptPreviewImageURL(rawURL)
  if (!normalizedURL || normalizedExisting.length >= 3) {
    return null
  }
  return [...normalizedExisting, normalizedURL]
}

export function parsePromptWebLinks(raw: string): PromptWebLink[] {
  if (!raw.trim()) {
    return []
  }
  return raw
    .replace(/\r\n/g, '\n')
    .split('\n')
    .map(parsePromptWebLinkLine)
    .filter((link): link is PromptWebLink => link !== null)
}

export function buildPromptLinksMarkdown(links: PromptWebLink[]) {
  return links
    .map(link => {
      const label = link.label.trim()
      const url = link.url.trim()
      if (!label || !url) {
        return ''
      }
      return `[${label}](${url})`
    })
    .filter(Boolean)
    .join('\n')
}

export function parsePromptWebLinkLine(raw: string): PromptWebLink | null {
  const line = raw.trim()
  if (!line) {
    return null
  }
  const matches = line.match(promptMarkdownLinkPattern)
  if (!matches) {
    return null
  }
  const label = matches[1].trim()
  const url = matches[2].trim()
  if (!label || !isPromptHTTPURL(url)) {
    return null
  }
  return { label, url }
}

export function normalizePromptPreviewImageURL(rawURL: string): string | null {
  const trimmed = rawURL.trim()
  if (!trimmed || !isPromptHTTPURL(trimmed)) {
    return null
  }
  return trimmed
}

function isPromptHTTPURL(rawURL: string) {
  try {
    const parsed = new URL(rawURL)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}
