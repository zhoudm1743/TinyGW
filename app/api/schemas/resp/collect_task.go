package resp

import "tinyGW/pkg/plugin"

type CollectTaskResp struct {
	Name       string        `json:"name" structs:"name"`
	Cron       string        `json:"cron" structs:"cron"`
	Status     int8          `json:"status" structs:"status"`
	DeviceList []string      `json:"deviceList" structs:"deviceList"`
	CreatedAt  plugin.TsTime `json:"createdAt" structs:"createdAt"`
	UpdatedAt  plugin.TsTime `json:"updatedAt" structs:"updatedAt"`
}
