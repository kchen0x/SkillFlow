export type SkillStatusKey = 'imported' | 'updatable' | 'pushedTools'
export type SkillStatusPageKey =
  | 'mySkills'
  | 'myTools'
  | 'pushToTool'
  | 'pullFromTool'
  | 'starredRepos'
  | 'githubInstall'

export type SkillStatusVisibilityConfig = Record<SkillStatusPageKey, SkillStatusKey[]>

export const SKILL_STATUS_ORDER: SkillStatusKey[] = ['imported', 'updatable', 'pushedTools']
export const SKILL_STATUS_PAGE_ORDER: SkillStatusPageKey[] = [
  'mySkills',
  'myTools',
  'pushToTool',
  'pullFromTool',
  'starredRepos',
  'githubInstall',
]

export const DEFAULT_SKILL_STATUS_VISIBILITY: SkillStatusVisibilityConfig = {
  mySkills: ['updatable', 'pushedTools'],
  myTools: ['imported', 'updatable', 'pushedTools'],
  pushToTool: ['pushedTools'],
  pullFromTool: ['imported'],
  starredRepos: ['imported', 'pushedTools'],
  githubInstall: ['imported', 'updatable', 'pushedTools'],
}

function isSkillStatusKey(value: unknown): value is SkillStatusKey {
  return value === 'imported' || value === 'updatable' || value === 'pushedTools'
}

export function normalizeSkillStatusVisibility(raw: any): SkillStatusVisibilityConfig {
  const next = {} as SkillStatusVisibilityConfig
  for (const page of SKILL_STATUS_PAGE_ORDER) {
    const source = Array.isArray(raw?.[page]) ? raw[page] : DEFAULT_SKILL_STATUS_VISIBILITY[page]
    const allowed = new Set(DEFAULT_SKILL_STATUS_VISIBILITY[page])
    const enabled = new Set(source.filter((value: unknown): value is SkillStatusKey => isSkillStatusKey(value) && allowed.has(value)))
    next[page] = SKILL_STATUS_ORDER.filter((status) => allowed.has(status) && enabled.has(status))
  }
  return next
}

export function hasSkillStatus(
  visibility: SkillStatusVisibilityConfig,
  page: SkillStatusPageKey,
  status: SkillStatusKey,
) {
  return visibility[page].includes(status)
}

export function toggleSkillStatusForPage(
  visibility: SkillStatusVisibilityConfig,
  page: SkillStatusPageKey,
  status: SkillStatusKey,
  enabled: boolean,
): SkillStatusVisibilityConfig {
  const pageStatuses = new Set(visibility[page])
  if (enabled) {
    pageStatuses.add(status)
  } else {
    pageStatuses.delete(status)
  }

  return {
    ...visibility,
    [page]: SKILL_STATUS_ORDER.filter((key) => pageStatuses.has(key)),
  }
}
