import { useRef, useState } from 'react'
import { FolderOpen, Github, FolderOpenDot, Copy, Check } from 'lucide-react'
import { OpenPath, GetSkillMeta, GetSkillMetaByPath, ReadSkillFileContent } from '../../wailsjs/go/main/App'
import { useLanguage } from '../contexts/LanguageContext'
import { copyTextToClipboard } from '../lib/clipboard'
import SkillTooltip from './SkillTooltip'
import SkillStatusStrip, { type SkillStatusBadge } from './SkillStatusStrip'

interface Props {
  name: string
  subtitle?: string
  source?: string
  path?: string
  id?: string
  showSelection?: boolean
  imported?: boolean
  updatable?: boolean
  pushedTools?: string[]
  showImported?: boolean
  showUpdatable?: boolean
  showPushedTools?: boolean
  selected: boolean
  onToggle: () => void
}

export default function SyncSkillCard({
  name, subtitle, source, path, id,
  showSelection = true,
  imported,
  updatable,
  pushedTools = [],
  showImported = false,
  showUpdatable = false,
  showPushedTools = false,
  selected,
  onToggle,
}: Props) {
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
        let m: any
        if (id) {
          m = await GetSkillMeta(id)
        } else if (path) {
          m = await GetSkillMetaByPath(path)
        }
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
    if (path) OpenPath(path)
  }

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation()
    if (!path) return
    try {
      const content = await ReadSkillFileContent(path)
      await copyTextToClipboard(content)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch { /* ignore */ }
  }

  const skillInfo = { Name: name, Source: source }
  const badges: SkillStatusBadge[] = [
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
  ]

  return (
    <>
      <div
        ref={cardRef}
        onClick={onToggle}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        className="card-base relative flex flex-col gap-2 p-3 cursor-pointer select-none group"
        style={selected ? {
          background: 'var(--accent-glow)',
          borderColor: 'var(--border-accent)',
          boxShadow: 'var(--glow-accent-sm)',
        } : undefined}
      >
        <div className="absolute top-2 right-2 flex items-center gap-0.5 z-10">
          {path && (
            <button
              onClick={handleCopy}
              title={t('toolSkills.copySkill')}
              className="p-1 rounded opacity-0 group-hover:opacity-100 transition-all"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => { e.currentTarget.style.background = 'var(--bg-overlay)'; e.currentTarget.style.color = 'var(--text-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.background = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
            >
              {copied
                ? <Check size={12} style={{ color: 'var(--color-success)' }} />
                : <Copy size={12} />}
            </button>
          )}
          {path && (
            <button
              onClick={handleOpen}
              title={t('toolSkills.openDir')}
              className="p-1 rounded opacity-0 group-hover:opacity-100 transition-all"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => { e.currentTarget.style.background = 'var(--bg-overlay)'; e.currentTarget.style.color = 'var(--text-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.background = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
            >
              <FolderOpenDot size={13} />
            </button>
          )}
        </div>

        <div className="flex items-start gap-1.5 pr-14">
          {source === 'github'
            ? <Github size={12} style={{ color: 'var(--text-muted)' }} className="mt-1 shrink-0" />
            : <FolderOpen size={12} style={{ color: 'var(--text-muted)' }} className="mt-1 shrink-0" />}
          <SkillStatusStrip
            badges={badges}
            pushedTools={showPushedTools ? pushedTools : []}
            className="flex-1 min-w-0"
          />
        </div>

        <p className="min-h-[2.75rem] pr-5 text-sm font-medium leading-snug line-clamp-2" style={{ color: 'var(--text-primary)' }}>{name}</p>

        {subtitle && (
          <p className="text-xs truncate" style={{ color: 'var(--text-muted)' }}>{subtitle}</p>
        )}

        {showSelection && (
          <div
            className="absolute bottom-2 right-2 w-4 h-4 rounded border-2 flex items-center justify-center transition-all duration-150"
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
        )}
      </div>

      {hoveredRect && (
        <SkillTooltip skill={skillInfo} meta={meta} anchorRect={hoveredRect} />
      )}
    </>
  )
}
