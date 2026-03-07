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

  return (
    <div className="flex h-full overflow-hidden">
      {/* Left: tool list */}
      <div className="w-48 shrink-0 border-r border-gray-800 p-3 flex flex-col gap-0.5 overflow-y-auto">
        <div className="px-3 py-1.5 text-xs font-medium tracking-wide text-gray-500 uppercase">
          工具列表
        </div>
        {tools.map(t => (
          <button
            key={t.name}
            onClick={() => selectTool(t.name)}
            className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-left transition-colors ${
              selectedTool === t.name
                ? 'bg-indigo-600 text-white'
                : 'text-gray-400 hover:bg-gray-800 hover:text-white'
            }`}
          >
            <ToolIcon name={t.name} size={20} />
            <span className="truncate">{t.name}</span>
          </button>
        ))}
        {tools.length === 0 && (
          <p className="px-3 text-xs text-gray-600 mt-2">没有启用的工具，请在设置中启用</p>
        )}
      </div>

      {/* Main content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Toolbar */}
        <div className="flex items-center gap-3 px-6 py-4 border-b border-gray-800 flex-wrap">
          {tool ? (
            <div className="flex items-center gap-2">
              <ToolIcon name={tool.name} size={22} />
              <span className="font-medium text-sm">{tool.name}</span>
            </div>
          ) : (
            <h2 className="text-sm font-medium flex items-center gap-2">
              <Wrench size={14} /> 我的工具
            </h2>
          )}
          <div className="flex-1" />
          {selectMode ? (
            <>
              <button
                onClick={toggleSelectAll}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800"
              >
                <CheckSquare size={14} />{allSelected ? '取消全选' : '全选'}
              </button>
              <button
                onClick={handleBatchDelete}
                disabled={selectedPaths.size === 0 || deleting}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-red-700 hover:bg-red-600 disabled:opacity-40 rounded-lg"
              >
                <Trash2 size={14} /> 删除 {selectedPaths.size > 0 ? `(${selectedPaths.size})` : ''}
              </button>
              <button
                onClick={() => { setSelectMode(false); setSelectedPaths(new Set()) }}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800"
              >
                取消
              </button>
            </>
          ) : (
            pushSkills.length > 0 && (
              <button
                onClick={() => setSelectMode(true)}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800"
              >
                <CheckSquare size={14} /> 批量删除
              </button>
            )
          )}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-8">
          {loading ? (
            <div className="flex items-center justify-center h-32 text-gray-500 text-sm">加载中...</div>
          ) : !tool ? (
            <div className="flex flex-col items-center justify-center h-48 text-gray-500">
              <Wrench size={32} className="mb-2 opacity-30" />
              <p className="text-sm">请先在左侧选择一个工具</p>
            </div>
          ) : (
            <>
              {/* Push dir section */}
              <section>
                <div className="flex items-center gap-2 mb-4">
                  <ArrowUpToLine size={14} className="text-emerald-400 shrink-0" />
                  <span className="text-sm font-medium text-gray-200">推送路径</span>
                  {tool.pushDir
                    ? <span className="text-xs text-gray-500 truncate" title={tool.pushDir}>{tool.pushDir}</span>
                    : <span className="text-xs text-gray-600">未配置</span>
                  }
                </div>
                {!tool.pushDir ? (
                  <p className="text-sm text-gray-600 pl-5">该工具未配置推送路径</p>
                ) : pushSkills.length === 0 ? (
                  <p className="text-sm text-gray-600 pl-5">推送路径下暂无 Skill</p>
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
                  <ScanLine size={14} className="text-indigo-400 shrink-0" />
                  <span className="text-sm font-medium text-gray-200">扫描路径</span>
                  {tool.scanDirs?.length > 0 && (
                    <span className="text-xs text-gray-500 truncate" title={tool.scanDirs.join(', ')}>
                      {tool.scanDirs.length} 个目录
                    </span>
                  )}
                </div>
                {scanOnlySkills.length === 0 ? (
                  <p className="text-sm text-gray-600 pl-5">扫描路径下暂无独立 Skill</p>
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
        className={`relative border rounded-xl transition-all duration-150 group p-4 ${
          selectMode && canDelete ? 'cursor-pointer' : 'cursor-default'
        } ${
          selected
            ? 'bg-gray-800 border-indigo-500 bg-indigo-900/20'
            : 'bg-gray-800 border-gray-700 hover:border-indigo-500'
        }`}
      >
        {/* Select checkbox */}
        {selectMode && canDelete && (
          <div className="absolute top-2 left-2 z-10">
            <div className={`w-4 h-4 rounded border-2 flex items-center justify-center ${
              selected ? 'bg-indigo-500 border-indigo-500' : 'border-gray-500 bg-gray-700'
            }`}>
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
            <button onClick={handleCopy} title="复制 skill.md"
              className="p-1 rounded text-gray-600 hover:text-gray-200 hover:bg-gray-700 transition-colors">
              {copied ? <Check size={13} className="text-green-400" /> : <Copy size={13} />}
            </button>
            <button onClick={handleOpen} title="打开目录"
              className="p-1 rounded text-gray-600 hover:text-gray-200 hover:bg-gray-700 transition-colors">
              <FolderOpenDot size={13} />
            </button>
          </div>
        )}

        <p className={`font-medium text-sm truncate mt-1 ${selectMode && canDelete ? 'pl-5' : ''} ${!selectMode ? 'pr-5' : ''}`}>
          {name}
        </p>

        {/* Delete button in hover actions (non-select mode) */}
        {!selectMode && canDelete && (
          <div className="mt-3 flex opacity-0 group-hover:opacity-100 transition-opacity">
            <button
              onClick={e => { e.stopPropagation(); onDelete() }}
              className="text-xs text-red-400 hover:text-red-300 ml-auto"
            >
              删除
            </button>
          </div>
        )}

        {/* Read-only badge for scan-only skills */}
        {!canDelete && (
          <div className="mt-3 flex">
            <span className="text-xs text-gray-600 ml-auto">只读</span>
          </div>
        )}
      </div>

      {hoveredRect && (
        <SkillTooltip skill={{ Name: name, Source: undefined }} meta={meta} anchorRect={hoveredRect} />
      )}
    </>
  )
}
