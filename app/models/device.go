package models

type Device struct {
	Model
	Name      string      `json:"name" gorm:"type:varchar(255);not null;unique;index:idx_devices_name"` // 设备名称
	Type      *DeviceType `json:"type" gorm:"type:json;serializer:json"`                                // 设备类型
	Address   string      `json:"address" gorm:"type:varchar(255);index:idx_devices_address"`           // 通讯地址
	Collector *Collector  `json:"collector" gorm:"type:json;serializer:json"`
	// Alone、Serial 可根据实际情况去掉，即一个采集接口有【固定】的波特率和驱动程序-----------
	Alone  bool   `json:"alone"` // 独立开关
	Serial Serial `json:"serial" gorm:"type:json;serializer:json"`
	//----------------------------------------------------------------------------
	Online         bool  `json:"online"`         // 设备在线
	CollectTime    int64 `json:"collectTime"`    // 最后采集时间
	CollectTotal   int   `json:"collectTotal"`   // 总采集次数
	CollectSuccess int   `json:"collectSuccess"` // 采集成功次数
	//----------------------------------------------------------------------------
	ReportTime    int64 `json:"reportTime"`    // 最后上报时间
	ReportTotal   int   `json:"reportTotal"`   // 总上报次数
	ReportSuccess int   `json:"reportSuccess"` // 上报成功次数
	//----------------------------------------------------------------------------
	AlarmStatus bool   `json:"alarmStatus"` // 报警状态
	AlarmReason string `json:"alarmReason"` // 报警原因
	AlarmTime   int64  `json:"alarmTime"`   // 报警时间
	AlarmTotal  int    `json:"alarmTotal"`  // 报警次数
	//----------------------------------------------------------------------------
	InitialVal float64 `json:"initialVal"`
	Scale      float64 `json:"scale"` // 倍率
}
