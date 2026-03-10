export type ListLoadState = 'loading' | 'empty' | 'ready'

export function getListLoadState({
  isLoading,
  itemCount,
}: {
  isLoading: boolean
  itemCount: number
}): ListLoadState {
  if (isLoading) return 'loading'
  if (itemCount === 0) return 'empty'
  return 'ready'
}
