package service

import (
	"fmt"
	"tinyGW/app/api/repository"
	"tinyGW/app/api/schemas/req"
	"tinyGW/app/api/schemas/resp"
	"tinyGW/app/models"
	"tinyGW/pkg/plugin/response"
)

type DeviceTypeService interface {
	Add(deviceType *req.DeviceTypeReq) error
	Update(deviceType *req.DeviceTypeReq) error
	Delete(name string) error
	Find(name string) (resp.DeviceTypeResp, error)
	FindAll() ([]resp.DeviceTypeResp, error)
	List(page *req.PageReq) (response.PageResp, error)
	// TODO: 增加其他方法
	AddProperties(deviceTypeName string, deviceProperty *req.DeviceProperty) error
	UpdateProperties(deviceTypeName string, devicePropertyId int, deviceProperty *req.DeviceProperty) error
	DeleteProperties(deviceTypeName string, devicePropertyId int) error
	FindProperty(deviceTypeName string, devicePropertyId int) (models.DeviceProperty, error)
	FindAllProperties(deviceTypeName string) ([]models.DeviceProperty, error)
}

type deviceTypeService struct {
	deviceRepository     repository.DeviceRepository
	deviceTypeRepository repository.DeviceTypeRepository
}

func (d deviceTypeService) AddProperties(deviceTypeName string, deviceProperty *req.DeviceProperty) error {
	var model models.DeviceProperty
	response.Copy(&model, deviceProperty)
	err := d.deviceTypeRepository.AddProperties(deviceTypeName, &model)
	if err != nil {
		return err
	}

	deviceType, _ := d.deviceTypeRepository.Find(deviceTypeName)
	d.deviceRepository.DeviceTypeChanged(deviceType)
	return err
}

func (d deviceTypeService) UpdateProperties(deviceTypeName string, devicePropertyId int, deviceProperty *req.DeviceProperty) error {
	var model models.DeviceProperty
	response.Copy(&model, deviceProperty)
	err := d.deviceTypeRepository.UpdateProperties(deviceTypeName, devicePropertyId, &model)
	if err != nil {
		return err
	}

	deviceType, _ := d.deviceTypeRepository.Find(deviceTypeName)
	d.deviceRepository.DeviceTypeChanged(deviceType)
	return err
}

func (d deviceTypeService) DeleteProperties(deviceTypeName string, devicePropertyId int) error {
	err := d.deviceTypeRepository.DeleteProperties(deviceTypeName, devicePropertyId)
	if err != nil {
		return err
	}

	deviceType, _ := d.deviceTypeRepository.Find(deviceTypeName)
	d.deviceRepository.DeviceTypeChanged(deviceType)
	return err
}

func (d deviceTypeService) FindProperty(deviceTypeName string, devicePropertyId int) (models.DeviceProperty, error) {
	return d.deviceTypeRepository.FindProperty(deviceTypeName, devicePropertyId)
}

func (d deviceTypeService) FindAllProperties(deviceTypeName string) ([]models.DeviceProperty, error) {
	return d.deviceTypeRepository.FindAllProperties(deviceTypeName)
}

func (d deviceTypeService) Add(deviceType *req.DeviceTypeReq) error {
	_, err := d.deviceTypeRepository.Find(deviceType.Name)
	if err == nil {
		return fmt.Errorf("仪表 %s <UNK>", deviceType.Name)
	}
	var dv models.DeviceType
	response.Copy(&dv, deviceType)
	return d.deviceTypeRepository.Save(&dv)
}

func (d deviceTypeService) Update(deviceType *req.DeviceTypeReq) error {
	var dv models.DeviceType
	response.Copy(&dv, deviceType)
	return d.deviceTypeRepository.Save(&dv)
}

func (d deviceTypeService) Delete(name string) error {
	return d.deviceTypeRepository.Delete(name)
}

func (d deviceTypeService) Find(name string) (resp.DeviceTypeResp, error) {
	f, err := d.deviceTypeRepository.Find(name)
	if err != nil {
		return resp.DeviceTypeResp{}, fmt.Errorf("仪表 %s 不存在", name)
	}
	var deviceType resp.DeviceTypeResp
	response.Copy(&deviceType, f)
	return deviceType, nil
}

func (d deviceTypeService) FindAll() ([]resp.DeviceTypeResp, error) {
	fs, err := d.deviceTypeRepository.FindAll()
	if err != nil {
		return nil, err
	}
	var deviceTypes []resp.DeviceTypeResp
	response.Copy(&deviceTypes, fs)
	return deviceTypes, nil
}

func (d deviceTypeService) List(page *req.PageReq) (response.PageResp, error) {
	limit := page.PageSize
	offset := page.PageSize * (page.PageNo - 1)
	list, total, err := d.deviceTypeRepository.List(limit, offset)
	if err != nil {
		return response.PageResp{}, err
	}
	var deviceTypes []req.DeviceTypeReq
	response.Copy(&deviceTypes, list)
	return response.PageResp{
		Count:    total,
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
		Lists:    list,
	}, nil
}

func NewDeviceTypeService(deviceRepository repository.DeviceRepository, deviceTypeRepository repository.DeviceTypeRepository) DeviceTypeService {
	return &deviceTypeService{
		deviceRepository:     deviceRepository,
		deviceTypeRepository: deviceTypeRepository,
	}
}
