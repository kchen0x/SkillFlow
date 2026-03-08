import type { CSSProperties } from 'react'
import { ArrowUpFromLine } from 'lucide-react'
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
}: Props) {
  const { t } = useLanguage()
  const visibleTools = pushedTools.slice(0, maxVisiblePushedTools)
  const hasOverflow = pushedTools.length > visibleTools.length

  return (
    <div className={`flex min-h-[1.75rem] flex-wrap content-start items-start gap-1.5 overflow-hidden ${className}`}>
      {badges.map((badge) => (
        <span
          key={badge.key}
          className="inline-flex max-w-full items-center rounded-full px-1.5 py-0.5 text-[11px] leading-4"
          style={toneStyles[badge.tone]}
          title={badge.label}
        >
          <span className="truncate max-w-[9rem]">{badge.label}</span>
        </span>
      ))}

      {pushedTools.length > 0 && (
        <span
          className="inline-flex items-center gap-1 rounded-full px-1.5 py-0.5 text-[11px] leading-4"
          style={toneStyles.success}
          title={t('common.pushedToTools', { tools: pushedTools.join(', ') })}
        >
          <ArrowUpFromLine size={11} className="shrink-0" />
          <span className="flex items-center gap-1 overflow-hidden">
            {visibleTools.map((toolName) => (
              <ToolIcon key={toolName} name={toolName} size={16} />
            ))}
            {hasOverflow && (
              <span
                className="inline-flex h-4 w-4 items-center justify-center rounded-full text-[11px] font-semibold"
                style={{ background: 'rgba(255,255,255,0.4)', color: 'currentColor' }}
              >
                …
              </span>
            )}
          </span>
        </span>
      )}
    </div>
  )
}
