package models

type CollectTask struct {
	Model
	Name       string   `json:"name" gorm:"type:varchar(255);unique;not null"`
	Cron       string   `json:"cron"`   // 定时策略
	Status     int8     `json:"status"` // 状态
	DeviceList []string `json:"deviceList" gorm:"type:json;serializer:json"`
}
