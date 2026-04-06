export type BackendClientConfig = {
  baseUrl: string
  token: string
}

export class BackendClientError extends Error {
  status: number

  constructor(message: string, status = 500) {
    super(message)
    this.name = 'BackendClientError'
    this.status = status
  }
}

type GetConfigFn = () => Promise<BackendClientConfig>
type FetchFn = typeof fetch

type CreateBackendClientOptions = {
  getConfig?: GetConfigFn
  fetchImpl?: FetchFn
}

function normalizeClientConfig(input: BackendClientConfig): BackendClientConfig {
  const baseUrl = typeof input?.baseUrl === 'string' ? input.baseUrl.trim().replace(/\/+$/, '') : ''
  const token = typeof input?.token === 'string' ? input.token.trim() : ''
  if (!baseUrl || !token) {
    throw new BackendClientError('backend client config unavailable')
  }
  return { baseUrl, token }
}

function defaultGetConfig(): Promise<BackendClientConfig> {
  return (window as any).go.main.App.GetBackendClientConfig()
}

export function createBackendClient(options: CreateBackendClientOptions = {}) {
  const getConfig = options.getConfig ?? defaultGetConfig
  const fetchImpl = options.fetchImpl ?? fetch
  let configPromise: Promise<BackendClientConfig> | null = null

  async function loadConfig(): Promise<BackendClientConfig> {
    if (!configPromise) {
      configPromise = Promise.resolve(getConfig()).then(normalizeClientConfig)
    }
    return configPromise
  }

  async function invoke<T>(method: string, ...params: unknown[]): Promise<T> {
    const config = await loadConfig()
    const response = await fetchImpl(`${config.baseUrl}/invoke`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-SkillFlow-Token': config.token,
      },
      body: JSON.stringify({
        method,
        params,
      }),
    })

    const rawText = await response.text()
    let payload: { ok?: boolean; result?: T; error?: string } | null = null
    if (rawText.trim()) {
      try {
        payload = JSON.parse(rawText)
      } catch {
        throw new BackendClientError(rawText, response.status)
      }
    }

    if (!response.ok) {
      throw new BackendClientError(payload?.error?.trim() || `backend status ${response.status}`, response.status)
    }
    if (!payload?.ok) {
      throw new BackendClientError(payload?.error?.trim() || 'backend call failed', response.status)
    }
    return payload.result as T
  }

  function reset() {
    configPromise = null
  }

  return {
    invoke,
    reset,
  }
}

export const backendClient = createBackendClient()
