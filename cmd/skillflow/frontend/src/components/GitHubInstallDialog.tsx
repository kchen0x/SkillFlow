import { useState } from 'react'
import { ScanGitHub, InstallFromGitHub, ListCategories } from '../../wailsjs/go/main/App'
import { Github, X, AlertCircle } from 'lucide-react'
import { useLanguage } from '../contexts/LanguageContext'
import { useSkillStatusVisibility } from '../contexts/SkillStatusVisibilityContext'
import AnimatedDialog from './ui/AnimatedDialog'
import SkillStatusStrip from './SkillStatusStrip'

interface Props { onClose: () => void; onDone: () => void }

export default function GitHubInstallDialog({ onClose, onDone }: Props) {
  const { t } = useLanguage()
  const visibility = useSkillStatusVisibility('githubInstall')
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
      setSelected(new Set(skills.filter((x: any) => !x.Installed).map((x: any) => x.Path)))
      setScannedOnce(true)
      if (skills.length === 0) setScanError(t('github.noSkills'))
    } catch (e: any) {
      setScanError(String(e?.message ?? e ?? t('github.scanFailed')))
    } finally {
      setScanning(false)
    }
  }

  const install = async () => {
    setInstalling(true)
    setInstallError('')
    try {
      const toInstall = candidates.filter(c => selected.has(c.Path))
      await InstallFromGitHub(url, toInstall, category)
      onDone()
    } catch (e: any) {
      setInstallError(String(e?.message ?? e ?? t('github.installFailed')))
    } finally {
      setInstalling(false)
    }
  }

  const toggle = (path: string) => {
    const next = new Set(selected)
    next.has(path) ? next.delete(path) : next.add(path)
    setSelected(next)
  }

  return (
    <AnimatedDialog open={true} onClose={onClose} width="w-[520px]">
      <div className="flex justify-between items-center mb-4">
        <h3 className="font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
          <Github size={16} /> {t('github.title')}
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
          placeholder={t('github.urlPlaceholder')}
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
              {t('github.scanning')}
            </span>
          ) : t('github.scan')}
        </button>
      </div>
      <p className="text-xs mb-3" style={{ color: 'var(--text-muted)' }}>{t('github.hint')}</p>

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
                key={c.Path}
                className="flex items-center gap-3 px-3 py-2 rounded-lg cursor-pointer transition-colors"
                style={{ color: 'var(--text-primary)' }}
                onMouseEnter={e => { e.currentTarget.style.background = 'var(--bg-hover)' }}
                onMouseLeave={e => { e.currentTarget.style.background = '' }}
              >
                <input
                  type="checkbox"
                  checked={selected.has(c.Path)}
                  onChange={() => toggle(c.Path)}
                  style={{ accentColor: 'var(--accent-secondary)' }}
                />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{c.Name}</p>
                  <p className="truncate text-xs" style={{ color: 'var(--text-muted)' }} title={c.Path}>{c.Path}</p>
                </div>
                <div className="max-w-[48%] shrink-0">
                  <SkillStatusStrip
                    className="justify-end"
                    maxVisiblePushedAgents={2}
                    badges={[
                      ...(visibility.includes('imported') && c.Installed ? [{
                        key: 'imported',
                        label: t('common.imported'),
                        tone: 'success' as const,
                      }] : []),
                      ...(visibility.includes('updatable') && c.Updatable ? [{
                        key: 'updatable',
                        label: t('common.updatable'),
                        tone: 'warning' as const,
                      }] : []),
                    ]}
                    pushedAgents={visibility.includes('pushedAgents') ? (c.PushedAgents ?? []) : []}
                  />
                </div>
              </label>
            ))}
          </div>
          <div className="flex items-center gap-3 mb-4">
            <span className="text-sm" style={{ color: 'var(--text-muted)' }}>{t('github.installTo')}</span>
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
            {installing ? t('github.installing') : t('github.installN', { count: selected.size })}
          </button>
        </>
      )}
    </AnimatedDialog>
  )
}
