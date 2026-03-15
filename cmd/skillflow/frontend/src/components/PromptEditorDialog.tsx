import { useEffect, useMemo, useState } from 'react'
import { FileText, Image as ImageIcon, Link2, Sparkles, Trash2, X } from 'lucide-react'
import { OpenURL } from '../../wailsjs/go/main/App'
import AnimatedDialog from './ui/AnimatedDialog'
import { useLanguage } from '../contexts/LanguageContext'
import { appendPromptImageURL, normalizePromptImageURLs, normalizePromptPreviewImageURL, parsePromptWebLinkLine, type PromptWebLink } from '../lib/promptRichContent'

export type PromptDraft = {
  originalName?: string
  name: string
  description: string
  category: string
  content: string
  imageURLs: string[]
  webLinks: PromptWebLink[]
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
  const [imageURLInput, setImageURLInput] = useState('')
  const [imageError, setImageError] = useState('')
  const [webLinkInput, setWebLinkInput] = useState('')
  const [webLinkError, setWebLinkError] = useState('')
  const [previewImageURL, setPreviewImageURL] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    setForm(draft)
    setImageURLInput('')
    setImageError('')
    setWebLinkInput('')
    setWebLinkError('')
    setPreviewImageURL(null)
  }, [draft, open])

  const lineCount = useMemo(() => {
    if (!form.content) return 1
    return form.content.replace(/\r\n/g, '\n').split('\n').length
  }, [form.content])
  const previewImages = useMemo(() => normalizePromptImageURLs(form.imageURLs ?? []), [form.imageURLs])

  const categoryOptions = useMemo(() => {
    const values = new Set(['Default', ...categories, form.category || 'Default'])
    return [...values]
  }, [categories, form.category])

  const addImageURL = () => {
    const normalizedURL = normalizePromptPreviewImageURL(imageURLInput)
    if (!normalizedURL) {
      setImageError(t('prompts.invalidImageError'))
      return
    }
    const nextImageURLs = appendPromptImageURL(form.imageURLs, normalizedURL)
    if (!nextImageURLs) {
      setImageError(t('prompts.tooManyImagesError'))
      return
    }
    setForm((current) => ({ ...current, imageURLs: nextImageURLs }))
    setImageURLInput('')
    setImageError('')
  }

  const addWebLink = () => {
    const nextLink = parsePromptWebLinkLine(webLinkInput)
    if (!nextLink) {
      setWebLinkError(t('prompts.invalidWebLinkError'))
      return
    }
    setForm((current) => ({ ...current, webLinks: [...current.webLinks, nextLink] }))
    setWebLinkInput('')
    setWebLinkError('')
  }

  return (
    <>
      <AnimatedDialog open={open} onClose={saving ? undefined : onClose} width="w-[820px] max-w-[calc(100vw-2rem)]" zIndex={60}>
        <div className="flex max-h-[min(82vh,760px)] flex-col">
        <div className="mb-5 flex items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="mb-3 inline-flex items-center gap-2 rounded-full px-3 py-1 text-[11px] font-medium uppercase tracking-[0.16em]" style={{ background: 'var(--accent-glow)', color: 'var(--accent-primary)', border: '1px solid var(--border-accent)' }}>
              <Sparkles size={12} />
              {t('prompts.systemFile')}
            </div>
            <div className="mb-1 flex items-center gap-2">
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
        <div className="min-h-0 flex-1 overflow-y-auto pr-1">
          <div className="mb-4 grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <label className="mb-2 block text-xs font-medium" style={{ color: 'var(--text-muted)' }}>{t('prompts.editorNameLabel')}</label>
              <input
                value={form.name}
                onChange={(event) => setForm(current => ({ ...current, name: event.target.value }))}
                placeholder={t('prompts.editorNamePlaceholder')}
                className="input-base"
              />
            </div>
            <div>
              <label className="mb-2 block text-xs font-medium" style={{ color: 'var(--text-muted)' }}>{t('prompts.editorCategoryLabel')}</label>
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
            <label className="mb-2 block text-xs font-medium" style={{ color: 'var(--text-muted)' }}>{t('prompts.editorDescriptionLabel')}</label>
            <input
              value={form.description}
              onChange={(event) => setForm(current => ({ ...current, description: event.target.value }))}
              placeholder={t('prompts.editorDescriptionPlaceholder')}
              className="input-base"
            />
          </div>

          <div className="mb-4">
            <div className="mb-2 flex items-center justify-between gap-3 text-xs" style={{ color: 'var(--text-muted)' }}>
              <span className="inline-flex items-center gap-2">
                <ImageIcon size={13} />
                {t('prompts.editorImageLabel')}
              </span>
              <span>{t('prompts.editorImageCount', { count: previewImages.length, max: 3 })}</span>
            </div>

            <div className="space-y-2">
              {form.imageURLs.length === 0 && (
                <div className="rounded-2xl border border-dashed px-3 py-3 text-sm" style={{ borderColor: 'var(--border-base)', color: 'var(--text-muted)', background: 'var(--bg-base)' }}>
                  {t('prompts.editorImageEmpty')}
                </div>
              )}
            </div>

            {previewImages.length > 0 && (
              <div className="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-3">
                {previewImages.map((imageURL, index) => (
                  <div
                    key={imageURL}
                    className="group relative overflow-hidden rounded-2xl transition-transform hover:scale-[1.01]"
                    style={{ border: '1px solid var(--border-base)', background: 'var(--bg-base)' }}
                  >
                    <button
                      type="button"
                      onClick={() => setPreviewImageURL(normalizePromptPreviewImageURL(imageURL))}
                      className="block w-full"
                      title={imageURL}
                    >
                      <img src={imageURL} alt={t('prompts.editorImagePreviewAlt')} className="h-28 w-full cursor-zoom-in object-cover" />
                    </button>
                    <button
                      type="button"
                      onClick={() => setForm((current) => ({
                        ...current,
                        imageURLs: current.imageURLs.filter((_, imageIndex) => imageIndex !== index),
                      }))}
                      className="absolute right-2 top-2 rounded-full p-1.5 opacity-0 pointer-events-none transition-all duration-150 group-hover:opacity-100 group-hover:pointer-events-auto group-focus-within:opacity-100 group-focus-within:pointer-events-auto focus:opacity-100 focus:pointer-events-auto"
                      style={{ color: 'white', background: 'rgba(15, 23, 42, 0.72)' }}
                      title={t('common.delete')}
                    >
                      <Trash2 size={12} />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="mb-4">
            <div className="mb-2 flex items-center justify-between gap-3 text-xs" style={{ color: 'var(--text-muted)' }}>
              <span className="inline-flex items-center gap-2">
                <Link2 size={13} />
                {t('prompts.editorLinkLabel')}
              </span>
              <span>{t('prompts.editorLinkHint')}</span>
            </div>

            {form.webLinks.length > 0 && (
              <div className="mt-3 flex flex-wrap gap-2">
                {form.webLinks.map((link, index) => (
                  <div
                    key={`${link.label}-${link.url}-${index}`}
                    className="inline-flex max-w-full items-center overflow-hidden rounded-full"
                    style={{ color: 'var(--accent-primary)', background: 'var(--accent-glow)', border: '1px solid var(--border-accent)' }}
                  >
                    <button
                      type="button"
                      onClick={() => OpenURL(link.url)}
                      className="inline-flex min-w-0 items-center gap-2 px-3 py-1.5 text-xs transition-colors"
                      title={link.url}
                    >
                      <Link2 size={12} />
                      <span className="truncate">{link.label}</span>
                    </button>
                    <button
                      type="button"
                      onClick={() => setForm((current) => ({
                        ...current,
                        webLinks: current.webLinks.filter((_, linkIndex) => linkIndex !== index),
                      }))}
                      className="border-l px-2 py-1.5 transition-colors"
                      style={{ borderColor: 'var(--border-accent)' }}
                      title={t('common.delete')}
                    >
                      <X size={12} />
                    </button>
                  </div>
                ))}
              </div>
            )}
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
              className="input-base min-h-[220px] max-h-[340px] resize-y overflow-y-auto border-0 bg-transparent p-3 font-mono text-sm leading-6"
              style={{ boxShadow: 'none' }}
            />
          </div>

          <div className="mt-4 space-y-4">
            <div>
              <div className="mb-2 flex items-center gap-2 text-xs" style={{ color: 'var(--text-muted)' }}>
                <ImageIcon size={13} />
                <span>{t('prompts.editorImageLabel')}</span>
              </div>
              <div className="rounded-2xl p-3" style={{ background: 'var(--bg-base)', border: '1px solid var(--border-base)' }}>
                <div className="flex items-center gap-2">
                  <input
                    value={imageURLInput}
                    onChange={(event) => {
                      setImageURLInput(event.target.value)
                      if (imageError) {
                        setImageError('')
                      }
                    }}
                    onKeyDown={(event) => {
                      if (event.key !== 'Enter') return
                      event.preventDefault()
                      addImageURL()
                    }}
                    placeholder={t('prompts.editorImagePlaceholder')}
                    className="input-base border-0 bg-transparent text-sm"
                    style={{ boxShadow: 'none' }}
                  />
                  <button
                    type="button"
                    onClick={addImageURL}
                    disabled={!imageURLInput.trim()}
                    className="btn-secondary shrink-0 text-sm disabled:opacity-50"
                  >
                    {t('common.add')}
                  </button>
                </div>
              </div>

              {imageError && (
                <div className="mt-2 rounded-xl px-3 py-2 text-sm" style={{ background: 'rgba(224, 123, 123, 0.12)', color: 'var(--color-error)', border: '1px solid rgba(224, 123, 123, 0.24)' }}>
                  {imageError}
                </div>
              )}
            </div>

            <div>
              <div className="mb-2 flex items-center gap-2 text-xs" style={{ color: 'var(--text-muted)' }}>
                <Link2 size={13} />
                <span>{t('prompts.editorLinkLabel')}</span>
              </div>
              <div className="rounded-2xl p-3" style={{ background: 'var(--bg-base)', border: '1px solid var(--border-base)' }}>
                <div className="flex items-center gap-2">
                  <input
                    value={webLinkInput}
                    onChange={(event) => {
                      setWebLinkInput(event.target.value)
                      if (webLinkError) {
                        setWebLinkError('')
                      }
                    }}
                    onKeyDown={(event) => {
                      if (event.key !== 'Enter') return
                      event.preventDefault()
                      addWebLink()
                    }}
                    placeholder={t('prompts.editorLinkPlaceholder')}
                    className="input-base border-0 bg-transparent font-mono text-sm"
                    style={{ boxShadow: 'none' }}
                  />
                  <button
                    type="button"
                    onClick={addWebLink}
                    disabled={!webLinkInput.trim()}
                    className="btn-secondary shrink-0 text-sm disabled:opacity-50"
                  >
                    {t('common.add')}
                  </button>
                </div>
              </div>

              {webLinkError && (
                <div className="mt-2 rounded-xl px-3 py-2 text-sm" style={{ background: 'rgba(224, 123, 123, 0.12)', color: 'var(--color-error)', border: '1px solid rgba(224, 123, 123, 0.24)' }}>
                  {webLinkError}
                </div>
              )}
            </div>
          </div>
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
        </div>
      </AnimatedDialog>

      <AnimatedDialog open={previewImageURL !== null} onClose={() => setPreviewImageURL(null)} width="w-[min(92vw,1080px)]" zIndex={70}>
        <div className="flex max-h-[82vh] flex-col">
          <div className="mb-4 flex items-center justify-between gap-3">
            <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
              {t('prompts.editorImageLabel')}
            </span>
            <button
              type="button"
              onClick={() => setPreviewImageURL(null)}
              className="rounded-lg p-1.5 transition-colors"
              style={{ color: 'var(--text-muted)' }}
              title={t('common.close')}
            >
              <X size={16} />
            </button>
          </div>

          {previewImageURL && (
            <div className="flex min-h-0 flex-1 items-center justify-center overflow-hidden rounded-2xl" style={{ background: 'var(--bg-base)', border: '1px solid var(--border-base)' }}>
              <img
                src={previewImageURL}
                alt={t('prompts.editorImagePreviewAlt')}
                className="max-h-[70vh] w-full object-contain"
              />
            </div>
          )}
        </div>
      </AnimatedDialog>
    </>
  )
}
