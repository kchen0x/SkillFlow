import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('../src/style.css', import.meta.url), 'utf8')

function getThemeVars(themeName) {
  const blockMatch = source.match(new RegExp(`\\[data-theme="${themeName}"\\] \\{([\\s\\S]*?)\\n\\}`))
  assert.ok(blockMatch, `expected to find ${themeName} theme block in style.css`)

  const vars = {}
  for (const match of blockMatch[1].matchAll(/--([\w-]+):\s*([^;]+);/g)) {
    vars[match[1]] = match[2].trim()
  }
  return vars
}

function parseColor(input, background = [255, 255, 255]) {
  const value = input.trim()

  if (value.startsWith('linear-gradient') || value.startsWith('radial-gradient')) {
    return background
  }

  if (value.startsWith('#')) {
    const hex = value.slice(1)
    if (hex.length === 6) {
      return [
        Number.parseInt(hex.slice(0, 2), 16),
        Number.parseInt(hex.slice(2, 4), 16),
        Number.parseInt(hex.slice(4, 6), 16),
      ]
    }
    if (hex.length === 3) {
      return [
        Number.parseInt(hex[0] + hex[0], 16),
        Number.parseInt(hex[1] + hex[1], 16),
        Number.parseInt(hex[2] + hex[2], 16),
      ]
    }
  }

  const rgbaMatch = value.match(/rgba?\(([^)]+)\)/)
  assert.ok(rgbaMatch, `unsupported color format: ${value}`)

  const parts = rgbaMatch[1].split(',').map(part => part.trim())
  const [r, g, b] = parts.slice(0, 3).map(Number)
  const alpha = parts[3] === undefined ? 1 : Number(parts[3])

  return [
    r * alpha + background[0] * (1 - alpha),
    g * alpha + background[1] * (1 - alpha),
    b * alpha + background[2] * (1 - alpha),
  ]
}

function relativeLuminance([r, g, b]) {
  const linear = [r, g, b].map(channel => {
    const normalized = channel / 255
    if (normalized <= 0.03928) {
      return normalized / 12.92
    }
    return ((normalized + 0.055) / 1.055) ** 2.4
  })

  return 0.2126 * linear[0] + 0.7152 * linear[1] + 0.0722 * linear[2]
}

function contrastRatio(firstColor, secondColor) {
  const first = relativeLuminance(firstColor)
  const second = relativeLuminance(secondColor)
  const lighter = Math.max(first, second)
  const darker = Math.min(first, second)
  return (lighter + 0.05) / (darker + 0.05)
}

for (const themeName of ['young', 'light']) {
  test(`${themeName} theme keeps text, button, and surface contrast above the accessibility floor`, () => {
    const vars = getThemeVars(themeName)
    const bgBase = parseColor(vars['bg-base'])
    const bgElevated = parseColor(vars['bg-elevated'])
    const btnPrimaryBg = parseColor(vars['btn-primary-bg'])
    const btnPrimaryHover = parseColor(vars['btn-primary-hover'])
    const borderBase = parseColor(vars['border-base'], bgElevated)
    const activeSurface = parseColor(vars['active-surface'], bgBase)
    const activeText = parseColor(vars['active-text'], activeSurface)
    const textMuted = parseColor(vars['text-muted'])
    const buttonText = parseColor('#ffffff')

    assert.ok(
      contrastRatio(textMuted, bgBase) >= 4.5,
      `${themeName} text-muted should keep at least 4.5:1 contrast on bg-base`,
    )
    assert.ok(
      contrastRatio(textMuted, bgElevated) >= 4.5,
      `${themeName} text-muted should keep at least 4.5:1 contrast on bg-elevated`,
    )
    assert.ok(
      contrastRatio(buttonText, btnPrimaryBg) >= 4.5,
      `${themeName} primary button label should keep at least 4.5:1 contrast`,
    )
    assert.ok(
      contrastRatio(buttonText, btnPrimaryHover) >= 4.5,
      `${themeName} primary button hover label should keep at least 4.5:1 contrast`,
    )
    assert.ok(
      contrastRatio(borderBase, bgElevated) >= 1.25,
      `${themeName} elevated surfaces should keep a visible border against the background`,
    )
    assert.ok(
      contrastRatio(activeSurface, bgBase) >= 1.2,
      `${themeName} active surface should remain visibly distinct from the base background`,
    )
    assert.ok(
      contrastRatio(activeText, activeSurface) >= 4.5,
      `${themeName} active text should stay readable on the active surface`,
    )
  })
}
