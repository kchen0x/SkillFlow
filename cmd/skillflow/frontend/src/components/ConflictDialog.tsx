import { useEffect } from 'react'
import { useLanguage } from '../contexts/LanguageContext'
import AnimatedDialog from './ui/AnimatedDialog'

interface Props<T> {
  conflicts: T[]
  onOverwrite: (conflict: T) => void
  onSkip: (conflict: T) => void
  onDone: () => void
  onCancel?: () => void
  labelForConflict?: (conflict: T) => string
  applyToRemainingLabel?: string
  applyToRemainingChecked?: boolean
  onApplyToRemainingChange?: (checked: boolean) => void
}

export default function ConflictDialog<T>({
  conflicts,
  onOverwrite,
  onSkip,
  onDone,
  onCancel,
  labelForConflict,
  applyToRemainingLabel,
  applyToRemainingChecked = false,
  onApplyToRemainingChange,
}: Props<T>) {
  const { t } = useLanguage()

  useEffect(() => {
    if (conflicts.length === 0) {
      onDone()
    }
  }, [conflicts, onDone])

  if (conflicts.length === 0) {
    return null
  }

  const current = conflicts[0]
  const label = labelForConflict ? labelForConflict(current) : String(current)

  return (
    <AnimatedDialog open={true} onClose={onCancel} width="w-96" zIndex={50}>
      <h3 className="text-base font-semibold mb-2" style={{ color: 'var(--text-primary)' }}>
        {t('conflictDialog.title')}
      </h3>
      <p className="text-sm mb-6" style={{ color: 'var(--text-muted)' }}>
        <span className="font-medium" style={{ color: 'var(--text-primary)' }}>{label}</span> {t('conflictDialog.existsSuffix')}
      </p>
      {applyToRemainingLabel && onApplyToRemainingChange && (
        <label className="mb-4 flex items-start gap-3 text-sm cursor-pointer" style={{ color: 'var(--text-secondary)' }}>
          <input
            type="checkbox"
            checked={applyToRemainingChecked}
            onChange={(event) => onApplyToRemainingChange(event.target.checked)}
            className="mt-0.5"
          />
          <span>{applyToRemainingLabel}</span>
        </label>
      )}
      <div className="flex gap-3 justify-end">
        {onCancel && (
          <button onClick={onCancel} className="btn-secondary px-4 py-2 text-sm rounded-lg">{t('common.cancel')}</button>
        )}
        <button onClick={() => onSkip(current)} className="btn-secondary px-4 py-2 text-sm rounded-lg">{t('conflictDialog.skip')}</button>
        <button onClick={() => onOverwrite(current)} className="btn-primary px-4 py-2 text-sm rounded-lg">{t('conflictDialog.overwrite')}</button>
      </div>
    </AnimatedDialog>
  )
}
