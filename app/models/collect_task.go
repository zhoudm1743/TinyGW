package models

type CollectTask struct {
	Model
	Name       string   `json:"name" gorm:"type:varchar(255);unique;not null;index:idx_collect_task_name"`
	Cron       string   `json:"cron"`   // 定时策略
	Status     int8     `json:"status"` // 状态
	DeviceList []string `json:"deviceList" gorm:"type:json;serializer:json"`
}
