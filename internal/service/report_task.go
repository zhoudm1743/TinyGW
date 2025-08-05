package service

import (
	"TinyGW/internal/repository"
	"TinyGW/models"
)

type ReportTaskService interface {
	Save(reportTask *models.ReportTask) error
	Delete(name string) error
	Find(name string) (models.ReportTask, error)
	FindAll() ([]models.ReportTask, error)
	List(limit int, offset int) ([]models.ReportTask, int64, error)
}

type reportTaskService struct {
	repo repository.ReportTaskRepository
}

func (s *reportTaskService) Save(reportTask *models.ReportTask) error {
	return s.repo.Save(reportTask)
}
func (s *reportTaskService) Delete(name string) error {
	return s.repo.Delete(name)
}
func (s *reportTaskService) Find(name string) (models.ReportTask, error) {
	return s.repo.Find(name)
}
func (s *reportTaskService) FindAll() ([]models.ReportTask, error) {
	return s.repo.FindAll()
}
func (s *reportTaskService) List(limit int, offset int) ([]models.ReportTask, int64, error) {
	return s.repo.List(limit, offset)
}

func NewReportTaskService(repo repository.ReportTaskRepository) ReportTaskService {
	return &reportTaskService{repo: repo}
}
