package sync

import (
	"context"

	"github.com/shinerio/skillflow/core/skillcatalog/domain"
)

type AgentAdapter interface {
	Name() string
	DefaultSkillsDir() string
	// Push copies skills into targetDir, flattened (no category subdirs)
	Push(ctx context.Context, skills []*domain.InstalledSkill, targetDir string) error
	// Pull scans sourceDir and returns skill candidates (not yet imported)
	Pull(ctx context.Context, sourceDir string) ([]*domain.InstalledSkill, error)
}
