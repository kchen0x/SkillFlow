import { useEffect, useRef, useState } from 'react'
import { GetEnabledTools, ListToolSkills, DeleteToolSkill, OpenPath, ReadSkillFileContent, GetSkillMetaByPath } from '../../wailsjs/go/main/App'
import { ToolIcon } from '../config/toolIcons'
import SkillTooltip from '../components/SkillTooltip'
import { Wrench, Trash2, FolderOpenDot, Copy, Check, CheckSquare, ArrowUpToLine, ScanLine } from 'lucide-react'

export default function ToolSkills() {
  const [tools, setTools] = useState<any[]>([])
  const [selectedTool, setSelectedTool] = useState<string>('')
  const [skills, setSkills] = useState<any[]>([])
  const [loading, setLoading] = useState(false)
  const [selectMode, setSelectMode] = useState(false)
  const [selectedPaths, setSelectedPaths] = useState<Set<string>>(new Set())
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    GetEnabledTools().then(t => {
      setTools(t ?? [])
      if (t && t.length > 0) {
        setSelectedTool(t[0].name)
        loadSkills(t[0].name)
      }
    })
  }, [])

  const loadSkills = async (toolName: string) => {
    setLoading(true)
    try {
      const s = await ListToolSkills(toolName)
      setSkills(s ?? [])
    } finally {
      setLoading(false)
    }
  }

  const selectTool = (toolName: string) => {
    setSelectedTool(toolName)
    setSelectMode(false)
    setSelectedPaths(new Set())
    loadSkills(toolName)
  }

  const handleDelete = async (skillPath: string) => {
    await DeleteToolSkill(selectedTool, skillPath)
    loadSkills(selectedTool)
  }

  const handleBatchDelete = async () => {
    if (selectedPaths.size === 0) return
    setDeleting(true)
    try {
      for (const path of selectedPaths) {
        await DeleteToolSkill(selectedTool, path)
      }
      setSelectedPaths(new Set())
      setSelectMode(false)
      loadSkills(selectedTool)
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

  const pushSkills = skills.filter(s => s.inPush)
  const scanOnlySkills = skills.filter(s => s.inScan && !s.inPush)

  const toggleSelectAll = () => {
    if (selectedPaths.size === pushSkills.length) {
      setSelectedPaths(new Set())
    } else {
      setSelectedPaths(new Set(pushSkills.map((s: any) => s.path)))
    }
  }

  const tool = tools.find(t => t.name === selectedTool)
  const allSelected = pushSkills.length > 0 && selectedPaths.size === pushSkills.length

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
          工具列表
        </div>
        {tools.map(t => (
          <button
            key={t.name}
            onClick={() => selectTool(t.name)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-left transition-all duration-150"
            style={getNavStyle(selectedTool === t.name)}
          >
            <ToolIcon name={t.name} size={20} />
            <span className="truncate">{t.name}</span>
          </button>
        ))}
        {tools.length === 0 && (
          <p className="px-3 text-xs mt-2" style={{ color: 'var(--text-disabled)' }}>没有启用的工具，请在设置中启用</p>
        )}
      </div>

      {/* Main content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Toolbar */}
        <div className="flex items-center gap-3 px-6 py-4 flex-wrap" style={{ borderBottom: '1px solid var(--border-base)' }}>
          {tool ? (
            <div className="flex items-center gap-2">
              <ToolIcon name={tool.name} size={22} />
              <span className="font-medium text-sm" style={{ color: 'var(--text-primary)' }}>{tool.name}</span>
            </div>
          ) : (
            <h2 className="text-sm font-medium flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
              <Wrench size={14} /> 我的工具
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
                <CheckSquare size={14} />{allSelected ? '取消全选' : '全选'}
              </button>
              <button
                onClick={handleBatchDelete}
                disabled={selectedPaths.size === 0 || deleting}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg disabled:opacity-40 transition-colors text-white"
                style={{ background: 'var(--color-error)' }}
              >
                <Trash2 size={14} /> 删除 {selectedPaths.size > 0 ? `(${selectedPaths.size})` : ''}
              </button>
              <button
                onClick={() => { setSelectMode(false); setSelectedPaths(new Set()) }}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
              >
                取消
              </button>
            </>
          ) : (
            pushSkills.length > 0 && (
              <button
                onClick={() => setSelectMode(true)}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
              >
                <CheckSquare size={14} /> 批量删除
              </button>
            )
          )}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-8">
          {loading ? (
            <div className="flex items-center justify-center h-32 text-sm" style={{ color: 'var(--text-muted)' }}>加载中...</div>
          ) : !tool ? (
            <div className="flex flex-col items-center justify-center h-48" style={{ color: 'var(--text-muted)' }}>
              <Wrench size={32} className="mb-2 opacity-30" />
              <p className="text-sm">请先在左侧选择一个工具</p>
            </div>
          ) : (
            <>
              {/* Push dir section */}
              <section>
                <div className="flex items-center gap-2 mb-4">
                  <ArrowUpToLine size={14} style={{ color: 'var(--color-success)' }} className="shrink-0" />
                  <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>推送路径</span>
                  {tool.pushDir
                    ? <span className="text-xs truncate" style={{ color: 'var(--text-muted)' }} title={tool.pushDir}>{tool.pushDir}</span>
                    : <span className="text-xs" style={{ color: 'var(--text-disabled)' }}>未配置</span>
                  }
                </div>
                {!tool.pushDir ? (
                  <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>该工具未配置推送路径</p>
                ) : pushSkills.length === 0 ? (
                  <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>推送路径下暂无 Skill</p>
                ) : (
                  <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
                    {pushSkills.map((sk: any) => (
                      <ToolSkillCard
                        key={sk.path}
                        name={sk.name}
                        path={sk.path}
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
                  <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>扫描路径</span>
                  {tool.scanDirs?.length > 0 && (
                    <span className="text-xs truncate" style={{ color: 'var(--text-muted)' }} title={tool.scanDirs.join(', ')}>
                      {tool.scanDirs.length} 个目录
                    </span>
                  )}
                </div>
                {scanOnlySkills.length === 0 ? (
                  <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>扫描路径下暂无独立 Skill</p>
                ) : (
                  <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
                    {scanOnlySkills.map((sk: any) => (
                      <ToolSkillCard
                        key={sk.path}
                        name={sk.name}
                        path={sk.path}
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
  canDelete: boolean
  selectMode: boolean
  selected: boolean
  onToggleSelect: () => void
  onDelete: () => void
}

function ToolSkillCard({ name, path, canDelete, selectMode, selected, onToggleSelect, onDelete }: ToolSkillCardProps) {
  const cardRef = useRef<HTMLDivElement>(null)
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const [hoveredRect, setHoveredRect] = useState<DOMRect | null>(null)
  const [meta, setMeta] = useState<any | null>(null)
  const [copied, setCopied] = useState(false)

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
      await navigator.clipboard.writeText(content)
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
              title="复制 skill.md"
              className="p-1 rounded transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-overlay)'; e.currentTarget.style.color = 'var(--text-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
            >
              {copied ? <Check size={13} style={{ color: 'var(--color-success)' }} /> : <Copy size={13} />}
            </button>
            <button
              onClick={handleOpen}
              title="打开目录"
              className="p-1 rounded transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-overlay)'; e.currentTarget.style.color = 'var(--text-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
            >
              <FolderOpenDot size={13} />
            </button>
          </div>
        )}

        <p
          className={`font-medium text-sm truncate mt-1 ${selectMode && canDelete ? 'pl-5' : ''} ${!selectMode ? 'pr-5' : ''}`}
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
              删除
            </button>
          </div>
        )}

        {/* Read-only badge */}
        {!canDelete && (
          <div className="mt-3 flex">
            <span className="text-xs ml-auto" style={{ color: 'var(--text-disabled)' }}>只读</span>
          </div>
        )}
      </div>

      {hoveredRect && (
        <SkillTooltip skill={{ Name: name, Source: undefined }} meta={meta} anchorRect={hoveredRect} />
      )}
    </>
  )
}
