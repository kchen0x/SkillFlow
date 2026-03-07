import { useEffect, useState } from 'react'
import { GetEnabledTools, ScanToolSkills, PullFromTool, PullFromToolForce, ListCategories } from '../../wailsjs/go/main/App'
import ConflictDialog from '../components/ConflictDialog'
import SyncSkillCard from '../components/SyncSkillCard'
import { ArrowDownToLine, AlertCircle, X, CheckSquare, Square } from 'lucide-react'
import { ToolIcon } from '../config/toolIcons'

export default function SyncPull() {
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
    if (selected.size === scanned.length) {
      setSelected(new Set())
    } else {
      setSelected(new Set(scanned.map((s: any) => s.Name)))
    }
  }

  const allSelected = scanned.length > 0 && selected.size === scanned.length

  const getNavStyle = (isActive: boolean) => isActive ? {
    background: 'var(--accent-glow)',
    color: 'var(--accent-primary)',
    border: '1px solid var(--border-accent)',
    boxShadow: 'var(--glow-accent-sm)',
  } : {
    color: 'var(--text-muted)',
    border: '1px solid transparent',
  }

  return (
    <div className="flex h-full overflow-hidden">
      <div className="w-48 shrink-0 p-3 flex flex-col gap-0.5" style={{ borderRight: '1px solid var(--border-base)' }}>
        <div className="px-3 py-1.5 text-xs font-medium tracking-wide uppercase" style={{ color: 'var(--text-muted)' }}>
          导入分类
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
            从工具拉取
          </div>

          <section>
            <p className="text-sm mb-3" style={{ color: 'var(--text-muted)' }}>来源工具</p>
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
            <p className="text-sm" style={{ color: 'var(--text-muted)' }}>扫描中...</p>
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
              <span>未发现任何 Skill，请确认工具目录中包含含有 skill.md 的子目录</span>
            </div>
          )}

          {scanned.length > 0 && (
            <div className="flex items-center justify-between gap-4">
              <div className="flex items-center gap-4">
                <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
                  选择要导入的 Skills
                  <span className="ml-1" style={{ color: 'var(--text-disabled)' }}>（{selected.size}/{scanned.length}）</span>
                </p>
                <button
                  onClick={toggleAll}
                  className="flex items-center gap-1.5 text-xs transition-colors"
                  style={{ color: 'var(--text-muted)' }}
                  onMouseEnter={e => { e.currentTarget.style.color = 'var(--text-primary)' }}
                  onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)' }}
                >
                  {allSelected ? <CheckSquare size={13} /> : <Square size={13} />}
                  {allSelected ? '取消全选' : '全选'}
                </button>
              </div>
              <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
                导入到分类「<span style={{ color: 'var(--text-primary)' }}>{targetCategory || defaultCategory}</span>」
              </p>
            </div>
          )}
        </div>

        {scanned.length > 0 && (
          <>
            <div className="flex-1 overflow-y-auto p-6">
              <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
                {scanned.map((sk: any) => (
                  <SyncSkillCard
                    key={sk.Name}
                    name={sk.Name}
                    path={sk.Path}
                    selected={selected.has(sk.Name)}
                    onToggle={() => toggle(sk.Name)}
                  />
                ))}
              </div>
            </div>

            <div className="px-6 py-4 flex items-center gap-4" style={{ borderTop: '1px solid var(--border-base)' }}>
              <button
                onClick={pull}
                disabled={pulling || selected.size === 0}
                className="btn-primary px-6 py-2 rounded-lg text-sm"
              >
                {pulling ? '拉取中...' : `开始拉取 (${selected.size})`}
              </button>
              {done && <span className="text-sm" style={{ color: 'var(--color-success)' }}>拉取完成 ✓</span>}
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
