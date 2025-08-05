package service

import (
	"TinyGW/internal/repository"
	"TinyGW/models"
)

type DeviceTypeService interface {
	Save(deviceType *models.DeviceType) error
	Delete(name string) error
	Find(name string) (models.DeviceType, error)
	FindAll() ([]models.DeviceType, error)
	List(limit int, offset int) ([]models.DeviceType, int64, error)
	AddProperties(deviceTypeName string, deviceProperty *models.DeviceProperty) error
	UpdateProperties(deviceTypeName string, devicePropertyId int, deviceProperty *models.DeviceProperty) error
	DeleteProperties(deviceTypeName string, devicePropertyId int) error
	FindProperty(deviceTypeName string, devicePropertyId int) (models.DeviceProperty, error)
	FindAllProperties(deviceTypeName string) ([]models.DeviceProperty, error)
}

type deviceTypeService struct {
	repo repository.DeviceTypeRepository
}

func (s *deviceTypeService) Save(deviceType *models.DeviceType) error {
	return s.repo.Save(deviceType)
}
func (s *deviceTypeService) Delete(name string) error {
	return s.repo.Delete(name)
}
func (s *deviceTypeService) Find(name string) (models.DeviceType, error) {
	return s.repo.Find(name)
}
func (s *deviceTypeService) FindAll() ([]models.DeviceType, error) {
	return s.repo.FindAll()
}
func (s *deviceTypeService) List(limit int, offset int) ([]models.DeviceType, int64, error) {
	return s.repo.List(limit, offset)
}
func (s *deviceTypeService) AddProperties(deviceTypeName string, deviceProperty *models.DeviceProperty) error {
	return s.repo.AddProperties(deviceTypeName, deviceProperty)
}
func (s *deviceTypeService) UpdateProperties(deviceTypeName string, devicePropertyId int, deviceProperty *models.DeviceProperty) error {
	return s.repo.UpdateProperties(deviceTypeName, devicePropertyId, deviceProperty)
}
func (s *deviceTypeService) DeleteProperties(deviceTypeName string, devicePropertyId int) error {
	return s.repo.DeleteProperties(deviceTypeName, devicePropertyId)
}
func (s *deviceTypeService) FindProperty(deviceTypeName string, devicePropertyId int) (models.DeviceProperty, error) {
	return s.repo.FindProperty(deviceTypeName, devicePropertyId)
}
func (s *deviceTypeService) FindAllProperties(deviceTypeName string) ([]models.DeviceProperty, error) {
	return s.repo.FindAllProperties(deviceTypeName)
}

func NewDeviceTypeService(repo repository.DeviceTypeRepository) DeviceTypeService {
	return &deviceTypeService{repo: repo}
}
