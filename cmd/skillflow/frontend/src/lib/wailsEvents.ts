export type WailsEventHandler = (...args: any[]) => void
export type WailsEventSubscription = readonly [eventName: string, handler: WailsEventHandler]
export type WailsEventRegister = (eventName: string, handler: WailsEventHandler) => void | (() => void)

export function subscribeToEvents(register: WailsEventRegister, subscriptions: readonly WailsEventSubscription[]) {
  const unsubscribers = subscriptions
    .map(([eventName, handler]) => register(eventName, handler))
    .filter((unsubscribe): unsubscribe is () => void => typeof unsubscribe === 'function')

  return () => {
    for (const unsubscribe of unsubscribers) {
      unsubscribe()
    }
  }
}
