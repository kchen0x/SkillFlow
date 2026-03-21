package gateway

import (
	"context"

	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
)

type AgentGateway interface {
	Name() string
	DefaultSkillsDir() string
	Push(ctx context.Context, skills []*skilldomain.InstalledSkill, targetDir string) error
	Pull(ctx context.Context, sourceDir string) ([]*skilldomain.InstalledSkill, error)
}

type MaxDepthPuller interface {
	PullWithMaxDepth(ctx context.Context, sourceDir string, maxDepth int) ([]*skilldomain.InstalledSkill, error)
}
