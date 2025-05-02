package repository

import (
	"fmt"
	"gorm.io/gorm"
	"tinyGW/app/models"
)

type CollectTaskRepository interface {
	Save(collectTask *models.CollectTask) error
	Delete(name string) error
	Find(name string) (*models.CollectTask, error)
	FindAll() ([]models.CollectTask, error)
}

type collectTaskRepository struct {
	db *gorm.DB
}

func (c collectTaskRepository) Save(collectTask *models.CollectTask) error {
	var err error
	var task models.CollectTask
	c.db.Where("name = ?", collectTask.Name).First(&task)
	if task.ID != 0 {
		err = c.db.Model(&task).Save(&collectTask).Error
		return err
	} else {
		err = c.db.Create(collectTask).Error
		return err
	}
}

func (c collectTaskRepository) Delete(name string) error {
	return c.db.Delete(&models.CollectTask{}, "name = ?", name).Error
}

func (c collectTaskRepository) Find(name string) (*models.CollectTask, error) {
	var collectTask models.CollectTask
	c.db.Where("name =?", name).First(&collectTask)
	if collectTask.ID == 0 {
		return nil, fmt.Errorf("collect task not found")
	}
	return &collectTask, nil
}

func (c collectTaskRepository) FindAll() ([]models.CollectTask, error) {
	var collectTasks []models.CollectTask
	err := c.db.Find(&collectTasks).Error
	if err != nil {
		return nil, err
	}
	return collectTasks, nil
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
