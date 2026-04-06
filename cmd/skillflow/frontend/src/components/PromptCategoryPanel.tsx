import { useState } from 'react'
import { AlertCircle, Plus } from 'lucide-react'
import ContextMenu from './ContextMenu'
import { CreatePromptCategory, RenamePromptCategory, DeletePromptCategory } from '../lib/backend'
import { useLanguage } from '../contexts/LanguageContext'
import AnimatedDialog from './ui/AnimatedDialog'

interface Props {
  categories: string[]
  promptCounts: Record<string, number>
  selected: string | null
  draggingPromptName: string | null
  onSelect: (cat: string | null) => void
  onCategoryDragStateChange: (active: boolean) => void
  onDrop: (promptName: string, category: string) => void
  onRefresh: () => void
}

const defaultCategoryName = 'Default'

export default function PromptCategoryPanel({
  categories, promptCounts, selected, draggingPromptName, onSelect, onCategoryDragStateChange, onDrop, onRefresh,
}: Props) {
  const { t } = useLanguage()
  const [menu, setMenu] = useState<{ x: number; y: number; cat: string } | null>(null)
  const [renaming, setRenaming] = useState<string | null>(null)
  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)
  const [createName, setCreateName] = useState('')
  const [dragTarget, setDragTarget] = useState<string | null>(null)
  const [deleteBlocked, setDeleteBlocked] = useState<{ cat: string; count: number } | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const acceptsPromptDrop = draggingPromptName !== null

  const handleDeleteCategory = async (cat: string) => {
    const count = promptCounts[cat] ?? 0
    if (count > 0) {
      setDeleteBlocked({ cat, count })
      return
    }
    try {
      await DeletePromptCategory(cat)
      if (selected === cat) onSelect(null)
      onRefresh()
    } catch (error: any) {
      const msg = String(error?.message ?? error ?? t('promptCategory.deleteFailed'))
      if (msg.includes('不可删除') || msg.includes('not empty') || msg.includes('先清空')) {
        setDeleteBlocked({ cat, count })
        return
      }
      setDeleteError(msg)
    }
  }

  const handleDrop = (event: React.DragEvent, cat: string) => {
    event.preventDefault()
    setDragTarget(null)
    onCategoryDragStateChange(false)
    const name = event.dataTransfer.getData('application/x-skillflow-prompt-name') || event.dataTransfer.getData('text/plain') || draggingPromptName || ''
    if (name) onDrop(name, cat)
  }

  const handleDragOver = (event: React.DragEvent, cat: string) => {
    if (!acceptsPromptDrop) return
    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'
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
    if (isDragTarget || isActive) {
      return {
        background: 'var(--active-surface)',
        color: 'var(--active-text)',
        border: '1px solid var(--active-border)',
        boxShadow: 'var(--active-shadow)',
      }
    }
    return {
      color: 'var(--text-muted)',
      border: '1px solid transparent',
    }
  }

  return (
    <div className="w-48 flex-shrink-0 p-3 flex flex-col gap-0.5" style={{ borderRight: '1px solid var(--border-base)' }}>
      <div
        onClick={() => onSelect(null)}
        onDragEnter={event => handleDragOver(event, '')}
        onDragOver={event => handleDragOver(event, '')}
        onDragLeave={() => handleDragLeave('')}
        onDrop={event => handleDrop(event, '')}
        className="px-3 py-2 rounded-lg text-sm cursor-pointer transition-all duration-150"
        style={getCategoryStyle(selected === null, dragTarget === '')}
      >
        {t('category.all')}
      </div>

      {categories.map((cat) => (
        renaming === cat ? (
          <input
            key={cat}
            autoFocus
            value={newName}
            onChange={event => setNewName(event.target.value)}
            onBlur={async () => {
              if (newName && newName !== cat) {
                await RenamePromptCategory(cat, newName)
                onRefresh()
              }
              setRenaming(null)
            }}
            onKeyDown={async (event) => {
              if (event.key === 'Enter') {
                await RenamePromptCategory(cat, newName)
                onRefresh()
                setRenaming(null)
              }
              if (event.key === 'Escape') setRenaming(null)
            }}
            className="input-base px-3 py-1.5 rounded-lg text-sm w-full"
          />
        ) : (
          <div
            key={cat}
            onClick={() => onSelect(cat)}
            onDragEnter={event => handleDragOver(event, cat)}
            onDragOver={event => handleDragOver(event, cat)}
            onDragLeave={() => handleDragLeave(cat)}
            onDrop={event => handleDrop(event, cat)}
            onContextMenu={(event) => {
              event.preventDefault()
              if (cat === defaultCategoryName) return
              setMenu({ x: event.clientX, y: event.clientY, cat })
            }}
            className="px-3 py-2 rounded-lg text-sm cursor-pointer transition-all duration-150"
            style={getCategoryStyle(selected === cat, dragTarget === cat)}
          >
            {cat}
          </div>
        )
      ))}

      {creating ? (
        <input
          autoFocus
          value={createName}
          onChange={event => setCreateName(event.target.value)}
          onBlur={async () => {
            if (createName) {
              await CreatePromptCategory(createName)
              onRefresh()
            }
            setCreating(false)
            setCreateName('')
          }}
          onKeyDown={async (event) => {
            if (event.key === 'Enter') {
              await CreatePromptCategory(createName)
              onRefresh()
              setCreating(false)
              setCreateName('')
            }
            if (event.key === 'Escape') {
              setCreating(false)
              setCreateName('')
            }
          }}
          className="input-base px-3 py-1.5 rounded-lg text-sm w-full"
        />
      ) : (
        <button
          onClick={() => setCreating(true)}
          className="flex items-center gap-1.5 px-3 py-2 text-sm mt-1 transition-colors"
          style={{ color: 'var(--text-muted)' }}
        >
          <Plus size={14} /> {t('category.newCategory')}
        </button>
      )}

      {menu && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          items={[
            { label: t('category.rename'), onClick: () => { setRenaming(menu.cat); setNewName(menu.cat) } },
            { label: t('category.delete'), onClick: async () => { await handleDeleteCategory(menu.cat) }, danger: true },
          ]}
          onClose={() => setMenu(null)}
        />
      )}

      <AnimatedDialog open={deleteBlocked !== null} onClose={() => setDeleteBlocked(null)} width="w-[420px]" zIndex={50}>
        <h3 className="font-semibold mb-2 flex items-center gap-2" style={{ color: 'var(--color-warning)' }}>
          <AlertCircle size={16} /> {t('promptCategory.cannotDelete')}
        </h3>
        <p className="text-sm mb-4" style={{ color: 'var(--text-secondary)' }}>
          {(deleteBlocked?.count ?? 0) > 0
            ? t('promptCategory.hasPrompts', { cat: deleteBlocked?.cat ?? '', count: deleteBlocked?.count ?? 0 })
            : t('promptCategory.hasPromptsGeneric', { cat: deleteBlocked?.cat ?? '' })}
        </p>
        <button onClick={() => setDeleteBlocked(null)} className="btn-secondary w-full py-2 rounded-lg text-sm">
          {t('common.gotIt')}
        </button>
      </AnimatedDialog>

      <AnimatedDialog open={deleteError !== null} onClose={() => setDeleteError(null)} width="w-[420px]" zIndex={50}>
        <h3 className="font-semibold mb-2 flex items-center gap-2" style={{ color: 'var(--color-error)' }}>
          <AlertCircle size={16} /> {t('promptCategory.deleteFailed')}
        </h3>
        <p className="text-sm mb-4" style={{ color: 'var(--text-secondary)' }}>{deleteError}</p>
        <button onClick={() => setDeleteError(null)} className="btn-secondary w-full py-2 rounded-lg text-sm">
          {t('common.close')}
        </button>
      </AnimatedDialog>
    </div>
  )
}
