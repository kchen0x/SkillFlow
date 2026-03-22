import { useEffect, useState } from 'react'
import { Brain, CheckSquare, ExternalLink, Plus, RefreshCw, Upload, X } from 'lucide-react'
import {
  CreateModuleMemory,
  DeleteModuleMemory,
  GetAllMemoryPushConfigs,
  GetAllMemoryPushStatuses,
  GetEnabledAgents,
  GetMainMemory,
  ListModuleMemories,
  OpenMemoryInEditor,
  PushSelectedMemory,
  SaveMainMemory,
  SaveMemoryPushConfig,
  SaveModuleMemory,
} from '../../wailsjs/go/main/App'
import { domain, main } from '../../wailsjs/go/models'
import { useLanguage } from '../contexts/LanguageContext'
import { renderMemoryMarkdown } from '../lib/memoryMarkdown'
import {
  createMemoryBatchPushState,
  getMemoryAutoSyncMode,
  getMemoryPushConfigForAutoSyncMode,
  isMemoryBatchPushReady,
  toggleMemoryBatchAgent,
  toggleMemoryBatchModule,
} from '../lib/memoryPageState'
import {
  buildMemoryPushStatusEntries,
  getMemoryAgentLabel,
  getMemoryDrawerMetrics,
  type MemoryPushStatus,
} from '../lib/memoryUi'

type MainMemoryItem = { content: string; updatedAt: string }
type ModuleItem = { name: string; content: string; updatedAt: string }
type PushStatus = MemoryPushStatus
type PushStatusMap = Record<string, PushStatus>
type PushConfigMap = Record<string, { mode: string; autoPush: boolean }>
type DrawerState =
  | { type: 'none' }
  | { type: 'main' }
  | { type: 'module'; name: string }
type NewModuleState = { open: boolean; name: string; content: string; nameError: string }

function statusColor(status: PushStatus): string {
  if (status === 'synced') return 'var(--color-success, #22c55e)'
  if (status === 'pendingPush') return 'var(--color-warning, #f97316)'
  return 'var(--text-muted)'
}

function getPreviewLines(content: string, maxLines: number): string {
  if (!content) return ''
  return content
    .split('\n')
    .filter(line => line.trim())
    .slice(0, maxLines)
    .join('\n')
}

export default function Memory() {
  const { t } = useLanguage()
  const [mainMemory, setMainMemory] = useState<MainMemoryItem | null>(null)
  const [modules, setModules] = useState<ModuleItem[]>([])
  const [agents, setAgents] = useState<domain.AgentProfile[]>([])
  const [pushStatuses, setPushStatuses] = useState<PushStatusMap>({})
  const [pushConfigs, setPushConfigs] = useState<PushConfigMap>({})
  const [drawerState, setDrawerState] = useState<DrawerState>({ type: 'none' })
  const [drawerContent, setDrawerContent] = useState('')
  const [drawerInitialContent, setDrawerInitialContent] = useState('')
  const [drawerTab, setDrawerTab] = useState<'edit' | 'preview'>('edit')
  const [batchPushMode, setBatchPushMode] = useState(false)
  const [batchPushState, setBatchPushState] = useState(createMemoryBatchPushState())
  const [searchQuery, setSearchQuery] = useState('')
  const [newModule, setNewModule] = useState<NewModuleState>({ open: false, name: '', content: '', nameError: '' })
  const [saving, setSaving] = useState(false)
  const [pushing, setPushing] = useState(false)
  const [pushMessage, setPushMessage] = useState('')
  const [deletingModule, setDeletingModule] = useState<string | null>(null)
  const [closeConfirmOpen, setCloseConfirmOpen] = useState(false)
  const [enterBatchPushAfterClose, setEnterBatchPushAfterClose] = useState(false)

  const availableAgents = agents ?? []
  const drawerModuleName = drawerState.type === 'module' ? drawerState.name : ''
  const drawerDirty = drawerState.type !== 'none' && drawerContent !== drawerInitialContent
  const previewHtml = drawerContent ? renderMemoryMarkdown(drawerContent) : ''
  const batchPushReady = isMemoryBatchPushReady(batchPushState)
  const selectedMemoryCount = 1 + batchPushState.selectedModules.length
  const memoryStatusEntries = buildMemoryPushStatusEntries(availableAgents, pushStatuses)
  const drawerMetrics = getMemoryDrawerMetrics(window.innerWidth)

  const loadAll = async () => {
    try {
      const [mm, mods, statuses, configs, enabledAgents] = await Promise.all([
        GetMainMemory(),
        ListModuleMemories(),
        GetAllMemoryPushStatuses(),
        GetAllMemoryPushConfigs(),
        GetEnabledAgents(),
      ])

      setMainMemory(mm as MainMemoryItem)
      setModules((mods ?? []) as ModuleItem[])
      setAgents((enabledAgents ?? []) as domain.AgentProfile[])

      const statusMap: PushStatusMap = {}
      for (const status of (statuses ?? []) as main.PushStatusDTO[]) {
        statusMap[status.agentType] = status.status as PushStatus
      }
      setPushStatuses(statusMap)

      const configMap: PushConfigMap = {}
      for (const config of (configs ?? []) as main.MemoryPushConfigDTO[]) {
        configMap[config.agentType] = { mode: config.mode, autoPush: config.autoPush }
      }
      setPushConfigs(configMap)
    } catch (error) {
      console.error('Failed to load memory data', error)
    }
  }

  useEffect(() => {
    void loadAll()
  }, [])

  const filteredModules = modules.filter(module => {
    const query = searchQuery.trim().toLowerCase()
    if (!query) {
      return true
    }
    return module.name.toLowerCase().includes(query) || module.content.toLowerCase().includes(query)
  })

  const closeDrawerImmediate = () => {
    setDrawerState({ type: 'none' })
    setDrawerContent('')
    setDrawerInitialContent('')
    setDrawerTab('edit')
    setCloseConfirmOpen(false)
    setEnterBatchPushAfterClose(false)
  }

  const enterBatchPushSelection = () => {
    closeDrawerImmediate()
    setBatchPushMode(true)
    setBatchPushState(createMemoryBatchPushState())
    setPushMessage('')
  }

  const requestCloseDrawer = (startBatchPushAfterClose = false) => {
    if (!drawerDirty) {
      closeDrawerImmediate()
      if (startBatchPushAfterClose) {
        enterBatchPushSelection()
      }
      return
    }
    setEnterBatchPushAfterClose(startBatchPushAfterClose)
    setCloseConfirmOpen(true)
  }

  const openDrawer = (state: DrawerState) => {
    if (batchPushMode) {
      return
    }

    let nextContent = ''
    if (state.type === 'main') {
      nextContent = mainMemory?.content ?? ''
    }
    if (state.type === 'module') {
      nextContent = modules.find(module => module.name === state.name)?.content ?? ''
    }

    setDrawerState(state)
    setDrawerContent(nextContent)
    setDrawerInitialContent(nextContent)
    setDrawerTab('edit')
    setCloseConfirmOpen(false)
    setEnterBatchPushAfterClose(false)
  }

  const handleSaveDrawer = async (closeAfterSave = false): Promise<boolean> => {
    setSaving(true)
    try {
      if (drawerState.type === 'main') {
        const result = await SaveMainMemory(drawerContent)
        setMainMemory(result as MainMemoryItem)
      } else if (drawerState.type === 'module') {
        const name = drawerState.name
        const result = await SaveModuleMemory(name, drawerContent)
        setModules(prev => prev.map(module => (module.name === name ? result as ModuleItem : module)))
      } else {
        return true
      }

      setDrawerInitialContent(drawerContent)
      await loadAll()
      if (closeAfterSave) {
        closeDrawerImmediate()
      }
      return true
    } catch (error) {
      console.error('Save failed', error)
      return false
    } finally {
      setSaving(false)
    }
  }

  const handleConfirmSaveAndClose = async () => {
    const shouldEnterBatchPush = enterBatchPushAfterClose
    setCloseConfirmOpen(false)
    setEnterBatchPushAfterClose(false)
    const saved = await handleSaveDrawer(true)
    if (saved && shouldEnterBatchPush) {
      enterBatchPushSelection()
    }
  }

  const handleDiscardAndClose = () => {
    const shouldEnterBatchPush = enterBatchPushAfterClose
    closeDrawerImmediate()
    if (shouldEnterBatchPush) {
      enterBatchPushSelection()
    }
  }

  const handleAutoSyncModeChange = async (agentType: string, nextMode: 'off' | 'merge' | 'takeover') => {
    const nextConfig = getMemoryPushConfigForAutoSyncMode(nextMode)
    try {
      await SaveMemoryPushConfig(agentType, nextConfig.mode, nextConfig.autoPush)
      setPushConfigs(prev => ({ ...prev, [agentType]: nextConfig }))
      await loadAll()
    } catch (error) {
      console.error('SaveMemoryPushConfig failed', error)
    }
  }

  const handleStartBatchPush = async () => {
    if (!batchPushReady) {
      return
    }

    setPushing(true)
    setPushMessage(t('memory.pushingBatch'))
    try {
      const results = await PushSelectedMemory(batchPushState.selectedAgents, batchPushState.selectedModules, batchPushState.mode)
      const hasFailures = ((results ?? []) as main.PushResultDTO[]).some(result => !result.success)
      await loadAll()
      if (hasFailures) {
        setPushMessage(t('memory.pushPartialFailed'))
        return
      }
      setPushMessage(t('memory.pushSuccess'))
      setBatchPushMode(false)
      setBatchPushState(createMemoryBatchPushState())
    } catch (error) {
      setPushMessage(t('memory.pushFailed'))
      console.error('PushSelectedMemory failed', error)
    } finally {
      setPushing(false)
    }
  }

  const handleOpenInEditor = async () => {
    if (drawerState.type === 'none') return
    const memoryType = drawerState.type === 'main' ? 'main' : 'module'
    const memoryName = drawerState.type === 'module' ? drawerState.name : ''
    try {
      await OpenMemoryInEditor(memoryType, memoryName)
    } catch (error) {
      console.error('OpenMemoryInEditor failed', error)
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
    } catch (error: unknown) {
      setNewModule(prev => ({ ...prev, nameError: String((error as Error)?.message ?? error) }))
    }
  }

  const handleDeleteModule = async (name: string) => {
    if (!window.confirm(t('memory.confirmDelete'))) return
    setDeletingModule(name)
    try {
      await DeleteModuleMemory(name)
      if (drawerState.type === 'module' && drawerState.name === name) {
        closeDrawerImmediate()
      }
      await loadAll()
    } catch (error) {
      console.error('DeleteModuleMemory failed', error)
    } finally {
      setDeletingModule(null)
    }
  }

  return (
    <div className="flex h-full flex-col" style={{ background: 'var(--bg-base)' }}>
      <div
        className="flex items-center gap-3 px-6 py-3"
        style={{ borderBottom: '1px solid var(--border-base)' }}
      >
        <input
          type="text"
          placeholder={t('memory.searchPlaceholder')}
          value={searchQuery}
          onChange={event => setSearchQuery(event.target.value)}
          className="px-3 py-1.5 rounded-lg text-sm outline-none"
          style={{
            background: 'var(--bg-surface)',
            border: '1px solid var(--border-base)',
            color: 'var(--text-primary)',
            width: 280,
          }}
        />
        <div className="flex-1" />
        {pushMessage && (
          <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{pushMessage}</span>
        )}
        {!batchPushMode && (
          <>
            <button
              onClick={() => setNewModule(prev => ({ ...prev, open: true }))}
              className="btn-secondary flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm"
            >
              <Plus size={14} />
              {t('memory.newModule')}
            </button>
            <button
              onClick={() => {
                if (drawerState.type !== 'none') {
                  requestCloseDrawer(true)
                  return
                }
                enterBatchPushSelection()
              }}
              className="btn-primary flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm"
            >
              <Upload size={14} />
              {t('memory.batchPush')}
            </button>
          </>
        )}
        {batchPushMode && (
          <>
            <button
              onClick={() => {
                setBatchPushMode(false)
                setBatchPushState(createMemoryBatchPushState())
              }}
              className="btn-secondary px-3 py-1.5 rounded-lg text-sm"
            >
              {t('common.cancel')}
            </button>
            <button
              onClick={handleStartBatchPush}
              disabled={!batchPushReady || pushing}
              className="btn-primary flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm"
            >
              {pushing && <RefreshCw size={14} className="animate-spin" />}
              {t('memory.startBatchPush')}
            </button>
          </>
        )}
      </div>

      <div className="px-6 pt-4">
        <div
          className="rounded-2xl p-4"
          style={{
            background: 'var(--bg-surface)',
            border: '1px solid var(--border-base)',
            boxShadow: 'var(--shadow-card)',
          }}
        >
          {!batchPushMode ? (
            <div className="flex flex-col gap-3">
              <div className="flex items-center justify-between gap-3">
                <div>
                  <p className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>
                    {t('memory.autoSyncPanelTitle')}
                  </p>
                  <p className="text-xs mt-1" style={{ color: 'var(--text-muted)' }}>
                    {t('memory.autoSyncPanelHint')}
                  </p>
                </div>
              </div>
              {availableAgents.length === 0 ? (
                <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
                  {t('memory.noAgentsConfigured')}
                </p>
              ) : (
                <div className="flex flex-col gap-2.5">
                  {availableAgents.map(agent => {
                    const currentMode = getMemoryAutoSyncMode(pushConfigs[agent.name])
                    const status = pushStatuses[agent.name] ?? 'neverPushed'
                    return (
                      <div
                        key={agent.name}
                        className="flex flex-wrap items-center justify-between gap-3 rounded-xl px-3 py-2"
                        style={{ background: 'var(--bg-hover)' }}
                      >
                        <div className="flex items-center gap-2">
                          <span
                            className="inline-block rounded-full"
                            style={{ width: 8, height: 8, background: statusColor(status) }}
                          />
                          <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                            {getMemoryAgentLabel(agent.name)}
                          </span>
                        </div>
                        <div className="flex flex-wrap gap-1.5">
                          {(['off', 'merge', 'takeover'] as const).map(mode => (
                            <button
                              key={mode}
                              onClick={() => void handleAutoSyncModeChange(agent.name, mode)}
                            className="text-xs px-2.5 py-1 rounded-lg"
                            style={{
                                background: currentMode === mode ? 'var(--active-surface)' : 'var(--bg-surface)',
                                color: currentMode === mode ? 'var(--active-text)' : 'var(--text-muted)',
                                border: currentMode === mode ? '1px solid var(--active-border)' : '1px solid var(--border-base)',
                              }}
                            >
                              {mode === 'off'
                                ? t('memory.autoSyncOff')
                                : mode === 'merge'
                                  ? t('memory.autoSyncMerge')
                                  : t('memory.autoSyncTakeover')}
                            </button>
                          ))}
                        </div>
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          ) : (
            <div className="flex flex-col gap-3">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>
                    {t('memory.batchPushTargets')}
                  </p>
                  <p className="text-xs mt-1" style={{ color: 'var(--text-muted)' }}>
                    {t('memory.batchPushHint', { count: selectedMemoryCount })}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{t('memory.pushMode')}:</span>
                  {(['merge', 'takeover'] as const).map(mode => (
                    <button
                      key={mode}
                      onClick={() => setBatchPushState(prev => ({ ...prev, mode }))}
                      className="text-xs px-2.5 py-1 rounded-lg"
                      style={{
                        background: batchPushState.mode === mode ? 'var(--active-surface)' : 'var(--bg-surface)',
                        color: batchPushState.mode === mode ? 'var(--active-text)' : 'var(--text-muted)',
                        border: batchPushState.mode === mode ? '1px solid var(--active-border)' : '1px solid var(--border-base)',
                      }}
                    >
                      {mode === 'merge' ? t('memory.mergeModeLabel') : t('memory.takeoverModeLabel')}
                    </button>
                  ))}
                </div>
              </div>
              {availableAgents.length === 0 ? (
                <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
                  {t('memory.noAgentsConfigured')}
                </p>
              ) : (
                <div className="flex flex-wrap gap-2">
                  {availableAgents.map(agent => {
                    const selected = batchPushState.selectedAgents.includes(agent.name)
                    return (
                      <button
                        key={agent.name}
                        onClick={() => setBatchPushState(prev => ({
                          ...prev,
                          selectedAgents: toggleMemoryBatchAgent(prev.selectedAgents, agent.name),
                        }))}
                        className="text-xs px-2.5 py-1.5 rounded-lg"
                        style={{
                          background: selected ? 'var(--active-surface)' : 'var(--bg-hover)',
                          color: selected ? 'var(--active-text)' : 'var(--text-muted)',
                          border: selected ? '1px solid var(--active-border)' : '1px solid var(--border-base)',
                        }}
                      >
                        {getMemoryAgentLabel(agent.name)}
                      </button>
                    )
                  })}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      <div className="flex-1 overflow-auto p-6">
        <div
          className="mb-6 rounded-xl p-4 relative"
          onClick={() => openDrawer({ type: 'main' })}
          style={{
            background: batchPushMode ? 'var(--active-surface)' : 'var(--bg-surface)',
            border: batchPushMode ? '1px solid var(--active-border)' : '1px solid var(--border-base)',
            borderLeft: '4px solid var(--accent-primary)',
            cursor: batchPushMode ? 'default' : 'pointer',
          }}
        >
          {batchPushMode && (
            <div className="absolute top-3 right-3 flex items-center gap-1.5">
              <CheckSquare size={16} style={{ color: 'var(--accent-primary)' }} />
              <span className="text-[10px]" style={{ color: 'var(--text-muted)' }}>
                {t('memory.required')}
              </span>
            </div>
          )}
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
              {memoryStatusEntries.map(entry => {
                const status = entry.status
                return (
                  <div key={entry.agentType} className="flex items-center gap-1" title={entry.label}>
                    <span
                      className="inline-block rounded-full"
                      style={{ width: 6, height: 6, background: statusColor(status) }}
                    />
                    <span className="text-xs" style={{ color: 'var(--text-muted)' }}>
                      {entry.label}
                    </span>
                  </div>
                )
              })}
            </div>
          </div>
        </div>

        {filteredModules.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 gap-2">
            <Brain size={32} style={{ color: 'var(--text-muted)' }} />
            <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
              {searchQuery ? t('memory.noSearchResults') : t('memory.noModules')}
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
            {filteredModules.map(module => {
              const selected = batchPushState.selectedModules.includes(module.name)
              return (
                <div
                  key={module.name}
                  className="rounded-xl p-4 relative"
                  onClick={() => {
                    if (batchPushMode) {
                      setBatchPushState(prev => ({
                        ...prev,
                        selectedModules: toggleMemoryBatchModule(prev.selectedModules, module.name),
                      }))
                      return
                    }
                    openDrawer({ type: 'module', name: module.name })
                  }}
                  style={{
                    background: selected ? 'var(--active-surface)' : 'var(--bg-surface)',
                    border: selected ? '1px solid var(--active-border)' : '1px solid var(--border-base)',
                    cursor: batchPushMode ? 'pointer' : 'pointer',
                  }}
                  onMouseEnter={event => { event.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.1)' }}
                  onMouseLeave={event => { event.currentTarget.style.boxShadow = '' }}
                >
                  {batchPushMode && (
                    <label
                      className="absolute top-3 right-3"
                      onClick={event => {
                        event.stopPropagation()
                        setBatchPushState(prev => ({
                          ...prev,
                          selectedModules: toggleMemoryBatchModule(prev.selectedModules, module.name),
                        }))
                      }}
                    >
                      <input type="checkbox" checked={selected} readOnly />
                    </label>
                  )}
                  <p className="font-medium text-sm mb-1 truncate pr-8" style={{ color: 'var(--text-primary)' }}>
                    {module.name}
                  </p>
                  {module.content && (
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
                      {getPreviewLines(module.content, 2)}
                    </pre>
                  )}
                  <div className="mb-2 flex flex-wrap gap-x-3 gap-y-1">
                    {memoryStatusEntries.map(entry => {
                      const status = entry.status
                      return (
                        <div key={entry.agentType} className="flex items-center gap-1" title={entry.label}>
                          <span
                            className="inline-block rounded-full"
                            style={{ width: 6, height: 6, background: statusColor(status) }}
                          />
                          <span className="text-[11px]" style={{ color: 'var(--text-muted)' }}>
                            {entry.label}
                          </span>
                        </div>
                      )
                    })}
                  </div>
                  <p className="text-[11px]" style={{ color: 'var(--text-muted)' }}>
                    {t('memory.moduleRefHint')}
                  </p>
                </div>
              )
            })}
          </div>
        )}
      </div>

      {drawerState.type !== 'none' && (
        <button
          type="button"
          aria-label={t('common.close')}
          onClick={() => requestCloseDrawer(false)}
          className="fixed inset-0"
          style={{
            background: 'rgba(15, 23, 42, 0.18)',
            border: 'none',
            zIndex: 30,
          }}
        />
      )}

      {drawerState.type !== 'none' && (
        <aside
          className="flex flex-col"
          style={{
            position: 'fixed',
            top: 0,
            right: 0,
            bottom: 0,
            width: drawerMetrics.width,
            maxWidth: drawerMetrics.maxWidth,
            minWidth: drawerMetrics.minWidth,
            borderLeft: '1px solid var(--border-base)',
            background: 'var(--bg-surface)',
            boxShadow: '-16px 0 36px rgba(15, 23, 42, 0.14)',
            zIndex: 40,
          }}
        >
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
                  onClick={() => void handleDeleteModule(drawerModuleName)}
                  disabled={deletingModule === drawerModuleName}
                  className="flex items-center gap-1 text-xs px-2 py-1 rounded-lg"
                  style={{ color: 'var(--color-error, #ef4444)', border: '1px solid var(--border-base)' }}
                >
                  {t('memory.deleteModule')}
                </button>
              )}
              <button onClick={() => requestCloseDrawer(false)} style={{ color: 'var(--text-muted)' }}>
                <X size={16} />
              </button>
            </div>
          </div>

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

          <div className="flex-1 overflow-auto">
            {drawerTab === 'edit' ? (
              <textarea
                value={drawerContent}
                onChange={event => setDrawerContent(event.target.value)}
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
              <div className="p-4">
                {previewHtml ? (
                  <div
                    className="memory-markdown-preview"
                    dangerouslySetInnerHTML={{ __html: previewHtml }}
                  />
                ) : (
                  <span className="text-sm" style={{ color: 'var(--text-muted)' }}>
                    {t('memory.contentPlaceholder')}
                  </span>
                )}
              </div>
            )}
          </div>

          <div
            className="p-4"
            style={{ borderTop: '1px solid var(--border-base)' }}
          >
            <button
              onClick={() => void handleSaveDrawer(false)}
              disabled={saving}
              className="btn-secondary flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm w-full justify-center"
            >
              {saving && <RefreshCw size={13} className="animate-spin" />}
              {saving ? t('common.saving') : t('common.save')}
            </button>
          </div>
        </aside>
      )}

      {closeConfirmOpen && (
        <div
          className="fixed inset-0 flex items-center justify-center z-50"
          style={{ background: 'rgba(0,0,0,0.4)' }}
          onClick={event => {
            if (event.target === event.currentTarget) {
              setCloseConfirmOpen(false)
              setEnterBatchPushAfterClose(false)
            }
          }}
        >
          <div
            className="rounded-2xl p-6 flex flex-col gap-4"
            style={{
              background: 'var(--bg-surface)',
              border: '1px solid var(--border-base)',
              width: 440,
              boxShadow: '0 8px 32px rgba(0,0,0,0.2)',
            }}
          >
            <div>
              <p className="font-semibold text-base" style={{ color: 'var(--text-primary)' }}>
                {t('memory.unsavedChangesTitle')}
              </p>
              <p className="text-sm mt-1" style={{ color: 'var(--text-muted)' }}>
                {t('memory.unsavedChangesHint')}
              </p>
            </div>
            <div className="flex gap-2 justify-end">
              <button
                onClick={() => {
                  setCloseConfirmOpen(false)
                  setEnterBatchPushAfterClose(false)
                }}
                className="btn-secondary px-4 py-2 rounded-lg text-sm"
              >
                {t('memory.keepEditing')}
              </button>
              <button
                onClick={handleDiscardAndClose}
                className="btn-secondary px-4 py-2 rounded-lg text-sm"
                style={{ color: 'var(--color-error)' }}
              >
                {t('memory.discardChanges')}
              </button>
              <button
                onClick={() => void handleConfirmSaveAndClose()}
                className="btn-primary px-4 py-2 rounded-lg text-sm"
              >
                {t('memory.saveAndClose')}
              </button>
            </div>
          </div>
        </div>
      )}

      {newModule.open && (
        <div
          className="fixed inset-0 flex items-center justify-center z-50"
          style={{ background: 'rgba(0,0,0,0.4)' }}
          onClick={event => { if (event.target === event.currentTarget) setNewModule(prev => ({ ...prev, open: false })) }}
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
                onChange={event => setNewModule(prev => ({ ...prev, name: event.target.value, nameError: '' }))}
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
                onChange={event => setNewModule(prev => ({ ...prev, content: event.target.value }))}
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
