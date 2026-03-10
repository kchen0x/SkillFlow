package main

import (
	"testing"

	"github.com/shinerio/skillflow/core/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveToolSkillCandidatesIncludesInstalledSource(t *testing.T) {
	installedDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nShared\n")
	scannedDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nShared\n")

	idx := skill.BuildInstalledIndex([]*skill.Skill{{
		ID:            "installed-github",
		Name:          "demo-skill",
		Path:          installedDir,
		Source:        skill.SourceGitHub,
		SourceURL:     "https://github.com/octo/demo",
		SourceSubPath: "skills/demo-skill",
	}})

	resolved := resolveToolSkillCandidates([]*skill.Skill{{
		Name: "demo-skill",
		Path: scannedDir,
	}}, idx, &toolPresenceIndex{toolsByKey: map[string][]string{}})

	require.Len(t, resolved, 1)
	assert.Equal(t, string(skill.SourceGitHub), resolved[0].Source)
	assert.True(t, resolved[0].Installed)
	assert.True(t, resolved[0].Imported)
}

func TestAggregateToolSkillEntriesPreservesInstalledSource(t *testing.T) {
	installedDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nShared\n")
	pushDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nShared\n")
	scanDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nShared\n")

	idx := skill.BuildInstalledIndex([]*skill.Skill{{
		ID:     "installed-manual",
		Name:   "demo-skill",
		Path:   installedDir,
		Source: skill.SourceManual,
	}})

	entries := aggregateToolSkillEntries(
		[]*skill.Skill{{Name: "demo-skill", Path: pushDir}},
		[]*skill.Skill{{Name: "demo-skill", Path: scanDir}},
		idx,
		&toolPresenceIndex{toolsByKey: map[string][]string{}},
	)

	require.Len(t, entries, 1)
	assert.Equal(t, string(skill.SourceManual), entries[0].Source)
	assert.True(t, entries[0].Pushed)
	assert.True(t, entries[0].SeenInToolScan)
}
