package backup

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v, output: %s", args, err, string(out))
	}
	return string(out)
}

func runGitWithError(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func writeRepoFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
}

func TestGitProviderSyncInitializesRepoAndPushes(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	base := t.TempDir()
	remoteDir := filepath.Join(base, "remote.git")
	runGit(t, "", "init", "--bare", remoteDir)

	localDir := filepath.Join(base, "skills")
	if err := os.MkdirAll(localDir, 0755); err != nil {
		t.Fatalf("mkdir localDir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "skill.md"), []byte("# test"), 0644); err != nil {
		t.Fatalf("write skill.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(localDir, "cache"), 0755); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "cache", "tmp.bin"), []byte("tmp"), 0644); err != nil {
		t.Fatalf("write cache file: %v", err)
	}

	p := NewGitProvider()
	if err := p.Init(map[string]string{
		"repo_url": remoteDir,
		"branch":   "main",
	}); err != nil {
		t.Fatalf("init provider: %v", err)
	}
	if err := p.Sync(context.Background(), localDir, "", "", func(string) {}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(localDir, ".git")); err != nil {
		t.Fatalf("expected .git directory: %v", err)
	}

	origin := strings.TrimSpace(runGit(t, localDir, "remote", "get-url", "origin"))
	if origin != remoteDir {
		t.Fatalf("unexpected origin: got %q want %q", origin, remoteDir)
	}
	gitignore, err := os.ReadFile(filepath.Join(localDir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !strings.Contains(string(gitignore), "cache/") {
		t.Fatalf("expected .gitignore to contain cache/, got: %q", string(gitignore))
	}
	if !strings.Contains(string(gitignore), "meta_local/") {
		t.Fatalf("expected .gitignore to contain meta_local/, got: %q", string(gitignore))
	}
	if !strings.Contains(string(gitignore), "star_repos_local.json") {
		t.Fatalf("expected .gitignore to contain star_repos_local.json, got: %q", string(gitignore))
	}

	_ = runGit(t, "", "--git-dir", remoteDir, "rev-parse", "--verify", "refs/heads/main")
	remoteFiles := runGit(t, "", "--git-dir", remoteDir, "ls-tree", "-r", "--name-only", "main")
	if strings.Contains(remoteFiles, "cache/") {
		t.Fatalf("cache should not be tracked, remote files: %s", remoteFiles)
	}
}

func TestGitProviderSyncAddsOriginForExistingRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	base := t.TempDir()
	remoteDir := filepath.Join(base, "remote.git")
	runGit(t, "", "init", "--bare", remoteDir)

	localDir := filepath.Join(base, "skills")
	if err := os.MkdirAll(localDir, 0755); err != nil {
		t.Fatalf("mkdir localDir: %v", err)
	}
	runGit(t, localDir, "init")
	if err := os.WriteFile(filepath.Join(localDir, "skill.md"), []byte("# test"), 0644); err != nil {
		t.Fatalf("write skill.md: %v", err)
	}

	p := NewGitProvider()
	if err := p.Init(map[string]string{
		"repo_url": remoteDir,
		"branch":   "main",
	}); err != nil {
		t.Fatalf("init provider: %v", err)
	}
	if err := p.Sync(context.Background(), localDir, "", "", func(string) {}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	origin := strings.TrimSpace(runGit(t, localDir, "remote", "get-url", "origin"))
	if origin != remoteDir {
		t.Fatalf("unexpected origin: got %q want %q", origin, remoteDir)
	}
}

func TestGitProviderSyncDisablesCommitSigningLocally(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	base := t.TempDir()
	remoteDir := filepath.Join(base, "remote.git")
	runGit(t, "", "init", "--bare", remoteDir)

	localDir := filepath.Join(base, "skills")
	if err := os.MkdirAll(localDir, 0755); err != nil {
		t.Fatalf("mkdir localDir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "skill.md"), []byte("# test"), 0644); err != nil {
		t.Fatalf("write skill.md: %v", err)
	}

	p := NewGitProvider()
	if err := p.Init(map[string]string{
		"repo_url": remoteDir,
		"branch":   "main",
	}); err != nil {
		t.Fatalf("init provider: %v", err)
	}
	if err := p.Sync(context.Background(), localDir, "", "", func(string) {}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	out, err := runGitWithError(localDir, "config", "--local", "--get", "commit.gpgsign")
	if err != nil {
		t.Fatalf("expected commit.gpgsign to be configured locally: %v, output: %s", err, out)
	}
	if strings.TrimSpace(out) != "false" {
		t.Fatalf("unexpected local commit.gpgsign: got %q want %q", strings.TrimSpace(out), "false")
	}
}

func TestGitProviderSyncRemovesTrackedExcludedRuntimeDir(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	base := t.TempDir()
	remoteDir := filepath.Join(base, "remote.git")
	runGit(t, "", "init", "--bare", remoteDir)

	localDir := filepath.Join(base, "skills")
	if err := os.MkdirAll(filepath.Join(localDir, "runtime"), 0755); err != nil {
		t.Fatalf("mkdir runtime dir: %v", err)
	}
	runGit(t, localDir, "init")
	runGit(t, localDir, "config", "user.email", "skillflow@test")
	runGit(t, localDir, "config", "user.name", "SkillFlow Test")
	writeRepoFiles(t, localDir, map[string]string{
		"skill.md":                    "# test\n",
		"runtime/helper-control.json": "{\"address\":\"127.0.0.1:9999\"}\n",
	})
	runGit(t, localDir, "add", "-A")
	runGit(t, localDir, "commit", "-m", "seed runtime")
	runGit(t, localDir, "remote", "add", "origin", remoteDir)
	runGit(t, localDir, "push", "-u", "origin", "HEAD:main")

	p := NewGitProvider()
	if err := p.Init(map[string]string{
		"repo_url": remoteDir,
		"branch":   "main",
	}); err != nil {
		t.Fatalf("init provider: %v", err)
	}
	if err := p.Sync(context.Background(), localDir, "", "", func(string) {}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	remoteFiles := runGit(t, "", "--git-dir", remoteDir, "ls-tree", "-r", "--name-only", "main")
	if strings.Contains(remoteFiles, "runtime/") {
		t.Fatalf("runtime should not remain tracked after sync, remote files: %s", remoteFiles)
	}

	gitignore, err := os.ReadFile(filepath.Join(localDir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !strings.Contains(string(gitignore), "runtime/") {
		t.Fatalf("expected .gitignore to contain runtime/, got: %q", string(gitignore))
	}
}

func TestGitProviderRestoreAllowsMissingRemoteBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	base := t.TempDir()
	remoteDir := filepath.Join(base, "remote.git")
	runGit(t, "", "init", "--bare", remoteDir)

	localDir := filepath.Join(base, "skills")
	p := NewGitProvider()
	if err := p.Init(map[string]string{
		"repo_url": remoteDir,
		"branch":   "main",
	}); err != nil {
		t.Fatalf("init provider: %v", err)
	}

	if err := p.Restore(context.Background(), "", "", localDir); err != nil {
		t.Fatalf("restore should allow missing remote branch, got: %v", err)
	}

	if _, err := os.Stat(filepath.Join(localDir, ".git")); err != nil {
		t.Fatalf("expected .git directory: %v", err)
	}
	origin := strings.TrimSpace(runGit(t, localDir, "remote", "get-url", "origin"))
	if origin != remoteDir {
		t.Fatalf("unexpected origin: got %q want %q", origin, remoteDir)
	}
}

func TestParseConflictFilesFromOutput(t *testing.T) {
	out := `
Auto-merging skills/a/skill.md
CONFLICT (content): Merge conflict in skills/a/skill.md
Auto-merging meta/123.json
CONFLICT (content): Merge conflict in meta/123.json
`
	files := parseConflictFilesFromOutput(out)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %#v", len(files), files)
	}
	if files[0] != "skills/a/skill.md" || files[1] != "meta/123.json" {
		t.Fatalf("unexpected files: %#v", files)
	}
}

func TestGitProviderResolveConflictUseRemoteWithoutInit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	base := t.TempDir()
	remoteDir := filepath.Join(base, "remote.git")
	runGit(t, "", "init", "--bare", remoteDir)

	seedDir := filepath.Join(base, "seed")
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		t.Fatalf("mkdir seedDir: %v", err)
	}
	runGit(t, seedDir, "init")
	runGit(t, seedDir, "config", "user.email", "skillflow@test")
	runGit(t, seedDir, "config", "user.name", "SkillFlow Test")
	runGit(t, seedDir, "checkout", "-b", "master")
	writeRepoFiles(t, seedDir, map[string]string{
		"config.json":     "{\n  \"source\": \"base\"\n}\n",
		"star_repos.json": "[\n  \"base\"\n]\n",
	})
	runGit(t, seedDir, "add", "-A")
	runGit(t, seedDir, "commit", "-m", "base")
	runGit(t, seedDir, "remote", "add", "origin", remoteDir)
	runGit(t, seedDir, "push", "-u", "origin", "master")
	runGit(t, "", "--git-dir", remoteDir, "symbolic-ref", "HEAD", "refs/heads/master")

	localDir := filepath.Join(base, "local")
	runGit(t, "", "clone", remoteDir, localDir)
	runGit(t, localDir, "config", "user.email", "skillflow@test")
	runGit(t, localDir, "config", "user.name", "SkillFlow Test")

	remoteWorkDir := filepath.Join(base, "remote-work")
	runGit(t, "", "clone", remoteDir, remoteWorkDir)
	runGit(t, remoteWorkDir, "config", "user.email", "skillflow@test")
	runGit(t, remoteWorkDir, "config", "user.name", "SkillFlow Test")
	writeRepoFiles(t, remoteWorkDir, map[string]string{
		"config.json":     "{\n  \"source\": \"remote\"\n}\n",
		"star_repos.json": "[\n  \"remote\"\n]\n",
	})
	runGit(t, remoteWorkDir, "add", "-A")
	runGit(t, remoteWorkDir, "commit", "-m", "remote change")
	runGit(t, remoteWorkDir, "push", "origin", "master")

	writeRepoFiles(t, localDir, map[string]string{
		"config.json":     "{\n  \"source\": \"local\"\n}\n",
		"star_repos.json": "[\n  \"local\"\n]\n",
	})
	runGit(t, localDir, "add", "-A")
	runGit(t, localDir, "commit", "-m", "local change")

	pullOut, err := runGitWithError(localDir, "pull", "--no-rebase", "--allow-unrelated-histories", "origin", "master")
	if err == nil {
		t.Fatal("expected git pull conflict")
	}
	if !strings.Contains(pullOut, "CONFLICT") {
		t.Fatalf("expected conflict output, got: %s", pullOut)
	}

	conflictedConfig, err := os.ReadFile(filepath.Join(localDir, "config.json"))
	if err != nil {
		t.Fatalf("read conflicted config.json: %v", err)
	}
	if !strings.Contains(string(conflictedConfig), "<<<<<<<") {
		t.Fatalf("expected conflict markers in config.json, got: %s", string(conflictedConfig))
	}

	p := NewGitProvider()
	if err := p.ResolveConflictUseRemote(localDir); err != nil {
		t.Fatalf("ResolveConflictUseRemote() failed: %v", err)
	}

	configData, err := os.ReadFile(filepath.Join(localDir, "config.json"))
	if err != nil {
		t.Fatalf("read resolved config.json: %v", err)
	}
	if got := strings.ReplaceAll(string(configData), "\r\n", "\n"); got != "{\n  \"source\": \"remote\"\n}\n" {
		t.Fatalf("resolved config.json=%q", got)
	}

	starData, err := os.ReadFile(filepath.Join(localDir, "star_repos.json"))
	if err != nil {
		t.Fatalf("read resolved star_repos.json: %v", err)
	}
	if got := strings.ReplaceAll(string(starData), "\r\n", "\n"); got != "[\n  \"remote\"\n]\n" {
		t.Fatalf("resolved star_repos.json=%q", got)
	}

	statusOut := strings.TrimSpace(runGit(t, localDir, "status", "--short"))
	if statusOut != "" {
		t.Fatalf("expected clean worktree after resolution, got: %s", statusOut)
	}
}

func TestGitProviderPendingChanges(t *testing.T) {
	remoteDir := t.TempDir()
	runGit(t, remoteDir, "init", "--bare")

	localDir := t.TempDir()
	p := NewGitProvider()
	if err := p.Init(map[string]string{
		"repo_url": remoteDir,
		"branch":   "main",
	}); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if err := p.ensureRepo(localDir); err != nil {
		t.Fatalf("ensureRepo failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(localDir, "existing.txt"), []byte("one"), 0644); err != nil {
		t.Fatalf("WriteFile existing.txt failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "delete.txt"), []byte("gone"), 0644); err != nil {
		t.Fatalf("WriteFile delete.txt failed: %v", err)
	}
	runGit(t, localDir, "add", "existing.txt", "delete.txt")
	runGit(t, localDir, "commit", "-m", "base")
	runGit(t, localDir, "push", "-u", "origin", "HEAD:main")

	if err := os.WriteFile(filepath.Join(localDir, "existing.txt"), []byte("two"), 0644); err != nil {
		t.Fatalf("WriteFile existing.txt update failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "added.txt"), []byte("new"), 0644); err != nil {
		t.Fatalf("WriteFile added.txt failed: %v", err)
	}
	if err := os.Remove(filepath.Join(localDir, "delete.txt")); err != nil {
		t.Fatalf("Remove delete.txt failed: %v", err)
	}

	changes, err := p.PendingChanges(localDir)
	if err != nil {
		t.Fatalf("PendingChanges failed: %v", err)
	}

	want := map[string]string{
		"added.txt":    "added",
		"delete.txt":   "deleted",
		"existing.txt": "modified",
	}
	if len(changes) != len(want) {
		t.Fatalf("expected %d changes, got %d: %+v", len(want), len(changes), changes)
	}
	for _, change := range changes {
		if want[change.Path] != change.Action {
			t.Fatalf("unexpected action for %s: got %s want %s", change.Path, change.Action, want[change.Path])
		}
	}
}
