package repo

import (
	"TinyGW/models"
	"errors"

	"gorm.io/gorm"
)

type CollectorRepo struct {
	db *gorm.DB
}

func NewCollectorRepo(db *gorm.DB) *CollectorRepo {
	db.AutoMigrate(&models.Collector{})
	return &CollectorRepo{db: db}
}

// 创建采集器
func (r *CollectorRepo) Create(collector *models.Collector) error {
	var count int64
	r.db.Model(&models.Collector{}).Where("name = ?", collector.Name).Count(&count)
	if count > 0 {
		return errors.New("采集器已存在")
	}
	return r.db.Create(collector).Error
}

// 更新采集器
func (r *CollectorRepo) Update(collector *models.Collector) error {
	var count int64
	r.db.Model(&models.Collector{}).Where("name = ?", collector.Name).Count(&count)
	if count == 0 {
		return errors.New("采集器不存在")
	}
	return r.db.Save(collector).Error
}

// 删除采集器
func (r *CollectorRepo) Delete(collector *models.Collector) error {
	return r.db.Delete(collector).Error
}

// 获取采集器
func (r *CollectorRepo) Get(name string) (*models.Collector, error) {
	var collector models.Collector
	if err := r.db.Where("name = ?", name).First(&collector).Error; err != nil {
		return nil, err
	}
	return &collector, nil
}

// 获取采集器列表
func (r *CollectorRepo) Find(page, pageSize int, conds map[string]interface{}) (int64, []models.Collector, error) {
	var collectors []models.Collector
	db := r.db.Model(&models.Collector{})
	for k, v := range conds {
		db = db.Where(k, v)
	}
	var count int64
	db.Count(&count)
	db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&collectors)
	return count, collectors, nil
}

// 获取采集器数量
func (r *CollectorRepo) Count(conds map[string]interface{}) (int64, error) {
	var count int64
	db := r.db.Model(&models.Collector{})
	for k, v := range conds {
		db = db.Where(k, v)
	}
	db.Count(&count)
	return count, nil
}
