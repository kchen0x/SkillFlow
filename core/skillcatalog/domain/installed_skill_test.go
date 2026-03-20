package domain_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/skillcatalog/domain"
	"github.com/stretchr/testify/assert"
)

func TestInstalledSkillSourceTypes(t *testing.T) {
	s := domain.InstalledSkill{
		ID:       "test-id",
		Name:     "my-skill",
		Source:   domain.SourceGitHub,
		Category: "coding",
	}
	assert.Equal(t, domain.SourceType("github"), s.Source)
	assert.True(t, s.IsGitHub())
	assert.False(t, s.IsManual())
}

func TestInstalledSkillIsManual(t *testing.T) {
	s := domain.InstalledSkill{Source: domain.SourceManual}
	assert.True(t, s.IsManual())
	assert.False(t, s.IsGitHub())
}

func TestInstalledSkillHasUpdate(t *testing.T) {
	s := domain.InstalledSkill{
		Source:    domain.SourceGitHub,
		SourceSHA: "abc123",
		LatestSHA: "def456",
	}
	assert.True(t, s.HasUpdate())

	s.LatestSHA = "abc123"
	assert.False(t, s.HasUpdate())
}
