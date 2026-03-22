export type ToolSkillsPanel = 'skills' | 'memory'
type ToolSkillsSortOrder = 'asc' | 'desc'

type SkillLike = {
  name?: string
}

type MemoryEntryLike = {
  key: string
  kind: 'main' | 'rule'
  title: string
  path: string
  content: string
  managed: boolean
}

type FilterToolSkillsPanelContentInput<TSkill extends SkillLike> = {
  activePanel: ToolSkillsPanel
  search: string
  sortOrder: ToolSkillsSortOrder
  pushSkills: TSkill[]
  scanOnlySkills: TSkill[]
  memoryEntries: MemoryEntryLike[]
}

type FilterToolSkillsPanelContentResult<TSkill extends SkillLike> = {
  filteredPushSkills: TSkill[]
  filteredScanOnlySkills: TSkill[]
  filteredMemoryEntries: MemoryEntryLike[]
}

type VisibleToolSkillsResultCountInput<TSkill extends SkillLike> = {
  activePanel: ToolSkillsPanel
  filteredPushSkills: TSkill[]
  filteredScanOnlySkills: TSkill[]
  filteredMemoryEntries: MemoryEntryLike[]
}

const nameCollator = new Intl.Collator(undefined, {
  numeric: true,
  sensitivity: 'base',
})

export function getDefaultToolSkillsPanel(): ToolSkillsPanel {
  return 'skills'
}

export function filterToolSkillsPanelContent<TSkill extends SkillLike>({
  activePanel,
  search,
  sortOrder,
  pushSkills,
  scanOnlySkills,
  memoryEntries,
}: FilterToolSkillsPanelContentInput<TSkill>): FilterToolSkillsPanelContentResult<TSkill> {
  if (activePanel === 'memory') {
    return {
      filteredPushSkills: [],
      filteredScanOnlySkills: [],
      filteredMemoryEntries: sortMemoryEntries(
        filterMemoryEntries(memoryEntries, search),
        sortOrder,
      ),
    }
  }

  return {
    filteredPushSkills: filterSkills(pushSkills, search, sortOrder),
    filteredScanOnlySkills: filterSkills(scanOnlySkills, search, sortOrder),
    filteredMemoryEntries: [],
  }
}

export function getVisibleToolSkillsResultCount<TSkill extends SkillLike>({
  activePanel,
  filteredPushSkills,
  filteredScanOnlySkills,
  filteredMemoryEntries,
}: VisibleToolSkillsResultCountInput<TSkill>): number {
  if (activePanel === 'memory') return filteredMemoryEntries.length
  return filteredPushSkills.length + filteredScanOnlySkills.length
}

function filterSkills<TSkill extends SkillLike>(items: TSkill[], search: string, sortOrder: ToolSkillsSortOrder): TSkill[] {
  const query = search.trim().toLocaleLowerCase()
  const filtered = query
    ? items.filter(item => (item.name ?? '').toLocaleLowerCase().includes(query))
    : items

  return [...filtered].sort((left, right) => {
    const result = nameCollator.compare((left.name ?? '').trim(), (right.name ?? '').trim())
    return sortOrder === 'asc' ? result : -result
  })
}

function filterMemoryEntries(entries: MemoryEntryLike[], search: string): MemoryEntryLike[] {
  const query = search.trim().toLowerCase()
  if (!query) return entries

  return entries.filter(entry =>
    entry.title.toLowerCase().includes(query) ||
    entry.path.toLowerCase().includes(query) ||
    entry.content.toLowerCase().includes(query),
  )
}

function sortMemoryEntries(entries: MemoryEntryLike[], sortOrder: ToolSkillsSortOrder): MemoryEntryLike[] {
  if (sortOrder === 'asc') return entries

  const mainEntry = entries.find(entry => entry.kind === 'main')
  const ruleEntries = entries.filter(entry => entry.kind === 'rule').reverse()

  return mainEntry ? [mainEntry, ...ruleEntries] : ruleEntries
}
