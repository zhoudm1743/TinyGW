package service

import (
	"TinyGW/internal/repository"
	"TinyGW/models"
)

type CollectorService interface {
	Save(collector *models.Collector) error
	Delete(name string) error
	Find(name string) (models.Collector, error)
	FindAll() ([]models.Collector, error)
	List(limit int, offset int) ([]models.Collector, int64, error)
}

type collectorService struct {
	repo repository.CollectorRepository
}

func (s *collectorService) Save(collector *models.Collector) error {
	return s.repo.Save(collector)
}
func (s *collectorService) Delete(name string) error {
	return s.repo.Delete(name)
}
func (s *collectorService) Find(name string) (models.Collector, error) {
	return s.repo.Find(name)
}
func (s *collectorService) FindAll() ([]models.Collector, error) {
	return s.repo.FindAll()
}
func (s *collectorService) List(limit int, offset int) ([]models.Collector, int64, error) {
	return s.repo.List(limit, offset)
}

func NewCollectorService(repo repository.CollectorRepository) CollectorService {
	return &collectorService{repo: repo}
}
