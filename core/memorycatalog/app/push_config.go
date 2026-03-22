package app

import "github.com/shinerio/skillflow/core/memorycatalog/domain"

// MemoryPushConfig is per-agent push configuration (local-only).
// Defined in domain to avoid import cycles with sub-packages.
type MemoryPushConfig = domain.MemoryPushConfig

// ModulePushTargets is per-module push target configuration (local-only).
type ModulePushTargets = domain.ModulePushTargets

// MemoryPushState tracks the last push to an agent (local-only).
type MemoryPushState = domain.MemoryPushState
