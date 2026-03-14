import { type CSSProperties } from 'react'
import { ToolIcon } from '../config/toolIcons'
import { useLanguage } from '../contexts/LanguageContext'
import { summarizePushedAgents } from '../lib/skillStatusStrip'

type BadgeTone = 'accent' | 'success' | 'warning' | 'muted' | 'info'

export type SkillStatusBadge = {
  key: string
  label: string
  tone: BadgeTone
}

interface Props {
  badges?: SkillStatusBadge[]
  pushedAgents?: string[]
  maxVisiblePushedAgents?: number
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
  pushedAgents = [],
  maxVisiblePushedAgents = 3,
  className = '',
  singleLine = true,
}: Props) {
  const { t } = useLanguage()
  const { visibleAgents, overflowCount } = summarizePushedAgents(pushedAgents, maxVisiblePushedAgents)

  return (
    <div
      className={`flex min-w-0 min-h-[1.75rem] gap-1.5 ${
        singleLine ? 'flex-nowrap items-center overflow-hidden' : 'flex-wrap content-start items-start overflow-visible'
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

      {pushedAgents.length > 0 && (
        <span
          className="inline-flex shrink-0 items-center gap-1 rounded-full px-1.5 py-0.5 text-[11px] leading-4"
          style={toneStyles.success}
          title={t('common.pushedToAgents', { agents: pushedAgents.join(', ') })}
        >
          <span className="flex min-w-0 max-w-[8rem] items-center gap-1 overflow-hidden">
            {visibleAgents.map((agentName) => (
              <ToolIcon key={agentName} name={agentName} size={16} />
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
