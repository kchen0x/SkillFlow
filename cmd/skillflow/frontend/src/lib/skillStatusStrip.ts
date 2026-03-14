export function summarizePushedAgents(pushedAgents: string[], maxVisiblePushedAgents: number) {
  const safeMaxVisible = Math.max(0, maxVisiblePushedAgents)
  const visibleAgents = pushedAgents.slice(0, safeMaxVisible)

  return {
    visibleAgents,
    overflowCount: Math.max(0, pushedAgents.length - visibleAgents.length),
  }
}
