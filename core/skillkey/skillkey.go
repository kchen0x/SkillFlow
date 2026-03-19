package skillkey

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	coregit "github.com/shinerio/skillflow/core/git"
)

type MatchStrength string

const (
	MatchStrengthNone     MatchStrength = ""
	MatchStrengthLogical  MatchStrength = "logical"
	MatchStrengthContent  MatchStrength = "content"
	MatchStrengthFallback MatchStrength = "fallback"
)

// Git returns the stable logical key for a git-backed skill using normalized
// repo source plus normalized repository subpath.
func Git(repoSource, subPath string) string {
	repoSource = strings.TrimSpace(strings.ToLower(repoSource))
	subPath = NormalizeRepoSubPath(subPath)
	if repoSource == "" || subPath == "" {
		return ""
	}
	return fmt.Sprintf("git:%s#%s", repoSource, subPath)
}

// GitFromRepoURL derives a git logical key from a remote URL and subpath.
func GitFromRepoURL(repoURL, subPath string) (string, error) {
	repoSource, err := coregit.RepoSource(repoURL)
	if err != nil {
		return "", err
	}
	return Git(repoSource, subPath), nil
}

func GitFromRepoURLOrEmpty(repoURL, subPath string) string {
	key, err := GitFromRepoURL(repoURL, subPath)
	if err != nil {
		return ""
	}
	return key
}

func NormalizeRepoSubPath(subPath string) string {
	cleaned := strings.TrimSpace(strings.ReplaceAll(subPath, "\\", "/"))
	if cleaned == "" {
		return ""
	}
	cleaned = strings.TrimPrefix(path.Clean("/"+cleaned), "/")
	if cleaned == "" {
		return "."
	}
	return cleaned
}

// ContentFromDir derives a stable logical key from the directory content so the
// same skill can still be correlated across local imports and tool scans.
func ContentFromDir(dir string) (string, error) {
	files, err := listFiles(dir)
	if err != nil {
		return "", err
	}
	hasher := sha256.New()
	for _, rel := range files {
		fullPath := filepath.Join(dir, filepath.FromSlash(rel))
		info, err := os.Lstat(fullPath)
		if err != nil {
			return "", err
		}

		if _, err := io.WriteString(hasher, "file:"+rel+"\n"); err != nil {
			return "", err
		}

		switch {
		case info.Mode()&os.ModeSymlink != 0:
			target, err := os.Readlink(fullPath)
			if err != nil {
				return "", err
			}
			if _, err := io.WriteString(hasher, "symlink:"+filepath.ToSlash(target)+"\n"); err != nil {
				return "", err
			}
		case info.Mode().IsRegular():
			f, err := os.Open(fullPath)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(hasher, f); err != nil {
				f.Close()
				return "", err
			}
			if err := f.Close(); err != nil {
				return "", err
			}
			if _, err := io.WriteString(hasher, "\n"); err != nil {
				return "", err
			}
		}
	}
	return "content:" + hex.EncodeToString(hasher.Sum(nil)), nil
}

func listFiles(root string) ([]string, error) {
	var files []string
	if _, err := os.Stat(root); err != nil {
		return nil, err
	}
	err := filepath.WalkDir(root, func(current string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if current == root {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&os.ModeType != 0 && d.Type()&os.ModeSymlink == 0 {
			return nil
		}
		rel, err := filepath.Rel(root, current)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}
