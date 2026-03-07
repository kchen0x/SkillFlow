import { useState } from 'react'
import { ScanGitHub, InstallFromGitHub, ListCategories } from '../../wailsjs/go/main/App'
import { Github, X, AlertCircle } from 'lucide-react'
import AnimatedDialog from './ui/AnimatedDialog'

interface Props { onClose: () => void; onDone: () => void }

export default function GitHubInstallDialog({ onClose, onDone }: Props) {
  const [url, setUrl] = useState('')
  const [candidates, setCandidates] = useState<any[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [categories, setCategories] = useState<string[]>([])
  const [category, setCategory] = useState('')
  const [scanning, setScanning] = useState(false)
  const [installing, setInstalling] = useState(false)
  const [scanError, setScanError] = useState('')
  const [installError, setInstallError] = useState('')
  const [scannedOnce, setScannedOnce] = useState(false)

  const scan = async () => {
    setScanning(true)
    setScanError('')
    setCandidates([])
    setScannedOnce(false)
    try {
      const [c, cats] = await Promise.all([ScanGitHub(url), ListCategories()])
      const skills = c ?? []
      const catList = cats ?? []
      setCandidates(skills)
      setCategories(catList)
      if (catList.length > 0 && category === '') setCategory(catList[0])
      setSelected(new Set(skills.filter((x: any) => !x.Installed).map((x: any) => x.Name)))
      setScannedOnce(true)
      if (skills.length === 0) setScanError('未发现任何 Skill，请确认该仓库包含含有 skill.md的子目录')
    } catch (e: any) {
      setScanError(String(e?.message ?? e ?? '扫描失败，请检查网络或仓库地址'))
    } finally {
      setScanning(false)
    }
  }

  const install = async () => {
    setInstalling(true)
    setInstallError('')
    try {
      const toInstall = candidates.filter(c => selected.has(c.Name))
      await InstallFromGitHub(url, toInstall, category)
      onDone()
    } catch (e: any) {
      setInstallError(String(e?.message ?? e ?? '安装失败'))
    } finally {
      setInstalling(false)
    }
  }

  const toggle = (name: string) => {
    const next = new Set(selected)
    next.has(name) ? next.delete(name) : next.add(name)
    setSelected(next)
  }

  return (
    <AnimatedDialog open={true} onClose={onClose} width="w-[520px]">
      <div className="flex justify-between items-center mb-4">
        <h3 className="font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
          <Github size={16} /> 从远程仓库安装
        </h3>
        <button onClick={onClose} style={{ color: 'var(--text-muted)' }} className="hover:opacity-80 transition-opacity">
          <X size={16} />
        </button>
      </div>

      <div className="flex gap-2 mb-4">
        <input
          value={url}
          onChange={e => setUrl(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && !scanning && url && scan()}
          placeholder="https://host/owner/repo.git 或 git@host:owner/repo.git"
          className="input-base flex-1"
        />
        <button
          onClick={scan}
          disabled={scanning || !url}
          className="btn-primary px-4 py-2 rounded-lg text-sm min-w-[72px]"
        >
          {scanning ? (
            <span className="flex items-center gap-1.5">
              <span className="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin inline-block" />
              克隆/更新中...
            </span>
          ) : '扫描'}
        </button>
      </div>
      <p className="text-xs mb-3" style={{ color: 'var(--text-muted)' }}>首次扫描会 git clone 仓库，后续自动 git pull 更新</p>

      {scanError && (
        <div
          className="flex items-start gap-2 rounded-lg px-4 py-3 text-sm mb-4"
          style={{ background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)', color: 'var(--color-error)' }}
        >
          <AlertCircle size={15} className="mt-0.5 shrink-0" />
          <span className="flex-1">{scanError}</span>
        </div>
      )}

      {candidates.length > 0 && (
        <>
          <div
            className="max-h-52 overflow-y-auto space-y-1 mb-4 rounded-lg p-1"
            style={{ background: 'var(--bg-surface)' }}
          >
            {candidates.map(c => (
              <label
                key={c.Name}
                className="flex items-center gap-3 px-3 py-2 rounded-lg cursor-pointer transition-colors"
                style={{ color: 'var(--text-primary)' }}
                onMouseEnter={e => { e.currentTarget.style.background = 'var(--bg-hover)' }}
                onMouseLeave={e => { e.currentTarget.style.background = '' }}
              >
                <input
                  type="checkbox"
                  checked={selected.has(c.Name)}
                  onChange={() => toggle(c.Name)}
                  style={{ accentColor: 'var(--accent-secondary)' }}
                />
                <span className="text-sm flex-1">{c.Name}</span>
                {c.Installed && (
                  <span
                    className="text-xs px-2 py-0.5 rounded"
                    style={{
                      background: 'rgba(14, 165, 233, 0.15)',
                      color: 'var(--accent-secondary)',
                      border: '1px solid rgba(14, 165, 233, 0.25)',
                    }}
                  >
                    已安装
                  </span>
                )}
              </label>
            ))}
          </div>
          <div className="flex items-center gap-3 mb-4">
            <span className="text-sm" style={{ color: 'var(--text-muted)' }}>安装到分类</span>
            <select
              value={category}
              onChange={e => setCategory(e.target.value)}
              className="select-base flex-1"
            >
              {categories.map(c => <option key={c} value={c}>{c}</option>)}
            </select>
          </div>
          {installError && (
            <div
              className="flex items-start gap-2 rounded-lg px-4 py-3 text-sm mb-3"
              style={{ background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)', color: 'var(--color-error)' }}
            >
              <AlertCircle size={15} className="mt-0.5 shrink-0" />
              <span>{installError}</span>
            </div>
          )}
          <button
            onClick={install}
            disabled={installing || selected.size === 0}
            className="btn-primary w-full py-2 rounded-lg text-sm"
          >
            {installing ? '安装中...' : `安装 ${selected.size} 个 Skill`}
          </button>
        </>
      )}
    </AnimatedDialog>
  )
}
