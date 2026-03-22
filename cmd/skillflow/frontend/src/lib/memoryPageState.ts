export type MemoryAutoSyncMode = 'off' | 'merge' | 'takeover'

export type MemoryPushConfigLike = {
  mode?: string
  autoPush?: boolean
}

export type MemoryBatchPushState = {
  mode: 'merge' | 'takeover'
  selectedAgents: string[]
  selectedModules: string[]
}

export function getMemoryAutoSyncMode(config?: MemoryPushConfigLike): MemoryAutoSyncMode {
  if (!config?.autoPush) {
    return 'off'
  }
  return config.mode === 'takeover' ? 'takeover' : 'merge'
}

export function getMemoryPushConfigForAutoSyncMode(mode: MemoryAutoSyncMode) {
  if (mode === 'off') {
    return {
      mode: 'merge',
      autoPush: false,
    }
  }
  return {
    mode,
    autoPush: true,
  }
}

export function createMemoryBatchPushState(): MemoryBatchPushState {
  return {
    mode: 'merge',
    selectedAgents: [],
    selectedModules: [],
  }
}

export function toggleMemoryBatchModule(selectedModules: string[], moduleName: string): string[] {
  if (selectedModules.includes(moduleName)) {
    return selectedModules.filter(name => name !== moduleName)
  }
  return [...selectedModules, moduleName]
}

export function toggleMemoryBatchAgent(selectedAgents: string[], agentType: string): string[] {
  if (selectedAgents.includes(agentType)) {
    return selectedAgents.filter(name => name !== agentType)
  }
  return [...selectedAgents, agentType]
}

export function isMemoryBatchSelected(kind: 'main' | 'module', name: string, selectedModules: string[]): boolean {
  if (kind === 'main') {
    return true
  }
  return selectedModules.includes(name)
}

export function isMemoryBatchPushReady(state: MemoryBatchPushState): boolean {
  return state.selectedAgents.length > 0 && state.selectedModules.length > 0
}
