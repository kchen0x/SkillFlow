package app

import (
	repoport "github.com/shinerio/skillflow/core/skillcatalog/app/port/repository"
	"github.com/shinerio/skillflow/core/skillcatalog/domain"
)

type Service struct {
	repo repoport.InstalledSkillRepository
}

func NewService(repo repoport.InstalledSkillRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateCategory(name string) error {
	return s.repo.CreateCategory(name)
}

func (s *Service) ListCategories() ([]string, error) {
	return s.repo.ListCategories()
}

func (s *Service) Import(srcDir, category string, source domain.SourceType, sourceURL, sourceSubPath string) (*domain.InstalledSkill, error) {
	return s.repo.Import(srcDir, category, source, sourceURL, sourceSubPath)
}

func (s *Service) Get(id string) (*domain.InstalledSkill, error) {
	return s.repo.Get(id)
}

func (s *Service) ListAll() ([]*domain.InstalledSkill, error) {
	return s.repo.ListAll()
}

func (s *Service) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *Service) MoveCategory(id, newCategory string) error {
	return s.repo.MoveCategory(id, newCategory)
}

func (s *Service) UpdateMeta(sk *domain.InstalledSkill) error {
	return s.repo.UpdateMeta(sk)
}

func (s *Service) SaveMeta(sk *domain.InstalledSkill) error {
	return s.repo.SaveMeta(sk)
}

func (s *Service) RenameCategory(oldName, newName string) error {
	return s.repo.RenameCategory(oldName, newName)
}

func (s *Service) DeleteCategory(name string) error {
	return s.repo.DeleteCategory(name)
}

func (s *Service) OverwriteFromDir(id, srcDir string) error {
	return s.repo.OverwriteFromDir(id, srcDir)
}
