import AnimatedDialog from './ui/AnimatedDialog'

interface Props {
  conflicts: string[]
  onOverwrite: (name: string) => void
  onSkip: (name: string) => void
  onDone: () => void
}

export default function ConflictDialog({ conflicts, onOverwrite, onSkip, onDone }: Props) {
  if (conflicts.length === 0) { onDone(); return null }
  const current = conflicts[0]
  return (
    <AnimatedDialog open={true} width="w-96" zIndex={50}>
      <h3 className="text-base font-semibold mb-2" style={{ color: 'var(--text-primary)' }}>冲突检测</h3>
      <p className="text-sm mb-6" style={{ color: 'var(--text-muted)' }}>
        <span className="font-medium" style={{ color: 'var(--text-primary)' }}>{current}</span> 已存在，如何处理？
      </p>
      <div className="flex gap-3 justify-end">
        <button onClick={() => onSkip(current)} className="btn-secondary px-4 py-2 text-sm rounded-lg">跳过</button>
        <button onClick={() => onOverwrite(current)} className="btn-primary px-4 py-2 text-sm rounded-lg">覆盖</button>
      </div>
    </AnimatedDialog>
  )
}
