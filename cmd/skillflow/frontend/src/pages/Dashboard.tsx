import { useEffect, useRef, useState, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  ListSkills, ListCategories, MoveSkillCategory,
  DeleteSkill, DeleteSkills, ImportLocal, UpdateSkill, CheckUpdates,
  OpenFolderDialog, GetSkillMeta,
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import CategoryPanel from '../components/CategoryPanel'
import SkillCard from '../components/SkillCard'
import SkillTooltip from '../components/SkillTooltip'
import GitHubInstallDialog from '../components/GitHubInstallDialog'
import { Github, FolderOpen, RefreshCw, Search, Trash2, CheckSquare } from 'lucide-react'
import { gridContainerVariants, cardVariants } from '../lib/motionVariants'

export default function Dashboard() {
  const [skills, setSkills] = useState<any[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [selectedCat, setSelectedCat] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [showGitHub, setShowGitHub] = useState(false)
  const [dragOver, setDragOver] = useState(false)
  const [draggingSkillID, setDraggingSkillID] = useState<string | null>(null)
  const [categoryDragActive, setCategoryDragActive] = useState(false)
  const [selectMode, setSelectMode] = useState(false)
  const [selectedIDs, setSelectedIDs] = useState<Set<string>>(new Set())

  // Hover tooltip state
  const [hoveredSkill, setHoveredSkill] = useState<{ skill: any; rect: DOMRect } | null>(null)
  const [hoveredMeta, setHoveredMeta] = useState<any | null>(null)
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  const load = useCallback(async () => {
    const [s, c] = await Promise.all([ListSkills(), ListCategories()])
    setSkills(s ?? [])
    setCategories(c ?? [])
  }, [])

  useEffect(() => {
    load()
    EventsOn('update.available', load)
  }, [load])

  const filtered = skills.filter(sk => {
    const matchCat = selectedCat === null || sk.Category === selectedCat
    const matchSearch = !search || sk.Name.toLowerCase().includes(search.toLowerCase())
    return matchCat && matchSearch
  })

  const skillCounts = skills.reduce((acc, sk) => {
    const category = sk.Category || 'Default'
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

  const handleImportButton = async () => {
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
    if (selectedIDs.size === filtered.length) {
      setSelectedIDs(new Set())
    } else {
      setSelectedIDs(new Set(filtered.map(sk => sk.ID)))
    }
  }

  const handleBatchDelete = async () => {
    if (selectedIDs.size === 0) return
    await DeleteSkills(Array.from(selectedIDs))
    setSelectedIDs(new Set())
    setSelectMode(false)
    load()
  }

  const allSelected = filtered.length > 0 && selectedIDs.size === filtered.length

  const clearHover = () => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current)
    setHoveredSkill(null)
    setHoveredMeta(null)
  }

  const handleHoverStart = (sk: any, rect: DOMRect) => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current)
    hoverTimer.current = setTimeout(async () => {
      setHoveredSkill({ skill: sk, rect })
      setHoveredMeta(null)
      const meta = await GetSkillMeta(sk.ID)
      setHoveredMeta(meta)
    }, 300)
  }

  const handleHoverEnd = () => {
    clearHover()
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
          style={{ background: 'var(--accent-glow)' }}>
          <p className="text-lg font-medium" style={{ color: 'var(--accent-primary)' }}>松开以导入 Skill</p>
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
          <div className="relative flex-1 min-w-[260px] max-w-[520px]">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2" style={{ color: 'var(--text-muted)' }} />
            <input
              value={search} onChange={e => setSearch(e.target.value)}
              placeholder="搜索 Skills..."
              className="input-base pl-10"
            />
          </div>

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
                {allSelected ? '取消全选' : '全选'}
              </button>
              <button
                onClick={handleBatchDelete}
                disabled={selectedIDs.size === 0}
                className="flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                style={{ backgroundColor: 'var(--color-error)', color: 'white' }}
              >
                <Trash2 size={14} /> 删除 {selectedIDs.size > 0 ? `(${selectedIDs.size})` : ''}
              </button>
              <button
                onClick={toggleSelectMode}
                className="flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
              >
                取消
              </button>
            </div>
          ) : (
            <div className="flex flex-wrap items-center gap-2 min-w-0">
              {[
                { icon: <RefreshCw size={14} />, label: '更新', onClick: () => CheckUpdates() },
                { icon: <CheckSquare size={14} />, label: '批删', onClick: toggleSelectMode },
                { icon: <FolderOpen size={14} />, label: '导入', onClick: handleImportButton },
              ].map(btn => (
                <button
                  key={btn.label}
                  onClick={btn.onClick}
                  className="flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg whitespace-nowrap transition-colors"
                  style={{ color: 'var(--text-muted)' }}
                  onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                  onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
                >
                  {btn.icon} {btn.label}
                </button>
              ))}
              <button
                onClick={() => setShowGitHub(true)}
                className="btn-primary flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg whitespace-nowrap"
              >
                <Github size={14} /> 远程安装
              </button>
            </div>
          )}
        </div>

        {/* Skills grid */}
        <div className="flex-1 overflow-y-auto p-6">
          <motion.div
            className="grid grid-cols-3 xl:grid-cols-4 gap-4"
            variants={containerVariants}
            initial="initial"
            animate="animate"
          >
            {filtered.map(sk => (
              <motion.div key={sk.ID} variants={filtered.length <= 30 ? cardVariants : undefined}>
                <SkillCard
                  skill={{ id: sk.ID, name: sk.Name, category: sk.Category, source: sk.Source, hasUpdate: !!sk.LatestSHA, path: sk.Path }}
                  categories={categories}
                  onDelete={async () => { await DeleteSkill(sk.ID); load() }}
                  onUpdate={async () => { await UpdateSkill(sk.ID); load() }}
                  onMoveCategory={async cat => { await MoveSkillCategory(sk.ID, cat); load() }}
                  dragging={draggingSkillID === sk.ID}
                  dropTargetActive={draggingSkillID === sk.ID && categoryDragActive}
                  onDragStateChange={(dragging) => {
                    setDraggingSkillID(dragging ? sk.ID : null)
                    if (!dragging) setCategoryDragActive(false)
                  }}
                  selectMode={selectMode}
                  selected={selectedIDs.has(sk.ID)}
                  onToggleSelect={() => toggleSelectID(sk.ID)}
                  onHoverStart={rect => handleHoverStart(sk, rect)}
                  onHoverEnd={handleHoverEnd}
                />
              </motion.div>
            ))}
          </motion.div>
          {filtered.length === 0 && (
            <div className="flex flex-col items-center justify-center h-48" style={{ color: 'var(--text-muted)' }}>
              <p className="text-sm">没有找到 Skills</p>
              <p className="text-xs mt-1">从远程仓库安装或拖拽文件夹到此处</p>
            </div>
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

      <AnimatePresence>
        {showGitHub && (
          <GitHubInstallDialog onClose={() => setShowGitHub(false)} onDone={() => { setShowGitHub(false); load() }} />
        )}
      </AnimatePresence>
    </div>
  )
}
