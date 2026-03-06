import { useEffect, useState } from 'react'
import { BackupNow, ListCloudFiles, RestoreFromCloud, GetConfig } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { Cloud, Upload, Download, RefreshCw } from 'lucide-react'

export default function Backup() {
  const [files, setFiles] = useState<Array<{ path: string; size: number }>>([])
  const [status, setStatus] = useState<'idle' | 'backing-up' | 'done' | 'error'>('idle')
  const [currentFile, setCurrentFile] = useState('')
  const [cloudEnabled, setCloudEnabled] = useState(false)
  const [isGit, setIsGit] = useState(false)

  useEffect(() => {
    GetConfig().then(cfg => {
      setCloudEnabled(cfg?.cloud?.enabled ?? false)
      setIsGit(cfg?.cloud?.provider === 'git')
    })

    EventsOn('backup.started', () => setStatus('backing-up'))
    EventsOn('backup.progress', (data: string) => {
      try { setCurrentFile(JSON.parse(data).currentFile ?? '') } catch {}
    })
    EventsOn('backup.completed', () => { setStatus('done'); loadFiles() })
    EventsOn('backup.failed', () => setStatus('error'))
    EventsOn('git.sync.started', () => setStatus('backing-up'))
    EventsOn('git.sync.completed', () => { setStatus('done'); loadFiles() })
    EventsOn('git.sync.failed', () => setStatus('error'))
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
      <h2 className="text-lg font-semibold mb-6 flex items-center gap-2"><Cloud size={18} /> 云备份</h2>

      {!cloudEnabled && (
        <div className="bg-yellow-900/30 border border-yellow-700/50 rounded-xl p-4 mb-6 text-sm text-yellow-300">
          云备份未启用。请前往设置 → 云存储完成配置。
        </div>
      )}

      <div className="flex gap-3 mb-8">
        <button
          onClick={async () => {
            try {
              setStatus('backing-up')
              await BackupNow()
            } catch {
              setStatus('error')
            }
          }}
          disabled={!cloudEnabled || status === 'backing-up'}
          className="flex items-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50"
        >
          {status === 'backing-up' ? <RefreshCw size={14} className="animate-spin" /> : <Upload size={14} />}
          {status === 'backing-up' ? `备份中 ${currentFile}` : '立即备份'}
        </button>
        <button
          onClick={async () => {
            try {
              setStatus('backing-up')
              await RestoreFromCloud()
              loadFiles()
              if (!isGit) setStatus('done')
            } catch {
              setStatus('error')
            }
          }}
          disabled={!cloudEnabled}
          className="flex items-center gap-2 px-5 py-2.5 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm disabled:opacity-50"
        ><Download size={14} /> {isGit ? '拉取远端' : '从云端恢复'}</button>
        <button onClick={loadFiles} className="flex items-center gap-2 px-4 py-2.5 text-gray-400 hover:text-white rounded-lg hover:bg-gray-800 text-sm">
          <RefreshCw size={14} /> 刷新
        </button>
      </div>

      {status === 'done' && <p className="mb-4 text-sm text-green-400">{isGit ? 'Git 同步完成' : '备份完成'}</p>}
      {status === 'error' && <p className="mb-4 text-sm text-red-400">{isGit ? 'Git 同步失败，请检查仓库配置' : '备份失败，请检查云存储配置'}</p>}

      {files.length > 0 && (
        <div>
          <p className="text-sm text-gray-400 mb-3">{isGit ? 'Git 跟踪文件' : '云端文件'}（{files.length} 个）</p>
          <div className="max-h-96 overflow-y-auto border border-gray-800 rounded-xl divide-y divide-gray-800">
            {files.map((f, i) => (
              <div key={i} className="flex items-center justify-between px-4 py-2.5 text-sm">
                <span className="text-gray-300 font-mono text-xs">{f.path}</span>
                <span className="text-gray-500 text-xs">{(f.size / 1024).toFixed(1)} KB</span>
              </div>
            ))}
          </div>
        </div>
      )}

    </div>
  )
}
