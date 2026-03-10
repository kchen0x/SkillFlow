import test from 'node:test'
import assert from 'node:assert/strict'

const { computeBuildPlan } = await import('./build-plan.mjs')

test('computeBuildPlan skips bindings and frontend when inputs and outputs are unchanged', () => {
  const plan = computeBuildPlan({
    bindingsChanged: false,
    frontendChanged: false,
    bindingsOutputsReady: true,
    frontendOutputsReady: true,
  })

  assert.deepEqual(plan, {
    skipBindings: true,
    skipFrontend: true,
  })
})

test('computeBuildPlan reruns both bindings and frontend when bindings changed', () => {
  const plan = computeBuildPlan({
    bindingsChanged: true,
    frontendChanged: false,
    bindingsOutputsReady: true,
    frontendOutputsReady: true,
  })

  assert.deepEqual(plan, {
    skipBindings: false,
    skipFrontend: false,
  })
})

test('computeBuildPlan reruns frontend but skips bindings when only frontend inputs changed', () => {
  const plan = computeBuildPlan({
    bindingsChanged: false,
    frontendChanged: true,
    bindingsOutputsReady: true,
    frontendOutputsReady: true,
  })

  assert.deepEqual(plan, {
    skipBindings: true,
    skipFrontend: false,
  })
})

test('computeBuildPlan reruns bindings when generated bindings are missing', () => {
  const plan = computeBuildPlan({
    bindingsChanged: false,
    frontendChanged: false,
    bindingsOutputsReady: false,
    frontendOutputsReady: true,
  })

  assert.deepEqual(plan, {
    skipBindings: false,
    skipFrontend: false,
  })
})

test('computeBuildPlan reruns frontend when frontend dist is missing', () => {
  const plan = computeBuildPlan({
    bindingsChanged: false,
    frontendChanged: false,
    bindingsOutputsReady: true,
    frontendOutputsReady: false,
  })

  assert.deepEqual(plan, {
    skipBindings: true,
    skipFrontend: false,
  })
})
