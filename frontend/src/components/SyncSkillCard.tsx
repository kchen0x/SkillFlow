import { FolderOpen, Github, FolderOpenDot } from 'lucide-react'
import { OpenPath } from '../../wailsjs/go/main/App'

interface Props {
  name: string
  subtitle?: string     // e.g. category or path hint
  source?: string       // 'github' | 'manual' | undefined
  path?: string         // filesystem path to open in file manager
  selected: boolean
  onToggle: () => void
}

export default function SyncSkillCard({ name, subtitle, source, path, selected, onToggle }: Props) {
  const handleOpen = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (path) OpenPath(path)
  }

  return (
    <div
      onClick={onToggle}
      className={`relative flex flex-col gap-2 p-3 rounded-xl border cursor-pointer transition-colors select-none ${
        selected
          ? 'bg-indigo-900/30 border-indigo-500'
          : 'bg-gray-800 border-gray-700 hover:border-gray-500'
      }`}
    >
      {/* Top-right: open folder button */}
      {path && (
        <button
          onClick={handleOpen}
          title="打开目录"
          className="absolute top-2 right-2 p-1 rounded text-gray-500 hover:text-gray-200 hover:bg-gray-700 transition-colors"
        >
          <FolderOpenDot size={13} />
        </button>
      )}

      {/* Source badge row */}
      <div className="flex items-center gap-1.5 pr-6">
        {source === 'github'
          ? <Github size={12} className="text-gray-400 shrink-0" />
          : <FolderOpen size={12} className="text-gray-400 shrink-0" />}
        {source && (
          <span className={`text-xs px-1.5 py-0.5 rounded ${
            source === 'github' ? 'bg-blue-900/50 text-blue-300' : 'bg-gray-700 text-gray-400'
          }`}>{source}</span>
        )}
      </div>

      {/* Skill name */}
      <p className="text-sm font-medium leading-snug truncate pr-5">{name}</p>

      {/* Subtitle (category or path) */}
      {subtitle && (
        <p className="text-xs text-gray-500 truncate">{subtitle}</p>
      )}

      {/* Selection indicator */}
      <div className={`absolute bottom-2 right-2 w-4 h-4 rounded border-2 flex items-center justify-center transition-colors ${
        selected ? 'bg-indigo-500 border-indigo-500' : 'border-gray-600 bg-gray-700'
      }`}>
        {selected && (
          <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
          </svg>
        )}
      </div>
    </div>
  )
}
