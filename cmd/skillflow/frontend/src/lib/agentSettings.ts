export type CustomAgentDraft = {
  name: string
  pushDir: string
  memoryPath: string
  rulesDir: string
}

type AgentNameLike = {
  name: string
}

type ValidationResult =
  | { ok: true }
  | { ok: false; reason: 'required_fields' | 'duplicate_name' }

export function createEmptyCustomAgentDraft(): CustomAgentDraft {
  return {
    name: '',
    pushDir: '',
    memoryPath: '',
    rulesDir: '',
  }
}

export function validateCustomAgentDraft(
  draft: CustomAgentDraft,
  agents: AgentNameLike[],
): ValidationResult {
  const normalizedName = draft.name.trim().toLowerCase()
  const pushDir = draft.pushDir.trim()
  const memoryPath = draft.memoryPath.trim()
  const rulesDir = draft.rulesDir.trim()

  if (!normalizedName || !pushDir || !memoryPath || !rulesDir) {
    return { ok: false, reason: 'required_fields' }
  }

  const duplicated = agents.some(agent => agent.name.trim().toLowerCase() === normalizedName)
  if (duplicated) {
    return { ok: false, reason: 'duplicate_name' }
  }

  return { ok: true }
}

export function buildCustomAgentProfile(draft: CustomAgentDraft) {
  const pushDir = draft.pushDir.trim()
  return {
    name: draft.name.trim(),
    pushDir,
    scanDirs: [pushDir],
    memoryPath: draft.memoryPath.trim(),
    rulesDir: draft.rulesDir.trim(),
    enabled: true,
    custom: true,
  }
}
