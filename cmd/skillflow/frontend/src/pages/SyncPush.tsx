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

type Scope = 'auto' | 'manual'

export default function SyncPush() {
  const [tools, setTools] = useState<any[]>([])
  const [selectedTools, setSelectedTools] = useState<Set<string>>(new Set())
  const [categories, setCategories] = useState<string[]>([])
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null)
  const [skills, setSkills] = useState<any[]>([])
  const [scope, setScope] = useState<Scope>('auto')
  const [selectedSkills, setSelectedSkills] = useState<Set<string>>(new Set())
  const [conflicts, setConflicts] = useState<string[]>([])
  const [pushing, setPushing] = useState(false)
  const [done, setDone] = useState(false)
  const [missingDirs, setMissingDirs] = useState<{ name: string; dir: string }[]>([])
  const [pendingPush, setPendingPush] = useState(false)

  useEffect(() => {
    Promise.all([GetEnabledTools(), ListSkills(), ListCategories()]).then(([t, s, c]) => {
      setTools(t ?? [])
      setSkills(s ?? [])
      setCategories(c ?? [])
    })
  }, [])

  const filteredSkills = useMemo(
    () => skills.filter((skill: any) => selectedCategory === null || skill.Category === selectedCategory),
    [skills, selectedCategory],
  )

  const pushIDs = useMemo(() => {
    if (scope === 'manual') return Array.from(selectedSkills)
    return filteredSkills.map((skill: any) => skill.ID)
  }, [filteredSkills, scope, selectedSkills])

  const pushCount = pushIDs.length
  const allManualSelected = filteredSkills.length > 0 && selectedSkills.size === filteredSkills.length

  const scopeLabel = scope === 'manual'
    ? `手动选择 ${selectedSkills.size}/${filteredSkills.length}`
    : selectedCategory === null
      ? `全部 Skills (${filteredSkills.length})`
      : `分类「${selectedCategory}」(${filteredSkills.length})`

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
    if (allManualSelected) {
      setSelectedSkills(new Set())
      return
    }
    setSelectedSkills(new Set(filteredSkills.map((skill: any) => skill.ID)))
  }

  const setAutoScope = () => {
    setScope('auto')
    setSelectedSkills(new Set())
  }

  const setManualScope = () => {
    setScope('manual')
    setSelectedSkills(new Set(filteredSkills.map((skill: any) => skill.ID)))
  }

  const getNavStyle = (isActive: boolean) => isActive ? {
    background: 'var(--accent-glow)',
    color: 'var(--accent-primary)',
    border: '1px solid var(--border-accent)',
    boxShadow: 'var(--glow-accent-sm)',
  } : {
    color: 'var(--text-muted)',
    border: '1px solid transparent',
  }

  const getScopeButtonStyle = (isActive: boolean) => isActive ? {
    background: 'var(--accent-glow)',
    color: 'var(--accent-primary)',
    border: '1px solid var(--border-accent)',
    boxShadow: 'var(--glow-accent-sm)',
  } : {
    background: 'var(--bg-elevated)',
    color: 'var(--text-secondary)',
    border: '1px solid var(--border-base)',
  }

  return (
    <div className="flex h-full overflow-hidden">
      <div className="w-48 shrink-0 p-3 flex flex-col gap-0.5" style={{ borderRight: '1px solid var(--border-base)' }}>
        <div className="px-3 py-1.5 text-xs font-medium tracking-wide uppercase" style={{ color: 'var(--text-muted)' }}>
          推送范围
        </div>
        <button
          onClick={() => setSelectedCategory(null)}
          className="px-3 py-2 rounded-lg text-sm text-left transition-all duration-150"
          style={getNavStyle(selectedCategory === null)}
        >
          全部
        </button>
        {categories.map(category => (
          <button
            key={category}
            onClick={() => setSelectedCategory(category)}
            className="px-3 py-2 rounded-lg text-sm text-left transition-all duration-150"
            style={getNavStyle(selectedCategory === category)}
          >
            {category}
          </button>
        ))}
      </div>

      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="px-6 py-4 flex flex-col gap-4" style={{ borderBottom: '1px solid var(--border-base)' }}>
          <div className="flex items-center gap-2 text-lg font-semibold" style={{ color: 'var(--text-primary)' }}>
            <ArrowUpFromLine size={18} />
            推送到工具
          </div>

          <section>
            <p className="text-sm mb-3" style={{ color: 'var(--text-muted)' }}>目标工具</p>
            <div className="flex flex-wrap gap-2">
              {tools.map(tool => (
                <button
                  key={tool.name}
                  onClick={() => toggleTool(tool.name)}
                  className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-all duration-200"
                  style={selectedTools.has(tool.name) ? {
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
                  <ToolIcon name={tool.name} size={20} />
                  {tool.name}
                </button>
              ))}
            </div>
          </section>

          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <button
                onClick={setAutoScope}
                className="px-3 py-1.5 rounded-lg text-sm transition-all duration-200"
                style={getScopeButtonStyle(scope === 'auto')}
              >
                {selectedCategory === null ? '推送全部' : '推送当前分类'}
              </button>
              <button
                onClick={setManualScope}
                className="px-3 py-1.5 rounded-lg text-sm transition-all duration-200"
                style={getScopeButtonStyle(scope === 'manual')}
              >
                手动选择 Skill
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
                {allManualSelected ? '取消全选' : '全选当前列表'}
              </button>
              <span style={{ color: 'var(--text-muted)' }}>
                当前可选 {filteredSkills.length} 个 Skill
              </span>
            </div>
          )}
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
            {filteredSkills.map((skill: any) => (
              <SyncSkillCard
                key={skill.ID}
                id={skill.ID}
                name={skill.Name}
                subtitle={skill.Category || undefined}
                source={skill.Source}
                path={skill.Path}
                selected={scope === 'manual' && selectedSkills.has(skill.ID)}
                showSelection={scope === 'manual'}
                onToggle={() => toggleSkill(skill.ID)}
              />
            ))}
          </div>

          {filteredSkills.length === 0 && (
            <div className="flex flex-col items-center justify-center h-48" style={{ color: 'var(--text-muted)' }}>
              <p className="text-sm">当前范围内没有 Skill</p>
              <p className="text-xs mt-1">选择"全部"或切换到其他分类后再试</p>
            </div>
          )}
        </div>

        <div className="px-6 py-4 shrink-0 flex items-center gap-4" style={{ borderTop: '1px solid var(--border-base)' }}>
          <button
            onClick={push}
            disabled={pushing || selectedTools.size === 0 || pushCount === 0}
            className="btn-primary px-6 py-2 rounded-lg text-sm"
          >
            {pushing ? '推送中...' : `开始推送 (${pushCount})`}
          </button>
          {done && <span className="text-sm" style={{ color: 'var(--color-success)' }}>推送完成</span>}
        </div>
      </div>

      {conflicts.length > 0 && (
        <ConflictDialog
          conflicts={conflicts}
          onOverwrite={async (name) => {
            const skill = skills.find(s => s.Name === name)
            if (skill) await PushToToolsForce([skill.ID], Array.from(selectedTools))
            setConflicts(prev => prev.filter(item => item !== name))
          }}
          onSkip={(name) => setConflicts(prev => prev.filter(item => item !== name))}
          onDone={() => setDone(true)}
        />
      )}

      <AnimatedDialog open={pendingPush} width="w-[460px]" zIndex={50}>
        <div className="flex justify-between items-center mb-1">
          <h3 className="font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
            <FolderPlus size={16} /> 目录不存在
          </h3>
          <button
            onClick={() => { setMissingDirs([]); setPendingPush(false) }}
            style={{ color: 'var(--text-muted)' }}
          >
            <X size={16} />
          </button>
        </div>
        <p className="text-xs mb-3" style={{ color: 'var(--text-muted)' }}>以下推送目录尚未创建，是否自动创建后继续推送？</p>
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
            创建并推送
          </button>
          <button
            onClick={() => { setMissingDirs([]); setPendingPush(false) }}
            className="btn-secondary flex-1 py-2 rounded-lg text-sm"
          >
            取消
          </button>
        </div>
      </AnimatedDialog>
    </div>
  )
}
