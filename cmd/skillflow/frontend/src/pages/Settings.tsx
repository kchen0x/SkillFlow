import { useEffect, useState } from 'react'
import { GetConfig, SaveConfig, ListCloudProviders, AddCustomTool, RemoveCustomTool, OpenFolderDialog, CheckAppUpdateAndNotify, GetAppVersion, GetLogDir, OpenLogDir } from '../../wailsjs/go/main/App'
import { Plus, Trash2, Settings, Globe, FolderOpen, RefreshCw, Sun, Moon } from 'lucide-react'
import { ToolIcon } from '../config/toolIcons'
import { useThemeContext } from '../contexts/ThemeContext'

type Tab = 'tools' | 'cloud' | 'general' | 'network'
type ProxyMode = 'none' | 'system' | 'manual'

function Toggle({ enabled, onToggle }: { enabled: boolean; onToggle: () => void }) {
  return (
    <div
      onClick={onToggle}
      className="w-9 h-5 rounded-full relative cursor-pointer transition-all duration-200"
      style={{
        background: enabled ? 'var(--accent-secondary)' : 'var(--bg-overlay)',
        boxShadow: enabled ? 'var(--glow-accent-sm)' : 'none',
        border: '1px solid var(--border-base)',
      }}
    >
      <div
        className={`absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform duration-200 ${enabled ? 'translate-x-4' : 'translate-x-0.5'}`}
        style={{ boxShadow: '0 1px 3px rgba(0,0,0,0.3)' }}
      />
    </div>
  )
}

export default function SettingsPage() {
  const { theme, toggleTheme } = useThemeContext()
  const [tab, setTab] = useState<Tab>('tools')
  const [cfg, setCfg] = useState<any>(null)
  const [providers, setProviders] = useState<any[]>([])
  const [saving, setSaving] = useState(false)
  const [newTool, setNewTool] = useState({ name: '', pushDir: '' })
  const [newScanDirs, setNewScanDirs] = useState<Record<string, string>>({})
  const [appVersion, setAppVersion] = useState('')
  const [logDir, setLogDir] = useState('')
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [updateResult, setUpdateResult] = useState<string | null>(null)

  useEffect(() => {
    Promise.all([GetConfig(), ListCloudProviders(), GetAppVersion(), GetLogDir()]).then(([c, p, v, logPath]) => {
      setCfg(c)
      setProviders(p ?? [])
      setAppVersion(v as string)
      setLogDir(logPath as string)
    })
  }, [])

  const checkUpdate = async () => {
    setCheckingUpdate(true)
    setUpdateResult(null)
    try {
      const info = await CheckAppUpdateAndNotify()
      if (info.hasUpdate) {
        setUpdateResult(`发现新版本 ${info.latestVersion}`)
      } else {
        setUpdateResult(`已是最新版本 (${info.currentVersion})`)
      }
    } catch (e: any) {
      setUpdateResult(`检测失败: ${e?.message ?? String(e)}`)
    } finally {
      setCheckingUpdate(false)
    }
  }

  const save = async () => {
    setSaving(true)
    await SaveConfig(cfg)
    setSaving(false)
  }

  const updateTool = (name: string, field: string, value: any) => {
    setCfg((prev: any) => ({
      ...prev,
      tools: prev.tools.map((t: any) => t.name === name ? { ...t, [field]: value } : t)
    }))
  }

  const addScanDir = (name: string) => {
    const path = (newScanDirs[name] ?? '').trim()
    if (!path) return
    setCfg((prev: any) => ({
      ...prev,
      tools: prev.tools.map((t: any) => {
        if (t.name !== name) return t
        const current = t.scanDirs ?? []
        if (current.includes(path)) return t
        return { ...t, scanDirs: [...current, path] }
      })
    }))
    setNewScanDirs((prev) => ({ ...prev, [name]: '' }))
  }

  const updateScanDir = (name: string, index: number, value: string) => {
    setCfg((prev: any) => ({
      ...prev,
      tools: prev.tools.map((t: any) => {
        if (t.name !== name) return t
        const next = [...(t.scanDirs ?? [])]
        next[index] = value
        return { ...t, scanDirs: next }
      })
    }))
  }

  const removeScanDir = (name: string, index: number) => {
    setCfg((prev: any) => ({
      ...prev,
      tools: prev.tools.map((t: any) => {
        if (t.name !== name) return t
        return { ...t, scanDirs: (t.scanDirs ?? []).filter((_: string, i: number) => i !== index) }
      })
    }))
  }

  const setProxyMode = (mode: ProxyMode) => {
    setCfg((prev: any) => ({ ...prev, proxy: { ...prev.proxy, Mode: mode } }))
  }

  const setProxyURL = (url: string) => {
    setCfg((prev: any) => ({ ...prev, proxy: { ...prev.proxy, URL: url } }))
  }

  const pickDir = async (onPick: (path: string) => void, currentPath = '') => {
    const dir = await OpenFolderDialog(currentPath)
    if (dir) onPick(dir)
  }

  const selectedProvider = providers.find((p: any) => p.name === cfg?.cloud?.provider)
  const proxyMode: ProxyMode = (cfg?.proxy?.Mode as ProxyMode) || 'none'

  if (!cfg) return <div className="p-8" style={{ color: 'var(--text-muted)' }}>加载中...</div>

  return (
    <div className="p-8 max-w-2xl">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-lg font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
          <Settings size={18} /> 设置
        </h2>
        <div className="flex items-center gap-3">
          {updateResult && (
            <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{updateResult}</span>
          )}
          {appVersion && (
            <span className="text-xs font-mono" style={{ color: 'var(--text-muted)' }}>
              {appVersion === 'dev' ? 'dev' : appVersion.startsWith('v') ? appVersion : `v${appVersion}`}
            </span>
          )}
          <button
            onClick={checkUpdate}
            disabled={checkingUpdate}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-lg transition-colors disabled:opacity-50"
            style={{ background: 'var(--bg-elevated)', color: 'var(--text-secondary)', border: '1px solid var(--border-base)' }}
          >
            <RefreshCw size={12} className={checkingUpdate ? 'animate-spin' : ''} />
            检测更新
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div
        className="flex gap-1 mb-6 rounded-xl p-1 w-fit"
        style={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-base)' }}
      >
        {([['tools', '工具路径'], ['cloud', '云存储'], ['general', '通用'], ['network', '网络']] as [Tab, string][]).map(([v, label]) => (
          <button
            key={v}
            onClick={() => setTab(v)}
            className="px-4 py-1.5 rounded-lg text-sm transition-all duration-200"
            style={tab === v ? {
              background: 'var(--bg-overlay)',
              color: 'var(--text-primary)',
              boxShadow: 'var(--glow-accent-sm)',
              border: '1px solid var(--border-accent)',
            } : {
              color: 'var(--text-muted)',
              border: '1px solid transparent',
            }}
          >{label}</button>
        ))}
      </div>

      {/* Tools tab */}
      {tab === 'tools' && (
        <div className="space-y-4">
          {(cfg.tools ?? []).map((t: any) => (
            <div
              key={t.name}
              className="rounded-xl p-4"
              style={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-base)' }}
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2.5">
                  <ToolIcon name={t.name} size={28} />
                  <span className="font-medium text-sm" style={{ color: 'var(--text-primary)' }}>{t.name}</span>
                </div>
                <label className="flex items-center gap-2 cursor-pointer">
                  <span className="text-xs" style={{ color: 'var(--text-muted)' }}>启用</span>
                  <Toggle enabled={t.enabled} onToggle={() => updateTool(t.name, 'enabled', !t.enabled)} />
                </label>
              </div>

              <div className="mb-3">
                <p className="text-xs mb-1.5" style={{ color: 'var(--text-muted)' }}>推送路径（仅 1 个）</p>
                <div className="flex gap-2">
                  <input
                    value={t.pushDir ?? ''}
                    onChange={e => updateTool(t.name, 'pushDir', e.target.value)}
                    className="input-base flex-1 font-mono"
                  />
                  <button
                    onClick={() => pickDir(dir => updateTool(t.name, 'pushDir', dir), t.pushDir ?? '')}
                    className="btn-secondary px-2.5 rounded-lg"
                    title="选择目录"
                  >
                    <FolderOpen size={14} />
                  </button>
                </div>
              </div>

              <div>
                <p className="text-xs mb-1.5" style={{ color: 'var(--text-muted)' }}>扫描路径（可多个）</p>
                <div className="space-y-2">
                  {(t.scanDirs ?? []).map((dir: string, idx: number) => (
                    <div key={`${t.name}-scan-${idx}`} className="flex gap-2">
                      <input
                        value={dir}
                        onChange={e => updateScanDir(t.name, idx, e.target.value)}
                        className="input-base flex-1 font-mono"
                      />
                      <button
                        onClick={() => pickDir(d => updateScanDir(t.name, idx, d), dir ?? '')}
                        className="btn-secondary px-2.5 rounded-lg"
                        title="选择目录"
                      >
                        <FolderOpen size={14} />
                      </button>
                      <button
                        onClick={() => removeScanDir(t.name, idx)}
                        className="btn-secondary px-2.5 rounded-lg"
                        title="删除扫描路径"
                      >
                        <Trash2 size={14} />
                      </button>
                    </div>
                  ))}
                </div>
                <div className="mt-2 flex gap-2">
                  <input
                    value={newScanDirs[t.name] ?? ''}
                    onChange={e => setNewScanDirs(prev => ({ ...prev, [t.name]: e.target.value }))}
                    placeholder="/path/to/scan"
                    className="input-base flex-1 font-mono"
                  />
                  <button
                    onClick={() => pickDir(d => setNewScanDirs(prev => ({ ...prev, [t.name]: d })), newScanDirs[t.name] ?? '')}
                    className="btn-secondary px-2.5 rounded-lg"
                    title="选择目录"
                  >
                    <FolderOpen size={14} />
                  </button>
                  <button
                    onClick={() => addScanDir(t.name)}
                    className="btn-secondary px-3 py-1.5 rounded-lg text-sm flex items-center gap-1"
                  >
                    <Plus size={14} /> 添加
                  </button>
                </div>
              </div>

              {t.custom && (
                <button
                  onClick={async () => { await RemoveCustomTool(t.name); const c = await GetConfig(); setCfg(c) }}
                  className="mt-2 text-xs flex items-center gap-1 transition-colors"
                  style={{ color: 'var(--color-error)' }}
                >
                  <Trash2 size={12} /> 删除
                </button>
              )}
            </div>
          ))}

          {/* Add custom tool */}
          <div
            className="rounded-xl p-4"
            style={{ border: '1px dashed var(--border-surface)', background: 'var(--bg-surface)' }}
          >
            <p className="text-sm mb-3" style={{ color: 'var(--text-muted)' }}>添加自定义工具</p>
            <div className="flex gap-2 mb-2">
              <input
                value={newTool.name}
                onChange={e => setNewTool(p => ({ ...p, name: e.target.value }))}
                placeholder="工具名称"
                className="input-base flex-1"
              />
            </div>
            <div className="flex gap-2">
              <input
                value={newTool.pushDir}
                onChange={e => setNewTool(p => ({ ...p, pushDir: e.target.value }))}
                placeholder="/path/to/push"
                className="input-base flex-1 font-mono"
              />
              <button
                onClick={() => pickDir(d => setNewTool(p => ({ ...p, pushDir: d })), newTool.pushDir)}
                className="btn-secondary px-2.5 rounded-lg"
                title="选择目录"
              >
                <FolderOpen size={14} />
              </button>
              <button
                onClick={async () => {
                  if (newTool.name && newTool.pushDir) {
                    await AddCustomTool(newTool.name, newTool.pushDir)
                    const c = await GetConfig(); setCfg(c)
                    setNewTool({ name: '', pushDir: '' })
                  }
                }}
                className="btn-primary px-3 py-1.5 rounded-lg text-sm flex items-center gap-1"
              >
                <Plus size={14} /> 添加
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Cloud tab */}
      {tab === 'cloud' && (
        <div className="space-y-4">
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>云厂商</p>
            <div className="flex gap-2">
              {providers.map((p: any) => (
                <button
                  key={p.name}
                  onClick={() => setCfg((prev: any) => ({ ...prev, cloud: { ...prev.cloud, provider: p.name } }))}
                  className="px-4 py-2 rounded-lg text-sm transition-all duration-200"
                  style={cfg.cloud?.provider === p.name ? {
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
                  {p.name}
                </button>
              ))}
            </div>
          </div>

          {selectedProvider && (
            <>
              {cfg.cloud?.provider !== 'git' && (
                <div>
                  <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>存储桶</p>
                  <input
                    value={cfg.cloud?.bucketName ?? ''}
                    onChange={e => setCfg((p: any) => ({ ...p, cloud: { ...p.cloud, bucketName: e.target.value } }))}
                    className="input-base"
                  />
                </div>
              )}
              {selectedProvider.fields.map((f: any) => (
                <div key={f.key}>
                  <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{f.label}</p>
                  <input
                    type={f.secret ? 'password' : 'text'}
                    placeholder={f.placeholder ?? ''}
                    value={cfg.cloud?.credentials?.[f.key] ?? ''}
                    onChange={e => setCfg((p: any) => ({
                      ...p, cloud: { ...p.cloud, credentials: { ...p.cloud?.credentials, [f.key]: e.target.value } }
                    }))}
                    className="input-base font-mono"
                  />
                </div>
              ))}
              <div>
                <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>定时自动同步间隔（分钟，0 表示仅在变更后同步）</p>
                <input
                  type="number"
                  min={0}
                  value={cfg.cloud?.syncIntervalMinutes ?? 0}
                  onChange={e => setCfg((p: any) => ({ ...p, cloud: { ...p.cloud, syncIntervalMinutes: parseInt(e.target.value) || 0 } }))}
                  className="input-base w-32"
                />
              </div>
              <label className="flex items-center gap-3 cursor-pointer">
                <Toggle
                  enabled={!!cfg.cloud?.enabled}
                  onToggle={() => setCfg((p: any) => ({ ...p, cloud: { ...p.cloud, enabled: !p.cloud?.enabled } }))}
                />
                <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>启用自动云备份</span>
              </label>
            </>
          )}
        </div>
      )}

      {/* General tab */}
      {tab === 'general' && (
        <div className="space-y-4">
          {/* Theme */}
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>外观主题</p>
            <div className="flex gap-2">
              {([['dark', '深色', <Moon size={14} />], ['light', '浅色', <Sun size={14} />]] as [string, string, React.ReactNode][]).map(([t, label, icon]) => (
                <button
                  key={t}
                  onClick={() => theme !== t && toggleTheme()}
                  className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm transition-all duration-200"
                  style={theme === t ? {
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
                  {icon} {label}
                </button>
              ))}
            </div>
          </div>

          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>日志打印级别</p>
            <div className="flex gap-2 mb-2">
              {([
                ['debug', 'Debug'],
                ['info', 'Info'],
                ['error', 'Error'],
              ] as [string, string][]).map(([level, label]) => (
                <button
                  key={level}
                  onClick={() => setCfg((p: any) => ({ ...p, logLevel: level }))}
                  className="px-3 py-1.5 rounded-lg text-sm transition-all duration-200"
                  style={(cfg.logLevel ?? 'error') === level ? {
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
                  {label}
                </button>
              ))}
            </div>
            <p className="text-xs" style={{ color: 'var(--text-muted)' }}>Debug 记录最详细，Info 记录常规信息，Error 仅记录错误。</p>
          </div>
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>日志目录</p>
            <div className="flex items-center gap-2">
              <button
                onClick={async () => { await OpenLogDir() }}
                className="btn-secondary px-3 py-1.5 rounded-lg text-sm"
              >
                打开日志目录
              </button>
              <span className="text-xs font-mono break-all" style={{ color: 'var(--text-muted)' }}>{logDir}</span>
            </div>
            <p className="mt-1.5 text-xs" style={{ color: 'var(--text-muted)' }}>日志文件最多 2 个，单文件最大 1MB，超限自动滚动覆盖旧日志。</p>
          </div>
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>本地 Skills 存储目录</p>
            <div className="flex gap-2">
              <input
                value={cfg.skillsStorageDir ?? ''}
                onChange={e => setCfg((p: any) => ({ ...p, skillsStorageDir: e.target.value }))}
                className="input-base flex-1 font-mono"
              />
              <button
                onClick={() => pickDir(d => setCfg((p: any) => ({ ...p, skillsStorageDir: d })), cfg.skillsStorageDir ?? '')}
                className="btn-secondary px-2.5 rounded-lg"
                title="选择目录"
              >
                <FolderOpen size={16} />
              </button>
            </div>
          </div>
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>从工具拉取时的默认分类</p>
            <div
              className="rounded-lg px-3 py-2 text-sm"
              style={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-base)', color: 'var(--text-secondary)' }}
            >
              Default
            </div>
            <p className="mt-1.5 text-xs" style={{ color: 'var(--text-muted)' }}>固定系统分类，用于未分类导入兜底，不可重命名或删除。</p>
          </div>
        </div>
      )}

      {/* Network tab */}
      {tab === 'network' && (
        <div className="space-y-6">
          <div>
            <p className="text-sm mb-1 flex items-center gap-1.5" style={{ color: 'var(--text-muted)' }}>
              <Globe size={14} /> 代理设置
            </p>
            <p className="text-xs mb-4" style={{ color: 'var(--text-muted)' }}>
              代理用于远程仓库相关操作（扫描仓库、安装 Skill、检查更新）
            </p>

            <div className="space-y-2">
              {([
                ['none',   '不使用代理',   '直连，不通过任何代理'],
                ['system', '使用系统代理', '读取 HTTP_PROXY / HTTPS_PROXY 环境变量'],
                ['manual', '手动配置',     '自定义代理地址'],
              ] as [ProxyMode, string, string][]).map(([mode, label, desc]) => (
                <div
                  key={mode}
                  onClick={() => setProxyMode(mode)}
                  className="flex items-start gap-3 p-3 rounded-xl cursor-pointer transition-all duration-200 select-none"
                  style={proxyMode === mode ? {
                    background: 'var(--accent-glow)',
                    border: '1px solid var(--border-accent)',
                  } : {
                    background: 'var(--bg-elevated)',
                    border: '1px solid var(--border-base)',
                  }}
                >
                  <div
                    className="mt-0.5 w-4 h-4 rounded-full border-2 flex items-center justify-center shrink-0 transition-all duration-200"
                    style={proxyMode === mode ? {
                      borderColor: 'var(--accent-secondary)',
                      background: 'var(--accent-secondary)',
                    } : {
                      borderColor: 'var(--text-muted)',
                    }}
                  >
                    {proxyMode === mode && <div className="w-1.5 h-1.5 bg-white rounded-full" />}
                  </div>
                  <div>
                    <p className="text-sm font-medium leading-snug" style={{ color: 'var(--text-primary)' }}>{label}</p>
                    <p className="text-xs mt-0.5" style={{ color: 'var(--text-muted)' }}>{desc}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {proxyMode === 'manual' && (
            <div>
              <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>代理地址</p>
              <input
                value={cfg.proxy?.URL ?? ''}
                onChange={e => setProxyURL(e.target.value)}
                placeholder="http://127.0.0.1:7890"
                className="input-base font-mono"
              />
              <p className="mt-1.5 text-xs" style={{ color: 'var(--text-muted)' }}>
                支持{' '}
                <code
                  className="px-1 rounded"
                  style={{ background: 'var(--bg-elevated)', color: 'var(--text-secondary)' }}
                >http://</code>、
                <code
                  className="px-1 rounded"
                  style={{ background: 'var(--bg-elevated)', color: 'var(--text-secondary)' }}
                >https://</code>、
                <code
                  className="px-1 rounded"
                  style={{ background: 'var(--bg-elevated)', color: 'var(--text-secondary)' }}
                >socks5://</code>{' '}格式
              </p>
            </div>
          )}
        </div>
      )}

      <button
        onClick={save}
        disabled={saving}
        className="btn-primary mt-8 px-6 py-2.5 rounded-lg text-sm"
      >
        {saving ? '保存中...' : '保存设置'}
      </button>
    </div>
  )
}
