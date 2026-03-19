# Proxy Connectivity Test Design

**Date:** 2026-03-17

## Context

The Settings → Network page currently lets users choose proxy mode and enter a manual proxy URL, but there is no direct way to verify whether the current proxy settings can actually reach a target site. Users have to infer success indirectly from later GitHub update checks or repo operations, which is slow and ambiguous when a proxy is misconfigured.

The requested behavior is:

- add a dedicated proxy connectivity test action in the Network settings page
- default the test target to `https://github.com`
- let the user enter any custom target URL
- enforce a 5-second timeout
- prefer the current in-memory form values even when they have not been saved yet
- if the user has not changed anything, the current form state should naturally match the saved config

The user also clarified that this is a small feature change and should not cause README churn. Only the detailed feature reference should change, and the repo instructions should make that rule more explicit.

## Decision

1. Add a backend Wails method dedicated to proxy connectivity testing instead of trying to reuse app update checks or frontend `fetch`.
2. Keep the test request separate from persisted config writes. The frontend passes the current `cfg.proxy` form state directly to the backend, so unsaved edits are honored without mutating `config.json` or `config_local.json`.
3. Add a small inline “Test Connection” section to Settings → Network with:
   - a target URL input prefilled with `https://github.com`
   - a test button
   - inline result feedback for success/failure
4. Use a 5-second context timeout for the test request.
5. Treat “received any HTTP response” as a successful connectivity test, even if the response code is `301`, `401`, or `403`. Fail only on transport-level problems such as timeout, DNS resolution failure, TLS failure, or proxy connection failure.
6. Update `docs/features.md` and `docs/features_zh.md`, but do not update `README.md` or `README_zh.md` for this small UX addition.
7. Strengthen `AGENTS.md` so the README rule more explicitly says that small feature additions should not trigger README edits unless the high-level product capability summary truly changes.

## UX Flow

1. User opens Settings → Network.
2. User sees the existing proxy mode controls and proxy URL field.
3. Below that, user sees a new proxy test area with:
   - target URL input
   - `Test Connection` action
   - short helper copy explaining the default GitHub target and 5-second timeout
4. Clicking the button sends the current page state’s proxy config plus the target URL to the backend.
5. While running, the button shows a loading state and disables repeat clicks.
6. On success, the page shows the target URL, HTTP status, and elapsed time.
7. On failure, the page shows the target URL, elapsed time when available, and the concrete error message.

## Backend API Shape

- Add a new exported Wails app method, for example:
  - `TestProxyConnection(targetURL string, proxy config.ProxyConfig) (*ProxyConnectionTestResult, error)`
- Add a result model containing:
  - normalized target URL
  - success flag
  - HTTP status code
  - elapsed milliseconds
  - human-readable message

The backend method should:

1. normalize and validate the target URL
2. build an HTTP client using the supplied proxy config instead of persisted config when provided
3. issue a simple GET request with a 5-second timeout
4. close the response body immediately after headers arrive
5. log started/completed/failed with target URL, proxy mode, status, and latency

## Error Handling

- Empty target URL falls back to `https://github.com`.
- Invalid target URLs return a validation error before any network call.
- Only `http` and `https` targets are accepted.
- Transport errors return a failure result with the underlying message.
- HTTP response codes do not count as transport failures.

## Scope

- `cmd/skillflow/app.go` or a new flat `cmd/skillflow/app_proxy.go` backend method/model
- `cmd/skillflow/frontend/src/pages/Settings.tsx`
- i18n strings in English and Chinese
- Wails bindings for the new app method
- `docs/features.md`
- `docs/features_zh.md`
- `AGENTS.md`

## Out Of Scope

- saving proxy test history
- automatically rewriting invalid target URLs
- changing README wording for this small feature
- adding proxy authentication UI
