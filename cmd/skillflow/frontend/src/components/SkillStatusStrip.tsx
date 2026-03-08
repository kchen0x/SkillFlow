import { useEffect, useRef, useState, type CSSProperties } from 'react'
import { ToolIcon } from '../config/toolIcons'
import { useLanguage } from '../contexts/LanguageContext'

type BadgeTone = 'accent' | 'success' | 'warning' | 'muted' | 'info'

export type SkillStatusBadge = {
  key: string
  label: string
  tone: BadgeTone
}

interface Props {
  badges?: SkillStatusBadge[]
  pushedTools?: string[]
  maxVisiblePushedTools?: number
  className?: string
  singleLine?: boolean
}

const toneStyles: Record<BadgeTone, CSSProperties> = {
  accent: {
    background: 'rgba(14, 165, 233, 0.15)',
    color: 'var(--accent-secondary)',
    border: '1px solid rgba(14, 165, 233, 0.25)',
  },
  success: {
    background: 'rgba(52, 211, 153, 0.15)',
    color: 'var(--color-success)',
    border: '1px solid rgba(52, 211, 153, 0.25)',
  },
  warning: {
    background: 'rgba(251, 191, 36, 0.16)',
    color: 'var(--color-warning)',
    border: '1px solid rgba(251, 191, 36, 0.28)',
  },
  muted: {
    background: 'var(--bg-overlay)',
    color: 'var(--text-muted)',
    border: '1px solid var(--border-base)',
  },
  info: {
    background: 'rgba(59, 130, 246, 0.12)',
    color: 'var(--accent-primary)',
    border: '1px solid rgba(59, 130, 246, 0.22)',
  },
}

export default function SkillStatusStrip({
  badges = [],
  pushedTools = [],
  maxVisiblePushedTools = 3,
  className = '',
  singleLine = true,
}: Props) {
  const { t } = useLanguage()
  const containerRef = useRef<HTMLDivElement>(null)
  const [shouldWrap, setShouldWrap] = useState(false)
  const visibleTools = pushedTools.slice(0, maxVisiblePushedTools)
  const overflowCount = pushedTools.length - visibleTools.length
  const effectiveSingleLine = singleLine && !shouldWrap

  useEffect(() => {
    if (!singleLine) {
      setShouldWrap(false)
      return
    }

    const el = containerRef.current
    if (!el) return

    let frame = 0
    let observer: ResizeObserver | null = null

    const measure = () => {
      frame = 0
      const styles = window.getComputedStyle(el)
      const gap = Number.parseFloat(styles.columnGap || styles.gap || '0') || 0
      const children = Array.from(el.children) as HTMLElement[]
      const totalWidth = children.reduce((sum, child) => sum + Math.max(child.getBoundingClientRect().width, child.scrollWidth), 0)
        + Math.max(children.length - 1, 0) * gap
      const hasTruncatedLabel = children.some((child) => {
        const label = child.querySelector('[data-status-label]') as HTMLElement | null
        return !!label && label.scrollWidth > label.clientWidth + 1
      })
      const nextShouldWrap = totalWidth > el.clientWidth + 1 || hasTruncatedLabel
      setShouldWrap(prev => (prev === nextShouldWrap ? prev : nextShouldWrap))
    }

    const scheduleMeasure = () => {
      if (frame) cancelAnimationFrame(frame)
      frame = requestAnimationFrame(measure)
    }

    scheduleMeasure()
    observer = new ResizeObserver(scheduleMeasure)
    observer.observe(el)
    Array.from(el.children).forEach((child) => observer?.observe(child))

    return () => {
      if (frame) cancelAnimationFrame(frame)
      observer?.disconnect()
    }
  }, [badges, pushedTools, maxVisiblePushedTools, singleLine])

  return (
    <div
      ref={containerRef}
      className={`flex min-w-0 min-h-[1.75rem] gap-1.5 ${
        effectiveSingleLine ? 'flex-nowrap items-center overflow-hidden' : 'flex-wrap content-start items-start overflow-visible'
      } ${className}`}
    >
      {badges.map((badge) => (
        <span
          key={badge.key}
          className="inline-flex max-w-full shrink-0 items-center rounded-full px-1.5 py-0.5 text-[11px] leading-4"
          style={toneStyles[badge.tone]}
          title={badge.label}
        >
          <span data-status-label className="max-w-[9rem] truncate">{badge.label}</span>
        </span>
      ))}

      {pushedTools.length > 0 && (
        <span
          className="inline-flex shrink-0 items-center gap-1 rounded-full px-1.5 py-0.5 text-[11px] leading-4"
          style={toneStyles.success}
          title={t('common.pushedToTools', { tools: pushedTools.join(', ') })}
        >
          <span className="flex min-w-0 items-center gap-1 overflow-hidden">
            {visibleTools.map((toolName) => (
              <ToolIcon key={toolName} name={toolName} size={16} />
            ))}
            {overflowCount > 0 && (
              <span
                className="inline-flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-semibold"
                style={{ background: 'rgba(255,255,255,0.4)', color: 'currentColor' }}
              >
                +{overflowCount}
              </span>
            )}
          </span>
        </span>
      )}
    </div>
  )
}
