import type { WailsEventSubscription } from './wailsEvents'

type DashboardSkillConfig = {
  autoPushAgents?: string[] | null
  autoUpdateSkills?: boolean | null
}

export type DashboardSkillSettings = {
  autoPushAgents: string[]
  autoUpdateSkills: boolean
}

export type DashboardToolbarActionKey = 'update' | 'manualPush' | 'batchDelete' | 'import' | 'autoUpdate'

export type DashboardAutoUpdateActionState = {
  buttonTone: 'primary' | 'secondary'
  visualState: 'enabled' | 'disabled'
}

const DASHBOARD_TOOLBAR_ACTION_KEYS: readonly DashboardToolbarActionKey[] = [
  'update',
  'manualPush',
  'batchDelete',
  'import',
  'autoUpdate',
]

export function readDashboardSkillSettings(cfg?: DashboardSkillConfig | null): DashboardSkillSettings {
  return {
    autoPushAgents: [...(cfg?.autoPushAgents ?? [])],
    autoUpdateSkills: !!cfg?.autoUpdateSkills,
  }
}

export function listDashboardToolbarActionKeys(): DashboardToolbarActionKey[] {
  return [...DASHBOARD_TOOLBAR_ACTION_KEYS]
}

export function getDashboardAutoUpdateActionState(autoUpdateSkills: boolean): DashboardAutoUpdateActionState {
  if (autoUpdateSkills) {
    return {
      buttonTone: 'primary',
      visualState: 'enabled',
    }
  }

  return {
    buttonTone: 'secondary',
    visualState: 'disabled',
  }
}

export function toggleDashboardAutoPushAgent(currentAgents: string[], agentName: string): string[] {
  const nextAgents = new Set(currentAgents.filter(Boolean))
  if (nextAgents.has(agentName)) nextAgents.delete(agentName)
  else nextAgents.add(agentName)
  return Array.from(nextAgents)
}

export function createDashboardSkillEventSubscriptions(load: () => void): readonly WailsEventSubscription[] {
  return [
    ['update.available', load],
    ['skills.updated', load],
  ] as const
}

export function createToolSkillsEventSubscriptions(load: () => void): readonly WailsEventSubscription[] {
  return [
    ['skills.updated', load],
  ] as const
}
