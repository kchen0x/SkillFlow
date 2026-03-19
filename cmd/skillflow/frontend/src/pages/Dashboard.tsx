import { useEffect, useRef, useState, useCallback, useMemo, type CSSProperties } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  ListSkills, ListCategories, MoveSkillCategory,
  DeleteSkill, DeleteSkills, ImportLocal, UpdateSkill, CheckUpdates,
  OpenFolderDialog, GetSkillMeta, GetConfig, SaveConfig,
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import CategoryPanel from '../components/CategoryPanel'
import SkillCard from '../components/SkillCard'
import SkillTooltip from '../components/SkillTooltip'
import { FolderOpen, RefreshCw, Trash2, CheckSquare, ArrowUpFromLine, AlertCircle, CheckCircle2, X } from 'lucide-react'
import {
  gridContainerVariants,
  cardVariants,
  shouldAnimateSkillCards,
  shouldAnimateSkillGridIntro,
} from '../lib/motionVariants'
import SkillListControls from '../components/SkillListControls'
import { getListLoadState } from '../lib/listLoadState'
import { SkillSortOrder, filterAndSortSkills } from '../lib/skillList'
import {
  createDashboardSkillEventSubscriptions,
  listDashboardToolbarActionKeys,
  readDashboardSkillSettings,
  toggleDashboardAutoPushAgent,
} from '../lib/dashboardSkillSettings'
import { useLanguage } from '../contexts/LanguageContext'
import { useSkillStatusVisibility } from '../contexts/SkillStatusVisibilityContext'
import { ToolIcon } from '../config/toolIcons'
import { subscribeToEvents } from '../lib/wailsEvents'

type SkillUpdateNotice = {
  tone: 'progress' | 'success' | 'error'
  message: string
}

export default function Dashboard() {
  const { t } = useLanguage()
  const visibility = useSkillStatusVisibility('mySkills')
  const [skills, setSkills] = useState<any[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [selectedCat, setSelectedCat] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [sortOrder, setSortOrder] = useState<SkillSortOrder>('asc')
  const [dragOver, setDragOver] = useState(false)
  const [draggingSkillID, setDraggingSkillID] = useState<string | null>(null)
  const [categoryDragActive, setCategoryDragActive] = useState(false)
  const [selectMode, setSelectMode] = useState(false)
  const [selectedIDs, setSelectedIDs] = useState<Set<string>>(new Set())
  const [agentOptions, setAgentOptions] = useState<any[]>([])
  const [autoPushAgents, setAutoPushAgents] = useState<Set<string>>(new Set())
  const [autoUpdateSkills, setAutoUpdateSkills] = useState(false)
  const [dashboardCfg, setDashboardCfg] = useState<any | null>(null)
  const [loading, setLoading] = useState(true)
  const [savingAutoPush, setSavingAutoPush] = useState(false)
  const [updatingSkillIDs, setUpdatingSkillIDs] = useState<Set<string>>(new Set())
  const [updateNotice, setUpdateNotice] = useState<SkillUpdateNotice | null>(null)

  // Hover tooltip state
  const [hoveredSkill, setHoveredSkill] = useState<{ skill: any; rect: DOMRect } | null>(null)
  const [hoveredMeta, setHoveredMeta] = useState<any | null>(null)
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const updateNoticeTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const hasRenderedSkillGrid = useRef(false)

  const syncDashboardConfigState = useCallback((cfg: any) => {
    const settings = readDashboardSkillSettings(cfg)
    setDashboardCfg(cfg)
    setAgentOptions((cfg?.agents ?? []).filter((agent: any) => agent.enabled))
    setAutoPushAgents(new Set(settings.autoPushAgents))
    setAutoUpdateSkills(settings.autoUpdateSkills)
  }, [])

  const load = useCallback(async (options?: { silent?: boolean }) => {
    if (!options?.silent) setLoading(true)
    try {
      const [s, c, cfg] = await Promise.all([ListSkills(), ListCategories(), GetConfig()])
      setSkills(s ?? [])
      setCategories(c ?? [])
      syncDashboardConfigState(cfg)
    } finally {
      if (!options?.silent) setLoading(false)
    }
  }, [syncDashboardConfigState])

  useEffect(() => {
    load()

    return subscribeToEvents(EventsOn, createDashboardSkillEventSubscriptions(load))
  }, [load])

  useEffect(() => {
    if (updateNoticeTimer.current) {
      clearTimeout(updateNoticeTimer.current)
      updateNoticeTimer.current = null
    }

    if (updateNotice && updateNotice.tone !== 'progress') {
      updateNoticeTimer.current = setTimeout(() => setUpdateNotice(null), 4200)
    }

    return () => {
      if (updateNoticeTimer.current) {
        clearTimeout(updateNoticeTimer.current)
        updateNoticeTimer.current = null
      }
    }
  }, [updateNotice])

  const filtered = useMemo(
    () => filterAndSortSkills(
      skills.filter(sk => selectedCat === null || sk.category === selectedCat),
      search,
      sortOrder,
      sk => sk.name ?? '',
    ),
    [skills, selectedCat, search, sortOrder],
  )
  const listState = getListLoadState({ isLoading: loading, itemCount: filtered.length })
  const animateCards = shouldAnimateSkillCards(filtered.length)
  const animateGridIntro = shouldAnimateSkillGridIntro(filtered.length, hasRenderedSkillGrid.current)

  useEffect(() => {
    if (listState !== 'loading') {
      // Only play the staggered card intro on the first settled render.
      hasRenderedSkillGrid.current = true
    }
  }, [listState])

  const skillCounts = skills.reduce((acc, sk) => {
    const category = sk.category || 'Default'
    acc[category] = (acc[category] ?? 0) + 1
    return acc
  }, {} as Record<string, number>)

  const handleDrop = async (skillId: string, category: string) => {
    await MoveSkillCategory(skillId, category)
    setDraggingSkillID(null)
    setCategoryDragActive(false)
    load()
  }

  const isFileDrag = (e: React.DragEvent) => e.dataTransfer.types.includes('Files')

  const handleWindowDragOver = (e: React.DragEvent) => {
    if (!isFileDrag(e)) return
    e.preventDefault()
    setDragOver(true)
  }
  const handleWindowDragLeave = (e: React.DragEvent) => {
    if (!isFileDrag(e)) return
    setDragOver(false)
  }
  const handleWindowDrop = async (e: React.DragEvent) => {
    if (!isFileDrag(e)) return
    e.preventDefault()
    setDragOver(false)
    const items = Array.from(e.dataTransfer.items)
    for (const item of items) {
      const entry = item.webkitGetAsEntry?.()
      if (entry?.isDirectory) {
        const file = item.getAsFile()
        if (file) {
          // @ts-ignore — Wails provides .path on File objects
          await ImportLocal(file.path ?? file.name, selectedCat ?? '')
          load()
        }
      }
    }
  }

  const handleImportButton = async (): Promise<void> => {
    const dir = await OpenFolderDialog('')
    if (dir) { await ImportLocal(dir, selectedCat ?? ''); load() }
  }

  const toggleSelectMode = () => {
    setSelectMode(prev => !prev)
    setSelectedIDs(new Set())
    clearHover()
  }

  const toggleSelectID = (id: string) => {
    setSelectedIDs(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const toggleSelectAll = () => {
    const filteredIDs = filtered.map(sk => sk.id)
    setSelectedIDs(prev => {
      const next = new Set(prev)
      if (filteredIDs.every(id => next.has(id))) {
        filteredIDs.forEach(id => next.delete(id))
      } else {
        filteredIDs.forEach(id => next.add(id))
      }
      return next
    })
  }

  const handleBatchDelete = async () => {
    if (selectedIDs.size === 0) return
    await DeleteSkills(Array.from(selectedIDs))
    setSelectedIDs(new Set())
    setSelectMode(false)
    load()
  }

  useEffect(() => {
    if (!selectMode) return
    const visibleIDs = new Set(filtered.map(sk => sk.id))
    setSelectedIDs(prev => {
      const next = new Set([...prev].filter(id => visibleIDs.has(id)))
      return next.size === prev.size ? prev : next
    })
  }, [filtered, selectMode])

  const allSelected = filtered.length > 0 && filtered.every(sk => selectedIDs.has(sk.id))

  const toolbarButtons = useMemo(() => listDashboardToolbarActionKeys().map((key) => {
    switch (key) {
      case 'update':
        return {
          key,
          icon: <RefreshCw size={14} />,
          label: t('dashboard.update'),
          onClick: async (): Promise<void> => { await CheckUpdates(); load() },
          disabled: false,
          className: 'flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg whitespace-nowrap transition-colors',
          style: { color: 'var(--text-muted)' },
          title: undefined as string | undefined,
        }
      case 'batchDelete':
        return {
          key,
          icon: <CheckSquare size={14} />,
          label: t('dashboard.batchDelete'),
          onClick: toggleSelectMode,
          disabled: false,
          className: 'flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg whitespace-nowrap transition-colors',
          style: { color: 'var(--text-muted)' },
          title: undefined as string | undefined,
        }
      case 'import':
        return {
          key,
          icon: <FolderOpen size={14} />,
          label: t('dashboard.import'),
          onClick: handleImportButton,
          disabled: false,
          className: 'flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg whitespace-nowrap transition-colors',
          style: { color: 'var(--text-muted)' },
          title: undefined as string | undefined,
        }
      case 'autoUpdate':
        return {
          key,
          icon: <RefreshCw size={14} className={autoUpdateSkills ? '' : 'opacity-80'} />,
          label: t('dashboard.autoUpdateTitle'),
          onClick: (): void => { void toggleAutoUpdateSetting() },
          disabled: savingAutoPush,
          className: `btn-primary flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg whitespace-nowrap transition-all disabled:cursor-not-allowed disabled:opacity-60 ${autoUpdateSkills ? '' : 'opacity-70'}`,
          style: undefined as CSSProperties | undefined,
          title: t('dashboard.autoUpdateDesc'),
        }
    }
  }), [autoUpdateSkills, load, savingAutoPush, t])

  const toggleAutoPushAgent = async (name: string) => {
    if (!dashboardCfg || savingAutoPush) return

    const nextAgents = toggleDashboardAutoPushAgent(Array.from(autoPushAgents), name)

    const nextCfg = {
      ...dashboardCfg,
      autoPushAgents: nextAgents,
    }

    setAutoPushAgents(new Set(nextAgents))
    setDashboardCfg(nextCfg)
    setSavingAutoPush(true)

    try {
      await SaveConfig(nextCfg)
      await load()
    } catch (error) {
      console.error('Save auto push agents failed:', error)
      const latestCfg = await GetConfig()
      syncDashboardConfigState(latestCfg)
    } finally {
      setSavingAutoPush(false)
    }
  }

  const toggleAutoUpdateSetting = async () => {
    if (!dashboardCfg || savingAutoPush) return

    const nextCfg = {
      ...dashboardCfg,
      autoUpdateSkills: !autoUpdateSkills,
    }

    setAutoUpdateSkills(!autoUpdateSkills)
    setDashboardCfg(nextCfg)
    setSavingAutoPush(true)

    try {
      await SaveConfig(nextCfg)
      await load()
    } catch (error) {
      console.error('Save auto update skills failed:', error)
      const latestCfg = await GetConfig()
      syncDashboardConfigState(latestCfg)
    } finally {
      setSavingAutoPush(false)
    }
  }

  const clearHover = () => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current)
    setHoveredSkill(null)
    setHoveredMeta(null)
  }

  const handleHoverStart = (sk: any, rect: DOMRect) => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current)
    hoverTimer.current = setTimeout(async () => {
      setHoveredSkill({
        skill: {
          Name: sk.name,
          Category: sk.category,
          Source: sk.source,
          SourceSHA: sk.sourceSha,
          LatestSHA: sk.latestSha,
        },
        rect,
      })
      setHoveredMeta(null)
      const meta = await GetSkillMeta(sk.id)
      setHoveredMeta(meta)
    }, 300)
  }

  const handleHoverEnd = () => {
    clearHover()
  }

  const handleSkillUpdate = async (sk: any) => {
    if (updatingSkillIDs.has(sk.id)) return

    setUpdatingSkillIDs(prev => {
      const next = new Set(prev)
      next.add(sk.id)
      return next
    })
    setUpdateNotice({
      tone: 'progress',
      message: t('dashboard.skillUpdateRunning', { name: sk.name }),
    })

    try {
      await UpdateSkill(sk.id)
      await load({ silent: true })
      setUpdateNotice({
        tone: 'success',
        message: t('dashboard.skillUpdateSuccess', { name: sk.name }),
      })
    } catch (error: any) {
      setUpdateNotice({
        tone: 'error',
        message: t('dashboard.skillUpdateFailed', {
          name: sk.name,
          msg: String(error?.message ?? error ?? t('common.confirm')),
        }),
      })
    } finally {
      setUpdatingSkillIDs(prev => {
        const next = new Set(prev)
        next.delete(sk.id)
        return next
      })
    }
  }

  const containerVariants = gridContainerVariants(filtered.length)

  return (
    <div
      className={`flex h-full relative ${dragOver ? 'ring-2 ring-inset' : ''}`}
      style={dragOver ? { '--tw-ring-color': 'var(--accent-primary)' } as any : {}}
      onDragOver={handleWindowDragOver}
      onDragLeave={handleWindowDragLeave}
      onDrop={handleWindowDrop}
    >
      {dragOver && (
        <div className="absolute inset-0 flex items-center justify-center z-40 pointer-events-none"
          style={{ background: 'var(--active-surface)', backdropFilter: 'blur(6px)' }}>
          <p className="text-lg font-medium" style={{ color: 'var(--active-text)' }}>{t('dashboard.dropToImport')}</p>
        </div>
      )}

      <CategoryPanel
        categories={categories}
        skillCounts={skillCounts}
        selected={selectedCat}
        draggingSkillId={draggingSkillID}
        onSelect={setSelectedCat}
        onCategoryDragStateChange={setCategoryDragActive}
        onDrop={handleDrop}
        onRefresh={load}
      />

      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Toolbar */}
        <div
          className="flex flex-wrap items-center gap-3 px-6 py-4"
          style={{ borderBottom: '1px solid var(--border-base)' }}
        >
          <SkillListControls
            search={search}
            onSearchChange={setSearch}
            sortOrder={sortOrder}
            onSortOrderChange={setSortOrder}
            placeholder={t('dashboard.searchPlaceholder')}
            resultLabel={t('common.showingNSkills', { count: filtered.length })}
          />

          {selectMode ? (
            <div className="flex flex-wrap items-center gap-2 min-w-0">
              <button
                onClick={toggleSelectAll}
                className="flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
              >
                <CheckSquare size={14} />
                {allSelected ? t('common.deselectAll') : t('common.selectAll')}
              </button>
              <button
                onClick={handleBatchDelete}
                disabled={selectedIDs.size === 0}
                className="flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                style={{ backgroundColor: 'var(--color-error)', color: 'white' }}
              >
                <Trash2 size={14} /> {t('common.delete')} {selectedIDs.size > 0 ? `(${selectedIDs.size})` : ''}
              </button>
              <button
                onClick={toggleSelectMode}
                className="flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
              >
                {t('common.cancel')}
              </button>
            </div>
          ) : (
            <div className="flex flex-wrap items-center gap-2 min-w-0">
              {toolbarButtons.map(btn => (
                <button
                  key={btn.key}
                  onClick={btn.onClick}
                  disabled={btn.disabled}
                  className={btn.className}
                  style={btn.style}
                  title={btn.title}
                  onMouseEnter={btn.style ? e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' } : undefined}
                  onMouseLeave={btn.style ? e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' } : undefined}
                >
                  {btn.icon} {btn.label}
                </button>
              ))}
            </div>
          )}
        </div>

        <AnimatePresence initial={false}>
          {updateNotice && (
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              transition={{ duration: 0.18, ease: 'easeOut' }}
              className="mx-6 mt-4 rounded-xl border px-4 py-3"
              style={updateNotice.tone === 'success' ? {
                background: 'color-mix(in srgb, var(--color-success) 14%, var(--bg-elevated) 86%)',
                borderColor: 'color-mix(in srgb, var(--color-success) 34%, transparent)',
              } : updateNotice.tone === 'error' ? {
                background: 'color-mix(in srgb, var(--color-error) 14%, var(--bg-elevated) 86%)',
                borderColor: 'color-mix(in srgb, var(--color-error) 34%, transparent)',
              } : {
                background: 'color-mix(in srgb, var(--accent-primary) 12%, var(--bg-elevated) 88%)',
                borderColor: 'color-mix(in srgb, var(--accent-primary) 28%, transparent)',
              }}
            >
              <div className="flex items-start gap-3">
                <div className="mt-0.5 shrink-0" style={updateNotice.tone === 'success' ? {
                  color: 'var(--color-success)',
                } : updateNotice.tone === 'error' ? {
                  color: 'var(--color-error)',
                } : {
                  color: 'var(--accent-primary)',
                }}>
                  {updateNotice.tone === 'success'
                    ? <CheckCircle2 size={16} />
                    : updateNotice.tone === 'error'
                      ? <AlertCircle size={16} />
                      : <RefreshCw size={16} className="animate-spin" />}
                </div>
                <div className="min-w-0 flex-1">
                  <p className="text-sm font-medium break-words" style={{ color: 'var(--text-primary)' }}>
                    {updateNotice.message}
                  </p>
                </div>
                <button
                  onClick={() => setUpdateNotice(null)}
                  className="shrink-0 rounded-md p-1 transition-colors"
                  style={{ color: 'var(--text-muted)' }}
                  title={t('common.close')}
                  onMouseEnter={e => {
                    e.currentTarget.style.backgroundColor = 'var(--bg-hover)'
                    e.currentTarget.style.color = 'var(--text-primary)'
                  }}
                  onMouseLeave={e => {
                    e.currentTarget.style.backgroundColor = ''
                    e.currentTarget.style.color = 'var(--text-muted)'
                  }}
                >
                  <X size={14} />
                </button>
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        <div
          className="px-6 py-3 flex flex-col gap-3"
          style={{
            borderBottom: '1px solid var(--border-base)',
            background: 'linear-gradient(135deg, color-mix(in srgb, var(--accent-glow) 42%, transparent) 0%, color-mix(in srgb, var(--bg-elevated) 94%, transparent) 100%)',
          }}
        >
          <div className="flex flex-wrap items-center gap-3">
            {loading ? (
              <p className="text-sm flex-1" style={{ color: 'var(--text-muted)' }}>
                {t('common.loading')}
              </p>
            ) : agentOptions.length > 0 ? (
              <div className="flex flex-wrap items-center gap-2 min-w-0 flex-1">
                {agentOptions.map(agent => {
                  const active = autoPushAgents.has(agent.name)
                  return (
                    <button
                      key={agent.name}
                      onClick={() => void toggleAutoPushAgent(agent.name)}
                      disabled={savingAutoPush}
                      className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm transition-all duration-200 disabled:opacity-60 disabled:cursor-not-allowed ${active ? 'font-semibold -translate-y-px' : ''}`}
                      style={active ? {
                        background: 'var(--active-surface)',
                        color: 'var(--active-text)',
                        border: '1px solid var(--active-border)',
                        boxShadow: 'var(--active-shadow)',
                      } : {
                        background: 'var(--bg-elevated)',
                        color: 'var(--text-secondary)',
                        border: '1px solid var(--border-base)',
                      }}
                    >
                      <ToolIcon name={agent.name} size={20} />
                      <span>{agent.name}</span>
                    </button>
                  )
                })}
                {savingAutoPush && (
                  <span className="text-xs whitespace-nowrap ml-1" style={{ color: 'var(--text-muted)' }}>
                    {t('common.saving')}
                  </span>
                )}
              </div>
            ) : (
              <p className="text-sm flex-1" style={{ color: 'var(--text-muted)' }}>
                {t('dashboard.autoPushEmpty')}
              </p>
            )}
            <div className="flex items-center gap-2 shrink-0 text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>
              <ArrowUpFromLine size={15} />
              {t('dashboard.autoPushTitle')}
            </div>
          </div>
        </div>

        {/* Skills grid */}
        <div className="flex-1 overflow-y-auto p-6">
          {listState === 'loading' ? (
            <div className="flex items-center justify-center h-48 text-sm" style={{ color: 'var(--text-muted)' }}>
              {t('common.loading')}
            </div>
          ) : listState === 'empty' ? (
            <div className="flex flex-col items-center justify-center h-48" style={{ color: 'var(--text-muted)' }}>
              <p className="text-sm">{t('dashboard.empty')}</p>
              <p className="text-xs mt-1">{t('dashboard.emptyHint')}</p>
            </div>
          ) : (
            <motion.div
              className="grid grid-cols-3 xl:grid-cols-4 gap-4"
              variants={containerVariants}
              initial={animateGridIntro ? 'initial' : false}
              animate="animate"
            >
              {filtered.map(sk => (
                <motion.div
                  key={sk.id}
                  variants={animateCards ? cardVariants : undefined}
                  initial={animateGridIntro ? 'initial' : false}
                >
                  <SkillCard
                    skill={{
                      id: sk.id,
                      name: sk.name,
                      category: sk.category,
                      source: sk.source,
                      hasUpdate: !!sk.updatable,
                      path: sk.path,
                      pushedAgents: sk.pushedAgents,
                    }}
                    showUpdatable={visibility.includes('updatable')}
                    showPushedAgents={visibility.includes('pushedAgents')}
                    categories={categories}
                    onDelete={async () => { await DeleteSkill(sk.id); load() }}
                    onUpdate={() => handleSkillUpdate(sk)}
                    onMoveCategory={async cat => { await MoveSkillCategory(sk.id, cat); load() }}
                    updating={updatingSkillIDs.has(sk.id)}
                    dragging={draggingSkillID === sk.id}
                    dropTargetActive={draggingSkillID === sk.id && categoryDragActive}
                    onDragStateChange={(dragging) => {
                      setDraggingSkillID(dragging ? sk.id : null)
                      if (!dragging) setCategoryDragActive(false)
                    }}
                    selectMode={selectMode}
                    selected={selectedIDs.has(sk.id)}
                    onToggleSelect={() => toggleSelectID(sk.id)}
                    onHoverStart={rect => handleHoverStart(sk, rect)}
                    onHoverEnd={handleHoverEnd}
                  />
                </motion.div>
              ))}
            </motion.div>
          )}
        </div>
      </div>

      {hoveredSkill && (
        <SkillTooltip
          skill={hoveredSkill.skill}
          meta={hoveredMeta}
          anchorRect={hoveredSkill.rect}
        />
      )}
    </div>
  )
}
