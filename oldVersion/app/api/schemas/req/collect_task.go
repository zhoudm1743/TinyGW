package req

type CollectTaskReq struct {
	Name       string   `json:"name" binding:"required" form:"name"`
	Cron       string   `json:"cron" binding:"required" form:"cron"`
	Status     int8     `json:"status" form:"status"`
	DeviceList []string `json:"deviceList" form:"deviceList"`
}
