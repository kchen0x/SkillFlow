import type { WailsEventSubscription } from './wailsEvents'

export function createMemoryEventSubscriptions(load: () => void): readonly WailsEventSubscription[] {
  return [
    ['memory:content:changed', load],
  ] as const
}
