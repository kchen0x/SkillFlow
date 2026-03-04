package sync

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/shinerio/skillflow/core/skill"
)

// FilesystemAdapter works for all tools — they all share the same file-based skills directory model.
type FilesystemAdapter struct {
	name             string
	defaultSkillsDir string
}

func NewFilesystemAdapter(name, defaultSkillsDir string) *FilesystemAdapter {
	return &FilesystemAdapter{name: name, defaultSkillsDir: defaultSkillsDir}
}

func (f *FilesystemAdapter) Name() string             { return f.name }
func (f *FilesystemAdapter) DefaultSkillsDir() string { return f.defaultSkillsDir }

func (f *FilesystemAdapter) Push(_ context.Context, skills []*skill.Skill, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}
	for _, sk := range skills {
		dst := filepath.Join(targetDir, sk.Name)
		if err := copyDir(sk.Path, dst); err != nil {
			return err
		}
	}
	return nil
}

func (f *FilesystemAdapter) Pull(_ context.Context, sourceDir string) ([]*skill.Skill, error) {
	validator := skill.NewValidator()
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, err
	}
	var skills []*skill.Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(sourceDir, e.Name())
		if err := validator.Validate(dir); err == nil {
			skills = append(skills, &skill.Skill{
				Name:   e.Name(),
				Path:   dir,
				Source: skill.SourceManual,
			})
		}
	}
	return skills, nil
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
