package req

type DeviceReq struct {
	Name        string  `json:"name" binding:"required"`
	TypeID      string  `json:"typeID" binding:"required"`
	Address     string  `json:"address" binding:"required"`
	CollectorID string  `json:"collectorID" binding:"required"`
	InitialVal  float64 `json:"initialVal"`
	Scale       float64 `json:"scale"`
}
