import { useRef, useState } from 'react'
import { Github, FolderOpen, RefreshCw, FolderOpenDot, Copy, Check } from 'lucide-react'
import ContextMenu from './ContextMenu'
import { OpenPath, ReadSkillFileContent } from '../../wailsjs/go/main/App'
import { useLanguage } from '../contexts/LanguageContext'

interface Skill { id: string; name: string; category: string; source: 'github' | 'manual'; hasUpdate: boolean; path?: string }
interface Props {
  skill: Skill
  categories: string[]
  onDelete: () => void
  onUpdate?: () => void
  onMoveCategory: (category: string) => void
  dragging?: boolean
  dropTargetActive?: boolean
  onDragStateChange?: (dragging: boolean) => void
  selectMode?: boolean
  selected?: boolean
  onToggleSelect?: () => void
  onHoverStart?: (rect: DOMRect) => void
  onHoverEnd?: () => void
}

export default function SkillCard({
  skill, categories, onDelete, onUpdate, onMoveCategory,
  dragging = false, dropTargetActive = false, onDragStateChange,
  selectMode, selected, onToggleSelect,
  onHoverStart, onHoverEnd,
}: Props) {
  const { t } = useLanguage()
  const [menu, setMenu] = useState<{ x: number; y: number } | null>(null)
  const [copied, setCopied] = useState(false)
  const cardRef = useRef<HTMLDivElement>(null)
  const dragGhostRef = useRef<HTMLDivElement | null>(null)

  const sourceLabel = skill.source === 'github' ? t('common.sourceGitHub') : t('common.sourceManual')

  const setCardDragImage = (e: React.DragEvent) => {
    if (!cardRef.current) return
    const clone = cardRef.current.cloneNode(true) as HTMLDivElement
    const rect = cardRef.current.getBoundingClientRect()
    clone.style.width = `${Math.max(rect.width * 0.82, 180)}px`
    clone.style.transform = 'scale(0.82)'
    clone.style.transformOrigin = 'top left'
    clone.style.opacity = '0.96'
    clone.style.pointerEvents = 'none'
    clone.style.position = 'fixed'
    clone.style.top = '-1000px'
    clone.style.left = '-1000px'
    clone.style.zIndex = '9999'
    document.body.appendChild(clone)
    dragGhostRef.current = clone
    e.dataTransfer.setDragImage(clone, 24, 18)
  }

  const cleanupDragGhost = () => {
    if (dragGhostRef.current?.parentNode) {
      dragGhostRef.current.parentNode.removeChild(dragGhostRef.current)
    }
    dragGhostRef.current = null
  }

  const handleContextMenu = (e: React.MouseEvent) => {
    if (selectMode) return
    e.preventDefault()
    setMenu({ x: e.clientX, y: e.clientY })
  }

  const handleClick = () => {
    if (selectMode) onToggleSelect?.()
  }

  const handleMouseEnter = () => {
    if (selectMode) return
    if (cardRef.current) onHoverStart?.(cardRef.current.getBoundingClientRect())
  }

  const handleMouseLeave = () => {
    onHoverEnd?.()
  }

  const handleOpenFolder = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (skill.path) OpenPath(skill.path)
  }

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation()
    if (!skill.path) return
    try {
      const content = await ReadSkillFileContent(skill.path)
      await navigator.clipboard.writeText(content)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch { /* ignore */ }
  }

  const menuItems = [
    ...(skill.hasUpdate ? [{ label: t('skillCard.update'), onClick: () => onUpdate?.() }] : []),
    ...categories.filter(c => c !== skill.category).map(c => ({
      label: t('skillCard.moveTo', { cat: c }),
      onClick: () => onMoveCategory(c),
    })),
    { label: t('skillCard.delete'), onClick: onDelete, danger: true },
  ]

  if (dragging && dropTargetActive) {
    return (
      <div className="relative min-h-[88px] rounded-xl border border-transparent bg-transparent">
        <div className="absolute inset-x-4 top-1/2 -translate-y-1/2 h-[2px] rounded-full"
          style={{ background: 'var(--accent-primary)', boxShadow: 'var(--glow-accent-sm)' }} />
      </div>
    )
  }

  return (
    <>
      <div
        ref={cardRef}
        draggable={!selectMode}
        onDragStart={e => {
          if (selectMode) return
          e.dataTransfer.setData('text/plain', skill.id)
          e.dataTransfer.setData('application/x-skillflow-skill-id', skill.id)
          e.dataTransfer.effectAllowed = 'move'
          setCardDragImage(e)
          onDragStateChange?.(true)
        }}
        onDragEnd={() => {
          cleanupDragGhost()
          onDragStateChange?.(false)
        }}
        onContextMenu={handleContextMenu}
        onClick={handleClick}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        className={`card-base relative p-4 ${selectMode ? 'cursor-pointer' : 'cursor-grab'} group ${
          dragging ? 'opacity-55 scale-[0.96]' : ''
        }`}
        style={selected ? {
          background: 'var(--accent-glow)',
          borderColor: 'var(--border-accent)',
          boxShadow: 'var(--glow-accent-sm)',
        } : undefined}
      >
        {selectMode && (
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

        {!selectMode && skill.path && (
          <button
            onClick={handleOpenFolder}
            title={t('skillCard.openDir')}
            className="absolute top-2 right-2 z-10 p-1 rounded opacity-0 group-hover:opacity-100 transition-all"
            style={{ color: 'var(--text-muted)' }}
            onMouseEnter={e => { e.currentTarget.style.background = 'var(--bg-overlay)'; e.currentTarget.style.color = 'var(--text-primary)' }}
            onMouseLeave={e => { e.currentTarget.style.background = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
          >
            <FolderOpenDot size={14} />
          </button>
        )}

        {skill.hasUpdate && !selectMode && (
          <span className="absolute top-2 right-8 w-2.5 h-2.5 rounded-full" style={{ background: 'var(--color-error)' }} />
        )}
        {skill.hasUpdate && selectMode && (
          <span className="absolute top-2 right-2 w-2.5 h-2.5 rounded-full" style={{ background: 'var(--color-error)' }} />
        )}

        <div className={`flex items-center gap-2 mb-2 ${selectMode ? 'pl-5' : ''}`}>
          {skill.source === 'github'
            ? <Github size={14} style={{ color: 'var(--text-muted)' }} />
            : <FolderOpen size={14} style={{ color: 'var(--text-muted)' }} />}
          <span
            className="text-xs px-1.5 py-0.5 rounded"
            style={skill.source === 'github' ? {
              background: 'rgba(14, 165, 233, 0.15)',
              color: 'var(--accent-secondary)',
              border: '1px solid rgba(14, 165, 233, 0.25)',
            } : {
              color: 'var(--text-muted)',
            }}
          >
            {sourceLabel}
          </span>
        </div>
        <p
          className={`font-medium text-sm truncate ${selectMode ? 'pl-5' : 'pr-5'}`}
          style={{ color: 'var(--text-primary)' }}
        >
          {skill.name}
        </p>
        {!selectMode && (
          <div className="mt-3 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
            {skill.hasUpdate && (
              <button
                onClick={e => { e.stopPropagation(); onUpdate?.() }}
                className="text-xs flex items-center gap-1 transition-colors"
                style={{ color: 'var(--accent-primary)' }}
              >
                <RefreshCw size={12} /> {t('skillCard.update')}
              </button>
            )}
            {skill.path && (
              <button
                onClick={handleCopy}
                className="text-xs flex items-center gap-1 transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.color = 'var(--text-primary)' }}
                onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)' }}
              >
                {copied
                  ? <><Check size={12} style={{ color: 'var(--color-success)' }} /> {t('skillCard.copied')}</>
                  : <><Copy size={12} /> {t('skillCard.copy')}</>}
              </button>
            )}
            <button
              onClick={e => { e.stopPropagation(); onDelete() }}
              className="text-xs ml-auto transition-colors"
              style={{ color: 'var(--color-error)' }}
            >
              {t('skillCard.delete')}
            </button>
          </div>
        )}
      </div>
      {menu && (
        <ContextMenu x={menu.x} y={menu.y} items={menuItems} onClose={() => setMenu(null)} />
      )}
    </>
  )
}
