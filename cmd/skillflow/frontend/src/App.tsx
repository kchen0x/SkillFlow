import { useState, useEffect } from 'react'
import { BrowserRouter, Route, Routes, NavLink, useLocation } from 'react-router-dom'
import { AnimatePresence, motion } from 'framer-motion'
import { Package, ArrowUpFromLine, ArrowDownToLine, Cloud, Settings, Star, X, Download, RefreshCw, AlertTriangle, GitMerge, MessageSquareWarning, ExternalLink, Wrench, Sun, Moon } from 'lucide-react'
import Dashboard from './pages/Dashboard'
import SyncPush from './pages/SyncPush'
import SyncPull from './pages/SyncPull'
import Backup from './pages/Backup'
import SettingsPage from './pages/Settings'
import StarredRepos from './pages/StarredRepos'
import ToolSkills from './pages/ToolSkills'
import { EventsOn } from '../wailsjs/runtime/runtime'
import { DownloadAppUpdate, ApplyAppUpdate, GetGitConflictPending, ResolveGitConflict, OpenURL, SetSkippedUpdateVersion } from '../wailsjs/go/main/App'
import { main } from '../wailsjs/go/models'
import { ThemeProvider, useThemeContext } from './contexts/ThemeContext'
import AnimatedDialog from './components/ui/AnimatedDialog'
import { pageVariants } from './lib/motionVariants'

type UpdateDialogState = 'idle' | 'available' | 'downloading' | 'ready_to_restart' | 'download_failed'

type GitConflictInfo = {
  message: string
  files: string[]
}

const feedbackIssueURL = 'https://github.com/shinerio/skillflow/issues/new/choose'

function parseConflictPayload(data: string): GitConflictInfo {
  try {
    const parsed = JSON.parse(data)
    if (typeof parsed === 'string') return { message: parsed, files: [] }
    return {
      message: parsed?.message ?? '',
      files: Array.isArray(parsed?.files) ? parsed.files.filter((f: any) => typeof f === 'string' && f.trim() !== '') : [],
    }
  } catch {
    return { message: data, files: [] }
  }
}

function parseAppUpdatePayload(data: unknown): main.AppUpdateInfo {
  return main.AppUpdateInfo.createFrom(data)
}

function AppContent() {
  const { theme, toggleTheme } = useThemeContext()
  const [dialogState, setDialogState] = useState<UpdateDialogState>('idle')
  const [updateInfo, setUpdateInfo] = useState<main.AppUpdateInfo | null>(null)

  const [conflictOpen, setConflictOpen] = useState(false)
  const [conflictInfo, setConflictInfo] = useState<GitConflictInfo>({ message: '', files: [] })
  const [resolving, setResolving] = useState(false)
  const [resolveError, setResolveError] = useState('')

  const handleResolve = async (useLocal: boolean) => {
    setResolving(true)
    setResolveError('')
    try {
      await ResolveGitConflict(useLocal)
      setConflictOpen(false)
    } catch (e: any) {
      setResolveError(String(e?.message ?? e ?? '操作失败，请重试'))
    } finally {
      setResolving(false)
    }
  }

  useEffect(() => {
    EventsOn('app.update.available', (data: unknown) => {
      setUpdateInfo(parseAppUpdatePayload(data))
      setDialogState('available')
    })
    EventsOn('app.update.download.done', () => {
      setDialogState('ready_to_restart')
    })
    EventsOn('app.update.download.fail', () => {
      setDialogState('download_failed')
    })
    EventsOn('git.conflict', (data: string) => {
      setConflictInfo(parseConflictPayload(data))
      setResolveError('')
      setConflictOpen(true)
    })
    GetGitConflictPending().then(pending => { if (pending) setConflictOpen(true) })
  }, [])

  const handleDownload = () => {
    if (!updateInfo?.downloadUrl) return
    setDialogState('downloading')
    DownloadAppUpdate(updateInfo.downloadUrl)
  }

  const handleRestart = () => {
    ApplyAppUpdate()
  }

  const handleSkip = async () => {
    if (updateInfo?.latestVersion) {
      await SetSkippedUpdateVersion(updateInfo.latestVersion)
    }
    setDialogState('idle')
  }

  const handleOpenRelease = () => {
    const releaseURL = updateInfo?.releaseUrl || 'https://github.com/shinerio/SkillFlow/releases/latest'
    if (releaseURL) {
      OpenURL(releaseURL)
    }
    setDialogState('idle')
  }

  return (
    <div
      className="flex h-screen flex-col relative"
      style={{ backgroundColor: 'var(--bg-base)', color: 'var(--text-primary)' }}
    >
      {/* Git conflict dialog */}
      <AnimatedDialog open={conflictOpen} width="w-[420px]" zIndex={50}>
        <div className="flex items-center gap-2 mb-3">
          <AlertTriangle size={18} style={{ color: 'var(--color-warning)' }} />
          <span className="font-semibold text-base">Git 同步冲突</span>
        </div>
        <p className="text-sm mb-2" style={{ color: 'var(--text-secondary)' }}>
          本地 Skills 与远端仓库存在冲突，请选择以哪一方为准：
        </p>
        {conflictInfo.files.length > 0 && (
          <div className="mb-3">
            <p className="text-xs mb-1.5" style={{ color: 'var(--text-muted)' }}>冲突相关文件（{conflictInfo.files.length}）</p>
            <div
              className="max-h-28 overflow-y-auto rounded-lg px-2 py-1.5"
              style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-base)' }}
            >
              {conflictInfo.files.slice(0, 30).map((f, i) => (
                <div key={`${f}-${i}`} className="font-mono text-[11px] truncate" style={{ color: 'var(--text-secondary)' }}>{f}</div>
              ))}
              {conflictInfo.files.length > 30 && (
                <div className="text-[11px]" style={{ color: 'var(--text-muted)' }}>... 还有 {conflictInfo.files.length - 30} 个文件</div>
              )}
            </div>
          </div>
        )}
        {conflictInfo.message && (
          <div
            className="mb-3 rounded-lg px-2 py-1.5"
            style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-base)' }}
          >
            <p className="text-[11px] mb-1" style={{ color: 'var(--text-muted)' }}>Git 输出</p>
            <pre className="text-[11px] whitespace-pre-wrap break-all max-h-20 overflow-y-auto" style={{ color: 'var(--text-secondary)' }}>{conflictInfo.message}</pre>
          </div>
        )}
        <ul className="text-xs list-disc list-inside mb-6 space-y-1" style={{ color: 'var(--text-muted)' }}>
          <li><span className="font-medium" style={{ color: 'var(--text-primary)' }}>以本地为准</span> — 保留本地内容，强制推送到远端</li>
          <li><span className="font-medium" style={{ color: 'var(--text-primary)' }}>以远端为准</span> — 丢弃本地冲突部分，恢复为远端内容</li>
        </ul>
        {resolveError && (
          <p
            className="mb-3 text-xs rounded-lg px-3 py-2 break-all"
            style={{ color: 'var(--color-error)', background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)' }}
          >{resolveError}</p>
        )}
        <div className="flex gap-3 justify-end">
          <button
            onClick={() => handleResolve(false)}
            disabled={resolving}
            className="btn-secondary flex items-center gap-1.5 px-4 py-2 text-sm rounded-lg"
          >
            {resolving ? <RefreshCw size={13} className="animate-spin" /> : <Download size={13} />}
            以远端为准
          </button>
          <button
            onClick={() => handleResolve(true)}
            disabled={resolving}
            className="btn-primary flex items-center gap-1.5 px-4 py-2 text-sm rounded-lg"
          >
            {resolving ? <RefreshCw size={13} className="animate-spin" /> : <GitMerge size={13} />}
            以本地为准
          </button>
        </div>
      </AnimatedDialog>

      {/* Update dialog */}
      <AnimatedDialog open={dialogState !== 'idle'} width="w-[440px]" zIndex={50}>
        <UpdateDialogContent
          state={dialogState}
          info={updateInfo}
          onDownload={handleDownload}
          onRestart={handleRestart}
          onOpenRelease={handleOpenRelease}
          onSkip={handleSkip}
          onClose={() => setDialogState('idle')}
        />
      </AnimatedDialog>

      <div className="flex flex-1 overflow-hidden relative">
        {/* Sidebar */}
        <aside
          className="w-56 flex flex-col p-4 gap-1 relative"
          style={{
            background: 'var(--bg-surface)',
            borderRight: '2px solid var(--sidebar-border)',
          }}
        >
          {/* Top glow divider */}
          <div
            className="absolute top-0 left-0 right-0 h-px"
            style={{ background: 'linear-gradient(90deg, transparent, var(--accent-primary), transparent)', opacity: 0.4 }}
          />
          <div className="flex items-center justify-between mb-6 px-2">
            <h1
              className="text-lg font-bold"
              style={{ color: 'var(--accent-primary)', textShadow: '0 0 12px var(--accent-glow)' }}
            >
              SkillFlow
            </h1>
            <button
              onClick={toggleTheme}
              className="p-1.5 rounded-lg transition-colors"
              style={{ color: 'var(--text-muted)' }}
              title={theme === 'dark' ? '切换到亮色模式' : '切换到暗色模式'}
            >
              {theme === 'dark' ? <Sun size={14} /> : <Moon size={14} />}
            </button>
          </div>
          <NavItem to="/" icon={<Package size={16} />} label="我的 Skills" />
          <NavItem to="/tools" icon={<Wrench size={16} />} label="我的工具" end={false} />
          <p className="text-xs px-2 mt-3 mb-1" style={{ color: 'var(--text-muted)' }}>同步管理</p>
          <NavItem to="/sync/push" icon={<ArrowUpFromLine size={16} />} label="推送到工具" />
          <NavItem to="/sync/pull" icon={<ArrowDownToLine size={16} />} label="从工具拉取" />
          <NavItem to="/starred" icon={<Star size={16} />} label="仓库收藏" end={false} />
          <div className="flex-1" />
          <div className="flex flex-col gap-1">
            <NavItem to="/backup" icon={<Cloud size={16} />} label="云备份" />
            <NavItem to="/settings" icon={<Settings size={16} />} label="设置" />
            <button
              onClick={() => OpenURL(feedbackIssueURL)}
              className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => {
                e.currentTarget.style.backgroundColor = 'var(--bg-hover)'
                e.currentTarget.style.color = 'var(--text-primary)'
              }}
              onMouseLeave={e => {
                e.currentTarget.style.backgroundColor = ''
                e.currentTarget.style.color = 'var(--text-muted)'
              }}
            >
              <MessageSquareWarning size={16} />
              意见反馈
            </button>
          </div>
        </aside>

        {/* Main content with page transitions */}
        <main className="flex-1 overflow-auto relative">
          <AnimatedRoutes />
        </main>
      </div>
    </div>
  )
}

function AnimatedRoutes() {
  const location = useLocation()
  return (
    <AnimatePresence mode="wait">
      <motion.div
        key={location.pathname}
        variants={pageVariants}
        initial="initial"
        animate="animate"
        exit="exit"
        className="h-full"
      >
        <Routes location={location}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/sync/push" element={<SyncPush />} />
          <Route path="/sync/pull" element={<SyncPull />} />
          <Route path="/backup" element={<Backup />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/starred" element={<StarredRepos />} />
          <Route path="/starred/:repoEncoded" element={<StarredRepos />} />
          <Route path="/tools" element={<ToolSkills />} />
        </Routes>
      </motion.div>
    </AnimatePresence>
  )
}

export default function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <AppContent />
      </BrowserRouter>
    </ThemeProvider>
  )
}

interface UpdateDialogContentProps {
  state: UpdateDialogState
  info: main.AppUpdateInfo | null
  onDownload: () => void
  onRestart: () => void
  onOpenRelease: () => void
  onSkip: () => void
  onClose: () => void
}

function UpdateDialogContent({ state, info, onDownload, onRestart, onOpenRelease, onSkip, onClose }: UpdateDialogContentProps) {
  const isDownloading = state === 'downloading'

  return (
    <>
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Download size={18} style={{ color: 'var(--accent-primary)' }} />
          <span className="font-semibold text-base">
            {state === 'ready_to_restart' ? '更新已就绪' : state === 'download_failed' ? '下载失败' : '发现新版本'}
          </span>
        </div>
        {!isDownloading && (
          <button
            onClick={onClose}
            style={{ color: 'var(--text-muted)' }}
            className="hover:opacity-80 transition-opacity"
          >
            <X size={16} />
          </button>
        )}
      </div>

      {(state === 'available' || state === 'downloading') && (
        <>
          <p className="text-sm mb-1" style={{ color: 'var(--text-secondary)' }}>
            最新版本：<span className="font-mono font-medium" style={{ color: 'var(--accent-primary)' }}>{info?.latestVersion}</span>
          </p>
          <p className="text-sm mb-4" style={{ color: 'var(--text-muted)' }}>
            当前版本：<span className="font-mono">{info?.currentVersion}</span>
          </p>
          {info?.releaseNotes && (
            <div
              className="mb-4 rounded-lg px-3 py-2 max-h-32 overflow-y-auto"
              style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-base)' }}
            >
              <p className="text-[11px] mb-1" style={{ color: 'var(--text-muted)' }}>更新说明</p>
              <pre className="text-xs whitespace-pre-wrap break-all" style={{ color: 'var(--text-secondary)' }}>{info.releaseNotes}</pre>
            </div>
          )}
        </>
      )}

      {state === 'downloading' && (
        <div className="flex items-center gap-2 mb-4 text-sm" style={{ color: 'var(--text-secondary)' }}>
          <RefreshCw size={14} className="animate-spin" style={{ color: 'var(--accent-primary)' }} />
          <span>正在下载 {info?.latestVersion}，请稍候...</span>
        </div>
      )}

      {state === 'ready_to_restart' && (
        <p className="text-sm mb-4" style={{ color: 'var(--text-secondary)' }}>
          新版本已下载完成，点击下方按钮重启应用以完成更新。
        </p>
      )}

      {state === 'download_failed' && (
        <p className="text-sm mb-4" style={{ color: 'var(--text-secondary)' }}>
          自动下载失败，请前往 Release 页面手动下载最新版本。
        </p>
      )}

      {state === 'available' && (
        <div className="flex flex-col gap-2">
          {info?.canAutoUpdate && (
            <button
              onClick={onDownload}
              className="btn-primary flex items-center justify-center gap-2 w-full px-4 py-2.5 rounded-xl text-sm font-medium"
            >
              <Download size={14} />
              下载并自动重启更新
            </button>
          )}
          <button
            onClick={onOpenRelease}
            className="btn-secondary flex items-center justify-center gap-2 w-full px-4 py-2.5 rounded-xl text-sm"
          >
            <ExternalLink size={14} />
            前往 Release 页面手动下载
          </button>
          <button
            onClick={onSkip}
            className="flex items-center justify-center gap-2 w-full px-4 py-2 text-sm transition-colors"
            style={{ color: 'var(--text-muted)' }}
          >
            跳过此版本（下次启动不再提示）
          </button>
        </div>
      )}

      {state === 'downloading' && (
        <p className="text-xs text-center" style={{ color: 'var(--text-muted)' }}>下载完成后将自动提示重启</p>
      )}

      {state === 'ready_to_restart' && (
        <div className="flex gap-3 justify-end">
          <button onClick={onClose} className="btn-secondary px-4 py-2 text-sm rounded-xl">
            稍后重启
          </button>
          <button onClick={onRestart} className="btn-primary flex items-center gap-2 px-4 py-2 text-sm rounded-xl">
            <RefreshCw size={13} />
            立即重启
          </button>
        </div>
      )}

      {state === 'download_failed' && (
        <div className="flex gap-3 justify-end">
          <button onClick={onClose} className="btn-secondary px-4 py-2 text-sm rounded-xl">
            关闭
          </button>
          <button onClick={onOpenRelease} className="btn-primary flex items-center gap-2 px-4 py-2 text-sm rounded-xl">
            <ExternalLink size={13} />
            前往下载页
          </button>
        </div>
      )}
    </>
  )
}

function NavItem({ to, icon, label, end = true }: { to: string; icon: React.ReactNode; label: string; end?: boolean }) {
  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        `flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-all duration-200 ${
          isActive ? 'nav-active' : 'nav-inactive'
        }`
      }
      style={({ isActive }) => isActive ? {
        backgroundColor: 'var(--accent-glow)',
        color: 'var(--accent-primary)',
        border: '1px solid var(--border-accent)',
        boxShadow: 'var(--glow-accent-sm)',
      } : {
        color: 'var(--text-muted)',
        border: '1px solid transparent',
      }}
      onMouseEnter={e => {
        const el = e.currentTarget
        if (!el.classList.contains('nav-active')) {
          el.style.backgroundColor = 'var(--bg-hover)'
          el.style.color = 'var(--text-primary)'
        }
      }}
      onMouseLeave={e => {
        const el = e.currentTarget
        if (!el.classList.contains('nav-active')) {
          el.style.backgroundColor = ''
          el.style.color = 'var(--text-muted)'
        }
      }}
    >
      {icon}
      {label}
    </NavLink>
  )
}
