package service

import (
	"TinyGW/internal/repository"
	"TinyGW/models"
)

type DeviceService interface {
	Save(device *models.Device) error
	Delete(name string) error
	Find(name string) (models.Device, error)
	FindAll() ([]models.Device, error)
	List(limit int, offset int) ([]models.Device, int64, error)
	FindByAddr(addr string) (models.Device, error)
	CollectorIsUsed(collector *models.Collector) bool
	DeviceTypeIsUsed(deviceType *models.DeviceType) bool
	CollectorChanged(collector models.Collector)
	DeviceTypeChanged(deviceType models.DeviceType)
}

type deviceService struct {
	repo repository.DeviceRepository
}

func (s *deviceService) Save(device *models.Device) error {
	return s.repo.Save(device)
}
func (s *deviceService) Delete(name string) error {
	return s.repo.Delete(name)
}
func (s *deviceService) Find(name string) (models.Device, error) {
	return s.repo.Find(name)
}
func (s *deviceService) FindAll() ([]models.Device, error) {
	return s.repo.FindAll()
}
func (s *deviceService) List(limit int, offset int) ([]models.Device, int64, error) {
	return s.repo.List(limit, offset)
}
func (s *deviceService) FindByAddr(addr string) (models.Device, error) {
	return s.repo.FindByAddr(addr)
}
func (s *deviceService) CollectorIsUsed(collector *models.Collector) bool {
	return s.repo.CollectorIsUsed(collector)
}
func (s *deviceService) DeviceTypeIsUsed(deviceType *models.DeviceType) bool {
	return s.repo.DeviceTypeIsUsed(deviceType)
}
func (s *deviceService) CollectorChanged(collector models.Collector) {
	s.repo.CollectorChanged(collector)
}
func (s *deviceService) DeviceTypeChanged(deviceType models.DeviceType) {
	s.repo.DeviceTypeChanged(deviceType)
}

func NewDeviceService(repo repository.DeviceRepository) DeviceService {
	return &deviceService{repo: repo}
}
