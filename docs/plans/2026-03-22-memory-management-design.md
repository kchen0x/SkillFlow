# Memory Management Feature Design

## Overview

Add a "My Memory" feature to SkillFlow that lets users manage personal AI coding agent memories (main memory + module memories) in a unified interface, and push them to various AI coding tools (Claude Code, Codex, Gemini CLI, OpenCode, OpenClaw, and custom agents).

SkillFlow is the source of truth for memory content. Memories are stored in SkillFlow's data directory and pushed to agent directories.

## Core Concepts

### Data Ownership

- SkillFlow owns and manages all memory source files
- Memory content is pushed (one-way) to agent directories
- Agents never write back to SkillFlow's memory storage

### Memory Types

- **Main Memory**: A single `main.md` file containing global instructions shared across all agents
- **Module Memories**: Flat list of `.md` files under `rules/` directory, each containing a specific topic (e.g., `coding-style.md`, `testing-rules.md`)

### Push Modes (per Agent, user-selectable)

- **Merge Mode**: SkillFlow manages a `<!-- SkillFlow Managed Start/End -->` marker block in the agent's main memory file. Content outside the block is preserved. If the block is corrupted or missing, SkillFlow recreates it at the end of the file.
- **Takeover Mode**: SkillFlow fully owns the agent's main memory file and rules directory. Push overwrites the entire file.

## Data Model

### Storage Layout

**Synced data** (included in cloud backup):

```
~/.skillflow/
  memory/
    main.md                    # Main memory content (pure markdown)
    rules/                     # Module memories (pure markdown)
      coding-style.md
      testing-rules.md
      project-conventions.md
```

**Local-only data** (excluded from sync):

```
~/.skillflow/
  memory/
    memory_local.json
```

`memory_local.json` schema (all three sections — `pushConfigs`, `modules`, `pushState` — live in one file for simplicity; writes are serialized through `MemoryService` to prevent concurrent mutation):

```json
{
  "pushConfigs": {
    "claude-code": { "mode": "merge", "autoPush": true },
    "gemini-cli": { "mode": "takeover", "autoPush": false }
  },
  "modules": {
    "coding-style": { "pushTargets": ["claude-code", "codex", "gemini-cli"] },
    "testing-rules": { "pushTargets": ["claude-code"] }
  },
  "pushState": {
    "claude-code": {
      "lastPushedAt": "2026-03-21T10:00:00Z",
      "lastPushedHash": "abc123"
    },
    "gemini-cli": {
      "lastPushedAt": "2026-03-20T15:00:00Z",
      "lastPushedHash": "def456"
    }
  }
}
```

### Domain Entities

Domain entities contain only synced, business-meaningful data. Local-only configuration (push targets, auto-push, push state) is managed by the application layer as configuration DTOs, not as domain entity fields.

**MainMemory**
- `Content string` — Markdown content
- `UpdatedAt time.Time`

**ModuleMemory**
- `Name string` — Filename without `.md` suffix, also serves as ID
- `Content string` — Markdown content
- `UpdatedAt time.Time`

Name validation rules:
- Allowed characters: lowercase letters, digits, hyphens (`a-z0-9-`)
- Must start with a letter, must not end with a hyphen
- Max length: 64 characters
- Case-insensitive uniqueness (stored as lowercase)
- Must not collide with existing non-`sf-` prefixed files in any target agent's rules directory

### Application-Layer Configuration (local-only)

**MemoryPushConfig** (per Agent, stored in `memory_local.json`)
- `AgentType string`
- `Mode PushMode` — `merge` or `takeover`
- `AutoPush bool` — Whether this agent auto-pushes when memory changes

**ModulePushTargets** (per Module, stored in `memory_local.json`)
- `ModuleName string`
- `PushTargets []AgentType`

**MemoryPushState** (per Agent, stored in `memory_local.json`)
- `LastPushedAt time.Time`
- `LastPushedHash string` — Per-agent content hash at last push

### Content Hash Computation

The push state hash is **per-agent**: it is computed from the specific content that was pushed to that agent. This includes main memory content plus only the module memories targeted at that agent.

Algorithm: SHA-256 of concatenated content, with modules sorted by name:
```
hash = SHA256(mainMemory.Content + "\n" + sortedModules[0].Content + "\n" + sortedModules[1].Content + ...)
```

This ensures that changing a module not targeted at agent X does not mark agent X as `pendingPush`.

### Push Status Model

Each agent's memory push status:

| Status | Meaning |
|--------|---------|
| `synced` | Pushed, content matches current local memory |
| `pendingPush` | Local memory has changed since last push |
| `neverPushed` | Push targets configured but never pushed |

## Backend Architecture

### New Bounded Context: `memorycatalog`

Memories are a distinct business domain with their own entities (main memory, module memories), lifecycle (edit, preview, push, adapt), and persistence needs. Unlike skills and prompts which are pushed as files to agent push directories, memories target different agent paths (main memory files + rules directories) with per-agent format adaptation. This justifies a separate bounded context rather than extending `agentintegration` or `promptcatalog`.

```
core/memorycatalog/
  domain/
    main_memory.go          # Main memory entity (content + timestamp)
    module_memory.go         # Module memory entity (name, content, timestamp)
    push_mode.go             # merge / takeover enum (domain value object)
    errors.go
  app/
    service.go               # MemoryService — CRUD, file watch, save (flat service pattern, consistent with existing contexts)
    push_service.go          # PushService — push execution, adapter dispatch
    push_config.go           # MemoryPushConfig, ModulePushTargets, MemoryPushState DTOs (local-only config, not domain)
    port/
      repository/
        storage.go           # MemoryStorage interface (main + module persistence)
      gateway/
        agent_config.go      # AgentConfigGateway interface (get agent paths)
        agent_push.go        # AgentMemoryPusher interface (push to agent filesystem)
  infra/
    repository/
      fs_storage.go          # Filesystem persistence (read/write memory/ directory)
    gateway/
      agent_config_gateway.go  # Reads agent config from agentintegration
    editor/
      launcher.go            # System default editor launcher
    adapters/
      claude_adapter.go      # Claude Code — no rules index needed
      codex_adapter.go       # Codex — explicit rules listing
      gemini_adapter.go      # Gemini CLI — explicit rules listing
      opencode_adapter.go    # OpenCode — explicit rules listing
      openclaw_adapter.go    # OpenClaw — explicit rules listing
      custom_adapter.go      # Custom agent — explicit rules listing (default)
```

### Core Interfaces

```go
// app/port/gateway/agent_push.go
// AgentMemoryPusher is implemented per agent type in infra/adapters/.
// It handles the filesystem details of writing to agent directories.
type AgentMemoryPusher interface {
    // Push main memory to agent's main memory file
    PushMainMemory(content string, mode PushMode, agentMemoryPath string) error
    // Push module memory to agent's rules directory
    PushModuleMemory(module ModuleMemory, agentRulesDir string) error
    // Remove a pushed module memory from agent's rules directory
    RemoveModuleMemory(moduleName string, agentRulesDir string) error
    // Build rules index block (Claude Code returns empty, others return listing)
    BuildRulesIndex(modules []ModuleMemory, agentRulesDir string) RulesIndex
    // Detect and repair corrupted marker blocks
    RepairManagedBlock(agentMemoryPath string) error
}

// RulesIndex is a structured representation of the rules index block,
// assembled by PushService when composing the main memory push content.
type RulesIndex struct {
    Header  string   // e.g. "The following rule files are managed by SkillFlow:"
    Entries []string // absolute paths to sf-*.md files
}
```

```go
// app/port/gateway/agent_config.go
// AgentConfigGateway reads agent configuration from agentintegration context.
type AgentConfigGateway interface {
    // ListEnabledAgents returns all enabled agents with their memory paths.
    ListEnabledAgents() ([]AgentMemoryConfig, error)
}

// AgentMemoryConfig is the cross-context read DTO for agent memory paths.
// This requires extending agentintegration's AgentProfile with MemoryDir and RulesDir fields.
type AgentMemoryConfig struct {
    AgentType  string
    MemoryPath string // e.g. ~/.claude/CLAUDE.md
    RulesDir   string // e.g. ~/.claude/rules/
}
```

### Push Flow

```
User edits and saves → MemoryService.Save()
  ├─ Persist to ~/.skillflow/memory/
  ├─ If AutoPush enabled for any target agent:
  │    └─ PushService.AutoPush(changedModule)
  │         ├─ Query module.PushTargets
  │         ├─ For each target agent:
  │         │    ├─ AgentGateway gets agent path config
  │         │    ├─ Select corresponding AgentMemoryAdapter
  │         │    ├─ adapter.PushModuleMemory() → write sf-xxx.md to rules/
  │         │    ├─ adapter.BuildRulesIndex() → generate index
  │         │    └─ adapter.PushMainMemory() → update main memory file (with index)
  │         └─ Emit event to notify frontend
  └─ Return result
```

### Module Memory Add/Remove Side Effects

| Operation | Push Behavior |
|-----------|--------------|
| Add module + configure push targets | Write `sf-xxx.md` to rules dir + update main memory index block |
| Delete module | Delete `sf-xxx.md` from rules dir + update index |
| Remove agent from module push targets | Delete `sf-xxx.md` from that agent's rules dir + update that agent's index |
| Edit module content | Overwrite `sf-xxx.md` in rules dir (index unchanged) |

### Cross-Context Collaboration

| Collaborator | Method | Purpose |
|-------------|--------|---------|
| `agentintegration` | Read interface | Get agent config (type, enabled, MemoryDir, RulesDir). **Requires extending `AgentProfile` with `MemoryDir` and `RulesDir` fields** alongside existing `ScanDirs` and `PushDir`. |
| `orchestration` | Write coordination | Manual batch push through orchestration layer |
| `readmodel` | Read composition | My Memory page composes memory data + agent push status |
| `backup` | Backup scope | `memory/` directory included in cloud backup scope |

## Agent Adapter Details

### Agent Path Defaults

| Agent | Main Memory File | Rules Directory | Needs Index |
|-------|-----------------|-----------------|-------------|
| Claude Code | `~/.claude/CLAUDE.md` | `~/.claude/rules/` | No (auto-scan) |
| Codex | `~/.codex/AGENTS.md` | `~/.codex/rules/` | Yes |
| Gemini CLI | `~/.gemini/GEMINI.md` | `~/.gemini/rules/` | Yes |
| OpenCode | `~/.config/opencode/AGENTS.md` | `~/.config/opencode/rules/` | Yes |
| OpenClaw | `~/.openclaw/workspace/MEMORY.md` | `~/.openclaw/workspace/rules/` | Yes |
| Custom Agent | User configured | User configured | Yes (default) |

Built-in agents support user-overriding default paths in Settings → Agents tab.

### Module File Naming

Module memories are pushed with `sf-` prefix to distinguish from user-manually-placed files:
- `coding-style.md` → `sf-coding-style.md` in agent rules directory

### Merge Mode Push Example (Gemini CLI)

After push, `~/.gemini/GEMINI.md`:

```markdown
(user's existing content preserved)

<!-- SkillFlow Managed Start - DO NOT EDIT THIS BLOCK -->
## My Global Instructions

Always use TypeScript strict mode...

## SkillFlow Rules Index

The following rule files are managed by SkillFlow:
- /Users/username/.gemini/rules/sf-coding-style.md
- /Users/username/.gemini/rules/sf-testing-rules.md
<!-- SkillFlow Managed End -->
```

### Takeover Mode Push Example

Same agent, takeover mode — entire file is overwritten:

```markdown
## My Global Instructions

Always use TypeScript strict mode...

## SkillFlow Rules Index

The following rule files are managed by SkillFlow:
- /Users/username/.gemini/rules/sf-coding-style.md
- /Users/username/.gemini/rules/sf-testing-rules.md
```

### Marker Block Repair Strategy

Before each merge-mode push:
1. Search for `<!-- SkillFlow Managed Start` and `<!-- SkillFlow Managed End -->`
2. If only one half found or format corrupted → delete remnants, recreate complete block at end of file
3. If neither found → append new block at end of file
4. If user manually edited content between markers → overwrite with SkillFlow content (markers exist to prevent this)
5. Proceed with normal content write

### Edge Case Handling

- **Agent memory file does not exist**: Create the file. In merge mode, create with only the marker block. In takeover mode, create with full content.
- **File is read-only or locked**: Log error, mark push as failed, surface error to user in UI. Do not retry automatically.
- **File encoding**: Always read/write as UTF-8 without BOM. If a file contains BOM, strip it on read and write without BOM.
- **`sf-` name collision**: If an agent's rules directory already contains a non-SkillFlow-managed file matching `sf-<name>.md`, warn the user and skip that module for that agent rather than overwriting.

## Frontend Design

### Navigation

New sidebar entry between My Prompts and My Agents:

```
My Skills        (/)
My Prompts       (/prompts)
My Memory        (/memory)      ← new, Brain/BookOpen icon
My Agents        (/tools)
Push to Agents   (/sync/push)
Pull from Agents (/sync/pull)
Starred Repos    (/starred)
Cloud Backup     (/backup)
Settings         (/settings)
```

Sidebar entry shows red dot when any agent has `pendingPush` status.

### Page Layout: Unified Card Grid

Follows existing Dashboard/Prompts card-grid pattern:

- **Left panel**: Filter by Agent (All, Claude Code, Codex, Gemini CLI, etc.)
- **Toolbar**: Search input + sort (A→Z) + "New Module" button + "Push All" button
- **Main memory card**: Prominent card at top with left accent border, shows content preview + per-agent push status chips
- **Module cards**: Two-column grid, each card shows:
  - Module name + auto-push indicator
  - Content preview (truncated)
  - Push target agent chips at bottom
  - Orange dot for `pendingPush` status

### Edit Drawer

Right-side drawer slides out when clicking any memory card:

- **Width**: ~55% of page
- **Background**: Card grid dims but remains visible
- **Header**: Memory name + "Open in Editor" button + close button
- **Tab bar**: Edit (raw markdown textarea) / Preview (rendered markdown)
- **Content area**: Scrollable editing or preview
- **Footer** (module memories):
  - Push target agent chips (click to toggle)
  - Auto-push toggle
  - "Push Now" button
- **Footer** (main memory):
  - No push target selection (pushes to all configured agents)
  - Auto-push toggle
  - "Push Now" button

### Edit Capabilities

- **Internal editor**: Basic markdown textarea for quick edits
- **External editor**: "Open in Editor" launches system default text editor; SkillFlow watches for file changes and refreshes preview
- **Preview**: Rendered markdown view

### Status Indicators

- **Green chip** (`✓ Agent`): synced — last push matches current content
- **Orange chip** (`⚠ Agent`): pendingPush — content has changed since last push
- **Gray chip** (`Agent`): not configured as push target for this module
- **Red dot on card**: at least one target agent has pendingPush
- **Red dot on sidebar**: at least one agent across all memories has pendingPush

## Settings Integration

### Settings → Agents Tab Extension

Each agent config (built-in and custom) gains:

- **Memory File**: Path to agent's main memory file (shows default, user-overridable)
- **Rules Dir**: Path to agent's rules directory (shows default, user-overridable)
- **Memory Push Mode**: Dropdown — `Merge` / `Takeover`
- **Auto Push Memory**: Toggle

These settings are stored in `memory_local.json`, not synced.

## Cloud Backup & Sync Integration

### Backup Scope

- `memory/main.md` + `memory/rules/*.md` → included in backup
- `memory/memory_local.json` → excluded from backup

### Restore / New Machine Flow

1. Memory content files restored to `~/.skillflow/memory/`
2. `memory_local.json` does not exist → all agents have empty pushState → all marked `neverPushed`
3. User configures push mode and auto-push per agent in Settings
4. User executes "Push All" or pushes individually

### Cross-Machine Sync Detection (on startup)

1. Compute content hash of `main.md` + all module files
2. Compare against each agent's `lastPushedHash` in `memory_local.json`
3. For agents with hash mismatch:
   - `autoPush: true` → auto-push, update pushState
   - `autoPush: false` → mark `pendingPush`, show red dot in frontend
4. Emit event to notify frontend of status changes

## Implementation Notes

### App Layer Pattern

The `app/` layer uses flat service files (`service.go`, `push_service.go`) without `command/`/`query/` sub-packages. This follows the existing convention in `skillcatalog`, `promptcatalog`, and `agentintegration`, which also use flat service files despite the `layers.md` documentation suggesting sub-packages.

### AgentType Identity Mapping

`AgentType string` used throughout `memorycatalog` maps to `AgentProfile.Name` in `agentintegration`. They are the same string value (e.g., `"claude-code"`, `"gemini-cli"`).

### Module Deletion Cleanup

When a module memory file is deleted:
1. Remove `sf-<name>.md` from all target agents' rules directories
2. Update rules index in each affected agent's main memory file
3. Remove the module's entry from `modules` in `memory_local.json`
4. Recompute and update `pushState` hashes for affected agents

### Wails Event Types

| Event | Payload | Trigger |
|-------|---------|---------|
| `memory:content:changed` | `{ type: "main" \| "module", name?: string }` | Memory file saved (internal or external editor) |
| `memory:push:completed` | `{ agent: string, success: bool, error?: string }` | Push to agent completed |
| `memory:status:changed` | `{ agent: string, status: "synced" \| "pendingPush" \| "neverPushed" }` | Push status changed (after sync detection or push) |
