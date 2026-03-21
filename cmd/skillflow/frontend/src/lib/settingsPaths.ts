export type SettingsPathKey = 'appDataDir' | 'repoCacheDir'

export type SettingsPathRow = {
  key: SettingsPathKey
  value: string
  action: 'open' | 'pick'
}

export function buildSettingsPathRows(cfg: { repoCacheDir?: string } | null | undefined, appDataDir: string): SettingsPathRow[] {
  return [
    {
      key: 'appDataDir',
      value: appDataDir ?? '',
      action: 'open',
    },
    {
      key: 'repoCacheDir',
      value: cfg?.repoCacheDir ?? '',
      action: 'pick',
    },
  ]
}
