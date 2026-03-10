package prompt

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	DefaultCategoryName = "Default"
	FileName            = "system.md"
	MetaFileName        = "prompt.json"
	exportVersion       = 1
)

var (
	ErrPromptNotFound   = errors.New("prompt not found")
	ErrPromptExists     = errors.New("prompt already exists")
	ErrEmptyContent     = errors.New("prompt content is empty")
	ErrInvalidName      = errors.New("invalid prompt name")
	ErrCategoryNotEmpty = errors.New("category not empty")
)

type Prompt struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Category    string    `json:"category"`
	Path        string    `json:"path"`
	FilePath    string    `json:"filePath"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Storage struct {
	root string
}

type promptMeta struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type exportBundle struct {
	Version    int            `json:"version"`
	ExportedAt time.Time      `json:"exportedAt"`
	Prompts    []exportPrompt `json:"prompts"`
}

type exportPrompt struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Content     string `json:"content"`
}

func NewStorage(root string) *Storage {
	return &Storage{root: filepath.Clean(root)}
}

func (s *Storage) ListCategories() ([]string, error) {
	if err := s.migrateLegacyLayout(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	categories := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(s.root, entry.Name())
		if fileExists(filepath.Join(dir, FileName)) {
			continue
		}
		categories = append(categories, entry.Name())
	}
	sort.Strings(categories)
	return categories, nil
}

func (s *Storage) CreateCategory(name string) error {
	category, err := normalizeCategoryName(name)
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(s.root, category), 0755)
}

func (s *Storage) RenameCategory(oldName, newName string) error {
	if err := s.migrateLegacyLayout(); err != nil {
		return err
	}
	oldCategory, err := normalizeCategoryName(oldName)
	if err != nil {
		return err
	}
	newCategory, err := normalizeCategoryName(newName)
	if err != nil {
		return err
	}
	if oldCategory == newCategory {
		return nil
	}
	oldPath := filepath.Join(s.root, oldCategory)
	newPath := filepath.Join(s.root, newCategory)
	if _, err := os.Stat(newPath); err == nil {
		entries, readErr := os.ReadDir(newPath)
		if readErr != nil {
			return readErr
		}
		if len(entries) > 0 {
			return fmt.Errorf("category already exists: %s", newCategory)
		}
	}
	return os.Rename(oldPath, newPath)
}

func (s *Storage) DeleteCategory(name string) error {
	if err := s.migrateLegacyLayout(); err != nil {
		return err
	}
	category, err := normalizeCategoryName(name)
	if err != nil {
		return err
	}
	prompts, err := s.ListAll()
	if err != nil {
		return err
	}
	for _, item := range prompts {
		if item.Category == category {
			return ErrCategoryNotEmpty
		}
	}
	return os.Remove(filepath.Join(s.root, category))
}

func (s *Storage) ListAll() ([]*Prompt, error) {
	if err := s.migrateLegacyLayout(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	prompts := make([]*Prompt, 0)
	for _, categoryEntry := range entries {
		if !categoryEntry.IsDir() {
			continue
		}
		categoryName := categoryEntry.Name()
		categoryPath := filepath.Join(s.root, categoryName)
		if fileExists(filepath.Join(categoryPath, FileName)) {
			continue
		}
		promptEntries, err := os.ReadDir(categoryPath)
		if err != nil {
			return nil, err
		}
		for _, promptEntry := range promptEntries {
			if !promptEntry.IsDir() {
				continue
			}
			item, err := s.readPromptDir(categoryName, filepath.Join(categoryPath, promptEntry.Name()))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) || errors.Is(err, ErrPromptNotFound) {
					continue
				}
				return nil, err
			}
			prompts = append(prompts, item)
		}
	}
	sort.Slice(prompts, func(i, j int) bool {
		left := prompts[i]
		right := prompts[j]
		if left.Category != right.Category {
			return left.Category < right.Category
		}
		return strings.ToLower(left.Name) < strings.ToLower(right.Name)
	})
	return prompts, nil
}

func (s *Storage) Get(name string) (*Prompt, error) {
	promptName, err := normalizePromptName(name)
	if err != nil {
		return nil, err
	}
	items, err := s.ListAll()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.Name == promptName {
			return item, nil
		}
	}
	return nil, ErrPromptNotFound
}

func (s *Storage) Create(name, description, category, content string) (*Prompt, error) {
	if err := s.migrateLegacyLayout(); err != nil {
		return nil, err
	}
	promptName, err := normalizePromptName(name)
	if err != nil {
		return nil, err
	}
	promptCategory, err := normalizeCategoryName(category)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(content) == "" {
		return nil, ErrEmptyContent
	}
	if _, err := s.Get(promptName); err == nil {
		return nil, ErrPromptExists
	} else if !errors.Is(err, ErrPromptNotFound) {
		return nil, err
	}
	promptDir := s.promptDir(promptCategory, promptName)
	if err := os.MkdirAll(promptDir, 0755); err != nil {
		return nil, err
	}
	now := time.Now()
	meta := promptMeta{
		Name:        promptName,
		Description: strings.TrimSpace(description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.writePromptFiles(promptDir, meta, content); err != nil {
		return nil, err
	}
	return s.readPromptDir(promptCategory, promptDir)
}

func (s *Storage) Update(originalName, name, description, category, content string) (*Prompt, error) {
	if err := s.migrateLegacyLayout(); err != nil {
		return nil, err
	}
	current, err := s.Get(originalName)
	if err != nil {
		return nil, err
	}
	promptName, err := normalizePromptName(name)
	if err != nil {
		return nil, err
	}
	promptCategory, err := normalizeCategoryName(category)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(content) == "" {
		return nil, ErrEmptyContent
	}
	if current.Name != promptName {
		if _, err := s.Get(promptName); err == nil {
			return nil, ErrPromptExists
		} else if !errors.Is(err, ErrPromptNotFound) {
			return nil, err
		}
	}
	currentDir := current.Path
	targetDir := s.promptDir(promptCategory, promptName)
	if currentDir != targetDir {
		targetExists, sameTarget, err := compareFilesystemEntries(currentDir, targetDir)
		if err != nil {
			return nil, err
		}
		if targetExists && !sameTarget {
			return nil, ErrPromptExists
		}
		if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
			return nil, err
		}
		if err := os.Rename(currentDir, targetDir); err != nil {
			return nil, err
		}
		currentDir = targetDir
	}
	meta, _, err := s.readPromptMeta(currentDir)
	if err != nil {
		return nil, err
	}
	meta.Name = promptName
	meta.Description = strings.TrimSpace(description)
	meta.UpdatedAt = time.Now()
	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = meta.UpdatedAt
	}
	if err := s.writePromptFiles(currentDir, meta, content); err != nil {
		return nil, err
	}
	return s.readPromptDir(promptCategory, currentDir)
}

func (s *Storage) Delete(name string) error {
	current, err := s.Get(name)
	if err != nil {
		return err
	}
	return os.RemoveAll(current.Path)
}

func (s *Storage) MoveCategory(name, category string) error {
	current, err := s.Get(name)
	if err != nil {
		return err
	}
	_, err = s.Update(current.Name, current.Name, current.Description, category, current.Content)
	return err
}

func (s *Storage) ExportJSON() ([]byte, error) {
	items, err := s.ListAll()
	if err != nil {
		return nil, err
	}
	bundle := exportBundle{
		Version:    exportVersion,
		ExportedAt: time.Now(),
		Prompts:    make([]exportPrompt, 0, len(items)),
	}
	for _, item := range items {
		bundle.Prompts = append(bundle.Prompts, exportPrompt{
			Name:        item.Name,
			Description: item.Description,
			Category:    item.Category,
			Content:     item.Content,
		})
	}
	return json.MarshalIndent(bundle, "", "  ")
}

func (s *Storage) ImportJSON(data []byte) (int, error) {
	bundle, err := parseImportJSON(data)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, item := range bundle.Prompts {
		if strings.TrimSpace(item.Name) == "" {
			return count, fmt.Errorf("import prompt missing name")
		}
		if strings.TrimSpace(item.Content) == "" {
			return count, fmt.Errorf("import prompt %s missing content", item.Name)
		}
		if _, err := s.Get(item.Name); err == nil {
			if _, err := s.Update(item.Name, item.Name, item.Description, item.Category, item.Content); err != nil {
				return count, err
			}
		} else if errors.Is(err, ErrPromptNotFound) {
			if _, err := s.Create(item.Name, item.Description, item.Category, item.Content); err != nil {
				return count, err
			}
		} else {
			return count, err
		}
		count++
	}
	return count, nil
}

func parseImportJSON(data []byte) (*exportBundle, error) {
	var bundle exportBundle
	if err := json.Unmarshal(data, &bundle); err == nil && len(bundle.Prompts) > 0 {
		return &bundle, nil
	}
	var items []exportPrompt
	if err := json.Unmarshal(data, &items); err == nil && len(items) > 0 {
		return &exportBundle{Version: exportVersion, ExportedAt: time.Now(), Prompts: items}, nil
	}
	return nil, fmt.Errorf("invalid prompt import file")
}

func (s *Storage) migrateLegacyLayout() error {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	existingNames := map[string]struct{}{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(s.root, entry.Name())
		if fileExists(filepath.Join(dir, FileName)) {
			continue
		}
		promptEntries, readErr := os.ReadDir(dir)
		if readErr != nil {
			return readErr
		}
		for _, promptEntry := range promptEntries {
			if promptEntry.IsDir() {
				existingNames[promptEntry.Name()] = struct{}{}
			}
		}
	}
	defaultDir := filepath.Join(s.root, DefaultCategoryName)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		legacyDir := filepath.Join(s.root, entry.Name())
		if !fileExists(filepath.Join(legacyDir, FileName)) {
			continue
		}
		if err := os.MkdirAll(defaultDir, 0755); err != nil {
			return err
		}
		name := entry.Name()
		for {
			if _, exists := existingNames[name]; !exists {
				break
			}
			name = nextAvailableName(entry.Name(), existingNames)
		}
		existingNames[name] = struct{}{}
		targetDir := filepath.Join(defaultDir, name)
		if err := os.Rename(legacyDir, targetDir); err != nil {
			return err
		}
		if _, _, err := s.readPromptMeta(targetDir); errors.Is(err, os.ErrNotExist) {
			contentBytes, readErr := os.ReadFile(filepath.Join(targetDir, FileName))
			if readErr != nil {
				return readErr
			}
			now := time.Now()
			if err := s.writePromptFiles(targetDir, promptMeta{Name: name, CreatedAt: now, UpdatedAt: now}, string(contentBytes)); err != nil {
				return err
			}
		} else if err != nil && !errors.Is(err, ErrPromptNotFound) {
			return err
		}
	}
	return nil
}

func (s *Storage) readPromptDir(category string, dir string) (*Prompt, error) {
	meta, content, err := s.readPromptMeta(dir)
	if err != nil {
		return nil, err
	}
	categoryName, err := normalizeCategoryName(category)
	if err != nil {
		return nil, err
	}
	return &Prompt{
		Name:        meta.Name,
		Description: meta.Description,
		Category:    categoryName,
		Path:        dir,
		FilePath:    filepath.Join(dir, FileName),
		Content:     content,
		CreatedAt:   meta.CreatedAt,
		UpdatedAt:   meta.UpdatedAt,
	}, nil
}

func (s *Storage) readPromptMeta(dir string) (promptMeta, string, error) {
	contentPath := filepath.Join(dir, FileName)
	contentBytes, err := os.ReadFile(contentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return promptMeta{}, "", ErrPromptNotFound
		}
		return promptMeta{}, "", err
	}
	metaPath := filepath.Join(dir, MetaFileName)
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return promptMeta{}, "", err
		}
		info, statErr := os.Stat(contentPath)
		if statErr != nil {
			return promptMeta{}, "", statErr
		}
		meta := promptMeta{
			Name:      filepath.Base(dir),
			CreatedAt: info.ModTime(),
			UpdatedAt: info.ModTime(),
		}
		if err := s.writePromptFiles(dir, meta, string(contentBytes)); err != nil {
			return promptMeta{}, "", err
		}
		return meta, string(contentBytes), nil
	}
	var meta promptMeta
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return promptMeta{}, "", err
	}
	if meta.Name == "" {
		meta.Name = filepath.Base(dir)
	}
	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = time.Now()
	}
	if meta.UpdatedAt.IsZero() {
		meta.UpdatedAt = meta.CreatedAt
	}
	return meta, string(contentBytes), nil
}

func (s *Storage) writePromptFiles(dir string, meta promptMeta, content string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(content), 0644); err != nil {
		return err
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, MetaFileName), data, 0644)
}

func (s *Storage) promptDir(category, name string) string {
	return filepath.Join(s.root, category, name)
}

func normalizeCategoryName(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || strings.EqualFold(trimmed, DefaultCategoryName) {
		return DefaultCategoryName, nil
	}
	if err := validatePathSegment(trimmed); err != nil {
		return "", err
	}
	return trimmed, nil
}

func normalizePromptName(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrInvalidName
	}
	if err := validatePathSegment(trimmed); err != nil {
		return "", err
	}
	return trimmed, nil
}

func validatePathSegment(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "." || trimmed == ".." {
		return ErrInvalidName
	}
	if strings.Contains(trimmed, "/") || strings.Contains(trimmed, string(filepath.Separator)) {
		return ErrInvalidName
	}
	invalidChars := `<>:"\\|?*`
	if strings.ContainsAny(trimmed, invalidChars) {
		return ErrInvalidName
	}
	if strings.HasSuffix(trimmed, ".") || strings.HasSuffix(trimmed, " ") {
		return ErrInvalidName
	}
	switch strings.ToUpper(trimmed) {
	case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return ErrInvalidName
	}
	return nil
}

func nextAvailableName(base string, existing map[string]struct{}) string {
	candidate := base
	index := 2
	for {
		candidate = fmt.Sprintf("%s-%d", base, index)
		if _, exists := existing[candidate]; !exists {
			return candidate
		}
		index++
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func compareFilesystemEntries(currentPath, targetPath string) (bool, bool, error) {
	currentInfo, err := os.Stat(currentPath)
	if err != nil {
		return false, false, err
	}
	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, os.SameFile(currentInfo, targetInfo), nil
}
