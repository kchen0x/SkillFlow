package install

import (
	"context"
	"path/filepath"

	"github.com/shinerio/skillflow/core/skill"
)

type LocalInstaller struct {
	validator *skill.Validator
}

func NewLocalInstaller() *LocalInstaller {
	return &LocalInstaller{validator: skill.NewValidator()}
}

func (l *LocalInstaller) Type() string { return "local" }

func (l *LocalInstaller) Scan(_ context.Context, source InstallSource) ([]SkillCandidate, error) {
	dir := source.URI
	if err := l.validator.Validate(dir); err != nil {
		return nil, nil // not a valid skill dir — return empty, not error
	}
	return []SkillCandidate{{Name: filepath.Base(dir), Path: dir}}, nil
}

func (l *LocalInstaller) Install(_ context.Context, _ InstallSource, _ []SkillCandidate, _ string) error {
	// Local install: the app layer copies from candidate.Path directly via Storage.Import
	return nil
}
