package repository

import (
	"fmt"
	"gorm.io/gorm"
	"tinyGW/app/models"
)

type CollectorRepository interface {
	Save(collector *models.Collector) error
	Delete(name string) error
	Find(name string) (*models.Collector, error)
	FindAll() ([]models.Collector, error)
}

type collectorRepository struct {
	db *gorm.DB
}

func (c collectorRepository) Save(collector *models.Collector) error {
	var err error
	var task models.Collector
	c.db.Where("name = ?", collector.Name).First(&task)
	if task.ID != 0 {
		err = c.db.Model(&task).Save(&collector).Error
		return err
	} else {
		err = c.db.Create(collector).Error
		return err
	}
}

func (c collectorRepository) Delete(name string) error {
	return c.db.Delete(&models.Collector{}, "name = ?", name).Error
}

func (c collectorRepository) Find(name string) (*models.Collector, error) {
	var collector models.Collector
	c.db.Where("name =?", name).First(&collector)
	if collector.ID == 0 {
		return nil, fmt.Errorf("collect task not found")
	}
	return &collector, nil
}

func (c collectorRepository) FindAll() ([]models.Collector, error) {
	var collectors []models.Collector
	err := c.db.Find(&collectors).Error
	if err != nil {
		return nil, err
	}
	return collectors, nil
}

func NewCollectorRepository(db *gorm.DB) CollectorRepository {
	db.AutoMigrate(&models.Collector{})
	return &collectorRepository{
		db: db,
	}
}
