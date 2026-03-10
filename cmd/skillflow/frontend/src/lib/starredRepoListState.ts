export type StarredRepoListState = 'loading' | 'empty' | 'ready'

export function getStarredRepoListState({
  isLoading,
  skillCount,
}: {
  isLoading: boolean
  skillCount: number
}): StarredRepoListState {
  if (isLoading) return 'loading'
  if (skillCount === 0) return 'empty'
  return 'ready'
}
