package repository

import (
	"errors"
	"gorm.io/gorm"
	"tinyGW/app/models"
	"tinyGW/pkg/service/conf"
)

type ReportTaskRepository interface {
	Save(reportTask *models.ReportTask) error
	Delete(name string) error
	Find(name string) (models.ReportTask, error)
	FindAll() ([]models.ReportTask, error)
	List(limit int, offset int) ([]models.ReportTask, int64, error)
}

type reportTaskRepository struct {
	db     *gorm.DB
	config *conf.Config
}

func (c reportTaskRepository) Save(reportTask *models.ReportTask) error {
	var err error
	var task models.ReportTask
	if err = c.db.Where("name = ?", reportTask.Name).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = c.db.Create(reportTask).Error
			return err
		}
	}
	err = c.db.Model(&task).Where("name = ?", reportTask.Name).Updates(reportTask).Error
	return err
	if err == nil {
		if task.ID != 0 {
			err = c.db.Model(&task).Save(&reportTask).Error
		} else {
			err = c.db.Create(reportTask).Error
		}
		if reportTask.Ip != "" { // 校验IP非空
			c.config.Cloud.Host = reportTask.Ip
		}
		if reportTask.Port > 0 { // 严格校验TCP端口有效性
			c.config.Cloud.Port = reportTask.Port
		}
		if reportTask.ClientID != "" { // 校验客户端ID非空
			c.config.Cloud.ClientId = reportTask.ClientID
		}
		if reportTask.Username != "" { // 校验用户名非空
			c.config.Cloud.Username = reportTask.Username
		}
		if reportTask.Password != "" { // 校验密码非空
			c.config.Cloud.Password = reportTask.Password
		}
		conf.SaveConfig(c.config)
	}
	return err
}

func (c reportTaskRepository) Delete(name string) error {
	return c.db.Delete(&models.ReportTask{}, "name = ?", name).Error
}

func (c reportTaskRepository) Find(name string) (models.ReportTask, error) {
	var reportTask models.ReportTask
	err := c.db.Where("name =?", name).First(&reportTask).Error
	return reportTask, err
}

func (c reportTaskRepository) FindAll() ([]models.ReportTask, error) {
	var reportTasks []models.ReportTask
	err := c.db.Find(&reportTasks).Error
	if err != nil {
		return nil, err
	}
	return reportTasks, nil
}

func (c reportTaskRepository) List(limit int, offset int) ([]models.ReportTask, int64, error) {
	var reportTasks []models.ReportTask
	var count int64
	c.db.Model(&models.ReportTask{}).Count(&count)
	err := c.db.Limit(limit).Offset(offset).Find(&reportTasks).Error
	if err != nil {
		return nil, 0, err
	}
	return reportTasks, count, nil
}

func NewReportTaskRepository(db *gorm.DB, config *conf.Config) ReportTaskRepository {
	db.AutoMigrate(&models.ReportTask{})
	var count int64
	db.Model(&models.ReportTask{}).Count(&count)
	if count == 0 {
		db.Create(&models.ReportTask{
			Name:       "数据上报",
			ReportName: "西奥物联网平台",
			Ip:         config.Cloud.Host,
			Port:       config.Cloud.Port,
			ClientID:   config.Cloud.ClientId,
			Username:   config.Cloud.Username,
			Password:   config.Cloud.Password,
			Cron:       "0 5 * * *",
			Status:     0,
		})
	}
	return &reportTaskRepository{
		db:     db,
		config: config,
	}
}
