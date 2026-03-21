package skills

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/readmodel/viewstate"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePresenceBuilder struct {
	keysByAgent map[string][]string
	calls       []string
	err         error
}

func (f *fakePresenceBuilder) BuildPresenceIndex(_ context.Context, profiles []agentdomain.AgentProfile, _ *skillquery.InstalledIndex, _ int) (*agentdomain.AgentPresenceIndex, error) {
	if f.err != nil {
		return nil, f.err
	}
	presence := agentdomain.NewAgentPresenceIndex()
	for _, profile := range profiles {
		f.calls = append(f.calls, profile.Name)
		presence.Add(profile.Name, f.keysByAgent[profile.Name]...)
	}
	return presence, nil
}

func TestPresenceResolverUsesSnapshotAndRebuildsOnlyChangedAgent(t *testing.T) {
	cache := viewstate.NewManager(t.TempDir())
	builder := &fakePresenceBuilder{
		keysByAgent: map[string][]string{
			"codex":  {"skill:a"},
			"claude": {"skill:b"},
		},
	}
	resolver := NewPresenceResolver(cache, builder)

	codexDir := t.TempDir()
	claudeDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(codexDir, "a.txt"), []byte("a"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "b.txt"), []byte("b"), 0644))

	profiles := []agentdomain.AgentProfile{
		{Name: "codex", PushDir: codexDir},
		{Name: "claude", PushDir: claudeDir},
	}
	idx := skillquery.BuildInstalledIndex(nil)

	first, err := resolver.Resolve(context.Background(), idx, 2, profiles)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"codex", "claude"}, builder.calls)
	assert.Equal(t, []string{"codex"}, first.Agents("skill:a"))
	assert.Equal(t, []string{"claude"}, first.Agents("skill:b"))

	builder.calls = nil
	second, err := resolver.Resolve(context.Background(), idx, 2, profiles)
	require.NoError(t, err)
	assert.Empty(t, builder.calls)
	assert.Equal(t, []string{"codex"}, second.Agents("skill:a"))
	assert.Equal(t, []string{"claude"}, second.Agents("skill:b"))

	require.NoError(t, os.WriteFile(filepath.Join(codexDir, "new.txt"), []byte("new"), 0644))
	builder.calls = nil
	_, err = resolver.Resolve(context.Background(), idx, 2, profiles)
	require.NoError(t, err)
	assert.Equal(t, []string{"codex"}, builder.calls)
}

func TestDirectorySummaryFingerprintHandlesEmptyAndMissing(t *testing.T) {
	empty, err := DirectorySummaryFingerprint("")
	require.NoError(t, err)

	missing, err := DirectorySummaryFingerprint(filepath.Join(t.TempDir(), "missing"))
	require.NoError(t, err)

	assert.NotEmpty(t, empty)
	assert.NotEmpty(t, missing)
	assert.NotEqual(t, empty, missing)
}
