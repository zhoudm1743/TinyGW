package req

type CollectTaskReq struct {
	Name       string   `json:"name" binding:"required"`
	Cron       string   `json:"cron" binding:"required"`
	Status     int8     `json:"status"`
	DeviceList []string `json:"deviceList"`
}
