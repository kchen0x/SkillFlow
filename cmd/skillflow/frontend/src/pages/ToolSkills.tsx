import { useEffect, useMemo, useRef, useState } from 'react'
import { GetEnabledAgents, ListAgentSkills, DeleteAgentSkill, OpenPath, ReadSkillFileContent, GetSkillMetaByPath } from '../../wailsjs/go/main/App'
import { ToolIcon } from '../config/toolIcons'
import SkillTooltip from '../components/SkillTooltip'
import SkillStatusStrip from '../components/SkillStatusStrip'
import SkillListControls from '../components/SkillListControls'
import { copyTextToClipboard } from '../lib/clipboard'
import { SkillSortOrder, filterAndSortSkills } from '../lib/skillList'
import { Wrench, Trash2, FolderOpenDot, Copy, Check, CheckSquare, ArrowUpToLine, ScanLine } from 'lucide-react'
import { useLanguage } from '../contexts/LanguageContext'
import { useSkillStatusVisibility } from '../contexts/SkillStatusVisibilityContext'

export default function ToolSkills() {
  const { t } = useLanguage()
  const visibility = useSkillStatusVisibility('myAgents')
  const [agents, setAgents] = useState<any[]>([])
  const [selectedAgent, setSelectedAgent] = useState<string>('')
  const [skills, setSkills] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [selectMode, setSelectMode] = useState(false)
  const [selectedPaths, setSelectedPaths] = useState<Set<string>>(new Set())
  const [deleting, setDeleting] = useState(false)
  const [search, setSearch] = useState('')
  const [sortOrder, setSortOrder] = useState<SkillSortOrder>('asc')

  useEffect(() => {
    let active = true

    const initialize = async () => {
      setLoading(true)
      try {
        const enabledAgents = await GetEnabledAgents()
        if (!active) return

        const nextAgents = enabledAgents ?? []
        setAgents(nextAgents)

        if (nextAgents.length === 0) {
          setSkills([])
          return
        }

        const initialAgent = nextAgents[0].name
        setSelectedAgent(initialAgent)

        const listedSkills = await ListAgentSkills(initialAgent)
        if (!active) return
        setSkills(listedSkills ?? [])
      } finally {
        if (active) setLoading(false)
      }
    }

    void initialize()

    return () => {
      active = false
    }
  }, [])

  const loadSkills = async (agentName: string) => {
    setLoading(true)
    try {
      const s = await ListAgentSkills(agentName)
      setSkills(s ?? [])
    } finally {
      setLoading(false)
    }
  }

  const selectAgent = (agentName: string) => {
    setSelectedAgent(agentName)
    setSelectMode(false)
    setSelectedPaths(new Set())
    loadSkills(agentName)
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

  const filteredPushSkills = useMemo(
    () => filterAndSortSkills(pushSkills, search, sortOrder, skill => skill.name ?? ''),
    [pushSkills, search, sortOrder],
  )

  const filteredScanOnlySkills = useMemo(
    () => filterAndSortSkills(scanOnlySkills, search, sortOrder, skill => skill.name ?? ''),
    [scanOnlySkills, search, sortOrder],
  )

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

  const agent = agents.find(t => t.name === selectedAgent)
  const allSelected = filteredPushSkills.length > 0 && filteredPushSkills.every((skill: any) => selectedPaths.has(skill.path))

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
              <div className="flex items-center gap-2">
                <ToolIcon name={agent.name} size={22} />
                <span className="font-medium text-sm" style={{ color: 'var(--text-primary)' }}>{agent.name}</span>
              </div>
            ) : (
              <h2 className="text-sm font-medium flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
                <Wrench size={14} /> {t('toolSkills.title')}
              </h2>
            )}
            <div className="flex-1" />
            {selectMode ? (
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
            ) : (
              filteredPushSkills.length > 0 && (
                <button
                  onClick={() => setSelectMode(true)}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
                  style={{ color: 'var(--text-muted)' }}
                  onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                  onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
                >
                  <CheckSquare size={14} /> {t('toolSkills.batchDelete')}
                </button>
              )
            )}
          </div>

          <SkillListControls
            search={search}
            onSearchChange={setSearch}
            sortOrder={sortOrder}
            onSortOrderChange={setSortOrder}
            placeholder={t('toolSkills.searchPlaceholder')}
            resultLabel={t('common.showingNSkills', { count: filteredPushSkills.length + filteredScanOnlySkills.length })}
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
          ) : (
            <>
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
