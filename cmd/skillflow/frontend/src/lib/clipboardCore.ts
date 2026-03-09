export type ClipboardFallbacks = {
  runtimeWriteText?: (text: string) => Promise<boolean | void>
  browserWriteText?: (text: string) => Promise<void>
  execCommandCopy?: (text: string) => boolean
}

export async function copyTextWithFallbacks(text: string, fallbacks: ClipboardFallbacks): Promise<void> {
  if (fallbacks.execCommandCopy?.(text)) {
    return
  }

  if (fallbacks.runtimeWriteText) {
    try {
      const ok = await fallbacks.runtimeWriteText(text)
      if (ok !== false) {
        return
      }
    } catch {
      // Fall through to browser/document clipboard fallbacks.
    }
  }

  if (fallbacks.browserWriteText) {
    await fallbacks.browserWriteText(text)
    return
  }

  throw new Error('clipboard unavailable')
}

type ClipboardDocument = {
  activeElement: Element | null
  body: {
    appendChild(node: unknown): void
    removeChild(node: unknown): void
  }
  createElement(tagName: string): {
    value: string
    setAttribute(name: string, value: string): void
    style: {
      position: string
      left: string
      top: string
      opacity: string
    }
    focus(): void
    select(): void
    setSelectionRange(start: number, end: number): void
  }
  execCommand(command: string): boolean
}

type FocusableElement = Element & {
  focus?: () => void
}

export function copyTextWithDocumentCommand(text: string, doc: ClipboardDocument): boolean {
  const textarea = doc.createElement('textarea')
  textarea.value = text
  textarea.setAttribute('readonly', '')
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  textarea.style.top = '0'
  textarea.style.opacity = '0'

  doc.body.appendChild(textarea)

  try {
    textarea.focus()
    textarea.select()
    textarea.setSelectionRange(0, text.length)
    return doc.execCommand('copy')
  } finally {
    doc.body.removeChild(textarea)
    ;(doc.activeElement as FocusableElement | null)?.focus?.()
  }
}
