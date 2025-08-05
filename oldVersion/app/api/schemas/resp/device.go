package resp

import "tinyGW/pkg/plugin"

type DeviceResp struct {
	Name      string          `json:"name"`    // 设备名称
	Type      *DeviceTypeResp `json:"type"`    // 设备类型
	Address   string          `json:"address"` // 通讯地址
	Collector *CollectorResp  `json:"collector"`
	// Alone、Serial 可根据实际情况去掉，即一个采集接口有【固定】的波特率和驱动程序-----------
	Alone  bool   `json:"alone" form:"alone"` // 独立开关
	Serial Serial `json:"serial" form:"serial"`
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
	InitialVal float64       `json:"initialVal"`
	Scale      float64       `json:"scale"` // 倍率
	CreatedAt  plugin.TsTime `json:"createdAt" structs:"createdAt"`
	UpdatedAt  plugin.TsTime `json:"updatedAt" structs:"updatedAt"`
}

// DeviceProperty 设备属性
type DeviceProperty struct {
	Name        string      `json:"name" form:"name"`               // 名称，英文标识
	Description string      `json:"description" form:"description"` // 描述，中文涵义
	Type        string      `json:"type" form:"type"`               // 类型:int,long,double,string
	Length      int         `json:"length" form:"length"`           // 长度
	Decimal     int         `json:"decimal" form:"decimal"`         // 小数位
	Unit        string      `json:"unit" form:"unit"`               // 计量单位
	Value       interface{} `json:"value" form:"value"`             // 数值
	Reported    bool        `json:"reported" form:"reported"`       // 是否上报
	IsAlarm     bool        `json:"isAlarm" form:"isAlarm"`         // 是否报警
	Threshold   float64     `json:"threshold" form:"threshold"`     // 预警值 +-区间
	Used        float64     `json:"used" form:"used"`
	AutoCalc    bool        `json:"autoCalc" form:"autoCalc"`
	Scale       float64     `json:"scale" form:"scale"` // 倍率
}
type DeviceTypeResp struct {
	Name       string           `json:"name" form:"name" binding:"required"` // 名称，英文标识
	Driver     string           `json:"driver" form:"driver"`                // 驱动程序目录、文件名
	Properties []DeviceProperty `json:"properties" form:"properties"`
	CreatedAt  plugin.TsTime    `json:"createdAt" structs:"createdAt"`
	UpdatedAt  plugin.TsTime    `json:"updatedAt" structs:"updatedAt"`
}
