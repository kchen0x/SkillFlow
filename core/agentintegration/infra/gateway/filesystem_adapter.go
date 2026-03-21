package gateway

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/shinerio/skillflow/core/config"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
)

type FilesystemAdapter struct {
	name             string
	defaultSkillsDir string
}

func NewFilesystemAdapter(name, defaultSkillsDir string) *FilesystemAdapter {
	return &FilesystemAdapter{name: name, defaultSkillsDir: defaultSkillsDir}
}

func (f *FilesystemAdapter) Name() string             { return f.name }
func (f *FilesystemAdapter) DefaultSkillsDir() string { return f.defaultSkillsDir }

func (f *FilesystemAdapter) Push(_ context.Context, skills []*skilldomain.InstalledSkill, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}
	for _, skill := range skills {
		dst := filepath.Join(targetDir, skill.Name)
		if err := copyDir(skill.Path, dst); err != nil {
			return err
		}
	}
	return nil
}

func (f *FilesystemAdapter) Pull(ctx context.Context, sourceDir string) ([]*skilldomain.InstalledSkill, error) {
	return f.PullWithMaxDepth(ctx, sourceDir, config.DefaultRepoScanMaxDepth)
}

func (f *FilesystemAdapter) PullWithMaxDepth(_ context.Context, sourceDir string, maxDepth int) ([]*skilldomain.InstalledSkill, error) {
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("目录不存在: %s", sourceDir)
	}
	if maxDepth < 0 {
		maxDepth = 0
	}
	var skills []*skilldomain.InstalledSkill
	var walk func(dir string, depth int)
	walk = func(dir string, depth int) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, entry := range entries {
			if !entry.IsDir() && isSkillMd(entry.Name()) {
				skills = append(skills, &skilldomain.InstalledSkill{
					Name:   filepath.Base(dir),
					Path:   dir,
					Source: skilldomain.SourceManual,
				})
				return
			}
		}
		if depth >= maxDepth {
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				walk(filepath.Join(dir, entry.Name()), depth+1)
			}
		}
	}
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			walk(filepath.Join(sourceDir, entry.Name()), 0)
		}
	}
	return skills, nil
}

func isSkillMd(name string) bool {
	return strings.ToLower(name) == "skill.md"
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
