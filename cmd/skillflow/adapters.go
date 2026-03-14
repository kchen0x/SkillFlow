package main

import (
	"github.com/shinerio/skillflow/core/config"
	"github.com/shinerio/skillflow/core/registry"
	agentsync "github.com/shinerio/skillflow/core/sync"
)

func registerAdapters() {
	agents := []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}
	for _, name := range agents {
		registry.RegisterAdapter(agentsync.NewFilesystemAdapter(name, config.DefaultAgentPushDir(name)))
	}
}
