import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const repoRoot = path.resolve(__dirname, '..', '..', '..')

test('Makefile packaged build uses go.mod maintenance skip flags', () => {
  const makefile = fs.readFileSync(path.join(repoRoot, 'Makefile'), 'utf8')

  assert.match(makefile, /WAILS_BUILD_FLAGS = .*?-m -nosyncgomod/)
})

test('frontend tsconfig enables incremental type-check cache', () => {
  const tsconfig = JSON.parse(
    fs.readFileSync(path.join(repoRoot, 'cmd/skillflow/frontend/tsconfig.json'), 'utf8'),
  )

  assert.equal(tsconfig.compilerOptions.incremental, true)
  assert.equal(
    tsconfig.compilerOptions.tsBuildInfoFile,
    '.cache/tsconfig.app.tsbuildinfo',
  )
})

test('frontend node tsconfig enables incremental cache for vite config checks', () => {
  const tsconfigNode = JSON.parse(
    fs.readFileSync(path.join(repoRoot, 'cmd/skillflow/frontend/tsconfig.node.json'), 'utf8'),
  )

  assert.equal(tsconfigNode.compilerOptions.incremental, true)
  assert.equal(
    tsconfigNode.compilerOptions.tsBuildInfoFile,
    '.cache/tsconfig.node.tsbuildinfo',
  )
})
