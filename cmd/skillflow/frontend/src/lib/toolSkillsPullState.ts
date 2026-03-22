export type ToolSkillsPullState = {
  targetCategory: string
  selectedPaths: string[]
}

type ToolSkillsPullEntry = {
  path?: string
  imported?: boolean
}

export function createToolSkillsPullState(targetCategory: string): ToolSkillsPullState {
  return {
    targetCategory,
    selectedPaths: [],
  }
}

export function toggleToolSkillsPullPath(selectedPaths: string[], path: string): string[] {
  if (selectedPaths.includes(path)) {
    return selectedPaths.filter(item => item !== path)
  }
  return [...selectedPaths, path]
}

export function toggleToolSkillsPullAllVisible(selectedPaths: string[], visiblePaths: string[]): string[] {
  const selectedSet = new Set(selectedPaths)
  if (visiblePaths.every(path => selectedSet.has(path))) {
    return selectedPaths.filter(path => !visiblePaths.includes(path))
  }
  const next = [...selectedPaths]
  for (const path of visiblePaths) {
    if (!selectedSet.has(path)) {
      next.push(path)
    }
  }
  return next
}

export function getToolSkillsVisibleNotImportedPaths(entries: ToolSkillsPullEntry[]): string[] {
  return entries
    .filter(entry => !entry.imported && typeof entry.path === 'string' && entry.path.trim() !== '')
    .map(entry => entry.path as string)
}

export function toggleToolSkillsPullVisibleNotImported(selectedPaths: string[], visibleNotImportedPaths: string[]): string[] {
  const selectedSet = new Set(selectedPaths)
  if (visibleNotImportedPaths.every(path => selectedSet.has(path))) {
    return selectedPaths.filter(path => !visibleNotImportedPaths.includes(path))
  }
  const next = [...selectedPaths]
  for (const path of visibleNotImportedPaths) {
    if (!selectedSet.has(path)) {
      next.push(path)
    }
  }
  return next
}

export function syncToolSkillsPullVisibleSelection(selectedPaths: string[], visiblePaths: string[]): string[] {
  const visibleSet = new Set(visiblePaths)
  return selectedPaths.filter(path => visibleSet.has(path))
}

export function isToolSkillsPullReady(state: Pick<ToolSkillsPullState, 'selectedPaths'>): boolean {
  return state.selectedPaths.length > 0
}
