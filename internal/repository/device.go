package repository

import (
	"TinyGW/models"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DeviceRepository interface {
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

type deviceRepository struct {
	db *gorm.DB
}

func (c *deviceRepository) CollectorIsUsed(collector *models.Collector) bool {
	if devices, err := c.FindAll(); err == nil {
		for _, device := range devices {
			if device.Collector.Name == collector.Name {
				return true
			}
		}
	}
	return false
}

func (c *deviceRepository) DeviceTypeIsUsed(deviceType *models.DeviceType) bool {
	if devices, err := c.FindAll(); err == nil {
		for _, device := range devices {
			if device.Type.Name == deviceType.Name {
				return true
			}
		}
	}
	return false
}

func (c *deviceRepository) CollectorChanged(collector models.Collector) {
	if devices, err := c.FindAll(); err == nil {
		for _, device := range devices {
			if device.Collector.Name == collector.Name {
				device.Collector = collector
				c.Save(&device)
			}
		}
	}
}

func (c *deviceRepository) DeviceTypeChanged(deviceType models.DeviceType) {
	if devices, err := c.FindAll(); err == nil {
		for _, device := range devices {
			if device.Type.Name == deviceType.Name {
				device.Type = deviceType
				c.Save(&device)
			}
		}
	}
}

func (c *deviceRepository) Save(device *models.Device) error {
	var err error
	var task models.Device
	if err = c.db.Where("name = ?", device.Name).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = c.db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "name"}},
				UpdateAll: true,
			}).Create(device).Error
			return err
		}
	}
	return c.db.Save(device).Error
}

func (c *deviceRepository) Delete(name string) error {
	return c.db.Delete(&models.Device{}, "name = ?", name).Error
}

func (c *deviceRepository) Find(name string) (models.Device, error) {
	var device models.Device
	err := c.db.Where("name = ?", name).First(&device).Error
	return device, err
}

func (c *deviceRepository) FindAll() ([]models.Device, error) {
	var devices []models.Device
	err := c.db.Find(&devices).Error
	return devices, err
}

func (c *deviceRepository) List(limit int, offset int) ([]models.Device, int64, error) {
	var devices []models.Device
	var count int64
	c.db.Model(&models.Device{}).Count(&count)
	err := c.db.Model(&models.Device{}).Limit(limit).Offset(offset).Find(&devices).Error
	return devices, count, err
}

func (c *deviceRepository) FindByAddr(addr string) (models.Device, error) {
	var device models.Device
	err := c.db.Where("address = ?", addr).First(&device).Error
	return device, err
}

func NewDeviceRepository(db *gorm.DB) DeviceRepository {
	db.AutoMigrate(&models.Device{})
	return &deviceRepository{db: db}
}
