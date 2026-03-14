package sync

import (
	"context"

	"github.com/shinerio/skillflow/core/skill"
)

type AgentAdapter interface {
	Name() string
	DefaultSkillsDir() string
	// Push copies skills into targetDir, flattened (no category subdirs)
	Push(ctx context.Context, skills []*skill.Skill, targetDir string) error
	// Pull scans sourceDir and returns skill candidates (not yet imported)
	Pull(ctx context.Context, sourceDir string) ([]*skill.Skill, error)
}
