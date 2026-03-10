import crypto from 'node:crypto'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const appDir = path.resolve(__dirname, '..')
const repoRoot = path.resolve(appDir, '..', '..')
const stateDir = path.join(appDir, '.build-cache')
const bindingsStateFile = path.join(stateDir, 'bindings.hash')
const frontendStateFile = path.join(stateDir, 'frontend.hash')

const bindingsOutputs = [
  path.join(appDir, 'frontend/wailsjs/go/main/App.js'),
  path.join(appDir, 'frontend/wailsjs/go/main/App.d.ts'),
  path.join(appDir, 'frontend/wailsjs/go/models.ts'),
]

const frontendOutputs = [
  path.join(appDir, 'frontend/dist/index.html'),
]

const bindingsInputs = [
  { kind: 'file', path: path.join(repoRoot, 'go.mod') },
  { kind: 'file', path: path.join(repoRoot, 'go.sum') },
  { kind: 'dir', path: path.join(repoRoot, 'cmd/skillflow'), extensions: new Set(['.go']) },
  { kind: 'dir', path: path.join(repoRoot, 'core'), extensions: new Set(['.go']) },
]

const frontendInputs = [
  { kind: 'file', path: path.join(appDir, 'frontend/index.html') },
  { kind: 'file', path: path.join(appDir, 'frontend/package.json') },
  { kind: 'file', path: path.join(appDir, 'frontend/package-lock.json') },
  { kind: 'file', path: path.join(appDir, 'frontend/postcss.config.js') },
  { kind: 'file', path: path.join(appDir, 'frontend/tailwind.config.js') },
  { kind: 'file', path: path.join(appDir, 'frontend/tsconfig.json') },
  { kind: 'file', path: path.join(appDir, 'frontend/tsconfig.node.json') },
  { kind: 'file', path: path.join(appDir, 'frontend/vite.config.ts') },
  { kind: 'dir', path: path.join(appDir, 'frontend/src') },
  { kind: 'dir', path: path.join(appDir, 'frontend/wailsjs') },
]

export function computeBuildPlan({
  bindingsChanged,
  frontendChanged,
  bindingsOutputsReady,
  frontendOutputsReady,
}) {
  if (bindingsChanged || !bindingsOutputsReady) {
    return {
      skipBindings: false,
      skipFrontend: false,
    }
  }

  return {
    skipBindings: true,
    skipFrontend: !frontendChanged && frontendOutputsReady,
  }
}

function walkFiles(rootDir, extensions) {
  if (!fs.existsSync(rootDir)) {
    return []
  }

  const files = []

  function visit(currentDir) {
    const entries = fs.readdirSync(currentDir, { withFileTypes: true })
      .sort((a, b) => a.name.localeCompare(b.name))

    for (const entry of entries) {
      const fullPath = path.join(currentDir, entry.name)
      if (entry.isDirectory()) {
        visit(fullPath)
        continue
      }
      if (!entry.isFile()) {
        continue
      }
      if (extensions && !extensions.has(path.extname(entry.name))) {
        continue
      }
      files.push(fullPath)
    }
  }

  visit(rootDir)
  return files
}

function expandInputs(inputs) {
  const files = []
  for (const input of inputs) {
    if (input.kind === 'file') {
      if (fs.existsSync(input.path)) {
        files.push(input.path)
      }
      continue
    }
    files.push(...walkFiles(input.path, input.extensions))
  }
  return files.sort()
}

function hashInputs(inputs) {
  const hash = crypto.createHash('sha256')
  for (const filePath of expandInputs(inputs)) {
    hash.update(path.relative(repoRoot, filePath))
    hash.update('\n')
    hash.update(fs.readFileSync(filePath))
    hash.update('\n')
  }
  return hash.digest('hex')
}

function readState(stateFile) {
  if (!fs.existsSync(stateFile)) {
    return null
  }
  return fs.readFileSync(stateFile, 'utf8').trim()
}

function ensureStateDir() {
  fs.mkdirSync(stateDir, { recursive: true })
}

function writeState(stateFile, value) {
  ensureStateDir()
  fs.writeFileSync(stateFile, `${value}\n`)
}

function outputsReady(paths) {
  return paths.every(filePath => fs.existsSync(filePath))
}

function currentState() {
  const bindingsHash = hashInputs(bindingsInputs)
  const frontendHash = hashInputs(frontendInputs)

  return {
    bindingsHash,
    frontendHash,
    bindingsOutputsReady: outputsReady(bindingsOutputs),
    frontendOutputsReady: outputsReady(frontendOutputs),
  }
}

function planFlags() {
  const state = currentState()
  const plan = computeBuildPlan({
    bindingsChanged: readState(bindingsStateFile) !== state.bindingsHash,
    frontendChanged: readState(frontendStateFile) !== state.frontendHash,
    bindingsOutputsReady: state.bindingsOutputsReady,
    frontendOutputsReady: state.frontendOutputsReady,
  })

  const flags = []
  if (plan.skipBindings) {
    flags.push('-skipbindings')
  }
  if (plan.skipFrontend) {
    flags.push('-s')
  }
  process.stdout.write(flags.join(' '))
}

function markBindings() {
  const { bindingsHash } = currentState()
  writeState(bindingsStateFile, bindingsHash)
}

function markAll() {
  const { bindingsHash, frontendHash } = currentState()
  writeState(bindingsStateFile, bindingsHash)
  writeState(frontendStateFile, frontendHash)
}

const command = process.argv[2]

if (command === 'plan') {
  planFlags()
} else if (command === 'mark-bindings') {
  markBindings()
} else if (command === 'mark') {
  markAll()
} else if (command) {
  process.stderr.write(`unknown command: ${command}\n`)
  process.exitCode = 1
}
