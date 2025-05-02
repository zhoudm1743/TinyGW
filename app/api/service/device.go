package service

import (
	"fmt"
	"github.com/gookit/color"
	"tinyGW/app/api/repository"
	"tinyGW/app/api/schemas/req"
	"tinyGW/app/api/schemas/resp"
	"tinyGW/app/models"
	"tinyGW/pkg/plugin/response"
)

type DeviceService interface {
	Add(device *req.DeviceReq) error
	Update(device *req.DeviceReq) error
	Delete(name string) error
	Find(name string) (resp.DeviceResp, error)
	FindAll() ([]resp.DeviceResp, error)
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

func (d deviceService) Find(name string) (resp.DeviceResp, error) {
	f, err := d.deviceRepo.Find(name)
	if err != nil {
		return resp.DeviceResp{}, fmt.Errorf("仪表 %s 不存在", name)
	}
	var device resp.DeviceResp
	response.Copy(&device, f)
	return device, nil
}

func (d deviceService) FindAll() ([]resp.DeviceResp, error) {
	fs, err := d.deviceRepo.FindAll()
	if err != nil {
		return nil, err
	}
	var devices []resp.DeviceResp
	response.Copy(&devices, fs)
	return devices, nil
}

func (d deviceService) List(page *req.PageReq) (response.PageResp, error) {
	limit := page.PageSize
	offset := page.PageSize * (page.PageNo - 1)
	color.Infoln("limit: %d, offset: %d", limit, offset)
	list, total, err := d.deviceRepo.List(limit, offset)
	if err != nil {
		return response.PageResp{}, err
	}
	var devices []resp.DeviceResp
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
