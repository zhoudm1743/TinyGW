package service

import (
	"TinyGW/internal/repository"
	"TinyGW/models"
)

type CollectTaskService interface {
	Save(collectTask *models.CollectTask) error
	Delete(name string) error
	Find(name string) (models.CollectTask, error)
	FindAll() ([]models.CollectTask, error)
	List(limit int, offset int) ([]models.CollectTask, int64, error)
}

type collectTaskService struct {
	repo repository.CollectTaskRepository
}

func (s *collectTaskService) Save(collectTask *models.CollectTask) error {
	return s.repo.Save(collectTask)
}
func (s *collectTaskService) Delete(name string) error {
	return s.repo.Delete(name)
}
func (s *collectTaskService) Find(name string) (models.CollectTask, error) {
	return s.repo.Find(name)
}
func (s *collectTaskService) FindAll() ([]models.CollectTask, error) {
	return s.repo.FindAll()
}
func (s *collectTaskService) List(limit int, offset int) ([]models.CollectTask, int64, error) {
	return s.repo.List(limit, offset)
}

func NewCollectTaskService(repo repository.CollectTaskRepository) CollectTaskService {
	return &collectTaskService{repo: repo}
}
