import { useEffect, useRef } from 'react'

interface MenuItem { label: string; onClick: () => void; danger?: boolean }
interface Props { x: number; y: number; items: MenuItem[]; onClose: () => void }

export default function ContextMenu({ x, y, items, onClose }: Props) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose()
    }
    document.addEventListener('click', handler)
    return () => document.removeEventListener('click', handler)
  }, [onClose])

  return (
    <div
      ref={ref}
      onClick={e => e.stopPropagation()}
      style={{
        position: 'fixed', top: y, left: x, zIndex: 9999,
        background: 'var(--bg-elevated)',
        border: '1px solid var(--border-accent)',
        boxShadow: 'var(--shadow-dialog), var(--glow-accent-sm)',
      }}
      className="rounded-lg py-1 min-w-36"
    >
      {items.map((item, i) => (
        <button
          key={i}
          onClick={() => { item.onClick(); onClose() }}
          className="w-full text-left px-4 py-2 text-sm transition-colors"
          style={{ color: item.danger ? 'var(--color-error)' : 'var(--text-secondary)' }}
          onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)' }}
          onMouseLeave={e => { e.currentTarget.style.backgroundColor = '' }}
        >
          {item.label}
        </button>
      ))}
    </div>
  )
}
