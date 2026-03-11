export const BACKGROUND_MEMORY_TRIM_DELAY_MS = 30000

export type AppActivityState = {
  windowVisible: boolean
  documentVisible: boolean
  focused: boolean
  foreground: boolean
  trimScheduled: boolean
  trimmed: boolean
  resumeToken: number
}

export type AppActivityEvent =
  | { type: 'window_visibility_changed'; visible: boolean }
  | { type: 'document_visibility_changed'; visible: boolean }
  | { type: 'focus_changed'; focused: boolean }
  | { type: 'trim_timeout_elapsed' }

function deriveForeground(windowVisible: boolean, documentVisible: boolean, focused: boolean) {
  return windowVisible && documentVisible && focused
}

export function createAppActivityState(overrides: Partial<Pick<AppActivityState, 'windowVisible' | 'documentVisible' | 'focused'>> = {}): AppActivityState {
  const windowVisible = overrides.windowVisible ?? true
  const documentVisible = overrides.documentVisible ?? true
  const focused = overrides.focused ?? true
  const foreground = deriveForeground(windowVisible, documentVisible, focused)

  return {
    windowVisible,
    documentVisible,
    focused,
    foreground,
    trimScheduled: !foreground,
    trimmed: false,
    resumeToken: 0,
  }
}

export function reduceAppActivityState(state: AppActivityState, event: AppActivityEvent): AppActivityState {
  let windowVisible = state.windowVisible
  let documentVisible = state.documentVisible
  let focused = state.focused
  let trimmed = state.trimmed
  let resumeToken = state.resumeToken

  switch (event.type) {
    case 'window_visibility_changed':
      windowVisible = event.visible
      if (event.visible) {
        // Desktop window restore events are more reliable than WebView focus/document visibility.
        documentVisible = true
        focused = true
      }
      break
    case 'document_visibility_changed':
      documentVisible = event.visible
      if (event.visible) windowVisible = true
      break
    case 'focus_changed':
      focused = event.focused
      if (event.focused) windowVisible = true
      break
    case 'trim_timeout_elapsed':
      if (!state.foreground) {
        trimmed = true
      }
      break
  }

  const foreground = deriveForeground(windowVisible, documentVisible, focused)
  if (foreground && trimmed) {
    trimmed = false
    resumeToken += 1
  }

  return {
    windowVisible,
    documentVisible,
    focused,
    foreground,
    trimScheduled: !foreground && !trimmed,
    trimmed,
    resumeToken,
  }
}
