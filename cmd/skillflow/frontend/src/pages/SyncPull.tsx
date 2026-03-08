import { useEffect, useMemo, useState } from 'react'
import { GetEnabledTools, ScanToolSkills, PullFromTool, PullFromToolForce, ListCategories } from '../../wailsjs/go/main/App'
import ConflictDialog from '../components/ConflictDialog'
import SyncSkillCard from '../components/SyncSkillCard'
import { ArrowDownToLine, AlertCircle, X, CheckSquare, Square } from 'lucide-react'
import { ToolIcon } from '../config/toolIcons'
import SkillListControls from '../components/SkillListControls'
import { useLanguage } from '../contexts/LanguageContext'
import { useSkillStatusVisibility } from '../contexts/SkillStatusVisibilityContext'
import { SkillSortOrder, filterAndSortSkills } from '../lib/skillList'

export default function SyncPull() {
  const { t } = useLanguage()
  const visibility = useSkillStatusVisibility('pullFromTool')
  const defaultCategory = 'Default'
  const [tools, setTools] = useState<any[]>([])
  const [selectedTool, setSelectedTool] = useState('')
  const [scanned, setScanned] = useState<any[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [categories, setCategories] = useState<string[]>([])
  const [targetCategory, setTargetCategory] = useState(defaultCategory)
  const [scanning, setScanning] = useState(false)
  const [pulling, setPulling] = useState(false)
  const [conflicts, setConflicts] = useState<string[]>([])
  const [done, setDone] = useState(false)
  const [scanError, setScanError] = useState('')
  const [scannedOnce, setScannedOnce] = useState(false)
  const [search, setSearch] = useState('')
  const [sortOrder, setSortOrder] = useState<SkillSortOrder>('asc')

  useEffect(() => {
    Promise.all([GetEnabledTools(), ListCategories()]).then(([t, c]) => {
      setTools(t ?? [])
      setCategories(c ?? [])
    })
  }, [])

  const scan = async (toolName: string) => {
    setScanning(true)
    setScanned([])
    setScanError('')
    setDone(false)
    try {
      const skills = await ScanToolSkills(toolName)
      setScanned(skills ?? [])
      setSelected(new Set())
      setScannedOnce(true)
    } catch (e: any) {
      setScanError(String(e?.message ?? e))
    } finally {
      setScanning(false)
    }
  }

  const pull = async () => {
    setPulling(true)
    const paths = [...selected]
    const result = await PullFromTool(selectedTool, paths, targetCategory)
    if (result && result.length > 0) {
      setConflicts(result)
    } else {
      setDone(true)
    }
    setPulling(false)
  }

  const toggle = (path: string) => {
    const next = new Set(selected)
    next.has(path) ? next.delete(path) : next.add(path)
    setSelected(next)
  }

  const toggleAll = () => {
    const visiblePaths = filteredScanned.map((skill: any) => skill.path)
    setSelected(prev => {
      const next = new Set(prev)
      if (visiblePaths.every(path => next.has(path))) {
        visiblePaths.forEach(path => next.delete(path))
      } else {
        visiblePaths.forEach(path => next.add(path))
      }
      return next
    })
  }
  const toggleNotImported = () => {
    if (visibleNotImportedPaths.length === 0) return
    setSelected(prev => {
      const next = new Set(prev)
      if (visibleNotImportedPaths.every(path => next.has(path))) {
        visibleNotImportedPaths.forEach(path => next.delete(path))
      } else {
        visibleNotImportedPaths.forEach(path => next.add(path))
      }
      return next
    })
  }


  const filteredScanned = useMemo(
    () => filterAndSortSkills(scanned, search, sortOrder, skill => skill.name ?? ''),
    [scanned, search, sortOrder],
  )

  const visibleNotImportedPaths = useMemo(
    () => filteredScanned.filter((skill: any) => !skill.imported).map((skill: any) => skill.path),
    [filteredScanned],
  )

  const allNotImportedSelected = visibleNotImportedPaths.length > 0 && visibleNotImportedPaths.every(path => selected.has(path))

  const allSelected = filteredScanned.length > 0 && filteredScanned.every((skill: any) => selected.has(skill.path))

  const getNavStyle = (isActive: boolean) => isActive ? {
    background: 'var(--active-surface)',
    color: 'var(--active-text)',
    border: '1px solid var(--active-border)',
    boxShadow: 'var(--active-shadow)',
  } : {
    color: 'var(--text-muted)',
    border: '1px solid transparent',
  }

  useEffect(() => {
    const visiblePaths = new Set(filteredScanned.map((skill: any) => skill.path))
    setSelected(prev => {
      const next = new Set([...prev].filter(path => visiblePaths.has(path)))
      return next.size === prev.size ? prev : next
    })
  }, [filteredScanned])

  return (
    <div className="flex h-full overflow-hidden">
      <div className="w-48 shrink-0 p-3 flex flex-col gap-0.5" style={{ borderRight: '1px solid var(--border-base)' }}>
        <div className="px-3 py-1.5 text-xs font-medium tracking-wide uppercase" style={{ color: 'var(--text-muted)' }}>
          {t('syncPull.importCategory')}
        </div>
        {categories.map(category => {
          const active = targetCategory === category
          return (
            <button
              key={category}
              onClick={() => setTargetCategory(category)}
              className={`px-3 py-2 rounded-lg text-sm text-left transition-all duration-150 ${active ? 'font-semibold -translate-y-px' : ''}`}
              style={getNavStyle(active)}
            >
              {category}
            </button>
          )
        })}
      </div>

      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="px-6 py-4 flex flex-col gap-4" style={{ borderBottom: '1px solid var(--border-base)' }}>
          <div className="flex items-center gap-2 text-lg font-semibold" style={{ color: 'var(--text-primary)' }}>
            <ArrowDownToLine size={18} />
            {t('syncPull.title')}
          </div>

          <section>
            <p className="text-sm mb-3" style={{ color: 'var(--text-muted)' }}>{t('syncPull.sourceTool')}</p>
            <div className="flex flex-wrap gap-2">
              {tools.map(t => {
                const active = selectedTool === t.name
                return (
                  <button
                    key={t.name}
                    onClick={() => {
                      setSelectedTool(t.name)
                      setScanned([])
                      setDone(false)
                      setScanError('')
                      setScannedOnce(false)
                      scan(t.name)
                    }}
                    className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-all duration-200 ${active ? 'font-semibold -translate-y-px' : ''}`}
                    style={active ? {
                      background: 'var(--active-surface)',
                      color: 'var(--active-text)',
                      border: '1px solid var(--active-border)',
                      boxShadow: 'var(--active-shadow)',
                    } : {
                      background: 'var(--bg-elevated)',
                      color: 'var(--text-secondary)',
                      border: '1px solid var(--border-base)',
                    }}
                  >
                    <ToolIcon name={t.name} size={20} />
                    <span>{t.name}</span>
                  </button>
                )
              })}
            </div>
          </section>

          {scanning && (
            <p className="text-sm" style={{ color: 'var(--text-muted)' }}>{t('syncPull.scanning')}</p>
          )}

          {scanError && (
            <div
              className="flex items-start gap-2 rounded-lg px-4 py-3 text-sm"
              style={{ background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)', color: 'var(--color-error)' }}
            >
              <AlertCircle size={16} className="mt-0.5 shrink-0" />
              <span className="flex-1">{scanError}</span>
              <button onClick={() => setScanError('')} className="shrink-0 hover:opacity-70">
                <X size={14} />
              </button>
            </div>
          )}

          {!scanError && !scanning && scannedOnce && scanned.length === 0 && (
            <div
              className="flex items-center gap-2 rounded-lg px-4 py-3 text-sm"
              style={{ background: 'rgba(251,191,36,0.1)', border: '1px solid rgba(251,191,36,0.3)', color: 'var(--color-warning)' }}
            >
              <AlertCircle size={16} className="shrink-0" />
              <span>{t('syncPull.emptyWarning')}</span>
            </div>
          )}

          {scanned.length > 0 && (
            <>
              <SkillListControls
                search={search}
                onSearchChange={setSearch}
                sortOrder={sortOrder}
                onSortOrderChange={setSortOrder}
                placeholder={t('syncPull.searchPlaceholder')}
                resultLabel={t('common.showingNSkills', { count: filteredScanned.length })}
              />
              <div className="flex items-center justify-between gap-4">
                <div className="flex items-center gap-4 flex-wrap">
                  <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
                    {t('syncPull.selectSkills')}
                    <span className="ml-1" style={{ color: 'var(--text-disabled)' }}>({selected.size}/{filteredScanned.length})</span>
                  </p>
                  <button
                    onClick={toggleAll}
                    className="flex items-center gap-1.5 text-xs transition-colors"
                    style={{ color: 'var(--text-muted)' }}
                    onMouseEnter={e => { e.currentTarget.style.color = 'var(--text-primary)' }}
                    onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)' }}
                  >
                    {allSelected ? <CheckSquare size={13} /> : <Square size={13} />}
                    {allSelected ? t('common.deselectAll') : t('common.selectAll')}
                  </button>
                  <button
                    onClick={toggleNotImported}
                    disabled={visibleNotImportedPaths.length === 0}
                    className="flex items-center gap-1.5 text-xs transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                    style={{ color: 'var(--text-muted)' }}
                    onMouseEnter={e => {
                      if (!e.currentTarget.disabled) e.currentTarget.style.color = 'var(--text-primary)'
                    }}
                    onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)' }}
                  >
                    {allNotImportedSelected ? <CheckSquare size={13} /> : <Square size={13} />}
                    {t('syncPull.selectNotImported')}
                  </button>
                </div>
                <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
                  {t('syncPull.importTo', { cat: targetCategory || defaultCategory })}
                </p>
              </div>
            </>
          )}
        </div>

        {scanned.length > 0 && (
          <>
            <div className="flex-1 overflow-y-auto p-6">
              {filteredScanned.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-48" style={{ color: 'var(--text-muted)' }}>
                  <p className="text-sm">{t('syncPull.noMatch')}</p>
                </div>
              ) : (
                <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
                  {filteredScanned.map((sk: any) => (
                    <SyncSkillCard
                      key={sk.path}
                      name={sk.name}
                      path={sk.path}
                      imported={sk.imported}
                      showImported={visibility.includes('imported')}
                      selected={selected.has(sk.path)}
                      onToggle={() => toggle(sk.path)}
                    />
                  ))}
                </div>
              )}
            </div>

            <div className="px-6 py-4 flex items-center gap-4" style={{ borderTop: '1px solid var(--border-base)' }}>
              <button
                onClick={pull}
                disabled={pulling || selected.size === 0}
                className="btn-primary px-6 py-2 rounded-lg text-sm"
              >
                {pulling ? t('syncPull.pulling') : t('syncPull.startPull', { count: selected.size })}
              </button>
              {done && <span className="text-sm" style={{ color: 'var(--color-success)' }}>{t('syncPull.done')}</span>}
            </div>
          </>
        )}
      </div>

      {conflicts.length > 0 && (
        <ConflictDialog
          conflicts={conflicts}
          labelForConflict={(path) => scanned.find((item: any) => item.path === path)?.name ?? path}
          onOverwrite={async (path) => {
            await PullFromToolForce(selectedTool, [path], targetCategory)
            setConflicts(prev => prev.filter(c => c !== path))
          }}
          onSkip={(path) => setConflicts(prev => prev.filter(c => c !== path))}
          onDone={() => setDone(true)}
        />
      )}
    </div>
  )
}
