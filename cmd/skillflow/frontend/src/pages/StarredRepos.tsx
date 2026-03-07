import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { AnimatePresence, motion } from 'framer-motion'
import {
  ListStarredRepos, AddStarredRepo, AddStarredRepoWithCredentials, RemoveStarredRepo,
  UpdateStarredRepo, UpdateAllStarredRepos,
  ListAllStarSkills, ListRepoStarSkills,
  ImportStarSkills, ListCategories, OpenURL,
  GetEnabledTools, PushStarSkillsToTools, PushStarSkillsToToolsForce, CheckMissingPushDirs,
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import {
  Star, RefreshCw, Plus, Trash2, LayoutGrid, Folder,
  ChevronLeft, CheckSquare, Download, AlertCircle, X, ExternalLink, ArrowUpToLine, Lock, KeyRound, FolderPlus, CheckCircle,
} from 'lucide-react'
import SyncSkillCard from '../components/SyncSkillCard'
import { ToolIcon } from '../config/toolIcons'
import AnimatedDialog from '../components/ui/AnimatedDialog'
import { toastVariants } from '../lib/motionVariants'

export default function StarredRepos() {
  const { repoEncoded } = useParams()
  const navigate = useNavigate()
  const currentRepo = repoEncoded ? decodeURIComponent(repoEncoded) : null

  const [repos, setRepos] = useState<any[]>([])
  const [repoSkills, setRepoSkills] = useState<any[]>([])
  const [allSkills, setAllSkills] = useState<any[]>([])
  const [view, setView] = useState<'folder' | 'flat'>('folder')
  const [syncing, setSyncing] = useState(false)
  const [addUrl, setAddUrl] = useState('')
  const [showAdd, setShowAdd] = useState(false)
  const [adding, setAdding] = useState(false)
  const [addError, setAddError] = useState('')
  const [selectMode, setSelectMode] = useState(false)
  const [selectedPaths, setSelectedPaths] = useState<Set<string>>(new Set())
  const [categories, setCategories] = useState<string[]>([])
  const [importCategory, setImportCategory] = useState('')
  const [showImportDialog, setShowImportDialog] = useState(false)
  const [importing, setImporting] = useState(false)
  const [tools, setTools] = useState<any[]>([])
  const [showPushToolDialog, setShowPushToolDialog] = useState(false)
  const [selectedTools, setSelectedTools] = useState<Set<string>>(new Set())
  const [pushingToTools, setPushingToTools] = useState(false)
  const [pushConflicts, setPushConflicts] = useState<string[]>([])
  const [showPushConflictDialog, setShowPushConflictDialog] = useState(false)
  const [missingDirs, setMissingDirs] = useState<{name: string, dir: string}[]>([])
  const [showMkdirDialog, setShowMkdirDialog] = useState(false)
  const [pushSuccessMsg, setPushSuccessMsg] = useState('')
  // Auth dialogs
  const [showHttpAuthDialog, setShowHttpAuthDialog] = useState(false)
  const [showSshErrorDialog, setShowSshErrorDialog] = useState(false)
  const [authUrl, setAuthUrl] = useState('')
  const [authUsername, setAuthUsername] = useState('')
  const [authPassword, setAuthPassword] = useState('')
  const [authError, setAuthError] = useState('')
  const [authAdding, setAuthAdding] = useState(false)

  const loadRepos = async () => {
    const r = await ListStarredRepos()
    setRepos(r ?? [])
  }

  const loadAllSkills = async () => {
    const s = await ListAllStarSkills()
    setAllSkills(s ?? [])
  }

  const loadRepoSkills = async (url: string) => {
    const s = await ListRepoStarSkills(url)
    setRepoSkills(s ?? [])
  }

  useEffect(() => {
    loadRepos()
    loadAllSkills()
    ListCategories().then(c => {
      setCategories(c ?? [])
      if (c && c.length > 0) setImportCategory(c[0])
    })
    GetEnabledTools().then(t => setTools(t ?? []))
    const off1 = EventsOn('star.sync.progress', () => loadRepos())
    const off2 = EventsOn('star.sync.done', () => { loadRepos(); loadAllSkills(); setSyncing(false) })
    return () => { off1?.(); off2?.() }
  }, [])

  useEffect(() => {
    if (currentRepo) loadRepoSkills(currentRepo)
  }, [currentRepo])

  const handleAddRepo = async () => {
    setAdding(true); setAddError('')
    try {
      await AddStarredRepo(addUrl)
      setShowAdd(false); setAddUrl('')
      await Promise.all([loadRepos(), loadAllSkills()])
    } catch (e: any) {
      const msg = String(e?.message ?? e ?? '添加失败')
      if (msg.startsWith('AUTH_SSH:')) {
        setShowAdd(false)
        setShowSshErrorDialog(true)
      } else if (msg.startsWith('AUTH_HTTP:')) {
        setAuthUrl(addUrl)
        setAuthUsername(''); setAuthPassword(''); setAuthError('')
        setShowHttpAuthDialog(true)
      } else {
        setAddError(msg)
      }
    } finally { setAdding(false) }
  }

  const handleAuthRetry = async () => {
    setAuthAdding(true); setAuthError('')
    try {
      await AddStarredRepoWithCredentials(authUrl, authUsername, authPassword)
      setShowHttpAuthDialog(false)
      setShowAdd(false); setAddUrl('')
      await Promise.all([loadRepos(), loadAllSkills()])
    } catch (e: any) {
      setAuthError(String(e?.message ?? e ?? '认证失败'))
    } finally { setAuthAdding(false) }
  }

  const handleUpdateAll = async () => {
    setSyncing(true)
    try {
      await UpdateAllStarredRepos()
    } finally {
      setSyncing(false)
      await Promise.all([loadRepos(), loadAllSkills()])
    }
  }

  const handleUpdateOne = async (url: string) => {
    await UpdateStarredRepo(url)
    await Promise.all([loadRepos(), loadAllSkills()])
  }

  const handleRemove = async (url: string) => {
    await RemoveStarredRepo(url)
    await Promise.all([loadRepos(), loadAllSkills()])
  }

  const toggleSelectPath = (path: string) => {
    setSelectedPaths(prev => {
      const next = new Set(prev)
      next.has(path) ? next.delete(path) : next.add(path)
      return next
    })
  }

  const toggleSelectAll = (skills: any[]) => {
    if (selectedPaths.size === skills.length) setSelectedPaths(new Set())
    else setSelectedPaths(new Set(skills.map((s: any) => s.path)))
  }

  const handleBatchImport = async () => {
    setImporting(true)
    try {
      const skills = currentRepo ? repoSkills : allSkills
      const byRepo = new Map<string, string[]>()
      for (const path of selectedPaths) {
        const sk = skills.find((s: any) => s.path === path)
        if (!sk) continue
        const arr = byRepo.get(sk.repoUrl) ?? []
        arr.push(path)
        byRepo.set(sk.repoUrl, arr)
      }
      for (const [rURL, paths] of byRepo) {
        await ImportStarSkills(paths, rURL, importCategory)
      }
      setShowImportDialog(false)
      setSelectMode(false)
      setSelectedPaths(new Set())
      if (currentRepo) loadRepoSkills(currentRepo); else loadAllSkills()
    } catch (e: any) {
      console.error('Import failed:', e)
    } finally { setImporting(false) }
  }

  const doPushToTools = async () => {
    setPushingToTools(true)
    try {
      const paths = [...selectedPaths]
      const toolNames = [...selectedTools]
      const conflicts = await PushStarSkillsToTools(paths, toolNames)
      setShowPushToolDialog(false)
      if (conflicts && conflicts.length > 0) {
        setPushConflicts(conflicts)
        setShowPushConflictDialog(true)
      } else {
        setSelectMode(false)
        setSelectedPaths(new Set())
        const count = paths.length
        const toolCount = toolNames.length
        setPushSuccessMsg(`已成功推送 ${count} 个 Skill 到 ${toolCount} 个工具`)
        setTimeout(() => setPushSuccessMsg(''), 3000)
      }
    } catch (e: any) {
      console.error('Push to tools failed:', e)
    } finally {
      setPushingToTools(false)
    }
  }

  const handlePushToTools = async () => {
    const toolNames = [...selectedTools]
    const missing = await CheckMissingPushDirs(toolNames)
    if (missing && missing.length > 0) {
      setMissingDirs(missing as {name: string, dir: string}[])
      setShowMkdirDialog(true)
    } else {
      await doPushToTools()
    }
  }

  const confirmMkdirAndPush = async () => {
    setShowMkdirDialog(false)
    setMissingDirs([])
    await doPushToTools()
  }

  const handlePushToToolsForce = async () => {
    try {
      const paths = [...selectedPaths]
      const toolNames = [...selectedTools]
      await PushStarSkillsToToolsForce(paths, toolNames)
      setShowPushConflictDialog(false)
      setSelectMode(false)
      setSelectedPaths(new Set())
      setPushConflicts([])
      setPushSuccessMsg(`已成功推送 ${paths.length} 个 Skill 到 ${toolNames.length} 个工具`)
      setTimeout(() => setPushSuccessMsg(''), 3000)
    } catch (e: any) {
      console.error('Force push failed:', e)
    }
  }

  const toggleTool = (name: string) => {
    setSelectedTools(prev => {
      const next = new Set(prev)
      next.has(name) ? next.delete(name) : next.add(name)
      return next
    })
  }

  const skills = currentRepo ? repoSkills : allSkills

  const toolBtnStyle = (active: boolean) => active ? {
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
    <div className="flex flex-col h-full">
      {/* Success toast */}
      <AnimatePresence>
        {pushSuccessMsg && (
          <motion.div
            variants={toastVariants}
            initial="initial"
            animate="animate"
            exit="exit"
            className="fixed top-4 left-1/2 -translate-x-1/2 z-50 flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm shadow-dialog"
            style={{
              background: 'rgba(52,211,153,0.15)',
              border: '1px solid rgba(52,211,153,0.4)',
              color: 'var(--color-success)',
              backdropFilter: 'blur(8px)',
            }}
          >
            <CheckCircle size={15} className="shrink-0" />
            {pushSuccessMsg}
          </motion.div>
        )}
      </AnimatePresence>

      {/* Toolbar */}
      <div className="flex items-center gap-3 px-6 py-4 flex-wrap" style={{ borderBottom: '1px solid var(--border-base)' }}>
        {currentRepo ? (
          <button
            onClick={() => { navigate('/starred'); setSelectMode(false); setSelectedPaths(new Set()) }}
            className="flex items-center gap-1 text-sm transition-colors"
            style={{ color: 'var(--text-muted)' }}
            onMouseEnter={e => { e.currentTarget.style.color = 'var(--text-primary)' }}
            onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)' }}
          >
            <ChevronLeft size={14} />
            <span>{currentRepo.split('/').slice(-2).join('/')}</span>
          </button>
        ) : (
          <h2 className="text-sm font-medium flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
            <Star size={14} /> 仓库收藏
          </h2>
        )}
        <div className="flex-1" />
        {selectMode ? (
          <>
            <button
              onClick={() => toggleSelectAll(skills)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
            >
              <CheckSquare size={14} />{selectedPaths.size === skills.length ? '取消全选' : '全选'}
            </button>
            <button
              onClick={() => { setSelectedTools(new Set()); setShowPushToolDialog(true) }}
              disabled={selectedPaths.size === 0}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg disabled:opacity-40 transition-colors"
              style={{ background: 'rgba(52,211,153,0.2)', color: 'var(--color-success)', border: '1px solid rgba(52,211,153,0.4)' }}
            >
              <ArrowUpToLine size={14} /> 推送到工具 {selectedPaths.size > 0 ? `(${selectedPaths.size})` : ''}
            </button>
            <button
              onClick={() => setShowImportDialog(true)}
              disabled={selectedPaths.size === 0}
              className="btn-primary flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg disabled:opacity-40"
            >
              <Download size={14} /> 导入到我的Skills {selectedPaths.size > 0 ? `(${selectedPaths.size})` : ''}
            </button>
            <button
              onClick={() => { setSelectMode(false); setSelectedPaths(new Set()) }}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
            >取消</button>
          </>
        ) : (
          <>
            {!currentRepo && (
              <>
                {[['folder', '文件夹', <Folder size={14} />], ['flat', '平铺', <LayoutGrid size={14} />]].map(([v, label, icon]) => (
                  <button
                    key={v as string}
                    onClick={() => setView(v as 'folder' | 'flat')}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-all duration-150"
                    style={view === v ? {
                      background: 'var(--bg-elevated)',
                      color: 'var(--text-primary)',
                      border: '1px solid var(--border-surface)',
                    } : {
                      color: 'var(--text-muted)',
                      border: '1px solid transparent',
                    }}
                  >
                    {icon} {label}
                  </button>
                ))}
              </>
            )}
            <button
              onClick={() => setSelectMode(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
            >
              <CheckSquare size={14} /> 批量导入
            </button>
            <button
              onClick={handleUpdateAll}
              disabled={syncing}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
            >
              <RefreshCw size={14} className={syncing ? 'animate-spin' : ''} /> 全部更新
            </button>
            {!currentRepo && (
              <button
                onClick={() => setShowAdd(true)}
                className="btn-primary flex items-center gap-1.5 px-4 py-1.5 text-sm rounded-lg"
              >
                <Plus size={14} /> 添加仓库
              </button>
            )}
          </>
        )}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-6">
        {currentRepo ? (
          <SkillGrid skills={repoSkills} selectMode={selectMode} selectedPaths={selectedPaths} onToggle={toggleSelectPath} showRepo />
        ) : view === 'folder' ? (
          <RepoGrid repos={repos}
            onEnter={url => navigate(`/starred/${encodeURIComponent(url)}`)}
            onUpdate={handleUpdateOne}
            onRemove={handleRemove} />
        ) : (
          <SkillGrid skills={allSkills} selectMode={selectMode} selectedPaths={selectedPaths} onToggle={toggleSelectPath} showRepo />
        )}
      </div>

      {/* Add repo dialog */}
      <AnimatedDialog open={showAdd} onClose={() => { setShowAdd(false); setAddError('') }} width="w-[460px]">
        <div className="flex justify-between items-center mb-4">
          <h3 className="font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
            <Star size={16} /> 添加远程仓库
          </h3>
          <button onClick={() => { setShowAdd(false); setAddError('') }} style={{ color: 'var(--text-muted)' }}>
            <X size={16} />
          </button>
        </div>
        <div className="flex gap-2 mb-3">
          <input
            value={addUrl}
            onChange={e => setAddUrl(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && !adding && addUrl && handleAddRepo()}
            placeholder="https://host/owner/repo.git 或 git@host:owner/repo.git"
            className="input-base flex-1"
          />
          <button
            onClick={handleAddRepo}
            disabled={adding || !addUrl}
            className="btn-primary px-4 py-2 rounded-lg text-sm min-w-[72px]"
          >
            {adding ? '克隆中...' : '添加'}
          </button>
        </div>
        <p className="text-xs mb-3" style={{ color: 'var(--text-muted)' }}>首次添加会 git clone 仓库，可能需要一些时间</p>
        {addError && (
          <div
            className="flex items-start gap-2 rounded-lg px-4 py-3 text-sm"
            style={{ background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)', color: 'var(--color-error)' }}
          >
            <AlertCircle size={15} className="mt-0.5 shrink-0" />
            <span>{addError}</span>
          </div>
        )}
      </AnimatedDialog>

      {/* Import to My Skills dialog */}
      <AnimatedDialog open={showImportDialog} onClose={() => setShowImportDialog(false)} width="w-[380px]">
        <h3 className="font-semibold mb-4" style={{ color: 'var(--text-primary)' }}>选择导入分类</h3>
        <select
          value={importCategory}
          onChange={e => setImportCategory(e.target.value)}
          className="select-base mb-4"
        >
          {categories.map(c => <option key={c} value={c}>{c}</option>)}
        </select>
        <div className="flex gap-3">
          <button
            onClick={handleBatchImport}
            disabled={importing}
            className="btn-primary flex-1 py-2 rounded-lg text-sm"
          >
            {importing ? '导入中...' : `导入 ${selectedPaths.size} 个`}
          </button>
          <button onClick={() => setShowImportDialog(false)} className="btn-secondary flex-1 py-2 rounded-lg text-sm">取消</button>
        </div>
      </AnimatedDialog>

      {/* Mkdir confirmation dialog */}
      <AnimatedDialog open={showMkdirDialog} onClose={() => { setShowMkdirDialog(false); setMissingDirs([]) }} width="w-[460px]" zIndex={60}>
        <div className="flex justify-between items-center mb-1">
          <h3 className="font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
            <FolderPlus size={16} /> 目录不存在
          </h3>
          <button onClick={() => { setShowMkdirDialog(false); setMissingDirs([]) }} style={{ color: 'var(--text-muted)' }}>
            <X size={16} />
          </button>
        </div>
        <p className="text-xs mb-3" style={{ color: 'var(--text-muted)' }}>以下推送目录尚未创建，是否自动创建后继续推送？</p>
        <ul className="space-y-1.5 mb-4 max-h-40 overflow-y-auto">
          {missingDirs.map(d => (
            <li key={d.name} className="text-sm rounded-lg px-3 py-2" style={{ background: 'var(--bg-surface)' }}>
              <span className="font-medium" style={{ color: 'var(--text-primary)' }}>{d.name}</span>
              <span className="text-xs block truncate" style={{ color: 'var(--text-muted)' }} title={d.dir}>{d.dir}</span>
            </li>
          ))}
        </ul>
        <div className="flex gap-3">
          <button onClick={confirmMkdirAndPush} className="btn-primary flex-1 py-2 rounded-lg text-sm">创建并推送</button>
          <button onClick={() => { setShowMkdirDialog(false); setMissingDirs([]) }} className="btn-secondary flex-1 py-2 rounded-lg text-sm">取消</button>
        </div>
      </AnimatedDialog>

      {/* Push to tool dialog */}
      <AnimatedDialog open={showPushToolDialog} onClose={() => setShowPushToolDialog(false)} width="w-[420px]">
        <div className="flex justify-between items-center mb-1">
          <h3 className="font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
            <ArrowUpToLine size={16} /> 推送到工具
          </h3>
          <button onClick={() => setShowPushToolDialog(false)} style={{ color: 'var(--text-muted)' }}><X size={16} /></button>
        </div>
        <p className="text-xs mb-4" style={{ color: 'var(--text-muted)' }}>将 Skill 直接复制到工具目录，无需导入到「我的Skills」</p>
        {tools.length === 0 ? (
          <p className="text-sm py-4 text-center" style={{ color: 'var(--text-muted)' }}>没有可用的工具，请在设置中启用工具</p>
        ) : (
          <div className="flex flex-wrap gap-2 mb-4">
            {tools.map((t: any) => (
              <button
                key={t.name}
                onClick={() => toggleTool(t.name)}
                className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-all duration-200"
                style={toolBtnStyle(selectedTools.has(t.name))}
              >
                <ToolIcon name={t.name} size={16} />
                {t.name}
              </button>
            ))}
          </div>
        )}
        <div className="flex gap-3">
          <button
            onClick={handlePushToTools}
            disabled={pushingToTools || selectedTools.size === 0 || tools.length === 0}
            className="btn-primary flex-1 py-2 rounded-lg text-sm"
          >
            {pushingToTools ? '推送中...' : `推送到 ${selectedTools.size} 个工具`}
          </button>
          <button onClick={() => setShowPushToolDialog(false)} className="btn-secondary flex-1 py-2 rounded-lg text-sm">取消</button>
        </div>
      </AnimatedDialog>

      {/* HTTP auth dialog */}
      <AnimatedDialog open={showHttpAuthDialog} onClose={() => setShowHttpAuthDialog(false)} width="w-[460px]">
        <div className="flex justify-between items-center mb-1">
          <h3 className="font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
            <Lock size={16} /> 需要认证
          </h3>
          <button onClick={() => setShowHttpAuthDialog(false)} style={{ color: 'var(--text-muted)' }}><X size={16} /></button>
        </div>
        <p className="text-xs mb-4" style={{ color: 'var(--text-muted)' }}>仓库需要用户名和密码（或 Access Token）才能访问</p>
        <div className="space-y-2 mb-4">
          <input
            value={authUsername}
            onChange={e => setAuthUsername(e.target.value)}
            placeholder="用户名"
            className="input-base"
          />
          <input
            type="password"
            value={authPassword}
            onChange={e => setAuthPassword(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && !authAdding && handleAuthRetry()}
            placeholder="密码 / Access Token"
            className="input-base"
          />
        </div>
        {authError && (
          <div
            className="flex items-start gap-2 rounded-lg px-4 py-3 text-sm mb-3"
            style={{ background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)', color: 'var(--color-error)' }}
          >
            <AlertCircle size={15} className="mt-0.5 shrink-0" />
            <span>{authError}</span>
          </div>
        )}
        <div className="flex gap-3">
          <button onClick={handleAuthRetry} disabled={authAdding} className="btn-primary flex-1 py-2 rounded-lg text-sm">
            {authAdding ? '连接中...' : '确认'}
          </button>
          <button onClick={() => setShowHttpAuthDialog(false)} className="btn-secondary flex-1 py-2 rounded-lg text-sm">取消</button>
        </div>
      </AnimatedDialog>

      {/* SSH auth error dialog */}
      <AnimatedDialog open={showSshErrorDialog} onClose={() => setShowSshErrorDialog(false)} width="w-[460px]">
        <h3 className="font-semibold mb-2 flex items-center gap-2" style={{ color: 'var(--color-warning)' }}>
          <KeyRound size={16} /> SSH 认证失败
        </h3>
        <p className="text-sm mb-3" style={{ color: 'var(--text-secondary)' }}>无法使用 SSH 访问远程仓库，请检查以下配置：</p>
        <ul className="text-sm space-y-1.5 list-disc list-inside mb-4" style={{ color: 'var(--text-muted)' }}>
          <li>SSH 密钥是否已生成（<code style={{ color: 'var(--text-secondary)' }}>ssh-keygen</code>）</li>
          <li>公钥是否已添加到 GitHub / GitLab 等远程仓库</li>
          <li>SSH Agent 是否正在运行（<code style={{ color: 'var(--text-secondary)' }}>ssh-add</code>）</li>
          <li>可尝试改用 HTTPS 协议克隆</li>
        </ul>
        <button onClick={() => setShowSshErrorDialog(false)} className="btn-secondary w-full py-2 rounded-lg text-sm">关闭</button>
      </AnimatedDialog>

      {/* Push conflict dialog */}
      <AnimatedDialog open={showPushConflictDialog} onClose={() => setShowPushConflictDialog(false)} width="w-[420px]">
        <h3 className="font-semibold mb-2 flex items-center gap-2" style={{ color: 'var(--color-warning)' }}>
          <AlertCircle size={16} /> 发现冲突
        </h3>
        <p className="text-sm mb-3" style={{ color: 'var(--text-muted)' }}>以下 Skill 在目标工具目录中已存在：</p>
        <ul className="space-y-1 mb-4 max-h-40 overflow-y-auto">
          {pushConflicts.map(c => (
            <li key={c} className="text-sm px-3 py-1.5 rounded" style={{ background: 'var(--bg-surface)', color: 'var(--text-secondary)' }}>{c}</li>
          ))}
        </ul>
        <div className="flex gap-3">
          <button
            onClick={handlePushToToolsForce}
            className="flex-1 py-2 rounded-lg text-sm text-white transition-colors"
            style={{ background: 'var(--color-warning)' }}
          >覆盖全部</button>
          <button
            onClick={() => { setShowPushConflictDialog(false); setSelectMode(false); setSelectedPaths(new Set()); setPushConflicts([]) }}
            className="btn-secondary flex-1 py-2 rounded-lg text-sm"
          >跳过冲突</button>
        </div>
      </AnimatedDialog>
    </div>
  )
}

function RepoGrid({ repos, onEnter, onUpdate, onRemove }: {
  repos: any[]
  onEnter: (url: string) => void
  onUpdate: (url: string) => void
  onRemove: (url: string) => void
}) {
  if (repos.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-48" style={{ color: 'var(--text-muted)' }}>
        <Star size={32} className="mb-2 opacity-30" />
        <p className="text-sm">还没有收藏的仓库</p>
        <p className="text-xs mt-1">点击「添加仓库」开始收藏</p>
      </div>
    )
  }
  return (
    <div className="grid grid-cols-2 xl:grid-cols-3 gap-4">
      {repos.map((r: any) => (
        <div
          key={r.url}
          onClick={() => onEnter(r.url)}
          className="card-base p-4 cursor-pointer"
        >
          <div className="flex justify-between items-start mb-2">
            <span className="font-medium text-sm truncate flex-1 mr-2" style={{ color: 'var(--text-primary)' }}>{r.name}</span>
            <div className="flex gap-1 shrink-0" onClick={e => e.stopPropagation()}>
              <button
                onClick={() => OpenURL(r.source ? `https://${r.source}` : r.url)}
                className="p-1 rounded transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.color = 'var(--accent-primary)' }}
                onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)' }}
                title="在浏览器中打开"
              >
                <ExternalLink size={12} />
              </button>
              <button
                onClick={() => onUpdate(r.url)}
                className="p-1 rounded transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.color = 'var(--text-primary)' }}
                onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)' }}
                title="更新"
              >
                <RefreshCw size={12} />
              </button>
              <button
                onClick={() => onRemove(r.url)}
                className="p-1 rounded transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.color = 'var(--color-error)' }}
                onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)' }}
                title="删除收藏"
              >
                <Trash2 size={12} />
              </button>
            </div>
          </div>
          {r.syncError ? (
            <p className="text-xs truncate" style={{ color: 'var(--color-error)' }} title={r.syncError}>{r.syncError}</p>
          ) : (
            <>
              <p className="text-xs truncate" style={{ color: 'var(--text-muted)' }} title={r.source || r.url}>{r.source || r.url}</p>
              <p className="text-xs mt-1" style={{ color: 'var(--text-muted)' }}>
                {r.lastSync && r.lastSync !== '0001-01-01T00:00:00Z'
                  ? `同步于 ${new Date(r.lastSync).toLocaleString()}`
                  : '未同步'}
              </p>
            </>
          )}
        </div>
      ))}
    </div>
  )
}

function SkillGrid({ skills, selectMode, selectedPaths, onToggle, showRepo = false }: {
  skills: any[]
  selectMode: boolean
  selectedPaths: Set<string>
  onToggle: (path: string) => void
  showRepo?: boolean
}) {
  if (skills.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-48" style={{ color: 'var(--text-muted)' }}>
        <p className="text-sm">没有找到 Skills</p>
      </div>
    )
  }
  return (
    <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
      {skills.map((sk: any) => {
        const src = sk.source || sk.repoUrl || ''
        const sourceType = src.includes('github.com') ? 'github' : src ? 'git' : undefined
        return (
          <SyncSkillCard
            key={sk.path}
            name={sk.name}
            path={sk.path}
            source={sourceType}
            subtitle={showRepo ? sk.repoName : undefined}
            imported={sk.imported}
            showSelection={selectMode}
            selected={selectedPaths.has(sk.path)}
            onToggle={() => selectMode && onToggle(sk.path)}
          />
        )
      })}
    </div>
  )
}
