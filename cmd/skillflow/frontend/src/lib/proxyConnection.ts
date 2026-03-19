export const DEFAULT_PROXY_TEST_URL = 'https://github.com'

export type ProxyConnectionProxyConfig = {
  mode?: string
  url?: string
}

export type ProxyConnectionPayload = {
  targetURL: string
  proxy: {
    mode: string
    url: string
  }
}

export function normalizeProxyTestTargetURL(value?: string | null) {
  const trimmed = (value ?? '').trim()
  return trimmed || DEFAULT_PROXY_TEST_URL
}

export function buildProxyConnectionPayload(
  targetURL: string | null | undefined,
  proxy: ProxyConnectionProxyConfig | null | undefined,
): ProxyConnectionPayload {
  return {
    targetURL: normalizeProxyTestTargetURL(targetURL),
    proxy: {
      mode: typeof proxy?.mode === 'string' && proxy.mode.trim() ? proxy.mode.trim() : 'none',
      url: typeof proxy?.url === 'string' ? proxy.url.trim() : '',
    },
  }
}
