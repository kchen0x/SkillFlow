# Memory Drawer And Status Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep memory editing in a right-side drawer with substantially more usable space and show pushed-agent status on module memory cards.

**Architecture:** Keep the existing `Memory.tsx` page as the integration point, but extract any reusable presentation calculations into a small frontend lib so the UI behavior can be tested without adding a new component-test stack. Reuse one status-rendering path for both main memory and module memory cards, and switch the drawer from layout flow to a fixed overlay anchored to the right edge.

**Tech Stack:** React 18, TypeScript, existing Node `--test` frontend test setup

---

### Task 1: Add failing tests for the new memory UI rules

**Files:**
- Create: `cmd/skillflow/frontend/tests/memoryUi.test.mjs`
- Create: `cmd/skillflow/frontend/tests/memoryPageUi.test.mjs`
- Test: `cmd/skillflow/frontend/src/lib/memoryUi.ts`
- Test: `cmd/skillflow/frontend/src/pages/Memory.tsx`

- [ ] **Step 1: Write the failing test for drawer width and shared status entries**

```js
test('getMemoryDrawerMetrics widens the editor while keeping a floor for narrow viewports', () => {
  assert.deepEqual(getMemoryDrawerMetrics(1600), { width: 960, maxWidth: 960, minWidth: 520 })
})

test('buildMemoryPushStatusEntries returns one status entry per enabled agent', () => {
  assert.deepEqual(buildMemoryPushStatusEntries([{ name: 'codex' }], { codex: 'synced' })[0], {
    agentType: 'codex',
    label: 'Codex',
    status: 'synced',
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `npm run test:unit -- memoryUi.test.mjs memoryPageUi.test.mjs`
Expected: FAIL because `memoryUi.ts` and the new `Memory.tsx` references do not exist yet.

- [ ] **Step 3: Add source assertions for the page integration**

```js
assert.match(source, /buildMemoryPushStatusEntries/)
assert.match(source, /position:\s*'fixed'/)
```

- [ ] **Step 4: Re-run tests to confirm the failures are still for missing implementation**

Run: `npm run test:unit -- memoryUi.test.mjs memoryPageUi.test.mjs`
Expected: FAIL on missing helper import or missing source matches.

- [ ] **Step 5: Commit**

```bash
git add cmd/skillflow/frontend/tests/memoryUi.test.mjs cmd/skillflow/frontend/tests/memoryPageUi.test.mjs
git commit -m "test: cover memory drawer and status ui rules"
```

### Task 2: Implement the widened fixed drawer and shared status display

**Files:**
- Create: `cmd/skillflow/frontend/src/lib/memoryUi.ts`
- Modify: `cmd/skillflow/frontend/src/pages/Memory.tsx`

- [ ] **Step 1: Write the minimal `memoryUi` helper implementation**

```ts
export function getMemoryDrawerMetrics(viewportWidth: number) {
  const maxWidth = 960
  const minWidth = 520
  return {
    width: Math.max(minWidth, Math.min(maxWidth, Math.round(viewportWidth * 0.72))),
    maxWidth,
    minWidth,
  }
}
```

- [ ] **Step 2: Add shared status-entry construction**

```ts
export function buildMemoryPushStatusEntries(agents, statuses) {
  return agents.map(agent => ({
    agentType: agent.name,
    label: getAgentLabel(agent.name),
    status: statuses[agent.name] ?? 'neverPushed',
  }))
}
```

- [ ] **Step 3: Update `Memory.tsx` to reuse the shared status entries on both card types**

```tsx
const memoryStatusEntries = buildMemoryPushStatusEntries(availableAgents, pushStatuses)
```

- [ ] **Step 4: Change the drawer container to a fixed right-side overlay and apply the wider metrics**

```tsx
style={{
  position: 'fixed',
  top: 0,
  right: 0,
  bottom: 0,
  width: drawerMetrics.width,
}}
```

- [ ] **Step 5: Commit**

```bash
git add cmd/skillflow/frontend/src/lib/memoryUi.ts cmd/skillflow/frontend/src/pages/Memory.tsx
git commit -m "feat: widen memory drawer and show module push status"
```

### Task 3: Verify the targeted frontend behavior

**Files:**
- Test: `cmd/skillflow/frontend/tests/memoryUi.test.mjs`
- Test: `cmd/skillflow/frontend/tests/memoryPageUi.test.mjs`
- Test: `cmd/skillflow/frontend/tests/memoryPageState.test.mjs`
- Test: `cmd/skillflow/frontend/tests/agentMemoryPreview.test.mjs`

- [ ] **Step 1: Run the targeted frontend unit tests**

Run: `npm run test:unit`
Expected: PASS for the existing suite plus the new memory UI tests.

- [ ] **Step 2: Review the resulting diff for unintended churn**

Run: `git diff -- cmd/skillflow/frontend/src/pages/Memory.tsx cmd/skillflow/frontend/src/lib/memoryUi.ts cmd/skillflow/frontend/tests/memoryUi.test.mjs cmd/skillflow/frontend/tests/memoryPageUi.test.mjs`
Expected: Only the drawer/status changes and supporting tests appear.

- [ ] **Step 3: Commit**

```bash
git add docs/superpowers/plans/2026-03-22-memory-drawer-and-status.md
git commit -m "docs: add memory drawer and status plan"
```
