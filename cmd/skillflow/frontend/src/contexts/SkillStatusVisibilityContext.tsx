import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { GetConfig } from '../../wailsjs/go/main/App'
import {
  DEFAULT_SKILL_STATUS_VISIBILITY,
  normalizeSkillStatusVisibility,
  type SkillStatusPageKey,
  type SkillStatusVisibilityConfig,
} from '../lib/skillStatusVisibility'

type SkillStatusVisibilityContextValue = {
  visibility: SkillStatusVisibilityConfig
  syncFromConfig: (cfg: any) => void
  refreshVisibility: () => Promise<void>
}

const SkillStatusVisibilityContext = createContext<SkillStatusVisibilityContextValue | null>(null)

export function SkillStatusVisibilityProvider({ children }: { children: ReactNode }) {
  const [visibility, setVisibility] = useState<SkillStatusVisibilityConfig>(DEFAULT_SKILL_STATUS_VISIBILITY)

  const syncFromConfig = (cfg: any) => {
    setVisibility(normalizeSkillStatusVisibility(cfg?.skillStatusVisibility))
  }

  const refreshVisibility = async () => {
    try {
      const cfg = await GetConfig()
      syncFromConfig(cfg)
    } catch {
      setVisibility(DEFAULT_SKILL_STATUS_VISIBILITY)
    }
  }

  useEffect(() => {
    void refreshVisibility()
  }, [])

  return (
    <SkillStatusVisibilityContext.Provider value={{ visibility, syncFromConfig, refreshVisibility }}>
      {children}
    </SkillStatusVisibilityContext.Provider>
  )
}

export function useSkillStatusVisibility(page: SkillStatusPageKey) {
  const ctx = useContext(SkillStatusVisibilityContext)
  if (!ctx) throw new Error('useSkillStatusVisibility must be used inside SkillStatusVisibilityProvider')
  return ctx.visibility[page]
}

export function useSkillStatusVisibilityContext() {
  const ctx = useContext(SkillStatusVisibilityContext)
  if (!ctx) throw new Error('useSkillStatusVisibilityContext must be used inside SkillStatusVisibilityProvider')
  return ctx
}
