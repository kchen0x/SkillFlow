import { useState, useEffect } from 'react'

export const THEMES = ['dark', 'young', 'light', 'sport'] as const

export type Theme = typeof THEMES[number]

const STORAGE_KEY = 'skillflow-theme-v2'
const LEGACY_STORAGE_KEY = 'skillflow-theme'

export const THEME_LABELS: Record<Theme, string> = {
  dark: 'Dark',
  young: 'Young',
  light: 'Light',
  sport: 'Sport',
}

function isTheme(value: string | null): value is Theme {
  return value !== null && THEMES.includes(value as Theme)
}

type ThemeStorage = Pick<Storage, 'getItem'>

export function resolveInitialThemeFromStorage(storage: ThemeStorage): Theme {
  const stored = storage.getItem(STORAGE_KEY)
  if (isTheme(stored)) return stored

  const legacyStored = storage.getItem(LEGACY_STORAGE_KEY)
  if (legacyStored === 'light') return 'young'
  if (legacyStored === 'dark') return 'dark'

  return 'young'
}

function getInitialTheme(): Theme {
  return resolveInitialThemeFromStorage(localStorage)
}

export function getNextTheme(theme: Theme): Theme {
  const currentIndex = THEMES.indexOf(theme)
  return THEMES[(currentIndex + 1) % THEMES.length]
}

export function useTheme() {
  const [theme, setTheme] = useState<Theme>(getInitialTheme)

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    document.documentElement.style.colorScheme = theme === 'dark' ? 'dark' : 'light'
    localStorage.setItem(STORAGE_KEY, theme)
    localStorage.removeItem(LEGACY_STORAGE_KEY)
  }, [theme])

  const cycleTheme = () => {
    setTheme(prev => getNextTheme(prev))
  }

  return { theme, setTheme, cycleTheme }
}
