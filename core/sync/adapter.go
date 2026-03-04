package sync

import (
	"context"

	"github.com/shinerio/skillflow/core/skill"
)

type ToolAdapter interface {
	Name() string
	SkillsDir() string
	Deploy(ctx context.Context, sk *skill.Skill) error
	Remove(ctx context.Context, sk *skill.Skill) error
	List(ctx context.Context) ([]string, error)
}
