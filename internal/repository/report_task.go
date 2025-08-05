package repository

import (
	"TinyGW/models"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReportTaskRepository interface {
	Save(reportTask *models.ReportTask) error
	Delete(name string) error
	Find(name string) (models.ReportTask, error)
	FindAll() ([]models.ReportTask, error)
	List(limit int, offset int) ([]models.ReportTask, int64, error)
}

type reportTaskRepository struct {
	db *gorm.DB
}

func (c *reportTaskRepository) Save(reportTask *models.ReportTask) error {
	var err error
	var task models.ReportTask
	if err = c.db.Where("name = ?", reportTask.Name).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = c.db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "name"}},
				UpdateAll: true,
			}).Create(reportTask).Error
			return err
		}
	}
	return c.db.Save(reportTask).Error
}

func (c *reportTaskRepository) Delete(name string) error {
	return c.db.Delete(&models.ReportTask{}, "name = ?", name).Error
}

func (c *reportTaskRepository) Find(name string) (models.ReportTask, error) {
	var reportTask models.ReportTask
	err := c.db.Where("name =?", name).First(&reportTask).Error
	return reportTask, err
}

func (c *reportTaskRepository) FindAll() ([]models.ReportTask, error) {
	var reportTasks []models.ReportTask
	err := c.db.Find(&reportTasks).Error
	return reportTasks, err
}

func (c *reportTaskRepository) List(limit int, offset int) ([]models.ReportTask, int64, error) {
	var reportTasks []models.ReportTask
	var count int64
	c.db.Model(&models.ReportTask{}).Count(&count)
	err := c.db.Limit(limit).Offset(offset).Find(&reportTasks).Error
	return reportTasks, count, err
}

func NewReportTaskRepository(db *gorm.DB) ReportTaskRepository {
	db.AutoMigrate(&models.ReportTask{})
	var count int64
	db.Model(&models.ReportTask{}).Count(&count)
	if count == 0 {
		db.Create(&models.ReportTask{
			Name:       "数据上报",
			ReportName: "西奥物联网平台",
			Ip:         "127.0.0.1",
			Port:       1883,
			ClientID:   "default-client",
			Username:   "admin",
			Password:   "admin",
			Cron:       "0 5 * * *",
			Status:     0,
		})
	}
	return &reportTaskRepository{db: db}
}
