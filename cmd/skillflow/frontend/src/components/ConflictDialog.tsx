import { useLanguage } from '../contexts/LanguageContext'
import AnimatedDialog from './ui/AnimatedDialog'

interface Props {
  conflicts: string[]
  onOverwrite: (name: string) => void
  onSkip: (name: string) => void
  onDone: () => void
}

export default function ConflictDialog({ conflicts, onOverwrite, onSkip, onDone }: Props) {
  const { t } = useLanguage()

  if (conflicts.length === 0) {
    onDone()
    return null
  }

  const current = conflicts[0]

  return (
    <AnimatedDialog open={true} width="w-96" zIndex={50}>
      <h3 className="text-base font-semibold mb-2" style={{ color: 'var(--text-primary)' }}>
        {t('conflictDialog.title')}
      </h3>
      <p className="text-sm mb-6" style={{ color: 'var(--text-muted)' }}>
        <span className="font-medium" style={{ color: 'var(--text-primary)' }}>{current}</span> {t('conflictDialog.existsSuffix')}
      </p>
      <div className="flex gap-3 justify-end">
        <button onClick={() => onSkip(current)} className="btn-secondary px-4 py-2 text-sm rounded-lg">{t('conflictDialog.skip')}</button>
        <button onClick={() => onOverwrite(current)} className="btn-primary px-4 py-2 text-sm rounded-lg">{t('conflictDialog.overwrite')}</button>
      </div>
    </AnimatedDialog>
  )
}
