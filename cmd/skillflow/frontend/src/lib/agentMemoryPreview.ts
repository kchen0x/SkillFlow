export type AgentMemoryRuleLike = {
  name: string
  path: string
  content?: string
  managed?: boolean
}

export type AgentMemoryPreviewLike = {
  memoryPath?: string
  mainExists?: boolean
  mainContent?: string
  rulesDir?: string
  rulesDirExists?: boolean
  rules?: AgentMemoryRuleLike[]
}

export type AgentMemoryEntry = {
  key: string
  kind: 'main' | 'rule'
  title: string
  path: string
  content: string
  managed: boolean
}

export const MAIN_MEMORY_ENTRY_TITLE = 'Main Memory'

export function buildAgentMemoryEntries(preview?: AgentMemoryPreviewLike | null): AgentMemoryEntry[] {
  if (!preview) return []

  const rules = [...(preview.rules ?? [])]
    .sort((a, b) => {
      if (!!a.managed !== !!b.managed) return a.managed ? -1 : 1
      return a.name.localeCompare(b.name)
    })
    .map(rule => ({
      key: `rule:${rule.path}`,
      kind: 'rule' as const,
      title: rule.name,
      path: rule.path,
      content: rule.content ?? '',
      managed: !!rule.managed,
    }))

  return [
    {
      key: 'main',
      kind: 'main',
      title: MAIN_MEMORY_ENTRY_TITLE,
      path: preview.memoryPath ?? '',
      content: preview.mainContent ?? '',
      managed: false,
    },
    ...rules,
  ]
}

export function filterAgentMemoryEntries(entries: AgentMemoryEntry[], search: string): AgentMemoryEntry[] {
  const query = search.trim().toLowerCase()
  if (!query) return entries

  return entries.filter(entry =>
    entry.title.toLowerCase().includes(query) ||
    entry.path.toLowerCase().includes(query) ||
    entry.content.toLowerCase().includes(query),
  )
}
