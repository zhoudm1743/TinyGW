package repo

import (
	"TinyGW/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type DeviceTypeRepo struct {
	db *gorm.DB
}

func NewDeviceTypeRepo(db *gorm.DB) *DeviceTypeRepo {
	db.AutoMigrate(&models.DeviceType{})
	return &DeviceTypeRepo{db: db}
}

// 创建设备类型
func (r *DeviceTypeRepo) Create(deviceType *models.DeviceType) error {
	var count int64
	r.db.Model(&models.DeviceType{}).Where("name = ?", deviceType.Name).Count(&count)
	if count > 0 {
		return errors.New("设备类型已存在")
	}
	return r.db.Create(deviceType).Error
}

// 更新设备类型
func (r *DeviceTypeRepo) Update(deviceType *models.DeviceType) error {
	var count int64
	r.db.Model(&models.DeviceType{}).Where("name = ?", deviceType.Name).Count(&count)
	if count == 0 {
		return errors.New("设备类型不存在")
	}
	return r.db.Save(deviceType).Error
}

// 删除设备类型
func (r *DeviceTypeRepo) Delete(deviceType *models.DeviceType) error {
	return r.db.Delete(deviceType).Error
}

// 获取设备类型
func (r *DeviceTypeRepo) Get(name string) (*models.DeviceType, error) {
	var deviceType models.DeviceType
	if err := r.db.Where("name = ?", name).First(&deviceType).Error; err != nil {
		return nil, err
	}
	return &deviceType, nil
}

// 获取设备类型列表
func (r *DeviceTypeRepo) Find(page, pageSize int, conds map[string]interface{}) (int64, []models.DeviceType, error) {
	var deviceTypes []models.DeviceType
	db := r.db.Model(&models.DeviceType{})
	for k, v := range conds {
		db = db.Where(k, v)
	}
	var count int64
	db.Count(&count)
	db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&deviceTypes)
	return count, deviceTypes, nil
}

// 获取设备类型数量
func (r *DeviceTypeRepo) Count(conds map[string]interface{}) (int64, error) {
	var count int64
	db := r.db.Model(&models.DeviceType{})
	for k, v := range conds {
		db = db.Where(k, v)
	}
	db.Count(&count)
	return count, nil
}

// 获取所有设备类型
func (r *DeviceTypeRepo) FindAll() ([]models.DeviceType, error) {
	var deviceTypes []models.DeviceType
	if err := r.db.Find(&deviceTypes).Error; err != nil {
		return nil, err
	}
	return deviceTypes, nil
}

// 添加设备类型属性
func (r *DeviceTypeRepo) AddProperty(deviceTypeName string, property *models.DeviceProperty) error {
	deviceType, err := r.Get(deviceTypeName)
	if err != nil {
		return errors.New("设备类型不存在")
	}

	// 检查属性名是否已存在
	for _, p := range deviceType.Properties {
		if p.Name == property.Name {
			return fmt.Errorf("属性 %s 已存在", property.Name)
		}
	}

	deviceType.Properties = append(deviceType.Properties, *property)
	return r.Update(deviceType)
}

// 更新设备类型属性
func (r *DeviceTypeRepo) UpdateProperty(deviceTypeName string, propertyName string, property *models.DeviceProperty) error {
	deviceType, err := r.Get(deviceTypeName)
	if err != nil {
		return errors.New("设备类型不存在")
	}

	found := false
	for i, p := range deviceType.Properties {
		if p.Name == propertyName {
			deviceType.Properties[i] = *property
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("属性 %s 不存在", propertyName)
	}

	return r.Update(deviceType)
}

// 删除设备类型属性
func (r *DeviceTypeRepo) DeleteProperty(deviceTypeName string, propertyName string) error {
	deviceType, err := r.Get(deviceTypeName)
	if err != nil {
		return errors.New("设备类型不存在")
	}

	found := false
	var newProperties []models.DeviceProperty
	for _, p := range deviceType.Properties {
		if p.Name == propertyName {
			found = true
			continue
		}
		newProperties = append(newProperties, p)
	}

	if !found {
		return fmt.Errorf("属性 %s 不存在", propertyName)
	}

	deviceType.Properties = newProperties
	return r.Update(deviceType)
}

// 获取设备类型属性
func (r *DeviceTypeRepo) GetProperty(deviceTypeName string, propertyName string) (*models.DeviceProperty, error) {
	deviceType, err := r.Get(deviceTypeName)
	if err != nil {
		return nil, errors.New("设备类型不存在")
	}

	for _, p := range deviceType.Properties {
		if p.Name == propertyName {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("属性 %s 不存在", propertyName)
}

// 获取设备类型所有属性
func (r *DeviceTypeRepo) GetAllProperties(deviceTypeName string) ([]models.DeviceProperty, error) {
	deviceType, err := r.Get(deviceTypeName)
	if err != nil {
		return nil, errors.New("设备类型不存在")
	}

	return deviceType.Properties, nil
}
