package domain

import "time"

// MemoryPushConfig is per-agent push configuration (local-only, stored in memory_local.json).
type MemoryPushConfig struct {
	AgentType string
	Mode      PushMode
	AutoPush  bool
}

// ModulePushTargets is per-module push target configuration (local-only).
type ModulePushTargets struct {
	ModuleName  string
	PushTargets []string // AgentType values
}

// MemoryPushState tracks the last push to an agent (local-only).
type MemoryPushState struct {
	LastPushedAt   time.Time
	LastPushedHash string // per-agent content hash at last push
}
