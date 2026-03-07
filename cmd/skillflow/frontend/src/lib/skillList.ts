export type SkillSortOrder = 'asc' | 'desc'

const nameCollator = new Intl.Collator(undefined, {
  numeric: true,
  sensitivity: 'base',
})

export function filterAndSortSkills<T>(
  items: T[],
  search: string,
  sortOrder: SkillSortOrder,
  getName: (item: T) => string,
) {
  const keyword = search.trim().toLocaleLowerCase()

  const filtered = keyword
    ? items.filter((item) => getName(item).toLocaleLowerCase().includes(keyword))
    : items

  return [...filtered].sort((left, right) => {
    const result = nameCollator.compare(getName(left).trim(), getName(right).trim())
    return sortOrder === 'asc' ? result : -result
  })
}
