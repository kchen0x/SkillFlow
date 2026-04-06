import { useEffect, useState } from 'react'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { Cloud, Download, RefreshCw, Upload } from 'lucide-react'
import { useLanguage } from '../contexts/LanguageContext'
import { subscribeToEvents } from '../lib/wailsEvents'
import { BackupNow, GetConfig, GetLastBackupChanges, GetLastBackupCompletedAt, RestoreFromCloud } from '../lib/backend'

type BackupChange = {
  path: string
  size: number
  action: string
}

function normalizeChanges(input: any): BackupChange[] {
  return (input ?? [])
    .map((item: any) => {
      const path = item?.path ?? item?.Path ?? ''
      const rawSize = item?.size ?? item?.Size ?? 0
      const size = typeof rawSize === 'number' ? rawSize : Number(rawSize) || 0
      const action = item?.action ?? item?.Action ?? ''
      return { path, size, action }
    })
    .filter((item: BackupChange) => item.path !== '')
}

function parseCompletedPayload(data: string): { files: BackupChange[]; completedAt: string } {
  try {
    const payload = JSON.parse(data)
    return {
      files: normalizeChanges(payload?.files ?? payload?.Files ?? []),
      completedAt: payload?.completedAt ?? payload?.CompletedAt ?? '',
    }
  } catch {
    return { files: [], completedAt: '' }
  }
}

function actionTone(action: string) {
  if (action === 'deleted') {
    return {
      color: 'var(--color-error)',
      background: 'rgba(248,113,113,0.12)',
    }
  }
  if (action === 'added') {
    return {
      color: 'var(--color-success)',
      background: 'rgba(74,222,128,0.12)',
    }
  }
  return {
    color: 'var(--accent-primary)',
    background: 'rgba(99,102,241,0.12)',
  }
}

export default function Backup() {
  const { t, lang } = useLanguage()
  const [files, setFiles] = useState<BackupChange[]>([])
  const [resultStatus, setResultStatus] = useState<'idle' | 'done' | 'error'>('idle')
  const [pushing, setPushing] = useState(false)
  const [pulling, setPulling] = useState(false)
  const [currentFile, setCurrentFile] = useState('')
  const [cloudEnabled, setCloudEnabled] = useState(false)
  const [isGit, setIsGit] = useState(false)
  const [lastCompletedAt, setLastCompletedAt] = useState('')
  const [loading, setLoading] = useState(true)

  const formatCompletedAt = (value: string) => {
    if (!value) return ''
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return ''
    return new Intl.DateTimeFormat(lang === 'zh' ? 'zh-CN' : 'en-US', {
      dateStyle: 'medium',
      timeStyle: 'medium',
    }).format(date)
  }

  const loadLastChanges = async () => {
    const [result, completedAt] = await Promise.all([GetLastBackupChanges(), GetLastBackupCompletedAt()])
    setFiles(normalizeChanges(result))
    setLastCompletedAt(completedAt ?? '')
  }

  useEffect(() => {
    let active = true

    Promise.all([GetConfig(), GetLastBackupChanges(), GetLastBackupCompletedAt()])
      .then(([cfg, changes, completedAt]) => {
        if (!active) return
        setCloudEnabled(cfg?.cloud?.enabled ?? false)
        setIsGit(cfg?.cloud?.provider === 'git')
        setFiles(normalizeChanges(changes))
        setLastCompletedAt(completedAt ?? '')
      })
      .finally(() => {
        if (active) setLoading(false)
      })

    const cleanup = subscribeToEvents(EventsOn, [
      ['backup.progress', (data: string) => {
        try { setCurrentFile(JSON.parse(data).currentFile ?? '') } catch {}
      }],
      ['backup.completed', (data: string) => {
        const payload = parseCompletedPayload(data)
        setResultStatus('done')
        setFiles(payload.files)
        setLastCompletedAt(payload.completedAt)
        setCurrentFile('')
      }],
      ['backup.failed', () => setResultStatus('error')],
      ['git.sync.completed', (data: string) => {
        const payload = parseCompletedPayload(data)
        setResultStatus('done')
        setFiles(payload.files)
        setLastCompletedAt(payload.completedAt)
        setCurrentFile('')
      }],
      ['git.sync.failed', () => setResultStatus('error')],
    ])

    return () => {
      active = false
      cleanup()
    }
  }, [])

  return (
    <div className="p-8 max-w-2xl">
      <h2 className="text-lg font-semibold mb-6 flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
        <Cloud size={18} /> {t('backup.title')}
      </h2>

      {loading && (
        <p className="mb-6 text-sm" style={{ color: 'var(--text-muted)' }}>
          {t('common.loading')}
        </p>
      )}

      {!loading && !cloudEnabled && (
        <div
          className="rounded-xl p-4 mb-6 text-sm"
          style={{
            background: 'rgba(251,191,36,0.1)',
            border: '1px solid rgba(251,191,36,0.3)',
            color: 'var(--color-warning)',
          }}
        >
          {t('backup.notEnabled')}
        </div>
      )}

      <div className="flex gap-3 mb-8">
        <button
          onClick={async () => {
            setPushing(true)
            setResultStatus('idle')
            setCurrentFile('')
            try {
              await BackupNow()
            } catch {
              setResultStatus('error')
            } finally {
              setPushing(false)
            }
          }}
          disabled={!cloudEnabled || pushing || pulling}
          className="btn-primary flex items-center gap-2 px-5 py-2.5 rounded-lg text-sm"
        >
          {pushing ? <RefreshCw size={14} className="animate-spin" /> : <Upload size={14} />}
          {pushing ? t('backup.backingUp', { file: currentFile }) : t('backup.backupNow')}
        </button>
        <button
          onClick={async () => {
            setPulling(true)
            setResultStatus('idle')
            try {
              await RestoreFromCloud()
              await loadLastChanges()
              setResultStatus('done')
            } catch {
              setResultStatus('error')
            } finally {
              setPulling(false)
            }
          }}
          disabled={!cloudEnabled || pushing || pulling}
          className="btn-secondary flex items-center gap-2 px-5 py-2.5 rounded-lg text-sm"
        >
          {pulling ? <RefreshCw size={14} className="animate-spin" /> : <Download size={14} />}
          {pulling ? t('backup.pulling') : (isGit ? t('backup.pullRemote') : t('backup.restore'))}
        </button>
        <button
          onClick={loadLastChanges}
          className="flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm transition-colors"
          style={{ color: 'var(--text-muted)' }}
          onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
          onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
        >
          <RefreshCw size={14} /> {t('backup.refresh')}
        </button>
      </div>

      {resultStatus === 'done' && (
        <p className="mb-4 text-sm" style={{ color: 'var(--color-success)' }}>
          {isGit ? t('backup.gitDone') : t('backup.done')}
        </p>
      )}
      {resultStatus === 'error' && (
        <p className="mb-4 text-sm" style={{ color: 'var(--color-error)' }}>
          {isGit ? t('backup.gitFailed') : t('backup.failed')}
        </p>
      )}

      {lastCompletedAt && (
        <p className="mb-4 text-sm" style={{ color: 'var(--text-muted)' }}>
          {t('backup.lastSyncAt', { time: formatCompletedAt(lastCompletedAt) })}
        </p>
      )}

      {files.length > 0 && (
        <div>
          <p className="text-sm mb-3" style={{ color: 'var(--text-muted)' }}>
            {isGit ? t('backup.gitChanges') : t('backup.cloudChanges')} ({files.length})
          </p>
          <div
            className="max-h-96 overflow-y-auto rounded-xl divide-y"
            style={{ border: '1px solid var(--border-base)', borderColor: 'var(--border-base)' }}
          >
            {files.map((f, i) => {
              const tone = actionTone(f.action)
              return (
                <div
                  key={`${f.path}-${f.action}-${i}`}
                  className="flex items-center justify-between gap-3 px-4 py-2.5 text-sm"
                  style={{ borderColor: 'var(--border-base)' }}
                >
                  <div className="flex min-w-0 items-center gap-2">
                    <span
                      className="rounded-full px-2 py-0.5 text-[10px] uppercase tracking-[0.08em]"
                      style={tone}
                    >
                      {t(`backup.action.${f.action || 'modified'}` as any)}
                    </span>
                    <span className="min-w-0 truncate font-mono text-xs" style={{ color: 'var(--text-secondary)' }}>
                      {f.path}
                    </span>
                  </div>
                  <span className="shrink-0 text-xs" style={{ color: 'var(--text-muted)' }}>
                    {f.action === 'deleted' ? t('backup.deleted') : `${(f.size / 1024).toFixed(1)} KB`}
                  </span>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {resultStatus === 'done' && files.length === 0 && (
        <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
          {t('backup.noChanges')}
        </p>
      )}
    </div>
  )
}
