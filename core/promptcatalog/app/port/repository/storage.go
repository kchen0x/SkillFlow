package repository

import "github.com/shinerio/skillflow/core/promptcatalog/domain"

type PromptRepository interface {
	ListCategories() ([]string, error)
	CreateCategory(name string) error
	RenameCategory(oldName, newName string) error
	DeleteCategory(name string) error
	ListAll() ([]*domain.Prompt, error)
	Get(name string) (*domain.Prompt, error)
	Create(name, description, category, content string, imageURLs []string, webLinks []domain.PromptLink) (*domain.Prompt, error)
	Update(originalName, name, description, category, content string, imageURLs []string, webLinks []domain.PromptLink) (*domain.Prompt, error)
	Delete(name string) error
	MoveCategory(name, category string) error
}
