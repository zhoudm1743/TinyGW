package repo

import (
	"TinyGW/models"
	"errors"

	"gorm.io/gorm"
)

type CollectTaskRepo struct {
	db *gorm.DB
}

func NewCollectTaskRepo(db *gorm.DB) *CollectTaskRepo {
	db.AutoMigrate(&models.CollectTask{})
	return &CollectTaskRepo{db: db}
}

// 创建采集任务
func (r *CollectTaskRepo) Create(task *models.CollectTask) error {
	var count int64
	r.db.Model(&models.CollectTask{}).Where("name = ?", task.Name).Count(&count)
	if count > 0 {
		return errors.New("采集任务已存在")
	}
	return r.db.Create(task).Error
}

// 更新采集任务
func (r *CollectTaskRepo) Update(task *models.CollectTask) error {
	var count int64
	r.db.Model(&models.CollectTask{}).Where("name = ?", task.Name).Count(&count)
	if count == 0 {
		return errors.New("采集任务不存在")
	}
	return r.db.Save(task).Error
}

// 删除采集任务
func (r *CollectTaskRepo) Delete(task *models.CollectTask) error {
	return r.db.Delete(task).Error
}

// 获取采集任务
func (r *CollectTaskRepo) Get(name string) (*models.CollectTask, error) {
	var task models.CollectTask
	if err := r.db.Where("name = ?", name).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// 获取采集任务列表
func (r *CollectTaskRepo) Find(page, pageSize int, conds map[string]interface{}) (int64, []models.CollectTask, error) {
	var tasks []models.CollectTask
	db := r.db.Model(&models.CollectTask{})
	for k, v := range conds {
		db = db.Where(k, v)
	}
	var count int64
	db.Count(&count)
	db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks)
	return count, tasks, nil
}

// 获取采集任务数量
func (r *CollectTaskRepo) Count(conds map[string]interface{}) (int64, error) {
	var count int64
	db := r.db.Model(&models.CollectTask{})
	for k, v := range conds {
		db = db.Where(k, v)
	}
	db.Count(&count)
	return count, nil
}
