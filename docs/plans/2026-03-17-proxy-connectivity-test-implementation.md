# Proxy Connectivity Test Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a Settings → Network proxy connectivity test that uses current unsaved proxy form values, defaults to `https://github.com`, and fails fast after 5 seconds.

**Architecture:** Add a small backend Wails method that accepts a target URL plus an explicit `config.ProxyConfig`, builds an HTTP client with the same proxy semantics as the app, and returns structured connectivity results. Wire the Settings network page to call that method with the current in-memory proxy form state and render inline feedback. Update detailed feature docs and strengthen `AGENTS.md` so small feature changes do not churn README files.

**Tech Stack:** Go, Wails, React, TypeScript, Markdown docs

---

### Task 1: Lock backend proxy test semantics in Go tests

**Files:**
- Create: `cmd/skillflow/app_proxy_test.go`
- Modify: `cmd/skillflow/app.go` or `cmd/skillflow/app_proxy.go`

**Step 1: Write the failing tests**

Add Go tests that prove:
- empty target URL falls back to `https://github.com`
- the explicit proxy config parameter overrides persisted proxy config
- a received non-2xx HTTP response still counts as a successful connectivity test
- the request stops after the 5-second timeout path

Suggested test names:
- `TestTestProxyConnectionDefaultsTargetURL`
- `TestTestProxyConnectionUsesProvidedProxyConfig`
- `TestTestProxyConnectionTreatsHTTPResponseAsSuccess`
- `TestTestProxyConnectionTimesOut`

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/skillflow -run 'TestTestProxyConnectionDefaultsTargetURL|TestTestProxyConnectionUsesProvidedProxyConfig|TestTestProxyConnectionTreatsHTTPResponseAsSuccess|TestTestProxyConnectionTimesOut'`

Expected: `FAIL` because the Wails-facing proxy test method does not exist yet.

**Step 3: Write minimal implementation**

Implement:
- a proxy test result model
- a helper that builds an HTTP client from an explicit `config.ProxyConfig`
- a 5-second timeout request path
- stable logging for start/completion/failure

Keep the method read-only. Do not persist any config changes.

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/skillflow -run 'TestTestProxyConnectionDefaultsTargetURL|TestTestProxyConnectionUsesProvidedProxyConfig|TestTestProxyConnectionTreatsHTTPResponseAsSuccess|TestTestProxyConnectionTimesOut'`

Expected: `ok`

**Step 5: Commit**

Run:

```bash
git add cmd/skillflow/app.go cmd/skillflow/app_proxy.go cmd/skillflow/app_proxy_test.go
git commit -m "feat: add proxy connectivity test backend"
```

### Task 2: Wire Wails bindings for the new backend method

**Files:**
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.js`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`
- Modify: `cmd/skillflow/frontend/wailsjs/go/models.ts`

**Step 1: Write the failing verification**

After the backend method is added, verify the frontend bindings do not expose it yet.

Run: `rg -n "TestProxyConnection|ProxyConnectionTestResult" cmd/skillflow/frontend/wailsjs/go`

Expected: no matches before generation/manual update.

**Step 2: Regenerate or update bindings**

Run Wails generation if available. If generation is unavailable, update the generated JS/TS binding files manually so they match the new exported method and result model.

**Step 3: Verify bindings**

Run: `rg -n "TestProxyConnection|ProxyConnectionTestResult" cmd/skillflow/frontend/wailsjs/go`

Expected: matches appear in the generated binding files.

**Step 4: Commit**

Run:

```bash
git add cmd/skillflow/frontend/wailsjs/go/main/App.js cmd/skillflow/frontend/wailsjs/go/main/App.d.ts cmd/skillflow/frontend/wailsjs/go/models.ts
git commit -m "chore: expose proxy connectivity bindings"
```

### Task 3: Add Settings network UI for proxy connectivity testing

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/Settings.tsx`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`

**Step 1: Write the failing frontend tests**

Add a small pure helper or reducer in the settings page area if needed, and test that:
- blank target URL falls back to `https://github.com`
- current form proxy state is the payload sent to the backend
- success and failure result text formatting is stable

Suggested test names:
- `normalizeProxyTestTarget defaults github url`
- `buildProxyTestPayload uses current proxy form state`

**Step 2: Run test to verify it fails**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: `FAIL` because the helper and UI wiring do not exist yet.

**Step 3: Write minimal implementation**

Implement:
- a local `proxyTestURL` state initialized to `https://github.com`
- a test button and loading state
- inline success/failure result display
- a backend call that always uses the current `cfg.proxy`
- localized copy for labels, helper text, and result text

Do not auto-save settings as part of the test action.

**Step 4: Run test to verify it passes**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: `PASS`

**Step 5: Commit**

Run:

```bash
git add cmd/skillflow/frontend/src/pages/Settings.tsx cmd/skillflow/frontend/src/i18n/en.ts cmd/skillflow/frontend/src/i18n/zh.ts
git commit -m "feat: add proxy connectivity test ui"
```

### Task 4: Update feature docs and reinforce README-churn guidance

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `AGENTS.md`

**Step 1: Write the failing verification**

Verify the current docs and repo instructions do not mention the new proxy connectivity test or the stronger README exception wording.

Run:

```bash
rg -n "proxy connectivity|测试连接|small feature|small UX|README churn" docs/features.md docs/features_zh.md AGENTS.md
```

Expected: no relevant matches.

**Step 2: Write minimal doc updates**

Update:
- the Settings / Network section in `docs/features.md`
- the corresponding Chinese section in `docs/features_zh.md`
- `AGENTS.md` so README updates are explicitly limited to high-level capability changes, and small feature additions should not touch README

Do not update `README.md` or `README_zh.md`.

**Step 3: Verify docs**

Run:

```bash
rg -n "proxy connectivity|测试连接|small feature|small UX|README" docs/features.md docs/features_zh.md AGENTS.md
```

Expected: the new feature and wording appear only in the intended files.

**Step 4: Commit**

Run:

```bash
git add docs/features.md docs/features_zh.md AGENTS.md
git commit -m "docs: add proxy test reference and tighten readme guidance"
```

### Task 5: Final verification

**Files:**
- Verify only

**Step 1: Run backend tests**

Run: `go test ./cmd/skillflow`

Expected: `ok`

**Step 2: Run frontend unit tests**

Run: `cd cmd/skillflow/frontend && npm run test:unit`

Expected: `PASS`

**Step 3: Build-check Wails bindings compile path**

Run: `cd cmd/skillflow/frontend && npm run build`

Expected: successful TypeScript + Vite production build

**Step 4: Review diff**

Run: `git diff --stat`

Expected: only the planned backend, frontend, binding, doc, and instruction files changed.
