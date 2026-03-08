import { useEffect, useMemo, useState } from 'react'
import {
  CheckMissingPushDirs,
  GetEnabledTools,
  ListCategories,
  ListSkills,
  PushToTools,
  PushToToolsForce,
} from '../../wailsjs/go/main/App'
import ConflictDialog from '../components/ConflictDialog'
import SyncSkillCard from '../components/SyncSkillCard'
import { ArrowUpFromLine, CheckSquare, FolderPlus, Square, X } from 'lucide-react'
import { ToolIcon } from '../config/toolIcons'
import AnimatedDialog from '../components/ui/AnimatedDialog'
import SkillListControls from '../components/SkillListControls'
import { useLanguage } from '../contexts/LanguageContext'
import { useSkillStatusVisibility } from '../contexts/SkillStatusVisibilityContext'
import { SkillSortOrder, filterAndSortSkills } from '../lib/skillList'

type Scope = 'auto' | 'manual'

export default function SyncPush() {
  const { t } = useLanguage()
  const visibility = useSkillStatusVisibility('pushToTool')
  const [tools, setTools] = useState<any[]>([])
  const [selectedTools, setSelectedTools] = useState<Set<string>>(new Set())
  const [categories, setCategories] = useState<string[]>([])
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null)
  const [skills, setSkills] = useState<any[]>([])
  const [scope, setScope] = useState<Scope>('auto')
  const [selectedSkills, setSelectedSkills] = useState<Set<string>>(new Set())
  const [conflicts, setConflicts] = useState<any[]>([])
  const [pushing, setPushing] = useState(false)
  const [done, setDone] = useState(false)
  const [missingDirs, setMissingDirs] = useState<{ name: string; dir: string }[]>([])
  const [pendingPush, setPendingPush] = useState(false)
  const [search, setSearch] = useState('')
  const [sortOrder, setSortOrder] = useState<SkillSortOrder>('asc')

  useEffect(() => {
    Promise.all([GetEnabledTools(), ListSkills(), ListCategories()]).then(([t, s, c]) => {
      setTools(t ?? [])
      setSkills(s ?? [])
      setCategories(c ?? [])
    })
  }, [])

  const filteredSkills = useMemo(
    () => skills.filter((skill: any) => selectedCategory === null || skill.category === selectedCategory),
    [skills, selectedCategory],
  )

  const visibleSkills = useMemo(
    () => filterAndSortSkills(filteredSkills, search, sortOrder, skill => skill.name ?? ''),
    [filteredSkills, search, sortOrder],
  )

  const pushIDs = useMemo(() => {
    if (scope === 'manual') return Array.from(selectedSkills)
    return visibleSkills.map((skill: any) => skill.id)
  }, [scope, selectedSkills, visibleSkills])

  const pushCount = pushIDs.length
  const allManualSelected = visibleSkills.length > 0 && visibleSkills.every((skill: any) => selectedSkills.has(skill.id))

  const scopeLabel = scope === 'manual'
    ? t('syncPush.scopeManual', { count: selectedSkills.size })
    : selectedCategory === null
      ? t('syncPush.scopeAll', { count: visibleSkills.length })
      : t('syncPush.scopeCategory', { cat: selectedCategory ?? '', count: visibleSkills.length })

  const doPush = async () => {
    setPushing(true)
    setDone(false)
    const toolNames = Array.from(selectedTools)
    const result = await PushToTools(pushIDs, toolNames)
    if (result && result.length > 0) {
      setConflicts(result)
    } else {
      setDone(true)
    }
    setPushing(false)
  }

  const push = async () => {
    const toolNames = Array.from(selectedTools)
    const missing = await CheckMissingPushDirs(toolNames)
    if (missing && missing.length > 0) {
      setMissingDirs(missing as { name: string; dir: string }[])
      setPendingPush(true)
      return
    }
    await doPush()
  }

  const confirmMkdirAndPush = async () => {
    setMissingDirs([])
    setPendingPush(false)
    await doPush()
  }

  const toggleTool = (name: string) => {
    setSelectedTools(prev => {
      const next = new Set(prev)
      next.has(name) ? next.delete(name) : next.add(name)
      return next
    })
  }

  const toggleSkill = (id: string) => {
    if (scope !== 'manual') return
    setSelectedSkills(prev => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })
  }

  const toggleAllManual = () => {
    const visibleIDs = visibleSkills.map((skill: any) => skill.id)
    setSelectedSkills(prev => {
      const next = new Set(prev)
      if (visibleIDs.every(id => next.has(id))) {
        visibleIDs.forEach(id => next.delete(id))
      } else {
        visibleIDs.forEach(id => next.add(id))
      }
      return next
    })
  }

  const setAutoScope = () => {
    setScope('auto')
    setSelectedSkills(new Set())
  }

  const setManualScope = () => {
    setScope('manual')
    setSelectedSkills(new Set(visibleSkills.map((skill: any) => skill.id)))
  }

  const getNavStyle = (isActive: boolean) => isActive ? {
    background: 'var(--active-surface)',
    color: 'var(--active-text)',
    border: '1px solid var(--active-border)',
    boxShadow: 'var(--active-shadow)',
  } : {
    color: 'var(--text-muted)',
    border: '1px solid transparent',
  }

  const getScopeButtonStyle = (isActive: boolean) => isActive ? {
    background: 'var(--active-surface)',
    color: 'var(--active-text)',
    border: '1px solid var(--active-border)',
    boxShadow: 'var(--active-shadow)',
  } : {
    background: 'var(--bg-elevated)',
    color: 'var(--text-secondary)',
    border: '1px solid var(--border-base)',
  }

  useEffect(() => {
    if (scope !== 'manual') return
    const visibleIDs = new Set(visibleSkills.map((skill: any) => skill.id))
    setSelectedSkills(prev => {
      const next = new Set([...prev].filter(id => visibleIDs.has(id)))
      return next.size === prev.size ? prev : next
    })
  }, [scope, visibleSkills])

  return (
    <div className="flex h-full overflow-hidden">
      <div className="w-48 shrink-0 p-3 flex flex-col gap-0.5" style={{ borderRight: '1px solid var(--border-base)' }}>
        <div className="px-3 py-1.5 text-xs font-medium tracking-wide uppercase" style={{ color: 'var(--text-muted)' }}>
          {t('syncPush.pushRange')}
        </div>
        <button
          onClick={() => setSelectedCategory(null)}
          className={`px-3 py-2 rounded-lg text-sm text-left transition-all duration-150 ${selectedCategory === null ? 'font-semibold -translate-y-px' : ''}`}
          style={getNavStyle(selectedCategory === null)}
        >
          {t('common.all')}
        </button>
        {categories.map(category => {
          const active = selectedCategory === category
          return (
            <button
              key={category}
              onClick={() => setSelectedCategory(category)}
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
            <ArrowUpFromLine size={18} />
            {t('syncPush.title')}
          </div>

          <section>
            <p className="text-sm mb-3" style={{ color: 'var(--text-muted)' }}>{t('syncPush.targetTool')}</p>
            <div className="flex flex-wrap gap-2">
              {tools.map(tool => {
                const active = selectedTools.has(tool.name)
                return (
                  <button
                    key={tool.name}
                    onClick={() => toggleTool(tool.name)}
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
                    <ToolIcon name={tool.name} size={20} />
                    <span>{tool.name}</span>
                  </button>
                )
              })}
            </div>
          </section>

          <SkillListControls
            search={search}
            onSearchChange={setSearch}
            sortOrder={sortOrder}
            onSortOrderChange={setSortOrder}
            placeholder={t('syncPush.searchPlaceholder')}
            resultLabel={t('common.showingNSkills', { count: visibleSkills.length })}
          />

          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <button
                onClick={setAutoScope}
                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm transition-all duration-200 ${scope === 'auto' ? 'font-semibold -translate-y-px' : ''}`}
                style={getScopeButtonStyle(scope === 'auto')}
              >
                {selectedCategory === null ? t('syncPush.pushAll') : t('syncPush.pushCategory')}
              </button>
              <button
                onClick={setManualScope}
                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm transition-all duration-200 ${scope === 'manual' ? 'font-semibold -translate-y-px' : ''}`}
                style={getScopeButtonStyle(scope === 'manual')}
              >
                {t('syncPush.manualSelect')}
              </button>
            </div>
            <p className="text-sm" style={{ color: 'var(--text-muted)' }}>{scopeLabel}</p>
          </div>

          {scope === 'manual' && (
            <div className="flex items-center gap-4 text-sm">
              <button
                onClick={toggleAllManual}
                className="flex items-center gap-1.5 transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.color = 'var(--text-primary)' }}
                onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)' }}
              >
                {allManualSelected ? <CheckSquare size={14} /> : <Square size={14} />}
                {allManualSelected ? t('common.deselectAll') : t('syncPush.selectAllList')}
              </button>
              <span style={{ color: 'var(--text-muted)' }}>
                {t('syncPush.nSkillsVisible', { count: visibleSkills.length })}
              </span>
            </div>
          )}
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
            {visibleSkills.map((skill: any) => (
              <SyncSkillCard
                key={skill.id}
                id={skill.id}
                name={skill.name}
                subtitle={skill.category || undefined}
                source={skill.source}
                path={skill.path}
                pushedTools={skill.pushedTools}
                showPushedTools={visibility.includes('pushedTools')}
                selected={scope === 'manual' && selectedSkills.has(skill.id)}
                showSelection={scope === 'manual'}
                onToggle={() => toggleSkill(skill.id)}
              />
            ))}
          </div>

          {visibleSkills.length === 0 && (
            <div className="flex flex-col items-center justify-center h-48" style={{ color: 'var(--text-muted)' }}>
              <p className="text-sm">{t('syncPush.empty')}</p>
              <p className="text-xs mt-1">{t('syncPush.emptyHint')}</p>
            </div>
          )}
        </div>

        <div className="px-6 py-4 shrink-0 flex items-center gap-4" style={{ borderTop: '1px solid var(--border-base)' }}>
          <button
            onClick={push}
            disabled={pushing || selectedTools.size === 0 || pushCount === 0}
            className="btn-primary px-6 py-2 rounded-lg text-sm"
          >
            {pushing ? t('syncPush.pushing') : t('syncPush.startPush', { count: pushCount })}
          </button>
          {done && <span className="text-sm" style={{ color: 'var(--color-success)' }}>{t('syncPush.done')}</span>}
        </div>
      </div>

      {conflicts.length > 0 && (
        <ConflictDialog
          conflicts={conflicts}
          labelForConflict={(conflict) => `${conflict.skillName} → ${conflict.toolName}`}
          onOverwrite={async (conflict) => {
            if (conflict.skillId) {
              await PushToToolsForce([conflict.skillId], [conflict.toolName])
            }
            setConflicts(prev => prev.filter(item => !(item.skillId === conflict.skillId && item.toolName === conflict.toolName)))
          }}
          onSkip={(conflict) => setConflicts(prev => prev.filter(item => !(item.skillId === conflict.skillId && item.toolName === conflict.toolName)))}
          onDone={() => setDone(true)}
        />
      )}

      <AnimatedDialog open={pendingPush} width="w-[460px]" zIndex={50}>
        <div className="flex justify-between items-center mb-1">
          <h3 className="font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
            <FolderPlus size={16} /> {t('syncPush.mkdirTitle')}
          </h3>
          <button
            onClick={() => { setMissingDirs([]); setPendingPush(false) }}
            style={{ color: 'var(--text-muted)' }}
          >
            <X size={16} />
          </button>
        </div>
        <p className="text-xs mb-3" style={{ color: 'var(--text-muted)' }}>{t('syncPush.mkdirDesc')}</p>
        <ul className="space-y-1.5 mb-4 max-h-40 overflow-y-auto">
          {missingDirs.map(dir => (
            <li
              key={dir.name}
              className="text-sm rounded-lg px-3 py-2"
              style={{ background: 'var(--bg-surface)' }}
            >
              <span className="font-medium" style={{ color: 'var(--text-primary)' }}>{dir.name}</span>
              <span className="text-xs block truncate" style={{ color: 'var(--text-muted)' }} title={dir.dir}>{dir.dir}</span>
            </li>
          ))}
        </ul>
        <div className="flex gap-3">
          <button onClick={confirmMkdirAndPush} className="btn-primary flex-1 py-2 rounded-lg text-sm">
            {t('syncPush.createAndPush')}
          </button>
          <button
            onClick={() => { setMissingDirs([]); setPendingPush(false) }}
            className="btn-secondary flex-1 py-2 rounded-lg text-sm"
          >
            {t('common.cancel')}
          </button>
        </div>
      </AnimatedDialog>
    </div>
  )
}
