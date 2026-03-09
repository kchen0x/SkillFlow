import { ClipboardSetText } from '../../wailsjs/runtime/runtime'
import { copyTextWithDocumentCommand, copyTextWithFallbacks } from './clipboardCore'

export async function copyTextToClipboard(text: string): Promise<void> {
  await copyTextWithFallbacks(text, {
    runtimeWriteText: async (value) => ClipboardSetText(value),
    browserWriteText: typeof navigator !== 'undefined' && navigator.clipboard?.writeText
      ? (value) => navigator.clipboard.writeText(value)
      : undefined,
    execCommandCopy: typeof document !== 'undefined'
      ? (value) => copyTextWithDocumentCommand(value, document)
      : undefined,
  })
}
