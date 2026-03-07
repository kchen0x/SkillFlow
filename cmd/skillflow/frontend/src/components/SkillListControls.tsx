import { ArrowDownAZ, ArrowUpAZ, Search } from 'lucide-react'
import { useLanguage } from '../contexts/LanguageContext'
import { SkillSortOrder } from '../lib/skillList'

type SkillListControlsProps = {
  search: string
  onSearchChange: (value: string) => void
  sortOrder: SkillSortOrder
  onSortOrderChange: (order: SkillSortOrder) => void
  placeholder?: string
  resultLabel?: string
  searchClassName?: string
}

const sortOptions: { value: SkillSortOrder; label: string; icon: JSX.Element }[] = [
  { value: 'asc', label: 'A-Z', icon: <ArrowUpAZ size={14} /> },
  { value: 'desc', label: 'Z-A', icon: <ArrowDownAZ size={14} /> },
]

export default function SkillListControls({
  search,
  onSearchChange,
  sortOrder,
  onSortOrderChange,
  placeholder,
  resultLabel,
  searchClassName = 'max-w-[520px]',
}: SkillListControlsProps) {
  const { t } = useLanguage()
  const effectivePlaceholder = placeholder ?? t('skillList.searchDefault')

  return (
    <div className="flex flex-1 flex-wrap items-center gap-3 min-w-[260px]">
      <div className={`relative flex-1 min-w-[220px] ${searchClassName}`}>
        <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2" style={{ color: 'var(--text-muted)' }} />
        <input
          type="search"
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder={effectivePlaceholder}
          className="input-base pl-10"
        />
      </div>

      <div
        className="flex items-center gap-1 rounded-xl p-1"
        style={{
          background: 'var(--bg-elevated)',
          border: '1px solid var(--border-base)',
        }}
      >
        {sortOptions.map((option) => {
          const active = sortOrder === option.value
          return (
            <button
              key={option.value}
              type="button"
              onClick={() => onSortOrderChange(option.value)}
              className="flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-sm transition-all duration-150"
              style={active ? {
                background: 'var(--active-surface)',
                color: 'var(--active-text)',
                border: '1px solid var(--active-border)',
                boxShadow: 'var(--active-shadow)',
              } : {
                color: 'var(--text-muted)',
                border: '1px solid transparent',
              }}
              title={option.value === 'asc' ? t('skillList.sortAscTitle') : t('skillList.sortDescTitle')}
            >
              {option.icon}
              <span className="hidden sm:inline">{option.label}</span>
            </button>
          )
        })}
      </div>

      {resultLabel && (
        <span className="text-xs whitespace-nowrap" style={{ color: 'var(--text-muted)' }}>
          {resultLabel}
        </span>
      )}
    </div>
  )
}
