import { getListLoadState, type ListLoadState } from './listLoadState.js'

export type StarredRepoListState = ListLoadState

export function getStarredRepoListState({
  isLoading,
  skillCount,
}: {
  isLoading: boolean
  skillCount: number
}): StarredRepoListState {
  return getListLoadState({ isLoading, itemCount: skillCount })
}
