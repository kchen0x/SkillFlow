import { createPortal } from 'react-dom'
import { Github, FolderOpen, Tag, Wrench, GitBranch, Calendar, Clock, Hash, ExternalLink } from 'lucide-react'
import { useLanguage } from '../contexts/LanguageContext'

export interface SkillInfo {
  Name: string
  Category?: string
  Source?: string
  SourceURL?: string
  SourceSubPath?: string
  SourceSHA?: string
  LatestSHA?: string
  InstalledAt?: string
  UpdatedAt?: string
}

interface SkillMeta {
  Name: string
  Description: string
  ArgumentHint: string
  AllowedTools: string
  Context: string
  DisableModelInvocation: boolean
}

interface Props {
  skill: SkillInfo
  meta: SkillMeta | null
  anchorRect: DOMRect
}

function fmt(dateStr: string | undefined, locale: string): string {
  if (!dateStr) return '—'
  try {
    return new Date(dateStr).toLocaleDateString(locale, { year: 'numeric', month: '2-digit', day: '2-digit' })
  } catch {
    return '—'
  }
}

function shortSHA(sha: string): string {
  return sha ? sha.slice(0, 7) : '—'
}

export default function SkillTooltip({ skill, meta, anchorRect }: Props) {
  const { t, lang } = useLanguage()
  const locale = lang === 'zh' ? 'zh-CN' : 'en-US'
  const sourceLabel = skill.Source === 'github'
    ? t('common.sourceGitHub')
    : skill.Source === 'manual'
      ? t('common.sourceManual')
      : skill.Source === 'git'
        ? t('common.sourceGit')
        : skill.Source

  const TOOLTIP_WIDTH = 300
  const TOOLTIP_MAX_HEIGHT = 400
  const GAP = 8

  let left = anchorRect.right + GAP
  if (left + TOOLTIP_WIDTH > window.innerWidth - 8) {
    left = anchorRect.left - TOOLTIP_WIDTH - GAP
  }

  let top = anchorRect.top
  if (top + TOOLTIP_MAX_HEIGHT > window.innerHeight - 8) {
    top = window.innerHeight - TOOLTIP_MAX_HEIGHT - 8
  }

  const displayName = meta?.Name || skill.Name
  const isGitHub = skill.Source === 'github'

  const tooltip = (
    <div
      style={{
        left, top, width: TOOLTIP_WIDTH, maxHeight: TOOLTIP_MAX_HEIGHT,
        background: 'var(--bg-elevated)',
        border: '1px solid var(--border-accent)',
        boxShadow: 'var(--shadow-dialog), var(--glow-accent-sm)',
      }}
      className="fixed z-50 overflow-y-auto rounded-xl text-sm pointer-events-none"
    >
      <div className="px-4 pt-4 pb-3" style={{ borderBottom: '1px solid var(--border-base)' }}>
        <div className="flex items-start gap-2">
          <div className="mt-0.5 shrink-0" style={{ color: 'var(--text-muted)' }}>
            {isGitHub ? <Github size={14} /> : <FolderOpen size={14} />}
          </div>
          <div className="min-w-0 flex-1">
            <p className="font-semibold leading-snug truncate" style={{ color: 'var(--text-primary)' }}>{displayName}</p>
            <div className="flex items-center gap-1.5 mt-1">
              {sourceLabel && (
                <span
                  className="text-xs px-1.5 py-0.5 rounded font-medium"
                  style={isGitHub ? {
                    background: 'rgba(14, 165, 233, 0.15)',
                    color: 'var(--accent-secondary)',
                  } : {
                    background: 'var(--bg-overlay)',
                    color: 'var(--text-muted)',
                  }}
                >
                  {sourceLabel}
                </span>
              )}
              {skill.Category && (
                <span className="text-xs truncate" style={{ color: 'var(--text-muted)' }}>{skill.Category}</span>
              )}
            </div>
          </div>
        </div>

        {meta === null ? (
          <p className="mt-3 text-xs italic" style={{ color: 'var(--text-muted)' }}>{t('skillTooltip.loading')}</p>
        ) : meta.Description ? (
          <p className="mt-3 text-xs leading-relaxed" style={{ color: 'var(--text-secondary)' }}>{meta.Description}</p>
        ) : (
          <p className="mt-3 text-xs italic" style={{ color: 'var(--text-disabled)' }}>{t('skillTooltip.noDescription')}</p>
        )}
      </div>

      {meta && (meta.ArgumentHint || meta.AllowedTools || meta.Context) && (
        <div className="px-4 py-3 space-y-2" style={{ borderBottom: '1px solid var(--border-base)' }}>
          {meta.ArgumentHint && (
            <Row icon={<Tag size={12} />} label={t('skillTooltip.argumentHint')}>
              <code
                className="text-xs px-1.5 py-0.5 rounded font-mono"
                style={{ background: 'var(--bg-surface)', color: 'var(--accent-primary)' }}
              >
                {meta.ArgumentHint}
              </code>
            </Row>
          )}
          {meta.AllowedTools && (
            <Row icon={<Wrench size={12} />} label={t('skillTooltip.allowedTools')}>
              <span className="text-xs" style={{ color: 'var(--text-secondary)' }}>{meta.AllowedTools}</span>
            </Row>
          )}
          {meta.Context && (
            <Row icon={<GitBranch size={12} />} label={t('skillTooltip.context')}>
              <span className="text-xs font-mono" style={{ color: 'var(--accent-primary)' }}>{meta.Context}</span>
            </Row>
          )}
        </div>
      )}

      {(isGitHub && skill.SourceURL || skill.SourceSHA || skill.InstalledAt) && (
        <div className="px-4 py-3 space-y-2">
          {isGitHub && skill.SourceURL && (
            <Row icon={<ExternalLink size={12} />} label={t('skillTooltip.repo')}>
              <span className="text-xs truncate max-w-[160px]" style={{ color: 'var(--accent-secondary)' }}>
                {skill.SourceURL.replace('https://github.com/', '')}
                {skill.SourceSubPath ? `/${skill.SourceSubPath}` : ''}
              </span>
            </Row>
          )}
          {skill.SourceSHA && (
            <Row icon={<Hash size={12} />} label={t('skillTooltip.version')}>
              <code className="text-xs font-mono" style={{ color: 'var(--text-secondary)' }}>{shortSHA(skill.SourceSHA)}</code>
              {skill.LatestSHA && skill.LatestSHA !== skill.SourceSHA && (
                <span className="ml-2 text-xs" style={{ color: 'var(--color-warning)' }}>{t('skillTooltip.updatable')} → {shortSHA(skill.LatestSHA)}</span>
              )}
            </Row>
          )}
          {skill.InstalledAt && (
            <Row icon={<Calendar size={12} />} label={t('skillTooltip.installedAt')}>
              <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{fmt(skill.InstalledAt, locale)}</span>
            </Row>
          )}
          {skill.UpdatedAt && skill.UpdatedAt !== skill.InstalledAt && (
            <Row icon={<Clock size={12} />} label={t('skillTooltip.updatedAt')}>
              <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{fmt(skill.UpdatedAt, locale)}</span>
            </Row>
          )}
        </div>
      )}
    </div>
  )

  return createPortal(tooltip, document.body)
}

function Row({ icon, label, children }: { icon: React.ReactNode; label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-start gap-2">
      <span className="mt-0.5 shrink-0" style={{ color: 'var(--text-muted)' }}>{icon}</span>
      <span className="shrink-0 w-16 text-xs leading-relaxed" style={{ color: 'var(--text-muted)' }}>{label}</span>
      <div className="flex items-center gap-1 min-w-0 flex-wrap">{children}</div>
    </div>
  )
}
