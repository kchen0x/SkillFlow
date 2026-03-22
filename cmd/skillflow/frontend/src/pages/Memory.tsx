import { useEffect, useState } from 'react'
import { Brain, ExternalLink, Plus, RefreshCw, X, ChevronDown } from 'lucide-react'
import {
  GetMainMemory,
  SaveMainMemory,
  ListModuleMemories,
  CreateModuleMemory,
  SaveModuleMemory,
  DeleteModuleMemory,
  GetAllMemoryPushStatuses,
  GetAllMemoryPushConfigs,
  GetAllModulePushTargets,
  SaveMemoryPushConfig,
  SaveModulePushTargets,
  PushMemoryToAgent,
  PushAllMemory,
  OpenMemoryInEditor,
} from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { useLanguage } from '../contexts/LanguageContext'

type MainMemoryItem = { content: string; updatedAt: string }
type ModuleItem = { name: string; content: string; updatedAt: string }
type PushStatus = 'synced' | 'pendingPush' | 'neverPushed'
// keyed by agentType
type PushStatusMap = Record<string, PushStatus>
// keyed by agentType
type PushConfigMap = Record<string, { mode: string; autoPush: boolean }>
// keyed by moduleName
type ModuleTargetsMap = Record<string, string[]>
type DrawerState =
  | { type: 'none' }
  | { type: 'main' }
  | { type: 'module'; name: string }
type NewModuleState = { open: boolean; name: string; content: string; nameError: string }

const ALL_AGENTS = ['claude-code', 'codex', 'gemini-cli', 'opencode', 'openclaw']
const agentDisplayName: Record<string, string> = {
  'claude-code': 'Claude Code',
  'codex': 'Codex',
  'gemini-cli': 'Gemini CLI',
  'opencode': 'OpenCode',
  'openclaw': 'OpenClaw',
}

function statusColor(status: PushStatus): string {
  if (status === 'synced') return 'var(--color-success, #22c55e)'
  if (status === 'pendingPush') return 'var(--color-warning, #f97316)'
  return 'var(--text-muted)'
}

function getPreviewLines(content: string, maxLines: number): string {
  if (!content) return ''
  return content
    .split('\n')
    .filter(l => l.trim())
    .slice(0, maxLines)
    .join('\n')
}

export default function Memory() {
  const { t } = useLanguage()
  const [mainMemory, setMainMemory] = useState<MainMemoryItem | null>(null)
  const [modules, setModules] = useState<ModuleItem[]>([])
  const [pushStatuses, setPushStatuses] = useState<PushStatusMap>({})
  const [pushConfigs, setPushConfigs] = useState<PushConfigMap>({})
  const [moduleTargets, setModuleTargets] = useState<ModuleTargetsMap>({})
  const [drawerState, setDrawerState] = useState<DrawerState>({ type: 'none' })
  const [drawerContent, setDrawerContent] = useState('')
  const [drawerTab, setDrawerTab] = useState<'edit' | 'preview'>('edit')
  const [drawerPushMode, setDrawerPushMode] = useState('merge')
  const [drawerAutoPush, setDrawerAutoPush] = useState(false)
  // agent type used for PushMemoryToAgent when push now is clicked
  const [drawerPushAgent, setDrawerPushAgent] = useState(ALL_AGENTS[0])
  const [filterAgent, setFilterAgent] = useState('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [newModule, setNewModule] = useState<NewModuleState>({ open: false, name: '', content: '', nameError: '' })
  const [saving, setSaving] = useState(false)
  const [pushing, setPushing] = useState(false)
  const [pushMessage, setPushMessage] = useState('')
  const [deletingModule, setDeletingModule] = useState<string | null>(null)

  const loadAll = async () => {
    try {
      const [mm, mods, statuses, configs, targets] = await Promise.all([
        GetMainMemory(),
        ListModuleMemories(),
        GetAllMemoryPushStatuses(),
        GetAllMemoryPushConfigs(),
        GetAllModulePushTargets(),
      ])
      setMainMemory(mm as MainMemoryItem)
      setModules((mods ?? []) as ModuleItem[])

      // PushStatusDTO: { agentType: string, status: string }
      const statusMap: PushStatusMap = {}
      for (const s of (statuses ?? []) as main.PushStatusDTO[]) {
        statusMap[s.agentType] = s.status as PushStatus
      }
      setPushStatuses(statusMap)

      // MemoryPushConfigDTO: { agentType: string, mode: string, autoPush: boolean }
      const configMap: PushConfigMap = {}
      for (const c of (configs ?? []) as main.MemoryPushConfigDTO[]) {
        configMap[c.agentType] = { mode: c.mode, autoPush: c.autoPush }
      }
      setPushConfigs(configMap)

      // ModulePushTargetsDTO: { moduleName: string, pushTargets: string[] }
      const targetsMap: ModuleTargetsMap = {}
      for (const tgt of (targets ?? []) as main.ModulePushTargetsDTO[]) {
        targetsMap[tgt.moduleName] = tgt.pushTargets ?? []
      }
      setModuleTargets(targetsMap)
    } catch (e) {
      console.error('Failed to load memory data', e)
    }
  }

  useEffect(() => {
    loadAll()
  }, [])

  const openDrawer = (state: DrawerState) => {
    setDrawerState(state)
    setDrawerTab('edit')
    if (state.type === 'main') {
      setDrawerContent(mainMemory?.content ?? '')
      // Push configs are per-agent; use first configured agent or defaults
      const firstAgent = ALL_AGENTS[0]
      const cfg = pushConfigs[firstAgent] ?? { mode: 'merge', autoPush: false }
      setDrawerPushMode(cfg.mode)
      setDrawerAutoPush(cfg.autoPush)
      setDrawerPushAgent(firstAgent)
    } else if (state.type === 'module') {
      const mod = modules.find(m => m.name === state.name)
      setDrawerContent(mod?.content ?? '')
      // For module drawer, agent config — use first enabled target agent or first agent
      const targets = moduleTargets[state.name] ?? []
      const agent = targets[0] ?? ALL_AGENTS[0]
      const cfg = pushConfigs[agent] ?? { mode: 'merge', autoPush: false }
      setDrawerPushMode(cfg.mode)
      setDrawerAutoPush(cfg.autoPush)
      setDrawerPushAgent(agent)
    }
  }

  const closeDrawer = () => {
    setDrawerState({ type: 'none' })
  }

  const handleSaveDrawer = async () => {
    setSaving(true)
    try {
      if (drawerState.type === 'main') {
        const result = await SaveMainMemory(drawerContent)
        setMainMemory(result as MainMemoryItem)
      } else if (drawerState.type === 'module') {
        const name = drawerState.name
        const result = await SaveModuleMemory(name, drawerContent)
        setModules(prev => prev.map(m => m.name === name ? result as ModuleItem : m))
      }
      await loadAll()
    } catch (e) {
      console.error('Save failed', e)
    } finally {
      setSaving(false)
    }
  }

  const handleSavePushConfig = async (agentType: string, mode: string, autoPush: boolean) => {
    try {
      await SaveMemoryPushConfig(agentType, mode, autoPush)
      setPushConfigs(prev => ({ ...prev, [agentType]: { mode, autoPush } }))
    } catch (e) {
      console.error('SaveMemoryPushConfig failed', e)
    }
  }

  const handleToggleTarget = async (moduleName: string, agent: string) => {
    const current = moduleTargets[moduleName] ?? []
    const next = current.includes(agent) ? current.filter(a => a !== agent) : [...current, agent]
    setModuleTargets(prev => ({ ...prev, [moduleName]: next }))
    try {
      await SaveModulePushTargets(moduleName, next)
    } catch (e) {
      console.error('SaveModulePushTargets failed', e)
    }
  }

  const handlePushNow = async () => {
    if (drawerState.type === 'none') return
    setPushing(true)
    setPushMessage('')
    try {
      await PushMemoryToAgent(drawerPushAgent)
      setPushMessage(t('memory.pushSuccess'))
      await loadAll()
    } catch (e) {
      setPushMessage(t('memory.pushFailed'))
      console.error('PushMemoryToAgent failed', e)
    } finally {
      setPushing(false)
    }
  }

  const handlePushAll = async () => {
    setPushing(true)
    setPushMessage(t('memory.pushingAll'))
    try {
      await PushAllMemory()
      setPushMessage(t('memory.pushSuccess'))
      await loadAll()
    } catch (e) {
      setPushMessage(t('memory.pushFailed'))
      console.error('PushAllMemory failed', e)
    } finally {
      setPushing(false)
    }
  }

  const handleOpenInEditor = async () => {
    if (drawerState.type === 'none') return
    const memType = drawerState.type === 'main' ? 'main' : 'module'
    const memName = drawerState.type === 'module' ? drawerState.name : ''
    try {
      await OpenMemoryInEditor(memType, memName)
    } catch (e) {
      console.error('OpenMemoryInEditor failed', e)
    }
  }

  const handleCreateModule = async () => {
    const name = newModule.name.trim()
    if (!/^[a-z0-9][a-z0-9-]*$/.test(name)) {
      setNewModule(prev => ({ ...prev, nameError: 'Invalid name. Use a-z, 0-9, -' }))
      return
    }
    try {
      await CreateModuleMemory(name, newModule.content)
      setNewModule({ open: false, name: '', content: '', nameError: '' })
      await loadAll()
    } catch (e: unknown) {
      setNewModule(prev => ({ ...prev, nameError: String((e as Error)?.message ?? e) }))
    }
  }

  const handleDeleteModule = async (name: string) => {
    if (!window.confirm(t('memory.confirmDelete'))) return
    setDeletingModule(name)
    try {
      await DeleteModuleMemory(name)
      if (drawerState.type === 'module' && drawerState.name === name) {
        closeDrawer()
      }
      await loadAll()
    } catch (e) {
      console.error('DeleteModuleMemory failed', e)
    } finally {
      setDeletingModule(null)
    }
  }

  const filteredModules = modules.filter(m => {
    const targets = moduleTargets[m.name] ?? []
    const matchAgent = filterAgent === 'all' || targets.includes(filterAgent)
    const matchSearch = !searchQuery || m.name.toLowerCase().includes(searchQuery.toLowerCase())
    return matchAgent && matchSearch
  })

  const drawerModuleName = drawerState.type === 'module' ? drawerState.name : ''
  const drawerModuleTargets = drawerState.type === 'module' ? (moduleTargets[drawerModuleName] ?? []) : []

  return (
    <div className="flex h-full" style={{ background: 'var(--bg-base)' }}>
      {/* Left filter panel */}
      <aside
        className="flex flex-col gap-1 p-3"
        style={{
          width: 160,
          minWidth: 160,
          borderRight: '1px solid var(--border-base)',
          background: 'var(--bg-surface)',
          overflowY: 'auto',
        }}
      >
        <p className="text-xs px-2 mb-1 font-medium" style={{ color: 'var(--text-muted)' }}>
          Agent
        </p>
        <button
          onClick={() => setFilterAgent('all')}
          className="flex items-center gap-2 px-2 py-1.5 rounded-lg text-sm text-left"
          style={{
            background: filterAgent === 'all' ? 'var(--active-surface)' : 'transparent',
            color: filterAgent === 'all' ? 'var(--active-text)' : 'var(--text-muted)',
            border: filterAgent === 'all' ? '1px solid var(--active-border)' : '1px solid transparent',
          }}
        >
          <Brain size={14} />
          {t('memory.filterAll')}
        </button>
        {ALL_AGENTS.map(agent => (
          <button
            key={agent}
            onClick={() => setFilterAgent(agent)}
            className="flex items-center gap-2 px-2 py-1.5 rounded-lg text-sm text-left truncate"
            style={{
              background: filterAgent === agent ? 'var(--active-surface)' : 'transparent',
              color: filterAgent === agent ? 'var(--active-text)' : 'var(--text-muted)',
              border: filterAgent === agent ? '1px solid var(--active-border)' : '1px solid transparent',
            }}
          >
            <span className="truncate">{agentDisplayName[agent] ?? agent}</span>
          </button>
        ))}
      </aside>

      {/* Main area */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Toolbar */}
        <div
          className="flex items-center gap-3 px-6 py-3"
          style={{ borderBottom: '1px solid var(--border-base)' }}
        >
          <input
            type="text"
            placeholder={t('memory.searchPlaceholder')}
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            className="px-3 py-1.5 rounded-lg text-sm outline-none"
            style={{
              background: 'var(--bg-surface)',
              border: '1px solid var(--border-base)',
              color: 'var(--text-primary)',
              width: 240,
            }}
          />
          <div className="flex-1" />
          {pushMessage && (
            <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{pushMessage}</span>
          )}
          <button
            onClick={() => setNewModule(prev => ({ ...prev, open: true }))}
            className="btn-secondary flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm"
          >
            <Plus size={14} />
            {t('memory.newModule')}
          </button>
          <button
            onClick={handlePushAll}
            disabled={pushing}
            className="btn-primary flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm"
          >
            {pushing ? <RefreshCw size={14} className="animate-spin" /> : <ChevronDown size={14} />}
            {t('memory.pushAll')}
          </button>
        </div>

        {/* Card grid */}
        <div className="flex-1 overflow-auto p-6">
          {/* Main memory card */}
          <div
            className="mb-6 rounded-xl p-4 cursor-pointer"
            onClick={() => openDrawer({ type: 'main' })}
            style={{
              background: 'var(--bg-surface)',
              border: '1px solid var(--border-base)',
              borderLeft: '4px solid var(--accent-primary)',
            }}
            onMouseEnter={e => { e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.1)' }}
            onMouseLeave={e => { e.currentTarget.style.boxShadow = '' }}
          >
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <Brain size={16} style={{ color: 'var(--accent-primary)' }} />
                  <span className="font-semibold text-sm" style={{ color: 'var(--text-primary)' }}>
                    {t('memory.mainMemory')}
                  </span>
                </div>
                <p className="text-xs mb-2" style={{ color: 'var(--text-muted)' }}>
                  {t('memory.mainMemoryDesc')}
                </p>
                {mainMemory?.content && (
                  <pre
                    className="text-xs whitespace-pre-wrap break-all"
                    style={{
                      color: 'var(--text-secondary)',
                      fontFamily: 'inherit',
                      display: '-webkit-box',
                      WebkitLineClamp: 3,
                      WebkitBoxOrient: 'vertical',
                      overflow: 'hidden',
                    } as React.CSSProperties}
                  >
                    {getPreviewLines(mainMemory.content, 3)}
                  </pre>
                )}
              </div>
              <div className="flex flex-col gap-1 items-end shrink-0">
                {ALL_AGENTS.map(agent => {
                  const status: PushStatus = pushStatuses[agent] ?? 'neverPushed'
                  return (
                    <div key={agent} className="flex items-center gap-1" title={agentDisplayName[agent]}>
                      <span
                        className="inline-block rounded-full"
                        style={{ width: 6, height: 6, background: statusColor(status) }}
                      />
                      <span className="text-xs" style={{ color: 'var(--text-muted)' }}>
                        {agentDisplayName[agent] ?? agent}
                      </span>
                    </div>
                  )
                })}
              </div>
            </div>
          </div>

          {/* Module cards */}
          {filteredModules.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 gap-2">
              <Brain size={32} style={{ color: 'var(--text-muted)' }} />
              <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
                {searchQuery || filterAgent !== 'all'
                  ? 'No matching modules'
                  : t('memory.noModules')}
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-4">
              {filteredModules.map(mod => {
                const targets = moduleTargets[mod.name] ?? []
                const hasPending = targets.some(a => pushStatuses[a] === 'pendingPush')
                return (
                  <div
                    key={mod.name}
                    className="rounded-xl p-4 cursor-pointer relative"
                    onClick={() => openDrawer({ type: 'module', name: mod.name })}
                    style={{
                      background: 'var(--bg-surface)',
                      border: '1px solid var(--border-base)',
                    }}
                    onMouseEnter={e => { e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.1)' }}
                    onMouseLeave={e => { e.currentTarget.style.boxShadow = '' }}
                  >
                    {hasPending && (
                      <span
                        className="absolute top-3 right-3 inline-block rounded-full"
                        style={{ width: 8, height: 8, background: 'var(--color-warning, #f97316)' }}
                      />
                    )}
                    <p className="font-medium text-sm mb-1 truncate pr-4" style={{ color: 'var(--text-primary)' }}>
                      {mod.name}
                    </p>
                    {mod.content && (
                      <pre
                        className="text-xs whitespace-pre-wrap break-all mb-2"
                        style={{
                          color: 'var(--text-secondary)',
                          fontFamily: 'inherit',
                          display: '-webkit-box',
                          WebkitLineClamp: 2,
                          WebkitBoxOrient: 'vertical',
                          overflow: 'hidden',
                        } as React.CSSProperties}
                      >
                        {getPreviewLines(mod.content, 2)}
                      </pre>
                    )}
                    {targets.length > 0 && (
                      <div className="flex flex-wrap gap-1 mt-2">
                        {targets.map(a => (
                          <span
                            key={a}
                            className="text-[10px] px-1.5 py-0.5 rounded"
                            style={{
                              background: 'var(--bg-hover)',
                              color: 'var(--text-muted)',
                              border: '1px solid var(--border-base)',
                            }}
                          >
                            {agentDisplayName[a] ?? a}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>

      {/* Right drawer */}
      {drawerState.type !== 'none' && (
        <aside
          className="flex flex-col"
          style={{
            width: '55%',
            minWidth: 380,
            borderLeft: '1px solid var(--border-base)',
            background: 'var(--bg-surface)',
          }}
        >
          {/* Drawer header */}
          <div
            className="flex items-center justify-between px-4 py-3"
            style={{ borderBottom: '1px solid var(--border-base)' }}
          >
            <span className="font-semibold text-sm" style={{ color: 'var(--text-primary)' }}>
              {drawerState.type === 'main' ? t('memory.mainMemory') : drawerModuleName}
            </span>
            <div className="flex items-center gap-2">
              <button
                onClick={handleOpenInEditor}
                className="flex items-center gap-1 text-xs px-2 py-1 rounded-lg"
                style={{ color: 'var(--text-muted)', border: '1px solid var(--border-base)' }}
              >
                <ExternalLink size={12} />
                {t('memory.openInEditor')}
              </button>
              {drawerState.type === 'module' && (
                <button
                  onClick={() => handleDeleteModule(drawerModuleName)}
                  disabled={deletingModule === drawerModuleName}
                  className="flex items-center gap-1 text-xs px-2 py-1 rounded-lg"
                  style={{ color: 'var(--color-error, #ef4444)', border: '1px solid var(--border-base)' }}
                >
                  {t('memory.deleteModule')}
                </button>
              )}
              <button onClick={closeDrawer} style={{ color: 'var(--text-muted)' }}>
                <X size={16} />
              </button>
            </div>
          </div>

          {/* Tab bar */}
          <div className="flex" style={{ borderBottom: '1px solid var(--border-base)' }}>
            {(['edit', 'preview'] as const).map(tab => (
              <button
                key={tab}
                onClick={() => setDrawerTab(tab)}
                className="px-4 py-2 text-sm"
                style={{
                  color: drawerTab === tab ? 'var(--active-text)' : 'var(--text-muted)',
                  borderBottom: drawerTab === tab ? '2px solid var(--accent-primary)' : '2px solid transparent',
                }}
              >
                {tab === 'edit' ? t('memory.editTab') : t('memory.previewTab')}
              </button>
            ))}
          </div>

          {/* Content */}
          <div className="flex-1 overflow-auto">
            {drawerTab === 'edit' ? (
              <textarea
                value={drawerContent}
                onChange={e => setDrawerContent(e.target.value)}
                placeholder={t('memory.contentPlaceholder')}
                className="w-full h-full resize-none outline-none p-4 text-sm"
                style={{
                  background: 'transparent',
                  color: 'var(--text-primary)',
                  fontFamily: 'monospace',
                  minHeight: 200,
                }}
              />
            ) : (
              <pre
                className="p-4 text-sm whitespace-pre-wrap break-words"
                style={{ color: 'var(--text-primary)', fontFamily: 'inherit' }}
              >
                {drawerContent || (
                  <span style={{ color: 'var(--text-muted)' }}>{t('memory.contentPlaceholder')}</span>
                )}
              </pre>
            )}
          </div>

          {/* Footer */}
          <div
            className="p-4 flex flex-col gap-3"
            style={{ borderTop: '1px solid var(--border-base)' }}
          >
            {/* Module push targets */}
            {drawerState.type === 'module' && (
              <div>
                <p className="text-xs mb-1.5 font-medium" style={{ color: 'var(--text-muted)' }}>
                  {t('memory.pushTargets')}
                </p>
                <div className="flex flex-wrap gap-1.5">
                  {ALL_AGENTS.map(agent => {
                    const selected = drawerModuleTargets.includes(agent)
                    return (
                      <button
                        key={agent}
                        onClick={() => handleToggleTarget(drawerModuleName, agent)}
                        className="text-xs px-2 py-1 rounded-lg"
                        style={{
                          background: selected ? 'var(--active-surface)' : 'var(--bg-hover)',
                          color: selected ? 'var(--active-text)' : 'var(--text-muted)',
                          border: selected ? '1px solid var(--active-border)' : '1px solid var(--border-base)',
                        }}
                      >
                        {agentDisplayName[agent] ?? agent}
                      </button>
                    )
                  })}
                </div>
              </div>
            )}

            {/* Agent selector for push (main memory drawer) */}
            {drawerState.type === 'main' && (
              <div className="flex items-center gap-2">
                <span className="text-xs" style={{ color: 'var(--text-muted)' }}>Push to:</span>
                <div className="flex flex-wrap gap-1.5">
                  {ALL_AGENTS.map(agent => (
                    <button
                      key={agent}
                      onClick={() => {
                        setDrawerPushAgent(agent)
                        const cfg = pushConfigs[agent] ?? { mode: 'merge', autoPush: false }
                        setDrawerPushMode(cfg.mode)
                        setDrawerAutoPush(cfg.autoPush)
                      }}
                      className="text-xs px-2 py-1 rounded-lg"
                      style={{
                        background: drawerPushAgent === agent ? 'var(--active-surface)' : 'var(--bg-hover)',
                        color: drawerPushAgent === agent ? 'var(--active-text)' : 'var(--text-muted)',
                        border: drawerPushAgent === agent ? '1px solid var(--active-border)' : '1px solid var(--border-base)',
                      }}
                    >
                      {agentDisplayName[agent] ?? agent}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Push mode + auto push */}
            <div className="flex items-center gap-3 flex-wrap">
              <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{t('memory.pushMode')}:</span>
              {['merge', 'takeover'].map(mode => (
                <button
                  key={mode}
                  onClick={() => {
                    setDrawerPushMode(mode)
                    handleSavePushConfig(drawerPushAgent, mode, drawerAutoPush)
                  }}
                  className="text-xs px-2 py-1 rounded"
                  style={{
                    background: drawerPushMode === mode ? 'var(--active-surface)' : 'var(--bg-hover)',
                    color: drawerPushMode === mode ? 'var(--active-text)' : 'var(--text-muted)',
                    border: drawerPushMode === mode ? '1px solid var(--active-border)' : '1px solid var(--border-base)',
                  }}
                >
                  {mode === 'merge' ? t('memory.mergeModeLabel') : t('memory.takeoverModeLabel')}
                </button>
              ))}
              <span className="text-xs ml-1" style={{ color: 'var(--text-muted)' }}>{t('memory.autoPush')}:</span>
              <button
                onClick={() => {
                  const next = !drawerAutoPush
                  setDrawerAutoPush(next)
                  handleSavePushConfig(drawerPushAgent, drawerPushMode, next)
                }}
                className="text-xs px-2 py-1 rounded"
                style={{
                  background: drawerAutoPush ? 'var(--active-surface)' : 'var(--bg-hover)',
                  color: drawerAutoPush ? 'var(--active-text)' : 'var(--text-muted)',
                  border: drawerAutoPush ? '1px solid var(--active-border)' : '1px solid var(--border-base)',
                }}
              >
                {drawerAutoPush ? 'ON' : 'OFF'}
              </button>
            </div>

            {/* Action buttons */}
            <div className="flex gap-2">
              <button
                onClick={handleSaveDrawer}
                disabled={saving}
                className="btn-secondary flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm flex-1 justify-center"
              >
                {saving && <RefreshCw size={13} className="animate-spin" />}
                {saving ? t('common.saving') : t('common.save')}
              </button>
              <button
                onClick={handlePushNow}
                disabled={pushing}
                className="btn-primary flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm flex-1 justify-center"
              >
                {pushing && <RefreshCw size={13} className="animate-spin" />}
                {pushing ? t('memory.pushingAll') : t('memory.pushNow')}
              </button>
            </div>
          </div>
        </aside>
      )}

      {/* New module dialog */}
      {newModule.open && (
        <div
          className="fixed inset-0 flex items-center justify-center z-50"
          style={{ background: 'rgba(0,0,0,0.4)' }}
          onClick={e => { if (e.target === e.currentTarget) setNewModule(prev => ({ ...prev, open: false })) }}
        >
          <div
            className="rounded-2xl p-6 flex flex-col gap-4"
            style={{
              background: 'var(--bg-surface)',
              border: '1px solid var(--border-base)',
              width: 420,
              boxShadow: '0 8px 32px rgba(0,0,0,0.2)',
            }}
          >
            <div className="flex items-center justify-between">
              <span className="font-semibold text-base" style={{ color: 'var(--text-primary)' }}>
                {t('memory.newModuleTitle')}
              </span>
              <button onClick={() => setNewModule(prev => ({ ...prev, open: false }))} style={{ color: 'var(--text-muted)' }}>
                <X size={16} />
              </button>
            </div>
            <div>
              <label className="text-xs block mb-1" style={{ color: 'var(--text-muted)' }}>Name</label>
              <input
                type="text"
                value={newModule.name}
                onChange={e => setNewModule(prev => ({ ...prev, name: e.target.value, nameError: '' }))}
                placeholder={t('memory.moduleNamePlaceholder')}
                className="w-full px-3 py-2 rounded-lg text-sm outline-none"
                style={{
                  background: 'var(--bg-hover)',
                  border: '1px solid var(--border-base)',
                  color: 'var(--text-primary)',
                }}
              />
              {newModule.nameError && (
                <p className="text-xs mt-1" style={{ color: 'var(--color-error, #ef4444)' }}>{newModule.nameError}</p>
              )}
            </div>
            <div>
              <label className="text-xs block mb-1" style={{ color: 'var(--text-muted)' }}>Content</label>
              <textarea
                value={newModule.content}
                onChange={e => setNewModule(prev => ({ ...prev, content: e.target.value }))}
                placeholder={t('memory.contentPlaceholder')}
                rows={5}
                className="w-full px-3 py-2 rounded-lg text-sm outline-none resize-none"
                style={{
                  background: 'var(--bg-hover)',
                  border: '1px solid var(--border-base)',
                  color: 'var(--text-primary)',
                  fontFamily: 'monospace',
                }}
              />
            </div>
            <div className="flex gap-2 justify-end">
              <button
                onClick={() => setNewModule({ open: false, name: '', content: '', nameError: '' })}
                className="btn-secondary px-4 py-2 rounded-lg text-sm"
              >
                {t('common.cancel')}
              </button>
              <button
                onClick={handleCreateModule}
                className="btn-primary px-4 py-2 rounded-lg text-sm"
              >
                {t('memory.createModule')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
