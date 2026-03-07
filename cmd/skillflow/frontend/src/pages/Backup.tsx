import { useEffect, useState } from 'react'
import { BackupNow, ListCloudFiles, RestoreFromCloud, GetConfig } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { Cloud, Upload, Download, RefreshCw } from 'lucide-react'
import { useLanguage } from '../contexts/LanguageContext'

export default function Backup() {
  const { t } = useLanguage()
  const [files, setFiles] = useState<Array<{ path: string; size: number }>>([])
  const [resultStatus, setResultStatus] = useState<'idle' | 'done' | 'error'>('idle')
  const [pushing, setPushing] = useState(false)
  const [pulling, setPulling] = useState(false)
  const [currentFile, setCurrentFile] = useState('')
  const [cloudEnabled, setCloudEnabled] = useState(false)
  const [isGit, setIsGit] = useState(false)

  useEffect(() => {
    GetConfig().then(cfg => {
      setCloudEnabled(cfg?.cloud?.enabled ?? false)
      setIsGit(cfg?.cloud?.provider === 'git')
    })

    EventsOn('backup.progress', (data: string) => {
      try { setCurrentFile(JSON.parse(data).currentFile ?? '') } catch {}
    })
    EventsOn('backup.completed', () => { setResultStatus('done'); loadFiles() })
    EventsOn('backup.failed', () => setResultStatus('error'))
    EventsOn('git.sync.completed', () => { setResultStatus('done'); loadFiles() })
    EventsOn('git.sync.failed', () => setResultStatus('error'))
  }, [])

  const loadFiles = async () => {
    const f = await ListCloudFiles()
    const normalized = (f ?? [])
      .map((item: any) => {
        const path = item?.path ?? item?.Path ?? ''
        const rawSize = item?.size ?? item?.Size ?? 0
        const size = typeof rawSize === 'number' ? rawSize : Number(rawSize) || 0
        return { path, size }
      })
      .filter((item: { path: string }) => item.path !== '')
    setFiles(normalized)
  }

  return (
    <div className="p-8 max-w-2xl">
      <h2 className="text-lg font-semibold mb-6 flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
        <Cloud size={18} /> {t('backup.title')}
      </h2>

      {!cloudEnabled && (
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
              loadFiles()
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
          onClick={loadFiles}
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

      {files.length > 0 && (
        <div>
          <p className="text-sm mb-3" style={{ color: 'var(--text-muted)' }}>
            {(isGit ? t('backup.gitFiles') : t('backup.cloudFiles'))}（{files.length}）
          </p>
          <div
            className="max-h-96 overflow-y-auto rounded-xl divide-y"
            style={{ border: '1px solid var(--border-base)', borderColor: 'var(--border-base)' }}
          >
            {files.map((f, i) => (
              <div
                key={i}
                className="flex items-center justify-between px-4 py-2.5 text-sm"
                style={{ borderColor: 'var(--border-base)' }}
              >
                <span className="font-mono text-xs" style={{ color: 'var(--text-secondary)' }}>{f.path}</span>
                <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{(f.size / 1024).toFixed(1)} KB</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
