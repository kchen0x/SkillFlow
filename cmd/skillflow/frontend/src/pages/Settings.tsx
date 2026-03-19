import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react'
import { GetConfig, SaveConfig, ListCloudProviders, AddCustomAgent, RemoveCustomAgent, OpenFolderDialog, CheckAppUpdateAndNotify, GetAppVersion, GetLogDir, OpenLogDir, TestProxyConnection } from '../../wailsjs/go/main/App'
import { Plus, Trash2, Settings, Globe, FolderOpen, RefreshCw, Sun, Moon, Sparkles, Check, Package, Wrench, ArrowUpFromLine, ArrowDownToLine, Star } from 'lucide-react'
import { ToolIcon } from '../config/toolIcons'
import { useThemeContext } from '../contexts/ThemeContext'
import { useSkillStatusVisibilityContext } from '../contexts/SkillStatusVisibilityContext'
import { type Theme } from '../hooks/useTheme'
import { useLanguage } from '../contexts/LanguageContext'
import type { TranslationKey } from '../i18n'
import {
  DEFAULT_SKILL_STATUS_VISIBILITY,
  normalizeSkillStatusVisibility,
  toggleSkillStatusForPage,
  type SkillStatusKey,
  type SkillStatusPageKey,
} from '../lib/skillStatusVisibility'
import { DEFAULT_PROXY_TEST_URL, buildProxyConnectionPayload } from '../lib/proxyConnection'

type Tab = 'agents' | 'cloud' | 'general' | 'network'
type ProxyMode = 'none' | 'system' | 'manual'

type ThemePreviewPalette = {
  shell: string
  sidebar: string
  sidebarSelection: string
  search: string
  panel: string
  accent: string
  accentGlow: string
  text: string
  textMuted: string
  divider: string
}

type ThemeOption = {
  id: Theme
  label: string
  tone: string
  description: string
  icon: ReactNode
  preview: ThemePreviewPalette
}

type StatusPageOption = {
  key: SkillStatusPageKey
  label: TranslationKey
  description: TranslationKey
  icon: ReactNode
}

type StatusOption = {
  key: SkillStatusKey
  label: TranslationKey
}

const CLOUD_PROVIDER_LABEL_KEYS: Record<string, TranslationKey> = {
  aliyun: 'settings.cloudProviderAliyun',
  aws: 'settings.cloudProviderAws',
  google: 'settings.cloudProviderGoogle',
  azure: 'settings.cloudProviderAzure',
  tencent: 'settings.cloudProviderTencent',
  huawei: 'settings.cloudProviderHuawei',
  git: 'settings.cloudProviderGit',
}

const CLOUD_PROVIDER_PRIORITY: Record<string, number> = {
  git: 0,
  huawei: 1,
}

const CLOUD_FIELD_LABEL_KEYS: Record<string, Record<string, TranslationKey>> = {
  aliyun: {
    access_key_id: 'settings.cloudFieldAccessKeyId',
    access_key_secret: 'settings.cloudFieldAccessKeySecret',
    endpoint: 'settings.cloudFieldEndpoint',
  },
  aws: {
    access_key_id: 'settings.cloudFieldAccessKeyId',
    secret_access_key: 'settings.cloudFieldSecretAccessKey',
    region: 'settings.cloudFieldRegion',
  },
  google: {
    service_account_json: 'settings.cloudFieldServiceAccountJson',
  },
  azure: {
    account_name: 'settings.cloudFieldAccountName',
    account_key: 'settings.cloudFieldAccountKey',
    service_url: 'settings.cloudFieldServiceUrl',
  },
  tencent: {
    secret_id: 'settings.cloudFieldSecretId',
    secret_key: 'settings.cloudFieldSecretKey',
    endpoint: 'settings.cloudFieldEndpoint',
  },
  huawei: {
    access_key_id: 'settings.cloudFieldAccessKeyId',
    secret_access_key: 'settings.cloudFieldSecretAccessKey',
    endpoint: 'settings.cloudFieldEndpoint',
  },
  git: {
    repo_url: 'settings.cloudFieldRepoUrl',
    branch: 'settings.cloudFieldBranch',
    username: 'settings.cloudFieldUsernameOptional',
    token: 'settings.cloudFieldTokenOptional',
  },
}

const defaultRepoScanMaxDepth = 5
const minRepoScanMaxDepth = 1
const maxRepoScanMaxDepth = 20
const CLOUD_REMOTE_ROOT_DIR = 'skillflow'
const STATUS_PAGE_OPTIONS: StatusPageOption[] = [
  { key: 'mySkills', label: 'settings.statusPageMySkills', description: 'settings.statusPageMySkillsDesc', icon: <Package size={14} /> },
  { key: 'myAgents', label: 'settings.statusPageMyTools', description: 'settings.statusPageMyToolsDesc', icon: <Wrench size={14} /> },
  { key: 'pushToAgent', label: 'settings.statusPagePushToTool', description: 'settings.statusPagePushToToolDesc', icon: <ArrowUpFromLine size={14} /> },
  { key: 'pullFromAgent', label: 'settings.statusPagePullFromTool', description: 'settings.statusPagePullFromToolDesc', icon: <ArrowDownToLine size={14} /> },
  { key: 'starredRepos', label: 'settings.statusPageStarredRepos', description: 'settings.statusPageStarredReposDesc', icon: <Star size={14} /> },
]
const STATUS_OPTIONS: StatusOption[] = [
  { key: 'imported', label: 'settings.statusImported' },
  { key: 'updatable', label: 'settings.statusUpdatable' },
  { key: 'pushedAgents', label: 'settings.statusPushedTools' },
]

function clampRepoScanMaxDepth(value: number) {
  if (!Number.isFinite(value) || value < minRepoScanMaxDepth) {
    return defaultRepoScanMaxDepth
  }
  if (value > maxRepoScanMaxDepth) {
    return maxRepoScanMaxDepth
  }
  return value
}

function getCloudProviderDisplayName(name: string, t: (key: TranslationKey) => string) {
  const key = CLOUD_PROVIDER_LABEL_KEYS[name]
  return key ? t(key) : name
}

function getCloudFieldLabel(providerName: string | undefined, fieldKey: string, fallback: string, t: (key: TranslationKey) => string) {
  const key = providerName ? CLOUD_FIELD_LABEL_KEYS[providerName]?.[fieldKey] : undefined
  return key ? t(key) : fallback
}

function orderCloudProviders(providerList: any[]) {
  return providerList
    .map((provider, index) => ({ provider, index }))
    .sort((left, right) => {
      const leftPriority = CLOUD_PROVIDER_PRIORITY[left.provider?.name] ?? Number.MAX_SAFE_INTEGER
      const rightPriority = CLOUD_PROVIDER_PRIORITY[right.provider?.name] ?? Number.MAX_SAFE_INTEGER
      if (leftPriority !== rightPriority) {
        return leftPriority - rightPriority
      }
      return left.index - right.index
    })
    .map(({ provider }) => provider)
}

function splitCloudRemoteSegments(value: string | undefined) {
  return (value ?? '')
    .replace(/\\/g, '/')
    .split('/')
    .map(part => part.trim())
    .filter(part => part && part !== '.')
}

function getCloudRemotePathInputValue(storedRemotePath: string | undefined) {
  const parts = splitCloudRemoteSegments(storedRemotePath)
  if (parts[parts.length - 1] === CLOUD_REMOTE_ROOT_DIR) {
    return parts.slice(0, -1).join('/')
  }
  return parts.join('/')
}

function buildStoredCloudRemotePath(inputValue: string | undefined) {
  const parts = splitCloudRemoteSegments(inputValue)
  if (parts[parts.length - 1] !== CLOUD_REMOTE_ROOT_DIR) {
    parts.push(CLOUD_REMOTE_ROOT_DIR)
  }
  return `${parts.join('/')}/`
}

function buildCloudBackupPreviewPath(bucketName: string | undefined, storedRemotePath: string | undefined, bucketPlaceholder: string) {
  const bucket = (bucketName ?? '').trim() || bucketPlaceholder
  return `${bucket}/${buildStoredCloudRemotePath(getCloudRemotePathInputValue(storedRemotePath))}`
}

function buildEmptyCloudProfile() {
  return {
    bucketName: '',
    remotePath: buildStoredCloudRemotePath(''),
    credentials: {} as Record<string, string>,
  }
}

function getCloudProfileDraft(source: any, providerName: string) {
  const profile = source?.cloudProfiles?.[providerName] ?? buildEmptyCloudProfile()
  return {
    bucketName: profile.bucketName ?? '',
    remotePath: profile.remotePath ?? buildStoredCloudRemotePath(''),
    credentials: { ...(profile.credentials ?? {}) },
  }
}

function syncActiveCloudProfile(source: any) {
  const provider = source?.cloud?.provider
  if (!provider) {
    return source
  }
  return {
    ...source,
    cloudProfiles: {
      ...(source?.cloudProfiles ?? {}),
      [provider]: {
        bucketName: source?.cloud?.bucketName ?? '',
        remotePath: source?.cloud?.remotePath ?? buildStoredCloudRemotePath(''),
        credentials: { ...(source?.cloud?.credentials ?? {}) },
      },
    },
  }
}

function buildCloudFromProfile(source: any, providerName: string) {
  const profile = getCloudProfileDraft(source, providerName)
  return {
    ...source?.cloud,
    provider: providerName,
    bucketName: profile.bucketName,
    remotePath: profile.remotePath,
    credentials: profile.credentials,
  }
}

function ThemeOptionCard({ option, active, onSelect }: { option: ThemeOption; active: boolean; onSelect: (theme: Theme) => void }) {
  return (
    <button
      onClick={() => onSelect(option.id)}
      className="group relative overflow-hidden rounded-2xl p-3 text-left transition-all duration-300"
      style={{
        background: active ? 'var(--bg-elevated)' : 'var(--bg-surface)',
        border: active ? '1px solid var(--border-accent)' : '1px solid var(--border-base)',
        boxShadow: active ? 'var(--shadow-card), var(--glow-accent-sm)' : 'var(--shadow-card)',
        transform: active ? 'translateY(-1px)' : 'none',
      }}
    >
      <div
        className="relative mb-3 h-28 overflow-hidden rounded-[18px]"
        style={{
          background: option.preview.shell,
          border: `1px solid ${option.preview.divider}`,
        }}
      >
        <div
          className="absolute inset-y-0 left-0"
          style={{
            width: '34%',
            background: option.preview.sidebar,
            borderRight: `1px solid ${option.preview.divider}`,
          }}
        />
        <div
          className="absolute left-3 top-3 h-6 rounded-xl"
          style={{
            width: 'calc(34% - 24px)',
            background: option.preview.sidebarSelection,
            boxShadow: `0 10px 22px ${option.preview.accentGlow}`,
          }}
        />
        <div
          className="absolute right-4 top-4 h-4 rounded-full"
          style={{
            left: '40%',
            background: option.preview.search,
          }}
        />
        <div
          className="absolute right-10 top-11 h-9 rounded-2xl"
          style={{
            left: '40%',
            background: option.preview.panel,
            boxShadow: `0 14px 28px ${option.preview.accentGlow}`,
          }}
        />
        <div
          className="absolute top-[53px] h-2 rounded-full"
          style={{
            left: '44%',
            width: '4rem',
            background: option.preview.text,
            opacity: 0.78,
          }}
        />
        <div
          className="absolute top-[66px] h-2 rounded-full"
          style={{
            left: '44%',
            width: '2.75rem',
            background: option.preview.textMuted,
            opacity: 0.55,
          }}
        />
        <div
          className="absolute bottom-4 right-4 h-9 w-9 rounded-2xl"
          style={{
            background: option.preview.accent,
            boxShadow: `0 12px 26px ${option.preview.accentGlow}`,
          }}
        />
        <div
          className="absolute bottom-6 h-2 rounded-full"
          style={{
            left: '52%',
            width: '3rem',
            background: option.preview.textMuted,
            opacity: 0.4,
          }}
        />
      </div>

      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span
              className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl"
              style={{
                background: active ? 'var(--accent-glow)' : 'var(--bg-overlay)',
                color: active ? 'var(--accent-primary)' : 'var(--text-secondary)',
                border: '1px solid var(--border-base)',
              }}
            >
              {option.icon}
            </span>
            <div className="min-w-0">
              <p className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>{option.label}</p>
              <p className="text-[11px] uppercase tracking-[0.18em]" style={{ color: 'var(--text-muted)' }}>{option.tone}</p>
            </div>
          </div>
          <p className="mt-3 text-xs leading-5" style={{ color: 'var(--text-secondary)' }}>{option.description}</p>
        </div>

        <span
          className="mt-0.5 flex h-6 w-6 items-center justify-center rounded-full"
          style={{
            background: active ? 'var(--accent-primary)' : 'transparent',
            color: active ? '#ffffff' : 'var(--text-disabled)',
            border: active ? 'none' : '1px solid var(--border-base)',
          }}
        >
          {active ? <Check size={14} /> : <div className="h-2.5 w-2.5 rounded-full" style={{ background: 'var(--bg-overlay)' }} />}
        </span>
      </div>
    </button>
  )
}

function Toggle({ enabled, onToggle }: { enabled: boolean; onToggle: () => void }) {
  return (
    <div
      onClick={onToggle}
      className="w-9 h-5 rounded-full relative cursor-pointer transition-all duration-200"
      style={{
        background: enabled ? 'var(--accent-secondary)' : 'var(--bg-overlay)',
        boxShadow: enabled ? 'var(--glow-accent-sm)' : 'none',
        border: '1px solid var(--border-base)',
      }}
    >
      <div
        className={`absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform duration-200 ${enabled ? 'translate-x-4' : 'translate-x-0.5'}`}
        style={{ boxShadow: '0 1px 3px rgba(0,0,0,0.3)' }}
      />
    </div>
  )
}

function StatusToggleChip({
  active,
  label,
  onClick,
}: {
  active: boolean
  label: string
  onClick: () => void
}) {
  return (
    <button
      onClick={onClick}
      className="rounded-lg px-3 py-1.5 text-sm transition-all duration-200"
      style={active ? {
        background: 'var(--accent-glow)',
        color: 'var(--text-primary)',
        border: '1px solid var(--border-accent)',
        boxShadow: 'var(--glow-accent-sm)',
      } : {
        background: 'var(--bg-elevated)',
        color: 'var(--text-secondary)',
        border: '1px solid var(--border-base)',
      }}
    >
      {label}
    </button>
  )
}

export default function SettingsPage() {
  const { theme, setTheme } = useThemeContext()
  const { t, lang, setLang } = useLanguage()
  const { syncFromConfig } = useSkillStatusVisibilityContext()
  const [tab, setTab] = useState<Tab>('agents')
  const [cfg, setCfg] = useState<any>(null)
  const [providers, setProviders] = useState<any[]>([])
  const [saving, setSaving] = useState(false)
  const [newAgent, setNewAgent] = useState({ name: '', pushDir: '' })
  const [newScanDirs, setNewScanDirs] = useState<Record<string, string>>({})
  const [appVersion, setAppVersion] = useState('')
  const [logDir, setLogDir] = useState('')
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [updateResult, setUpdateResult] = useState<string | null>(null)
  const [proxyTestURL, setProxyTestURL] = useState(DEFAULT_PROXY_TEST_URL)
  const [testingProxy, setTestingProxy] = useState(false)
  const [proxyTestResult, setProxyTestResult] = useState<{ tone: 'success' | 'error', text: string } | null>(null)
  const cfgRef = useRef<any>(null)
  const savingRef = useRef(false)

  const themeOptions: ThemeOption[] = [
    {
      id: 'dark',
      label: 'Dark',
      tone: 'Ink Slate',
      description: t('settings.themeDark'),
      icon: <Moon size={15} />,
      preview: {
        shell: 'radial-gradient(circle at top right, rgba(154,168,193,0.12), transparent 28%), linear-gradient(180deg, #13171d 0%, #0f1318 100%)',
        sidebar: 'rgba(20, 24, 31, 0.94)',
        sidebarSelection: 'rgba(167, 183, 207, 0.12)',
        search: 'rgba(255,255,255,0.06)',
        panel: 'rgba(29, 35, 44, 0.94)',
        accent: '#a7b7cf',
        accentGlow: 'rgba(116, 132, 159, 0.22)',
        text: '#edf1f7',
        textMuted: '#7e8a9c',
        divider: 'rgba(255,255,255,0.07)',
      },
    },
    {
      id: 'young',
      label: 'Young',
      tone: 'Breeze Paper',
      description: t('settings.themeYoung'),
      icon: <Sparkles size={15} />,
      preview: {
        shell: 'radial-gradient(circle at top right, rgba(147,197,253,0.22), transparent 30%), radial-gradient(circle at bottom left, rgba(251,191,36,0.1), transparent 32%), linear-gradient(180deg, #f6faff 0%, #edf4fc 52%, #fffdf8 100%)',
        sidebar: 'rgba(230, 238, 248, 0.98)',
        sidebarSelection: 'rgba(76, 127, 198, 0.22)',
        search: 'rgba(67, 94, 131, 0.18)',
        panel: 'rgba(255, 255, 255, 0.99)',
        accent: '#4c7fc6',
        accentGlow: 'rgba(76, 127, 198, 0.22)',
        text: '#203854',
        textMuted: '#586f8a',
        divider: 'rgba(86, 111, 138, 0.18)',
      },
    },
    {
      id: 'light',
      label: 'Light',
      tone: 'Messor Calm',
      description: t('settings.themeLight'),
      icon: <Sun size={15} />,
      preview: {
        shell: 'linear-gradient(180deg, #f7f9fc 0%, #edf1f7 100%)',
        sidebar: 'rgba(231, 235, 242, 0.97)',
        sidebarSelection: 'rgba(20, 85, 217, 0.18)',
        search: 'rgba(54, 65, 82, 0.12)',
        panel: 'rgba(255, 255, 255, 0.99)',
        accent: '#1f5ce0',
        accentGlow: 'rgba(20, 85, 217, 0.22)',
        text: '#172434',
        textMuted: '#59687b',
        divider: 'rgba(15, 23, 42, 0.12)',
      },
    },
  ]

  useEffect(() => {
    Promise.all([GetConfig(), ListCloudProviders(), GetAppVersion(), GetLogDir()]).then(([c, p, v, logPath]) => {
      setCfg(syncActiveCloudProfile(c))
      setProviders(orderCloudProviders(p ?? []))
      setAppVersion(v as string)
      setLogDir(logPath as string)
    })
  }, [])

  useEffect(() => {
    cfgRef.current = cfg
    savingRef.current = saving
  }, [cfg, saving])

  const checkUpdate = async () => {
    setCheckingUpdate(true)
    setUpdateResult(null)
    try {
      const info = await CheckAppUpdateAndNotify()
      if (info.hasUpdate) {
        setUpdateResult(t('settings.updateFound', { version: info.latestVersion }))
      } else {
        setUpdateResult(t('settings.updateLatest', { version: info.currentVersion }))
      }
    } catch (e: any) {
      setUpdateResult(t('settings.updateFailed', { msg: e?.message ?? String(e) }))
    } finally {
      setCheckingUpdate(false)
    }
  }

  const testProxyConnection = async () => {
    const payload = buildProxyConnectionPayload(proxyTestURL, cfg?.proxy)
    setTestingProxy(true)
    setProxyTestResult(null)
    try {
      const result = await TestProxyConnection(payload.targetURL, payload.proxy)
      if (result.success) {
        setProxyTestResult({
          tone: 'success',
          text: t('settings.proxyTestSuccess', {
            url: result.targetURL,
            status: String(result.statusCode),
            ms: String(result.elapsedMs),
          }),
        })
      } else {
        setProxyTestResult({
          tone: 'error',
          text: t('settings.proxyTestFailure', {
            url: result.targetURL,
            msg: result.message,
            ms: String(result.elapsedMs),
          }),
        })
      }
    } catch (e: any) {
      setProxyTestResult({
        tone: 'error',
        text: t('settings.proxyTestFailure', {
          url: payload.targetURL,
          msg: e?.message ?? String(e),
          ms: '0',
        }),
      })
    } finally {
      setTestingProxy(false)
    }
  }

  const save = useCallback(async () => {
    const currentCfg = cfgRef.current
    if (!currentCfg || savingRef.current) return

    savingRef.current = true
    setSaving(true)
    try {
      const nextCfg = syncActiveCloudProfile({
        ...currentCfg,
        skillStatusVisibility: normalizeSkillStatusVisibility(currentCfg?.skillStatusVisibility),
      })
      await SaveConfig(nextCfg)
      setCfg(nextCfg)
      syncFromConfig(nextCfg)
    } finally {
      savingRef.current = false
      setSaving(false)
    }
  }, [syncFromConfig])

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      const key = event.key.toLowerCase()
      const isSaveShortcut = (event.ctrlKey || event.metaKey)
        && !event.altKey
        && !event.shiftKey
        && (event.code === 'KeyS' || key === 's')

      if (!isSaveShortcut || event.defaultPrevented || event.isComposing) {
        return
      }

      event.preventDefault()
      event.stopPropagation()
      void save()
    }

    document.addEventListener('keydown', handleKeyDown, true)
    return () => document.removeEventListener('keydown', handleKeyDown, true)
  }, [save])

  const updateActiveCloud = (patch: Record<string, any>) => {
    setCfg((prev: any) => syncActiveCloudProfile({
      ...prev,
      cloud: {
        ...prev.cloud,
        ...patch,
      },
    }))
  }

  const updateAgent = (name: string, field: string, value: any) => {
    setCfg((prev: any) => ({
      ...prev,
      agents: prev.agents.map((agent: any) => agent.name === name ? { ...agent, [field]: value } : agent)
    }))
  }

  const addScanDir = (name: string) => {
    const path = (newScanDirs[name] ?? '').trim()
    if (!path) return
    setCfg((prev: any) => ({
      ...prev,
      agents: prev.agents.map((agent: any) => {
        if (agent.name !== name) return agent
        const current = agent.scanDirs ?? []
        if (current.includes(path)) return agent
        return { ...agent, scanDirs: [...current, path] }
      })
    }))
    setNewScanDirs((prev) => ({ ...prev, [name]: '' }))
  }

  const updateScanDir = (name: string, index: number, value: string) => {
    setCfg((prev: any) => ({
      ...prev,
      agents: prev.agents.map((agent: any) => {
        if (agent.name !== name) return agent
        const next = [...(agent.scanDirs ?? [])]
        next[index] = value
        return { ...agent, scanDirs: next }
      })
    }))
  }

  const removeScanDir = (name: string, index: number) => {
    setCfg((prev: any) => ({
      ...prev,
      agents: prev.agents.map((agent: any) => {
        if (agent.name !== name) return agent
        return { ...agent, scanDirs: (agent.scanDirs ?? []).filter((_: string, i: number) => i !== index) }
      })
    }))
  }

  const setProxyMode = (mode: ProxyMode) => {
    setCfg((prev: any) => ({ ...prev, proxy: { ...prev.proxy, mode } }))
  }

  const setProxyURL = (url: string) => {
    setCfg((prev: any) => ({ ...prev, proxy: { ...prev.proxy, url } }))
  }

  const pickDir = async (onPick: (path: string) => void, currentPath = '') => {
    const dir = await OpenFolderDialog(currentPath)
    if (dir) onPick(dir)
  }

  const selectedProvider = providers.find((p: any) => p.name === cfg?.cloud?.provider)
  const proxyMode: ProxyMode = (cfg?.proxy?.mode as ProxyMode) || 'none'
  const cloudRemotePathInput = getCloudRemotePathInputValue(cfg?.cloud?.remotePath)
  const cloudBackupPreviewPath = buildCloudBackupPreviewPath(cfg?.cloud?.bucketName, cfg?.cloud?.remotePath, t('settings.remotePathBucketPlaceholder'))
  const statusVisibility = normalizeSkillStatusVisibility(cfg?.skillStatusVisibility)

  if (!cfg) return <div className="p-8" style={{ color: 'var(--text-muted)' }}>{t('common.loading')}</div>

  return (
    <div className="min-h-full w-full max-w-6xl p-6 md:p-8">
      <div className="mb-6 flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
        <h2 className="text-lg font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
          <Settings size={18} /> {t('settings.title')}
        </h2>
        <div className="flex flex-wrap items-center gap-3 xl:justify-end">
          {updateResult && (
            <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{updateResult}</span>
          )}
          {appVersion && (
            <span className="text-xs font-mono" style={{ color: 'var(--text-muted)' }}>
              {appVersion === 'dev' ? 'dev' : appVersion.startsWith('v') ? appVersion : `v${appVersion}`}
            </span>
          )}
          <button
            onClick={checkUpdate}
            disabled={checkingUpdate}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-lg transition-colors disabled:opacity-50"
            style={{ background: 'var(--bg-elevated)', color: 'var(--text-secondary)', border: '1px solid var(--border-base)' }}
          >
            <RefreshCw size={12} className={checkingUpdate ? 'animate-spin' : ''} />
            {checkingUpdate ? t('settings.checkingUpdate') : t('settings.checkUpdate')}
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div
        className="mb-6 flex w-fit max-w-full flex-wrap gap-1 rounded-xl p-1"
        style={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-base)' }}
      >
        {(['agents', 'cloud', 'network', 'general'] as Tab[]).map(tabKey => {
          const labels: Record<Tab, string> = {
            agents: t('settings.tabTools'),
            cloud: t('settings.tabCloud'),
            general: t('settings.tabGeneral'),
            network: t('settings.tabNetwork'),
          }
          return (
            <button
              key={tabKey}
              onClick={() => setTab(tabKey)}
              className="px-4 py-1.5 rounded-lg text-sm transition-all duration-200"
              style={tab === tabKey ? {
                background: 'var(--bg-overlay)',
                color: 'var(--text-primary)',
                boxShadow: 'var(--glow-accent-sm)',
                border: '1px solid var(--border-accent)',
              } : {
                color: 'var(--text-muted)',
                border: '1px solid transparent',
              }}
            >{labels[tabKey]}</button>
          )
        })}
      </div>

      {/* Agents tab */}
      {tab === 'agents' && (
        <div className="space-y-4">
          {(cfg.agents ?? []).map((agent: any) => (
            <div
              key={agent.name}
              className="rounded-xl p-4"
              style={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-base)' }}
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2.5">
                  <ToolIcon name={agent.name} size={28} />
                  <span className="font-medium text-sm" style={{ color: 'var(--text-primary)' }}>{agent.name}</span>
                </div>
                <label className="flex items-center gap-2 cursor-pointer">
                  <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{t('settings.toolEnabled')}</span>
                  <Toggle enabled={agent.enabled} onToggle={() => updateAgent(agent.name, 'enabled', !agent.enabled)} />
                </label>
              </div>

              <div className="mb-3">
                <p className="text-xs mb-1.5" style={{ color: 'var(--text-muted)' }}>{t('settings.pushPath')}</p>
                <div className="flex gap-2">
                  <input
                    value={agent.pushDir ?? ''}
                    onChange={e => updateAgent(agent.name, 'pushDir', e.target.value)}
                    className="input-base flex-1 font-mono"
                  />
                  <button
                    onClick={() => pickDir(dir => updateAgent(agent.name, 'pushDir', dir), agent.pushDir ?? '')}
                    className="btn-secondary px-2.5 rounded-lg"
                    title={t('settings.selectDir')}
                  >
                    <FolderOpen size={14} />
                  </button>
                </div>
              </div>

              <div>
                <p className="text-xs mb-1.5" style={{ color: 'var(--text-muted)' }}>{t('settings.scanPaths')}</p>
                <div className="space-y-2">
                  {(agent.scanDirs ?? []).map((dir: string, idx: number) => (
                    <div key={`${agent.name}-scan-${idx}`} className="flex gap-2">
                      <input
                        value={dir}
                        onChange={e => updateScanDir(agent.name, idx, e.target.value)}
                        className="input-base flex-1 font-mono"
                      />
                      <button
                        onClick={() => pickDir(d => updateScanDir(agent.name, idx, d), dir ?? '')}
                        className="btn-secondary px-2.5 rounded-lg"
                        title={t('settings.selectDir')}
                      >
                        <FolderOpen size={14} />
                      </button>
                      <button
                        onClick={() => removeScanDir(agent.name, idx)}
                        className="btn-secondary px-2.5 rounded-lg"
                        title={t('settings.deleteScanPath')}
                      >
                        <Trash2 size={14} />
                      </button>
                    </div>
                  ))}
                </div>
                <div className="mt-2 flex gap-2">
                  <input
                    value={newScanDirs[agent.name] ?? ''}
                    onChange={e => setNewScanDirs(prev => ({ ...prev, [agent.name]: e.target.value }))}
                    placeholder="/path/to/scan"
                    className="input-base flex-1 font-mono"
                  />
                  <button
                    onClick={() => pickDir(d => setNewScanDirs(prev => ({ ...prev, [agent.name]: d })), newScanDirs[agent.name] ?? '')}
                    className="btn-secondary px-2.5 rounded-lg"
                    title={t('settings.selectDir')}
                  >
                    <FolderOpen size={14} />
                  </button>
                  <button
                    onClick={() => addScanDir(agent.name)}
                    className="btn-secondary px-3 py-1.5 rounded-lg text-sm flex items-center gap-1"
                  >
                    <Plus size={14} /> {t('settings.addPath')}
                  </button>
                </div>
              </div>

              {agent.custom && (
                <button
                  onClick={async () => { await RemoveCustomAgent(agent.name); const c = await GetConfig(); setCfg(syncActiveCloudProfile(c)) }}
                  className="mt-2 text-xs flex items-center gap-1 transition-colors"
                  style={{ color: 'var(--color-error)' }}
                >
                  <Trash2 size={12} /> {t('settings.deleteTool')}
                </button>
              )}
            </div>
          ))}

          {/* Add custom agent */}
          <div
            className="rounded-xl p-4"
            style={{ border: '1px dashed var(--border-surface)', background: 'var(--bg-surface)' }}
          >
            <p className="text-sm mb-3" style={{ color: 'var(--text-muted)' }}>{t('settings.addCustomTool')}</p>
            <div className="flex gap-2 mb-2">
              <input
                value={newAgent.name}
                onChange={e => setNewAgent(p => ({ ...p, name: e.target.value }))}
                placeholder={t('settings.toolName')}
                className="input-base flex-1"
              />
            </div>
            <div className="flex gap-2">
              <input
                value={newAgent.pushDir}
                onChange={e => setNewAgent(p => ({ ...p, pushDir: e.target.value }))}
                placeholder="/path/to/push"
                className="input-base flex-1 font-mono"
              />
              <button
                onClick={() => pickDir(d => setNewAgent(p => ({ ...p, pushDir: d })), newAgent.pushDir)}
                className="btn-secondary px-2.5 rounded-lg"
                title={t('settings.selectDir')}
              >
                <FolderOpen size={14} />
              </button>
              <button
                onClick={async () => {
                  if (newAgent.name && newAgent.pushDir) {
                    await AddCustomAgent(newAgent.name, newAgent.pushDir)
                    const c = await GetConfig(); setCfg(syncActiveCloudProfile(c))
                    setNewAgent({ name: '', pushDir: '' })
                  }
                }}
                className="btn-primary px-3 py-1.5 rounded-lg text-sm flex items-center gap-1"
              >
                <Plus size={14} /> {t('settings.addPath')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Cloud tab */}
      {tab === 'cloud' && (
        <div className="space-y-4">
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.cloudProvider')}</p>
            <div className="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
              {providers.map((p: any) => (
                <button
                  key={p.name}
                  onClick={() => setCfg((prev: any) => {
                    const synced = syncActiveCloudProfile(prev)
                    return {
                      ...synced,
                      cloud: buildCloudFromProfile(synced, p.name),
                    }
                  })}
                  className="flex min-h-[84px] items-center justify-center rounded-xl px-4 py-3 text-center text-sm leading-snug transition-all duration-200 whitespace-normal"
                  style={cfg.cloud?.provider === p.name ? {
                    background: 'var(--accent-glow)',
                    color: 'var(--accent-primary)',
                    border: '1px solid var(--border-accent)',
                    boxShadow: 'var(--glow-accent-sm)',
                  } : {
                    background: 'var(--bg-elevated)',
                    color: 'var(--text-secondary)',
                    border: '1px solid var(--border-base)',
                  }}
                >
                  {getCloudProviderDisplayName(p.name, t)}
                </button>
              ))}
            </div>
          </div>

          {selectedProvider && (
            <>
              {cfg.cloud?.provider !== 'git' && (
                <>
                  <div>
                    <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.bucket')}</p>
                    <input
                      value={cfg.cloud?.bucketName ?? ''}
                      onChange={e => updateActiveCloud({ bucketName: e.target.value })}
                      className="input-base"
                    />
                  </div>

                  <div>
                    <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.remotePath')}</p>
                    <input
                      value={cloudRemotePathInput}
                      onChange={e => updateActiveCloud({ remotePath: buildStoredCloudRemotePath(e.target.value) })}
                      placeholder="team-a/nightly"
                      className="input-base font-mono"
                    />
                    <p className="mt-2 text-xs leading-5" style={{ color: 'var(--text-muted)' }}>
                      {t('settings.remotePathHint')}
                    </p>
                  </div>

                  <div
                    className="relative overflow-hidden rounded-2xl px-4 py-3"
                    style={{
                      background: 'linear-gradient(135deg, color-mix(in srgb, var(--accent-glow) 58%, transparent) 0%, color-mix(in srgb, var(--bg-elevated) 92%, transparent) 100%)',
                      border: '1px solid color-mix(in srgb, var(--border-accent) 68%, var(--border-base) 32%)',
                      boxShadow: 'var(--shadow-card)',
                    }}
                  >
                    <div
                      className="pointer-events-none absolute inset-y-0 right-0 w-28"
                      style={{
                        background: 'radial-gradient(circle at top right, color-mix(in srgb, var(--accent-primary) 28%, transparent) 0%, transparent 72%)',
                        opacity: 0.8,
                      }}
                    />
                    <p className="text-xs uppercase tracking-[0.18em]" style={{ color: 'var(--text-muted)' }}>
                      {t('settings.remotePathPreview')}
                    </p>
                    <p className="mt-2 text-sm font-mono break-all" style={{ color: 'var(--text-primary)' }}>
                      {cloudBackupPreviewPath}
                    </p>
                  </div>
                </>
              )}
              {selectedProvider.fields.map((f: any) => (
                <div key={f.key}>
                  <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{getCloudFieldLabel(cfg.cloud?.provider, f.key, f.label, t)}</p>
                  <input
                    type={f.secret ? 'password' : 'text'}
                    placeholder={f.placeholder ?? ''}
                    value={cfg.cloud?.credentials?.[f.key] ?? ''}
                    onChange={e => updateActiveCloud({
                      credentials: { ...(cfg.cloud?.credentials ?? {}), [f.key]: e.target.value },
                    })}
                    className="input-base font-mono"
                  />
                </div>
              ))}
              <div>
                <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.syncInterval')}</p>
                <input
                  type="number"
                  min={0}
                  value={cfg.cloud?.syncIntervalMinutes ?? 0}
                  onChange={e => updateActiveCloud({ syncIntervalMinutes: parseInt(e.target.value) || 0 })}
                  className="input-base w-32"
                />
              </div>
              <label className="flex items-center gap-3 cursor-pointer">
                <Toggle
                  enabled={!!cfg.cloud?.enabled}
                  onToggle={() => updateActiveCloud({ enabled: !cfg.cloud?.enabled })}
                />
                <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>{t('settings.enableAutoBackup')}</span>
              </label>
            </>
          )}
        </div>
      )}

      {/* General tab */}
      {tab === 'general' && (
        <div className="space-y-4">
          {/* Language */}
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.language')}</p>
            <div className="flex gap-2">
              {(['zh', 'en'] as const).map(l => (
                <button
                  key={l}
                  onClick={() => setLang(l)}
                  className="px-4 py-1.5 rounded-lg text-sm transition-all duration-200"
                  style={lang === l ? {
                    background: 'var(--accent-glow)',
                    color: 'var(--accent-primary)',
                    border: '1px solid var(--border-accent)',
                    boxShadow: 'var(--glow-accent-sm)',
                  } : {
                    background: 'var(--bg-elevated)',
                    color: 'var(--text-secondary)',
                    border: '1px solid var(--border-base)',
                  }}
                >
                  {l === 'zh' ? '中文' : 'English'}
                </button>
              ))}
            </div>
          </div>

          {/* Theme */}
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.theme')}</p>
            <div className="grid gap-3 md:grid-cols-3">
              {themeOptions.map((option) => (
                <ThemeOptionCard
                  key={option.id}
                  option={option}
                  active={theme === option.id}
                  onSelect={setTheme}
                />
              ))}
            </div>
            <p className="mt-2 text-xs leading-5" style={{ color: 'var(--text-muted)' }}>
              {t('settings.themeHint')}
            </p>
          </div>

          <div>
            <div className="mb-2 flex items-center justify-between gap-3">
              <div>
                <p className="text-sm" style={{ color: 'var(--text-muted)' }}>{t('settings.skillStatusVisibility')}</p>
                <p className="mt-1 text-xs leading-5" style={{ color: 'var(--text-muted)' }}>{t('settings.skillStatusVisibilityHint')}</p>
              </div>
            </div>
            <div
              className="overflow-hidden rounded-2xl"
              style={{ border: '1px solid var(--border-base)', background: 'var(--bg-surface)' }}
            >
              {STATUS_PAGE_OPTIONS.map((page) => (
                <div
                  key={page.key}
                  className="flex flex-col gap-3 px-4 py-3 md:flex-row md:items-center md:justify-between"
                  style={{
                    borderBottom: page.key === STATUS_PAGE_OPTIONS[STATUS_PAGE_OPTIONS.length - 1].key ? 'none' : '1px solid var(--border-base)',
                  }}
                >
                  <div className="min-w-0 flex items-start gap-3">
                    <div className="mt-0.5 flex h-8 w-8 items-center justify-center rounded-xl" style={{
                      background: 'var(--bg-elevated)',
                      border: '1px solid var(--border-base)',
                      color: 'var(--accent-primary)',
                    }}>
                      {page.icon}
                    </div>
                    <div className="min-w-0">
                      <p className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t(page.label)}</p>
                      <p className="mt-1 text-xs leading-5" style={{ color: 'var(--text-muted)' }}>{t(page.description)}</p>
                    </div>
                  </div>
                  <div className="flex flex-wrap gap-2 md:justify-end">
                    {STATUS_OPTIONS.filter((status) => DEFAULT_SKILL_STATUS_VISIBILITY[page.key].includes(status.key)).map((status) => (
                      <StatusToggleChip
                        key={`${page.key}-${status.key}`}
                        active={statusVisibility[page.key].includes(status.key)}
                        label={t(status.label)}
                        onClick={() => setCfg((prev: any) => ({
                          ...prev,
                          skillStatusVisibility: (() => {
                            const prevVisibility = normalizeSkillStatusVisibility(prev?.skillStatusVisibility)
                            const currentlyEnabled = prevVisibility[page.key].includes(status.key)
                            return toggleSkillStatusForPage(
                              prevVisibility,
                              page.key,
                              status.key,
                              !currentlyEnabled,
                            )
                          })(),
                        }))}
                      />
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.logLevel')}</p>
            <div className="flex gap-2 mb-2">
              {([
                ['debug', 'Debug'],
                ['info', 'Info'],
                ['error', 'Error'],
              ] as [string, string][]).map(([level, label]) => (
                <button
                  key={level}
                  onClick={() => setCfg((p: any) => ({ ...p, logLevel: level }))}
                  className="px-3 py-1.5 rounded-lg text-sm transition-all duration-200"
                  style={(cfg.logLevel ?? 'error') === level ? {
                    background: 'var(--accent-glow)',
                    color: 'var(--accent-primary)',
                    border: '1px solid var(--border-accent)',
                    boxShadow: 'var(--glow-accent-sm)',
                  } : {
                    background: 'var(--bg-elevated)',
                    color: 'var(--text-secondary)',
                    border: '1px solid var(--border-base)',
                  }}
                >
                  {label}
                </button>
              ))}
            </div>
            <p className="text-xs" style={{ color: 'var(--text-muted)' }}>{t('settings.logLevelHint')}</p>
          </div>
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.launchAtLogin')}</p>
            <label className="flex items-center gap-3 cursor-pointer">
              <Toggle
                enabled={!!cfg.launchAtLogin}
                onToggle={() => setCfg((p: any) => ({ ...p, launchAtLogin: !p.launchAtLogin }))}
              />
              <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>{t('settings.launchAtLoginHint')}</span>
            </label>
          </div>
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.logDir')}</p>
            <div className="flex items-center gap-2">
              <button
                onClick={async () => { await OpenLogDir() }}
                className="btn-secondary px-3 py-1.5 rounded-lg text-sm"
              >
                {t('settings.openLogDir')}
              </button>
              <span className="text-xs font-mono break-all" style={{ color: 'var(--text-muted)' }}>{logDir}</span>
            </div>
            <p className="mt-1.5 text-xs" style={{ color: 'var(--text-muted)' }}>{t('settings.logDirHint')}</p>
          </div>
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.skillsDir')}</p>
            <div className="flex gap-2">
              <input
                value={cfg.skillsStorageDir ?? ''}
                onChange={e => setCfg((p: any) => ({ ...p, skillsStorageDir: e.target.value }))}
                className="input-base flex-1 font-mono"
              />
              <button
                onClick={() => pickDir(d => setCfg((p: any) => ({ ...p, skillsStorageDir: d })), cfg.skillsStorageDir ?? '')}
                className="btn-secondary px-2.5 rounded-lg"
                title={t('settings.selectDir')}
              >
                <FolderOpen size={16} />
              </button>
            </div>
          </div>
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.repoScanMaxDepth')}</p>
            <input
              type="number"
              min={minRepoScanMaxDepth}
              max={maxRepoScanMaxDepth}
              value={cfg.repoScanMaxDepth ?? defaultRepoScanMaxDepth}
              onChange={e => setCfg((p: any) => ({
                ...p,
                repoScanMaxDepth: clampRepoScanMaxDepth(parseInt(e.target.value, 10)),
              }))}
              className="input-base w-32"
            />
            <p className="mt-1.5 text-xs" style={{ color: 'var(--text-muted)' }}>{t('settings.repoScanMaxDepthHint')}</p>
          </div>
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.defaultCategory')}</p>
            <div
              className="rounded-lg px-3 py-2 text-sm"
              style={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-base)', color: 'var(--text-secondary)' }}
            >
              {t('settings.defaultCategoryValue')}
            </div>
            <p className="mt-1.5 text-xs" style={{ color: 'var(--text-muted)' }}>{t('settings.defaultCategoryHint')}</p>
          </div>
        </div>
      )}

      {/* Network tab */}
      {tab === 'network' && (
        <div className="space-y-6">
          <div>
            <p className="text-sm mb-1 flex items-center gap-1.5" style={{ color: 'var(--text-muted)' }}>
              <Globe size={14} /> {t('settings.proxy')}
            </p>
            <p className="text-xs mb-4" style={{ color: 'var(--text-muted)' }}>
              {t('settings.proxyHint')}
            </p>

            <div className="space-y-2">
              {([
                ['none',   t('settings.proxyNone'),   t('settings.proxyNoneDesc')],
                ['system', t('settings.proxySystem'), t('settings.proxySystemDesc')],
                ['manual', t('settings.proxyManual'), t('settings.proxyManualDesc')],
              ] as [ProxyMode, string, string][]).map(([mode, label, desc]) => (
                <div
                  key={mode}
                  onClick={() => setProxyMode(mode)}
                  className="flex items-start gap-3 p-3 rounded-xl cursor-pointer transition-all duration-200 select-none"
                  style={proxyMode === mode ? {
                    background: 'var(--accent-glow)',
                    border: '1px solid var(--border-accent)',
                  } : {
                    background: 'var(--bg-elevated)',
                    border: '1px solid var(--border-base)',
                  }}
                >
                  <div
                    className="mt-0.5 w-4 h-4 rounded-full border-2 flex items-center justify-center shrink-0 transition-all duration-200"
                    style={proxyMode === mode ? {
                      borderColor: 'var(--accent-secondary)',
                      background: 'var(--accent-secondary)',
                    } : {
                      borderColor: 'var(--text-muted)',
                    }}
                  >
                    {proxyMode === mode && <div className="w-1.5 h-1.5 bg-white rounded-full" />}
                  </div>
                  <div>
                    <p className="text-sm font-medium leading-snug" style={{ color: 'var(--text-primary)' }}>{label}</p>
                    <p className="text-xs mt-0.5" style={{ color: 'var(--text-muted)' }}>{desc}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {proxyMode === 'manual' && (
            <div>
              <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.proxyUrl')}</p>
              <input
                value={cfg.proxy?.url ?? ''}
                onChange={e => setProxyURL(e.target.value)}
                placeholder="http://127.0.0.1:7890"
                className="input-base font-mono"
              />
              <p className="mt-1.5 text-xs" style={{ color: 'var(--text-muted)' }}>
                {t('settings.proxyUrlHint')}
              </p>
            </div>
          )}

          <div
            className="rounded-xl p-4 space-y-3"
            style={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-base)' }}
          >
            <div>
              <p className="text-sm mb-1" style={{ color: 'var(--text-primary)' }}>{t('settings.proxyTestTitle')}</p>
              <p className="text-xs" style={{ color: 'var(--text-muted)' }}>{t('settings.proxyTestHint')}</p>
            </div>
            <div>
              <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.proxyTestUrl')}</p>
              <input
                value={proxyTestURL}
                onChange={e => setProxyTestURL(e.target.value)}
                placeholder={DEFAULT_PROXY_TEST_URL}
                className="input-base font-mono"
              />
              <p className="mt-1.5 text-xs" style={{ color: 'var(--text-muted)' }}>
                {t('settings.proxyTestUrlHint')}
              </p>
            </div>
            <div className="flex flex-wrap items-center gap-3">
              <button
                onClick={testProxyConnection}
                disabled={testingProxy || !cfg}
                className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-lg transition-colors disabled:opacity-50"
                style={{ background: 'var(--bg-overlay)', color: 'var(--text-primary)', border: '1px solid var(--border-base)' }}
              >
                <RefreshCw size={12} className={testingProxy ? 'animate-spin' : ''} />
                {testingProxy ? t('settings.proxyTesting') : t('settings.proxyTestAction')}
              </button>
              {proxyTestResult && (
                <span
                  className="text-xs"
                  style={proxyTestResult.tone === 'success'
                    ? { color: 'var(--success)' }
                    : { color: 'var(--danger)' }}
                >
                  {proxyTestResult.text}
                </span>
              )}
            </div>
          </div>
        </div>
      )}

      <button
        onClick={save}
        disabled={saving}
        className="btn-primary mt-8 px-6 py-2.5 rounded-lg text-sm"
      >
        {saving ? t('common.saving') : t('settings.saveSettings')}
      </button>
    </div>
  )
}
