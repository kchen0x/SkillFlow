package skill

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shinerio/skillflow/core/pathutil"
)

var ErrSkillExists = errors.New("skill already exists in target location")
var ErrSkillNotFound = errors.New("skill not found")
var ErrCategoryNotEmpty = errors.New("category not empty")

type Storage struct {
	root     string
	metaDir  string
	syncRoot string
}

func NewStorage(root string) *Storage {
	cleanRoot := filepath.Clean(root)
	syncRoot := filepath.Dir(cleanRoot)
	return &Storage{
		root:     cleanRoot,
		metaDir:  filepath.Join(syncRoot, "meta"),
		syncRoot: syncRoot,
	}
}

func (s *Storage) CreateCategory(name string) error {
	return os.MkdirAll(filepath.Join(s.root, name), 0755)
}

func (s *Storage) ListCategories() ([]string, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var cats []string
	for _, e := range entries {
		if e.IsDir() {
			cats = append(cats, e.Name())
		}
	}
	return cats, nil
}

func (s *Storage) Import(srcDir, category string, source SourceType, sourceURL, sourceSubPath string) (*Skill, error) {
	name := filepath.Base(srcDir)
	targetDir := filepath.Join(s.root, category, name)
	if _, err := os.Stat(targetDir); err == nil {
		return nil, ErrSkillExists
	}
	if err := copyDir(srcDir, targetDir); err != nil {
		return nil, err
	}
	sk := &Skill{
		ID:            uuid.New().String(),
		Name:          name,
		Path:          targetDir,
		Category:      category,
		Source:        source,
		SourceURL:     sourceURL,
		SourceSubPath: sourceSubPath,
		InstalledAt:   time.Now(),
		UpdatedAt:     time.Now(),
	}
	return sk, s.saveMeta(sk)
}

func (s *Storage) Get(id string) (*Skill, error) {
	skills, err := s.ListAll()
	if err != nil {
		return nil, err
	}
	for _, sk := range skills {
		if sk.ID == id {
			return sk, nil
		}
	}
	return nil, ErrSkillNotFound
}

func (s *Storage) ListAll() ([]*Skill, error) {
	if err := os.MkdirAll(s.metaDir, 0755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.metaDir)
	if err != nil {
		return nil, err
	}
	var skills []*Skill
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.metaDir, e.Name()))
		if err != nil {
			continue
		}
		var sk Skill
		if err := json.Unmarshal(data, &sk); err == nil {
			if s.resolveLoadedSkillPath(&sk) {
				_ = s.saveMeta(&sk)
			}
			skills = append(skills, &sk)
		}
	}
	return skills, nil
}

func (s *Storage) Delete(id string) error {
	sk, err := s.Get(id)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(sk.Path); err != nil {
		return err
	}
	return os.Remove(filepath.Join(s.metaDir, id+".json"))
}

func (s *Storage) MoveCategory(id, newCategory string) error {
	sk, err := s.Get(id)
	if err != nil {
		return err
	}
	newPath := filepath.Join(s.root, newCategory, sk.Name)
	if err := os.MkdirAll(filepath.Join(s.root, newCategory), 0755); err != nil {
		return err
	}
	if err := os.Rename(sk.Path, newPath); err != nil {
		return err
	}
	sk.Path = newPath
	sk.Category = newCategory
	sk.UpdatedAt = time.Now()
	return s.saveMeta(sk)
}

func (s *Storage) UpdateMeta(sk *Skill) error {
	sk.UpdatedAt = time.Now()
	return s.saveMeta(sk)
}

func (s *Storage) SaveMeta(sk *Skill) error {
	return s.saveMeta(sk)
}

func (s *Storage) RenameCategory(oldName, newName string) error {
	oldPath := filepath.Join(s.root, oldName)
	newPath := filepath.Join(s.root, newName)
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	skills, err := s.ListAll()
	if err != nil {
		return err
	}
	for _, sk := range skills {
		if sk.Category == oldName {
			sk.Category = newName
			sk.Path = filepath.Join(newPath, sk.Name)
			sk.UpdatedAt = time.Now()
			if err := s.saveMeta(sk); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Storage) DeleteCategory(name string) error {
	skills, err := s.ListAll()
	if err != nil {
		return err
	}
	for _, sk := range skills {
		if sk.Category == name {
			return ErrCategoryNotEmpty
		}
	}
	return os.Remove(filepath.Join(s.root, name))
}

// OverwriteFromDir replaces an existing skill's directory contents from srcDir, used for updates.
func (s *Storage) OverwriteFromDir(id, srcDir string) error {
	sk, err := s.Get(id)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(sk.Path); err != nil {
		return err
	}
	return copyDir(srcDir, sk.Path)
}

func (s *Storage) saveMeta(sk *Skill) error {
	if err := os.MkdirAll(s.metaDir, 0755); err != nil {
		return err
	}
	snapshot := *sk
	snapshot.Path = pathutil.StorePath(s.syncRoot, sk.Path, s.skillPath(sk.Category, sk.Name))
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.metaDir, sk.ID+".json"), data, 0644)
}

func (s *Storage) resolveLoadedSkillPath(sk *Skill) bool {
	expected := s.skillPath(sk.Category, sk.Name)
	resolved, needsMigration := pathutil.ResolveStoredPath(s.syncRoot, sk.Path, expected)
	sk.Path = resolved
	return needsMigration
}

func (s *Storage) skillPath(category, name string) string {
	if strings.TrimSpace(name) == "" {
		return ""
	}
	if strings.TrimSpace(category) == "" {
		return filepath.Join(s.root, name)
	}
	return filepath.Join(s.root, category, name)
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
