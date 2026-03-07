import { useState } from 'react'
import { AlertCircle, Plus } from 'lucide-react'
import ContextMenu from './ContextMenu'
import { CreateCategory, RenameCategory, DeleteCategory } from '../../wailsjs/go/main/App'
import AnimatedDialog from './ui/AnimatedDialog'

interface Props {
  categories: string[]
  skillCounts: Record<string, number>
  selected: string | null
  draggingSkillId: string | null
  onSelect: (cat: string | null) => void
  onCategoryDragStateChange: (active: boolean) => void
  onDrop: (skillId: string, category: string) => void
  onRefresh: () => void
}

const defaultCategoryName = 'Default'

export default function CategoryPanel({
  categories, skillCounts, selected, draggingSkillId, onSelect, onCategoryDragStateChange, onDrop, onRefresh,
}: Props) {
  const [menu, setMenu] = useState<{ x: number; y: number; cat: string } | null>(null)
  const [renaming, setRenaming] = useState<string | null>(null)
  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)
  const [createName, setCreateName] = useState('')
  const [dragTarget, setDragTarget] = useState<string | null>(null)
  const [deleteBlocked, setDeleteBlocked] = useState<{ cat: string; count: number } | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const acceptsSkillDrop = draggingSkillId !== null

  const handleDeleteCategory = async (cat: string) => {
    const count = skillCounts[cat] ?? 0
    if (count > 0) {
      setDeleteBlocked({ cat, count })
      return
    }
    try {
      await DeleteCategory(cat)
      if (selected === cat) onSelect(null)
      onRefresh()
    } catch (e: any) {
      const msg = String(e?.message ?? e ?? '删除分类失败')
      if (msg.includes('先清空')) {
        setDeleteBlocked({ cat, count })
        return
      }
      setDeleteError(msg)
    }
  }

  const handleDrop = (e: React.DragEvent, cat: string) => {
    e.preventDefault()
    setDragTarget(null)
    onCategoryDragStateChange(false)
    const id = e.dataTransfer.getData('application/x-skillflow-skill-id') || e.dataTransfer.getData('text/plain') || draggingSkillId || ''
    if (id) onDrop(id, cat)
  }

  const handleDragOver = (e: React.DragEvent, cat: string) => {
    if (!acceptsSkillDrop) return
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    setDragTarget(cat)
    onCategoryDragStateChange(true)
  }

  const handleDragLeave = (cat: string) => {
    if (dragTarget === cat) {
      setDragTarget(null)
      onCategoryDragStateChange(false)
    }
  }

  const getCategoryStyle = (isActive: boolean, isDragTarget: boolean) => {
    if (isDragTarget) return {
      background: 'var(--accent-glow)',
      color: 'var(--accent-primary)',
      border: '1px solid var(--border-accent)',
      boxShadow: 'var(--glow-accent-sm)',
    }
    if (isActive) return {
      background: 'var(--accent-glow)',
      color: 'var(--accent-primary)',
      border: '1px solid var(--border-accent)',
      boxShadow: 'var(--glow-accent-sm)',
    }
    return {
      color: 'var(--text-muted)',
      border: '1px solid transparent',
    }
  }

  return (
    <div
      className="w-48 flex-shrink-0 p-3 flex flex-col gap-0.5"
      style={{ borderRight: '1px solid var(--border-base)' }}
    >
      {/* All */}
      <div
        onClick={() => onSelect(null)}
        onDragEnter={e => handleDragOver(e, '')}
        onDragOver={e => handleDragOver(e, '')}
        onDragLeave={() => handleDragLeave('')}
        onDrop={e => handleDrop(e, '')}
        className="px-3 py-2 rounded-lg text-sm cursor-pointer transition-all duration-150"
        style={getCategoryStyle(selected === null, dragTarget === '')}
        onMouseEnter={e => {
          if (selected !== null && dragTarget !== '') {
            e.currentTarget.style.backgroundColor = 'var(--bg-hover)'
            e.currentTarget.style.color = 'var(--text-primary)'
          }
        }}
        onMouseLeave={e => {
          if (selected !== null && dragTarget !== '') {
            e.currentTarget.style.backgroundColor = ''
            e.currentTarget.style.color = 'var(--text-muted)'
          }
        }}
      >全部</div>

      {/* Categories */}
      {categories.map(cat => (
        renaming === cat
          ? <input
              key={cat} autoFocus value={newName}
              onChange={e => setNewName(e.target.value)}
              onBlur={async () => {
                if (newName && newName !== cat) { await RenameCategory(cat, newName); onRefresh() }
                setRenaming(null)
              }}
              onKeyDown={async e => {
                if (e.key === 'Enter') { await RenameCategory(cat, newName); onRefresh(); setRenaming(null) }
                if (e.key === 'Escape') setRenaming(null)
              }}
              className="input-base px-3 py-1.5 rounded-lg text-sm w-full"
            />
          : <div
              key={cat}
              onClick={() => onSelect(cat)}
              onDragEnter={e => handleDragOver(e, cat)}
              onDragOver={e => handleDragOver(e, cat)}
              onDragLeave={() => handleDragLeave(cat)}
              onDrop={e => handleDrop(e, cat)}
              onContextMenu={e => {
                e.preventDefault()
                if (cat === defaultCategoryName) return
                setMenu({ x: e.clientX, y: e.clientY, cat })
              }}
              className="px-3 py-2 rounded-lg text-sm cursor-pointer transition-all duration-150"
              style={getCategoryStyle(selected === cat, dragTarget === cat)}
              onMouseEnter={e => {
                if (selected !== cat && dragTarget !== cat) {
                  e.currentTarget.style.backgroundColor = 'var(--bg-hover)'
                  e.currentTarget.style.color = 'var(--text-primary)'
                }
              }}
              onMouseLeave={e => {
                if (selected !== cat && dragTarget !== cat) {
                  e.currentTarget.style.backgroundColor = ''
                  e.currentTarget.style.color = 'var(--text-muted)'
                }
              }}
            >{cat}</div>
      ))}

      {/* New category input */}
      {creating
        ? <input
            autoFocus value={createName}
            onChange={e => setCreateName(e.target.value)}
            onBlur={async () => {
              if (createName) { await CreateCategory(createName); onRefresh() }
              setCreating(false); setCreateName('')
            }}
            onKeyDown={async e => {
              if (e.key === 'Enter') { await CreateCategory(createName); onRefresh(); setCreating(false); setCreateName('') }
              if (e.key === 'Escape') { setCreating(false); setCreateName('') }
            }}
            className="input-base px-3 py-1.5 rounded-lg text-sm w-full"
          />
        : <button
            onClick={() => setCreating(true)}
            className="flex items-center gap-1.5 px-3 py-2 text-sm mt-1 transition-colors"
            style={{ color: 'var(--text-muted)' }}
            onMouseEnter={e => { e.currentTarget.style.color = 'var(--text-primary)' }}
            onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)' }}
          >
            <Plus size={14} /> 新建分类
          </button>
      }

      {/* Context menu */}
      {menu && (
        <ContextMenu
          x={menu.x} y={menu.y}
          items={[
            { label: '重命名', onClick: () => { setRenaming(menu.cat); setNewName(menu.cat) } },
            { label: '删除', onClick: async () => { await handleDeleteCategory(menu.cat) }, danger: true },
          ]}
          onClose={() => setMenu(null)}
        />
      )}

      <AnimatedDialog open={deleteBlocked !== null} onClose={() => setDeleteBlocked(null)} width="w-[420px]" zIndex={50}>
        <h3 className="font-semibold mb-2 flex items-center gap-2" style={{ color: 'var(--color-warning)' }}>
          <AlertCircle size={16} /> 无法删除分类
        </h3>
        <p className="text-sm mb-4" style={{ color: 'var(--text-secondary)' }}>
          分类「<span className="font-medium" style={{ color: 'var(--text-primary)' }}>{deleteBlocked?.cat}</span>」下
          {(deleteBlocked?.count ?? 0) > 0 ? `还有 ${deleteBlocked?.count} 个 Skill，` : '还有 Skill，'}
          请先清空该分类后再删除。
        </p>
        <button onClick={() => setDeleteBlocked(null)} className="btn-secondary w-full py-2 rounded-lg text-sm">
          我知道了
        </button>
      </AnimatedDialog>

      <AnimatedDialog open={deleteError !== null} onClose={() => setDeleteError(null)} width="w-[420px]" zIndex={50}>
        <h3 className="font-semibold mb-2 flex items-center gap-2" style={{ color: 'var(--color-error)' }}>
          <AlertCircle size={16} /> 删除失败
        </h3>
        <p className="text-sm mb-4" style={{ color: 'var(--text-secondary)' }}>{deleteError}</p>
        <button onClick={() => setDeleteError(null)} className="btn-secondary w-full py-2 rounded-lg text-sm">
          关闭
        </button>
      </AnimatedDialog>
    </div>
  )
}
