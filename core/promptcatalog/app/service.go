package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	repoport "github.com/shinerio/skillflow/core/promptcatalog/app/port/repository"
	"github.com/shinerio/skillflow/core/promptcatalog/domain"
)

const exportVersion = 2

type ImportPrompt struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Category    string              `json:"category"`
	Content     string              `json:"content"`
	ImageURLs   []string            `json:"imageURLs,omitempty"`
	WebLinks    []domain.PromptLink `json:"webLinks,omitempty"`
}

type ImportPreview struct {
	Creates   []ImportPrompt `json:"creates"`
	Conflicts []ImportPrompt `json:"conflicts"`
}

type exportBundle struct {
	Version    int            `json:"version"`
	ExportedAt time.Time      `json:"exportedAt"`
	Prompts    []exportPrompt `json:"prompts"`
}

type exportPrompt struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Category    string              `json:"category,omitempty"`
	Content     string              `json:"content"`
	ImageURLs   []string            `json:"imageURLs,omitempty"`
	WebLinks    []domain.PromptLink `json:"webLinks,omitempty"`
}

type Service struct {
	repo repoport.PromptRepository
}

func NewService(repo repoport.PromptRepository) *Service {
	return &Service{repo: repo}
}

func ParseWebLinksMarkdown(raw string) ([]domain.PromptLink, error) {
	return domain.ParseWebLinksMarkdown(raw)
}

func (s *Service) ListPrompts() ([]*domain.Prompt, error) {
	return s.repo.ListAll()
}

func (s *Service) GetPrompt(name string) (*domain.Prompt, error) {
	return s.repo.Get(name)
}

func (s *Service) ListPromptCategories() ([]string, error) {
	return s.repo.ListCategories()
}

func (s *Service) CreatePrompt(name, description, category, content string, imageURLs []string, webLinks []domain.PromptLink) (*domain.Prompt, error) {
	return s.repo.Create(name, description, category, content, imageURLs, webLinks)
}

func (s *Service) UpdatePrompt(originalName, name, description, category, content string, imageURLs []string, webLinks []domain.PromptLink) (*domain.Prompt, error) {
	return s.repo.Update(originalName, name, description, category, content, imageURLs, webLinks)
}

func (s *Service) DeletePrompt(name string) error {
	return s.repo.Delete(name)
}

func (s *Service) CreatePromptCategory(name string) error {
	return s.repo.CreateCategory(name)
}

func (s *Service) RenamePromptCategory(oldName, newName string) error {
	return s.repo.RenameCategory(oldName, newName)
}

func (s *Service) DeletePromptCategory(name string) error {
	return s.repo.DeleteCategory(name)
}

func (s *Service) MovePromptToCategory(name, category string) error {
	return s.repo.MoveCategory(name, category)
}

func (s *Service) ExportPromptBundle(names []string) ([]byte, error) {
	items, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	if len(names) > 0 {
		allowed := make(map[string]struct{}, len(names))
		for _, rawName := range names {
			name, err := domain.NormalizePromptName(rawName)
			if err != nil {
				return nil, err
			}
			allowed[name] = struct{}{}
		}
		filtered := make([]*domain.Prompt, 0, len(allowed))
		for _, item := range items {
			if _, ok := allowed[item.Name]; ok {
				filtered = append(filtered, item)
			}
		}
		items = filtered
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
			ImageURLs:   item.ImageURLs,
			WebLinks:    item.WebLinks,
		})
	}
	return json.MarshalIndent(bundle, "", "  ")
}

func (s *Service) PreviewPromptImport(data []byte) (*ImportPreview, error) {
	bundle, err := parseImportJSON(data)
	if err != nil {
		return nil, err
	}
	preview := &ImportPreview{
		Creates:   make([]ImportPrompt, 0, len(bundle.Prompts)),
		Conflicts: make([]ImportPrompt, 0, len(bundle.Prompts)),
	}
	createIndexByName := make(map[string]int, len(bundle.Prompts))
	for _, item := range bundle.Prompts {
		importItem, err := normalizeImportPrompt(item)
		if err != nil {
			return nil, err
		}
		if index, ok := createIndexByName[importItem.Name]; ok {
			preview.Creates[index] = importItem
			continue
		}
		if _, err := s.repo.Get(importItem.Name); err == nil {
			preview.Conflicts = append(preview.Conflicts, importItem)
		} else if err == domain.ErrPromptNotFound {
			createIndexByName[importItem.Name] = len(preview.Creates)
			preview.Creates = append(preview.Creates, importItem)
		} else {
			return nil, err
		}
	}
	return preview, nil
}

func (s *Service) ApplyPromptImport(preview *ImportPreview, overwriteNames []string) (int, error) {
	if preview == nil {
		return 0, fmt.Errorf("import preview is nil")
	}
	count := 0
	for _, item := range preview.Creates {
		if _, err := s.repo.Create(item.Name, item.Description, item.Category, item.Content, item.ImageURLs, item.WebLinks); err != nil {
			return count, err
		}
		count++
	}
	overwriteSet := make(map[string]struct{}, len(overwriteNames))
	for _, rawName := range overwriteNames {
		name, err := domain.NormalizePromptName(rawName)
		if err != nil {
			return count, err
		}
		overwriteSet[name] = struct{}{}
	}
	for _, item := range preview.Conflicts {
		if _, ok := overwriteSet[item.Name]; !ok {
			continue
		}
		if _, err := s.repo.Update(item.Name, item.Name, item.Description, item.Category, item.Content, item.ImageURLs, item.WebLinks); err != nil {
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

func normalizeImportPrompt(item exportPrompt) (ImportPrompt, error) {
	if strings.TrimSpace(item.Name) == "" {
		return ImportPrompt{}, fmt.Errorf("import prompt missing name")
	}
	if strings.TrimSpace(item.Content) == "" {
		return ImportPrompt{}, fmt.Errorf("import prompt %s missing content", item.Name)
	}
	name, err := domain.NormalizePromptName(item.Name)
	if err != nil {
		return ImportPrompt{}, err
	}
	category, err := domain.NormalizeCategoryName(item.Category)
	if err != nil {
		return ImportPrompt{}, err
	}
	imageURLs, err := domain.NormalizePromptImageURLs(item.ImageURLs)
	if err != nil {
		return ImportPrompt{}, err
	}
	webLinks, err := domain.NormalizePromptLinks(item.WebLinks)
	if err != nil {
		return ImportPrompt{}, err
	}
	return ImportPrompt{
		Name:        name,
		Description: strings.TrimSpace(item.Description),
		Category:    category,
		Content:     item.Content,
		ImageURLs:   imageURLs,
		WebLinks:    webLinks,
	}, nil
}
