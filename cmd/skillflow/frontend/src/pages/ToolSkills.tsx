import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  GetEnabledAgents,
  GetAgentMemoryPreview,
  ListAgentSkills,
  DeleteAgentSkill,
  OpenPath,
  ReadSkillFileContent,
  GetSkillMetaByPath,
  ListCategories,
  ScanAgentSkills,
  PullFromAgent,
  PullFromAgentForce,
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import ConflictDialog from '../components/ConflictDialog'
import { ToolIcon } from '../config/toolIcons'
import SkillTooltip from '../components/SkillTooltip'
import SkillStatusStrip from '../components/SkillStatusStrip'
import SkillListControls from '../components/SkillListControls'
import SyncSkillCard from '../components/SyncSkillCard'
import { buildAgentMemoryEntries } from '../lib/agentMemoryPreview'
import { copyTextToClipboard } from '../lib/clipboard'
import { createToolSkillsEventSubscriptions } from '../lib/dashboardSkillSettings'
import { SkillSortOrder, filterAndSortSkills } from '../lib/skillList'
import { ToolSkillsPanel, filterToolSkillsPanelContent, getDefaultToolSkillsPanel, getVisibleToolSkillsResultCount } from '../lib/toolSkillsPanels'
import {
  createToolSkillsPullState,
  toggleToolSkillsPullPath,
  toggleToolSkillsPullAllVisible,
  toggleToolSkillsPullVisibleNotImported,
  syncToolSkillsPullVisibleSelection,
  getToolSkillsVisibleNotImportedPaths,
  isToolSkillsPullReady,
} from '../lib/toolSkillsPullState'
import { subscribeToEvents } from '../lib/wailsEvents'
import { Wrench, Trash2, FolderOpenDot, Copy, Check, CheckSquare, ArrowUpToLine, ScanLine, Brain, RefreshCw, ArrowDownToLine, AlertCircle } from 'lucide-react'
import { useLanguage } from '../contexts/LanguageContext'
import { useSkillStatusVisibility } from '../contexts/SkillStatusVisibilityContext'

type AgentMemoryRuleItem = {
  name: string
  path: string
  content: string
  managed: boolean
}

type AgentMemoryPreviewItem = {
  agentName: string
  memoryPath: string
  rulesDir: string
  mainExists: boolean
  mainContent: string
  rulesDirExists: boolean
  rules: AgentMemoryRuleItem[]
}

export default function ToolSkills() {
  const { t } = useLanguage()
  const visibility = useSkillStatusVisibility('myAgents')
  const defaultCategory = 'Default'
  const [agents, setAgents] = useState<any[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [selectedAgent, setSelectedAgent] = useState<string>('')
  const [skills, setSkills] = useState<any[]>([])
  const [memoryPreview, setMemoryPreview] = useState<AgentMemoryPreviewItem | null>(null)
  const [memoryLoading, setMemoryLoading] = useState(false)
  const [memoryError, setMemoryError] = useState('')
  const [loading, setLoading] = useState(true)
  const [selectMode, setSelectMode] = useState(false)
  const [selectedPaths, setSelectedPaths] = useState<Set<string>>(new Set())
  const [deleting, setDeleting] = useState(false)
  const [search, setSearch] = useState('')
  const [sortOrder, setSortOrder] = useState<SkillSortOrder>('asc')
  const [activePanel, setActivePanel] = useState<ToolSkillsPanel>(getDefaultToolSkillsPanel())
  const [pullMode, setPullMode] = useState(false)
  const [pullState, setPullState] = useState(createToolSkillsPullState(defaultCategory))
  const [scanned, setScanned] = useState<any[]>([])
  const [scanning, setScanning] = useState(false)
  const [pulling, setPulling] = useState(false)
  const [scanError, setScanError] = useState('')
  const [pullConflicts, setPullConflicts] = useState<string[]>([])
  const [pullDone, setPullDone] = useState(false)

  const loadSkills = useCallback(async (agentName: string) => {
    setLoading(true)
    try {
      const s = await ListAgentSkills(agentName)
      setSkills(s ?? [])
    } finally {
      setLoading(false)
    }
  }, [])

  const loadMemoryPreview = useCallback(async (agentName: string) => {
    setMemoryLoading(true)
    setMemoryError('')
    try {
      const preview = await GetAgentMemoryPreview(agentName)
      setMemoryPreview((preview ?? null) as AgentMemoryPreviewItem | null)
    } catch (error) {
      console.error('Failed to load agent memory preview', error)
      setMemoryPreview(null)
      setMemoryError(t('toolSkills.memoryLoadFailed'))
    } finally {
      setMemoryLoading(false)
    }
  }, [t])

  useEffect(() => {
    let active = true

    const initialize = async () => {
      setLoading(true)
      try {
        const [enabledAgents, listedCategories] = await Promise.all([GetEnabledAgents(), ListCategories()])
        if (!active) return

        const nextAgents = enabledAgents ?? []
        const nextCategories = listedCategories ?? []
        setAgents(nextAgents)
        setCategories(nextCategories)

        if (nextAgents.length === 0) {
          setSkills([])
          setMemoryPreview(null)
          setMemoryError('')
          return
        }

        const initialAgent = nextAgents[0].name
        setSelectedAgent(initialAgent)

        const listedSkills = await ListAgentSkills(initialAgent)
        if (!active) return
        setSkills(listedSkills ?? [])
        await loadMemoryPreview(initialAgent)
      } finally {
        if (active) setLoading(false)
      }
    }

    void initialize()

    return () => {
      active = false
    }
  }, [loadMemoryPreview])

  useEffect(() => {
    if (!selectedAgent) return
    return subscribeToEvents(EventsOn, createToolSkillsEventSubscriptions(() => { void loadSkills(selectedAgent) }))
  }, [loadSkills, selectedAgent])

  const selectAgent = (agentName: string) => {
    setSelectedAgent(agentName)
    setSelectMode(false)
    setSelectedPaths(new Set())
    setPullMode(false)
    setPullState(createToolSkillsPullState(categories[0] ?? defaultCategory))
    setScanned([])
    setScanError('')
    setPullConflicts([])
    setPullDone(false)
    setMemoryPreview(null)
    setMemoryError('')
    void loadSkills(agentName)
    void loadMemoryPreview(agentName)
  }

  const selectPanel = (panel: ToolSkillsPanel) => {
    setActivePanel(panel)
    if (panel !== 'skills') {
      setSelectMode(false)
      setSelectedPaths(new Set())
      setPullMode(false)
      setPullState(createToolSkillsPullState(categories[0] ?? defaultCategory))
      setScanned([])
      setScanError('')
      setPullConflicts([])
      setPullDone(false)
    }
  }

  const handleDelete = async (skillPath: string) => {
    await DeleteAgentSkill(selectedAgent, skillPath)
    loadSkills(selectedAgent)
  }

  const handleBatchDelete = async () => {
    if (selectedPaths.size === 0) return
    setDeleting(true)
    try {
      for (const path of selectedPaths) {
        await DeleteAgentSkill(selectedAgent, path)
      }
      setSelectedPaths(new Set())
      setSelectMode(false)
      loadSkills(selectedAgent)
    } finally {
      setDeleting(false)
    }
  }

  const toggleSelectPath = (path: string) => {
    setSelectedPaths(prev => {
      const next = new Set(prev)
      next.has(path) ? next.delete(path) : next.add(path)
      return next
    })
  }

  const pushSkills = skills.filter(s => s.pushed)
  const scanOnlySkills = skills.filter(s => s.seenInAgentScan && !s.pushed)

  const memoryEntries = useMemo(
    () => buildAgentMemoryEntries(memoryPreview),
    [memoryPreview],
  )

  const filteredScanned = useMemo(
    () => filterAndSortSkills(scanned, search, sortOrder, skill => skill.name ?? ''),
    [scanned, search, sortOrder],
  )

  const { filteredPushSkills, filteredScanOnlySkills, filteredMemoryEntries } = useMemo(
    () => filterToolSkillsPanelContent({
      activePanel,
      search,
      sortOrder,
      pushSkills,
      scanOnlySkills,
      memoryEntries,
    }),
    [activePanel, search, sortOrder, pushSkills, scanOnlySkills, memoryEntries],
  )

  const mainMemoryEntry = useMemo(
    () => filteredMemoryEntries.find(entry => entry.kind === 'main') ?? null,
    [filteredMemoryEntries],
  )

  const ruleEntries = useMemo(
    () => filteredMemoryEntries.filter(entry => entry.kind === 'rule'),
    [filteredMemoryEntries],
  )

  const availableCategories = categories.length > 0 ? categories : [defaultCategory]
  const visibleNotImportedPaths = useMemo(
    () => getToolSkillsVisibleNotImportedPaths(filteredScanned),
    [filteredScanned],
  )
  const pullReady = isToolSkillsPullReady(pullState)

  const toggleSelectAll = () => {
    const visiblePaths = filteredPushSkills.map((skill: any) => skill.path)
    setSelectedPaths(prev => {
      const next = new Set(prev)
      if (visiblePaths.every(path => next.has(path))) {
        visiblePaths.forEach(path => next.delete(path))
      } else {
        visiblePaths.forEach(path => next.add(path))
      }
      return next
    })
  }

  useEffect(() => {
    if (!selectMode) return
    const visiblePaths = new Set(filteredPushSkills.map((skill: any) => skill.path))
    setSelectedPaths(prev => {
      const next = new Set([...prev].filter(path => visiblePaths.has(path)))
      return next.size === prev.size ? prev : next
    })
  }, [filteredPushSkills, selectMode])

  useEffect(() => {
    if (!pullMode) return
    const nextPaths = syncToolSkillsPullVisibleSelection(
      pullState.selectedPaths,
      filteredScanned.map((skill: any) => skill.path),
    )
    const same = nextPaths.length === pullState.selectedPaths.length
      && nextPaths.every((path, index) => path === pullState.selectedPaths[index])
    if (same) return
    setPullState(prev => ({
      ...prev,
      selectedPaths: nextPaths,
    }))
  }, [filteredScanned, pullMode, pullState.selectedPaths])

  const scanForPull = useCallback(async (agentName: string) => {
    setScanning(true)
    setScanned([])
    setScanError('')
    setPullConflicts([])
    setPullDone(false)
    try {
      const scannedSkills = await ScanAgentSkills(agentName)
      setScanned(scannedSkills ?? [])
      setPullState(prev => ({ ...prev, selectedPaths: [] }))
    } catch (error: any) {
      setScanError(String(error?.message ?? error ?? t('toolSkills.pullFailed')))
    } finally {
      setScanning(false)
    }
  }, [t])

  const enterPullMode = useCallback(async () => {
    if (!selectedAgent) return
    setSelectMode(false)
    setSelectedPaths(new Set())
    setPullMode(true)
    setPullState(createToolSkillsPullState(availableCategories[0] ?? defaultCategory))
    await scanForPull(selectedAgent)
  }, [availableCategories, scanForPull, selectedAgent])

  const exitPullMode = () => {
    setPullMode(false)
    setPullState(createToolSkillsPullState(availableCategories[0] ?? defaultCategory))
    setScanned([])
    setScanError('')
    setPullConflicts([])
  }

  const handleStartPull = async () => {
    if (!selectedAgent || !pullReady) return
    setPulling(true)
    setScanError('')
    setPullDone(false)
    try {
      const conflicts = await PullFromAgent(selectedAgent, pullState.selectedPaths, pullState.targetCategory)
      if (conflicts && conflicts.length > 0) {
        setPullConflicts(conflicts)
        return
      }
      setPullDone(true)
      exitPullMode()
      await loadSkills(selectedAgent)
    } catch (error: any) {
      setScanError(String(error?.message ?? error ?? t('toolSkills.pullFailed')))
    } finally {
      setPulling(false)
    }
  }

  const agent = agents.find(t => t.name === selectedAgent)
  const allSelected = filteredPushSkills.length > 0 && filteredPushSkills.every((skill: any) => selectedPaths.has(skill.path))
  const allPullSelected = filteredScanned.length > 0 && filteredScanned.every((skill: any) => pullState.selectedPaths.includes(skill.path))
  const allNotImportedSelected = visibleNotImportedPaths.length > 0 && visibleNotImportedPaths.every(path => pullState.selectedPaths.includes(path))
  const visibleResultCount = pullMode ? filteredScanned.length : getVisibleToolSkillsResultCount({
    activePanel,
    filteredPushSkills,
    filteredScanOnlySkills,
    filteredMemoryEntries,
  })
  const searchPlaceholder = pullMode
    ? t('toolSkills.searchPullPlaceholder')
    : activePanel === 'memory'
    ? t('toolSkills.searchMemoryPlaceholder')
    : t('toolSkills.searchSkillsPlaceholder')

  const getNavStyle = (isActive: boolean) => isActive ? {
    background: 'var(--accent-glow)',
    color: 'var(--accent-primary)',
    border: '1px solid var(--border-accent)',
    boxShadow: 'var(--glow-accent-sm)',
  } : {
    color: 'var(--text-muted)',
    border: '1px solid transparent',
  }

  return (
    <div className="flex h-full overflow-hidden">
      {/* Left: tool list */}
      <div className="w-48 shrink-0 p-3 flex flex-col gap-0.5 overflow-y-auto" style={{ borderRight: '1px solid var(--border-base)' }}>
        <div className="px-3 py-1.5 text-xs font-medium tracking-wide uppercase" style={{ color: 'var(--text-muted)' }}>
          {t('toolSkills.toolList')}
        </div>
        {agents.map(t => (
          <button
            key={t.name}
            onClick={() => selectAgent(t.name)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-left transition-all duration-150"
            style={getNavStyle(selectedAgent === t.name)}
          >
            <ToolIcon name={t.name} size={20} />
            <span className="truncate">{t.name}</span>
          </button>
        ))}
        {!loading && agents.length === 0 && (
          <p className="px-3 text-xs mt-2" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noTools')}</p>
        )}
      </div>

      {/* Main content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Toolbar */}
        <div className="px-6 py-4 flex flex-col gap-4" style={{ borderBottom: '1px solid var(--border-base)' }}>
          <div className="flex items-center gap-3 flex-wrap">
            {agent ? (
              <div className="flex items-center gap-3 flex-wrap">
                <div className="flex items-center gap-2">
                  <ToolIcon name={agent.name} size={22} />
                  <span className="font-medium text-sm" style={{ color: 'var(--text-primary)' }}>{agent.name}</span>
                </div>
                <div
                  className="flex items-center gap-1 rounded-xl p-1"
                  style={{
                    background: 'var(--bg-elevated)',
                    border: '1px solid var(--border-base)',
                  }}
                >
                  {([
                    ['skills', t('toolSkills.panelSkills')],
                    ['memory', t('toolSkills.panelMemory')],
                  ] as const).map(([panel, label]) => {
                    const selected = activePanel === panel
                    return (
                      <button
                        key={panel}
                        type="button"
                        onClick={() => selectPanel(panel)}
                        className="rounded-lg px-3 py-1.5 text-sm transition-all duration-150"
                        style={selected ? {
                          background: 'var(--active-surface)',
                          color: 'var(--active-text)',
                          border: '1px solid var(--active-border)',
                          boxShadow: 'var(--active-shadow)',
                        } : {
                          color: 'var(--text-muted)',
                          border: '1px solid transparent',
                        }}
                      >
                        {label}
                      </button>
                    )
                  })}
                </div>
              </div>
            ) : (
              <h2 className="text-sm font-medium flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
                <Wrench size={14} /> {t('toolSkills.title')}
              </h2>
            )}
            <div className="flex-1" />
            {activePanel === 'skills' && selectMode ? (
              <>
                <button
                  onClick={toggleSelectAll}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
                  style={{ color: 'var(--text-muted)' }}
                  onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                  onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
                >
                  <CheckSquare size={14} />{allSelected ? t('common.deselectAll') : t('common.selectAll')}
                </button>
                <button
                  onClick={handleBatchDelete}
                  disabled={selectedPaths.size === 0 || deleting}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg disabled:opacity-40 transition-colors text-white"
                  style={{ background: 'var(--color-error)' }}
                >
                  <Trash2 size={14} /> {t('common.delete')} {selectedPaths.size > 0 ? `(${selectedPaths.size})` : ''}
                </button>
                <button
                  onClick={() => { setSelectMode(false); setSelectedPaths(new Set()) }}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
                  style={{ color: 'var(--text-muted)' }}
                  onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                  onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
                >
                  {t('common.cancel')}
                </button>
              </>
            ) : activePanel === 'skills' && pullMode ? (
              <>
                <label
                  className="flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg"
                  style={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-base)', color: 'var(--text-secondary)' }}
                >
                  <span>{t('toolSkills.pullTargetCategory')}</span>
                  <select
                    value={pullState.targetCategory}
                    onChange={event => setPullState(prev => ({ ...prev, targetCategory: event.target.value }))}
                    className="rounded-md px-2 py-1 text-sm outline-none"
                    style={{ background: 'var(--bg-surface)', color: 'var(--text-primary)' }}
                  >
                    {availableCategories.map(category => (
                      <option key={category} value={category}>{category}</option>
                    ))}
                  </select>
                </label>
                <button
                  onClick={() => setPullState(prev => ({
                    ...prev,
                    selectedPaths: toggleToolSkillsPullAllVisible(prev.selectedPaths, filteredScanned.map((skill: any) => skill.path)),
                  }))}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
                  style={{ color: 'var(--text-muted)' }}
                  onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                  onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
                >
                  <CheckSquare size={14} />
                  {allPullSelected ? t('common.deselectAll') : t('common.selectAll')}
                </button>
                <button
                  onClick={() => setPullState(prev => ({
                    ...prev,
                    selectedPaths: toggleToolSkillsPullVisibleNotImported(prev.selectedPaths, visibleNotImportedPaths),
                  }))}
                  disabled={visibleNotImportedPaths.length === 0}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors disabled:cursor-not-allowed disabled:opacity-40"
                  style={{ color: 'var(--text-muted)' }}
                  onMouseEnter={e => {
                    if (!e.currentTarget.disabled) {
                      e.currentTarget.style.backgroundColor = 'var(--bg-hover)'
                      e.currentTarget.style.color = 'var(--text-primary)'
                    }
                  }}
                  onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
                >
                  <CheckSquare size={14} />
                  {allNotImportedSelected ? t('common.deselectAll') : t('syncPull.selectNotImported')}
                </button>
                <button
                  onClick={() => void handleStartPull()}
                  disabled={!pullReady || pulling}
                  className="btn-primary flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {pulling ? <RefreshCw size={14} className="animate-spin" /> : <ArrowDownToLine size={14} />}
                  {t('toolSkills.manualPullStart', { count: pullState.selectedPaths.length })}
                </button>
                <button
                  onClick={exitPullMode}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
                  style={{ color: 'var(--text-muted)' }}
                  onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                  onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
                >
                  {t('common.cancel')}
                </button>
              </>
            ) : activePanel === 'skills' ? (
              <>
                <button
                  onClick={() => void enterPullMode()}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
                  style={{ color: 'var(--text-muted)' }}
                  onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                  onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
                >
                  <ArrowDownToLine size={14} /> {t('toolSkills.manualPull')}
                </button>
                {filteredPushSkills.length > 0 && (
                  <button
                    onClick={() => setSelectMode(true)}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
                    style={{ color: 'var(--text-muted)' }}
                    onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                    onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
                  >
                    <CheckSquare size={14} /> {t('toolSkills.batchDelete')}
                  </button>
                )}
              </>
            ) : null
            }
          </div>

          <SkillListControls
            search={search}
            onSearchChange={setSearch}
            sortOrder={sortOrder}
            onSortOrderChange={setSortOrder}
            placeholder={searchPlaceholder}
            resultLabel={t('toolSkills.showingNResults', { count: visibleResultCount })}
          />
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-8">
          {loading ? (
            <div className="flex items-center justify-center h-32 text-sm" style={{ color: 'var(--text-muted)' }}>{t('common.loading')}</div>
          ) : !agent ? (
            <div className="flex flex-col items-center justify-center h-48" style={{ color: 'var(--text-muted)' }}>
              <Wrench size={32} className="mb-2 opacity-30" />
              <p className="text-sm">{t('toolSkills.selectToolFirst')}</p>
            </div>
          ) : activePanel === 'memory' ? (
            <section>
              <div className="flex items-center gap-2 mb-4">
                <Brain size={14} style={{ color: 'var(--accent-primary)' }} className="shrink-0" />
                <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t('toolSkills.memoryTitle')}</span>
                <div className="flex-1" />
                <button
                  onClick={() => void loadMemoryPreview(selectedAgent)}
                  disabled={memoryLoading}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors disabled:opacity-40"
                  style={{ color: 'var(--text-muted)' }}
                  onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                  onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
                  title={t('toolSkills.memoryRefresh')}
                >
                  <RefreshCw size={14} className={memoryLoading ? 'animate-spin' : ''} />
                  {t('toolSkills.memoryRefresh')}
                </button>
              </div>

              {memoryError ? (
                <p className="text-sm pl-5" style={{ color: 'var(--color-error)' }}>{memoryError}</p>
              ) : memoryLoading && !memoryPreview ? (
                <p className="text-sm pl-5" style={{ color: 'var(--text-muted)' }}>{t('common.loading')}</p>
              ) : memoryPreview ? (
                <div className="space-y-4">
                  {mainMemoryEntry && (
                    <div className="card-base p-4">
                      <div className="flex items-start gap-3">
                        <div className="min-w-0 flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t('toolSkills.memoryFile')}</span>
                          </div>
                          {memoryPreview.memoryPath
                            ? <p className="text-xs break-all" style={{ color: 'var(--text-muted)' }}>{memoryPreview.memoryPath}</p>
                            : <p className="text-xs" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.memoryNotConfigured')}</p>
                          }
                        </div>
                        <button
                          onClick={() => { if (memoryPreview.memoryPath) void OpenPath(memoryPreview.memoryPath) }}
                          disabled={!memoryPreview.memoryPath}
                          className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors disabled:opacity-40"
                          style={{ color: 'var(--text-muted)' }}
                          onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                          onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
                          title={t('toolSkills.openFile')}
                        >
                          <FolderOpenDot size={14} />
                          {t('toolSkills.openFile')}
                        </button>
                      </div>

                      <div className="mt-4 rounded-lg border p-3 max-h-56 overflow-y-auto" style={{ borderColor: 'var(--border-base)', background: 'var(--bg-panel)' }}>
                        {!memoryPreview.memoryPath ? (
                          <p className="text-sm" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.memoryNotConfigured')}</p>
                        ) : !memoryPreview.mainExists ? (
                          <p className="text-sm" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.memoryFileMissing')}</p>
                        ) : (
                          <pre className="whitespace-pre-wrap break-words text-sm m-0" style={{ color: 'var(--text-primary)' }}>{mainMemoryEntry.content || t('toolSkills.emptyContent')}</pre>
                        )}
                      </div>
                    </div>
                  )}

                  <div>
                    <div className="flex items-center gap-2 mb-4">
                      <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t('toolSkills.rulesDir')}</span>
                      {memoryPreview.rulesDir
                        ? <span className="text-xs truncate" style={{ color: 'var(--text-muted)' }} title={memoryPreview.rulesDir}>{memoryPreview.rulesDir}</span>
                        : <span className="text-xs" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noPushDir')}</span>
                      }
                      <div className="flex-1" />
                      <button
                        onClick={() => { if (memoryPreview.rulesDir) void OpenPath(memoryPreview.rulesDir) }}
                        disabled={!memoryPreview.rulesDir}
                        className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors disabled:opacity-40"
                        style={{ color: 'var(--text-muted)' }}
                        onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                        onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
                        title={t('toolSkills.openDir')}
                      >
                        <FolderOpenDot size={14} />
                        {t('toolSkills.openDir')}
                      </button>
                    </div>

                    {!memoryPreview.rulesDir ? (
                      <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.rulesNotConfigured')}</p>
                    ) : !memoryPreview.rulesDirExists ? (
                      <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.rulesDirMissing')}</p>
                    ) : ruleEntries.length === 0 ? (
                      <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>
                        {search.trim() ? t('toolSkills.noMemoryMatch') : t('toolSkills.noRuleFiles')}
                      </p>
                    ) : (
                      <div className="grid grid-cols-2 xl:grid-cols-3 gap-4">
                        {ruleEntries.map(entry => (
                          <div key={entry.key} className="card-base p-4">
                            <div className="flex items-start gap-3">
                              <div className="min-w-0 flex-1">
                                <div className="flex items-center gap-2 mb-1">
                                  <span className="text-sm font-medium truncate" style={{ color: 'var(--text-primary)' }}>{entry.title}</span>
                                  {entry.managed && (
                                    <span
                                      className="px-2 py-0.5 rounded-full text-[11px] font-medium"
                                      style={{ background: 'var(--accent-glow)', color: 'var(--accent-primary)', border: '1px solid var(--border-accent)' }}
                                    >
                                      {t('toolSkills.managedRule')}
                                    </span>
                                  )}
                                </div>
                                <p className="text-xs break-all" style={{ color: 'var(--text-muted)' }}>{entry.path}</p>
                              </div>
                              <button
                                onClick={() => void OpenPath(entry.path)}
                                className="p-1 rounded transition-colors"
                                style={{ color: 'var(--text-muted)' }}
                                onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-overlay)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                                onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
                                title={t('toolSkills.openFile')}
                              >
                                <FolderOpenDot size={13} />
                              </button>
                            </div>

                            <div className="mt-3 rounded-lg border p-3 max-h-40 overflow-y-auto" style={{ borderColor: 'var(--border-base)', background: 'var(--bg-panel)' }}>
                              <pre className="whitespace-pre-wrap break-words text-sm m-0" style={{ color: 'var(--text-primary)' }}>{entry.content || t('toolSkills.emptyContent')}</pre>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              ) : null}
            </section>
          ) : pullMode ? (
            <section>
              <div className="flex items-center gap-2 mb-4">
                <ArrowDownToLine size={14} style={{ color: 'var(--accent-primary)' }} className="shrink-0" />
                <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t('toolSkills.manualPullTitle')}</span>
              </div>

              {scanning ? (
                <p className="text-sm pl-5" style={{ color: 'var(--text-muted)' }}>{t('syncPull.scanning')}</p>
              ) : scanError ? (
                <div
                  className="flex items-start gap-2 rounded-lg px-4 py-3 text-sm"
                  style={{ background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)', color: 'var(--color-error)' }}
                >
                  <AlertCircle size={16} className="mt-0.5 shrink-0" />
                  <span className="flex-1">{scanError}</span>
                </div>
              ) : scanned.length === 0 ? (
                <div
                  className="flex items-center gap-2 rounded-lg px-4 py-3 text-sm"
                  style={{ background: 'rgba(251,191,36,0.1)', border: '1px solid rgba(251,191,36,0.3)', color: 'var(--color-warning)' }}
                >
                  <AlertCircle size={16} className="shrink-0" />
                  <span>{t('syncPull.emptyWarning')}</span>
                </div>
              ) : filteredScanned.length === 0 ? (
                <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('syncPull.noMatch')}</p>
              ) : (
                <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
                  {filteredScanned.map((sk: any) => (
                    <SyncSkillCard
                      key={sk.path}
                      name={sk.name}
                      source={sk.source}
                      path={sk.path}
                      imported={sk.imported}
                      updatable={sk.updatable}
                      pushedAgents={(sk.pushedAgents ?? []).filter((agentName: string) => agentName !== selectedAgent)}
                      showImported={visibility.includes('imported')}
                      showUpdatable={visibility.includes('updatable')}
                      showPushedAgents={visibility.includes('pushedAgents')}
                      selected={pullState.selectedPaths.includes(sk.path)}
                      onToggle={() => setPullState(prev => ({
                        ...prev,
                        selectedPaths: toggleToolSkillsPullPath(prev.selectedPaths, sk.path),
                      }))}
                    />
                  ))}
                </div>
              )}
            </section>
          ) : (
            <>
              {pullDone && (
                <div
                  className="flex items-center gap-2 rounded-lg px-4 py-3 text-sm"
                  style={{ background: 'rgba(52,211,153,0.1)', border: '1px solid rgba(52,211,153,0.3)', color: 'var(--color-success)' }}
                >
                  <Check size={16} className="shrink-0" />
                  <span>{t('toolSkills.pullDone')}</span>
                </div>
              )}

              {/* Push dir section */}
              <section>
                <div className="flex items-center gap-2 mb-4">
                  <ArrowUpToLine size={14} style={{ color: 'var(--color-success)' }} className="shrink-0" />
                  <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t('toolSkills.pushPath')}</span>
                  {agent.pushDir
                    ? <span className="text-xs truncate" style={{ color: 'var(--text-muted)' }} title={agent.pushDir}>{agent.pushDir}</span>
                    : <span className="text-xs" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noPushDir')}</span>
                  }
                </div>
                {!agent.pushDir ? (
                  <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noPushDirDesc')}</p>
                ) : pushSkills.length === 0 ? (
                  <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noPushSkills')}</p>
                ) : filteredPushSkills.length === 0 ? (
                  <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noMatch')}</p>
                ) : (
                  <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
                    {filteredPushSkills.map((sk: any) => (
                      <ToolSkillCard
                        key={sk.path}
                        name={sk.name}
                        path={sk.path}
                        source={sk.source}
                        imported={sk.imported}
                        updatable={sk.updatable}
                        pushedAgents={(sk.pushedAgents ?? []).filter((agentName: string) => agentName !== selectedAgent)}
                        showImported={visibility.includes('imported')}
                        showUpdatable={visibility.includes('updatable')}
                        showPushedAgents={visibility.includes('pushedAgents')}
                        canDelete
                        selectMode={selectMode}
                        selected={selectedPaths.has(sk.path)}
                        onToggleSelect={() => toggleSelectPath(sk.path)}
                        onDelete={() => handleDelete(sk.path)}
                      />
                    ))}
                  </div>
                )}
              </section>

              {/* Scan-only section */}
              <section>
                <div className="flex items-center gap-2 mb-4">
                  <ScanLine size={14} style={{ color: 'var(--accent-primary)' }} className="shrink-0" />
                  <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t('toolSkills.scanPath')}</span>
                  {agent.scanDirs?.length > 0 && (
                    <span className="text-xs truncate" style={{ color: 'var(--text-muted)' }} title={agent.scanDirs.join(', ')}>
                      {t('toolSkills.nDirs', { count: agent.scanDirs.length })}
                    </span>
                  )}
                </div>
                <p className="mb-4 pl-5 text-xs" style={{ color: 'var(--text-muted)' }}>{t('toolSkills.scanPathHint')}</p>
                {scanOnlySkills.length === 0 ? (
                  <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noScanSkills')}</p>
                ) : filteredScanOnlySkills.length === 0 ? (
                  <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noMatch')}</p>
                ) : (
                  <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
                    {filteredScanOnlySkills.map((sk: any) => (
                      <ToolSkillCard
                        key={sk.path}
                        name={sk.name}
                        path={sk.path}
                        source={sk.source}
                        imported={sk.imported}
                        updatable={sk.updatable}
                        pushedAgents={(sk.pushedAgents ?? []).filter((agentName: string) => agentName !== selectedAgent)}
                        showImported={visibility.includes('imported')}
                        showUpdatable={visibility.includes('updatable')}
                        showPushedAgents={visibility.includes('pushedAgents')}
                        canDelete={false}
                        selectMode={false}
                        selected={false}
                        onToggleSelect={() => {}}
                        onDelete={() => {}}
                      />
                    ))}
                  </div>
                )}
              </section>
            </>
          )}
        </div>
      </div>

      <ConflictDialog
        conflicts={pullConflicts}
        labelForConflict={(path) => scanned.find((item: any) => item.path === path)?.name ?? path}
        onOverwrite={async (path) => {
          await PullFromAgentForce(selectedAgent, [path], pullState.targetCategory)
          setPullConflicts(prev => prev.filter(conflict => conflict !== path))
        }}
        onSkip={(path) => setPullConflicts(prev => prev.filter(conflict => conflict !== path))}
        onDone={() => {
          setPullDone(true)
          exitPullMode()
          void loadSkills(selectedAgent)
        }}
      />
    </div>
  )
}

interface ToolSkillCardProps {
  name: string
  path: string
  source?: string
  imported?: boolean
  updatable?: boolean
  pushedAgents?: string[]
  showImported: boolean
  showUpdatable: boolean
  showPushedAgents: boolean
  canDelete: boolean
  selectMode: boolean
  selected: boolean
  onToggleSelect: () => void
  onDelete: () => void
}

function ToolSkillCard({
  name,
  path,
  source,
  imported,
  updatable,
  pushedAgents = [],
  showImported,
  showUpdatable,
  showPushedAgents,
  canDelete,
  selectMode,
  selected,
  onToggleSelect,
  onDelete,
}: ToolSkillCardProps) {
  const { t } = useLanguage()
  const cardRef = useRef<HTMLDivElement>(null)
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const [hoveredRect, setHoveredRect] = useState<DOMRect | null>(null)
  const [meta, setMeta] = useState<any | null>(null)
  const [copied, setCopied] = useState(false)
  const sourceLabel = source === 'github'
    ? t('common.sourceGitHub')
    : source === 'manual'
      ? t('common.sourceManual')
      : source === 'git'
        ? t('common.sourceGit')
        : source

  const handleMouseEnter = () => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current)
    hoverTimer.current = setTimeout(async () => {
      if (!cardRef.current) return
      setHoveredRect(cardRef.current.getBoundingClientRect())
      setMeta(null)
      try {
        const m = await GetSkillMetaByPath(path)
        setMeta(m ?? {})
      } catch {
        setMeta({})
      }
    }, 300)
  }

  const handleMouseLeave = () => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current)
    setHoveredRect(null)
    setMeta(null)
  }

  const handleOpen = (e: React.MouseEvent) => {
    e.stopPropagation()
    OpenPath(path)
  }

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation()
    try {
      const content = await ReadSkillFileContent(path)
      await copyTextToClipboard(content)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch { /* ignore */ }
  }

  const handleClick = () => {
    if (selectMode && canDelete) onToggleSelect()
  }

  return (
    <>
      <div
        ref={cardRef}
        onClick={handleClick}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        className={`card-base relative p-4 group ${selectMode && canDelete ? 'cursor-pointer' : 'cursor-default'}`}
        style={selected ? {
          background: 'var(--accent-glow)',
          borderColor: 'var(--border-accent)',
          boxShadow: 'var(--glow-accent-sm)',
        } : undefined}
      >
        {/* Select checkbox */}
        {selectMode && canDelete && (
          <div className="absolute top-2 left-2 z-10">
            <div
              className="w-4 h-4 rounded border-2 flex items-center justify-center transition-all duration-150"
              style={selected ? {
                background: 'var(--accent-secondary)',
                borderColor: 'var(--accent-secondary)',
                boxShadow: 'var(--glow-accent-sm)',
              } : {
                borderColor: 'var(--text-muted)',
                background: 'var(--bg-elevated)',
              }}
            >
              {selected && (
                <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                </svg>
              )}
            </div>
          </div>
        )}

        {/* Hover action buttons */}
        {!selectMode && (
          <div className="absolute top-2 right-2 flex items-center gap-0.5 z-10 opacity-0 group-hover:opacity-100 transition-opacity">
            <button
              onClick={handleCopy}
              title={t('toolSkills.copySkill')}
              className="p-1 rounded transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-overlay)'; e.currentTarget.style.color = 'var(--text-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
            >
              {copied ? <Check size={13} style={{ color: 'var(--color-success)' }} /> : <Copy size={13} />}
            </button>
            <button
              onClick={handleOpen}
              title={t('toolSkills.openDir')}
              className="p-1 rounded transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-overlay)'; e.currentTarget.style.color = 'var(--text-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
            >
              <FolderOpenDot size={13} />
            </button>
          </div>
        )}

        <SkillStatusStrip
          className={`${selectMode && canDelete ? 'pl-5' : ''} pr-12`}
          badges={[
            ...(sourceLabel ? [{
              key: `source:${sourceLabel}`,
              label: sourceLabel,
              tone: source === 'github' ? ('accent' as const) : ('muted' as const),
            }] : []),
            ...(showImported && imported ? [{
              key: 'imported',
              label: t('common.imported'),
              tone: 'success' as const,
            }] : []),
            ...(showUpdatable && updatable ? [{
              key: 'updatable',
              label: t('common.updatable'),
              tone: 'warning' as const,
            }] : []),
          ]}
          pushedAgents={showPushedAgents ? pushedAgents : []}
        />

        <p
          className={`mt-1 min-h-[2.75rem] font-medium text-sm leading-snug line-clamp-2 ${selectMode && canDelete ? 'pl-5' : ''} ${!selectMode ? 'pr-5' : ''}`}
          style={{ color: 'var(--text-primary)' }}
        >
          {name}
        </p>

        {/* Delete button */}
        {!selectMode && canDelete && (
          <div className="mt-3 flex opacity-0 group-hover:opacity-100 transition-opacity">
            <button
              onClick={e => { e.stopPropagation(); onDelete() }}
              className="text-xs ml-auto transition-colors"
              style={{ color: 'var(--color-error)' }}
            >
              {t('toolSkills.delete')}
            </button>
          </div>
        )}

        {/* Read-only badge */}
        {!canDelete && (
          <div className="mt-3 flex">
            <span className="text-xs ml-auto" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.readOnly')}</span>
          </div>
        )}
      </div>

      {hoveredRect && (
        <SkillTooltip skill={{ Name: name, Source: source }} meta={meta} anchorRect={hoveredRect} />
      )}
    </>
  )
}
