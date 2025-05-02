package repository

import (
	"errors"
	"gorm.io/gorm"
	"tinyGW/app/models"
)

type CollectTaskRepository interface {
	Save(collectTask *models.CollectTask) error
	Delete(name string) error
	Find(name string) (models.CollectTask, error)
	FindAll() ([]models.CollectTask, error)
	List(limit int, offset int) ([]models.CollectTask, int64, error)
}

type collectTaskRepository struct {
	db *gorm.DB
}

func (c collectTaskRepository) Save(collectTask *models.CollectTask) error {
	var err error
	var task models.CollectTask
	if err = c.db.Where("name = ?", collectTask.Name).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = c.db.Create(collectTask).Error
			return err
		}
	}
	err = c.db.Model(&task).Where("name = ?", collectTask.Name).Updates(collectTask).Error
	return err
}

func (c collectTaskRepository) Delete(name string) error {
	return c.db.Delete(&models.CollectTask{}, "name = ?", name).Error
}

func (c collectTaskRepository) Find(name string) (models.CollectTask, error) {
	var collectTask models.CollectTask
	err := c.db.Where("name =?", name).First(&collectTask).Error
	return collectTask, err
}

func (c collectTaskRepository) FindAll() ([]models.CollectTask, error) {
	var collectTasks []models.CollectTask
	err := c.db.Find(&collectTasks).Error
	if err != nil {
		return nil, err
	}
	return collectTasks, nil
}

func (c collectTaskRepository) List(limit int, offset int) ([]models.CollectTask, int64, error) {
	var collectTasks []models.CollectTask
	var count int64
	c.db.Model(&models.CollectTask{}).Count(&count)
	err := c.db.Limit(limit).Offset(offset).Find(&collectTasks).Error
	if err != nil {
		return nil, 0, err
	}
	return collectTasks, count, nil
}

func NewCollectTaskRepository(db *gorm.DB) CollectTaskRepository {
	db.AutoMigrate(&models.CollectTask{})
	var count int64
	db.Model(&models.CollectTask{}).Count(&count)
	if count == 0 {
		db.Create(&models.CollectTask{
			Name:       "数据采集",
			Cron:       "0 2 * * *",
			Status:     0,
			DeviceList: []string{"*"},
		})
	}
	return &collectTaskRepository{
		db: db,
	}
}
