import { useEffect, useMemo, useState } from 'react'
import { Check, Copy, Download, FileText, Plus, Trash2, Upload } from 'lucide-react'
import {
  CreatePrompt,
  DeletePrompt,
  ExportPromptsByNames,
  ImportPrompts,
  ListPromptCategories,
  ListPrompts,
  MovePromptCategory,
  UpdatePrompt,
} from '../../wailsjs/go/main/App'
import PromptEditorDialog, { type PromptDraft } from '../components/PromptEditorDialog'
import PromptCategoryPanel from '../components/PromptCategoryPanel'
import AnimatedDialog from '../components/ui/AnimatedDialog'
import SkillListControls from '../components/SkillListControls'
import { useLanguage } from '../contexts/LanguageContext'
import { copyTextToClipboard } from '../lib/clipboard'
import { buildPromptExportActions, canExportPromptSelection, listPromptExportCandidates, resolvePromptExportNames, type PromptExportMode } from '../lib/promptExport'
import { buildPromptLinksMarkdown, normalizePromptImageURLs, type PromptWebLink } from '../lib/promptRichContent'
import { SkillSortOrder } from '../lib/skillList'
import { matchesKeywordExpression } from '../lib/search'

type PromptItem = {
  name: string
  description: string
  category: string
  path: string
  filePath: string
  content: string
  imageURLs?: string[]
  webLinks?: PromptWebLink[]
  createdAt?: string
  updatedAt?: string
}

type StatusState = {
  type: 'success' | 'error'
  message: string
} | null

const defaultCategoryName = 'Default'
const collator = new Intl.Collator(undefined, { numeric: true, sensitivity: 'base' })

function createEmptyDraft(category = defaultCategoryName): PromptDraft {
  return {
    name: '',
    description: '',
    category,
    content: '',
    imageURLs: [],
    webLinks: [],
  }
}

function getPromptPreview(content: string) {
  const normalized = content.replace(/\r\n/g, '\n').trim()
  if (!normalized) return { preview: '', truncated: false }
  const compact = normalized.replace(/\s+/g, ' ')
  if (compact.length <= 88) return { preview: compact, truncated: false }
  return {
    preview: `${compact.slice(0, 88).trimEnd()}…`,
    truncated: true,
  }
}

export default function Prompts() {
  const { t } = useLanguage()
  const [prompts, setPrompts] = useState<PromptItem[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [selectedCat, setSelectedCat] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [sortOrder, setSortOrder] = useState<SkillSortOrder>('asc')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [importing, setImporting] = useState(false)
  const [exporting, setExporting] = useState(false)
  const [exportBarOpen, setExportBarOpen] = useState(false)
  const [exportMode, setExportMode] = useState<PromptExportMode>('all')
  const [selectedExportNames, setSelectedExportNames] = useState<Set<string>>(new Set())
  const [editorOpen, setEditorOpen] = useState(false)
  const [editorMode, setEditorMode] = useState<'create' | 'edit'>('create')
  const [draft, setDraft] = useState<PromptDraft>(createEmptyDraft())
  const [deleteTarget, setDeleteTarget] = useState<PromptItem | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [saveError, setSaveError] = useState('')
  const [status, setStatus] = useState<StatusState>(null)
  const [copiedName, setCopiedName] = useState<string | null>(null)
  const [draggingPromptName, setDraggingPromptName] = useState<string | null>(null)
  const [categoryDragActive, setCategoryDragActive] = useState(false)

  const load = async () => {
    setLoading(true)
    try {
      const [promptItems, promptCategories] = await Promise.all([ListPrompts(), ListPromptCategories()])
      const nextPrompts = (promptItems ?? []) as PromptItem[]
      const nextCategories = Array.from(new Set([defaultCategoryName, ...((promptCategories ?? []) as string[])]))
      setPrompts(nextPrompts)
      setCategories(nextCategories)
      setSelectedCat(current => current && !nextCategories.includes(current) ? null : current)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const sortPromptItems = (items: PromptItem[]) => [...items].sort((left, right) => {
    const result = collator.compare(left.name.trim(), right.name.trim())
    return sortOrder === 'asc' ? result : -result
  })

  const filteredPrompts = useMemo(() => {
    const items = prompts.filter((item) => selectedCat === null || item.category === selectedCat)
    const searched = search.trim()
      ? items.filter((item) => matchesKeywordExpression(search, [
        item.name,
        item.description,
        item.content,
        (item.webLinks ?? []).map(link => `${link.label} ${link.url}`).join('\n'),
      ].join('\n')))
      : items
    return sortPromptItems(searched)
  }, [prompts, search, selectedCat, sortOrder])

  const exportScopePrompts = useMemo(() => sortPromptItems(listPromptExportCandidates(prompts, selectedCat)), [prompts, selectedCat, sortOrder])

  const exportActions = useMemo(() => buildPromptExportActions(selectedCat, {
    all: t('category.all'),
    selected: t('prompts.exportSpecified'),
  }), [selectedCat, t])

  const promptCounts = useMemo(() => prompts.reduce<Record<string, number>>((counts, item) => {
    counts[item.category] = (counts[item.category] ?? 0) + 1
    return counts
  }, {}), [prompts])

  useEffect(() => {
    setSelectedExportNames((current) => {
      if (current.size === 0) return current
      const next = new Set(exportScopePrompts.filter((item) => current.has(item.name)).map((item) => item.name))
      return next.size === current.size ? current : next
    })
  }, [exportScopePrompts])

  const formatPromptError = (message: unknown) => {
    const raw = String((message as any)?.message ?? message ?? '').trim()
    if (!raw) return t('common.confirm')
    if (raw.includes('prompt content is empty')) return t('prompts.emptyContentError')
    if (raw.includes('prompt not found')) return t('prompts.notFoundError')
    if (raw.includes('prompt already exists')) return t('prompts.duplicateNameError')
    if (raw.includes('invalid prompt name')) return t('prompts.invalidNameError')
    if (raw.includes('prompt has too many image urls')) return t('prompts.tooManyImagesError')
    if (raw.includes('prompt image url is invalid')) return t('prompts.invalidImageError')
    if (raw.includes('prompt web link is invalid')) return t('prompts.invalidWebLinkError')
    return raw
  }

  const openCreateDialog = () => {
    setSaveError('')
    setEditorMode('create')
    setDraft(createEmptyDraft(selectedCat ?? defaultCategoryName))
    setEditorOpen(true)
  }

  const openEditDialog = (item: PromptItem) => {
    setSaveError('')
    setEditorMode('edit')
    setDraft({
      originalName: item.name,
      name: item.name,
      description: item.description ?? '',
      category: item.category || defaultCategoryName,
      content: item.content,
      imageURLs: item.imageURLs ?? [],
      webLinks: item.webLinks ?? [],
    })
    setEditorOpen(true)
  }

  const handleSave = async (nextDraft: PromptDraft) => {
    setSaving(true)
    setSaveError('')
    try {
      const imageURLs = normalizePromptImageURLs(nextDraft.imageURLs)
      const webLinksMarkdown = buildPromptLinksMarkdown(nextDraft.webLinks)
      if (editorMode === 'create') {
        await CreatePrompt(nextDraft.name, nextDraft.description, nextDraft.category, nextDraft.content, imageURLs, webLinksMarkdown)
      } else {
        await UpdatePrompt(nextDraft.originalName ?? nextDraft.name, nextDraft.name, nextDraft.description, nextDraft.category, nextDraft.content, imageURLs, webLinksMarkdown)
      }
      await load()
      setEditorOpen(false)
      setDraft(createEmptyDraft(selectedCat ?? defaultCategoryName))
    } catch (error) {
      setSaveError(formatPromptError(error))
    } finally {
      setSaving(false)
    }
  }

  const handleCopy = async (item: PromptItem) => {
    try {
      await copyTextToClipboard(item.content)
      setCopiedName(item.name)
      window.setTimeout(() => {
        setCopiedName(current => current === item.name ? null : current)
      }, 1800)
    } catch {
      setCopiedName(null)
    }
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await DeletePrompt(deleteTarget.name)
      await load()
      setDeleteTarget(null)
    } finally {
      setDeleting(false)
    }
  }

  const handleMoveCategory = async (name: string, category: string) => {
    await MovePromptCategory(name, category || defaultCategoryName)
    await load()
  }

  const handleImport = async () => {
    setImporting(true)
    setStatus(null)
    try {
      const count = await ImportPrompts()
      if (count > 0) {
        await load()
        setStatus({ type: 'success', message: t('prompts.importedN', { count }) })
      }
    } catch (error) {
      setStatus({ type: 'error', message: formatPromptError(error) })
    } finally {
      setImporting(false)
    }
  }

  const handleExport = async () => {
    setStatus(null)
    if (exportBarOpen) {
      closeExportBar()
      return
    }
    setExportBarOpen(true)
  }

  const closeExportBar = () => {
    setExportBarOpen(false)
    setExportMode('all')
    setSelectedExportNames(new Set())
  }

  const runExport = async (mode: PromptExportMode) => {
    if (!canExportPromptSelection(mode, selectedExportNames)) {
      setStatus({ type: 'error', message: t('prompts.exportSelectAtLeastOne') })
      return
    }
    setExporting(true)
    setStatus(null)
    try {
      const names = resolvePromptExportNames(mode, prompts, selectedCat, selectedExportNames)
      const path = await ExportPromptsByNames(mode === 'all' ? [] : names)
      if (path) {
        setStatus({ type: 'success', message: t('prompts.exportedTo', { path }) })
        closeExportBar()
      }
    } catch (error) {
      setStatus({ type: 'error', message: formatPromptError(error) })
    } finally {
      setExporting(false)
    }
  }

  return (
    <div className="flex h-full overflow-hidden">
      <PromptCategoryPanel
        categories={categories}
        promptCounts={promptCounts}
        selected={selectedCat}
        draggingPromptName={draggingPromptName}
        onSelect={setSelectedCat}
        onCategoryDragStateChange={setCategoryDragActive}
        onDrop={handleMoveCategory}
        onRefresh={load}
      />

      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="px-6 py-4 flex flex-col gap-4" style={{ borderBottom: '1px solid var(--border-base)' }}>
          <div className="flex items-center gap-3 flex-wrap">
            <h2 className="text-sm font-medium flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
              <FileText size={14} /> {t('prompts.title')}
            </h2>
            <div className="flex-1" />
            <button
              onClick={handleImport}
              disabled={importing || exporting}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors disabled:opacity-50"
              style={{ color: 'var(--text-muted)' }}
            >
              <Upload size={14} /> {importing ? t('prompts.importing') : t('prompts.import')}
            </button>
            <button
              onClick={handleExport}
              disabled={importing || exporting}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors disabled:opacity-50"
              style={{ color: 'var(--text-muted)' }}
            >
              <Download size={14} /> {exporting ? t('prompts.exporting') : t('prompts.export')}
            </button>
            <button onClick={openCreateDialog} className="btn-primary flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg whitespace-nowrap">
              <Plus size={14} /> {t('prompts.add')}
            </button>
          </div>

          <SkillListControls
            search={search}
            onSearchChange={setSearch}
            sortOrder={sortOrder}
            onSortOrderChange={setSortOrder}
            placeholder={t('prompts.searchPlaceholder')}
            resultLabel={t('prompts.showingNPrompts', { count: filteredPrompts.length })}
          />

          {exportBarOpen && (
            <div className="rounded-2xl px-3 py-3 flex flex-col gap-3" style={{ border: '1px solid var(--border-base)', background: 'var(--bg-base)' }}>
              <div className="flex flex-wrap items-center gap-2">
                {exportActions.map((action) => (
                  <button
                    key={action.key}
                    type="button"
                    disabled={exporting}
                    onClick={() => {
                      if (action.key === 'selected') {
                        setExportMode('selected')
                        return
                      }
                      void runExport(action.key)
                    }}
                    className="px-3 py-1.5 rounded-lg text-sm transition-colors disabled:opacity-50"
                    style={exportMode === action.key
                      ? { background: 'var(--active-surface)', color: 'var(--active-text)', border: '1px solid var(--active-border)' }
                      : { color: 'var(--text-muted)', border: '1px solid var(--border-base)' }}
                  >
                    {action.label}
                  </button>
                ))}
                <div className="flex-1" />
                <button type="button" onClick={closeExportBar} disabled={exporting} className="btn-secondary text-sm">
                  {t('common.cancel')}
                </button>
              </div>

              {exportMode === 'selected' && (
                <>
                  <div className="text-sm" style={{ color: 'var(--text-secondary)' }}>
                    {t('prompts.exportSelectedCount', { count: selectedExportNames.size })}
                  </div>
                  <div className="max-h-44 overflow-y-auto rounded-xl" style={{ border: '1px solid var(--border-base)' }}>
                    {exportScopePrompts.map((item) => {
                      const checked = selectedExportNames.has(item.name)
                      return (
                        <label
                          key={item.name}
                          className="flex items-center gap-3 px-3 py-2 text-sm cursor-pointer"
                          style={{ color: 'var(--text-primary)', borderBottom: '1px solid var(--border-base)' }}
                        >
                          <input
                            type="checkbox"
                            checked={checked}
                            onChange={(event) => {
                              const next = new Set(selectedExportNames)
                              if (event.target.checked) next.add(item.name)
                              else next.delete(item.name)
                              setSelectedExportNames(next)
                            }}
                          />
                          <span className="min-w-0 flex-1 truncate">{item.name}</span>
                          <span className="shrink-0 text-[11px]" style={{ color: 'var(--text-muted)' }}>{item.category}</span>
                        </label>
                      )
                    })}
                  </div>
                  <div className="flex items-center justify-end gap-2">
                    <button
                      type="button"
                      onClick={() => void runExport('selected')}
                      disabled={exporting || !canExportPromptSelection('selected', selectedExportNames)}
                      className="btn-primary text-sm disabled:opacity-50"
                    >
                      {exporting ? t('prompts.exporting') : t('prompts.export')}
                    </button>
                  </div>
                </>
              )}
            </div>
          )}

          {status && (
            <div
              className="rounded-xl px-3 py-2 text-sm"
              style={status.type === 'success'
                ? { background: 'rgba(64, 181, 138, 0.12)', color: 'var(--color-success)', border: '1px solid rgba(64, 181, 138, 0.24)' }
                : { background: 'rgba(224, 123, 123, 0.12)', color: 'var(--color-error)', border: '1px solid rgba(224, 123, 123, 0.24)' }}
            >
              {status.message}
            </div>
          )}
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          {loading ? (
            <div className="flex items-center justify-center h-32 text-sm" style={{ color: 'var(--text-muted)' }}>{t('common.loading')}</div>
          ) : filteredPrompts.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-48" style={{ color: 'var(--text-muted)' }}>
              <FileText size={32} className="mb-2 opacity-30" />
              <p className="text-sm">{prompts.length === 0 ? t('prompts.emptyTitle') : t('prompts.noMatch')}</p>
            </div>
          ) : (
            <div className="grid gap-4 [grid-template-columns:repeat(auto-fill,minmax(220px,1fr))]">
              {filteredPrompts.map((item) => (
                <PromptCard
                  key={item.name}
                  item={item}
                  copied={copiedName === item.name}
                  dragging={draggingPromptName === item.name}
                  dropTargetActive={draggingPromptName === item.name && categoryDragActive}
                  onDragStateChange={(dragging) => {
                    setDraggingPromptName(dragging ? item.name : null)
                    if (!dragging) setCategoryDragActive(false)
                  }}
                  onOpen={() => openEditDialog(item)}
                  onCopy={() => handleCopy(item)}
                  onDelete={() => setDeleteTarget(item)}
                />
              ))}
            </div>
          )}
        </div>
      </div>

      <PromptEditorDialog
        open={editorOpen}
        mode={editorMode}
        draft={draft}
        categories={categories}
        saving={saving}
        error={saveError}
        onClose={() => {
          if (saving) return
          setEditorOpen(false)
          setSaveError('')
        }}
        onSave={handleSave}
      />

      <AnimatedDialog open={deleteTarget !== null} onClose={deleting ? undefined : () => setDeleteTarget(null)} width="w-[420px]" zIndex={65}>
        <div className="flex items-center gap-2 mb-3">
          <Trash2 size={18} style={{ color: 'var(--color-error)' }} />
          <span className="text-base font-semibold" style={{ color: 'var(--text-primary)' }}>{t('prompts.deleteConfirmTitle')}</span>
        </div>
        <p className="text-sm leading-6" style={{ color: 'var(--text-secondary)' }}>{t('prompts.deleteConfirmDesc')}</p>
        {deleteTarget && (
          <div className="mt-4 rounded-xl px-3 py-3 text-sm" style={{ background: 'var(--bg-base)', color: 'var(--text-primary)', border: '1px solid var(--border-base)' }}>
            <div className="font-medium mb-1">{deleteTarget.name}</div>
            {deleteTarget.description && (
              <div className="text-xs mb-2" style={{ color: 'var(--text-muted)' }}>{deleteTarget.description}</div>
            )}
            <div className="whitespace-pre-line">{getPromptPreview(deleteTarget.content).preview}</div>
          </div>
        )}
        <div className="mt-6 flex items-center justify-end gap-3">
          <button onClick={() => setDeleteTarget(null)} disabled={deleting} className="btn-secondary">
            {t('common.cancel')}
          </button>
          <button onClick={handleDelete} disabled={deleting} className="btn-primary" style={{ background: 'var(--color-error)' }}>
            {deleting ? t('common.delete') : t('common.confirm')}
          </button>
        </div>
      </AnimatedDialog>
    </div>
  )
}

type PromptCardProps = {
  item: PromptItem
  copied: boolean
  dragging: boolean
  dropTargetActive: boolean
  onDragStateChange: (dragging: boolean) => void
  onOpen: () => void
  onCopy: () => void
  onDelete: () => void
}

function PromptCard({ item, copied, dragging, dropTargetActive, onDragStateChange, onOpen, onCopy, onDelete }: PromptCardProps) {
  const { t } = useLanguage()
  const { preview, truncated } = getPromptPreview(item.content)
  const tooltipTitle = item.description ? `${item.name}
${item.description}` : item.name
  const stopActionPointer = (event: React.PointerEvent<HTMLButtonElement>) => {
    event.stopPropagation()
  }

  if (dragging && dropTargetActive) {
    return (
      <div className="relative min-h-[76px] rounded-xl border border-transparent bg-transparent">
        <div className="absolute inset-x-4 top-1/2 h-[2px] -translate-y-1/2 rounded-full" style={{ background: 'var(--accent-primary)', boxShadow: 'var(--glow-accent-sm)' }} />
      </div>
    )
  }

  return (
    <div
      draggable
      title={tooltipTitle}
      onDragStart={(event) => {
        event.dataTransfer.setData('text/plain', item.name)
        event.dataTransfer.setData('application/x-skillflow-prompt-name', item.name)
        event.dataTransfer.effectAllowed = 'move'
        onDragStateChange(true)
      }}
      onDragEnd={() => onDragStateChange(false)}
      onClick={onOpen}
      className={`card-base relative flex min-h-[118px] flex-col p-4 group cursor-grab ${dragging ? 'opacity-55 scale-[0.96]' : ''}`}
    >
      <button
        type="button"
        draggable={false}
        onPointerDown={stopActionPointer}
        onDragStart={(event) => event.preventDefault()}
        onClick={(event) => {
          event.stopPropagation()
          onCopy()
        }}
        title={copied ? t('prompts.copied') : t('prompts.copy')}
        className="absolute top-2 right-2 z-10 rounded p-1 opacity-0 transition-opacity group-hover:opacity-100"
        style={{ color: copied ? 'var(--color-success)' : 'var(--text-muted)' }}
      >
        {copied ? <Check size={12} /> : <Copy size={12} />}
      </button>

      <div className="min-w-0 pr-5">
        <p className="truncate text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{item.name}</p>
        {item.description ? (
          <p className="mt-1 truncate text-[11px]" style={{ color: 'var(--text-secondary)' }}>{item.description}</p>
        ) : (
          <div className="mt-1 h-[16px]" />
        )}

        {preview && (
          <p className="mt-3 line-clamp-2 text-[11px]" style={{ color: 'var(--text-muted)' }}>{preview}</p>
        )}
      </div>

      <div className="mt-auto flex min-w-0 items-center gap-2 pt-3">
        <span className="max-w-[96px] shrink-0 truncate rounded px-1.5 py-0.5 text-[10px]" style={{ color: 'var(--text-muted)', background: 'var(--bg-base)', border: '1px solid var(--border-base)' }}>
          {item.category}
        </span>
        {truncated && (
          <span className="truncate text-[11px]" style={{ color: 'var(--accent-primary)' }}>{t('prompts.viewMore')}</span>
        )}
        <button
          type="button"
          draggable={false}
          onPointerDown={stopActionPointer}
          onDragStart={(event) => event.preventDefault()}
          onClick={(event) => {
            event.stopPropagation()
            onDelete()
          }}
          title={t('common.delete')}
          className="ml-auto rounded p-1 opacity-0 transition-opacity group-hover:opacity-100"
          style={{ color: 'var(--color-error)' }}
        >
          <Trash2 size={12} />
        </button>
      </div>
    </div>
  )
}
