import { createContext, useContext, type ReactNode } from 'react'
import {
  DEFAULT_SKILL_STATUS_VISIBILITY,
  type SkillStatusPageKey,
  type SkillStatusVisibilityConfig,
} from '../lib/skillStatusVisibility'

type SkillStatusVisibilityContextValue = {
  visibility: SkillStatusVisibilityConfig
}

const SkillStatusVisibilityContext = createContext<SkillStatusVisibilityContextValue | null>(null)

export function SkillStatusVisibilityProvider({ children }: { children: ReactNode }) {
  return (
    <SkillStatusVisibilityContext.Provider value={{ visibility: DEFAULT_SKILL_STATUS_VISIBILITY }}>
      {children}
    </SkillStatusVisibilityContext.Provider>
  )
}

export function useSkillStatusVisibility(page: SkillStatusPageKey) {
  const ctx = useContext(SkillStatusVisibilityContext)
  if (!ctx) throw new Error('useSkillStatusVisibility must be used inside SkillStatusVisibilityProvider')
  return ctx.visibility[page]
}
