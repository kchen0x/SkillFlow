package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	backupdomain "github.com/shinerio/skillflow/core/backup/domain"
	snapshotinfra "github.com/shinerio/skillflow/core/backup/infra/snapshot"
)

type GitProvider struct {
	repoURL  string
	branch   string
	username string
	token    string
	localDir string
}

func NewGitProvider() *GitProvider { return &GitProvider{} }

func init() {
	RegisterProviderFactory(func() backupdomain.CloudProvider { return NewGitProvider() })
}

func (p *GitProvider) Name() string { return backupdomain.GitProviderName }

func (p *GitProvider) RequiredCredentials() []backupdomain.CredentialField {
	return []backupdomain.CredentialField{
		{Key: "repo_url", Label: "Git 仓库地址", Placeholder: "https://github.com/user/my-backup.git"},
		{Key: "branch", Label: "分支（留空默认 main）", Placeholder: "main"},
		{Key: "username", Label: "用户名（HTTPS 认证，可选）", Placeholder: "your-username"},
		{Key: "token", Label: "访问令牌（HTTPS 认证，可选）", Placeholder: "ghp_xxxx", Secret: true},
	}
}

func (p *GitProvider) Init(credentials map[string]string) error {
	p.repoURL = strings.TrimSpace(credentials["repo_url"])
	if p.repoURL == "" {
		return fmt.Errorf("git 仓库地址不能为空")
	}
	p.branch = strings.TrimSpace(credentials["branch"])
	if p.branch == "" {
		p.branch = "main"
	}
	p.username = strings.TrimSpace(credentials["username"])
	p.token = strings.TrimSpace(credentials["token"])
	return nil
}

func (p *GitProvider) authenticatedURL() string {
	if p.username != "" && p.token != "" && strings.HasPrefix(p.repoURL, "https://") {
		return "https://" + p.username + ":" + p.token + "@" + p.repoURL[8:]
	}
	return p.repoURL
}

func (p *GitProvider) run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	hideConsole(cmd)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func ensureIgnoredPath(localDir, ignoredPath string) error {
	gitignorePath := filepath.Join(localDir, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	normalized := strings.ReplaceAll(string(content), "\r\n", "\n")
	for _, line := range strings.Split(normalized, "\n") {
		if strings.TrimSpace(line) == ignoredPath {
			return nil
		}
	}

	if len(normalized) > 0 && !strings.HasSuffix(normalized, "\n") {
		normalized += "\n"
	}
	normalized += ignoredPath + "\n"
	return os.WriteFile(gitignorePath, []byte(normalized), 0644)
}

func (p *GitProvider) isGitRepo(localDir string) bool {
	out, err := p.run(localDir, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) == "true"
}

func (p *GitProvider) ensureExistingRepo(localDir string) error {
	if !p.isGitRepo(localDir) {
		return fmt.Errorf("git repo not initialized: %s", localDir)
	}
	return nil
}

func (p *GitProvider) resolveRepoBranch(localDir string) (string, error) {
	if out, err := p.run(localDir, "branch", "--show-current"); err == nil {
		branch := strings.TrimSpace(out)
		if branch != "" {
			return branch, nil
		}
	}
	if out, err := p.run(localDir, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		branch := strings.TrimSpace(out)
		if branch != "" && branch != "HEAD" {
			return branch, nil
		}
	}
	if p.branch != "" {
		return p.branch, nil
	}
	return "", fmt.Errorf("git branch not found for repo: %s", localDir)
}

func isMissingRemoteRef(out string) bool {
	lower := strings.ToLower(out)
	return strings.Contains(lower, "couldn't find remote ref") ||
		strings.Contains(lower, "remote ref does not exist") ||
		strings.Contains(lower, "no such ref was fetched")
}

func isGitConflictOutput(out string) bool {
	lower := strings.ToLower(out)
	return strings.Contains(lower, "conflict") ||
		strings.Contains(lower, "non-fast-forward") ||
		strings.Contains(lower, "rejected") ||
		strings.Contains(lower, "divergent") ||
		strings.Contains(lower, "not possible to fast-forward") ||
		strings.Contains(lower, "need to specify how to reconcile divergent branches")
}

func parseConflictFilesFromOutput(out string) []string {
	lines := strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "CONFLICT") {
			continue
		}
		if idx := strings.LastIndex(strings.ToLower(line), " in "); idx >= 0 && idx+4 < len(line) {
			path := strings.TrimSpace(line[idx+4:])
			if path != "" {
				files = append(files, path)
			}
		}
	}
	return uniqueStrings(files)
}

func parseNameOnlyOutput(out string) []string {
	var files []string
	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		files = append(files, line)
	}
	return files
}

func parsePorcelainConflictFiles(out string) []string {
	var files []string
	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if len(line) < 4 {
			continue
		}
		status := line[:2]
		if status == "UU" || status == "AA" || status == "DD" ||
			status == "AU" || status == "UA" || status == "DU" || status == "UD" {
			files = append(files, strings.TrimSpace(line[3:]))
		}
	}
	return files
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func (p *GitProvider) collectConflictFiles(localDir, branch, output string) []string {
	files := parseConflictFilesFromOutput(output)
	if out, err := p.run(localDir, "diff", "--name-only", "--diff-filter=U"); err == nil {
		files = append(files, parseNameOnlyOutput(out)...)
	}
	p.run(localDir, "fetch", "origin", branch) //nolint
	if _, err := p.run(localDir, "rev-parse", "--verify", "HEAD"); err == nil {
		if out, err := p.run(localDir, "diff", "--name-only", "HEAD..origin/"+branch); err == nil {
			files = append(files, parseNameOnlyOutput(out)...)
		}
		if out, err := p.run(localDir, "diff", "--name-only", "origin/"+branch+"..HEAD"); err == nil {
			files = append(files, parseNameOnlyOutput(out)...)
		}
	}
	if out, err := p.run(localDir, "status", "--porcelain"); err == nil {
		files = append(files, parsePorcelainConflictFiles(out)...)
	}
	return uniqueStrings(files)
}

func (p *GitProvider) ensureRepo(localDir string) error {
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}
	gitDir := filepath.Join(localDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) || !p.isGitRepo(localDir) {
		if out, err := p.run(localDir, "init"); err != nil {
			return fmt.Errorf("git init 失败: %s", out)
		}
	}
	authURL := p.authenticatedURL()
	remoteOut, remoteErr := p.run(localDir, "remote", "get-url", "origin")
	if remoteErr != nil {
		if out, err := p.run(localDir, "remote", "add", "origin", authURL); err != nil {
			return fmt.Errorf("git remote add 失败: %s", out)
		}
	} else if strings.TrimSpace(remoteOut) != authURL {
		if out, err := p.run(localDir, "remote", "set-url", "origin", authURL); err != nil {
			return fmt.Errorf("git remote set-url 失败: %s", out)
		}
	}
	if out, err := p.run(localDir, "config", "user.email", "skillflow@local"); err != nil {
		return fmt.Errorf("git config user.email 失败: %s", out)
	}
	if out, err := p.run(localDir, "config", "user.name", "SkillFlow"); err != nil {
		return fmt.Errorf("git config user.name 失败: %s", out)
	}
	if out, err := p.run(localDir, "config", "commit.gpgsign", "false"); err != nil {
		return fmt.Errorf("git config commit.gpgsign 失败: %s", out)
	}
	for _, dir := range snapshotinfra.ExcludedDirectories() {
		if err := ensureIgnoredPath(localDir, dir+"/"); err != nil {
			return fmt.Errorf("写入 .gitignore 失败: %w", err)
		}
	}
	for _, file := range snapshotinfra.ExcludedFiles() {
		if err := ensureIgnoredPath(localDir, file); err != nil {
			return fmt.Errorf("写入 .gitignore 失败: %w", err)
		}
	}
	for _, pattern := range snapshotinfra.ExcludedPatterns() {
		if err := ensureIgnoredPath(localDir, pattern); err != nil {
			return fmt.Errorf("写入 .gitignore 失败: %w", err)
		}
	}
	return nil
}

// removeExcludedFromCache untracks any files that are currently indexed but
// should not be backed up (e.g. *local.json files that were committed before
// the .gitignore entry was added).
func (p *GitProvider) removeExcludedFromCache(localDir string) {
	out, err := p.run(localDir, "ls-files")
	if err != nil {
		return
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && snapshotinfra.ShouldSkipBackupPath(line) {
			p.run(localDir, "rm", "--cached", "--ignore-unmatch", line) //nolint
		}
	}
}

func (p *GitProvider) Sync(_ context.Context, localDir, _, _ string, onProgress func(file string)) error {
	p.localDir = localDir
	if err := p.ensureRepo(localDir); err != nil {
		return err
	}
	onProgress("git add")
	if out, err := p.run(localDir, "add", "-A"); err != nil {
		return fmt.Errorf("git add 失败: %s", out)
	}
	for _, dir := range snapshotinfra.ExcludedDirectories() {
		if out, err := p.run(localDir, "rm", "-r", "--cached", "--ignore-unmatch", dir); err != nil {
			return fmt.Errorf("git rm --cached %s 失败: %s", dir, out)
		}
	}
	p.removeExcludedFromCache(localDir)
	statusOut, _ := p.run(localDir, "status", "--porcelain")
	_, headErr := p.run(localDir, "rev-parse", "--verify", "HEAD")
	if strings.TrimSpace(statusOut) == "" && headErr != nil {
		onProgress("up-to-date")
		return nil
	}
	if strings.TrimSpace(statusOut) != "" {
		onProgress("git commit")
		if out, err := p.run(localDir, "commit", "-m", "SkillFlow auto-backup"); err != nil {
			return fmt.Errorf("git commit 失败: %s", out)
		}
	}
	onProgress("git push")
	out, err := p.run(localDir, "push", "origin", "HEAD:"+p.branch)
	if err != nil {
		if out2, err2 := p.run(localDir, "push", "--set-upstream", "origin", "HEAD:"+p.branch); err2 != nil {
			if isGitConflictOutput(out) || isGitConflictOutput(out2) {
				return &backupdomain.GitConflictError{
					Output: out + out2,
					Files:  p.collectConflictFiles(localDir, p.branch, out+out2),
				}
			}
			return fmt.Errorf("git push 失败: %s %s", out, out2)
		}
	}
	return nil
}

func (p *GitProvider) autoCommitLocal(localDir string) {
	p.run(localDir, "add", "-A") //nolint
	for _, dir := range snapshotinfra.ExcludedDirectories() {
		p.run(localDir, "rm", "-r", "--cached", "--ignore-unmatch", dir) //nolint
	}
	p.removeExcludedFromCache(localDir)
	statusOut, _ := p.run(localDir, "status", "--porcelain")
	if strings.TrimSpace(statusOut) != "" {
		p.run(localDir, "commit", "-m", "SkillFlow: pre-pull auto-commit") //nolint
	}
}

func (p *GitProvider) Restore(_ context.Context, _, _, localDir string) error {
	p.localDir = localDir
	if err := p.ensureRepo(localDir); err != nil {
		return err
	}
	p.autoCommitLocal(localDir)
	out, err := p.run(localDir, "pull", "origin", p.branch, "--allow-unrelated-histories")
	if err != nil {
		if isMissingRemoteRef(out) {
			return nil
		}
		if isGitConflictOutput(out) || strings.Contains(out, "Automatic merge failed") {
			return &backupdomain.GitConflictError{
				Output: out,
				Files:  p.collectConflictFiles(localDir, p.branch, out),
			}
		}
		return fmt.Errorf("git pull 失败: %s", out)
	}
	return nil
}

func (p *GitProvider) PendingChanges(localDir string) ([]backupdomain.RemoteFile, error) {
	if err := p.ensureRepo(localDir); err != nil {
		return nil, err
	}
	out, err := p.run(localDir, "status", "--porcelain", "-uall")
	if err != nil {
		return nil, fmt.Errorf("git status failed: %s", out)
	}
	return parseGitStatusChanges(localDir, out), nil
}

func (p *GitProvider) List(_ context.Context, _, _ string) ([]backupdomain.RemoteFile, error) {
	if p.localDir == "" {
		return nil, nil
	}
	out, err := p.run(p.localDir, "ls-files")
	if err != nil {
		return nil, nil
	}
	var files []backupdomain.RemoteFile
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		full := filepath.Join(p.localDir, line)
		info, statErr := os.Stat(full)
		var size int64
		if statErr == nil {
			size = info.Size()
		}
		files = append(files, backupdomain.RemoteFile{Path: line, Size: size})
	}
	return files, nil
}

func (p *GitProvider) ResolveConflictUseLocal(localDir string) error {
	if err := p.ensureExistingRepo(localDir); err != nil {
		return err
	}
	branch, err := p.resolveRepoBranch(localDir)
	if err != nil {
		return err
	}
	p.run(localDir, "merge", "--abort")                                        //nolint
	p.run(localDir, "add", "-A")                                               //nolint
	p.run(localDir, "commit", "-m", "SkillFlow: resolve conflict (use local)") //nolint
	out, err := p.run(localDir, "push", "origin", "HEAD:"+branch, "--force-with-lease")
	if err != nil {
		if out2, err2 := p.run(localDir, "push", "origin", "HEAD:"+branch, "--force"); err2 != nil {
			return fmt.Errorf("git push --force 失败: %s %s", out, out2)
		}
	}
	return nil
}

func (p *GitProvider) ResolveConflictUseRemote(localDir string) error {
	if err := p.ensureExistingRepo(localDir); err != nil {
		return err
	}
	branch, err := p.resolveRepoBranch(localDir)
	if err != nil {
		return err
	}
	p.run(localDir, "merge", "--abort") //nolint
	if out, err := p.run(localDir, "fetch", "origin", branch); err != nil {
		return fmt.Errorf("git fetch 失败: %s", out)
	}
	if out, err := p.run(localDir, "reset", "--hard", "origin/"+branch); err != nil {
		return fmt.Errorf("git reset 失败: %s", out)
	}
	return nil
}

func parseGitStatusChanges(localDir, output string) []backupdomain.RemoteFile {
	changes := make([]backupdomain.RemoteFile, 0)
	normalized := strings.ReplaceAll(output, "\r\n", "\n")
	for _, line := range strings.Split(normalized, "\n") {
		line = strings.TrimRight(line, "\r")
		if len(line) < 3 {
			continue
		}
		status := line[:2]
		path := strings.TrimSpace(line[3:])
		if idx := strings.LastIndex(path, " -> "); idx >= 0 {
			path = strings.TrimSpace(path[idx+4:])
		}
		if path == "" || path == ".gitignore" || snapshotinfra.ShouldSkipBackupPath(path) {
			continue
		}
		action := "modified"
		switch {
		case status == "??":
			action = "added"
		case strings.Contains(status, "D"):
			action = "deleted"
		case strings.Contains(status, "A"):
			action = "added"
		}
		var size int64
		if action != "deleted" {
			if info, err := os.Stat(filepath.Join(localDir, filepath.FromSlash(path))); err == nil {
				size = info.Size()
			}
		}
		changes = append(changes, backupdomain.RemoteFile{Path: path, Size: size, Action: action})
	}
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Path == changes[j].Path {
			return changes[i].Action < changes[j].Action
		}
		return changes[i].Path < changes[j].Path
	})
	return changes
}
