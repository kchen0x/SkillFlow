package repository

import "github.com/shinerio/skillflow/core/skillcatalog/domain"

type InstalledSkillRepository interface {
	CreateCategory(name string) error
	ListCategories() ([]string, error)
	Import(srcDir, category string, source domain.SourceType, sourceURL, sourceSubPath string) (*domain.InstalledSkill, error)
	Get(id string) (*domain.InstalledSkill, error)
	ListAll() ([]*domain.InstalledSkill, error)
	Delete(id string) error
	MoveCategory(id, newCategory string) error
	UpdateMeta(sk *domain.InstalledSkill) error
	SaveMeta(sk *domain.InstalledSkill) error
	RenameCategory(oldName, newName string) error
	DeleteCategory(name string) error
	OverwriteFromDir(id, srcDir string) error
}
