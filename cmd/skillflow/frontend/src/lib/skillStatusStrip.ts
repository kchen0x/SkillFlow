export function summarizePushedTools(pushedTools: string[], maxVisiblePushedTools: number) {
  const safeMaxVisible = Math.max(0, maxVisiblePushedTools)
  const visibleTools = pushedTools.slice(0, safeMaxVisible)

  return {
    visibleTools,
    overflowCount: Math.max(0, pushedTools.length - visibleTools.length),
  }
}
