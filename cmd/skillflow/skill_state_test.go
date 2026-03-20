package main

import (
	"testing"

	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAgentSkillCandidatesIncludesInstalledSource(t *testing.T) {
	installedDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nShared\n")
	scannedDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nShared\n")

	idx := skillquery.BuildInstalledIndex([]*skilldomain.InstalledSkill{{
		ID:            "installed-github",
		Name:          "demo-skill",
		Path:          installedDir,
		Source:        skilldomain.SourceGitHub,
		SourceURL:     "https://github.com/octo/demo",
		SourceSubPath: "skills/demo-skill",
	}})

	resolved := resolveAgentSkillCandidates([]*skilldomain.InstalledSkill{{
		Name: "demo-skill",
		Path: scannedDir,
	}}, idx, &agentPresenceIndex{agentsByKey: map[string][]string{}})

	require.Len(t, resolved, 1)
	assert.Equal(t, string(skilldomain.SourceGitHub), resolved[0].Source)
	assert.True(t, resolved[0].Installed)
	assert.True(t, resolved[0].Imported)
}

func TestAggregateAgentSkillEntriesPreservesInstalledSource(t *testing.T) {
	installedDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nShared\n")
	pushDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nShared\n")
	scanDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nShared\n")

	idx := skillquery.BuildInstalledIndex([]*skilldomain.InstalledSkill{{
		ID:     "installed-manual",
		Name:   "demo-skill",
		Path:   installedDir,
		Source: skilldomain.SourceManual,
	}})

	entries := aggregateAgentSkillEntries(
		[]*skilldomain.InstalledSkill{{Name: "demo-skill", Path: pushDir}},
		[]*skilldomain.InstalledSkill{{Name: "demo-skill", Path: scanDir}},
		idx,
		&agentPresenceIndex{agentsByKey: map[string][]string{}},
	)

	require.Len(t, entries, 1)
	assert.Equal(t, string(skilldomain.SourceManual), entries[0].Source)
	assert.True(t, entries[0].Pushed)
	assert.True(t, entries[0].SeenInAgentScan)
}

func TestAggregateAgentSkillEntriesMarksStalePushedCopyUpdatable(t *testing.T) {
	installedDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nNew\n")
	pushDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nOld\n")

	idx := skillquery.BuildInstalledIndex([]*skilldomain.InstalledSkill{{
		ID:            "installed-github",
		Name:          "demo-skill",
		Path:          installedDir,
		Source:        skilldomain.SourceGitHub,
		SourceURL:     "https://github.com/octo/demo",
		SourceSubPath: "skills/demo-skill",
	}})

	entries := aggregateAgentSkillEntries(
		[]*skilldomain.InstalledSkill{{Name: "demo-skill", Path: pushDir}},
		nil,
		idx,
		&agentPresenceIndex{agentsByKey: map[string][]string{}},
	)

	require.Len(t, entries, 1)
	assert.True(t, entries[0].Pushed)
	assert.True(t, entries[0].Updatable)
}

func TestAggregateAgentSkillEntriesDoesNotMarkMatchingPushedCopyUpdatable(t *testing.T) {
	installedDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nSame\n")
	pushDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nSame\n")

	idx := skillquery.BuildInstalledIndex([]*skilldomain.InstalledSkill{{
		ID:            "installed-github",
		Name:          "demo-skill",
		Path:          installedDir,
		Source:        skilldomain.SourceGitHub,
		SourceURL:     "https://github.com/octo/demo",
		SourceSubPath: "skills/demo-skill",
	}})

	entries := aggregateAgentSkillEntries(
		[]*skilldomain.InstalledSkill{{Name: "demo-skill", Path: pushDir}},
		nil,
		idx,
		&agentPresenceIndex{agentsByKey: map[string][]string{}},
	)

	require.Len(t, entries, 1)
	assert.False(t, entries[0].Updatable)
}
