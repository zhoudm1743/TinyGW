package models

type Device struct {
	Model
	Name        string      `gorm:"type:varchar(255);primary_key;" json:"name"`
	TypeID      string      `json:"typeID" gorm:"type:varchar(255);index:idx_devices_type_id"` // 设备类型ID
	Type        *DeviceType `json:"type" gorm:"foreignKey:TypeID"`
	Address     string      `json:"address" gorm:"type:varchar(255);index:idx_devices_address"`          // 通讯地址
	CollectorID string      `json:"collectorID" gorm:"type:varchar(255);index:idx_devices_collector_id"` // 采集器ID
	Collector   *Collector  `json:"collector" gorm:"foreignKey:CollectorID"`
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
