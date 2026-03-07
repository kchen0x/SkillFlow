import { useEffect, useMemo, useState } from 'react'
import { FileText, Sparkles, X } from 'lucide-react'
import AnimatedDialog from './ui/AnimatedDialog'
import { useLanguage } from '../contexts/LanguageContext'

export type PromptDraft = {
  originalName?: string
  name: string
  description: string
  category: string
  content: string
}

type PromptEditorDialogProps = {
  open: boolean
  mode: 'create' | 'edit'
  draft: PromptDraft
  categories: string[]
  saving?: boolean
  error?: string
  onClose: () => void
  onSave: (draft: PromptDraft) => void
}

export default function PromptEditorDialog({
  open,
  mode,
  draft,
  categories,
  saving = false,
  error = '',
  onClose,
  onSave,
}: PromptEditorDialogProps) {
  const { t } = useLanguage()
  const [form, setForm] = useState<PromptDraft>(draft)

  useEffect(() => {
    if (!open) return
    setForm(draft)
  }, [draft, open])

  const lineCount = useMemo(() => {
    if (!form.content) return 1
    return form.content.replace(/\r\n/g, '\n').split('\n').length
  }, [form.content])

  const categoryOptions = useMemo(() => {
    const values = new Set(['Default', ...categories, form.category || 'Default'])
    return [...values]
  }, [categories, form.category])

  return (
    <AnimatedDialog open={open} onClose={saving ? undefined : onClose} width="w-[820px]" zIndex={60}>
      <div className="flex items-start justify-between gap-3 mb-5">
        <div className="min-w-0">
          <div className="inline-flex items-center gap-2 rounded-full px-3 py-1 text-[11px] font-medium tracking-[0.16em] uppercase mb-3" style={{ background: 'var(--accent-glow)', color: 'var(--accent-primary)', border: '1px solid var(--border-accent)' }}>
            <Sparkles size={12} />
            {t('prompts.systemFile')}
          </div>
          <div className="flex items-center gap-2 mb-1">
            <FileText size={18} style={{ color: 'var(--accent-primary)' }} />
            <h3 className="text-lg font-semibold" style={{ color: 'var(--text-primary)' }}>
              {mode === 'create' ? t('prompts.editorCreateTitle') : t('prompts.editorEditTitle')}
            </h3>
          </div>
        </div>

        <button onClick={onClose} disabled={saving} className="shrink-0 rounded-lg p-1.5 transition-colors disabled:opacity-40" style={{ color: 'var(--text-muted)' }}>
          <X size={16} />
        </button>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 mb-4">
        <div>
          <label className="block text-xs font-medium mb-2" style={{ color: 'var(--text-muted)' }}>{t('prompts.editorNameLabel')}</label>
          <input
            value={form.name}
            onChange={(event) => setForm(current => ({ ...current, name: event.target.value }))}
            placeholder={t('prompts.editorNamePlaceholder')}
            className="input-base"
          />
        </div>
        <div>
          <label className="block text-xs font-medium mb-2" style={{ color: 'var(--text-muted)' }}>{t('prompts.editorCategoryLabel')}</label>
          <select
            value={form.category || 'Default'}
            onChange={(event) => setForm(current => ({ ...current, category: event.target.value }))}
            className="select-base"
          >
            {categoryOptions.map((category) => (
              <option key={category} value={category}>{category}</option>
            ))}
          </select>
        </div>
      </div>

      <div className="mb-4">
        <label className="block text-xs font-medium mb-2" style={{ color: 'var(--text-muted)' }}>{t('prompts.editorDescriptionLabel')}</label>
        <input
          value={form.description}
          onChange={(event) => setForm(current => ({ ...current, description: event.target.value }))}
          placeholder={t('prompts.editorDescriptionPlaceholder')}
          className="input-base"
        />
      </div>

      <div className="mb-2 flex items-center justify-between gap-3 text-xs" style={{ color: 'var(--text-muted)' }}>
        <span>{t('prompts.editorContentLabel')}</span>
        <span>{t('prompts.editorStats', { count: form.content.length, lines: lineCount })}</span>
      </div>
      <div className="rounded-2xl p-3" style={{ background: 'var(--bg-base)', border: '1px solid var(--border-base)' }}>
        <textarea
          value={form.content}
          onChange={(event) => setForm(current => ({ ...current, content: event.target.value }))}
          placeholder={t('prompts.editorPlaceholder')}
          className="input-base min-h-[320px] resize-y border-0 bg-transparent p-3 font-mono text-sm leading-6"
          style={{ boxShadow: 'none' }}
        />
      </div>

      {error && (
        <div className="mt-4 rounded-xl px-3 py-2 text-sm" style={{ background: 'rgba(224, 123, 123, 0.12)', color: 'var(--color-error)', border: '1px solid rgba(224, 123, 123, 0.24)' }}>
          {error}
        </div>
      )}

      <div className="mt-6 flex items-center justify-end gap-3">
        <button onClick={onClose} disabled={saving} className="btn-secondary">
          {t('common.cancel')}
        </button>
        <button onClick={() => onSave(form)} disabled={saving} className="btn-primary min-w-[120px]">
          {saving ? t('common.saving') : mode === 'create' ? t('prompts.createAction') : t('common.save')}
        </button>
      </div>
    </AnimatedDialog>
  )
}
