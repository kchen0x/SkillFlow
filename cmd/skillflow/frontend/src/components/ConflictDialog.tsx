import { useLanguage } from '../contexts/LanguageContext'
import AnimatedDialog from './ui/AnimatedDialog'

interface Props<T> {
  conflicts: T[]
  onOverwrite: (conflict: T) => void
  onSkip: (conflict: T) => void
  onDone: () => void
  labelForConflict?: (conflict: T) => string
}

export default function ConflictDialog<T>({ conflicts, onOverwrite, onSkip, onDone, labelForConflict }: Props<T>) {
  const { t } = useLanguage()

  if (conflicts.length === 0) {
    onDone()
    return null
  }

  const current = conflicts[0]
  const label = labelForConflict ? labelForConflict(current) : String(current)

  return (
    <AnimatedDialog open={true} width="w-96" zIndex={50}>
      <h3 className="text-base font-semibold mb-2" style={{ color: 'var(--text-primary)' }}>
        {t('conflictDialog.title')}
      </h3>
      <p className="text-sm mb-6" style={{ color: 'var(--text-muted)' }}>
        <span className="font-medium" style={{ color: 'var(--text-primary)' }}>{label}</span> {t('conflictDialog.existsSuffix')}
      </p>
      <div className="flex gap-3 justify-end">
        <button onClick={() => onSkip(current)} className="btn-secondary px-4 py-2 text-sm rounded-lg">{t('conflictDialog.skip')}</button>
        <button onClick={() => onOverwrite(current)} className="btn-primary px-4 py-2 text-sm rounded-lg">{t('conflictDialog.overwrite')}</button>
      </div>
    </AnimatedDialog>
  )
}
