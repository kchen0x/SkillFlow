import { useState, useEffect } from 'react'
import { BrowserRouter, Route, Routes, NavLink } from 'react-router-dom'
import { Package, ArrowUpFromLine, ArrowDownToLine, Cloud, Settings, Star, X, Download, RefreshCw } from 'lucide-react'
import Dashboard from './pages/Dashboard'
import SyncPush from './pages/SyncPush'
import SyncPull from './pages/SyncPull'
import Backup from './pages/Backup'
import SettingsPage from './pages/Settings'
import StarredRepos from './pages/StarredRepos'
import { EventsOn } from '../wailsjs/runtime/runtime'
import { DownloadAppUpdate, ApplyAppUpdate } from '../wailsjs/go/main/App'
import { main } from '../wailsjs/go/models'

type BannerState = 'idle' | 'available' | 'downloading' | 'ready_to_restart' | 'download_failed'

export default function App() {
  const [bannerState, setBannerState] = useState<BannerState>('idle')
  const [updateInfo, setUpdateInfo] = useState<main.AppUpdateInfo | null>(null)
  const [dismissed, setDismissed] = useState(false)

  useEffect(() => {
    EventsOn('app.update.available', (data: main.AppUpdateInfo) => {
      setUpdateInfo(data)
      setBannerState('available')
    })
    EventsOn('app.update.download.done', () => {
      setBannerState('ready_to_restart')
    })
    EventsOn('app.update.download.fail', () => {
      setBannerState('download_failed')
    })
  }, [])

  const handleDownload = () => {
    if (!updateInfo?.downloadUrl) return
    setBannerState('downloading')
    DownloadAppUpdate(updateInfo.downloadUrl)
  }

  const handleRestart = () => {
    ApplyAppUpdate()
  }

  const showBanner = !dismissed && bannerState !== 'idle'

  return (
    <BrowserRouter>
      <div className="flex h-screen bg-gray-950 text-gray-100 flex-col">
        {showBanner && (
          <UpdateBanner
            state={bannerState}
            info={updateInfo}
            onDownload={handleDownload}
            onRestart={handleRestart}
            onDismiss={() => setDismissed(true)}
          />
        )}
        <div className="flex flex-1 overflow-hidden">
          <aside className="w-56 bg-gray-900 border-r border-gray-800 flex flex-col p-4 gap-1">
            <h1 className="text-lg font-bold mb-6 px-2">SkillFlow</h1>
            <NavItem to="/" icon={<Package size={16} />} label="我的 Skills" />
            <p className="text-xs text-gray-500 px-2 mt-3 mb-1">同步管理</p>
            <NavItem to="/sync/push" icon={<ArrowUpFromLine size={16} />} label="推送到工具" />
            <NavItem to="/sync/pull" icon={<ArrowDownToLine size={16} />} label="从工具拉取" />
            <NavItem to="/starred" icon={<Star size={16} />} label="仓库收藏" end={false} />
            <div className="flex-1" />
            <NavItem to="/backup" icon={<Cloud size={16} />} label="云备份" />
            <NavItem to="/settings" icon={<Settings size={16} />} label="设置" />
          </aside>
          <main className="flex-1 overflow-auto">
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/sync/push" element={<SyncPush />} />
              <Route path="/sync/pull" element={<SyncPull />} />
              <Route path="/backup" element={<Backup />} />
              <Route path="/settings" element={<SettingsPage />} />
              <Route path="/starred" element={<StarredRepos />} />
              <Route path="/starred/:repoEncoded" element={<StarredRepos />} />
            </Routes>
          </main>
        </div>
      </div>
    </BrowserRouter>
  )
}

interface UpdateBannerProps {
  state: BannerState
  info: main.AppUpdateInfo | null
  onDownload: () => void
  onRestart: () => void
  onDismiss: () => void
}

function UpdateBanner({ state, info, onDownload, onRestart, onDismiss }: UpdateBannerProps) {
  return (
    <div className="flex items-center justify-between px-4 py-2 bg-indigo-700 text-white text-sm shrink-0">
      <div className="flex items-center gap-3">
        {state === 'available' && (
          <>
            <span>新版本可用: {info?.latestVersion}</span>
            {info?.canAutoUpdate ? (
              <button
                onClick={onDownload}
                className="flex items-center gap-1 px-2 py-0.5 bg-white text-indigo-700 rounded text-xs font-medium hover:bg-indigo-100"
              >
                <Download size={12} />
                立即下载
              </button>
            ) : (
              <a
                href={info?.releaseUrl}
                target="_blank"
                rel="noreferrer"
                className="flex items-center gap-1 px-2 py-0.5 bg-white text-indigo-700 rounded text-xs font-medium hover:bg-indigo-100"
              >
                查看详情
              </a>
            )}
          </>
        )}
        {state === 'downloading' && (
          <>
            <RefreshCw size={14} className="animate-spin" />
            <span>正在下载 {info?.latestVersion}...</span>
          </>
        )}
        {state === 'ready_to_restart' && (
          <>
            <span>已下载完成，点击重启以完成更新</span>
            <button
              onClick={onRestart}
              className="flex items-center gap-1 px-2 py-0.5 bg-white text-indigo-700 rounded text-xs font-medium hover:bg-indigo-100"
            >
              立即重启
            </button>
          </>
        )}
        {state === 'download_failed' && (
          <>
            <span>下载失败，请手动下载</span>
            <a
              href={info?.releaseUrl}
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-1 px-2 py-0.5 bg-white text-indigo-700 rounded text-xs font-medium hover:bg-indigo-100"
            >
              前往下载页
            </a>
          </>
        )}
      </div>
      {state !== 'downloading' && (
        <button onClick={onDismiss} className="text-white hover:text-indigo-200 ml-4">
          <X size={14} />
        </button>
      )}
    </div>
  )
}

function NavItem({ to, icon, label, end = true }: { to: string; icon: React.ReactNode; label: string; end?: boolean }) {
  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        `flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
          isActive ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:bg-gray-800 hover:text-white'
        }`
      }
    >
      {icon}
      {label}
    </NavLink>
  )
}
