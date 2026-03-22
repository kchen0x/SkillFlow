export function buildModuleDeletePreview(content: string, maxLines = 3): string {
  if (!content) return ''

  return content
    .split('\n')
    .filter(line => line.trim())
    .slice(0, maxLines)
    .join('\n')
}
