export type SkillStatusKey = 'imported' | 'updatable' | 'pushedAgents'
export type SkillStatusPageKey =
  | 'mySkills'
  | 'myAgents'
  | 'pushToAgent'
  | 'pullFromAgent'
  | 'starredRepos'

export type SkillStatusVisibilityConfig = Record<SkillStatusPageKey, SkillStatusKey[]>

export const DEFAULT_SKILL_STATUS_VISIBILITY: SkillStatusVisibilityConfig = {
  mySkills: ['updatable', 'pushedAgents'],
  myAgents: ['imported', 'updatable', 'pushedAgents'],
  pushToAgent: ['pushedAgents'],
  pullFromAgent: ['imported'],
  starredRepos: ['imported', 'pushedAgents'],
}

export function hasSkillStatus(
  visibility: SkillStatusVisibilityConfig,
  page: SkillStatusPageKey,
  status: SkillStatusKey,
) {
  return visibility[page].includes(status)
}
