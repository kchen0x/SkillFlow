function stripWrappingQuotes(term: string) {
  return term.replace(/^['"]+|['"]+$/g, '').trim()
}

export function matchesKeywordExpression(query: string, rawText: string) {
  const normalizedQuery = query.trim().toLocaleLowerCase()
  if (!normalizedQuery) return true

  const normalizedText = rawText.toLocaleLowerCase()
  const hasLogicalOperator = /\b(?:and|or)\b/i.test(normalizedQuery)
  if (!hasLogicalOperator) {
    return normalizedText.includes(normalizedQuery)
  }

  const orGroups = normalizedQuery
    .split(/\s+or\s+/i)
    .map(group => group.trim())
    .filter(Boolean)

  return orGroups.some((group) => {
    const andTerms = group
      .split(/\s+and\s+/i)
      .map(stripWrappingQuotes)
      .filter(Boolean)

    return andTerms.length > 0 && andTerms.every(term => normalizedText.includes(term))
  })
}
