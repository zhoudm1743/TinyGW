package models

// ReportTask 定时任务表
type ReportTask struct {
	Model
	Name       string `json:"name" gorm:"type:varchar(255);unique;not null;index:idx_report_name"`
	ReportName string `json:"reportName"` // 平台名称
	Ip         string `json:"ip"`         // ip地址
	Port       int    `json:"port"`       // 端口
	ClientID   string `json:"clientID"`   // 网关编号
	Username   string `json:"username"`   // 用户名
	Password   string `json:"password"`   // 密码
	Cron       string `json:"cron"`       // cron表达式
	Status     int8   `json:"status"`     // 运行状态 0：停止 1：运行
}
