package models

type DeviceLog struct {
	Model
	DeviceID  string  `json:"deviceID" gorm:"type:varchar(255);index:idx_device_logs_device_id"` // 设备ID
	Device    *Device `json:"device" gorm:"foreignKey:DeviceID"`
	Time      int64   `json:"time"`      // 时间
	OriginCmd string  `json:"originCmd"` // 原始命令
	OriginRes string  `json:"originRes"` // 原始响应
	Res       string  `json:"res"`       // 响应
	Property  string  `json:"property"`  // 属性
	Value     string  `json:"value"`     // 值
	Status    int     `json:"status"`    // 状态
	Error     string  `json:"error"`     // 错误
}
