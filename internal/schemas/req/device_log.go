package req

type DeviceLogReq struct {
	DeviceID  string `json:"deviceID" binding:"required"`
	Time      int64  `json:"time" binding:"required"`
	OriginCmd string `json:"originCmd"`
	OriginRes string `json:"originRes"`
	Res       string `json:"res"`
	Property  string `json:"property"`
	Value     string `json:"value"`
	Status    int    `json:"status"`
	Error     string `json:"error"`
}
