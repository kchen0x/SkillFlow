import { ClipboardSetText } from '../../wailsjs/runtime/runtime'

export async function copyTextToClipboard(text: string): Promise<void> {
  try {
    const ok = await ClipboardSetText(text)
    if (ok) {
      return
    }
  } catch {
    // Fall back to the browser clipboard when the runtime API is unavailable.
  }

  if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text)
    return
  }

  throw new Error('clipboard unavailable')
}
