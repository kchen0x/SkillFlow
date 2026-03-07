import { createContext, useContext, ReactNode } from 'react'
import { Theme, useTheme } from '../hooks/useTheme'

interface ThemeContextValue {
  theme: Theme
  setTheme: (theme: Theme) => void
  cycleTheme: () => void
}

const ThemeContext = createContext<ThemeContextValue>({
  theme: 'dark',
  setTheme: () => {},
  cycleTheme: () => {},
})

export function ThemeProvider({ children }: { children: ReactNode }) {
  const { theme, setTheme, cycleTheme } = useTheme()
  return (
    <ThemeContext.Provider value={{ theme, setTheme: (nextTheme) => setTheme(nextTheme), cycleTheme }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useThemeContext() {
  return useContext(ThemeContext)
}
