package repo

import (
	"TinyGW/models"

	"gorm.io/gorm"
)

type DeviceLogRepo struct {
	db *gorm.DB
}

func NewDeviceLogRepo(db *gorm.DB) *DeviceLogRepo {
	db.AutoMigrate(&models.DeviceLog{})
	return &DeviceLogRepo{db: db}
}

// 创建设备日志
func (r *DeviceLogRepo) Create(log *models.DeviceLog) error {
	return r.db.Create(log).Error
}

// 删除设备日志
func (r *DeviceLogRepo) Delete(log *models.DeviceLog) error {
	return r.db.Delete(log).Error
}

// 获取设备日志列表
func (r *DeviceLogRepo) Find(page, pageSize int, conds map[string]interface{}) (int64, []models.DeviceLog, error) {
	var logs []models.DeviceLog
	db := r.db.Model(&models.DeviceLog{})
	for k, v := range conds {
		db = db.Where(k, v)
	}
	var count int64
	db.Count(&count)
	db.Offset((page - 1) * pageSize).Limit(pageSize).Order("time desc").Find(&logs)
	return count, logs, nil
}

// 获取设备日志数量
func (r *DeviceLogRepo) Count(conds map[string]interface{}) (int64, error) {
	var count int64
	db := r.db.Model(&models.DeviceLog{})
	for k, v := range conds {
		db = db.Where(k, v)
	}
	db.Count(&count)
	return count, nil
}

// 清除指定时间之前的日志
func (r *DeviceLogRepo) ClearBefore(time int64) error {
	return r.db.Where("time < ?", time).Delete(&models.DeviceLog{}).Error
}
