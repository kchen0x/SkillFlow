package domain

import (
	"strings"
	"time"

	"github.com/shinerio/skillflow/core/skillkey"
)

type SourceType string

const (
	SourceGitHub SourceType = "github"
	SourceManual SourceType = "manual"
)

type InstalledSkill struct {
	ID            string
	Name          string
	Path          string
	Category      string
	Source        SourceType
	SourceURL     string
	SourceSubPath string
	SourceSHA     string
	LatestSHA     string
	InstalledAt   time.Time
	UpdatedAt     time.Time
	LastCheckedAt time.Time
}

func (s *InstalledSkill) IsGitHub() bool { return s.Source == SourceGitHub }

func (s *InstalledSkill) IsManual() bool { return s.Source == SourceManual }

func (s *InstalledSkill) HasUpdate() bool {
	return s.IsGitHub() && s.LatestSHA != "" && s.LatestSHA != s.SourceSHA
}

func LogicalKey(sk *InstalledSkill) (string, error) {
	if sk == nil {
		return "", nil
	}
	if sk.IsGitHub() {
		if logicalKey, err := skillkey.GitFromRepoURL(sk.SourceURL, sk.SourceSubPath); err == nil && strings.TrimSpace(logicalKey) != "" {
			return logicalKey, nil
		}
	}
	if strings.TrimSpace(sk.Path) == "" {
		return "", nil
	}
	return skillkey.ContentFromDir(sk.Path)
}
