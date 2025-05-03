package repository

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"tinyGW/app/models"
	"tinyGW/pkg/service/conf"
	"tinyGW/pkg/service/event"
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
	event  *event.EventService
}

func (c reportTaskRepository) Save(reportTask *models.ReportTask) error {
	var err error
	var task models.ReportTask
	if err = c.db.Where("name = ?", reportTask.Name).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = c.db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "name"}},
				UpdateAll: true,
			}).Create(reportTask).Error
			if err != nil {
				return err
			}
			c.event.Publish(event.Event{
				Name: "ReportTask_Add",
				Data: reportTask,
			})
			return nil
		}
	}
	err = c.db.Model(&task).Where("name = ?", reportTask.Name).Updates(reportTask).Error
	if err != nil {
		return fmt.Errorf("更新采集任务失败: %s", err.Error())
	}
	c.event.Publish(event.Event{
		Name: "ReportTask_Update",
		Data: reportTask,
	})

	if len(task.Name) != 0 {
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

	return err
}

func (c reportTaskRepository) Delete(name string) error {
	err := c.db.Delete(&models.ReportTask{}, "name = ?", name).Error
	if err != nil {
		return fmt.Errorf("删除采集任务失败: %s", err.Error())
	}
	c.event.Publish(event.Event{
		Name: "ReportTask_Delete",
		Data: name,
	})
	return nil
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

func NewReportTaskRepository(db *gorm.DB, config *conf.Config, event *event.EventService) ReportTaskRepository {
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
		event:  event,
	}
}
