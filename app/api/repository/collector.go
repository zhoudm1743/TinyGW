package repository

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"tinyGW/app/models"
	"tinyGW/pkg/service/event"
)

type CollectorRepository interface {
	Save(collector *models.Collector) error
	Delete(name string) error
	Find(name string) (models.Collector, error)
	FindAll() ([]models.Collector, error)
	List(limit int, offset int) ([]models.Collector, int64, error)
}

type collectorRepository struct {
	db    *gorm.DB
	event *event.EventService
}

func (c collectorRepository) Save(collector *models.Collector) error {
	var err error
	var task models.Collector
	if err = c.db.Where("name = ?", collector.Name).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = c.db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "name"}},
				UpdateAll: true,
			}).Create(collector).Error
			if err != nil {
				return err
			}
			c.event.Publish(event.Event{
				Name: "collector_add",
				Data: *collector,
			})
			return nil
		}
	}
	err = c.db.Save(collector).Error
	if err != nil {
		return err
	}
	c.event.Publish(event.Event{
		Name: "collector_update",
		Data: *collector,
	})
	return nil
}

func (c collectorRepository) Delete(name string) error {
	err := c.db.Delete(&models.Collector{}, "name = ?", name).Error
	if err != nil {
		return err
	}
	c.event.Publish(event.Event{
		Name: "collector_delete",
		Data: name,
	})
	return nil
}

func (c collectorRepository) Find(name string) (models.Collector, error) {
	var collector models.Collector
	err := c.db.Where("name =?", name).First(&collector).Error
	return collector, err
}

func (c collectorRepository) FindAll() ([]models.Collector, error) {
	var collectors []models.Collector
	err := c.db.Find(&collectors).Error
	if err != nil {
		return nil, err
	}
	return collectors, nil
}

func (c collectorRepository) List(limit int, offset int) ([]models.Collector, int64, error) {
	var collectors []models.Collector
	var total int64
	c.db.Model(&models.Collector{}).Count(&total)
	err := c.db.Limit(limit).Offset(offset).Find(&collectors).Error
	if err != nil {
		return nil, 0, err
	}
	return collectors, total, nil
}

func NewCollectorRepository(db *gorm.DB, event *event.EventService) CollectorRepository {
	db.AutoMigrate(&models.Collector{})
	return &collectorRepository{
		db:    db,
		event: event,
	}
}
