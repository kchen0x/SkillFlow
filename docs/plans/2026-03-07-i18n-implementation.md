# i18n (Chinese / English) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add bilingual (zh/en) support to the SkillFlow frontend via a lightweight custom i18n context, with a language toggle in both the sidebar and Settings > General.

**Architecture:** Create `src/i18n/` with `zh.ts`, `en.ts`, and `index.ts` (flat key-value translation maps + `t(key, vars?)` interpolation). Wrap the app in `LanguageProvider` that stores language preference in `localStorage`. Replace all hardcoded Chinese strings across 14 files with `t()` calls.

**Tech Stack:** React context (same pattern as ThemeContext), TypeScript, no new npm dependencies.

---

## Task 1: Create i18n infrastructure

**Files:**
- Create: `cmd/skillflow/frontend/src/i18n/index.ts`
- Create: `cmd/skillflow/frontend/src/i18n/zh.ts`
- Create: `cmd/skillflow/frontend/src/i18n/en.ts`
- Create: `cmd/skillflow/frontend/src/contexts/LanguageContext.tsx`

### Step 1: Create `src/i18n/zh.ts`

```ts
export const zh = {
  // Nav
  'nav.mySkills': '我的 Skills',
  'nav.myTools': '我的工具',
  'nav.syncSection': '同步管理',
  'nav.pushToTool': '推送到工具',
  'nav.pullFromTool': '从工具拉取',
  'nav.starred': '仓库收藏',
  'nav.backup': '云备份',
  'nav.settings': '设置',
  'nav.feedback': '意见反馈',
  'nav.switchLangTitle': '切换到英文',
  'nav.switchThemeTitle': '切换到 {theme} 主题',

  // Common
  'common.save': '保存',
  'common.cancel': '取消',
  'common.delete': '删除',
  'common.close': '关闭',
  'common.confirm': '确认',
  'common.loading': '加载中...',
  'common.saving': '保存中...',
  'common.all': '全部',
  'common.selectAll': '全选',
  'common.deselectAll': '取消全选',
  'common.add': '添加',
  'common.skip': '跳过',
  'common.overwrite': '覆盖',
  'common.gotIt': '我知道了',
  'common.showingNSkills': '当前显示 {count} 个 Skill',

  // Update dialog
  'update.newVersion': '发现新版本',
  'update.ready': '更新已就绪',
  'update.failed': '下载失败',
  'update.latestLabel': '最新版本：',
  'update.currentLabel': '当前版本：',
  'update.notes': '更新说明',
  'update.download': '下载并自动重启更新',
  'update.openRelease': '前往 Release 页面手动下载',
  'update.skip': '跳过此版本（下次启动不再提示）',
  'update.downloading': '正在下载 {version}，请稍候...',
  'update.downloadHint': '下载完成后将自动提示重启',
  'update.restartNow': '立即重启',
  'update.restartLater': '稍后重启',
  'update.restartDesc': '新版本已下载完成，点击下方按钮重启应用以完成更新。',
  'update.downloadFailDesc': '自动下载失败，请前往 Release 页面手动下载最新版本。',
  'update.goDownload': '前往下载页',

  // Git conflict dialog
  'conflict.title': 'Git 同步冲突',
  'conflict.desc': '本地 Skills 与远端仓库存在冲突，请选择以哪一方为准：',
  'conflict.filesLabel': '冲突相关文件（{count}）',
  'conflict.output': 'Git 输出',
  'conflict.keepLocal': '以本地为准',
  'conflict.keepRemote': '以远端为准',
  'conflict.keepLocalDesc': '保留本地内容，强制推送到远端',
  'conflict.keepRemoteDesc': '丢弃本地冲突部分，恢复为远端内容',
  'conflict.moreFiles': '... 还有 {count} 个文件',

  // Dashboard
  'dashboard.searchPlaceholder': '搜索 Skills...',
  'dashboard.update': '更新',
  'dashboard.batchDelete': '批删',
  'dashboard.import': '导入',
  'dashboard.remoteInstall': '远程安装',
  'dashboard.dropToImport': '松开以导入 Skill',
  'dashboard.empty': '没有找到 Skills',
  'dashboard.emptyHint': '从远程仓库安装或拖拽文件夹到此处',

  // SyncPush
  'syncPush.title': '推送到工具',
  'syncPush.pushRange': '推送范围',
  'syncPush.targetTool': '目标工具',
  'syncPush.pushAll': '推送全部',
  'syncPush.pushCategory': '推送当前分类',
  'syncPush.manualSelect': '手动选择 Skill',
  'syncPush.selectAllList': '全选当前列表',
  'syncPush.nSkillsVisible': '当前可选 {count} 个 Skill',
  'syncPush.startPush': '开始推送 ({count})',
  'syncPush.pushing': '推送中...',
  'syncPush.done': '推送完成',
  'syncPush.empty': '当前范围内没有 Skill',
  'syncPush.emptyHint': '选择"全部"或切换到其他分类后再试',
  'syncPush.scopeManual': '手动选择 {count} 个 Skill',
  'syncPush.scopeAll': '全部 Skills ({count})',
  'syncPush.scopeCategory': '分类「{cat}」({count})',
  'syncPush.searchPlaceholder': '搜索待推送的 Skills...',
  'syncPush.mkdirTitle': '目录不存在',
  'syncPush.mkdirDesc': '以下推送目录尚未创建，是否自动创建后继续推送？',
  'syncPush.createAndPush': '创建并推送',

  // SyncPull
  'syncPull.title': '从工具拉取',
  'syncPull.importCategory': '导入分类',
  'syncPull.sourceTool': '来源工具',
  'syncPull.scanning': '扫描中...',
  'syncPull.emptyWarning': '未发现任何 Skill，请确认工具目录中包含含有 skill.md 的子目录',
  'syncPull.selectSkills': '选择要导入的 Skills',
  'syncPull.importTo': '导入到分类「{cat}」',
  'syncPull.startPull': '开始拉取 ({count})',
  'syncPull.pulling': '拉取中...',
  'syncPull.done': '拉取完成 ✓',
  'syncPull.noMatch': '当前筛选下没有匹配的 Skill',
  'syncPull.searchPlaceholder': '搜索扫描到的 Skills...',

  // StarredRepos
  'starred.title': '仓库收藏',
  'starred.addRepo': '添加仓库',
  'starred.updateAll': '全部更新',
  'starred.batchImport': '批量导入',
  'starred.folder': '文件夹',
  'starred.flat': '平铺',
  'starred.addRepoTitle': '添加远程仓库',
  'starred.addHint': '首次添加会 git clone 仓库，可能需要一些时间',
  'starred.cloning': '克隆中...',
  'starred.addBtn': '添加',
  'starred.emptyTitle': '还没有收藏的仓库',
  'starred.emptyHint': '点击「添加仓库」开始收藏',
  'starred.noSkills': '没有找到 Skills',
  'starred.selectCategory': '选择导入分类',
  'starred.importing': '导入中...',
  'starred.importN': '导入 {count} 个',
  'starred.pushToTools': '推送到工具',
  'starred.pushingToTools': '推送中...',
  'starred.pushToNTools': '推送到 {count} 个工具',
  'starred.noTools': '没有可用的工具，请在设置中启用工具',
  'starred.mkdirTitle': '目录不存在',
  'starred.mkdirDesc': '以下推送目录尚未创建，是否自动创建后继续推送？',
  'starred.createAndPush': '创建并推送',
  'starred.authTitle': '需要认证',
  'starred.authDesc': '仓库需要用户名和密码（或 Access Token）才能访问',
  'starred.username': '用户名',
  'starred.password': '密码 / Access Token',
  'starred.connecting': '连接中...',
  'starred.sshTitle': 'SSH 认证失败',
  'starred.sshDesc': '无法使用 SSH 访问远程仓库，请检查以下配置：',
  'starred.conflictsTitle': '发现冲突',
  'starred.conflictsDesc': '以下 Skill 在目标工具目录中已存在：',
  'starred.overwriteAll': '覆盖全部',
  'starred.skipConflicts': '跳过冲突',
  'starred.syncAt': '同步于',
  'starred.notSynced': '未同步',
  'starred.openInBrowser': '在浏览器中打开',
  'starred.updateBtn': '更新',
  'starred.removeStarred': '删除收藏',
  'starred.searchCurrentRepo': '搜索当前仓库中的 Skills...',
  'starred.searchAllRepos': '搜索收藏仓库中的 Skills...',
  'starred.importToMySkills': '导入到我的Skills',
  'starred.pushToToolsCount': '推送到工具',
  'starred.successMsg': '已成功推送 {count} 个 Skill 到 {toolCount} 个工具',

  // ToolSkills
  'toolSkills.title': '我的工具',
  'toolSkills.toolList': '工具列表',
  'toolSkills.noTools': '没有启用的工具，请在设置中启用',
  'toolSkills.selectToolFirst': '请先在左侧选择一个工具',
  'toolSkills.batchDelete': '批量删除',
  'toolSkills.pushPath': '推送路径',
  'toolSkills.noPushDir': '未配置',
  'toolSkills.noPushDirDesc': '该工具未配置推送路径',
  'toolSkills.noPushSkills': '推送路径下暂无 Skill',
  'toolSkills.noMatch': '当前筛选下没有匹配的 Skill',
  'toolSkills.scanPath': '扫描路径',
  'toolSkills.nDirs': '{count} 个目录',
  'toolSkills.noScanSkills': '扫描路径下暂无独立 Skill',
  'toolSkills.delete': '删除',
  'toolSkills.readOnly': '只读',
  'toolSkills.copySkill': '复制 skill.md',
  'toolSkills.openDir': '打开目录',
  'toolSkills.searchPlaceholder': '搜索当前工具中的 Skills...',

  // Backup
  'backup.title': '云备份',
  'backup.notEnabled': '云备份未启用。请前往设置 → 云存储完成配置。',
  'backup.backupNow': '立即备份',
  'backup.backingUp': '备份中 {file}',
  'backup.pullRemote': '拉取远端',
  'backup.restore': '从云端恢复',
  'backup.pulling': '拉取中...',
  'backup.refresh': '刷新',
  'backup.gitDone': 'Git 同步完成',
  'backup.done': '备份完成',
  'backup.gitFailed': 'Git 同步失败，请检查仓库配置',
  'backup.failed': '备份失败，请检查云存储配置',
  'backup.gitFiles': 'Git 跟踪文件',
  'backup.cloudFiles': '云端文件',

  // Settings
  'settings.title': '设置',
  'settings.checkUpdate': '检测更新',
  'settings.checkingUpdate': '检测中...',
  'settings.tabTools': '工具路径',
  'settings.tabCloud': '云存储',
  'settings.tabGeneral': '通用',
  'settings.tabNetwork': '网络',
  'settings.toolEnabled': '启用',
  'settings.pushPath': '推送路径（仅 1 个）',
  'settings.scanPaths': '扫描路径（可多个）',
  'settings.selectDir': '选择目录',
  'settings.deleteScanPath': '删除扫描路径',
  'settings.addPath': '添加',
  'settings.deleteTool': '删除',
  'settings.addCustomTool': '添加自定义工具',
  'settings.toolName': '工具名称',
  'settings.cloudProvider': '云厂商',
  'settings.bucket': '存储桶',
  'settings.syncInterval': '定时自动同步间隔（分钟，0 表示仅在变更后同步）',
  'settings.enableAutoBackup': '启用自动云备份',
  'settings.theme': '外观主题',
  'settings.themeHint': '旧版「浅色」会自动迁移到 Young；侧边栏右上角快捷按钮会按 Dark → Young → Light 顺序轮换。',
  'settings.themeDark': '重做为石墨黑与雾蓝点缀，更安静、更耐看的暗色工作氛围。',
  'settings.themeYoung': '保留原浅色的轻盈蓝调，但收敛成更柔和的纸感与晨光层次。',
  'settings.themeLight': '新增低饱和灰白主题，参考图中的侧栏灰蓝与苹果蓝操作色。',
  'settings.logLevel': '日志打印级别',
  'settings.logLevelHint': 'Debug 记录最详细，Info 记录常规信息，Error 仅记录错误。',
  'settings.logDir': '日志目录',
  'settings.openLogDir': '打开日志目录',
  'settings.logDirHint': '日志文件最多 2 个，单文件最大 1MB，超限自动滚动覆盖旧日志。',
  'settings.skillsDir': '本地 Skills 存储目录',
  'settings.defaultCategory': '从工具拉取时的默认分类',
  'settings.defaultCategoryHint': '固定系统分类，用于未分类导入兜底，不可重命名或删除。',
  'settings.proxy': '代理设置',
  'settings.proxyHint': '代理用于远程仓库相关操作（扫描仓库、安装 Skill、检查更新）',
  'settings.proxyNone': '不使用代理',
  'settings.proxyNoneDesc': '直连，不通过任何代理',
  'settings.proxySystem': '使用系统代理',
  'settings.proxySystemDesc': '读取 HTTP_PROXY / HTTPS_PROXY 环境变量',
  'settings.proxyManual': '手动配置',
  'settings.proxyManualDesc': '自定义代理地址',
  'settings.proxyUrl': '代理地址',
  'settings.proxyUrlHint': '支持 http://、https://、socks5:// 格式',
  'settings.saveSettings': '保存设置',
  'settings.language': '语言',
  'settings.updateFound': '发现新版本 {version}',
  'settings.updateLatest': '已是最新版本 ({version})',
  'settings.updateFailed': '检测失败: {msg}',

  // CategoryPanel
  'category.all': '全部',
  'category.newCategory': '新建分类',
  'category.rename': '重命名',
  'category.delete': '删除',
  'category.cannotDelete': '无法删除分类',
  'category.hasSkills': '分类「{cat}」下还有 {count} 个 Skill，请先清空该分类后再删除。',
  'category.hasSkillsGeneric': '分类「{cat}」下还有 Skill，请先清空该分类后再删除。',
  'category.deleteFailed': '删除失败',

  // SkillCard
  'skillCard.update': '更新',
  'skillCard.copy': '复制',
  'skillCard.copied': '已复制',
  'skillCard.delete': '删除',
  'skillCard.moveTo': '移动到 {cat}',
  'skillCard.openDir': '打开目录',

  // ConflictDialog
  'conflictDialog.title': '冲突检测',
  'conflictDialog.desc': '{name} 已存在，如何处理？',
  'conflictDialog.skip': '跳过',
  'conflictDialog.overwrite': '覆盖',

  // GitHubInstallDialog
  'github.title': '从远程仓库安装',
  'github.scanning': '克隆/更新中...',
  'github.scan': '扫描',
  'github.hint': '首次扫描会 git clone 仓库，后续自动 git pull 更新',
  'github.noSkills': '未发现任何 Skill，请确认该仓库包含含有 skill.md 的子目录',
  'github.installTo': '安装到分类',
  'github.installing': '安装中...',
  'github.installN': '安装 {count} 个 Skill',
  'github.installed': '已安装',

  // SkillListControls
  'skillList.sortAscTitle': '按名称首字母正序排序',
  'skillList.sortDescTitle': '按名称首字母逆序排序',
  'skillList.searchDefault': '搜索 Skills...',
} as const
```

### Step 2: Create `src/i18n/en.ts`

```ts
export const en = {
  // Nav
  'nav.mySkills': 'My Skills',
  'nav.myTools': 'My Tools',
  'nav.syncSection': 'Sync',
  'nav.pushToTool': 'Push to Tool',
  'nav.pullFromTool': 'Pull from Tool',
  'nav.starred': 'Starred Repos',
  'nav.backup': 'Cloud Backup',
  'nav.settings': 'Settings',
  'nav.feedback': 'Feedback',
  'nav.switchLangTitle': 'Switch to Chinese',
  'nav.switchThemeTitle': 'Switch to {theme} theme',

  // Common
  'common.save': 'Save',
  'common.cancel': 'Cancel',
  'common.delete': 'Delete',
  'common.close': 'Close',
  'common.confirm': 'Confirm',
  'common.loading': 'Loading...',
  'common.saving': 'Saving...',
  'common.all': 'All',
  'common.selectAll': 'Select All',
  'common.deselectAll': 'Deselect All',
  'common.add': 'Add',
  'common.skip': 'Skip',
  'common.overwrite': 'Overwrite',
  'common.gotIt': 'Got It',
  'common.showingNSkills': 'Showing {count} Skills',

  // Update dialog
  'update.newVersion': 'Update Available',
  'update.ready': 'Ready to Restart',
  'update.failed': 'Download Failed',
  'update.latestLabel': 'Latest: ',
  'update.currentLabel': 'Current: ',
  'update.notes': 'Release Notes',
  'update.download': 'Download & Auto-restart',
  'update.openRelease': 'Open Release Page',
  'update.skip': 'Skip this version (don\'t remind on next start)',
  'update.downloading': 'Downloading {version}, please wait...',
  'update.downloadHint': 'You will be prompted to restart after download',
  'update.restartNow': 'Restart Now',
  'update.restartLater': 'Later',
  'update.restartDesc': 'The update has been downloaded. Click below to restart and apply it.',
  'update.downloadFailDesc': 'Auto-download failed. Please visit the Release page to download manually.',
  'update.goDownload': 'Go to Downloads',

  // Git conflict dialog
  'conflict.title': 'Git Sync Conflict',
  'conflict.desc': 'Local Skills conflict with remote. Choose which version to keep:',
  'conflict.filesLabel': 'Conflict files ({count})',
  'conflict.output': 'Git Output',
  'conflict.keepLocal': 'Keep Local',
  'conflict.keepRemote': 'Keep Remote',
  'conflict.keepLocalDesc': 'Keep local content and force-push to remote',
  'conflict.keepRemoteDesc': 'Discard local changes and restore remote content',
  'conflict.moreFiles': '... and {count} more files',

  // Dashboard
  'dashboard.searchPlaceholder': 'Search Skills...',
  'dashboard.update': 'Update',
  'dashboard.batchDelete': 'Batch Del',
  'dashboard.import': 'Import',
  'dashboard.remoteInstall': 'Remote Install',
  'dashboard.dropToImport': 'Drop to Import Skill',
  'dashboard.empty': 'No Skills Found',
  'dashboard.emptyHint': 'Install from a remote repo or drag a folder here',

  // SyncPush
  'syncPush.title': 'Push to Tool',
  'syncPush.pushRange': 'Push Scope',
  'syncPush.targetTool': 'Target Tool',
  'syncPush.pushAll': 'Push All',
  'syncPush.pushCategory': 'Push Current Category',
  'syncPush.manualSelect': 'Manual Select',
  'syncPush.selectAllList': 'Select All in List',
  'syncPush.nSkillsVisible': '{count} Skills available',
  'syncPush.startPush': 'Start Push ({count})',
  'syncPush.pushing': 'Pushing...',
  'syncPush.done': 'Push Complete',
  'syncPush.empty': 'No Skills in this scope',
  'syncPush.emptyHint': 'Select "All" or switch to another category',
  'syncPush.scopeManual': '{count} Skills selected',
  'syncPush.scopeAll': 'All Skills ({count})',
  'syncPush.scopeCategory': 'Category "{cat}" ({count})',
  'syncPush.searchPlaceholder': 'Search Skills to push...',
  'syncPush.mkdirTitle': 'Directory Not Found',
  'syncPush.mkdirDesc': 'The following push directories don\'t exist. Create them and continue?',
  'syncPush.createAndPush': 'Create & Push',

  // SyncPull
  'syncPull.title': 'Pull from Tool',
  'syncPull.importCategory': 'Import Category',
  'syncPull.sourceTool': 'Source Tool',
  'syncPull.scanning': 'Scanning...',
  'syncPull.emptyWarning': 'No Skills found. Make sure the tool directory has subdirectories containing skill.md',
  'syncPull.selectSkills': 'Select Skills to import',
  'syncPull.importTo': 'Import to category "{cat}"',
  'syncPull.startPull': 'Start Pull ({count})',
  'syncPull.pulling': 'Pulling...',
  'syncPull.done': 'Pull complete ✓',
  'syncPull.noMatch': 'No Skills match the current filter',
  'syncPull.searchPlaceholder': 'Search scanned Skills...',

  // StarredRepos
  'starred.title': 'Starred Repos',
  'starred.addRepo': 'Add Repository',
  'starred.updateAll': 'Update All',
  'starred.batchImport': 'Batch Import',
  'starred.folder': 'Folder',
  'starred.flat': 'Flat',
  'starred.addRepoTitle': 'Add Remote Repository',
  'starred.addHint': 'First add will git clone the repo, may take a moment',
  'starred.cloning': 'Cloning...',
  'starred.addBtn': 'Add',
  'starred.emptyTitle': 'No starred repositories yet',
  'starred.emptyHint': 'Click "Add Repository" to get started',
  'starred.noSkills': 'No Skills found',
  'starred.selectCategory': 'Select Import Category',
  'starred.importing': 'Importing...',
  'starred.importN': 'Import {count}',
  'starred.pushToTools': 'Push to Tools',
  'starred.pushingToTools': 'Pushing...',
  'starred.pushToNTools': 'Push to {count} tools',
  'starred.noTools': 'No tools available. Enable tools in Settings.',
  'starred.mkdirTitle': 'Directory Not Found',
  'starred.mkdirDesc': 'The following push directories don\'t exist. Create them and continue?',
  'starred.createAndPush': 'Create & Push',
  'starred.authTitle': 'Authentication Required',
  'starred.authDesc': 'Repository requires username and password (or Access Token)',
  'starred.username': 'Username',
  'starred.password': 'Password / Access Token',
  'starred.connecting': 'Connecting...',
  'starred.sshTitle': 'SSH Auth Failed',
  'starred.sshDesc': 'Cannot access repository via SSH. Check:',
  'starred.conflictsTitle': 'Conflicts Found',
  'starred.conflictsDesc': 'The following Skills already exist in the target directory:',
  'starred.overwriteAll': 'Overwrite All',
  'starred.skipConflicts': 'Skip Conflicts',
  'starred.syncAt': 'Synced at',
  'starred.notSynced': 'Not synced',
  'starred.openInBrowser': 'Open in browser',
  'starred.updateBtn': 'Update',
  'starred.removeStarred': 'Remove starred',
  'starred.searchCurrentRepo': 'Search Skills in this repo...',
  'starred.searchAllRepos': 'Search starred repo Skills...',
  'starred.importToMySkills': 'Import to My Skills',
  'starred.pushToToolsCount': 'Push to Tools',
  'starred.successMsg': 'Successfully pushed {count} Skills to {toolCount} tools',

  // ToolSkills
  'toolSkills.title': 'My Tools',
  'toolSkills.toolList': 'Tools',
  'toolSkills.noTools': 'No tools enabled. Enable in Settings.',
  'toolSkills.selectToolFirst': 'Select a tool on the left',
  'toolSkills.batchDelete': 'Batch Delete',
  'toolSkills.pushPath': 'Push Path',
  'toolSkills.noPushDir': 'Not configured',
  'toolSkills.noPushDirDesc': 'No push path configured for this tool',
  'toolSkills.noPushSkills': 'No Skills in push path',
  'toolSkills.noMatch': 'No matching Skills',
  'toolSkills.scanPath': 'Scan Paths',
  'toolSkills.nDirs': '{count} directories',
  'toolSkills.noScanSkills': 'No standalone Skills in scan paths',
  'toolSkills.delete': 'Delete',
  'toolSkills.readOnly': 'Read-only',
  'toolSkills.copySkill': 'Copy skill.md',
  'toolSkills.openDir': 'Open Directory',
  'toolSkills.searchPlaceholder': 'Search Skills in tool...',

  // Backup
  'backup.title': 'Cloud Backup',
  'backup.notEnabled': 'Cloud backup not enabled. Go to Settings → Cloud Storage to configure.',
  'backup.backupNow': 'Backup Now',
  'backup.backingUp': 'Backing up {file}',
  'backup.pullRemote': 'Pull Remote',
  'backup.restore': 'Restore from Cloud',
  'backup.pulling': 'Pulling...',
  'backup.refresh': 'Refresh',
  'backup.gitDone': 'Git Sync Complete',
  'backup.done': 'Backup Complete',
  'backup.gitFailed': 'Git sync failed. Check repository config.',
  'backup.failed': 'Backup failed. Check cloud storage config.',
  'backup.gitFiles': 'Git Tracked Files',
  'backup.cloudFiles': 'Cloud Files',

  // Settings
  'settings.title': 'Settings',
  'settings.checkUpdate': 'Check for Updates',
  'settings.checkingUpdate': 'Checking...',
  'settings.tabTools': 'Tool Paths',
  'settings.tabCloud': 'Cloud Storage',
  'settings.tabGeneral': 'General',
  'settings.tabNetwork': 'Network',
  'settings.toolEnabled': 'Enable',
  'settings.pushPath': 'Push Path (single)',
  'settings.scanPaths': 'Scan Paths (multiple)',
  'settings.selectDir': 'Select Directory',
  'settings.deleteScanPath': 'Remove scan path',
  'settings.addPath': 'Add',
  'settings.deleteTool': 'Delete',
  'settings.addCustomTool': 'Add Custom Tool',
  'settings.toolName': 'Tool name',
  'settings.cloudProvider': 'Cloud Provider',
  'settings.bucket': 'Bucket',
  'settings.syncInterval': 'Auto-sync interval (minutes, 0 = on change only)',
  'settings.enableAutoBackup': 'Enable auto cloud backup',
  'settings.theme': 'Appearance Theme',
  'settings.themeHint': 'Legacy "Light" is auto-migrated to Young; the sidebar shortcut cycles Dark → Young → Light.',
  'settings.themeDark': 'Graphite black with hazy blue accents — a quiet, enduring dark workspace.',
  'settings.themeYoung': 'Light blue tones softened into a paper-like feel with morning light layers.',
  'settings.themeLight': 'Low-saturation gray-white theme inspired by sidebar gray-blue and Apple blue.',
  'settings.logLevel': 'Log Level',
  'settings.logLevelHint': 'Debug is most verbose, Info logs key events, Error logs failures only.',
  'settings.logDir': 'Log Directory',
  'settings.openLogDir': 'Open Log Directory',
  'settings.logDirHint': 'Max 2 log files, 1MB each, auto-rotated when size limit is reached.',
  'settings.skillsDir': 'Local Skills Storage Directory',
  'settings.defaultCategory': 'Default Category for Pull',
  'settings.defaultCategoryHint': 'Fixed system category for uncategorized imports. Cannot be renamed or deleted.',
  'settings.proxy': 'Proxy Settings',
  'settings.proxyHint': 'Proxy is used for remote repo operations (scan, install, update check)',
  'settings.proxyNone': 'No Proxy',
  'settings.proxyNoneDesc': 'Direct connection, no proxy',
  'settings.proxySystem': 'System Proxy',
  'settings.proxySystemDesc': 'Reads HTTP_PROXY / HTTPS_PROXY environment variables',
  'settings.proxyManual': 'Manual',
  'settings.proxyManualDesc': 'Custom proxy address',
  'settings.proxyUrl': 'Proxy URL',
  'settings.proxyUrlHint': 'Supports http://, https://, socks5:// formats',
  'settings.saveSettings': 'Save Settings',
  'settings.language': 'Language',
  'settings.updateFound': 'Found new version {version}',
  'settings.updateLatest': 'Already up to date ({version})',
  'settings.updateFailed': 'Check failed: {msg}',

  // CategoryPanel
  'category.all': 'All',
  'category.newCategory': 'New Category',
  'category.rename': 'Rename',
  'category.delete': 'Delete',
  'category.cannotDelete': 'Cannot Delete Category',
  'category.hasSkills': 'Category "{cat}" has {count} Skill(s). Clear it first.',
  'category.hasSkillsGeneric': 'Category "{cat}" still has Skills. Clear it first.',
  'category.deleteFailed': 'Delete Failed',

  // SkillCard
  'skillCard.update': 'Update',
  'skillCard.copy': 'Copy',
  'skillCard.copied': 'Copied',
  'skillCard.delete': 'Delete',
  'skillCard.moveTo': 'Move to {cat}',
  'skillCard.openDir': 'Open Directory',

  // ConflictDialog
  'conflictDialog.title': 'Conflict Detected',
  'conflictDialog.desc': '{name} already exists. How to proceed?',
  'conflictDialog.skip': 'Skip',
  'conflictDialog.overwrite': 'Overwrite',

  // GitHubInstallDialog
  'github.title': 'Install from Remote Repository',
  'github.scanning': 'Cloning/Updating...',
  'github.scan': 'Scan',
  'github.hint': 'First scan will git clone; subsequent scans auto-update via git pull',
  'github.noSkills': 'No Skills found. Make sure the repo has subdirectories containing skill.md',
  'github.installTo': 'Install to Category',
  'github.installing': 'Installing...',
  'github.installN': 'Install {count} Skills',
  'github.installed': 'Installed',

  // SkillListControls
  'skillList.sortAscTitle': 'Sort A to Z',
  'skillList.sortDescTitle': 'Sort Z to A',
  'skillList.searchDefault': 'Search Skills...',
} as const
```

### Step 3: Create `src/i18n/index.ts`

```ts
import { zh } from './zh'
import { en } from './en'

export type Lang = 'zh' | 'en'
export type Translations = typeof zh

export const locales: Record<Lang, Translations> = { zh, en }
```

### Step 4: Create `src/contexts/LanguageContext.tsx`

```tsx
import { createContext, useContext, useState, type ReactNode } from 'react'
import { type Lang, type Translations, locales } from '../i18n'

const STORAGE_KEY = 'skillflow-lang'

function getInitialLang(): Lang {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'zh' || stored === 'en') return stored
  } catch {}
  return 'zh'
}

interface LanguageContextValue {
  lang: Lang
  setLang: (lang: Lang) => void
  t: (key: keyof Translations, vars?: Record<string, string | number>) => string
}

const LanguageContext = createContext<LanguageContextValue | null>(null)

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(getInitialLang)

  const setLang = (newLang: Lang) => {
    setLangState(newLang)
    try { localStorage.setItem(STORAGE_KEY, newLang) } catch {}
  }

  const t = (key: keyof Translations, vars?: Record<string, string | number>): string => {
    let str = (locales[lang][key] ?? locales['zh'][key] ?? key) as string
    if (vars) {
      Object.entries(vars).forEach(([k, v]) => {
        str = str.replace(new RegExp(`\\{${k}\\}`, 'g'), String(v))
      })
    }
    return str
  }

  return (
    <LanguageContext.Provider value={{ lang, setLang, t }}>
      {children}
    </LanguageContext.Provider>
  )
}

export function useLanguage() {
  const ctx = useContext(LanguageContext)
  if (!ctx) throw new Error('useLanguage must be used inside LanguageProvider')
  return ctx
}
```

### Step 5: Commit

```bash
cd /path/to/repo
git add cmd/skillflow/frontend/src/i18n/ cmd/skillflow/frontend/src/contexts/LanguageContext.tsx
git commit -m "feat(i18n): add LanguageContext and zh/en translation maps"
```

---

## Task 2: Update App.tsx — provider, sidebar language button, all dialog strings

**Files:**
- Modify: `cmd/skillflow/frontend/src/App.tsx`

### Step 1: Import LanguageProvider and useLanguage, add Languages icon

In the imports block at the top of `App.tsx`:

```tsx
// Add to existing lucide-react import:
import { ..., Languages } from 'lucide-react'

// Add new import after ThemeProvider import:
import { LanguageProvider, useLanguage } from './contexts/LanguageContext'
```

### Step 2: Wrap app in LanguageProvider

Change the default export:

```tsx
export default function App() {
  return (
    <LanguageProvider>
      <ThemeProvider>
        <BrowserRouter>
          <AppContent />
        </BrowserRouter>
      </ThemeProvider>
    </LanguageProvider>
  )
}
```

### Step 3: Update AppContent to consume `t`

At the top of `AppContent()`, add:

```tsx
const { t, lang, setLang } = useLanguage()
```

Replace the `handleResolve` error fallback:
```tsx
// Before:
setResolveError(String(e?.message ?? e ?? '操作失败，请重试'))
// After:
setResolveError(String(e?.message ?? e ?? t('common.confirm')))
```

### Step 4: Replace all Chinese strings in Git Conflict dialog

```tsx
{/* Git conflict dialog */}
<AnimatedDialog open={conflictOpen} width="w-[420px]" zIndex={50}>
  <div className="flex items-center gap-2 mb-3">
    <AlertTriangle size={18} style={{ color: 'var(--color-warning)' }} />
    <span className="font-semibold text-base">{t('conflict.title')}</span>
  </div>
  <p className="text-sm mb-2" style={{ color: 'var(--text-secondary)' }}>
    {t('conflict.desc')}
  </p>
  {conflictInfo.files.length > 0 && (
    <div className="mb-3">
      <p className="text-xs mb-1.5" style={{ color: 'var(--text-muted)' }}>
        {t('conflict.filesLabel', { count: conflictInfo.files.length })}
      </p>
      <div className="max-h-28 overflow-y-auto rounded-lg px-2 py-1.5"
        style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-base)' }}>
        {conflictInfo.files.slice(0, 30).map((f, i) => (
          <div key={`${f}-${i}`} className="font-mono text-[11px] truncate"
            style={{ color: 'var(--text-secondary)' }}>{f}</div>
        ))}
        {conflictInfo.files.length > 30 && (
          <div className="text-[11px]" style={{ color: 'var(--text-muted)' }}>
            {t('conflict.moreFiles', { count: conflictInfo.files.length - 30 })}
          </div>
        )}
      </div>
    </div>
  )}
  {conflictInfo.message && (
    <div className="mb-3 rounded-lg px-2 py-1.5"
      style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-base)' }}>
      <p className="text-[11px] mb-1" style={{ color: 'var(--text-muted)' }}>{t('conflict.output')}</p>
      <pre className="text-[11px] whitespace-pre-wrap break-all max-h-20 overflow-y-auto"
        style={{ color: 'var(--text-secondary)' }}>{conflictInfo.message}</pre>
    </div>
  )}
  <ul className="text-xs list-disc list-inside mb-6 space-y-1" style={{ color: 'var(--text-muted)' }}>
    <li><span className="font-medium" style={{ color: 'var(--text-primary)' }}>{t('conflict.keepLocal')}</span> — {t('conflict.keepLocalDesc')}</li>
    <li><span className="font-medium" style={{ color: 'var(--text-primary)' }}>{t('conflict.keepRemote')}</span> — {t('conflict.keepRemoteDesc')}</li>
  </ul>
  {resolveError && (
    <p className="mb-3 text-xs rounded-lg px-3 py-2 break-all"
      style={{ color: 'var(--color-error)', background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)' }}>
      {resolveError}
    </p>
  )}
  <div className="flex gap-3 justify-end">
    <button onClick={() => handleResolve(false)} disabled={resolving}
      className="btn-secondary flex items-center gap-1.5 px-4 py-2 text-sm rounded-lg">
      {resolving ? <RefreshCw size={13} className="animate-spin" /> : <Download size={13} />}
      {t('conflict.keepRemote')}
    </button>
    <button onClick={() => handleResolve(true)} disabled={resolving}
      className="btn-primary flex items-center gap-1.5 px-4 py-2 text-sm rounded-lg">
      {resolving ? <RefreshCw size={13} className="animate-spin" /> : <GitMerge size={13} />}
      {t('conflict.keepLocal')}
    </button>
  </div>
</AnimatedDialog>
```

### Step 5: Replace sidebar nav labels and add language button

```tsx
<aside ...>
  ...
  <div className="flex items-center justify-between mb-6 px-2">
    <h1 className="text-[17px] font-semibold tracking-[0.08em]"
      style={{ color: 'var(--brand-color)', textShadow: 'var(--brand-shadow)' }}>
      SkillFlow
    </h1>
    <div className="flex items-center gap-1">
      <button
        onClick={() => setLang(lang === 'zh' ? 'en' : 'zh')}
        className="p-1.5 rounded-lg transition-colors"
        style={{ color: 'var(--text-muted)' }}
        title={t('nav.switchLangTitle')}
      >
        <Languages size={14} />
      </button>
      <button
        onClick={cycleTheme}
        className="p-1.5 rounded-lg transition-colors"
        style={{ color: 'var(--text-muted)' }}
        title={t('nav.switchThemeTitle', { theme: THEME_LABELS[nextTheme] })}
      >
        <Palette size={14} />
      </button>
    </div>
  </div>
  <NavItem to="/" icon={<Package size={16} />} label={t('nav.mySkills')} />
  <NavItem to="/tools" icon={<Wrench size={16} />} label={t('nav.myTools')} end={false} />
  <p className="text-xs px-2 mt-3 mb-1" style={{ color: 'var(--text-muted)' }}>{t('nav.syncSection')}</p>
  <NavItem to="/sync/push" icon={<ArrowUpFromLine size={16} />} label={t('nav.pushToTool')} />
  <NavItem to="/sync/pull" icon={<ArrowDownToLine size={16} />} label={t('nav.pullFromTool')} />
  <NavItem to="/starred" icon={<Star size={16} />} label={t('nav.starred')} end={false} />
  <div className="flex-1" />
  <div className="flex flex-col gap-1">
    <NavItem to="/backup" icon={<Cloud size={16} />} label={t('nav.backup')} />
    <NavItem to="/settings" icon={<Settings size={16} />} label={t('nav.settings')} />
    <button
      onClick={() => OpenURL(feedbackIssueURL)}
      className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors"
      style={{ color: 'var(--text-muted)' }}
      onMouseEnter={...}
      onMouseLeave={...}
    >
      <MessageSquareWarning size={16} />
      {t('nav.feedback')}
    </button>
  </div>
</aside>
```

### Step 6: Update UpdateDialogContent to accept and use `t`

Add `t` to the props interface and pass it from caller:

```tsx
interface UpdateDialogContentProps {
  // ...existing...
  t: (key: keyof Translations, vars?: Record<string, string | number>) => string
}
```

In `AppContent`, pass `t` when rendering `<UpdateDialogContent ... t={t} />`.

Replace all Chinese strings in `UpdateDialogContent` using `t()`:
- `'更新已就绪'` → `t('update.ready')`
- `'下载失败'` → `t('update.failed')`
- `'发现新版本'` → `t('update.newVersion')`
- `'最新版本：'` → `t('update.latestLabel')`
- `'当前版本：'` → `t('update.currentLabel')`
- `'更新说明'` → `t('update.notes')`
- `'正在下载 {version}，请稍候...'` → `t('update.downloading', { version: info?.latestVersion ?? '' })`
- `'新版本已下载完成...'` → `t('update.restartDesc')`
- `'自动下载失败...'` → `t('update.downloadFailDesc')`
- `'下载并自动重启更新'` → `t('update.download')`
- `'前往 Release 页面手动下载'` → `t('update.openRelease')`
- `'跳过此版本...'` → `t('update.skip')`
- `'下载完成后将自动提示重启'` → `t('update.downloadHint')`
- `'稍后重启'` → `t('update.restartLater')`
- `'立即重启'` → `t('update.restartNow')`
- `'关闭'` → `t('common.close')`
- `'前往下载页'` → `t('update.goDownload')`

Also add the `Translations` type import: `import type { Translations } from './i18n'`

### Step 7: Commit

```bash
git add cmd/skillflow/frontend/src/App.tsx
git commit -m "feat(i18n): update App.tsx — sidebar language toggle, dialog strings"
```

---

## Task 3: Update Settings.tsx

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/Settings.tsx`

### Step 1: Import useLanguage and update themeOptions descriptions

```tsx
import { useLanguage } from '../contexts/LanguageContext'
```

At the top of `SettingsPage()`, add:
```tsx
const { t, lang, setLang } = useLanguage()
```

Move `themeOptions` inside `SettingsPage` (or make it a function of `t`) so descriptions use `t()`:

```tsx
const themeOptions: ThemeOption[] = [
  { id: 'dark', label: 'Dark', tone: 'Ink Slate', description: t('settings.themeDark'), icon: <Moon size={15} />, preview: { ... } },
  { id: 'young', label: 'Young', tone: 'Breeze Paper', description: t('settings.themeYoung'), icon: <Sparkles size={15} />, preview: { ... } },
  { id: 'light', label: 'Light', tone: 'Messor Calm', description: t('settings.themeLight'), icon: <Sun size={15} />, preview: { ... } },
]
```

(Keep all preview palette objects exactly as-is, only change `description`.)

### Step 2: Update checkUpdate result strings

```tsx
// Before:
setUpdateResult(`发现新版本 ${info.latestVersion}`)
// After:
setUpdateResult(t('settings.updateFound', { version: info.latestVersion }))

// Before:
setUpdateResult(`已是最新版本 (${info.currentVersion})`)
// After:
setUpdateResult(t('settings.updateLatest', { version: info.currentVersion }))

// Before:
setUpdateResult(`检测失败: ${e?.message ?? String(e)}`)
// After:
setUpdateResult(t('settings.updateFailed', { msg: e?.message ?? String(e) }))
```

### Step 3: Replace loading state string

```tsx
// Before:
if (!cfg) return <div className="p-8" style={{ color: 'var(--text-muted)' }}>加载中...</div>
// After:
if (!cfg) return <div className="p-8" style={{ color: 'var(--text-muted)' }}>{t('common.loading')}</div>
```

### Step 4: Replace all Chinese strings in JSX

**Header:**
```tsx
<h2 ...><Settings size={18} /> {t('settings.title')}</h2>
<button onClick={checkUpdate} disabled={checkingUpdate} ...>
  {checkingUpdate ? t('settings.checkingUpdate') : t('settings.checkUpdate')}
</button>
```

**Tabs:**
```tsx
{([['tools', t('settings.tabTools')], ['cloud', t('settings.tabCloud')], ['general', t('settings.tabGeneral')], ['network', t('settings.tabNetwork')]] as [Tab, string][]).map(...)}
```

**Tools tab:**
- `'启用'` → `t('settings.toolEnabled')`
- `'推送路径（仅 1 个）'` → `t('settings.pushPath')`
- `'扫描路径（可多个）'` → `t('settings.scanPaths')`
- `title="选择目录"` → `title={t('settings.selectDir')}`
- `title="删除扫描路径"` → `title={t('settings.deleteScanPath')}`
- `'添加'` (scan path button) → `t('settings.addPath')`
- `'删除'` (custom tool) → `t('settings.deleteTool')`
- `'添加自定义工具'` → `t('settings.addCustomTool')`
- `placeholder="工具名称"` → `placeholder={t('settings.toolName')}`
- `'添加'` (add tool button) → `t('settings.addPath')`

**Cloud tab:**
- `'云厂商'` → `t('settings.cloudProvider')`
- `'存储桶'` → `t('settings.bucket')`
- `'定时自动同步...'` → `t('settings.syncInterval')`
- `'启用自动云备份'` → `t('settings.enableAutoBackup')`

**General tab — add Language section ABOVE theme section:**
```tsx
{/* Language */}
<div>
  <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.language')}</p>
  <div className="flex gap-2">
    {(['zh', 'en'] as const).map(l => (
      <button
        key={l}
        onClick={() => setLang(l)}
        className="px-4 py-1.5 rounded-lg text-sm transition-all duration-200"
        style={lang === l ? {
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
        {l === 'zh' ? '中文' : 'English'}
      </button>
    ))}
  </div>
</div>
```

**General tab — remaining strings:**
- `'外观主题'` → `t('settings.theme')`
- theme hint text → `t('settings.themeHint')`
- `'日志打印级别'` → `t('settings.logLevel')`
- log level hint → `t('settings.logLevelHint')`
- `'日志目录'` → `t('settings.logDir')`
- `'打开日志目录'` → `t('settings.openLogDir')`
- log dir hint → `t('settings.logDirHint')`
- `'本地 Skills 存储目录'` → `t('settings.skillsDir')`
- `'从工具拉取时的默认分类'` → `t('settings.defaultCategory')`
- default category hint → `t('settings.defaultCategoryHint')`

**Network tab:**
- `'代理设置'` → `t('settings.proxy')`
- proxy hint → `t('settings.proxyHint')`
- proxy mode tuples: replace each label/desc with `t('settings.proxyNone')`, `t('settings.proxyNoneDesc')`, etc.
- `'代理地址'` → `t('settings.proxyUrl')`
- proxy url hint → `t('settings.proxyUrlHint')`

**Save button:**
```tsx
<button onClick={save} disabled={saving} className="btn-primary mt-8 px-6 py-2.5 rounded-lg text-sm">
  {saving ? t('common.saving') : t('settings.saveSettings')}
</button>
```

### Step 5: Commit

```bash
git add cmd/skillflow/frontend/src/pages/Settings.tsx
git commit -m "feat(i18n): update Settings.tsx — language selector and all tab strings"
```

---

## Task 4: Update Dashboard.tsx

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/Dashboard.tsx`

### Step 1: Import and consume `useLanguage`

```tsx
import { useLanguage } from '../contexts/LanguageContext'
// Inside Dashboard():
const { t } = useLanguage()
```

### Step 2: Replace all Chinese strings

- `placeholder="搜索 Skills..."` in `<SkillListControls>` → `placeholder={t('dashboard.searchPlaceholder')}`
- `resultLabel={`当前显示 ${filtered.length} 个 Skill`}` → `resultLabel={t('common.showingNSkills', { count: filtered.length })}`
- `{allSelected ? '取消全选' : '全选'}` → `{allSelected ? t('common.deselectAll') : t('common.selectAll')}`
- `<Trash2 size={14} /> 删除 {selectedIDs.size > 0 ? ...}` → `<Trash2 size={14} /> {t('common.delete')} {selectedIDs.size > 0 ? ...}`
- `'取消'` (cancel batch) → `t('common.cancel')`
- Button labels array: `'更新'` → `t('dashboard.update')`, `'批删'` → `t('dashboard.batchDelete')`, `'导入'` → `t('dashboard.import')`
- `'远程安装'` → `t('dashboard.remoteInstall')`
- `'松开以导入 Skill'` → `t('dashboard.dropToImport')`
- `'没有找到 Skills'` → `t('dashboard.empty')`
- `'从远程仓库安装或拖拽文件夹到此处'` → `t('dashboard.emptyHint')`

### Step 3: Commit

```bash
git add cmd/skillflow/frontend/src/pages/Dashboard.tsx
git commit -m "feat(i18n): update Dashboard.tsx"
```

---

## Task 5: Update SyncPush.tsx

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/SyncPush.tsx`

### Step 1: Import and consume `useLanguage`

```tsx
import { useLanguage } from '../contexts/LanguageContext'
// Inside SyncPush():
const { t } = useLanguage()
```

### Step 2: Update `scopeLabel` computation

```tsx
const scopeLabel = scope === 'manual'
  ? t('syncPush.scopeManual', { count: selectedSkills.size })
  : selectedCategory === null
    ? t('syncPush.scopeAll', { count: visibleSkills.length })
    : t('syncPush.scopeCategory', { cat: selectedCategory, count: visibleSkills.length })
```

### Step 3: Replace all Chinese strings

- Left panel header `'推送范围'` → `t('syncPush.pushRange')`
- `'全部'` (nav all) → `t('common.all')`
- Page title `'推送到工具'` → `t('syncPush.title')`
- `'目标工具'` → `t('syncPush.targetTool')`
- `<SkillListControls placeholder="搜索待推送的 Skills..." resultLabel={...}/>` → use `t('syncPush.searchPlaceholder')` and `t('common.showingNSkills', { count: visibleSkills.length })`
- `{selectedCategory === null ? '推送全部' : '推送当前分类'}` → `{selectedCategory === null ? t('syncPush.pushAll') : t('syncPush.pushCategory')}`
- `'手动选择 Skill'` → `t('syncPush.manualSelect')`
- `{allManualSelected ? '取消全选' : '全选当前列表'}` → `{allManualSelected ? t('common.deselectAll') : t('syncPush.selectAllList')}`
- `'当前可选 {n} 个 Skill'` → `t('syncPush.nSkillsVisible', { count: visibleSkills.length })`
- `'当前范围内没有 Skill'` → `t('syncPush.empty')`
- `'选择"全部"或切换到其他分类后再试'` → `t('syncPush.emptyHint')`
- Push button: `{pushing ? '推送中...' : `开始推送 (${pushCount})`}` → `{pushing ? t('syncPush.pushing') : t('syncPush.startPush', { count: pushCount })}`
- `'推送完成'` → `t('syncPush.done')`
- mkdir dialog: title `'目录不存在'` → `t('syncPush.mkdirTitle')`, desc → `t('syncPush.mkdirDesc')`, `'创建并推送'` → `t('syncPush.createAndPush')`, `'取消'` → `t('common.cancel')`

### Step 4: Commit

```bash
git add cmd/skillflow/frontend/src/pages/SyncPush.tsx
git commit -m "feat(i18n): update SyncPush.tsx"
```

---

## Task 6: Update SyncPull.tsx

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/SyncPull.tsx`

### Step 1: Import and consume `useLanguage`

```tsx
import { useLanguage } from '../contexts/LanguageContext'
const { t } = useLanguage()
```

### Step 2: Replace all Chinese strings

- Left panel header `'导入分类'` → `t('syncPull.importCategory')`
- Page title `'从工具拉取'` → `t('syncPull.title')`
- `'来源工具'` → `t('syncPull.sourceTool')`
- `'扫描中...'` → `t('syncPull.scanning')`
- Warning `'未发现任何 Skill...'` → `t('syncPull.emptyWarning')`
- `<SkillListControls placeholder="搜索扫描到的 Skills..." resultLabel={...}/>` → use translations
- `'选择要导入的 Skills'` → `t('syncPull.selectSkills')`
- `{allSelected ? '取消全选' : '全选'}` → `{allSelected ? t('common.deselectAll') : t('common.selectAll')}`
- `'导入到分类「{cat}」'` → `t('syncPull.importTo', { cat: targetCategory || defaultCategory })`
- Pull button: `{pulling ? '拉取中...' : `开始拉取 (${selected.size})`}` → `{pulling ? t('syncPull.pulling') : t('syncPull.startPull', { count: selected.size })}`
- `'拉取完成 ✓'` → `t('syncPull.done')`
- No match: `'当前筛选下没有匹配的 Skill'` → `t('syncPull.noMatch')`

### Step 3: Commit

```bash
git add cmd/skillflow/frontend/src/pages/SyncPull.tsx
git commit -m "feat(i18n): update SyncPull.tsx"
```

---

## Task 7: Update StarredRepos.tsx

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/StarredRepos.tsx`

### Step 1: Import and consume `useLanguage`

```tsx
import { useLanguage } from '../contexts/LanguageContext'
const { t } = useLanguage()
```

Also pass `t` down to `RepoGrid` and `SkillGrid` as a prop.

### Step 2: Update success message

```tsx
// Before:
setPushSuccessMsg(`已成功推送 ${count} 个 Skill 到 ${toolCount} 个工具`)
// After:
setPushSuccessMsg(t('starred.successMsg', { count, toolCount }))
```

### Step 3: Replace all Chinese strings in toolbar and dialogs

Key replacements (apply `t()` throughout):
- `'仓库收藏'` (title h2) → `t('starred.title')`
- `'文件夹'`, `'平铺'` view buttons → `t('starred.folder')`, `t('starred.flat')`
- `{selectedPaths.size === skills.length ? '取消全选' : '全选'}` → using `t('common.deselectAll')` / `t('common.selectAll')`
- `'推送到工具'` count button → include selected count
- `'导入到我的Skills'` → `t('starred.importToMySkills')` + count
- `'取消'` → `t('common.cancel')`
- `'批量导入'` → `t('starred.batchImport')`
- `'全部更新'` → `t('starred.updateAll')`
- `'添加仓库'` → `t('starred.addRepo')`
- SkillListControls: use `t('starred.searchCurrentRepo')` or `t('starred.searchAllRepos')` for placeholder
- All dialog titles/descriptions/buttons using `t()` as per key mapping

**RepoGrid** — pass `t` as prop and update:
- `'还没有收藏的仓库'` → `t('starred.emptyTitle')`
- `'点击「添加仓库」开始收藏'` → `t('starred.emptyHint')`
- `title="在浏览器中打开"` → `title={t('starred.openInBrowser')}`
- `title="更新"` → `title={t('starred.updateBtn')}`
- `title="删除收藏"` → `title={t('starred.removeStarred')}`
- `'同步于'` → `t('starred.syncAt')`
- `'未同步'` → `t('starred.notSynced')`

**SkillGrid** — pass `t` as prop and update:
- `'没有找到 Skills'` → `t('starred.noSkills')`

### Step 4: Commit

```bash
git add cmd/skillflow/frontend/src/pages/StarredRepos.tsx
git commit -m "feat(i18n): update StarredRepos.tsx"
```

---

## Task 8: Update ToolSkills.tsx

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`

### Step 1: Import and consume `useLanguage`

```tsx
import { useLanguage } from '../contexts/LanguageContext'
// In ToolSkills():
const { t } = useLanguage()
// Pass t to ToolSkillCard via prop
```

Add `t` prop to `ToolSkillCardProps` interface:
```tsx
interface ToolSkillCardProps {
  // ...existing...
  t: (key: any, vars?: Record<string, string | number>) => string
}
```

### Step 2: Replace all Chinese strings in ToolSkills

- Left panel header `'工具列表'` → `t('toolSkills.toolList')`
- `'没有启用的工具，请在设置中启用'` → `t('toolSkills.noTools')`
- `tool ? ... : <h2>...<Wrench/> 我的工具</h2>` → `t('toolSkills.title')`
- `{allSelected ? '取消全选' : '全选'}` → `t('common.deselectAll')` / `t('common.selectAll')`
- `'删除'` (batch delete button) → `t('common.delete')`
- `'取消'` → `t('common.cancel')`
- `'批量删除'` → `t('toolSkills.batchDelete')`
- `<SkillListControls placeholder="搜索当前工具中的 Skills..." resultLabel={...}/>` → use translations
- `'加载中...'` → `t('common.loading')`
- `'请先在左侧选择一个工具'` → `t('toolSkills.selectToolFirst')`
- `'推送路径'` section label → `t('toolSkills.pushPath')`
- `'未配置'` → `t('toolSkills.noPushDir')`
- `'该工具未配置推送路径'` → `t('toolSkills.noPushDirDesc')`
- `'推送路径下暂无 Skill'` → `t('toolSkills.noPushSkills')`
- `'当前筛选下没有匹配的 Skill'` (×2) → `t('toolSkills.noMatch')`
- `'扫描路径'` section label → `t('toolSkills.scanPath')`
- `'{n} 个目录'` → `t('toolSkills.nDirs', { count: tool.scanDirs.length })`
- `'扫描路径下暂无独立 Skill'` → `t('toolSkills.noScanSkills')`

### Step 3: Replace Chinese strings in ToolSkillCard

Pass `t` from ToolSkills to ToolSkillCard, then replace:
- `title="复制 skill.md"` → `title={t('toolSkills.copySkill')}`
- `title="打开目录"` → `title={t('toolSkills.openDir')}`
- `'删除'` (hover delete) → `t('toolSkills.delete')`
- `'只读'` (badge) → `t('toolSkills.readOnly')`

### Step 4: Commit

```bash
git add cmd/skillflow/frontend/src/pages/ToolSkills.tsx
git commit -m "feat(i18n): update ToolSkills.tsx"
```

---

## Task 9: Update Backup.tsx

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/Backup.tsx`

### Step 1: Import and consume `useLanguage`

```tsx
import { useLanguage } from '../contexts/LanguageContext'
const { t } = useLanguage()
```

### Step 2: Replace all Chinese strings

- `'云备份'` (title) → `t('backup.title')`
- `'云备份未启用。请前往设置 → 云存储完成配置。'` → `t('backup.notEnabled')`
- Backup button: `{pushing ? `备份中 ${currentFile}` : '立即备份'}` → `{pushing ? t('backup.backingUp', { file: currentFile }) : t('backup.backupNow')}`
- Restore button: `{pulling ? '拉取中...' : (isGit ? '拉取远端' : '从云端恢复')}` → `{pulling ? t('backup.pulling') : (isGit ? t('backup.pullRemote') : t('backup.restore'))}`
- `'刷新'` → `t('backup.refresh')`
- Success: `{isGit ? 'Git 同步完成' : '备份完成'}` → `{isGit ? t('backup.gitDone') : t('backup.done')}`
- Error: `{isGit ? 'Git 同步失败...' : '备份失败...'}` → `{isGit ? t('backup.gitFailed') : t('backup.failed')}`
- Files header: `{isGit ? 'Git 跟踪文件' : '云端文件'}（{files.length} 个）` → `{isGit ? t('backup.gitFiles') : t('backup.cloudFiles')}（{files.length}）`

### Step 3: Commit

```bash
git add cmd/skillflow/frontend/src/pages/Backup.tsx
git commit -m "feat(i18n): update Backup.tsx"
```

---

## Task 10: Update CategoryPanel.tsx

**Files:**
- Modify: `cmd/skillflow/frontend/src/components/CategoryPanel.tsx`

### Step 1: Import and consume `useLanguage`

```tsx
import { useLanguage } from '../contexts/LanguageContext'
// Inside CategoryPanel():
const { t } = useLanguage()
```

### Step 2: Replace all Chinese strings

- `'全部'` category item → `t('category.all')`
- `<Plus size={14} /> 新建分类` → `<Plus size={14} /> {t('category.newCategory')}`
- Context menu items:
  ```tsx
  items={[
    { label: t('category.rename'), onClick: () => { setRenaming(menu.cat); setNewName(menu.cat) } },
    { label: t('category.delete'), onClick: async () => { await handleDeleteCategory(menu.cat) }, danger: true },
  ]}
  ```
- Error message fallback: `String(e?.message ?? e ?? '删除分类失败')` → use a generic error or `t('category.deleteFailed')`
- `'无法删除分类'` dialog title → `t('category.cannotDelete')`
- Dialog body:
  ```tsx
  {(deleteBlocked?.count ?? 0) > 0
    ? t('category.hasSkills', { cat: deleteBlocked?.cat ?? '', count: deleteBlocked?.count ?? 0 })
    : t('category.hasSkillsGeneric', { cat: deleteBlocked?.cat ?? '' })}
  ```
- `'我知道了'` → `t('common.gotIt')`
- `'删除失败'` dialog title → `t('category.deleteFailed')`
- `'关闭'` → `t('common.close')`

### Step 3: Commit

```bash
git add cmd/skillflow/frontend/src/components/CategoryPanel.tsx
git commit -m "feat(i18n): update CategoryPanel.tsx"
```

---

## Task 11: Update SkillCard.tsx and ConflictDialog.tsx

**Files:**
- Modify: `cmd/skillflow/frontend/src/components/SkillCard.tsx`
- Modify: `cmd/skillflow/frontend/src/components/ConflictDialog.tsx`

### Step 1: Update SkillCard.tsx

```tsx
import { useLanguage } from '../contexts/LanguageContext'
// Inside SkillCard():
const { t } = useLanguage()
```

Replace:
```tsx
const menuItems = [
  ...(skill.hasUpdate ? [{ label: t('skillCard.update'), onClick: () => onUpdate?.() }] : []),
  ...categories.filter(c => c !== skill.category).map(c => ({
    label: t('skillCard.moveTo', { cat: c }),
    onClick: () => onMoveCategory(c),
  })),
  { label: t('skillCard.delete'), onClick: onDelete, danger: true },
]
```

Other replacements:
- `title="打开目录"` → `title={t('skillCard.openDir')}`
- `<RefreshCw size={12} /> 更新` → `<RefreshCw size={12} /> {t('skillCard.update')}`
- `<><Check .../> 已复制</>` → `<><Check .../> {t('skillCard.copied')}</>`
- `<><Copy .../> 复制</>` → `<><Copy .../> {t('skillCard.copy')}</>`
- `'删除'` (button) → `t('skillCard.delete')`

### Step 2: Update ConflictDialog.tsx

```tsx
import { useLanguage } from '../contexts/LanguageContext'
// Inside ConflictDialog():
const { t } = useLanguage()
```

Replace:
```tsx
<h3 ...>{t('conflictDialog.title')}</h3>
<p ...>
  <span ...>{current}</span> {/* keep dynamic part */}
  {t('conflictDialog.desc', { name: '' }).replace(current + ' ', '')}
  {/* Simpler: render as: */}
</p>
```

Actually for the conflict description, use:
```tsx
<p className="text-sm mb-6" style={{ color: 'var(--text-muted)' }}>
  <span className="font-medium" style={{ color: 'var(--text-primary)' }}>{current}</span>
  {' '}{t('conflictDialog.desc', { name: '' }).split('{name}').pop()?.trim() ?? '已存在，如何处理？'}
</p>
```

Simpler approach — have the description key NOT include `{name}` and split into two parts:
In `zh.ts` / `en.ts`, define:
- `'conflictDialog.existsSuffix': '已存在，如何处理？'` / `' already exists. How to proceed?'`

Then in JSX:
```tsx
<p className="text-sm mb-6" style={{ color: 'var(--text-muted)' }}>
  <span className="font-medium" style={{ color: 'var(--text-primary)' }}>{current}</span>
  {' '}{t('conflictDialog.existsSuffix')}
</p>
```

Add `'conflictDialog.existsSuffix'` to both `zh.ts` and `en.ts`:
- zh: `'已存在，如何处理？'`
- en: `'already exists. How to proceed?'`

Replace buttons:
- `'跳过'` → `t('conflictDialog.skip')`
- `'覆盖'` → `t('conflictDialog.overwrite')`

### Step 3: Commit

```bash
git add cmd/skillflow/frontend/src/components/SkillCard.tsx cmd/skillflow/frontend/src/components/ConflictDialog.tsx
git commit -m "feat(i18n): update SkillCard and ConflictDialog"
```

---

## Task 12: Update GitHubInstallDialog.tsx and SkillListControls.tsx

**Files:**
- Modify: `cmd/skillflow/frontend/src/components/GitHubInstallDialog.tsx`
- Modify: `cmd/skillflow/frontend/src/components/SkillListControls.tsx`

### Step 1: Update GitHubInstallDialog.tsx

```tsx
import { useLanguage } from '../contexts/LanguageContext'
// Inside GitHubInstallDialog():
const { t } = useLanguage()
```

Replace:
- `'从远程仓库安装'` title → `t('github.title')`
- Scan button: `{scanning ? <span>...'克隆/更新中...'</span> : '扫描'}` → `{scanning ? <span>...{t('github.scanning')}</span> : t('github.scan')}`
- `'首次扫描会 git clone 仓库，后续自动 git pull 更新'` → `t('github.hint')`
- `'未发现任何 Skill...'` (setScanError call) → `t('github.noSkills')`
- `'扫描失败，请检查网络或仓库地址'` fallback → keep as-is (it's a dynamic error message)
- `'安装失败'` fallback → keep as-is
- `'已安装'` badge → `t('github.installed')`
- `'安装到分类'` label → `t('github.installTo')`
- Install button: `{installing ? '安装中...' : `安装 ${selected.size} 个 Skill`}` → `{installing ? t('github.installing') : t('github.installN', { count: selected.size })}`

### Step 2: Update SkillListControls.tsx

```tsx
import { useLanguage } from '../contexts/LanguageContext'
// Inside SkillListControls():
const { t } = useLanguage()
```

The `placeholder` and `resultLabel` props are already passed from callers — the callers will pass translated strings (done in prior tasks). Only the sort button titles need updating here:

```tsx
title={option.value === 'asc' ? t('skillList.sortAscTitle') : t('skillList.sortDescTitle')}
```

Also update the default value for `placeholder` prop:
```tsx
placeholder = t('skillList.searchDefault'),
```

Wait — default prop values can't use hooks. Instead, make the default `undefined` and handle it in the component:
```tsx
placeholder?: string
// ...
<input placeholder={placeholder ?? t('skillList.searchDefault')} ... />
```

### Step 3: Commit

```bash
git add cmd/skillflow/frontend/src/components/GitHubInstallDialog.tsx cmd/skillflow/frontend/src/components/SkillListControls.tsx
git commit -m "feat(i18n): update GitHubInstallDialog and SkillListControls"
```

---

## Task 13: Update documentation

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `README.md`
- Modify: `README_zh.md`

### Step 1: Update `docs/features.md`

Add a new section under Settings:

```markdown
### Language Settings

A language toggle is available in two places:
- **Sidebar header**: Globe icon (🌐) next to the theme button — click to cycle between Chinese and English instantly.
- **Settings → General → Language**: Select `中文` or `English` using styled buttons.

The selected language is persisted to `localStorage` and survives app restarts. Default is Chinese.
```

Update "Last updated" date at bottom to 2026-03-07.

### Step 2: Update `docs/features_zh.md`

Add the equivalent Chinese section:

```markdown
### 语言设置

语言切换入口有两处：
- **侧边栏右上角**：地球仪图标（Language 按钮），紧邻主题切换按钮，点击即可在中英文之间切换。
- **设置 → 通用 → 语言**：点击 `中文` 或 `English` 按钮切换。

语言偏好保存于 `localStorage`，应用重启后不丢失，默认为中文。
```

Update "最后更新" date.

### Step 3: Update README.md and README_zh.md

In the features table, update the Settings row to mention language switching:

README.md — add to the relevant features row or add a new row:
```markdown
| 🌐 Bilingual UI | Switch between Chinese and English — sidebar button or Settings → General |
```

README_zh.md:
```markdown
| 🌐 双语界面 | 支持中英文切换，侧边栏一键切换或在「设置 → 通用」选择语言 |
```

### Step 4: Commit

```bash
git add docs/features.md docs/features_zh.md README.md README_zh.md
git commit -m "docs: update feature docs and README for i18n language switching"
```

---

## Summary

All 13 tasks together implement the full i18n feature:

1. **Task 1** — i18n infrastructure (zero dependencies, ~240 translation keys)
2. **Task 2** — App.tsx: sidebar language button + all dialog strings
3. **Task 3** — Settings.tsx: language selector row + all settings strings
4. **Tasks 4–9** — Page components: Dashboard, SyncPush, SyncPull, StarredRepos, ToolSkills, Backup
5. **Tasks 10–12** — Shared components: CategoryPanel, SkillCard, ConflictDialog, GitHubInstallDialog, SkillListControls
6. **Task 13** — Documentation sync

Each task is independently committable. Work through them in order since Tasks 4–12 all depend on Task 1 being complete.
