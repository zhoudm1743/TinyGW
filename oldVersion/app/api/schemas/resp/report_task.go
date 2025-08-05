package resp

import "tinyGW/pkg/plugin"

type ReportTaskResp struct {
	Name       string        `json:"name" structs:"name"`
	ReportName string        `json:"reportName" structs:"reportName"` // 平台名称
	Ip         string        `json:"ip" structs:"ip"`                 // ip地址
	Port       int           `json:"port" structs:"port"`             // 端口
	ClientID   string        `json:"clientID" structs:"clientID"`     // 网关编号
	Username   string        `json:"username" structs:"username"`     // 用户名
	Password   string        `json:"password" structs:"password"`     // 密码
	Cron       string        `json:"cron" structs:"cron"`             // cron表达式
	Status     int8          `json:"status" structs:"status"`         // 运行状态 0：停止 1：运行
	CreatedAt  plugin.TsTime `json:"createdAt" structs:"createdAt"`
	UpdatedAt  plugin.TsTime `json:"updatedAt" structs:"updatedAt"`
}
