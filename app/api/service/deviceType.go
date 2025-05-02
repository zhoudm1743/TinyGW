package service

import (
	"fmt"
	"tinyGW/app/api/repository"
	"tinyGW/app/api/schemas/req"
	"tinyGW/app/models"
	"tinyGW/pkg/plugin/response"
)

type DeviceService interface {
	Add(device *req.DeviceReq) error
	Update(device *req.DeviceReq) error
	Delete(name string) error
	Find(name string) (req.DeviceReq, error)
	FindAll() ([]req.DeviceReq, error)
	List(page *req.PageReq) (response.PageResp, error)
}

type deviceService struct {
	deviceRepo repository.DeviceRepository
}

func (d deviceService) Add(device *req.DeviceReq) error {
	_, err := d.deviceRepo.Find(device.Name)
	if err == nil {
		return fmt.Errorf("仪表 %s <UNK>", device.Name)
	}
	var dv models.Device
	response.Copy(&dv, device)
	return d.deviceRepo.Save(&dv)
}

func (d deviceService) Update(device *req.DeviceReq) error {
	var dv models.Device
	response.Copy(&dv, device)
	return d.deviceRepo.Save(&dv)
}

func (d deviceService) Delete(name string) error {
	return d.deviceRepo.Delete(name)
}

func (d deviceService) Find(name string) (req.DeviceReq, error) {
	f, err := d.deviceRepo.Find(name)
	if err != nil {
		return req.DeviceReq{}, fmt.Errorf("仪表 %s 不存在", name)
	}
	var device req.DeviceReq
	response.Copy(&device, f)
	return device, nil
}

func (d deviceService) FindAll() ([]req.DeviceReq, error) {
	fs, err := d.deviceRepo.FindAll()
	if err != nil {
		return nil, err
	}
	var devices []req.DeviceReq
	response.Copy(&devices, fs)
	return devices, nil
}

func (d deviceService) List(page *req.PageReq) (response.PageResp, error) {
	limit := page.PageSize
	offset := page.PageSize * (page.PageNo - 1)
	list, total, err := d.deviceRepo.List(offset, limit)
	if err != nil {
		return response.PageResp{}, err
	}
	var devices []req.DeviceReq
	response.Copy(&devices, list)
	return response.PageResp{
		Count:    total,
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
		Lists:    list,
	}, nil
}

func NewDeviceService(deviceRepo repository.DeviceRepository) DeviceService {
	return &deviceService{
		deviceRepo: deviceRepo,
	}
}
