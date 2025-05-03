package repository

import (
	"errors"
	"github.com/gookit/color"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"tinyGW/app/models"
)

type DeviceTypeRepository interface {
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

type deviceTypeRepository struct {
	db *gorm.DB
}

func (c deviceTypeRepository) AddProperties(deviceTypeName string, deviceProperty *models.DeviceProperty) error {
	dt, err := c.Find(deviceTypeName)
	if err != nil {
		return err
	}
	dt.Properties = append(dt.Properties, *deviceProperty)

	return c.Save(&dt)
}

func (c deviceTypeRepository) UpdateProperties(deviceTypeName string, devicePropertyId int, deviceProperty *models.DeviceProperty) error {
	dt, err := c.Find(deviceTypeName)
	if err != nil {
		return err
	}
	dt.Properties[devicePropertyId-1] = *deviceProperty
	return c.Save(&dt)
}

func (c deviceTypeRepository) DeleteProperties(deviceTypeName string, devicePropertyId int) error {
	dt, err := c.Find(deviceTypeName)
	if err != nil {
		return err
	}
	dt.Properties = append(dt.Properties[:devicePropertyId-1], dt.Properties[devicePropertyId:]...)
	return c.Save(&dt)
}

func (c deviceTypeRepository) FindProperty(deviceTypeName string, devicePropertyId int) (models.DeviceProperty, error) {
	dt, err := c.Find(deviceTypeName)
	if err != nil {
		return models.DeviceProperty{}, err
	}
	return dt.Properties[devicePropertyId-1], nil
}

func (c deviceTypeRepository) FindAllProperties(deviceTypeName string) ([]models.DeviceProperty, error) {
	dt, err := c.Find(deviceTypeName)
	if err != nil {
		return nil, err
	}
	return dt.Properties, nil
}

func (c deviceTypeRepository) Save(deviceType *models.DeviceType) error {
	var err error
	var task models.DeviceType
	if err = c.db.Where("name = ?", deviceType.Name).First(&task).Error; err != nil {
		color.Infoln(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = c.db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "name"}},
				UpdateAll: true,
			}).Create(deviceType).Error
			return err
		}
	}
	err = c.db.Model(&task).Where("name = ?", deviceType.Name).Updates(deviceType).Error
	return err
}

func (c deviceTypeRepository) Delete(name string) error {
	return c.db.Delete(&models.DeviceType{}, "name = ?", name).Error
}

func (c deviceTypeRepository) Find(name string) (models.DeviceType, error) {
	var deviceType models.DeviceType
	err := c.db.Where("name = ?", name).First(&deviceType).Error
	return deviceType, err
}

func (c deviceTypeRepository) FindAll() ([]models.DeviceType, error) {
	var deviceTypes []models.DeviceType
	err := c.db.Find(&deviceTypes).Error
	if err != nil {
		return nil, err
	}
	return deviceTypes, nil
}

func (c deviceTypeRepository) List(limit int, offset int) ([]models.DeviceType, int64, error) {
	var deviceTypes []models.DeviceType
	var count int64
	err := c.db.Model(&models.DeviceType{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	err = c.db.Model(&models.DeviceType{}).Limit(limit).Offset(offset).Find(&deviceTypes).Error
	if err != nil {
		return nil, 0, err
	}
	return deviceTypes, count, nil
}

func NewDeviceTypeRepository(db *gorm.DB) DeviceTypeRepository {
	db.AutoMigrate(&models.DeviceType{})
	return &deviceTypeRepository{
		db: db,
	}
}
