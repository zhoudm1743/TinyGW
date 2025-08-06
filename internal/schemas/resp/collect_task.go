package resp

type CollectTaskResp struct {
	Name       string   `json:"name"`
	Cron       string   `json:"cron"`
	Status     int8     `json:"status"`
	DeviceList []string `json:"deviceList"`
	CreatedAt  int64    `json:"createdAt"`
	UpdatedAt  int64    `json:"updatedAt"`
}
