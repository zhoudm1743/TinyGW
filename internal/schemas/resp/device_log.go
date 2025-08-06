package resp

type DeviceLogResp struct {
	ID         uint   `json:"id"`
	DeviceID   string `json:"deviceID"`
	DeviceName string `json:"deviceName,omitempty"`
	Time       int64  `json:"time"`
	OriginCmd  string `json:"originCmd"`
	OriginRes  string `json:"originRes"`
	Res        string `json:"res"`
	Property   string `json:"property"`
	Value      string `json:"value"`
	Status     int    `json:"status"`
	Error      string `json:"error"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}
