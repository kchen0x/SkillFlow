export type DashboardManualPushState = {
  selectedAgents: string[]
  selectedSkillIDs: string[]
}

export function createDashboardManualPushState(): DashboardManualPushState {
  return {
    selectedAgents: [],
    selectedSkillIDs: [],
  }
}

export function toggleDashboardManualPushAgent(selectedAgents: string[], agentName: string): string[] {
  if (selectedAgents.includes(agentName)) {
    return selectedAgents.filter(name => name !== agentName)
  }
  return [...selectedAgents, agentName]
}

export function toggleDashboardManualPushSkill(selectedSkillIDs: string[], skillID: string): string[] {
  if (selectedSkillIDs.includes(skillID)) {
    return selectedSkillIDs.filter(id => id !== skillID)
  }
  return [...selectedSkillIDs, skillID]
}

export function toggleDashboardManualPushAllVisible(selectedSkillIDs: string[], visibleSkillIDs: string[]): string[] {
  const selectedSet = new Set(selectedSkillIDs)
  if (visibleSkillIDs.every(id => selectedSet.has(id))) {
    return selectedSkillIDs.filter(id => !visibleSkillIDs.includes(id))
  }
  const next = [...selectedSkillIDs]
  for (const id of visibleSkillIDs) {
    if (!selectedSet.has(id)) {
      next.push(id)
    }
  }
  return next
}

export function syncDashboardManualPushVisibleSelection(selectedSkillIDs: string[], visibleSkillIDs: string[]): string[] {
  const visibleSet = new Set(visibleSkillIDs)
  return selectedSkillIDs.filter(id => visibleSet.has(id))
}

export function isDashboardManualPushReady(state: DashboardManualPushState): boolean {
  return state.selectedAgents.length > 0 && state.selectedSkillIDs.length > 0
}
