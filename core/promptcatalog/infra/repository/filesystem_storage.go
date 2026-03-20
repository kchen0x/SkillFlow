package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/shinerio/skillflow/core/promptcatalog/domain"
)

type FilesystemStorage struct {
	root string
}

type promptMeta struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	ImageURLs   []string            `json:"imageURLs,omitempty"`
	WebLinks    []domain.PromptLink `json:"webLinks,omitempty"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}

func NewFilesystemStorage(root string) *FilesystemStorage {
	return &FilesystemStorage{root: filepath.Clean(root)}
}

func (s *FilesystemStorage) ListCategories() ([]string, error) {
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
		if fileExists(filepath.Join(dir, domain.FileName)) {
			continue
		}
		categories = append(categories, entry.Name())
	}
	sort.Strings(categories)
	return categories, nil
}

func (s *FilesystemStorage) CreateCategory(name string) error {
	category, err := domain.NormalizeCategoryName(name)
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(s.root, category), 0755)
}

func (s *FilesystemStorage) RenameCategory(oldName, newName string) error {
	if err := s.migrateLegacyLayout(); err != nil {
		return err
	}
	oldCategory, err := domain.NormalizeCategoryName(oldName)
	if err != nil {
		return err
	}
	newCategory, err := domain.NormalizeCategoryName(newName)
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

func (s *FilesystemStorage) DeleteCategory(name string) error {
	if err := s.migrateLegacyLayout(); err != nil {
		return err
	}
	category, err := domain.NormalizeCategoryName(name)
	if err != nil {
		return err
	}
	prompts, err := s.ListAll()
	if err != nil {
		return err
	}
	for _, item := range prompts {
		if item.Category == category {
			return domain.ErrCategoryNotEmpty
		}
	}
	return os.Remove(filepath.Join(s.root, category))
}

func (s *FilesystemStorage) ListAll() ([]*domain.Prompt, error) {
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
	prompts := make([]*domain.Prompt, 0)
	for _, categoryEntry := range entries {
		if !categoryEntry.IsDir() {
			continue
		}
		categoryName := categoryEntry.Name()
		categoryPath := filepath.Join(s.root, categoryName)
		if fileExists(filepath.Join(categoryPath, domain.FileName)) {
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
				if errors.Is(err, os.ErrNotExist) || errors.Is(err, domain.ErrPromptNotFound) {
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

func (s *FilesystemStorage) Get(name string) (*domain.Prompt, error) {
	promptName, err := domain.NormalizePromptName(name)
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
	return nil, domain.ErrPromptNotFound
}

func (s *FilesystemStorage) Create(name, description, category, content string, imageURLs []string, webLinks []domain.PromptLink) (*domain.Prompt, error) {
	if err := s.migrateLegacyLayout(); err != nil {
		return nil, err
	}
	promptName, err := domain.NormalizePromptName(name)
	if err != nil {
		return nil, err
	}
	promptCategory, err := domain.NormalizeCategoryName(category)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(content) == "" {
		return nil, domain.ErrEmptyContent
	}
	normalizedImages, err := domain.NormalizePromptImageURLs(imageURLs)
	if err != nil {
		return nil, err
	}
	normalizedLinks, err := domain.NormalizePromptLinks(webLinks)
	if err != nil {
		return nil, err
	}
	if _, err := s.Get(promptName); err == nil {
		return nil, domain.ErrPromptExists
	} else if !errors.Is(err, domain.ErrPromptNotFound) {
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
		ImageURLs:   normalizedImages,
		WebLinks:    normalizedLinks,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.writePromptFiles(promptDir, meta, content); err != nil {
		return nil, err
	}
	return s.readPromptDir(promptCategory, promptDir)
}

func (s *FilesystemStorage) Update(originalName, name, description, category, content string, imageURLs []string, webLinks []domain.PromptLink) (*domain.Prompt, error) {
	if err := s.migrateLegacyLayout(); err != nil {
		return nil, err
	}
	current, err := s.Get(originalName)
	if err != nil {
		return nil, err
	}
	promptName, err := domain.NormalizePromptName(name)
	if err != nil {
		return nil, err
	}
	promptCategory, err := domain.NormalizeCategoryName(category)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(content) == "" {
		return nil, domain.ErrEmptyContent
	}
	normalizedImages, err := domain.NormalizePromptImageURLs(imageURLs)
	if err != nil {
		return nil, err
	}
	normalizedLinks, err := domain.NormalizePromptLinks(webLinks)
	if err != nil {
		return nil, err
	}
	if current.Name != promptName {
		if _, err := s.Get(promptName); err == nil {
			return nil, domain.ErrPromptExists
		} else if !errors.Is(err, domain.ErrPromptNotFound) {
			return nil, err
		}
	}
	currentDir := current.Path
	targetDir := s.promptDir(promptCategory, promptName)
	if currentDir != targetDir {
		targetExists, sameTarget, err := domain.CompareFilesystemEntries(currentDir, targetDir)
		if err != nil {
			return nil, err
		}
		if targetExists && !sameTarget {
			return nil, domain.ErrPromptExists
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
	meta.ImageURLs = normalizedImages
	meta.WebLinks = normalizedLinks
	meta.UpdatedAt = time.Now()
	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = meta.UpdatedAt
	}
	if err := s.writePromptFiles(currentDir, meta, content); err != nil {
		return nil, err
	}
	return s.readPromptDir(promptCategory, currentDir)
}

func (s *FilesystemStorage) Delete(name string) error {
	current, err := s.Get(name)
	if err != nil {
		return err
	}
	return os.RemoveAll(current.Path)
}

func (s *FilesystemStorage) MoveCategory(name, category string) error {
	current, err := s.Get(name)
	if err != nil {
		return err
	}
	_, err = s.Update(current.Name, current.Name, current.Description, category, current.Content, current.ImageURLs, current.WebLinks)
	return err
}

func (s *FilesystemStorage) migrateLegacyLayout() error {
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
		if fileExists(filepath.Join(dir, domain.FileName)) {
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
	defaultDir := filepath.Join(s.root, domain.DefaultCategoryName)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		legacyDir := filepath.Join(s.root, entry.Name())
		if !fileExists(filepath.Join(legacyDir, domain.FileName)) {
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
			name = domain.NextAvailableName(entry.Name(), existingNames)
		}
		existingNames[name] = struct{}{}
		targetDir := filepath.Join(defaultDir, name)
		if err := os.Rename(legacyDir, targetDir); err != nil {
			return err
		}
		if _, _, err := s.readPromptMeta(targetDir); errors.Is(err, os.ErrNotExist) {
			contentBytes, readErr := os.ReadFile(filepath.Join(targetDir, domain.FileName))
			if readErr != nil {
				return readErr
			}
			now := time.Now()
			if err := s.writePromptFiles(targetDir, promptMeta{Name: name, CreatedAt: now, UpdatedAt: now}, string(contentBytes)); err != nil {
				return err
			}
		} else if err != nil && !errors.Is(err, domain.ErrPromptNotFound) {
			return err
		}
	}
	return nil
}

func (s *FilesystemStorage) readPromptDir(category string, dir string) (*domain.Prompt, error) {
	meta, content, err := s.readPromptMeta(dir)
	if err != nil {
		return nil, err
	}
	categoryName, err := domain.NormalizeCategoryName(category)
	if err != nil {
		return nil, err
	}
	return &domain.Prompt{
		Name:        meta.Name,
		Description: meta.Description,
		Category:    categoryName,
		Path:        dir,
		FilePath:    filepath.Join(dir, domain.FileName),
		Content:     content,
		ImageURLs:   meta.ImageURLs,
		WebLinks:    meta.WebLinks,
		CreatedAt:   meta.CreatedAt,
		UpdatedAt:   meta.UpdatedAt,
	}, nil
}

func (s *FilesystemStorage) readPromptMeta(dir string) (promptMeta, string, error) {
	contentPath := filepath.Join(dir, domain.FileName)
	contentBytes, err := os.ReadFile(contentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return promptMeta{}, "", domain.ErrPromptNotFound
		}
		return promptMeta{}, "", err
	}
	metaPath := filepath.Join(dir, domain.MetaFileName)
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
	normalizedImages, err := domain.NormalizePromptImageURLs(meta.ImageURLs)
	if err != nil {
		return promptMeta{}, "", err
	}
	normalizedLinks, err := domain.NormalizePromptLinks(meta.WebLinks)
	if err != nil {
		return promptMeta{}, "", err
	}
	meta.ImageURLs = normalizedImages
	meta.WebLinks = normalizedLinks
	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = time.Now()
	}
	if meta.UpdatedAt.IsZero() {
		meta.UpdatedAt = meta.CreatedAt
	}
	return meta, string(contentBytes), nil
}

func (s *FilesystemStorage) writePromptFiles(dir string, meta promptMeta, content string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, domain.FileName), []byte(content), 0644); err != nil {
		return err
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, domain.MetaFileName), data, 0644)
}

func (s *FilesystemStorage) promptDir(category, name string) string {
	return domain.PromptDir(s.root, category, name)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
