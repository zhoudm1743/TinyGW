package repo

import (
	"TinyGW/models"
	"errors"

	"gorm.io/gorm"
)

type DeviceRepo struct {
	db *gorm.DB
}

func NewDeviceRepo(db *gorm.DB) *DeviceRepo {
	db.AutoMigrate(&models.Device{})
	return &DeviceRepo{db: db}
}

// 创建设备
func (r *DeviceRepo) Create(device *models.Device) error {
	var count int64
	r.db.Model(&models.Device{}).Where("name = ?", device.Name).Count(&count)
	if count > 0 {
		return errors.New("设备已存在")
	}
	return r.db.Create(device).Error
}

// 更新设备
func (r *DeviceRepo) Update(device *models.Device) error {
	var count int64
	r.db.Model(&models.Device{}).Where("name = ?", device.Name).Count(&count)
	if count == 0 {
		return errors.New("设备不存在")
	}
	return r.db.Save(device).Error
}

// 删除设备
func (r *DeviceRepo) Delete(device *models.Device) error {
	return r.db.Delete(device).Error
}

// 获取设备
func (r *DeviceRepo) Get(name string) (*models.Device, error) {
	var device models.Device
	if err := r.db.Where("name = ?", name).First(&device).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

// 获取设备列表
func (r *DeviceRepo) Find(page, pageSize int, conds map[string]interface{}) (int64, []models.Device, error) {
	var devices []models.Device
	db := r.db.Model(&models.Device{})
	for k, v := range conds {
		db = db.Where(k, v)
	}
	var count int64
	db.Count(&count)
	db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&devices)
	return count, devices, nil
}

// 获取设备数量
func (r *DeviceRepo) Count(conds map[string]interface{}) (int64, error) {
	var count int64
	db := r.db.Model(&models.Device{})
	for k, v := range conds {
		db = db.Where(k, v)
	}
	db.Count(&count)
	return count, nil
}
