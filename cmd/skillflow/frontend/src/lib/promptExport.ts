export type PromptExportMode = 'all' | 'current' | 'selected'

export type PromptExportItem = {
  name: string
  category: string
}

export type PromptExportAction = {
  key: PromptExportMode
  label: string
}

export type PromptExportLabels = {
  all: string
  selected: string
}

const defaultLabels: PromptExportLabels = {
  all: '全部',
  selected: '指定',
}

export function buildPromptExportActions(
  selectedCategory: string | null,
  labels: PromptExportLabels = defaultLabels,
): PromptExportAction[] {
  const actions: PromptExportAction[] = [{ key: 'all', label: labels.all }]
  if (selectedCategory) {
    actions.push({ key: 'current', label: selectedCategory })
  }
  actions.push({ key: 'selected', label: labels.selected })
  return actions
}

export function listPromptExportCandidates<T extends PromptExportItem>(prompts: T[], selectedCategory: string | null): T[] {
  if (!selectedCategory) {
    return [...prompts]
  }
  return prompts.filter((item) => item.category === selectedCategory)
}

export function resolvePromptExportNames<T extends PromptExportItem>(
  mode: PromptExportMode,
  prompts: T[],
  selectedCategory: string | null,
  selectedNames: Set<string>,
): string[] {
  if (mode === 'all') {
    return prompts.map((item) => item.name)
  }

  const scoped = listPromptExportCandidates(prompts, selectedCategory)
  if (mode === 'current') {
    return scoped.map((item) => item.name)
  }

  return scoped
    .filter((item) => selectedNames.has(item.name))
    .map((item) => item.name)
}

export function canExportPromptSelection(mode: PromptExportMode, selectedNames: Set<string>) {
  return mode !== 'selected' || selectedNames.size > 0
}
