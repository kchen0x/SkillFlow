package git

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const defaultMaxRecursiveScanDepth = 5

// scanTree walks a directory tree and returns every directory that contains a
// skill.md file (case-insensitive). Once a directory is identified as a skill,
// the walk stops descending into that subtree.
func scanTree(repoDir, dir, repoURL, repoName, source string, depth, maxDepth int) ([]StarSkill, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	if hasSkillMd(entries) {
		rel, err := filepath.Rel(repoDir, dir)
		if err != nil {
			return nil, err
		}
		return []StarSkill{{
			Name:     filepath.Base(dir),
			Path:     dir,
			SubPath:  filepath.ToSlash(rel),
			RepoURL:  repoURL,
			RepoName: repoName,
			Source:   source,
		}}, nil
	}
	if depth >= maxDepth {
		return nil, nil
	}

	var result []StarSkill
	for _, e := range entries {
		if !e.IsDir() || shouldSkipScanDir(e.Name()) {
			continue
		}
		skills, err := scanTree(repoDir, filepath.Join(dir, e.Name()), repoURL, repoName, source, depth+1, maxDepth)
		if err != nil {
			return nil, err
		}
		result = append(result, skills...)
	}
	return result, nil
}

func hasSkillMd(entries []os.DirEntry) bool {
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(e.Name(), "skill.md") {
			return true
		}
	}
	return false
}

func shouldSkipScanDir(name string) bool {
	return strings.HasPrefix(name, ".")
}

// ScanSkills looks for skill directories anywhere inside the given repo clone
// using the default recursive depth limit.
func ScanSkills(repoDir, repoURL, repoName, source string) ([]StarSkill, error) {
	return ScanSkillsWithMaxDepth(repoDir, repoURL, repoName, source, defaultMaxRecursiveScanDepth)
}

// ScanSkillsWithMaxDepth looks for skill directories anywhere inside the given
// repo clone. It scans recursively, supports skill.md in any casing, stops
// descending once a directory is identified as a skill, and bounds recursion
// depth to protect against pathological nested trees.
func ScanSkillsWithMaxDepth(repoDir, repoURL, repoName, source string, maxDepth int) ([]StarSkill, error) {
	if maxDepth < 0 {
		maxDepth = 0
	}
	entries, err := os.ReadDir(repoDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if hasSkillMd(entries) {
		return scanTree(repoDir, repoDir, repoURL, repoName, source, 0, maxDepth)
	}

	var roots []string
	skillsRoot := filepath.Join(repoDir, "skills")
	if info, err := os.Stat(skillsRoot); err == nil && info.IsDir() {
		roots = append(roots, skillsRoot)
	}
	for _, e := range entries {
		if !e.IsDir() || shouldSkipScanDir(e.Name()) {
			continue
		}
		dir := filepath.Join(repoDir, e.Name())
		if dir == skillsRoot {
			continue
		}
		roots = append(roots, dir)
	}

	var result []StarSkill
	for _, root := range roots {
		skills, err := scanTree(repoDir, root, repoURL, repoName, source, 0, maxDepth)
		if err != nil {
			return nil, err
		}
		result = append(result, skills...)
	}
	return result, nil
}
