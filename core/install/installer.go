package install

import "context"

type InstallSource struct {
	Type string // "github" | "local"
	URI  string
}

type SkillCandidate struct {
	Name       string
	Path       string // relative path within source
	LogicalKey string
	Installed  bool
	Updatable  bool
}

type Installer interface {
	Type() string
	Scan(ctx context.Context, source InstallSource) ([]SkillCandidate, error)
	Install(ctx context.Context, source InstallSource, selected []SkillCandidate, category string) error
}
