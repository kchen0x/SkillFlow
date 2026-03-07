import { useEffect, useMemo, useState } from 'react'
import { GetEnabledTools, ScanToolSkills, PullFromTool, PullFromToolForce, ListCategories } from '../../wailsjs/go/main/App'
import ConflictDialog from '../components/ConflictDialog'
import SyncSkillCard from '../components/SyncSkillCard'
import { ArrowDownToLine, AlertCircle, X, CheckSquare, Square } from 'lucide-react'
import { ToolIcon } from '../config/toolIcons'
import SkillListControls from '../components/SkillListControls'
import { useLanguage } from '../contexts/LanguageContext'
import { SkillSortOrder, filterAndSortSkills } from '../lib/skillList'

export default function SyncPull() {
  const { t } = useLanguage()
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
      setSelected(new Set((skills ?? []).map((s: any) => s.Name)))
      setScannedOnce(true)
    } catch (e: any) {
      setScanError(String(e?.message ?? e))
    } finally {
      setScanning(false)
    }
  }

  const pull = async () => {
    setPulling(true)
    const names = [...selected]
    const result = await PullFromTool(selectedTool, names, targetCategory)
    if (result && result.length > 0) {
      setConflicts(result)
    } else {
      setDone(true)
    }
    setPulling(false)
  }

  const toggle = (name: string) => {
    const next = new Set(selected)
    next.has(name) ? next.delete(name) : next.add(name)
    setSelected(next)
  }

  const toggleAll = () => {
    const visibleNames = filteredScanned.map((skill: any) => skill.Name)
    setSelected(prev => {
      const next = new Set(prev)
      if (visibleNames.every(name => next.has(name))) {
        visibleNames.forEach(name => next.delete(name))
      } else {
        visibleNames.forEach(name => next.add(name))
      }
      return next
    })
  }

  const filteredScanned = useMemo(
    () => filterAndSortSkills(scanned, search, sortOrder, skill => skill.Name ?? ''),
    [scanned, search, sortOrder],
  )

  const allSelected = filteredScanned.length > 0 && filteredScanned.every((skill: any) => selected.has(skill.Name))

  const getNavStyle = (isActive: boolean) => isActive ? {
    background: 'var(--accent-glow)',
    color: 'var(--accent-primary)',
    border: '1px solid var(--border-accent)',
    boxShadow: 'var(--glow-accent-sm)',
  } : {
    color: 'var(--text-muted)',
    border: '1px solid transparent',
  }

  useEffect(() => {
    const visibleNames = new Set(filteredScanned.map((skill: any) => skill.Name))
    setSelected(prev => {
      const next = new Set([...prev].filter(name => visibleNames.has(name)))
      return next.size === prev.size ? prev : next
    })
  }, [filteredScanned])

  return (
    <div className="flex h-full overflow-hidden">
      <div className="w-48 shrink-0 p-3 flex flex-col gap-0.5" style={{ borderRight: '1px solid var(--border-base)' }}>
        <div className="px-3 py-1.5 text-xs font-medium tracking-wide uppercase" style={{ color: 'var(--text-muted)' }}>
          {t('syncPull.importCategory')}
        </div>
        {categories.map(category => (
          <button
            key={category}
            onClick={() => setTargetCategory(category)}
            className="px-3 py-2 rounded-lg text-sm text-left transition-all duration-150"
            style={getNavStyle(targetCategory === category)}
          >
            {category}
          </button>
        ))}
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
              {tools.map(t => (
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
                  className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-all duration-200"
                  style={selectedTool === t.name ? {
                    background: 'var(--accent-glow)',
                    color: 'var(--accent-primary)',
                    border: '1px solid var(--border-accent)',
                    boxShadow: 'var(--glow-accent-sm)',
                  } : {
                    background: 'var(--bg-elevated)',
                    color: 'var(--text-secondary)',
                    border: '1px solid var(--border-base)',
                  }}
                >
                  <ToolIcon name={t.name} size={20} />
                  {t.name}
                </button>
              ))}
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
                <div className="flex items-center gap-4">
                  <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
                    {t('syncPull.selectSkills')}
                    <span className="ml-1" style={{ color: 'var(--text-disabled)' }}>（{selected.size}/{filteredScanned.length}）</span>
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
                      key={sk.Name}
                      name={sk.Name}
                      path={sk.Path}
                      selected={selected.has(sk.Name)}
                      onToggle={() => toggle(sk.Name)}
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
          onOverwrite={async (name) => {
            await PullFromToolForce(selectedTool, [name], targetCategory)
            setConflicts(prev => prev.filter(c => c !== name))
          }}
          onSkip={(name) => setConflicts(prev => prev.filter(c => c !== name))}
          onDone={() => setDone(true)}
        />
      )}
    </div>
  )
}
