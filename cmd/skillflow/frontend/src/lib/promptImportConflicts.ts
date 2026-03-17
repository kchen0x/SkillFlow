export type PromptImportConflictItem = {
  name: string
}

export type PromptImportDecision = 'skip' | 'overwrite'

export type PromptImportDecisionResult<T extends PromptImportConflictItem> = {
  overwriteNames: string[]
  remainingConflicts: T[]
}

export function applyPromptImportDecision<T extends PromptImportConflictItem>(
  conflicts: T[],
  overwriteNames: string[],
  decision: PromptImportDecision,
  applyToRemaining: boolean,
): PromptImportDecisionResult<T> {
  if (conflicts.length === 0) {
    return {
      overwriteNames: [...overwriteNames],
      remainingConflicts: [],
    }
  }

  const handled = applyToRemaining ? conflicts : conflicts.slice(0, 1)
  const nextOverwriteNames = new Set(overwriteNames)
  if (decision === 'overwrite') {
    handled.forEach((item) => nextOverwriteNames.add(item.name))
  }

  return {
    overwriteNames: [...nextOverwriteNames],
    remainingConflicts: applyToRemaining ? [] : conflicts.slice(1),
  }
}
