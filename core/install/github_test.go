package install_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shinerio/skillflow/core/install"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockGitHubServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	// Mock: list skills directory contents
	mux.HandleFunc("/repos/user/repo/contents/skills", func(w http.ResponseWriter, r *http.Request) {
		items := []map[string]any{
			{"name": "skill-a", "type": "dir", "path": "skills/skill-a"},
			{"name": "skill-b", "type": "dir", "path": "skills/skill-b"},
			{"name": "readme.md", "type": "file", "path": "skills/readme.md"},
		}
		json.NewEncoder(w).Encode(items)
	})
	// Mock: check skill.mdexistence for skill-a (returns file info)
	mux.HandleFunc("/repos/user/repo/contents/skills/skill-a/skill.md", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"name": "skill.md", "type": "file"})
	})
	// Mock: skill-b has no skill.md(404)
	mux.HandleFunc("/repos/user/repo/contents/skills/skill-b/skill.md", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	return httptest.NewServer(mux)
}

func TestGitHubInstallerScan(t *testing.T) {
	srv := mockGitHubServer(t)
	defer srv.Close()

	installer := install.NewGitHubInstaller(srv.URL, nil)
	candidates, err := installer.Scan(context.Background(), install.InstallSource{
		Type: "github",
		URI:  srv.URL + "/repos/user/repo",
	})
	require.NoError(t, err)
	// Only skill-a has skill.md, skill-b does not
	assert.Len(t, candidates, 1)
	assert.Equal(t, "skill-a", candidates[0].Name)
}

func TestGitHubInstallerScanSSHURI(t *testing.T) {
	srv := mockGitHubServer(t)
	defer srv.Close()

	installer := install.NewGitHubInstaller(srv.URL, nil)
	candidates, err := installer.Scan(context.Background(), install.InstallSource{
		Type: "github",
		URI:  "git@github.com:user/repo.git",
	})
	require.NoError(t, err)
	assert.Len(t, candidates, 1)
	assert.Equal(t, "skill-a", candidates[0].Name)
}
